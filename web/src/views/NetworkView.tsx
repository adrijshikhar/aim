import * as React from 'react';
import {
  Radio,
  Clock,
  Laptop,
  RotateCw,
  Trash2,
  Copy,
  Check,
  KeyRound,
  Search,
  Plus,
  X,
  Share2,
  ShieldAlert,
  Activity,
  Calendar,
  Inbox,
  FileCheck,
  Ban,
} from 'lucide-react';
import {
  getShareStatus,
  setGlobalShareMode,
  toggleShareProfile,
  revokeShareLease,
  addSharePeer,
  requestShareLease,
  releaseShareLease,
  cleanExpiredLeases,
  saveSchedule,
  deleteSchedule,
  approveShareRequest,
  denyShareRequest,
  ShareStatusDTO,
  PeerInfoDTO,
  ProfileDTO,
  SchedulePolicyDTO,
} from '@/lib/api';
import { useToast } from '@/components/ui/use-toast';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { AgentBadge } from '@/components/AgentBadge';
import { SelectDropdown } from '@/components/ui/select-dropdown';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';

type ShareMode = 'off' | 'request' | 'auto';

interface NetworkViewProps {
  profiles?: ProfileDTO[];
}

export function NetworkView({ profiles = [] }: NetworkViewProps) {
  const { toast } = useToast();
  const [data, setData] = React.useState<ShareStatusDTO | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [filterQuery, setFilterQuery] = React.useState('');
  const [copiedCmd, setCopiedCmd] = React.useState<string | null>(null);
  const [activeTab, setActiveTab] = React.useState<string>('teammates');

  // Global Sharing Slider state
  const [isShareUpdating, setIsShareUpdating] = React.useState(false);

  // Connect Peer Modal state
  const [isConnectOpen, setIsConnectOpen] = React.useState(false);
  const [addHost, setAddHost] = React.useState('');
  const [isAdding, setIsAdding] = React.useState(false);

  // Request Lease Modal state
  const [selectedPeer, setSelectedPeer] = React.useState<PeerInfoDTO | null>(null);
  const [selectedProfile, setSelectedProfile] = React.useState<string>('');
  const [selectedAgent, setSelectedAgent] = React.useState<string>('claude');
  const [selectedIsAuto, setSelectedIsAuto] = React.useState<boolean>(false);
  const [leaseHours, setLeaseHours] = React.useState<number>(4);
  const [isRequesting, setIsRequesting] = React.useState(false);

  // Schedule & Automation state (Duration & Calendar Date-Time Picker)
  const [schedScope, setSchedScope] = React.useState<string>('global');
  const [schedMode, setSchedMode] = React.useState<'auto' | 'request'>('auto');
  const [schedType, setSchedType] = React.useState<'recurring' | 'calendar'>('recurring');
  const [schedStartTime, setSchedStartTime] = React.useState('00:00');
  const [schedEndTime, setSchedEndTime] = React.useState('08:00');
  const [schedStartDate, setSchedStartDate] = React.useState(() => new Date().toISOString().split('T')[0]);
  const [schedEndDate, setSchedEndDate] = React.useState(() => new Date().toISOString().split('T')[0]);
  const [isSavingSchedule, setIsSavingSchedule] = React.useState(false);
  const [currentTime, setCurrentTime] = React.useState(() => Date.now());

  // Requests Tab Filter state
  const [requestsFilter, setRequestsFilter] = React.useState<'all' | 'pending' | 'approved' | 'revoked'>('all');

  // Audit Log filter state
  const [auditSearchQuery, setAuditSearchQuery] = React.useState('');
  const [auditActionFilter, setAuditActionFilter] = React.useState<string>('all');

  React.useEffect(() => {
    const timer = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);

  const searchInputRef = React.useRef<HTMLInputElement>(null);

  const loadStatus = React.useCallback(async (refresh = false) => {
    try {
      const res = await getShareStatus(refresh);
      setData(res);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch share status';
      toast({
        title: 'Network status error',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  }, [toast]);

  React.useEffect(() => {
    loadStatus(true);
    let count = 0;
    const interval = setInterval(() => {
      count++;
      loadStatus(count % 3 === 0);
    }, 5000);
    return () => clearInterval(interval);
  }, [loadStatus]);

  const ownProfiles = React.useMemo(
    () => profiles.filter((p) => !p.name.includes(':')),
    [profiles]
  );

  const handleSetLocalProfileShareMode = async (profileName: string, mode: ShareMode) => {
    if (!profileName) return;
    setIsShareUpdating(true);
    const enabled = mode !== 'off';
    const auto_approve = mode === 'auto';
    const profObj = profiles.find((p) => p.name === profileName);
    const agent = profObj?.agent || 'claude';
    const existingConfig = data?.shared_profiles?.find((s) => s.name === profileName);
    const ttlStr = existingConfig?.max_lease_ttl
      ? `${Math.max(1, Math.round(existingConfig.max_lease_ttl / 1e9 / 3600))}h`
      : '4h';
    try {
      await toggleShareProfile({
        name: profileName,
        agent,
        auto_approve,
        ttl: ttlStr,
        enabled,
      });
      toast({
        title: mode === 'off' ? 'Profile Private' : `Profile Shared (${mode === 'auto' ? 'Auto-Approve' : 'Request Mode'})`,
        description: `Updated local sharing policy for "${profileName}"`,
        variant: mode === 'off' ? 'default' : 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not update profile sharing';
      toast({
        title: 'Update failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsShareUpdating(false);
    }
  };

  const handleCleanExpired = async () => {
    try {
      const res = await cleanExpiredLeases();
      toast({
        title: 'Expired Leases Cleaned',
        description: `Removed ${res.cleaned} expired local profile lease(s).`,
        variant: 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to clean expired leases';
      toast({
        title: 'Clean Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const handleManualRefresh = async () => {
    setIsLoading(true);
    try {
      await loadStatus(true);
      toast({
        title: 'Network Refreshed',
        description: 'Polled all discovered peers for latest profile statuses.',
        variant: 'success',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleApproveRequest = async (requestId: string) => {
    try {
      await approveShareRequest(requestId);
      toast({
        title: 'Lease Approved',
        description: `Granted credentials for request ${requestId}`,
        variant: 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to approve request';
      toast({
        title: 'Approval Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const handleDenyRequest = async (requestId: string) => {
    try {
      await denyShareRequest(requestId);
      toast({
        title: 'Request Denied',
        description: `Declined lease request ${requestId}`,
        variant: 'default',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to deny request';
      toast({
        title: 'Denial Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const handleRevokeLease = async (leaseId: string) => {
    try {
      await revokeShareLease(leaseId);
      toast({
        title: 'Lease Revoked',
        description: `Revoked credential lease ${leaseId}. Teammate sandbox access disabled.`,
        variant: 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to revoke lease';
      toast({
        title: 'Revoke failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  // Keyboard shortcut for search
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (document.querySelector('[role="dialog"]')) return;
      if (
        e.key === '/' &&
        document.activeElement?.tagName !== 'INPUT' &&
        document.activeElement?.tagName !== 'TEXTAREA'
      ) {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const isScheduleActiveNow = React.useCallback((s: SchedulePolicyDTO) => {
    if (!s.enabled) return false;
    const now = new Date();
    if (s.start_date && s.end_date) {
      const start = new Date(`${s.start_date}T${s.start_time}:00`);
      const end = new Date(`${s.end_date}T${s.end_time}:00`);
      return now >= start && now <= end;
    }
    if (s.start_time && s.end_time) {
      const [sh, sm] = s.start_time.split(':').map(Number);
      const [eh, em] = s.end_time.split(':').map(Number);
      const nowMin = now.getHours() * 60 + now.getMinutes();
      const startMin = sh * 60 + sm;
      const endMin = eh * 60 + em;
      if (startMin <= endMin) {
        return nowMin >= startMin && nowMin < endMin;
      }
      return nowMin >= startMin || nowMin < endMin;
    }
    return false;
  }, []);

  const activeSchedulesCount = React.useMemo(() => {
    return (data?.schedules || []).filter(isScheduleActiveNow).length;
  }, [data?.schedules, isScheduleActiveNow]);

  const currentGlobalMode: ShareMode = data?.global_mode || 'off';
  const effectiveGlobalMode: ShareMode = (data?.effective_global_mode || currentGlobalMode) as ShareMode;

  const handleSaveSchedulePolicy = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingSchedule(true);
    try {
      const payload: Partial<SchedulePolicyDTO> = {
        enabled: true,
        scope: schedScope,
        mode: schedMode,
        start_time: schedStartTime,
        end_time: schedEndTime,
      };
      if (schedType === 'calendar') {
        payload.start_date = schedStartDate;
        payload.end_date = schedEndDate;
      }
      await saveSchedule(payload);
      toast({
        title: 'Schedule Policy Saved',
        description: `Configured ${schedScope === 'global' ? 'Global' : schedScope} ${schedMode === 'auto' ? 'Auto-Approve' : 'Request'} (${schedStartTime} - ${schedEndTime})`,
        variant: 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to save schedule';
      toast({
        title: 'Save Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsSavingSchedule(false);
    }
  };

  const handleDeleteSchedulePolicy = async (id: string) => {
    try {
      await deleteSchedule(id);
      toast({
        title: 'Schedule Removed',
        description: 'Scheduled sharing policy removed',
        variant: 'success',
      });
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete schedule';
      toast({
        title: 'Delete Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const handleGlobalModeChange = async (mode: ShareMode) => {
    if (isShareUpdating) return;
    setIsShareUpdating(true);
    try {
      await setGlobalShareMode(mode);
      const titles: Record<ShareMode, string> = {
        off: 'Global Sharing Disabled',
        request: 'Global Sharing: Request Mode',
        auto: 'Global Sharing: Auto Approve',
      };
      const descs: Record<ShareMode, string> = {
        off: 'Profile sharing is disabled on this node.',
        request: 'Global sharing active. Teammate lease requests require approval.',
        auto: 'Global sharing active. Teammate lease requests are granted automatically.',
      };
      toast({
        title: titles[mode],
        description: descs[mode],
        variant: mode === 'off' ? 'default' : 'success',
      });
      await loadStatus();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not update global sharing mode';
      toast({
        title: 'Sharing update failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsShareUpdating(false);
    }
  };

  const handleAddPeer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!addHost.trim()) return;
    setIsAdding(true);
    try {
      const peer = await addSharePeer(addHost.trim());
      toast({
        title: 'Teammate Connected',
        description: `Verified ${peer.hostname} (${peer.profiles?.length || 0} shared profiles)`,
        variant: 'success',
      });
      setAddHost('');
      setIsConnectOpen(false);
      await loadStatus();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not reach teammate';
      toast({
        title: 'Connection Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsAdding(false);
    }
  };

  const handleRequestLease = async () => {
    if (!selectedPeer || !selectedProfile) return;
    setIsRequesting(true);
    try {
      const res = await requestShareLease({
        peer: selectedPeer.hostname,
        profile: selectedProfile,
        agent: selectedAgent,
        hours: leaseHours,
      });

      if (res.ok && res.local_profile) {
        const grantedAgent = res.agent || selectedAgent;
        toast({
          title: 'Lease Granted',
          description: `Installed as ${res.local_profile}. Run with: aim run ${grantedAgent} ${res.local_profile}`,
          variant: 'success',
        });
        setSelectedPeer(null);
        setSelectedProfile('');
        await loadStatus(false);
      } else {
        toast({
          title: 'Request Submitted',
          description: res.message || 'Waiting for teammate approval. Lease will activate once approved.',
        });
        setSelectedPeer(null);
        setSelectedProfile('');
        await loadStatus(false);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to obtain lease';
      toast({
        title: 'Lease Request Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsRequesting(false);
    }
  };

  const handleRelease = async (profileName: string) => {
    try {
      await releaseShareLease(profileName);
      toast({
        title: 'Lease Released',
        description: `Cleaned up local profile ${profileName}`,
        variant: 'success',
      });
      await loadStatus();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not release lease';
      toast({
        title: 'Release Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedCmd(id);
    toast({
      title: 'Command Copied',
      description: text,
      variant: 'success',
    });
    setTimeout(() => setCopiedCmd(null), 2000);
  };

  const formatExpiresIn = (expiresAtStr: string) => {
    const diff = new Date(expiresAtStr).getTime() - currentTime;
    if (diff <= 0) return 'Expired';
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    if (hours === 0 && mins === 0) return '< 1m remaining';
    if (hours === 0) return `${mins}m remaining`;
    return `${hours}h ${mins}m remaining`;
  };

  const formatLastSeen = (lastSeenStr?: string) => {
    if (!lastSeenStr) return 'unknown';
    const diff = currentTime - new Date(lastSeenStr).getTime();
    if (diff < 60 * 1000) return 'just now';
    const mins = Math.floor(diff / (60 * 1000));
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  };

  const pendingRequests = data?.pending_requests || [];

  const activeBorrowedLeases = React.useMemo(() => {
    return (data?.active_leases || []).filter(
      (l) => !l.is_expired && new Date(l.expires_at).getTime() > currentTime
    );
  }, [data?.active_leases, currentTime]);

  const expiredBorrowedLeases = React.useMemo(() => {
    return (data?.active_leases || []).filter(
      (l) => l.is_expired || new Date(l.expires_at).getTime() <= currentTime
    );
  }, [data?.active_leases, currentTime]);

  const revokedOrExpiredLeases = React.useMemo(() => {
    if (!data?.granted_leases) return [];
    return data.granted_leases.filter(
      (l) => l.revoked || new Date(l.expires_at).getTime() <= currentTime
    );
  }, [data?.granted_leases, currentTime]);

  const activeGrantedLeases = React.useMemo(() => {
    if (!data?.granted_leases) return [];
    return data.granted_leases.filter(
      (l) => !l.revoked && new Date(l.expires_at).getTime() > currentTime
    );
  }, [data?.granted_leases, currentTime]);

  const vpn = data?.vpn;
  const isVpnActive = vpn && vpn.connected;

  const filteredPeers = React.useMemo(() => {
    if (!data?.peers) return [];
    if (!filterQuery.trim()) return data.peers;
    const q = filterQuery.toLowerCase();
    return data.peers.filter(
      (p) =>
        p.hostname.toLowerCase().includes(q) ||
        p.vpn_ip.toLowerCase().includes(q) ||
        (p.profiles && p.profiles.some((prof) => prof.toLowerCase().includes(q)))
    );
  }, [data?.peers, filterQuery]);

  const filteredAuditLogs = React.useMemo(() => {
    if (!data?.audit_logs) return [];
    return data.audit_logs.filter((log) => {
      if (auditActionFilter !== 'all') {
        if (auditActionFilter === 'granted' && log.action !== 'LEASE_GRANTED') return false;
        if (auditActionFilter === 'requested' && log.action !== 'LEASE_REQUESTED') return false;
        if (auditActionFilter === 'revoked' && log.action !== 'LEASE_REVOKED') return false;
        if (
          auditActionFilter === 'policy' &&
          !['GLOBAL_MODE_CHANGED', 'PROFILE_SHARING_ON', 'PROFILE_SHARING_OFF'].includes(log.action)
        ) {
          return false;
        }
        if (
          auditActionFilter === 'schedule' &&
          !['SCHEDULE_SAVED', 'SCHEDULE_DELETED'].includes(log.action)
        ) {
          return false;
        }
      }

      if (auditSearchQuery.trim()) {
        const q = auditSearchQuery.toLowerCase();
        const matchesDetails = log.details?.toLowerCase().includes(q) ?? false;
        const matchesActor = log.actor?.toLowerCase().includes(q) ?? false;
        const matchesProfile = log.profile?.toLowerCase().includes(q) ?? false;
        const matchesAction = log.action?.toLowerCase().includes(q) ?? false;
        if (!matchesDetails && !matchesActor && !matchesProfile && !matchesAction) {
          return false;
        }
      }

      return true;
    });
  }, [data?.audit_logs, auditActionFilter, auditSearchQuery]);

  return (
    <div className="space-y-6">
      {/* Top Header bar with 3-Level Segmented Slider in the top right */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 pb-2 border-b border-[#222222]">
        <div className="space-y-1">
          <div className="flex items-center gap-2 flex-wrap">
            <Radio className="h-4 w-4 text-accent-cyan" />
            <h1 className="text-xl font-semibold tracking-tight text-foreground font-sans">
              P2P Network Sharing
            </h1>
            <Badge
              variant="outline"
              className="font-mono text-xs uppercase tracking-wider py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
            >
              Mesh
            </Badge>

            {/* Inline VPN & P2P status badge */}
            {isVpnActive ? (
              <span
                className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full border border-border bg-[#141414] text-xs font-mono text-muted-foreground"
                title={`Corporate VPN Active: Interface ${vpn.interface_name}, Subnet ${vpn.broadcast_ip}`}
              >
                <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                <span className="text-foreground font-medium">{vpn.interface_name}</span>
                <span className="text-muted-foreground/60">•</span>
                <span>{vpn.ip}:4319</span>
              </span>
            ) : (
              <span
                className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full border border-status-warning-border bg-status-warning-bg text-xs font-mono text-status-warning"
                title="Connect to Pritunl / OpenVPN to discover teammates across your corporate subnet"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                <span>VPN Offline</span>
              </span>
            )}
          </div>
        </div>

        {/* TOP RIGHT: Global Sharing 3-Level Segmented Control & Settings */}
        <div className="flex items-center gap-2 shrink-0">
          <div className="inline-flex items-center p-0.5 rounded-geist bg-[#141414] border border-[#262626] shadow-xs shrink-0">
            <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground px-2 flex items-center gap-1.5 shrink-0">
              <Radio className={cn('h-3 w-3', effectiveGlobalMode !== 'off' ? 'text-accent-cyan' : 'text-muted-foreground')} />
              <span>Global:</span>
            </span>

            <button
              type="button"
              disabled={isShareUpdating}
              onClick={() => handleGlobalModeChange('off')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium flex items-center gap-1.5 border',
                effectiveGlobalMode === 'off'
                  ? 'bg-[#262626] text-[#ededed] border-[#383838] shadow-xs'
                  : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
              )}
              title="Global sharing is OFF — machine is private and unreachable by teammates"
            >
              <span className={cn('h-1.5 w-1.5 rounded-full', effectiveGlobalMode === 'off' ? 'bg-[#999999]' : 'bg-[#555555]')} />
              <span>Off</span>
            </button>

            <button
              type="button"
              disabled={isShareUpdating}
              onClick={() => handleGlobalModeChange('request')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium flex items-center gap-1.5 border',
                effectiveGlobalMode === 'request'
                  ? 'bg-[#142334] text-[#52a8ff] border-[#1f4068] shadow-xs'
                  : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
              )}
              title="Global sharing is in Request mode — teammate lease requests require manual approval"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  effectiveGlobalMode === 'request' ? 'bg-[#52a8ff]' : 'bg-[#555555]'
                )}
              />
              <span>Request</span>
            </button>

            <button
              type="button"
              disabled={isShareUpdating}
              onClick={() => handleGlobalModeChange('auto')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium flex items-center gap-1.5 border',
                effectiveGlobalMode === 'auto'
                  ? 'bg-[#183925] text-[#98c379] border-[#235836] shadow-xs'
                  : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
              )}
              title="Global sharing is in Auto mode — teammate lease requests are automatically approved"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  effectiveGlobalMode === 'auto' ? 'bg-[#30a46c]' : 'bg-[#555555]'
                )}
              />
              <span>Auto Approve</span>
            </button>
          </div>
        </div>
      </div>

      {/* Tab Navigation System */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#222222] pb-3">
          <TabsList className="bg-[#0c0c0c] border border-[#222222] p-1 h-9 rounded-geist gap-1 w-full sm:w-auto">
            <TabsTrigger value="teammates" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
              <span>Teammates</span>
              {filteredPeers.length > 0 && (
                <span className="ml-0.5 px-1.5 py-0.2 rounded-full text-xs font-mono bg-[#1f1f1f] text-foreground border border-[#333333]">
                  {filteredPeers.length}
                </span>
              )}
            </TabsTrigger>

            <TabsTrigger value="requests" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Inbox className="h-3.5 w-3.5 text-[#e5c07b]" />
              <span>Requests</span>
              {pendingRequests.length > 0 ? (
                <span className="ml-0.5 px-1.5 py-0.2 rounded-full text-xs font-mono bg-[#142334] text-[#52a8ff] border border-[#1f4068] font-bold animate-pulse">
                  {pendingRequests.length}
                </span>
              ) : (activeGrantedLeases.length + activeBorrowedLeases.length) > 0 ? (
                <span className="ml-0.5 px-1.5 py-0.2 rounded-full text-xs font-mono bg-[#1f1f1f] text-foreground border border-[#333333]">
                  {activeGrantedLeases.length + activeBorrowedLeases.length}
                </span>
              ) : null}
            </TabsTrigger>

            <TabsTrigger value="my-sharing" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Share2 className="h-3.5 w-3.5 text-accent-cyan" />
              <span>My Sharing</span>
              <span className="ml-0.5 px-1.5 py-0.2 rounded-full text-xs font-mono bg-[#1f1f1f] text-foreground border border-[#333333]">
                {data?.shared_profiles?.filter((sp) => sp.enabled)?.length || 0}/{ownProfiles.length}
              </span>
            </TabsTrigger>

            <TabsTrigger value="audit" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Activity className="h-3.5 w-3.5 text-accent-cyan" />
              <span>Audit Log</span>
              {data?.audit_logs && data.audit_logs.length > 0 && (
                <span className="ml-0.5 px-1.5 py-0.2 rounded-full text-xs font-mono bg-[#1f1f1f] text-foreground border border-[#333333]">
                  {data.audit_logs.length}
                </span>
              )}
            </TabsTrigger>
          </TabsList>

          <div className="hidden md:flex items-center gap-3">
            {data?.active_leases && data.active_leases.length > 0 && (
              <span className="text-xs font-mono text-muted-foreground flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5 text-[#61afef]" />
                <span>
                  {data.active_leases.filter((l) => !l.is_expired && new Date(l.expires_at).getTime() > currentTime).length} borrowed
                </span>
              </span>
            )}
            {activeGrantedLeases.length > 0 && (
              <span className="text-xs font-mono text-muted-foreground flex items-center gap-1.5">
                <Share2 className="h-3.5 w-3.5 text-accent-cyan" />
                <span>{activeGrantedLeases.length} lending</span>
              </span>
            )}
          </div>
        </div>

        {/* TAB 1: TEAMMATES */}
        <TabsContent value="teammates" className="space-y-6 focus-visible:outline-none">
          {/* Active Borrowed Leases Section */}
          {data && data.active_leases && data.active_leases.length > 0 && (
            <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
              <CardHeader className="p-5 pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-[#61afef]" />
                    <CardTitle className="text-sm font-semibold text-foreground">
                      Borrowed Profiles
                    </CardTitle>
                    <Badge
                      variant="outline"
                      className="text-xs font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                    >
                      {data.active_leases.filter((l) => !l.is_expired && new Date(l.expires_at).getTime() > currentTime).length} active
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    {data.active_leases.some((l) => l.is_expired || new Date(l.expires_at).getTime() <= currentTime) && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={async () => {
                          const expired = data.active_leases.filter((l) => l.is_expired || new Date(l.expires_at).getTime() <= currentTime);
                          for (const l of expired) {
                            await releaseShareLease(l.local_profile).catch(() => {});
                          }
                          toast({ title: 'Cleaned up expired leases' });
                          await loadStatus();
                        }}
                        className="h-6 px-2 text-xs font-mono text-muted-foreground hover:text-status-danger cursor-pointer"
                      >
                        <Trash2 className="h-3 w-3 mr-1" />
                        Clean expired
                      </Button>
                    )}
                    <span className="text-xs text-muted-foreground font-mono">
                      Sandboxed locally
                    </span>
                  </div>
                </div>
                <CardDescription className="text-xs text-muted-foreground pt-1">
                  External teammate profiles currently leased and sandboxed locally on this machine.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {data.active_leases.map((lease) => {
                    const isExpired = lease.is_expired || new Date(lease.expires_at).getTime() <= currentTime;
                    const runCmd = `aim run ${lease.agent} ${lease.local_profile}`;
                    return (
                      <div
                        key={lease.local_profile}
                        className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3 shadow-xs hover:border-[#383838] transition-colors"
                      >
                        <div className="space-y-1.5">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span className="text-xs font-mono font-bold text-accent-cyan tracking-tight">
                                {lease.local_profile}
                              </span>
                              <AgentBadge agent={lease.agent} size="sm" />
                            </div>
                            <span
                              className={cn(
                                'text-xs font-mono px-2 py-0.5 rounded font-medium',
                                isExpired
                                  ? 'bg-status-danger-bg text-status-danger border border-status-danger-border'
                                  : 'bg-status-success-bg text-status-success border border-status-success-border'
                              )}
                            >
                              {formatExpiresIn(lease.expires_at)}
                            </span>
                          </div>
                          <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-mono">
                            <span>Source:</span>
                            <span className="text-foreground font-semibold">{lease.source_peer}</span>
                            <span className="text-[#666666]">/</span>
                            <span className="text-foreground">{lease.source_profile}</span>
                          </div>
                        </div>

                        <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                          <button
                            onClick={() => copyToClipboard(runCmd, lease.local_profile)}
                            className="inline-flex items-center gap-1.5 text-xs font-mono text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
                            title="Click to copy command"
                          >
                            {copiedCmd === lease.local_profile ? (
                              <Check className="h-3 w-3 text-status-success" />
                            ) : (
                              <Copy className="h-3 w-3" />
                            )}
                            <code>{runCmd}</code>
                          </button>

                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRelease(lease.local_profile)}
                            className="h-6 px-2 text-xs text-status-danger hover:bg-status-danger-bg cursor-pointer"
                          >
                            <Trash2 className="h-3 w-3 mr-1" />
                            Release
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Discovered Teammates & Shared Profiles Section */}
          <section className="space-y-4">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3 pb-3 border-b border-border/40">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Laptop className="h-4 w-4 text-[#61afef]" />
                  <h2 className="text-sm font-semibold text-foreground">
                    Discovered Teammates
                  </h2>
                  <Badge
                    variant="outline"
                    className="text-xs font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                  >
                    {filteredPeers.length} online
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground">
                  Teammates actively sharing AI profiles over the VPN subnet or local Wi-Fi.
                </p>
              </div>

              {/* Filter Input + Connect Peer + Refresh Button Group */}
              <div className="flex items-center gap-2 flex-wrap sm:flex-nowrap">
                <div className="relative w-full sm:w-64">
                  <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60" />
                  <Input
                    ref={searchInputRef}
                    type="text"
                    value={filterQuery}
                    onChange={(e) => setFilterQuery(e.target.value)}
                    placeholder="Filter teammates..."
                    className="pl-8 pr-8 h-8 text-xs font-sans bg-[#0c0c0c] border-[#222222] text-[#ededed] placeholder:text-muted-foreground rounded-geist focus:border-[#333333]"
                  />
                  {filterQuery ? (
                    <button
                      type="button"
                      onClick={() => setFilterQuery('')}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
                      title="Clear filter"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  ) : (
                    <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 border border-[#2a2a2a] bg-[#161616] px-1 py-[1px] rounded text-xs font-mono text-muted-foreground pointer-events-none">
                      /
                    </kbd>
                  )}
                </div>

                {/* Connect Peer Icon Button */}
                <Button
                  variant="default"
                  size="icon"
                  onClick={() => setIsConnectOpen(true)}
                  className="h-8 w-8 rounded-geist bg-[#ededed] text-[#0a0a0a] hover:bg-white hover:text-black flex items-center justify-center shrink-0 cursor-pointer shadow-xs"
                  title="Connect Peer"
                >
                  <Plus className="h-4 w-4" />
                </Button>

                {/* Refresh Icon Button */}
                <Button
                  variant="outline"
                  size="sm"
                  disabled={isLoading}
                  onClick={handleManualRefresh}
                  className="h-8 w-8 p-0 rounded-geist border-border bg-[#141414] hover:bg-[#1f1f1f] text-foreground flex items-center justify-center shrink-0 cursor-pointer"
                  title="Refresh Teammates"
                >
                  <RotateCw className={cn('h-3.5 w-3.5', isLoading && 'animate-spin text-accent-cyan')} />
                </Button>
              </div>
            </div>

            <div>
              {!data || !data.peers || data.peers.length === 0 ? (
                <div className="p-10 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] space-y-2.5">
                  <Radio className="h-7 w-7 text-muted-foreground/30 mx-auto" />
                  <div className="text-xs font-semibold text-foreground">No teammates discovered yet</div>
                  <p className="text-xs text-muted-foreground max-w-sm mx-auto leading-relaxed">
                    Teammates connected to the internal VPN who share a profile will be automatically discovered via mDNS, or you can click &quot;Connect Peer&quot; above to add their IP/hostname directly.
                  </p>
                </div>
              ) : filteredPeers.length === 0 ? (
                <div className="p-6 text-center border border-dashed border-border rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground">
                  No teammates match filter &ldquo;{filterQuery}&rdquo;.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filteredPeers.map((peer) => {
                    const profileCount = peer.profiles?.length || 0;
                    const isPeerOnline = peer.last_seen
                      ? (currentTime - new Date(peer.last_seen).getTime()) < 5 * 60 * 1000
                      : true;
                    return (
                      <Card
                        key={peer.id || peer.hostname}
                        className="border border-border bg-card hover:border-[#383838] transition-all rounded-geist flex flex-col justify-between"
                      >
                        <div>
                          <CardHeader className="p-4 pb-2">
                            <div className="flex items-start justify-between gap-2">
                              <div className="space-y-0.5 min-w-0">
                                <CardTitle className="font-semibold text-sm text-foreground tracking-tight truncate flex items-center gap-1.5">
                                  <Laptop className="h-4 w-4 text-[#61afef] shrink-0" />
                                  <span className="truncate">{peer.hostname}</span>
                                </CardTitle>
                                <CardDescription className="text-xs font-mono text-muted-foreground truncate">
                                  {peer.vpn_ip}
                                </CardDescription>
                              </div>
                              {isPeerOnline ? (
                                <Badge
                                  variant="outline"
                                  className="badge-success text-xs font-sans px-2 py-0.5 shrink-0 flex items-center gap-1.5"
                                >
                                  <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                                  <span>Online</span>
                                </Badge>
                              ) : (
                                <Badge
                                  variant="outline"
                                  className="text-xs font-sans px-2 py-0.5 shrink-0 flex items-center gap-1.5 border-border/60 bg-[#161616] text-muted-foreground"
                                  title={`Last seen: ${peer.last_seen}`}
                                >
                                  <span className="h-1.5 w-1.5 rounded-full bg-[#555555]" />
                                  <span>{formatLastSeen(peer.last_seen)}</span>
                                </Badge>
                              )}
                            </div>
                          </CardHeader>

                          <CardContent className="p-4 pt-2 space-y-2.5">
                            <div className="flex items-center justify-between text-xs font-mono uppercase text-muted-foreground tracking-wider font-medium">
                              <span>Offered Profiles</span>
                              <span>{profileCount} available</span>
                            </div>

                            {peer.profiles && peer.profiles.length > 0 ? (
                              <div className="space-y-1.5">
                                {peer.profiles.map((prof) => {
                                  const sp = peer.shared_profiles?.find((p) => p.name === prof);
                                  const agent = sp?.agent || 'claude';
                                  const isAuto = sp ? sp.auto_approve : true;
                                  return (
                                    <div
                                      key={prof}
                                      className="group flex items-center justify-between gap-2 px-2.5 py-2 rounded-geist border border-[#222222] bg-[#121212]/80 hover:bg-[#181818] hover:border-[#383838] transition-all"
                                    >
                                      <div className="flex items-center gap-2 truncate min-w-0">
                                        <span className="font-mono text-xs font-semibold text-foreground group-hover:text-[#61afef] transition-colors truncate">
                                          {prof}
                                        </span>
                                        <div className="opacity-0 group-hover:opacity-100 group-focus-within:opacity-100 transition-opacity duration-150 shrink-0">
                                          <AgentBadge agent={agent} size="sm" />
                                        </div>
                                        {sp && (
                                          <span
                                            className={cn(
                                              'text-[10px] font-mono uppercase tracking-wider px-1.5 py-0.5 rounded border shrink-0 font-medium flex items-center gap-1',
                                              isAuto
                                                ? 'bg-[#183925]/60 text-[#98c379] border-[#235836]/70'
                                                : 'bg-[#142334]/60 text-[#52a8ff] border-[#1f4068]/70'
                                            )}
                                          >
                                            <span className={cn('h-1 w-1 rounded-full', isAuto ? 'bg-[#98c379]' : 'bg-[#52a8ff]')} />
                                            <span>{isAuto ? 'Auto' : 'Request'}</span>
                                          </span>
                                        )}
                                      </div>
                                      <Button
                                        size="sm"
                                        variant="outline"
                                        onClick={() => {
                                          setSelectedPeer(peer);
                                          setSelectedProfile(prof);
                                          setSelectedAgent(agent);
                                          setSelectedIsAuto(isAuto);
                                        }}
                                        className="h-6 px-2 text-[11px] font-mono font-medium rounded-geist border-[#2a2a2a] bg-[#1a1a1a] hover:bg-[#252525] hover:border-[#404040] text-foreground hover:text-[#61afef] flex items-center gap-1.5 cursor-pointer shrink-0 shadow-xs transition-colors"
                                        title={`Borrow ${prof} from ${peer.hostname}`}
                                      >
                                        <KeyRound className="h-3 w-3 text-muted-foreground group-hover:text-[#61afef] transition-colors" />
                                        <span>Borrow</span>
                                      </Button>
                                    </div>
                                  );
                                })}
                              </div>
                            ) : (
                              <div className="p-3 rounded-geist bg-[#141414] border border-[#1e1e1e] text-xs text-muted-foreground text-center italic">
                                No profiles currently shared
                              </div>
                            )}
                          </CardContent>
                        </div>
                      </Card>
                    );
                  })}
                </div>
              )}
            </div>
          </section>
        </TabsContent>

        {/* TAB 2: REQUESTS (Pending, Approved, Revoked / Expired) */}
        <TabsContent value="requests" className="space-y-5 focus-visible:outline-none">
          {/* Sub-filter row */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-2 border-b border-border/40">
            <div className="flex items-center gap-1.5 flex-wrap">
              <button
                type="button"
                onClick={() => setRequestsFilter('all')}
                className={cn(
                  'px-3 py-1 rounded-geist text-xs font-mono transition-colors cursor-pointer border',
                  requestsFilter === 'all'
                    ? 'bg-[#222222] text-[#ededed] border-[#383838] font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-[#141414] border-transparent'
                )}
              >
                All Requests ({pendingRequests.length + activeGrantedLeases.length + activeBorrowedLeases.length + revokedOrExpiredLeases.length + expiredBorrowedLeases.length})
              </button>

              <button
                type="button"
                onClick={() => setRequestsFilter('pending')}
                className={cn(
                  'px-3 py-1 rounded-geist text-xs font-mono transition-colors cursor-pointer border flex items-center gap-1.5',
                  requestsFilter === 'pending'
                    ? 'bg-[#142334] text-[#52a8ff] border-[#1f4068] font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-[#141414] border-transparent'
                )}
              >
                <span className={cn('h-1.5 w-1.5 rounded-full', pendingRequests.length > 0 ? 'bg-[#52a8ff]' : 'bg-[#555555]')} />
                <span>Pending ({pendingRequests.length})</span>
              </button>

              <button
                type="button"
                onClick={() => setRequestsFilter('approved')}
                className={cn(
                  'px-3 py-1 rounded-geist text-xs font-mono transition-colors cursor-pointer border flex items-center gap-1.5',
                  requestsFilter === 'approved'
                    ? 'bg-[#183925] text-[#98c379] border-[#235836] font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-[#141414] border-transparent'
                )}
              >
                <span className="h-1.5 w-1.5 rounded-full bg-[#30a46c]" />
                <span>Approved ({activeGrantedLeases.length + activeBorrowedLeases.length})</span>
              </button>

              <button
                type="button"
                onClick={() => setRequestsFilter('revoked')}
                className={cn(
                  'px-3 py-1 rounded-geist text-xs font-mono transition-colors cursor-pointer border flex items-center gap-1.5',
                  requestsFilter === 'revoked'
                    ? 'bg-[#331c1e] text-[#e06c75] border-[#552528] font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-[#141414] border-transparent'
                )}
              >
                <span className="h-1.5 w-1.5 rounded-full bg-[#e06c75]" />
                <span>Revoked / Expired ({revokedOrExpiredLeases.length + expiredBorrowedLeases.length})</span>
              </button>
            </div>
          </div>

          {/* Pending Requests Section */}
          {(requestsFilter === 'all' || requestsFilter === 'pending') && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-[#e5c07b]" />
                <h3 className="text-xs font-mono uppercase tracking-wider font-semibold text-foreground">
                  Incoming Requests Pending Approval ({pendingRequests.length})
                </h3>
              </div>

              {pendingRequests.length === 0 ? (
                requestsFilter === 'pending' && (
                  <div className="p-8 text-center border border-dashed border-[#222222] rounded-geist bg-[#0d0d0d] text-xs font-mono text-muted-foreground">
                    No pending approval requests. Teammate requests will appear here when sharing in Request mode.
                  </div>
                )
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {pendingRequests.map((pr) => (
                    <div
                      key={pr.request_id}
                      className="p-4 rounded-geist border border-[#2a2a2a] bg-[#111111] flex flex-col justify-between gap-3 shadow-sm hover:border-[#3a3a3a] transition-colors"
                    >
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <AgentBadge agent={pr.request.agent} size="sm" />
                            <span className="font-mono font-bold text-xs text-foreground">
                              {pr.request.profile_name}
                            </span>
                          </div>
                          <Badge variant="outline" className="text-xs font-mono border-status-warning-border bg-status-warning-bg text-status-warning">
                            Pending Approval
                          </Badge>
                        </div>

                        <div className="space-y-1 text-xs font-mono text-muted-foreground">
                          <div className="flex items-center gap-1.5">
                            <Laptop className="h-3 w-3 text-[#61afef]" />
                            <span className="text-foreground font-semibold">{pr.request.requester_id}</span>
                            <span className="text-[#666666]">({pr.request.requester_ip})</span>
                          </div>
                          <div className="flex items-center justify-between text-[#888888] pt-1">
                            <span>Requested Duration:</span>
                            <span className="text-foreground font-semibold">{pr.request.requested_hours} hours</span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#1f1f1f]">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDenyRequest(pr.request_id)}
                          className="h-7 px-2.5 text-xs font-mono text-status-danger hover:bg-status-danger-bg cursor-pointer"
                        >
                          <Ban className="h-3.5 w-3.5 mr-1" />
                          Deny
                        </Button>
                        <Button
                          size="sm"
                          onClick={() => handleApproveRequest(pr.request_id)}
                          className="h-7 px-3 text-xs font-mono bg-[#183925] hover:bg-[#204a31] text-[#98c379] border border-[#235836] font-medium cursor-pointer shadow-xs"
                        >
                          <FileCheck className="h-3.5 w-3.5 mr-1" />
                          Approve
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Approved Requests / Active Leases Section */}
          {(requestsFilter === 'all' || requestsFilter === 'approved') && (
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <FileCheck className="h-4 w-4 text-[#98c379]" />
                <h3 className="text-xs font-mono uppercase tracking-wider font-semibold text-foreground">
                  Active Approved Leases ({activeGrantedLeases.length + activeBorrowedLeases.length})
                </h3>
              </div>

              {activeGrantedLeases.length === 0 && activeBorrowedLeases.length === 0 ? (
                requestsFilter === 'approved' && (
                  <div className="p-8 text-center border border-dashed border-[#222222] rounded-geist bg-[#0d0d0d] text-xs font-mono text-muted-foreground">
                    No active approved leases.
                  </div>
                )
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {/* Granted to teammates */}
                  {activeGrantedLeases.map((lease) => (
                    <div
                      key={lease.lease_id}
                      className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3 shadow-xs hover:border-[#383838] transition-colors"
                    >
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-mono font-bold text-foreground">
                              {lease.profile_name}
                            </span>
                            <AgentBadge agent={lease.agent} size="sm" />
                            <span
                              className={cn(
                                'text-xs font-mono px-1.5 py-0.5 rounded border',
                                lease.auto_approved
                                  ? 'bg-[#183925]/50 text-[#98c379] border-[#235836]/60'
                                  : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                              )}
                            >
                              {lease.auto_approved ? 'Auto Approved' : 'Manual Approved'}
                            </span>
                          </div>
                          <span className="text-xs font-mono px-2 py-0.5 rounded font-medium bg-status-success-bg text-status-success border border-status-success-border">
                            {formatExpiresIn(lease.expires_at)}
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-mono">
                          <Laptop className="h-3 w-3 text-[#61afef]" />
                          <span>Borrower:</span>
                          <span className="text-foreground font-semibold truncate max-w-[220px]" title={lease.borrower_id}>
                            {lease.borrower_id}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                        <span className="text-xs font-mono text-[#666666]">
                          Lease ID: {lease.lease_id}
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRevokeLease(lease.lease_id)}
                          className="h-6 px-2 text-xs text-status-danger hover:bg-status-danger-bg cursor-pointer"
                        >
                          <ShieldAlert className="h-3 w-3 mr-1" />
                          Revoke Lease
                        </Button>
                      </div>
                    </div>
                  ))}

                  {/* Borrowed by you */}
                  {activeBorrowedLeases.map((lease) => {
                    const runCmd = `aim run ${lease.agent} ${lease.local_profile}`;
                    return (
                      <div
                        key={lease.local_profile}
                        className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3 shadow-xs hover:border-[#383838] transition-colors"
                      >
                        <div className="space-y-1.5">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span className="text-xs font-mono font-bold text-accent-cyan tracking-tight">
                                {lease.local_profile}
                              </span>
                              <AgentBadge agent={lease.agent} size="sm" />
                            </div>
                            <span className="text-xs font-mono px-2 py-0.5 rounded font-medium bg-status-success-bg text-status-success border border-status-success-border">
                              {formatExpiresIn(lease.expires_at)}
                            </span>
                          </div>
                          <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-mono">
                            <span>Host:</span>
                            <span className="text-foreground font-semibold">{lease.source_peer}</span>
                            <span className="text-[#666666]">/</span>
                            <span className="text-foreground">{lease.source_profile}</span>
                          </div>
                        </div>

                        <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                          <button
                            onClick={() => copyToClipboard(runCmd, lease.local_profile)}
                            className="inline-flex items-center gap-1.5 text-xs font-mono text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
                            title="Click to copy command"
                          >
                            {copiedCmd === lease.local_profile ? (
                              <Check className="h-3 w-3 text-status-success" />
                            ) : (
                              <Copy className="h-3 w-3" />
                            )}
                            <code>{runCmd}</code>
                          </button>

                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRelease(lease.local_profile)}
                            className="h-6 px-2 text-xs text-status-danger hover:bg-status-danger-bg cursor-pointer"
                          >
                            <Trash2 className="h-3 w-3 mr-1" />
                            Release
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* Revoked / Expired Leases Section */}
          {(requestsFilter === 'all' || requestsFilter === 'revoked') && (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Ban className="h-4 w-4 text-muted-foreground" />
                  <h3 className="text-xs font-mono uppercase tracking-wider font-semibold text-foreground">
                    Revoked & Expired Leases ({revokedOrExpiredLeases.length + expiredBorrowedLeases.length})
                  </h3>
                </div>
                {expiredBorrowedLeases.length > 0 && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleCleanExpired}
                    className="h-6 px-2.5 text-xs font-mono border-border bg-[#141414] hover:bg-[#1a1a1a] text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-1.5"
                  >
                    <Trash2 className="h-3 w-3 text-status-danger" />
                    <span>Clean Expired ({expiredBorrowedLeases.length})</span>
                  </Button>
                )}
              </div>

              {revokedOrExpiredLeases.length === 0 && expiredBorrowedLeases.length === 0 ? (
                requestsFilter === 'revoked' && (
                  <div className="p-8 text-center border border-dashed border-[#222222] rounded-geist bg-[#0d0d0d] text-xs font-mono text-muted-foreground">
                    No revoked or expired leases.
                  </div>
                )
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {/* Expired Leases Borrowed from Peers */}
                  {expiredBorrowedLeases.map((lease) => (
                    <div
                      key={lease.local_profile}
                      className="p-3.5 rounded-geist border border-[#1f1f1f] bg-[#0c0c0c] flex items-center justify-between gap-3 text-xs font-mono opacity-80"
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-foreground truncate">{lease.local_profile}</span>
                            <AgentBadge agent={lease.agent} size="sm" />
                          </div>
                          <div className="text-muted-foreground text-xs truncate">from: {lease.source_peer}/{lease.source_profile}</div>
                        </div>
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <Badge variant="outline" className="text-xs font-mono border-status-danger-border text-status-danger bg-status-danger-bg">
                          Expired
                        </Badge>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRelease(lease.local_profile)}
                          className="h-6 px-2 text-xs text-muted-foreground hover:text-foreground cursor-pointer"
                          title="Purge expired lease credentials"
                        >
                          <Trash2 className="h-3 w-3 mr-1" />
                          Purge
                        </Button>
                      </div>
                    </div>
                  ))}

                  {/* Leases Granted to Teammates that are Revoked / Expired */}
                  {revokedOrExpiredLeases.map((lease) => (
                    <div
                      key={lease.lease_id}
                      className="p-3.5 rounded-geist border border-[#1f1f1f] bg-[#0c0c0c] flex items-center justify-between gap-3 text-xs font-mono opacity-80"
                    >
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-semibold text-foreground truncate">{lease.profile_name}</span>
                            <AgentBadge agent={lease.agent} size="sm" />
                          </div>
                          <div className="text-muted-foreground text-xs truncate">borrower: {lease.borrower_id}</div>
                        </div>
                      </div>
                      <Badge variant="outline" className={cn("text-xs font-mono", lease.revoked ? "border-status-danger-border text-status-danger bg-status-danger-bg" : "border-border text-muted-foreground")}>
                        {lease.revoked ? 'Revoked' : 'Expired'}
                      </Badge>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </TabsContent>

        {/* TAB 2: MY SHARING */}
        <TabsContent value="my-sharing" className="space-y-6 focus-visible:outline-none">
          {effectiveGlobalMode === 'off' && (
            <div className="p-4 rounded-geist border border-[#262626] bg-[#0f0f0f] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <ShieldAlert className="h-4 w-4 text-[#e5c07b]" />
                  <span className="text-xs font-semibold text-foreground">Global Sharing is Currently Off</span>
                </div>
                <p className="text-xs text-muted-foreground max-w-xl">
                  Your node is completely private. Teammates on your subnet cannot discover your profiles or request leases until sharing is enabled.
                </p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleGlobalModeChange('request')}
                  className="rounded-geist text-xs font-mono border-border bg-[#141414] hover:bg-[#1a1a1a] cursor-pointer"
                >
                  Enable Request Mode
                </Button>
                <Button
                  size="sm"
                  onClick={() => handleGlobalModeChange('auto')}
                  className="rounded-geist text-xs font-mono bg-[#183925] hover:bg-[#204a31] text-[#98c379] border border-[#235836] cursor-pointer"
                >
                  Enable Auto-Approve
                </Button>
              </div>
            </div>
          )}

          {/* Active Granted Leases Section (Teammates Borrowing Your Profiles) */}
          {activeGrantedLeases.length > 0 && (
            <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
              <CardHeader className="p-5 pb-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <Share2 className="h-4 w-4 text-accent-cyan" />
                    <CardTitle className="text-sm font-semibold text-foreground">
                      Active Borrowers
                    </CardTitle>
                    <Badge
                      variant="outline"
                      className="text-xs font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                    >
                      {activeGrantedLeases.length} active
                    </Badge>
                  </div>
                  <span className="text-xs text-muted-foreground font-mono">
                    Revoke access instantly anytime
                  </span>
                </div>
                <CardDescription className="text-xs text-muted-foreground pt-1">
                  Teammates currently authenticated with your shared profile credentials.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {activeGrantedLeases.map((lease) => (
                    <div
                      key={lease.lease_id}
                      className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3 shadow-xs hover:border-[#383838] transition-colors"
                    >
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <AgentBadge agent={lease.agent} size="sm" />
                            <span className="text-xs font-mono font-bold text-foreground">
                              {lease.profile_name}
                            </span>
                            <span
                              className={cn(
                                'text-xs font-mono px-1.5 py-0.5 rounded border',
                                lease.auto_approved
                                  ? 'bg-[#183925]/50 text-[#98c379] border-[#235836]/60'
                                  : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                              )}
                            >
                              {lease.auto_approved ? 'Auto Approved' : 'Manual Approved'}
                            </span>
                          </div>
                          <span className="text-xs font-mono px-2 py-0.5 rounded font-medium bg-status-success-bg text-status-success border border-status-success-border">
                            {formatExpiresIn(lease.expires_at)}
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-mono">
                          <Laptop className="h-3 w-3 text-[#61afef]" />
                          <span>Borrower:</span>
                          <span className="text-foreground font-semibold truncate max-w-[220px]" title={lease.borrower_id}>
                            {lease.borrower_id}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                        <span className="text-xs font-mono text-[#666666]">
                          Lease ID: {lease.lease_id}
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRevokeLease(lease.lease_id)}
                          className="h-6 px-2 text-xs text-status-danger hover:bg-status-danger-bg cursor-pointer"
                        >
                          <ShieldAlert className="h-3 w-3 mr-1" />
                          Revoke Lease
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Local Offered Profiles Section */}
          <section className="space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-2 border-b border-border/40">
              <div className="space-y-0.5">
                <div className="flex items-center gap-2">
                  <Share2 className="h-4 w-4 text-accent-cyan" />
                  <h2 className="text-sm font-semibold text-foreground">
                    Your Offered Profiles
                  </h2>
                  <Badge
                    variant="outline"
                    className="text-xs font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                  >
                    {data?.shared_profiles?.filter((sp) => sp.enabled)?.length || 0} of {ownProfiles.length} shared
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground">
                  Profiles on this machine that teammates on the VPN mesh can discover and lease.
                </p>
              </div>
            </div>

            {ownProfiles.length === 0 ? (
              <div className="p-8 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground font-mono">
                No local profiles found on this machine to share.
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {ownProfiles.map((p) => {
                  const sp = data?.shared_profiles?.find((s) => s.name === p.name);
                  const isShared = sp ? sp.enabled : false;
                  const mode: ShareMode = !isShared ? 'off' : (sp?.auto_approve ? 'auto' : 'request');
                  return (
                    <div
                      key={p.name}
                      className="p-3.5 rounded-geist border border-border bg-card flex items-center justify-between gap-3 shadow-xs hover:border-[#383838] transition-colors"
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <div className="font-semibold text-xs text-foreground truncate font-mono">
                          {p.name}
                        </div>
                        <AgentBadge agent={p.agent} size="sm" />
                      </div>

                      {/* Segmented slider for this profile */}
                      <div className="inline-flex items-center p-0.5 rounded-geist bg-[#141414] border border-[#262626] shrink-0">
                        <button
                          type="button"
                          disabled={isShareUpdating}
                          onClick={() => handleSetLocalProfileShareMode(p.name, 'off')}
                          className={cn(
                            'px-2 py-0.5 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium border flex items-center gap-1',
                            mode === 'off'
                              ? 'bg-[#262626] text-[#ededed] border-[#383838] shadow-xs'
                              : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
                          )}
                          title="Keep this profile private"
                        >
                          <span className={cn('h-1.5 w-1.5 rounded-full', mode === 'off' ? 'bg-[#999999]' : 'bg-[#555555]')} />
                          <span>Off</span>
                        </button>
                        <button
                          type="button"
                          disabled={isShareUpdating}
                          onClick={() => handleSetLocalProfileShareMode(p.name, 'request')}
                          className={cn(
                            'px-2 py-0.5 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium flex items-center gap-1 border',
                            mode === 'request'
                              ? 'bg-[#142334] text-[#52a8ff] border-[#1f4068] shadow-xs'
                              : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
                          )}
                          title="Require manual lease request confirmation"
                        >
                          <span className={cn('h-1.5 w-1.5 rounded-full', mode === 'request' ? 'bg-[#52a8ff]' : 'bg-[#555555]')} />
                          <span>Req</span>
                        </button>
                        <button
                          type="button"
                          disabled={isShareUpdating}
                          onClick={() => handleSetLocalProfileShareMode(p.name, 'auto')}
                          className={cn(
                            'px-2 py-0.5 rounded-sm text-xs font-mono transition-colors cursor-pointer font-medium flex items-center gap-1 border',
                            mode === 'auto'
                              ? 'bg-[#183925] text-[#98c379] border-[#235836] shadow-xs'
                              : 'text-muted-foreground hover:text-foreground hover:bg-[#1a1a1a] border-transparent'
                          )}
                          title="Auto-approve lease requests for this profile"
                        >
                          <span className={cn('h-1.5 w-1.5 rounded-full', mode === 'auto' ? 'bg-[#30a46c]' : 'bg-[#555555]')} />
                          <span>Auto</span>
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          {/* Scheduled Sharing & Automation Card (Housed directly in My Sharing) */}
          <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
            <CardHeader className="p-5 pb-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-accent-cyan" />
                  <CardTitle className="text-sm font-semibold text-foreground">
                    Automation & Scheduled Windows
                  </CardTitle>
                  {activeSchedulesCount > 0 && (
                    <Badge
                      variant="outline"
                      className="text-xs font-mono py-0 px-2 rounded-full border-border bg-[#183925] text-[#98c379]"
                    >
                      {activeSchedulesCount} active now
                    </Badge>
                  )}
                </div>
                <span className="text-xs text-muted-foreground font-mono">
                  Time-based sharing policies
                </span>
              </div>
              <CardDescription className="text-xs text-muted-foreground pt-1">
                Automatically switch sharing modes or auto-approve teammate requests during off-peak hours (e.g. nights or weekends).
              </CardDescription>
            </CardHeader>
            <CardContent className="p-5 pt-0 space-y-4">
              <form onSubmit={handleSaveSchedulePolicy} className="space-y-4 p-4 rounded-geist bg-[#0f0f0f] border border-[#1f1f1f]">
                {/* Scope & Policy Mode */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground font-medium">
                      Target Scope
                    </label>
                    <SelectDropdown
                      value={schedScope}
                      onChange={(val) => setSchedScope(String(val))}
                      size="sm"
                      options={[
                        { value: 'global', label: 'Global (All Profiles)' },
                        ...ownProfiles.map((p) => ({ value: p.name, label: `Profile: ${p.name}` })),
                      ]}
                    />
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground font-medium">
                      Policy Mode in Window
                    </label>
                    <SelectDropdown
                      value={schedMode}
                      onChange={(val) => setSchedMode(val as 'auto' | 'request')}
                      size="sm"
                      options={[
                        { value: 'auto', label: 'Auto Approve (Instant)' },
                        { value: 'request', label: 'Request Mode (Manual)' },
                      ]}
                    />
                  </div>
                </div>

                {/* Schedule Type Selection */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono uppercase tracking-wider text-muted-foreground font-medium">
                    Schedule Window Type
                  </label>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => setSchedType('recurring')}
                      className={cn(
                        'px-3 py-2 text-xs font-mono rounded-geist border transition-all text-left flex items-center gap-2 cursor-pointer',
                        schedType === 'recurring'
                          ? 'bg-[#181818] border-accent-cyan text-foreground font-semibold'
                          : 'bg-[#121212] border-[#262626] text-muted-foreground hover:bg-[#161616]'
                      )}
                    >
                      <Clock className="h-3.5 w-3.5 text-accent-cyan" />
                      <span>Daily Recurring Hours</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setSchedType('calendar')}
                      className={cn(
                        'px-3 py-2 text-xs font-mono rounded-geist border transition-all text-left flex items-center gap-2 cursor-pointer',
                        schedType === 'calendar'
                          ? 'bg-[#181818] border-accent-cyan text-foreground font-semibold'
                          : 'bg-[#121212] border-[#262626] text-muted-foreground hover:bg-[#161616]'
                      )}
                    >
                      <Calendar className="h-3.5 w-3.5 text-accent-cyan" />
                      <span>Calendar Date Range</span>
                    </button>
                  </div>
                </div>

                {/* Recurring Time Pickers with Quick Chips */}
                {schedType === 'recurring' ? (
                  <div className="space-y-2.5 p-3 rounded-geist bg-[#141414] border border-[#222222]">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs font-mono text-muted-foreground">
                      <span>Recurring Daily Hours</span>
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <button
                          type="button"
                          onClick={() => { setSchedStartTime('00:00'); setSchedEndTime('08:00'); }}
                          className="px-2 py-0.5 rounded text-xs font-mono bg-[#1f1f1f] hover:bg-[#282828] text-accent-cyan transition-colors cursor-pointer"
                        >
                          12am - 8am
                        </button>
                        <button
                          type="button"
                          onClick={() => { setSchedStartTime('22:00'); setSchedEndTime('06:00'); }}
                          className="px-2 py-0.5 rounded text-xs font-mono bg-[#1f1f1f] hover:bg-[#282828] text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                        >
                          10pm - 6am
                        </button>
                        <button
                          type="button"
                          onClick={() => { setSchedStartTime('09:00'); setSchedEndTime('18:00'); }}
                          className="px-2 py-0.5 rounded text-xs font-mono bg-[#1f1f1f] hover:bg-[#282828] text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
                        >
                          9am - 6pm
                        </button>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div className="space-y-1">
                        <span className="text-xs font-mono text-muted-foreground">From Time</span>
                        <input
                          type="time"
                          value={schedStartTime}
                          onChange={(e) => setSchedStartTime(e.target.value)}
                          required
                          className="w-full h-8 px-2.5 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                        />
                      </div>
                      <div className="space-y-1">
                        <span className="text-xs font-mono text-muted-foreground">To Time</span>
                        <input
                          type="time"
                          value={schedEndTime}
                          onChange={(e) => setSchedEndTime(e.target.value)}
                          required
                          className="w-full h-8 px-2.5 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                        />
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-2.5 p-3 rounded-geist bg-[#141414] border border-[#222222]">
                    <span className="text-xs font-mono text-muted-foreground block">
                      Calendar Date & Time Interval
                    </span>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <div className="space-y-1">
                        <span className="text-xs font-mono text-muted-foreground">Start Date & Time</span>
                        <div className="flex gap-1.5">
                          <input
                            type="date"
                            value={schedStartDate}
                            onChange={(e) => setSchedStartDate(e.target.value)}
                            required
                            className="w-2/3 h-8 px-2 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                          />
                          <input
                            type="time"
                            value={schedStartTime}
                            onChange={(e) => setSchedStartTime(e.target.value)}
                            required
                            className="w-1/3 h-8 px-1.5 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                          />
                        </div>
                      </div>
                      <div className="space-y-1">
                        <span className="text-xs font-mono text-muted-foreground">End Date & Time</span>
                        <div className="flex gap-1.5">
                          <input
                            type="date"
                            value={schedEndDate}
                            onChange={(e) => setSchedEndDate(e.target.value)}
                            required
                            className="w-2/3 h-8 px-2 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                          />
                          <input
                            type="time"
                            value={schedEndTime}
                            onChange={(e) => setSchedEndTime(e.target.value)}
                            required
                            className="w-1/3 h-8 px-1.5 rounded-geist bg-[#0c0c0c] border border-[#262626] text-xs font-mono text-[#ededed] focus:outline-none focus:border-accent-cyan"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                <div className="flex justify-end pt-1">
                  <Button
                    type="submit"
                    size="sm"
                    disabled={isSavingSchedule}
                    className="h-8 rounded-geist bg-[#0070f3] hover:bg-[#0060df] text-xs font-mono text-white flex items-center gap-1.5 cursor-pointer shadow-xs"
                  >
                    <Check className="h-3.5 w-3.5" />
                    <span>{isSavingSchedule ? 'Saving...' : 'Add Scheduled Policy'}</span>
                  </Button>
                </div>
              </form>

              {/* Existing Scheduled Policies */}
              {data?.schedules && data.schedules.length > 0 ? (
                <div className="space-y-2 pt-2 border-t border-[#222222]">
                  <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground font-medium block">
                    Active & Configured Policies ({data.schedules.length})
                  </span>
                  <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
                    {data.schedules.map((sc) => {
                      const isActive = isScheduleActiveNow(sc);
                      return (
                        <div
                          key={sc.id}
                          className="p-2.5 rounded-geist bg-[#111111] border border-[#222222] flex items-center justify-between text-xs font-mono"
                        >
                          <div className="flex items-center gap-2">
                            <span
                              className={cn(
                                'h-2 w-2 rounded-full shrink-0',
                                isActive ? 'bg-[#10b981]' : 'bg-[#666666]'
                              )}
                              title={isActive ? 'Active Now' : 'Scheduled'}
                            />
                            <span className="font-bold text-foreground">
                              {sc.scope === 'global' ? 'Global' : sc.scope}
                            </span>
                            <span
                              className={cn(
                                'text-xs px-1.5 py-0.5 rounded border',
                                sc.mode === 'auto'
                                  ? 'bg-[#183925] text-[#98c379] border-[#235836]'
                                  : 'bg-[#142334] text-[#52a8ff] border-[#1f4068]'
                              )}
                            >
                              {sc.mode === 'auto' ? 'Auto' : 'Request'}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              {sc.start_date ? `${sc.start_date} ` : ''}{sc.start_time} – {sc.end_date ? `${sc.end_date} ` : ''}{sc.end_time}
                            </span>
                          </div>
                          <button
                            type="button"
                            onClick={() => handleDeleteSchedulePolicy(sc.id)}
                            className="text-muted-foreground hover:text-status-danger cursor-pointer p-1"
                            title="Delete schedule"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      );
                    })}
                  </div>
                </div>
              ) : (
                <div className="p-4 text-center border border-dashed border-[#222222] rounded-geist bg-[#0d0d0d] text-xs text-muted-foreground font-mono">
                  No scheduled policies configured. Sharing respects the global and per-profile switches above.
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* TAB 3: AUDIT LOG */}
        <TabsContent value="audit" className="space-y-4 focus-visible:outline-none">
          {/* P2P Activity & Audit Log Card */}
          <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist flex flex-col h-[calc(100vh-230px)] min-h-[500px]">
            <CardHeader className="p-5 pb-3 space-y-3 shrink-0">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-accent-cyan" />
                  <CardTitle className="text-sm font-semibold text-foreground">
                    Audit Log
                  </CardTitle>
                  <Badge
                    variant="outline"
                    className="text-xs font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                  >
                    {filteredAuditLogs.length === (data?.audit_logs?.length || 0)
                      ? `${filteredAuditLogs.length} events`
                      : `${filteredAuditLogs.length} of ${data?.audit_logs?.length || 0} events`}
                  </Badge>
                </div>
                <span className="text-xs text-muted-foreground font-mono">
                  Tamper-evident mesh audit trail
                </span>
              </div>

              <CardDescription className="text-xs text-muted-foreground pt-0">
                Real-time audit log of all lease requests, approvals, revocations, and sharing policy switches.
              </CardDescription>

              {/* Audit Log Filters */}
              <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 pt-1">
                <div className="relative flex-1">
                  <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                  <Input
                    value={auditSearchQuery}
                    onChange={(e) => setAuditSearchQuery(e.target.value)}
                    placeholder="Filter audit events (actor, details, action)..."
                    className="h-8 pl-8 pr-7 text-xs font-mono bg-[#141414] border-border rounded-geist placeholder:text-muted-foreground"
                  />
                  {auditSearchQuery && (
                    <button
                      type="button"
                      onClick={() => setAuditSearchQuery('')}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground text-xs cursor-pointer"
                      title="Clear search"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  )}
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <div className="w-[160px]">
                    <SelectDropdown
                      value={auditActionFilter}
                      onChange={(val) => setAuditActionFilter(String(val))}
                      size="sm"
                      triggerClassName="h-8 bg-[#141414] border-border text-xs font-mono"
                      options={[
                        { value: 'all', label: 'All Actions' },
                        { value: 'granted', label: 'Leases Granted' },
                        { value: 'requested', label: 'Lease Requests' },
                        { value: 'revoked', label: 'Leases Revoked' },
                        { value: 'policy', label: 'Policy Changes' },
                        { value: 'schedule', label: 'Schedules' },
                      ]}
                    />
                  </div>

                  {(auditSearchQuery || auditActionFilter !== 'all') && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        setAuditSearchQuery('');
                        setAuditActionFilter('all');
                      }}
                      className="h-8 px-2 text-xs font-mono text-muted-foreground hover:text-foreground cursor-pointer"
                      title="Reset Filters"
                    >
                      <X className="h-3.5 w-3.5 mr-1" />
                      Reset
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            <CardContent className="p-5 pt-0 flex-1 min-h-0 flex flex-col">
              {!data?.audit_logs || data.audit_logs.length === 0 ? (
                <div className="flex-1 flex items-center justify-center p-6 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground font-mono">
                  No sharing events recorded yet.
                </div>
              ) : filteredAuditLogs.length === 0 ? (
                <div className="flex-1 flex flex-col items-center justify-center p-6 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground font-mono space-y-2">
                  <p>No audit events match the selected filters.</p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setAuditSearchQuery('');
                      setAuditActionFilter('all');
                    }}
                    className="h-6 px-2.5 text-xs font-mono cursor-pointer"
                  >
                    Clear filters
                  </Button>
                </div>
              ) : (
                <div className="flex-1 overflow-y-auto space-y-1.5 pr-1 font-mono text-xs min-h-0">
                  {filteredAuditLogs.map((log) => {
                    const actionColors: Record<string, string> = {
                      LEASE_GRANTED: 'text-[#98c379] bg-[#183925]/40 border-[#235836]',
                      LEASE_REQUESTED: 'text-[#52a8ff] bg-[#142334]/40 border-[#1f4068]',
                      LEASE_REVOKED: 'text-[#e06c75] bg-[#3a1d20]/40 border-[#5c282d]',
                      LEASE_DENIED: 'text-[#e06c75] bg-[#3a1d20]/40 border-[#5c282d]',
                      GLOBAL_MODE_CHANGED: 'text-[#e5c07b] bg-[#332b1a]/40 border-[#594926]',
                      PROFILE_SHARING_ON: 'text-[#98c379] bg-[#183925]/40 border-[#235836]',
                      PROFILE_SHARING_OFF: 'text-[#888888] bg-[#1e1e1e] border-[#333333]',
                      SCHEDULE_SAVED: 'text-[#56b6c2] bg-[#1a2d33]/40 border-[#264e59]',
                      SCHEDULE_DELETED: 'text-[#e5c07b] bg-[#332b1a]/40 border-[#594926]',
                    };
                    const color = actionColors[log.action] || 'text-muted-foreground border-border bg-[#141414]';
                    const timeStr = new Date(log.timestamp).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit',
                      second: '2-digit',
                    });
                    return (
                      <div
                        key={log.id}
                        className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-2 px-2.5 rounded-geist bg-[#111111] border border-[#1f1f1f] text-xs"
                      >
                        <div className="flex items-center gap-2 truncate min-w-0">
                          <span className="text-muted-foreground shrink-0">{timeStr}</span>
                          <span className={cn('px-1.5 py-0.5 rounded border text-xs shrink-0 font-medium', color)}>
                            {log.action.replace(/_/g, ' ')}
                          </span>
                          <span className="text-foreground truncate">{log.details}</span>
                        </div>
                        {log.actor && (
                          <span className="text-muted-foreground text-xs shrink-0 truncate max-w-[150px]">
                            actor: {log.actor}
                          </span>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Connect Peer Dialog Modal */}
      <Dialog open={isConnectOpen} onOpenChange={setIsConnectOpen}>
        <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
          <form onSubmit={handleAddPeer}>
            <DialogHeader className="space-y-1 pb-3 border-b border-[#222222] pr-6">
              <div className="flex items-center gap-2">
                <Laptop className="h-4 w-4 text-[#61afef]" />
                <DialogTitle className="text-base font-semibold text-[#ededed]">
                  Connect Teammate Peer
                </DialogTitle>
              </div>
              <DialogDescription className="text-xs text-[#888888]">
                Enter a teammate&apos;s hostname (e.g. <code>Adrijs-MacBook-Pro.local</code>) or their VPN tunnel IP (e.g. <code>10.8.16.9</code>).
              </DialogDescription>
            </DialogHeader>

            <div className="py-4 space-y-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-[#ededed]">
                  Teammate Hostname or IP
                </label>
                <Input
                  autoFocus
                  type="text"
                  value={addHost}
                  onChange={(e) => setAddHost(e.target.value)}
                  placeholder="Adrijs-MacBook-Pro.local or 10.8.16.9"
                  className="h-9 text-xs font-mono bg-[#141414] border-[#262626] rounded-geist text-foreground"
                />
              </div>
              <p className="text-[11px] font-mono text-[#666666]">
                AIM will verify connectivity over port :4319 and query available shared profiles.
              </p>
            </div>

            <DialogFooter className="pt-2 border-t border-[#222222]">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setIsConnectOpen(false)}
                className="rounded-geist text-xs border-border"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                size="sm"
                disabled={isAdding || !addHost.trim()}
                className="rounded-geist text-xs bg-foreground text-background hover:bg-foreground/90 font-medium"
              >
                {isAdding ? 'Connecting...' : 'Connect Peer'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Request Profile Lease Dialog Modal */}
      <Dialog
        open={!!(selectedPeer && selectedProfile)}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedPeer(null);
            setSelectedProfile('');
          }
        }}
      >
        <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
          <DialogHeader className="space-y-1 pb-3 border-b border-[#222222] pr-6">
            <div className="flex items-center gap-2">
              <Radio className="h-4 w-4 text-accent-cyan" />
              <DialogTitle className="text-base font-semibold text-[#ededed]">
                Request Profile Lease
              </DialogTitle>
            </div>
            <DialogDescription className="text-xs text-[#888888]">
              Borrow credentials from teammate on internal VPN subnet.
            </DialogDescription>
          </DialogHeader>

          <div className="py-2 space-y-4">
            {/* Metadata Strip matching Geist standards */}
            <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-2">
              <div className="flex items-center justify-between text-[#888888]">
                <span>Remote Host</span>
                <span className="font-mono text-[#ededed] font-medium flex items-center gap-1.5">
                  <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
                  {selectedPeer?.hostname}
                </span>
              </div>
              <div className="flex items-center justify-between text-[#888888]">
                <span>Offered Profile</span>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[#ededed] font-medium">
                    {selectedProfile}
                  </span>
                  <AgentBadge agent={selectedAgent} size="sm" />
                </div>
              </div>
              <div className="flex items-center justify-between text-[#888888]">
                <span>Approval Policy</span>
                <span
                  className={cn(
                    'font-mono text-[11px] font-medium px-1.5 py-0.5 rounded border',
                    selectedIsAuto
                      ? 'bg-[#183925]/50 text-[#98c379] border-[#235836]/60'
                      : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                  )}
                >
                  {selectedIsAuto ? 'Instant Auto-Approve' : 'Manual Approval Required'}
                </span>
              </div>
              <div className="pt-2 border-t border-[#1f1f1f] flex items-center justify-between">
                <span className="text-[#888888]">Local Profile ID</span>
                <span className="font-mono font-bold text-accent-cyan">
                  {selectedPeer?.hostname}:{selectedProfile}
                </span>
              </div>
            </div>

            {/* Lease Duration with SelectDropdown */}
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="text-xs font-semibold text-[#ededed]">
                  Lease Duration
                </label>
                <span className="text-xs text-muted-foreground font-mono">
                  Auto-purged after expiry
                </span>
              </div>
              <SelectDropdown
                value={leaseHours}
                onChange={(val) => setLeaseHours(Number(val))}
                size="sm"
                options={[
                  { value: 1, label: '1 hour' },
                  { value: 2, label: '2 hours' },
                  { value: 4, label: '4 hours (Recommended)' },
                  { value: 8, label: '8 hours (Full Workday)' },
                ]}
              />
              <p className="text-[11px] text-[#666666] leading-relaxed pt-1">
                Once granted, credentials operate 100% locally in your isolated profile directory. The teammate&apos;s machine can sleep or disconnect without interrupting your active sessions.
              </p>
            </div>
          </div>

          <DialogFooter className="pt-2 border-t border-[#222222]">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setSelectedPeer(null);
                setSelectedProfile('');
              }}
              className="rounded-geist text-xs border-border"
            >
              Cancel
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={isRequesting}
              onClick={handleRequestLease}
              className="rounded-geist text-xs font-medium bg-foreground text-background hover:bg-foreground/90"
            >
              {isRequesting
                ? 'Requesting...'
                : selectedIsAuto
                ? 'Instant Borrow'
                : 'Confirm Lease'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

    </div>
  );
}
