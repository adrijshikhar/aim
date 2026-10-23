import * as React from 'react';
import {
  Wifi,
  AlertTriangle,
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
  Calendar as CalendarIcon,
  Inbox,
  FileCheck,
  Ban,
  Send,
  Zap,
  Lock,
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
  cancelShareRequest,
  pollShareRequest,
  ShareStatusDTO,
  PeerInfoDTO,
  ProfileDTO,
  ActiveLeaseDTO,
  SchedulePolicyDTO,
} from '@/lib/api';
import { useToast } from '@/components/ui/use-toast';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { AgentBadge, getAgentDotColor } from '@/components/AgentBadge';
import { SelectDropdown, type SelectOption } from '@/components/ui/select-dropdown';
import { SegmentedControl } from '@/components/ui/segmented-control';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from '@/components/ui/tooltip';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';
import { format, parseISO, isValid } from 'date-fns';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
import { TeammateCardsSkeleton } from '@/components/Skeletons';

const TIME_OPTIONS: SelectOption<string>[] = Array.from({ length: 48 }, (_, i) => {
  const h = Math.floor(i / 2);
  const m = i % 2 === 0 ? '00' : '30';
  const val = `${h.toString().padStart(2, '0')}:${m}`;
  const period = h < 12 ? 'AM' : 'PM';
  const displayH = h % 12 === 0 ? 12 : h % 12;
  const label = `${displayH.toString().padStart(2, '0')}:${m} ${period}`;
  return { value: val, label };
});

function getTimeOptions(currentVal: string): SelectOption<string>[] {
  const exists = TIME_OPTIONS.some((o) => o.value === currentVal);
  if (!exists && currentVal) {
    return [{ value: currentVal, label: currentVal }, ...TIME_OPTIONS];
  }
  return TIME_OPTIONS;
}

function ScheduleDatePicker({
  value,
  onChange,
  placeholder = 'Pick date',
}: {
  value: string;
  onChange: (val: string) => void;
  placeholder?: string;
}) {
  const [open, setOpen] = React.useState(false);
  const selectedDate = React.useMemo(() => {
    if (!value) return undefined;
    try {
      const parsed = parseISO(value);
      return isValid(parsed) ? parsed : undefined;
    } catch {
      return undefined;
    }
  }, [value]);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          className={cn(
            'w-full h-9 justify-start text-left font-mono text-xs border-border bg-secondary hover:bg-muted text-foreground flex items-center gap-2 rounded-geist cursor-pointer',
            !selectedDate && 'text-muted-foreground'
          )}
        >
          <CalendarIcon className="h-3.5 w-3.5 text-accent-cyan shrink-0" />
          <span>{selectedDate ? format(selectedDate, 'MMM d, yyyy') : placeholder}</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0 z-50 bg-[#111317] border-border shadow-2xl" align="start">
        <Calendar
          mode="single"
          selected={selectedDate}
          onSelect={(d) => {
            if (d) {
              onChange(format(d, 'yyyy-MM-dd'));
              setOpen(false);
            }
          }}
        />
      </PopoverContent>
    </Popover>
  );
}

type ShareMode = 'off' | 'request' | 'auto';

// Module-level in-memory cache and sessionStorage persistence to guarantee 0ms instant view switching and reload with zero skeleton flash
let networkMemoryCache: ShareStatusDTO | null = null;
const NETWORK_STORAGE_KEY = 'aim_network_status_cache_v1';

function getInitialNetworkStatus(): ShareStatusDTO | null {
  if (networkMemoryCache) {
    return networkMemoryCache;
  }
  try {
    const raw = sessionStorage.getItem(NETWORK_STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (parsed && typeof parsed === 'object') {
        networkMemoryCache = parsed;
        return parsed;
      }
    }
  } catch {
    // Ignore storage/parse errors
  }
  return null;
}

function saveNetworkStatusToCache(status: ShareStatusDTO) {
  networkMemoryCache = status;
  try {
    sessionStorage.setItem(NETWORK_STORAGE_KEY, JSON.stringify(status));
  } catch {
    // Ignore storage errors
  }
}

interface BorrowedLeaseCardProps {
  lease: ActiveLeaseDTO;
  currentTime: number;
  formatExpiresIn: (expiresAt: string) => string;
  copiedCmd: string | null;
  onCopy: (text: string, key: string) => void;
  onRelease: (localProfile: string) => void;
}

function BorrowedLeaseCard({
  lease,
  currentTime,
  formatExpiresIn,
  copiedCmd,
  onCopy,
  onRelease,
}: BorrowedLeaseCardProps) {
  const isRevoked = Boolean(lease.is_revoked);
  const isExpired = lease.is_expired || isRevoked || new Date(lease.expires_at).getTime() <= currentTime;
  const installedAgents = (lease.agents && lease.agents.length > 0) ? lease.agents : [lease.agent || 'claude'];

  const [selectedAgent, setSelectedAgent] = React.useState<string>(installedAgents[0] || 'claude');

  React.useEffect(() => {
    if (!installedAgents.includes(selectedAgent)) {
      setSelectedAgent(installedAgents[0] || 'claude');
    }
  }, [installedAgents, selectedAgent]);

  const activeRunCmd = `aim run ${selectedAgent} ${lease.local_profile}`;
  const copyKey = `${lease.local_profile}-${selectedAgent}`;

  return (
    <div className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3.5 shadow-xs hover:border-[#383838] transition-colors">
      <div className="space-y-1.5">
        <div className="flex items-start justify-between gap-2">
          <div className="space-y-0.5 min-w-0 flex-1">
            <span
              className="text-sm font-mono font-bold text-accent-cyan tracking-tight block truncate select-all"
              title={lease.local_profile}
            >
              {lease.local_profile}
            </span>
            <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-mono truncate">
              <span>Source:</span>
              <span className="text-foreground font-semibold truncate">{lease.source_peer}</span>
              <span className="text-muted-foreground/60">/</span>
              <span className="text-foreground truncate">{lease.source_profile}</span>
            </div>
          </div>

          <div className="flex items-center gap-1.5 shrink-0">
            <span
              className={cn(
                'text-xs font-mono px-2 py-0.5 rounded-[4px] font-medium inline-flex items-center gap-1.5',
                isRevoked || isExpired
                  ? 'bg-status-danger-bg text-status-danger border border-status-danger-border'
                  : 'bg-status-success-bg text-status-success border border-status-success-border'
              )}
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full shrink-0',
                  isRevoked || isExpired ? 'bg-status-danger-solid' : 'bg-status-success-solid'
                )}
              />
              <span>{isRevoked ? 'Revoked' : formatExpiresIn(lease.expires_at)}</span>
            </span>

            <Button
              variant="outline"
              size="icon"
              onClick={() => onRelease(lease.local_profile)}
              className="h-6 w-6 rounded-md border-border bg-secondary hover:bg-status-danger/10 text-muted-foreground hover:text-status-danger hover:border-status-danger/30 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
              title="Release Lease"
              aria-label="Release Lease"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </div>

      <div className="space-y-2 pt-2 border-t border-border/40">
        {installedAgents.length > 1 && (
          <div className="space-y-1.5">
            <div className="flex items-center justify-between text-[10px] font-mono uppercase tracking-wider text-muted-foreground font-medium select-none">
              <span>Runnable Engines ({installedAgents.length}):</span>
              <span className="text-[10px] text-muted-foreground/70 normal-case font-sans">Click engine to switch</span>
            </div>
            <SegmentedControl
              size="sm"
              value={selectedAgent}
              onChange={(ag) => setSelectedAgent(ag)}
              options={installedAgents.map((ag) => ({
                value: ag,
                label: ag,
                dotColor: getAgentDotColor(ag),
              }))}
            />
          </div>
        )}

        <div className="flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md bg-[#111111] border border-border font-mono text-xs shadow-xs">
          <div className="flex items-center gap-2 min-w-0 flex-1 text-muted-foreground select-all truncate">
            <span className="text-accent-cyan font-bold select-none shrink-0">$</span>
            <span className="text-foreground truncate whitespace-nowrap">
              aim run <strong className="text-accent-cyan font-semibold">{selectedAgent}</strong> {lease.local_profile}
            </span>
          </div>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => onCopy(activeRunCmd, copyKey)}
            className="h-6 px-2 text-[11px] font-sans text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer shrink-0 transition-colors"
            title={`Copy: ${activeRunCmd}`}
          >
            {copiedCmd === copyKey ? (
              <>
                <Check className="h-3 w-3 text-status-success mr-1" />
                <span className="text-status-success">Copied</span>
              </>
            ) : (
              <>
                <Copy className="h-3 w-3 mr-1" />
                <span>Copy</span>
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}

interface NetworkViewProps {
  profiles?: ProfileDTO[];
}

export function NetworkView({ profiles = [] }: NetworkViewProps) {
  const { toast } = useToast();
  const initialData = React.useMemo(() => getInitialNetworkStatus(), []);
  const [data, setData] = React.useState<ShareStatusDTO | null>(initialData);
  const [isLoading, setIsLoading] = React.useState(!initialData);
  const [isRefreshing, setIsRefreshing] = React.useState(false);
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
  const [selectedAgents, setSelectedAgents] = React.useState<string[]>([]);
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

  const [cancellingReqId, setCancellingReqId] = React.useState<string | null>(null);
  const [pollingReqId, setPollingReqId] = React.useState<string | null>(null);

  // Revoke Lease Confirmation Modal state
  const [leaseToRevoke, setLeaseToRevoke] = React.useState<{ leaseId: string; profileName: string; borrowerId?: string } | null>(null);
  const [isRevoking, setIsRevoking] = React.useState(false);

  // Release Lease Confirmation Modal state
  const [leaseToRelease, setLeaseToRelease] = React.useState<string | null>(null);
  const [isReleasing, setIsReleasing] = React.useState(false);

  // Requests Tab Filter state
  const [requestsFilter, setRequestsFilter] = React.useState<'all' | 'pending' | 'outgoing' | 'approved' | 'revoked'>('all');
  const [requestSearchQuery, setRequestSearchQuery] = React.useState('');

  // Audit Log filter state
  const [auditSearchQuery, setAuditSearchQuery] = React.useState('');
  const [auditActionFilter, setAuditActionFilter] = React.useState<string>('all');
  const [visibleAuditCount, setVisibleAuditCount] = React.useState(50);
  const auditSentinelRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const timer = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);

  const searchInputRef = React.useRef<HTMLInputElement>(null);

  const loadStatus = React.useCallback(async (refresh = false) => {
    try {
      const res = await getShareStatus(refresh);
      setData(res);
      saveNetworkStatusToCache(res);

      if (res.outgoing_requests && res.outgoing_requests.some((r) => r.status === 'pending')) {
        for (const r of res.outgoing_requests) {
          if (r.status === 'pending') {
            try {
              const pollRes = await pollShareRequest(r.id);
              if (pollRes && pollRes.status === 'approved') {
                toast({
                  title: 'Lease Approved!',
                  description: `Teammate approved lease for ${r.profile_name}. Profile is active.`,
                  variant: 'success',
                });
              }
            } catch {
              // Ignore until host approves
            }
          }
        }
      }
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
    // 1. Fast local daemon fetch (<10ms) to ensure instant UI rendering and sync with local daemon
    loadStatus(false).then(() => {
      // 2. Background subnet mesh discovery without blocking UI or showing skeletons
      setIsRefreshing(true);
      getShareStatus(true)
        .then((fresh) => {
          setData(fresh);
          saveNetworkStatusToCache(fresh);
        })
        .catch(() => {})
        .finally(() => setIsRefreshing(false));
    });

    let count = 0;
    const interval = setInterval(() => {
      if (document.hidden) return;
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
    setIsRefreshing(true);
    try {
      await loadStatus(true);
    } finally {
      setIsRefreshing(false);
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

  const handleCancelRequest = async (requestId: string) => {
    setCancellingReqId(requestId);
    try {
      const res = await cancelShareRequest(requestId);
      if (res?.warning) {
        toast({
          title: 'Request Cancelled Locally',
          description: res.warning,
          variant: 'default',
        });
      } else {
        toast({
          title: 'Request Cancelled',
          description: `Cancelled lease request ${requestId}`,
          variant: 'default',
        });
      }
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to cancel request';
      toast({
        title: 'Cancellation Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setCancellingReqId(null);
    }
  };

  const handleManualPoll = async (requestId: string) => {
    setPollingReqId(requestId);
    try {
      const res = await pollShareRequest(requestId);
      if (res && res.status === 'approved') {
        toast({
          title: 'Lease Approved!',
          description: `Teammate approved request. Profile is now ready to use.`,
          variant: 'success',
        });
      } else if (res && res.status === 'denied') {
        toast({
          title: 'Request Denied',
          description: `Teammate declined request.`,
          variant: 'destructive',
        });
      } else {
        toast({
          title: 'Status Polled',
          description: `Request is still pending teammate approval.`,
          variant: 'default',
        });
      }
      await loadStatus(false);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to poll status';
      toast({
        title: 'Poll Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setPollingReqId(null);
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
    const totalHours = Math.floor(diff / (1000 * 60 * 60));
    const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    if (totalHours >= 24) {
      const days = Math.floor(totalHours / 24);
      const remHours = totalHours % 24;
      return remHours > 0 ? `${days}d ${remHours}h remaining` : `${days}d remaining`;
    }
    if (totalHours === 0 && mins === 0) return '< 1m remaining';
    if (totalHours === 0) return `${mins}m remaining`;
    return `${totalHours}h ${mins}m remaining`;
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
  const outgoingRequests = data?.outgoing_requests || [];
  const outgoingPending = React.useMemo(() => {
    return outgoingRequests.filter((r) => r.status === 'pending');
  }, [outgoingRequests]);

  const activeBorrowedLeases = React.useMemo(() => {
    return (data?.active_leases || []).filter(
      (l) => !l.is_expired && !l.is_revoked && new Date(l.expires_at).getTime() > currentTime
    );
  }, [data?.active_leases, currentTime]);

  const expiredBorrowedLeases = React.useMemo(() => {
    return (data?.active_leases || []).filter(
      (l) => l.is_expired || l.is_revoked || new Date(l.expires_at).getTime() <= currentTime
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
        if (auditActionFilter === 'granted' && !['LEASE_GRANTED', 'LEASE_APPROVED', 'LEASE_BORROWED'].includes(log.action)) return false;
        if (auditActionFilter === 'requested' && log.action !== 'LEASE_REQUESTED') return false;
        if (auditActionFilter === 'cancelled' && log.action !== 'LEASE_CANCELLED') return false;
        if (auditActionFilter === 'revoked' && !['LEASE_REVOKED', 'LEASE_DENIED'].includes(log.action)) return false;
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

  // Filtered requests by search query
  const filteredPendingRequests = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return pendingRequests;
    const q = requestSearchQuery.toLowerCase();
    return pendingRequests.filter(
      (pr) =>
        pr.request.profile_name.toLowerCase().includes(q) ||
        pr.request.requester_id.toLowerCase().includes(q) ||
        pr.request.requester_ip.toLowerCase().includes(q) ||
        pr.request_id.toLowerCase().includes(q)
    );
  }, [pendingRequests, requestSearchQuery]);

  const filteredOutgoingRequests = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return outgoingPending;
    const q = requestSearchQuery.toLowerCase();
    return outgoingPending.filter(
      (r) =>
        r.profile_name.toLowerCase().includes(q) ||
        r.peer_id.toLowerCase().includes(q) ||
        r.peer_ip.toLowerCase().includes(q) ||
        r.id.toLowerCase().includes(q)
    );
  }, [outgoingPending, requestSearchQuery]);

  const filteredActiveGranted = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return activeGrantedLeases;
    const q = requestSearchQuery.toLowerCase();
    return activeGrantedLeases.filter(
      (l) =>
        l.profile_name.toLowerCase().includes(q) ||
        l.borrower_id.toLowerCase().includes(q) ||
        l.lease_id.toLowerCase().includes(q)
    );
  }, [activeGrantedLeases, requestSearchQuery]);

  const filteredActiveBorrowed = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return activeBorrowedLeases;
    const q = requestSearchQuery.toLowerCase();
    return activeBorrowedLeases.filter(
      (l) =>
        l.local_profile.toLowerCase().includes(q) ||
        l.source_peer.toLowerCase().includes(q) ||
        l.source_profile.toLowerCase().includes(q)
    );
  }, [activeBorrowedLeases, requestSearchQuery]);

  const filteredExpiredBorrowed = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return expiredBorrowedLeases;
    const q = requestSearchQuery.toLowerCase();
    return expiredBorrowedLeases.filter(
      (l) =>
        l.local_profile.toLowerCase().includes(q) ||
        l.source_peer.toLowerCase().includes(q) ||
        l.source_profile.toLowerCase().includes(q)
    );
  }, [expiredBorrowedLeases, requestSearchQuery]);

  const filteredRevokedOrExpired = React.useMemo(() => {
    if (!requestSearchQuery.trim()) return revokedOrExpiredLeases;
    const q = requestSearchQuery.toLowerCase();
    return revokedOrExpiredLeases.filter(
      (l) =>
        l.profile_name.toLowerCase().includes(q) ||
        l.borrower_id.toLowerCase().includes(q) ||
        l.lease_id.toLowerCase().includes(q)
    );
  }, [revokedOrExpiredLeases, requestSearchQuery]);

  const totalRequestsCount = React.useMemo(() => {
    return (
      pendingRequests.length +
      outgoingPending.length +
      activeGrantedLeases.length +
      activeBorrowedLeases.length +
      expiredBorrowedLeases.length +
      revokedOrExpiredLeases.length
    );
  }, [
    pendingRequests.length,
    outgoingPending.length,
    activeGrantedLeases.length,
    activeBorrowedLeases.length,
    expiredBorrowedLeases.length,
    revokedOrExpiredLeases.length,
  ]);

  // Audit Log Pagination (50 records per page + Infinite Scrolling)
  React.useEffect(() => {
    setVisibleAuditCount(50);
  }, [auditSearchQuery, auditActionFilter]);

  React.useEffect(() => {
    const sentinel = auditSentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setVisibleAuditCount((prev) => {
            if (prev < filteredAuditLogs.length) {
              return Math.min(prev + 50, filteredAuditLogs.length);
            }
            return prev;
          });
        }
      },
      { rootMargin: '200px' }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [filteredAuditLogs.length]);

  const paginatedAuditLogs = React.useMemo(() => {
    return filteredAuditLogs.slice(0, visibleAuditCount);
  }, [filteredAuditLogs, visibleAuditCount]);

  return (
    <div className="space-y-6">
      {/* Top Header bar with Global Policy Segmented Control (Always Visible in All Cases) */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border">
        <div className="space-y-1">
          <div className="flex items-center gap-2 flex-wrap">
            <Radio className="h-4 w-4 text-accent-cyan" />
            <h1 className="text-xl font-semibold tracking-tight text-foreground font-sans">
              P2P Network Sharing
            </h1>
            <Badge
              variant="outline"
              className="font-mono text-[10px] uppercase tracking-wider py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground font-medium"
            >
              Mesh
            </Badge>
          </div>
        </div>

        {/* Global Sharing Policy Segmented Control — Always visible regardless of teammates count or active tab */}
        <div className="flex items-center gap-2 shrink-0">
          <SegmentedControl
            value={effectiveGlobalMode}
            onChange={(val) => handleGlobalModeChange(val as 'off' | 'request' | 'auto')}
            disabled={!data || isShareUpdating}
            prefix={
              <>
                <Radio className={cn('h-3 w-3', effectiveGlobalMode !== 'off' ? 'text-accent-cyan' : 'text-muted-foreground')} />
                <span>Global Policy:</span>
              </>
            }
            options={[
              {
                value: 'off',
                label: 'Off',
                dotColor: effectiveGlobalMode === 'off' ? 'bg-[#999999]' : 'bg-[#555555]',
                title: 'Global sharing is OFF — node is private and unreachable by teammates',
              },
              {
                value: 'request',
                label: 'Request',
                dotColor: effectiveGlobalMode === 'request' ? 'bg-[#52a8ff]' : 'bg-[#555555]',
                activeClassName: 'bg-[#142334] text-[#52a8ff] border-[#1f4068]',
                title: 'Global sharing is in Request mode — teammate lease requests require manual approval',
              },
              {
                value: 'auto',
                label: 'Auto Approve',
                dotColor: effectiveGlobalMode === 'auto' ? 'bg-[#30a46c]' : 'bg-[#555555]',
                activeClassName: 'bg-[#183925] text-[#98c379] border-[#235836]',
                title: 'Global sharing is in Auto mode — teammate lease requests are automatically approved',
              },
            ]}
          />
        </div>
      </div>

      {/* VPN Tunnel Status Strip */}
      <div className="p-4 rounded-geist border border-border bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0">
          <div className="flex items-center justify-center h-8 w-8 rounded-full border border-border bg-secondary shrink-0">
            {isVpnActive ? (
              <Wifi className="h-4 w-4 text-status-success" />
            ) : isLoading && !data ? (
              <RotateCw className="h-4 w-4 text-muted-foreground animate-spin" />
            ) : (
              <AlertTriangle className="h-4 w-4 text-status-warning" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono font-medium text-foreground">
                {isVpnActive
                  ? 'Subnet Tunnel Connected'
                  : isLoading && !data
                  ? 'Checking VPN Tunnel...'
                  : 'VPN Tunnel Inactive'}
              </span>
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  isVpnActive
                    ? 'bg-status-success animate-pulse'
                    : isLoading && !data
                    ? 'bg-muted-foreground animate-pulse'
                    : 'bg-status-warning'
                )}
              />
            </div>
            <div className="text-xs text-muted-foreground font-mono mt-0.5 flex items-center flex-wrap gap-x-2 gap-y-1">
              {isVpnActive && vpn ? (
                <>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="cursor-help underline decoration-dotted underline-offset-2">
                          Assigned IP: {vpn.ip} ({vpn.interface_name})
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs text-xs font-sans">
                        Your device&apos;s private IP address on the VPN network tunnel interface.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  <span>•</span>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="cursor-help underline decoration-dotted underline-offset-2">
                          Broadcast: {vpn.broadcast_ip}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs text-xs font-sans">
                        Subnet broadcast address used to automatically announce and discover teammates without a centralized server.
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </>
              ) : isLoading && !data ? (
                <span>Detecting network tunnel interface and assigned subnet IP...</span>
              ) : (
                <span>Connect to your corporate VPN (Pritunl / OpenVPN / Wireguard) to discover teammates across your subnet.</span>
              )}
            </div>
          </div>
        </div>

        {isVpnActive && (
          <div className="flex items-center gap-2 shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                if (vpn?.ip) {
                  navigator.clipboard.writeText(vpn.ip);
                  toast({ title: 'VPN IP Copied', description: vpn.ip, variant: 'success' });
                }
              }}
              className="h-8 text-xs font-mono rounded-geist border-border bg-secondary hover:bg-muted text-foreground flex items-center gap-1.5 cursor-pointer"
            >
              <Copy className="h-3 w-3" />
              <span>Copy IP</span>
            </Button>
          </div>
        )}
      </div>

      {/* Tab Navigation System */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-1">
          <TabsList className="bg-secondary/40 border border-border p-1 h-9 rounded-geist gap-1 w-full sm:w-auto">
            <TabsTrigger value="teammates" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
              <span>Teammates</span>
              {filteredPeers.length > 0 && (
                <span className="ml-1 px-1.5 py-0.5 rounded-[4px] text-[11px] font-mono bg-secondary/80 text-muted-foreground border border-border/60">
                  {filteredPeers.length}
                </span>
              )}
            </TabsTrigger>

            <TabsTrigger value="requests" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Inbox className="h-3.5 w-3.5 text-[#e5c07b]" />
              <span>Requests</span>
              {(pendingRequests.length + outgoingPending.length) > 0 ? (
                <span className="ml-1 px-1.5 py-0.5 rounded-[4px] text-[11px] font-mono bg-[#142334] text-[#52a8ff] border border-[#1f4068] font-bold animate-pulse">
                  {pendingRequests.length + outgoingPending.length}
                </span>
              ) : (activeGrantedLeases.length + activeBorrowedLeases.length) > 0 ? (
                <span className="ml-1 px-1.5 py-0.5 rounded-[4px] text-[11px] font-mono bg-secondary/80 text-muted-foreground border border-border/60">
                  {activeGrantedLeases.length + activeBorrowedLeases.length}
                </span>
              ) : null}
            </TabsTrigger>

            <TabsTrigger value="my-sharing" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Share2 className="h-3.5 w-3.5 text-accent-cyan" />
              <span>My Sharing</span>
              <span className="ml-1 px-1.5 py-0.5 rounded-[4px] text-[11px] font-mono bg-secondary/80 text-muted-foreground border border-border/60">
                {data?.shared_profiles?.filter((sp) => sp.enabled)?.length || 0}/{ownProfiles.length}
              </span>
            </TabsTrigger>

            <TabsTrigger value="audit" className="gap-2 text-xs flex-1 sm:flex-initial">
              <Activity className="h-3.5 w-3.5 text-accent-cyan" />
              <span>Audit Log</span>
              {((data?.total_audit_count && data.total_audit_count > 0) || (data?.audit_logs && data.audit_logs.length > 0)) ? (
                <span className="ml-1 px-1.5 py-0.5 rounded-[4px] text-[11px] font-mono bg-secondary/80 text-muted-foreground border border-border/60">
                  {data?.total_audit_count || data?.audit_logs?.length}
                </span>
              ) : null}
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
            <Card className="border-border bg-card rounded-geist">
              <CardHeader className="p-5 pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-[#61afef]" />
                    <CardTitle className="text-sm font-semibold text-foreground">
                      Borrowed Profiles
                    </CardTitle>
                    <Badge
                      variant="outline"
                      className="text-[11px] font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground"
                    >
                      {data.active_leases.filter((l) => !l.is_expired && !l.is_revoked && new Date(l.expires_at).getTime() > currentTime).length} active
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    {data.active_leases.some((l) => l.is_expired || l.is_revoked || new Date(l.expires_at).getTime() <= currentTime) && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleCleanExpired}
                        className="h-7 px-2.5 rounded-geist border-border bg-secondary hover:bg-muted hover:border-status-danger/40 hover:text-status-danger text-xs font-mono font-medium text-foreground inline-flex items-center gap-1.5 cursor-pointer shadow-xs transition-colors"
                      >
                        <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                        <span>Clean expired</span>
                      </Button>
                    )}
                    <Badge
                      variant="outline"
                      className="text-[11px] font-mono py-0.5 px-2 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground font-normal"
                    >
                      Sandboxed locally
                    </Badge>
                  </div>
                </div>
                <CardDescription className="text-xs text-muted-foreground pt-1">
                  External teammate profiles currently leased and sandboxed locally on this machine.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {data.active_leases.map((lease) => (
                    <BorrowedLeaseCard
                      key={lease.local_profile}
                      lease={lease}
                      currentTime={currentTime}
                      formatExpiresIn={formatExpiresIn}
                      copiedCmd={copiedCmd}
                      onCopy={copyToClipboard}
                      onRelease={(prof) => setLeaseToRelease(prof)}
                    />
                  ))}
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
                    className="text-[11px] font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground"
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
                    className="pl-8 pr-8 h-8 text-xs font-sans bg-secondary border-border text-foreground placeholder:text-muted-foreground rounded-geist focus:border-border/80"
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
                    <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 border border-border bg-muted px-1 py-[1px] rounded text-xs font-mono text-muted-foreground pointer-events-none">
                      /
                    </kbd>
                  )}
                </div>

                {/* Connect Peer Icon Button */}
                <Button
                  variant="default"
                  size="icon"
                  onClick={() => setIsConnectOpen(true)}
                  className="h-8 w-8 rounded-geist bg-primary text-primary-foreground hover:bg-primary/90 flex items-center justify-center shrink-0 cursor-pointer shadow-xs"
                  title="Connect Peer"
                >
                  <Plus className="h-4 w-4" />
                </Button>

                {/* Refresh Icon Button */}
                <Button
                  variant="outline"
                  size="sm"
                  disabled={isLoading || isRefreshing}
                  onClick={handleManualRefresh}
                  className="h-8 w-8 p-0 rounded-geist border-border bg-secondary hover:bg-muted text-foreground flex items-center justify-center shrink-0 cursor-pointer"
                  title="Refresh Teammates"
                >
                  <RotateCw className={cn('h-3.5 w-3.5', (isLoading || isRefreshing) && 'animate-spin text-accent-cyan')} />
                </Button>
              </div>
            </div>

            <div>
              {/* Subtle Scanning Indicator when refreshing existing data */}
              {isRefreshing && (
                <div className="mb-4 flex items-center justify-between px-3.5 py-2 rounded-geist border border-accent-cyan/30 bg-accent-cyan/5 text-accent-cyan text-xs font-mono animate-in fade-in duration-150">
                  <div className="flex items-center gap-2">
                    <RotateCw className="h-3.5 w-3.5 animate-spin shrink-0" />
                    <span>Scanning subnet mesh for teammates &amp; profile statuses...</span>
                  </div>
                  <span className="text-[11px] opacity-70">Live polling</span>
                </div>
              )}

              {isLoading && !data ? (
                <TeammateCardsSkeleton count={3} />
              ) : !data || !data.peers || data.peers.length === 0 ? (
                <div className="space-y-4">
                  {/* Your Node & Global Sharing Policy Card (Always visible even when 0 teammates discovered) */}
                  <div className="p-5 rounded-geist border border-border bg-card space-y-4 shadow-xs">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border/40">
                      <div className="space-y-0.5 min-w-0">
                        <div className="flex items-center gap-2">
                          <Laptop className="h-4 w-4 text-[#61afef] shrink-0" />
                          <h3 className="text-sm font-semibold text-foreground font-sans truncate">
                            Your Local Node &amp; Sharing Policy
                          </h3>
                          <Badge
                            variant="outline"
                            className={cn(
                              'text-xs font-mono py-0 px-2 rounded-full font-medium',
                              effectiveGlobalMode === 'auto'
                                ? 'border-[#235836] bg-[#183925]/50 text-[#98c379]'
                                : effectiveGlobalMode === 'request'
                                  ? 'border-[#1f4068] bg-[#142334]/50 text-[#52a8ff]'
                                  : 'border-border bg-secondary text-muted-foreground'
                            )}
                          >
                            {effectiveGlobalMode === 'auto'
                              ? 'Global: Auto-Approve'
                              : effectiveGlobalMode === 'request'
                                ? 'Global: Request Mode'
                                : 'Global: Private (Off)'}
                          </Badge>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          {isVpnActive
                            ? `Active on subnet as ${vpn?.ip}:4319 • Ready for teammate discovery`
                            : 'Local mesh node ready. Connect to VPN to enable automated teammate discovery.'}
                        </p>
                      </div>

                      {isVpnActive && vpn?.ip && (
                        <div className="flex items-center gap-2 shrink-0">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              const addr = `${vpn.ip}:4319`;
                              navigator.clipboard.writeText(addr);
                              toast({
                                title: 'Peer Address Copied',
                                description: addr,
                                variant: 'success',
                              });
                            }}
                            className="h-8 text-xs font-mono rounded-geist border-border bg-secondary hover:bg-muted text-foreground flex items-center gap-1.5 cursor-pointer"
                            title="Copy your node address for teammates"
                          >
                            <Copy className="h-3 w-3" />
                            <span>Copy My Endpoint</span>
                          </Button>
                        </div>
                      )}
                    </div>

                    {/* Node Shared Profiles Breakdown */}
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-1">
                      <div className="space-y-1">
                        <span className="text-xs font-sans font-medium text-foreground">
                          Offered Profiles ({data?.shared_profiles?.filter((sp) => sp.enabled)?.length || 0} of {ownProfiles.length} shared)
                        </span>
                        <div className="flex items-center gap-1.5 flex-wrap pt-0.5">
                          {ownProfiles.length === 0 ? (
                            <span className="text-xs text-muted-foreground">No local profiles created yet.</span>
                          ) : (
                            ownProfiles.map((p) => {
                              const sp = data?.shared_profiles?.find((s) => s.name === p.name);
                              const isEnabled = sp ? sp.enabled : false;
                              const isAuto = sp ? sp.auto_approve : false;
                              return (
                                <Badge
                                  key={p.name}
                                  variant="outline"
                                  className={cn(
                                    'text-xs font-mono py-0.5 px-2 rounded-geist flex items-center gap-1.5',
                                    isEnabled
                                      ? 'border-border bg-secondary text-foreground'
                                      : 'border-border/40 bg-card text-muted-foreground/60 line-through'
                                  )}
                                >
                                  <AgentBadge agent={p.agent} size="sm" />
                                  <span>{p.name}</span>
                                  {isEnabled && (
                                    <span
                                      className={cn(
                                        'text-[10px] px-1 py-0 rounded',
                                        isAuto
                                          ? 'bg-[#183925] text-[#98c379]'
                                          : 'bg-[#142334] text-[#52a8ff]'
                                      )}
                                    >
                                      {isAuto ? 'Auto' : 'Req'}
                                    </span>
                                  )}
                                </Badge>
                              );
                            })
                          )}
                        </div>
                      </div>

                      <div className="flex items-center gap-2 shrink-0">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setActiveTab('my-sharing')}
                          className="h-8 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-muted text-foreground cursor-pointer"
                        >
                          Manage Profiles &amp; Schedules
                        </Button>
                      </div>
                    </div>

                    {/* Warning if Global Sharing is OFF */}
                    {effectiveGlobalMode === 'off' && (
                      <div className="p-3 rounded-geist border border-status-warning-border/60 bg-status-warning-bg/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
                        <div className="flex items-center gap-2">
                          <AlertTriangle className="h-4 w-4 text-status-warning shrink-0" />
                          <span className="text-foreground">
                            Global Sharing is currently <strong>Off</strong>. Teammates cannot discover your node or request leases until sharing is enabled.
                          </span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <Button
                            size="sm"
                            onClick={() => handleGlobalModeChange('request')}
                            className="h-7 text-xs font-sans rounded-geist bg-[#142334] hover:bg-[#1c3048] text-[#52a8ff] border border-[#1f4068] cursor-pointer"
                          >
                            Enable Request Mode
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => handleGlobalModeChange('auto')}
                            className="h-7 text-xs font-sans rounded-geist bg-[#183925] hover:bg-[#204a31] text-[#98c379] border border-[#235836] cursor-pointer"
                          >
                            Enable Auto-Approve
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Empty Teammate Mesh Discovery Card */}
                  <div className="p-10 text-center border border-dashed border-border rounded-geist bg-card/40 space-y-3">
                    <div className="h-10 w-10 rounded-full bg-secondary border border-border flex items-center justify-center mx-auto text-muted-foreground">
                      <Radio className="h-5 w-5 text-accent-cyan animate-pulse" />
                    </div>
                    <div className="space-y-1">
                      <h4 className="text-sm font-semibold text-foreground font-sans">
                        No Teammates Discovered Yet
                      </h4>
                      <p className="text-xs text-muted-foreground max-w-md mx-auto leading-relaxed">
                        AIM actively listens for teammates broadcasting on the internal VPN subnet via mDNS. When a teammate starts sharing an agent profile, they will appear here automatically.
                      </p>
                    </div>
                    <div className="flex items-center justify-center gap-2 pt-1">
                      <Button
                        size="sm"
                        onClick={() => setIsConnectOpen(true)}
                        className="h-8 text-xs font-sans rounded-geist bg-foreground text-background hover:bg-foreground/90 font-medium cursor-pointer"
                      >
                        Connect Teammate Peer Manually
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={isLoading || isRefreshing}
                        onClick={handleManualRefresh}
                        className="h-8 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-muted text-foreground cursor-pointer flex items-center gap-1.5"
                      >
                        <RotateCw className={cn('h-3.5 w-3.5', (isLoading || isRefreshing) && 'animate-spin text-accent-cyan')} />
                        <span>Refresh Mesh</span>
                      </Button>
                    </div>
                  </div>
                </div>
              ) : filteredPeers.length === 0 ? (
                <div className="p-6 text-center border border-dashed border-border rounded-geist bg-card/30 text-xs text-muted-foreground">
                  No teammates match filter &ldquo;{filterQuery}&rdquo;.
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-4">
                  {filteredPeers.map((peer) => {
                    const profileCount = peer.profiles?.length || 0;
                    const isPeerOnline = peer.last_seen
                      ? (currentTime - new Date(peer.last_seen).getTime()) < 5 * 60 * 1000
                      : true;
                    const displayIp = peer.vpn_ip ? peer.vpn_ip.replace(/:\d+$/, '') : '';
                    return (
                      <Card
                        key={peer.id || peer.hostname}
                        className="border border-border bg-card hover:border-[#383838] transition-all rounded-geist flex flex-col justify-between overflow-hidden"
                      >
                        <div>
                          <CardHeader className="p-4 pb-3 border-b border-border/40">
                            <div className="flex items-center justify-between gap-2">
                              <div className="space-y-0.5 min-w-0 flex-1">
                                <CardTitle className="font-semibold text-sm text-foreground tracking-tight flex items-center gap-2">
                                  <Laptop className="h-4 w-4 text-[#61afef] shrink-0" />
                                  <span className="truncate font-sans font-semibold text-sm" title={peer.hostname}>
                                    {peer.hostname}
                                  </span>
                                </CardTitle>
                                {displayIp && displayIp !== peer.hostname && (
                                  <CardDescription className="text-xs font-mono text-muted-foreground truncate pl-6">
                                    {displayIp}
                                  </CardDescription>
                                )}
                              </div>
                              {isPeerOnline ? (
                                <div className="flex items-center gap-1.5 text-xs font-sans text-muted-foreground shrink-0">
                                  <span className="relative flex h-1.5 w-1.5">
                                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#98c379] opacity-75" />
                                    <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-[#98c379]" />
                                  </span>
                                  <span className="text-[#98c379] font-medium text-[11px]">Online</span>
                                </div>
                              ) : (
                                <div
                                  className="flex items-center gap-1.5 text-xs font-sans text-muted-foreground shrink-0"
                                  title={`Last seen: ${peer.last_seen}`}
                                >
                                  <span className="h-1.5 w-1.5 rounded-full bg-[#555555]" />
                                  <span className="text-[11px]">{formatLastSeen(peer.last_seen)}</span>
                                </div>
                              )}
                            </div>
                          </CardHeader>

                          <div className="px-4 py-2 bg-secondary/15 flex items-center justify-between text-[11px] font-mono uppercase text-muted-foreground tracking-wider font-medium border-b border-border/30">
                            <span>Offered Profiles</span>
                            <span>{profileCount} available</span>
                          </div>

                          {peer.profiles && peer.profiles.length > 0 ? (
                            <div className="divide-y divide-border/40">
                              {peer.profiles.map((prof) => {
                                const sp = peer.shared_profiles?.find((p) => p.name === prof);
                                const offeredAgents = (sp?.agents && sp.agents.length > 0) ? sp.agents : [sp?.agent || 'claude'];
                                const agent = offeredAgents[0];
                                const isAuto = sp ? sp.auto_approve : true;
                                return (
                                  <div
                                    key={prof}
                                    className="px-4 py-2.5 hover:bg-secondary/20 transition-colors flex items-center justify-between gap-3"
                                  >
                                    <div className="min-w-0 flex-1 space-y-1">
                                      <div className="flex items-center gap-2 min-w-0">
                                        <span
                                          className="font-mono text-xs font-semibold text-foreground group-hover:text-accent-blue transition-colors truncate select-all"
                                          title={prof}
                                        >
                                          {prof}
                                        </span>
                                      </div>

                                      {/* Clean typographic dot-list: no capsule borders, no awkward slashes */}
                                      <div className="flex items-center gap-2.5 text-xs font-mono text-muted-foreground flex-wrap">
                                        {offeredAgents.map((ag) => (
                                          <span key={ag} className="inline-flex items-center gap-1 text-foreground/75">
                                            <span className={cn('h-1.5 w-1.5 rounded-full shrink-0', getAgentDotColor(ag))} />
                                            <span>{ag}</span>
                                          </span>
                                        ))}
                                      </div>
                                    </div>

                                    <Button
                                      size="sm"
                                      variant="outline"
                                      onClick={() => {
                                        setSelectedPeer(peer);
                                        setSelectedProfile(prof);
                                        setSelectedAgent(agent);
                                        setSelectedAgents(offeredAgents);
                                        setSelectedIsAuto(isAuto);
                                      }}
                                      className={cn(
                                        "h-7 px-2.5 text-xs font-sans font-medium rounded-geist border flex items-center gap-1.5 cursor-pointer shrink-0 shadow-xs transition-colors",
                                        isAuto
                                          ? "border-border bg-secondary/50 hover:bg-secondary hover:border-[#98c379]/40 hover:text-[#98c379] text-foreground/90"
                                          : "border-border bg-secondary/50 hover:bg-secondary hover:border-accent-blue/40 hover:text-accent-blue text-foreground/90"
                                      )}
                                      title={isAuto ? `Instant borrow ${prof} from ${peer.hostname}` : `Request lease for ${prof} from ${peer.hostname}`}
                                    >
                                      {isAuto ? (
                                        <>
                                          <Zap className="h-3 w-3 text-[#98c379]" />
                                          <span>Borrow</span>
                                        </>
                                      ) : (
                                        <>
                                          <KeyRound className="h-3 w-3 text-muted-foreground" />
                                          <span>Request</span>
                                        </>
                                      )}
                                    </Button>
                                  </div>
                                );
                              })}
                            </div>
                          ) : (
                            <div className="p-6 text-xs text-muted-foreground text-center">
                              No profiles currently shared
                            </div>
                          )}
                        </div>
                      </Card>
                    );
                  })}
                </div>
              )}
            </div>
          </section>
        </TabsContent>

        {/* TAB 2: REQUESTS (Executive KPI Summary Board + Focused List View) */}
        <TabsContent value="requests" className="space-y-6 focus-visible:outline-none">
          {/* Top KPI Scorecards & Category Filter Board */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {/* Card 1: Incoming Pending */}
            <div
              onClick={() => setRequestsFilter(requestsFilter === 'pending' ? 'all' : 'pending')}
              className={cn(
                'p-4 rounded-geist border transition-all cursor-pointer relative overflow-hidden group select-none shadow-xs',
                requestsFilter === 'pending'
                  ? 'bg-card border-status-warning ring-1 ring-status-warning/40 shadow-sm'
                  : 'bg-card border-border hover:border-border/80'
              )}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-sans font-medium text-muted-foreground">Incoming Requests</span>
                <div
                  className={cn(
                    'h-7 w-7 rounded-geist flex items-center justify-center transition-colors',
                    pendingRequests.length > 0 ? 'bg-[#142334] text-[#52a8ff] border border-[#1f4068]' : 'bg-secondary text-muted-foreground'
                  )}
                >
                  <Inbox className="h-3.5 w-3.5" />
                </div>
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight text-foreground">
                  {pendingRequests.length}
                </span>
                {pendingRequests.length > 0 ? (
                  <Badge variant="outline" className="text-xs font-mono border-status-warning-border bg-status-warning-bg text-status-warning py-0 px-1.5 animate-pulse">
                    Action required
                  </Badge>
                ) : (
                  <span className="text-xs font-sans text-muted-foreground">All reviewed</span>
                )}
              </div>
              <p className="mt-1 text-xs font-sans text-muted-foreground truncate">
                Leases requested by teammates
              </p>
            </div>

            {/* Card 2: My Outgoing */}
            <div
              onClick={() => setRequestsFilter(requestsFilter === 'outgoing' ? 'all' : 'outgoing')}
              className={cn(
                'p-4 rounded-geist border transition-all cursor-pointer relative overflow-hidden group select-none shadow-xs',
                requestsFilter === 'outgoing'
                  ? 'bg-card border-accent-cyan ring-1 ring-accent-cyan/40 shadow-sm'
                  : 'bg-card border-border hover:border-border/80'
              )}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-sans font-medium text-muted-foreground">My Outgoing</span>
                <div
                  className={cn(
                    'h-7 w-7 rounded-geist flex items-center justify-center transition-colors',
                    outgoingPending.length > 0 ? 'bg-[#142334] text-[#52a8ff] border border-[#1f4068]' : 'bg-secondary text-muted-foreground'
                  )}
                >
                  <Send className="h-3.5 w-3.5" />
                </div>
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight text-foreground">
                  {outgoingPending.length}
                </span>
                {outgoingPending.length > 0 ? (
                  <Badge variant="outline" className="text-xs font-mono border-[#1f4068] bg-[#142334] text-[#52a8ff] py-0 px-1.5">
                    Awaiting host
                  </Badge>
                ) : (
                  <span className="text-xs font-sans text-muted-foreground">None pending</span>
                )}
              </div>
              <p className="mt-1 text-xs font-sans text-muted-foreground truncate">
                Requests sent to peers
              </p>
            </div>

            {/* Card 3: Active Leases */}
            <div
              onClick={() => setRequestsFilter(requestsFilter === 'approved' ? 'all' : 'approved')}
              className={cn(
                'p-4 rounded-geist border transition-all cursor-pointer relative overflow-hidden group select-none shadow-xs',
                requestsFilter === 'approved'
                  ? 'bg-card border-status-success ring-1 ring-status-success/40 shadow-sm'
                  : 'bg-card border-border hover:border-border/80'
              )}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-sans font-medium text-muted-foreground">Active Leases</span>
                <div
                  className={cn(
                    'h-7 w-7 rounded-geist flex items-center justify-center transition-colors',
                    activeGrantedLeases.length + activeBorrowedLeases.length > 0
                      ? 'bg-[#183925] text-[#98c379] border border-[#235836]'
                      : 'bg-secondary text-muted-foreground border border-border'
                  )}
                >
                  <FileCheck className="h-3.5 w-3.5" />
                </div>
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight text-foreground">
                  {activeGrantedLeases.length + activeBorrowedLeases.length}
                </span>
                <span
                  className={cn(
                    'text-xs font-sans font-medium',
                    activeGrantedLeases.length + activeBorrowedLeases.length > 0
                      ? 'text-status-success'
                      : 'text-muted-foreground'
                  )}
                >
                  Active
                </span>
              </div>
              <p className="mt-1 text-xs font-sans text-muted-foreground truncate">
                {activeGrantedLeases.length} lending • {activeBorrowedLeases.length} borrowed
              </p>
            </div>

            {/* Card 4: Past / Expired */}
            <div
              onClick={() => setRequestsFilter(requestsFilter === 'revoked' ? 'all' : 'revoked')}
              className={cn(
                'p-4 rounded-geist border transition-all cursor-pointer relative overflow-hidden group select-none shadow-xs',
                requestsFilter === 'revoked'
                  ? 'bg-card border-muted-foreground ring-1 ring-muted-foreground/40 shadow-sm'
                  : 'bg-card border-border hover:border-border/80'
              )}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-sans font-medium text-muted-foreground">Past / Expired</span>
                <div className="h-7 w-7 rounded-geist bg-secondary text-muted-foreground border border-border flex items-center justify-center">
                  <Ban className="h-3.5 w-3.5" />
                </div>
              </div>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-2xl font-bold font-mono tracking-tight text-foreground">
                  {revokedOrExpiredLeases.length + expiredBorrowedLeases.length}
                </span>
                <span className="text-xs font-sans text-muted-foreground">Historical</span>
              </div>
              <p className="mt-1 text-xs font-sans text-muted-foreground truncate">
                Revoked or timed-out leases
              </p>
            </div>
          </div>

          {/* Sub-toolbar with Active Filter Status & Search Bar */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-1 border-b border-border/40 pb-3">
            <div className="flex items-center gap-2">
              <span className="text-xs font-sans text-muted-foreground">Viewing:</span>
              <div className="flex items-center gap-1.5">
                <Badge
                  variant="outline"
                  className="text-xs font-sans py-0.5 px-2 rounded-geist border-border bg-secondary text-foreground flex items-center gap-1.5"
                >
                  <span className="capitalize">{requestsFilter === 'all' ? 'All Requests & Leases' : requestsFilter === 'pending' ? 'Incoming Pending' : requestsFilter === 'outgoing' ? 'My Outgoing' : requestsFilter === 'approved' ? 'Active Leases' : 'Past / Expired'}</span>
                  {requestsFilter !== 'all' && (
                    <button
                      type="button"
                      onClick={() => setRequestsFilter('all')}
                      className="text-muted-foreground hover:text-foreground cursor-pointer ml-0.5"
                      title="Clear filter"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  )}
                </Badge>
                {requestsFilter !== 'all' && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setRequestsFilter('all')}
                    className="h-6 px-2 text-xs font-sans text-muted-foreground hover:text-foreground cursor-pointer"
                  >
                    Show All
                  </Button>
                )}
              </div>
            </div>

            <div className="relative w-full sm:w-64">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60" />
              <Input
                type="text"
                value={requestSearchQuery}
                onChange={(e) => setRequestSearchQuery(e.target.value)}
                placeholder="Search requests..."
                className="pl-8 pr-7 h-8 text-xs font-sans bg-secondary border-border text-foreground placeholder:text-muted-foreground rounded-geist"
              />
              {requestSearchQuery && (
                <button
                  type="button"
                  onClick={() => setRequestSearchQuery('')}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
                  title="Clear search"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
          </div>

          {/* Empty State when zero total requests exist across everything */}
          {totalRequestsCount === 0 ? (
            <div className="p-12 text-center border border-dashed border-border rounded-geist bg-card/30 space-y-3">
              <div className="h-10 w-10 rounded-full bg-secondary border border-border flex items-center justify-center mx-auto text-muted-foreground">
                <Inbox className="h-5 w-5" />
              </div>
              <div className="space-y-1">
                <h3 className="text-sm font-sans font-semibold text-foreground">No Lease Requests or Active Shares</h3>
                <p className="text-xs font-sans text-muted-foreground max-w-sm mx-auto leading-relaxed">
                  When teammates request access to your shared profiles, or when you borrow credentials across the mesh network, they will appear here.
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setActiveTab('teammates')}
                className="h-8 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-muted text-foreground cursor-pointer"
              >
                Browse Discovered Teammates
              </Button>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Section 1: Incoming Pending Requests */}
              {(requestsFilter === 'all' ? filteredPendingRequests.length > 0 : requestsFilter === 'pending') && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Inbox className="h-4 w-4 text-[#e5c07b]" />
                    <h3 className="text-sm font-sans font-semibold text-foreground">
                      Incoming Requests Pending Approval
                    </h3>
                    <Badge variant="outline" className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground">
                      {filteredPendingRequests.length}
                    </Badge>
                  </div>

                  {filteredPendingRequests.length === 0 ? (
                    <div className="p-8 text-center border border-dashed border-border rounded-geist bg-secondary/30 text-xs font-sans text-muted-foreground">
                      {requestSearchQuery ? 'No incoming requests match search.' : 'No pending incoming requests from teammates.'}
                    </div>
                  ) : (
                    <div className="space-y-2.5">
                      {filteredPendingRequests.map((pr) => (
                        <div
                          key={pr.request_id}
                          className="p-4 rounded-geist border border-border bg-card flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-xs hover:border-border/80 transition-colors"
                        >
                          <div className="flex items-start sm:items-center gap-3 min-w-0">
                            <div className="h-8 w-8 rounded-geist bg-[#142334] border border-[#1f4068] flex items-center justify-center shrink-0">
                              <Inbox className="h-4 w-4 text-[#52a8ff]" />
                            </div>
                            <div className="space-y-1 min-w-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span className="font-mono font-bold text-sm text-foreground">
                                  {pr.request.profile_name}
                                </span>
                                <AgentBadge agent={pr.request.agent} size="sm" />
                                <Badge variant="outline" className="text-[11px] font-sans font-medium px-2 py-0.5 rounded-[4px] border-status-warning-border bg-status-warning-bg text-status-warning">
                                  Awaiting Your Approval
                                </Badge>
                              </div>
                              <div className="flex items-center gap-2 text-xs font-sans text-muted-foreground flex-wrap">
                                <span className="flex items-center gap-1 text-foreground font-medium">
                                  <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
                                  <span className="font-mono">{pr.request.requester_id}</span>
                                </span>
                                <span className="text-muted-foreground/60 font-mono text-[11px]">({pr.request.requester_ip})</span>
                                <span className="text-muted-foreground/40">•</span>
                                <span>Duration: <strong className="text-foreground font-mono">{pr.request.requested_hours >= 24 ? `${pr.request.requested_hours / 24}d` : `${pr.request.requested_hours}h`}</strong></span>
                                <span className="text-muted-foreground/40">•</span>
                                <span className="flex items-center gap-1">
                                  <Clock className="h-3 w-3 text-muted-foreground/70" />
                                  <span className="font-mono">{formatExpiresIn(pr.expires_at)}</span>
                                </span>
                              </div>
                            </div>
                          </div>

                          <div className="flex items-center gap-1.5 shrink-0">
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() => handleDenyRequest(pr.request_id)}
                              className="h-8 w-8 rounded-geist border-border bg-secondary hover:bg-status-danger/10 text-muted-foreground hover:text-status-danger hover:border-status-danger/30 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                              title="Deny Request"
                              aria-label="Deny Request"
                            >
                              <Ban className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              size="icon"
                              onClick={() => handleApproveRequest(pr.request_id)}
                              className="h-8 w-8 rounded-geist bg-[#183925] hover:bg-[#204a31] text-[#98c379] border border-[#235836] flex items-center justify-center shrink-0 cursor-pointer shadow-xs transition-colors"
                              title="Approve Lease"
                              aria-label="Approve Lease"
                            >
                              <Check className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Section 2: My Outgoing Requests */}
              {(requestsFilter === 'all' ? filteredOutgoingRequests.length > 0 : requestsFilter === 'outgoing') && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Send className="h-4 w-4 text-[#52a8ff]" />
                    <h3 className="text-sm font-sans font-semibold text-foreground">
                      My Outgoing Requests
                    </h3>
                    <Badge variant="outline" className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground">
                      {filteredOutgoingRequests.length}
                    </Badge>
                  </div>

                  {filteredOutgoingRequests.length === 0 ? (
                    <div className="p-8 text-center border border-dashed border-border rounded-geist bg-secondary/30 text-xs font-sans text-muted-foreground">
                      {requestSearchQuery ? 'No outgoing requests match search.' : 'No active outgoing requests sent to teammates.'}
                    </div>
                  ) : (
                    <div className="space-y-2.5">
                      {filteredOutgoingRequests.map((req) => (
                        <div
                          key={req.id}
                          className="p-4 rounded-geist border border-border bg-card flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-xs hover:border-border/80 transition-colors"
                        >
                          <div className="flex items-start sm:items-center gap-3 min-w-0">
                            <div className="h-8 w-8 rounded-geist bg-[#142334] border border-[#1f4068] flex items-center justify-center shrink-0">
                              <Send className="h-4 w-4 text-[#52a8ff]" />
                            </div>
                            <div className="space-y-1 min-w-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span className="font-mono font-bold text-sm text-foreground">
                                  {req.profile_name}
                                </span>
                                <AgentBadge agent={req.agent} size="sm" />
                                <Badge variant="outline" className="text-[11px] font-sans font-medium px-2 py-0.5 rounded-[4px] border-[#1f4068] bg-[#142334] text-[#52a8ff] flex items-center gap-1.5">
                                  <span className="h-1.5 w-1.5 rounded-full bg-[#52a8ff] animate-pulse" />
                                  <span>Awaiting Teammate Approval</span>
                                </Badge>
                              </div>
                              <div className="flex items-center gap-2 text-xs font-sans text-muted-foreground flex-wrap">
                                <span className="flex items-center gap-1 text-foreground font-medium">
                                  <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
                                  <span className="font-mono">Host: {req.peer_id}</span>
                                </span>
                                <span className="text-muted-foreground/60 font-mono text-[11px]">({req.peer_ip})</span>
                                <span className="text-muted-foreground/40">•</span>
                                <span>Duration: <strong className="text-foreground font-mono">{req.requested_hours >= 24 ? `${req.requested_hours / 24}d` : `${req.requested_hours}h`}</strong></span>
                                <span className="text-muted-foreground/40">•</span>
                                <span className="flex items-center gap-1">
                                  <Clock className="h-3 w-3 text-muted-foreground/70" />
                                  <span className="font-mono">{formatExpiresIn(req.expires_at)}</span>
                                </span>
                              </div>
                            </div>
                          </div>

                          <div className="flex items-center gap-1.5 shrink-0">
                            <Button
                              variant="outline"
                              size="icon"
                              disabled={pollingReqId === req.id}
                              onClick={() => handleManualPoll(req.id)}
                              className="h-8 w-8 rounded-geist border-border bg-secondary hover:bg-muted text-muted-foreground hover:text-foreground flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                              title={pollingReqId === req.id ? 'Checking status...' : 'Check Status'}
                              aria-label="Check Status"
                            >
                              <RotateCw className={cn('h-3.5 w-3.5', pollingReqId === req.id && 'animate-spin text-accent-cyan')} />
                            </Button>
                            <Button
                              variant="outline"
                              size="icon"
                              disabled={cancellingReqId === req.id}
                              onClick={() => handleCancelRequest(req.id)}
                              className="h-8 w-8 rounded-geist border-border bg-secondary hover:bg-status-danger/10 text-muted-foreground hover:text-status-danger hover:border-status-danger/30 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                              title={cancellingReqId === req.id ? 'Cancelling request...' : 'Cancel Request'}
                              aria-label="Cancel Request"
                            >
                              <X className={cn('h-3.5 w-3.5', cancellingReqId === req.id && 'animate-pulse text-status-danger')} />
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Section 3: Active Approved Leases */}
              {(requestsFilter === 'all' ? (filteredActiveGranted.length + filteredActiveBorrowed.length) > 0 : requestsFilter === 'approved') && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <FileCheck
                      className={cn(
                        'h-4 w-4',
                        filteredActiveGranted.length + filteredActiveBorrowed.length > 0
                          ? 'text-[#98c379]'
                          : 'text-muted-foreground'
                      )}
                    />
                    <h3 className="text-sm font-sans font-semibold text-foreground">
                      Active Approved Leases
                    </h3>
                    <Badge variant="outline" className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground">
                      {filteredActiveGranted.length + filteredActiveBorrowed.length}
                    </Badge>
                  </div>

                  {filteredActiveGranted.length === 0 && filteredActiveBorrowed.length === 0 ? (
                    <div className="p-8 text-center border border-dashed border-border rounded-geist bg-secondary/30 text-xs font-sans text-muted-foreground">
                      {requestSearchQuery ? 'No active leases match search.' : 'No active approved leases.'}
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {/* Granted to teammates */}
                      {filteredActiveGranted.map((lease) => (
                        <div
                          key={lease.lease_id}
                          className="p-4 rounded-geist border border-border bg-card flex flex-col justify-between gap-3 shadow-xs hover:border-border/80 transition-colors"
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
                                    'text-[11px] font-sans font-medium px-2 py-0.5 rounded-[4px] border',
                                    lease.auto_approved
                                      ? 'bg-[#183925]/50 text-[#98c379] border-[#235836]/60'
                                      : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                                  )}
                                >
                                  {lease.auto_approved ? 'Auto-Approved' : 'Manual Approved'}
                                </span>
                              </div>
                              <span className="text-xs font-sans px-2 py-0.5 rounded-[4px] font-medium bg-status-success-bg text-status-success border border-status-success-border flex items-center gap-1">
                                <Clock className="h-3 w-3" />
                                <span className="font-mono tabular-nums">{formatExpiresIn(lease.expires_at)}</span>
                              </span>
                            </div>
                            <div className="text-xs text-muted-foreground flex items-center gap-1.5 font-sans">
                              <Laptop className="h-3.5 w-3.5 text-[#61afef]" />
                              <span>Borrower:</span>
                              <span className="text-foreground font-mono font-medium" title={lease.borrower_id}>
                                {lease.borrower_id}
                              </span>
                            </div>
                          </div>

                          <div className="flex items-center justify-between pt-2.5 border-t border-border/50 text-xs font-sans">
                            <span className="text-muted-foreground font-sans">
                              Lease ID: <span className="font-mono text-muted-foreground/80">{lease.lease_id}</span>
                            </span>
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() => setLeaseToRevoke({
                                leaseId: lease.lease_id,
                                profileName: lease.profile_name,
                                borrowerId: lease.borrower_id,
                              })}
                              className="h-8 w-8 rounded-geist border-border bg-secondary hover:bg-status-danger/10 text-muted-foreground hover:text-status-danger hover:border-status-danger/30 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                              title="Revoke Lease"
                              aria-label="Revoke Lease"
                            >
                              <ShieldAlert className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </div>
                      ))}

                      {/* Borrowed by you */}
                      {filteredActiveBorrowed.map((lease) => (
                        <BorrowedLeaseCard
                          key={lease.local_profile}
                          lease={lease}
                          currentTime={currentTime}
                          formatExpiresIn={formatExpiresIn}
                          copiedCmd={copiedCmd}
                          onCopy={copyToClipboard}
                          onRelease={(prof) => setLeaseToRelease(prof)}
                        />
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Section 4: Revoked / Expired Leases */}
              {(requestsFilter === 'all' ? (filteredRevokedOrExpired.length + filteredExpiredBorrowed.length) > 0 : requestsFilter === 'revoked') && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Ban className="h-4 w-4 text-muted-foreground" />
                      <h3 className="text-sm font-sans font-semibold text-foreground">
                        Revoked & Expired Leases
                      </h3>
                      <Badge variant="outline" className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground">
                        {filteredRevokedOrExpired.length + filteredExpiredBorrowed.length}
                      </Badge>
                    </div>
                    {filteredExpiredBorrowed.length > 0 && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleCleanExpired}
                        className="h-7 px-2.5 text-xs font-mono border-border bg-secondary hover:bg-muted text-muted-foreground hover:text-foreground cursor-pointer flex items-center gap-1.5"
                      >
                        <Trash2 className="h-3 w-3 text-status-danger" />
                        <span>Clean Expired ({filteredExpiredBorrowed.length})</span>
                      </Button>
                    )}
                  </div>

                  {filteredRevokedOrExpired.length === 0 && filteredExpiredBorrowed.length === 0 ? (
                    <div className="p-8 text-center border border-dashed border-border rounded-geist bg-secondary/30 text-xs font-sans text-muted-foreground">
                      {requestSearchQuery ? 'No expired leases match search.' : 'No revoked or expired leases.'}
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {/* Expired Leases Borrowed from Peers */}
                      {filteredExpiredBorrowed.map((lease) => (
                        <div
                          key={lease.local_profile}
                          className="p-3.5 rounded-geist border border-border bg-card flex items-center justify-between gap-3 text-xs shadow-xs"
                        >
                          <div className="flex items-center gap-2.5 min-w-0">
                            <div className="min-w-0 space-y-0.5">
                              <div className="flex items-center gap-2">
                                <span className="font-mono font-semibold text-foreground truncate">{lease.local_profile}</span>
                                <AgentBadge agent={lease.agent} size="sm" />
                              </div>
                              <div className="text-muted-foreground text-xs font-sans flex items-center gap-1">
                                <span>from:</span>
                                <span className="font-mono text-foreground/80">{lease.source_peer}/{lease.source_profile}</span>
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-2 shrink-0">
                            <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-sans">
                              <Clock className="h-3 w-3 text-muted-foreground/70" />
                              <span>Expired {formatLastSeen(lease.expires_at)}</span>
                            </div>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleRelease(lease.local_profile)}
                              className="h-7 px-2 text-xs font-sans text-muted-foreground hover:text-foreground cursor-pointer"
                              title="Purge expired lease credentials"
                            >
                              <Trash2 className="h-3.5 w-3.5 mr-1 text-status-danger" />
                              Purge
                            </Button>
                          </div>
                        </div>
                      ))}

                      {/* Leases Granted to Teammates that are Revoked / Expired */}
                      {filteredRevokedOrExpired.map((lease) => (
                        <div
                          key={lease.lease_id}
                          className="p-3.5 rounded-geist border border-border bg-card flex items-center justify-between gap-3 text-xs shadow-xs"
                        >
                          <div className="flex items-center gap-2.5 min-w-0">
                            <div className="min-w-0 space-y-0.5">
                              <div className="flex items-center gap-2">
                                <span className="font-mono font-semibold text-foreground truncate">{lease.profile_name}</span>
                                <AgentBadge agent={lease.agent} size="sm" />
                              </div>
                              <div className="text-muted-foreground text-xs font-sans flex items-center gap-1.5">
                                <span>Borrower:</span>
                                <span className="font-mono text-foreground/80" title={lease.borrower_id}>{lease.borrower_id}</span>
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-sans shrink-0">
                            <Clock className="h-3 w-3 text-muted-foreground/70" />
                            <span>{lease.revoked ? 'Revoked' : 'Expired'} {formatLastSeen(lease.expires_at)}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
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
            <Card className="border-border bg-card rounded-geist">
              <CardHeader className="p-5 pb-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <Share2 className="h-4 w-4 text-accent-cyan" />
                    <CardTitle className="text-sm font-semibold text-foreground">
                      Active Borrowers
                    </CardTitle>
                    <Badge
                      variant="outline"
                      className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground"
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
                          <span className="text-foreground font-semibold" title={lease.borrower_id}>
                            {lease.borrower_id}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                        <span className="text-xs font-mono text-[#666666]">
                          Lease ID: {lease.lease_id}
                        </span>
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => setLeaseToRevoke({
                            leaseId: lease.lease_id,
                            profileName: lease.profile_name,
                            borrowerId: lease.borrower_id,
                          })}
                          className="h-8 w-8 rounded-geist border-border bg-secondary hover:bg-status-danger/10 text-muted-foreground hover:text-status-danger hover:border-status-danger/30 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                          title="Revoke Lease"
                          aria-label="Revoke Lease"
                        >
                          <ShieldAlert className="h-3.5 w-3.5" />
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
                    className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground"
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

                      {/* Sharing Mode Dropdown */}
                      <div className="w-36 shrink-0">
                        <SelectDropdown
                          value={mode}
                          onChange={(val) => handleSetLocalProfileShareMode(p.name, val as ShareMode)}
                          disabled={isShareUpdating}
                          size="sm"
                          options={[
                            {
                              value: 'off',
                              label: 'Private',
                              icon: <Lock className="h-3.5 w-3.5 text-muted-foreground" />,
                              description: 'Hidden from peers',
                            },
                            {
                              value: 'request',
                              label: 'Request',
                              icon: <Clock className="h-3.5 w-3.5 text-[#52a8ff]" />,
                              description: 'Requires approval',
                            },
                            {
                              value: 'auto',
                              label: 'Auto',
                              icon: <Zap className="h-3.5 w-3.5 text-[#98c379]" />,
                              description: 'Instant auto-approve',
                            },
                          ]}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          {/* Scheduled Sharing & Automation Card (Housed directly in My Sharing) */}
          <Card className="border border-border bg-card rounded-geist shadow-xs">
            <CardHeader className="p-5 pb-4 border-b border-border/50">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <CalendarIcon className="h-4 w-4 text-accent-cyan" />
                  <CardTitle className="text-sm font-semibold text-foreground">
                    Automation & Scheduled Windows
                  </CardTitle>
                  {activeSchedulesCount > 0 && (
                    <Badge
                      variant="outline"
                      className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-[#183925] text-[#98c379]"
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
            <CardContent className="p-5">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 items-start">
                {/* Left Column: Schedule Builder Form (50% / 6 cols) */}
                <form onSubmit={handleSaveSchedulePolicy} className="space-y-4 p-4 rounded-geist bg-secondary/30 border border-border/60 shadow-xs">
                  <div className="flex items-center justify-between pb-1 border-b border-border/40">
                    <span className="text-xs font-mono uppercase tracking-wider text-foreground font-semibold flex items-center gap-1.5">
                      <Zap className="h-3.5 w-3.5 text-accent-cyan" />
                      <span>New Schedule Rule</span>
                    </span>
                    <span className="text-[11px] font-sans text-muted-foreground">Applies automatically</span>
                  </div>

                  {/* Scope & Policy Mode */}
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div className="space-y-1.5">
                      <label className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground font-medium">
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
                      <label className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground font-medium">
                        Policy Mode in Window
                      </label>
                      <SelectDropdown
                        value={schedMode}
                        onChange={(val) => setSchedMode(val as 'auto' | 'request')}
                        size="sm"
                        options={[
                          {
                            value: 'auto',
                            label: 'Auto Approve (Instant)',
                            icon: <Zap className="h-3.5 w-3.5 text-[#98c379]" />,
                          },
                          {
                            value: 'request',
                            label: 'Request Mode (Manual)',
                            icon: <Clock className="h-3.5 w-3.5 text-[#52a8ff]" />,
                          },
                        ]}
                      />
                    </div>
                  </div>

                  {/* Schedule Type Selection — Compact SegmentedControl */}
                  <div className="space-y-2">
                    <label className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground font-medium block">
                      Schedule Window Type
                    </label>
                    <SegmentedControl
                      size="sm"
                      value={schedType}
                      onChange={(val) => setSchedType(val as 'recurring' | 'calendar')}
                      options={[
                        {
                          value: 'recurring',
                          label: 'Daily Recurring Hours',
                          icon: <Clock className="h-3 w-3 text-accent-cyan" />,
                        },
                        {
                          value: 'calendar',
                          label: 'Calendar Date Range',
                          icon: <CalendarIcon className="h-3 w-3 text-accent-cyan" />,
                        },
                      ]}
                    />
                  </div>

                  {/* Recurring Time Content */}
                  {schedType === 'recurring' ? (
                    <div className="space-y-3 p-3 rounded-geist bg-card border border-border/80">
                      <div className="flex items-center justify-between text-[11px] font-mono text-muted-foreground flex-wrap gap-1">
                        <span className="font-medium text-foreground">Quick Presets:</span>
                        <div className="flex items-center gap-1.5 flex-wrap">
                          <button
                            type="button"
                            onClick={() => { setSchedStartTime('00:00'); setSchedEndTime('08:00'); }}
                            className={cn(
                              "px-2 py-0.5 rounded text-[11px] font-mono transition-colors cursor-pointer border",
                              schedStartTime === '00:00' && schedEndTime === '08:00'
                                ? "bg-secondary text-accent-cyan border-accent-cyan/40 font-semibold"
                                : "bg-secondary/60 hover:bg-secondary text-muted-foreground hover:text-foreground border-border/50"
                            )}
                          >
                            12am - 8am
                          </button>
                          <button
                            type="button"
                            onClick={() => { setSchedStartTime('22:00'); setSchedEndTime('06:00'); }}
                            className={cn(
                              "px-2 py-0.5 rounded text-[11px] font-mono transition-colors cursor-pointer border",
                              schedStartTime === '22:00' && schedEndTime === '06:00'
                                ? "bg-secondary text-accent-cyan border-accent-cyan/40 font-semibold"
                                : "bg-secondary/60 hover:bg-secondary text-muted-foreground hover:text-foreground border-border/50"
                            )}
                          >
                            10pm - 6am
                          </button>
                          <button
                            type="button"
                            onClick={() => { setSchedStartTime('09:00'); setSchedEndTime('18:00'); }}
                            className={cn(
                              "px-2 py-0.5 rounded text-[11px] font-mono transition-colors cursor-pointer border",
                              schedStartTime === '09:00' && schedEndTime === '18:00'
                                ? "bg-secondary text-accent-cyan border-accent-cyan/40 font-semibold"
                                : "bg-secondary/60 hover:bg-secondary text-muted-foreground hover:text-foreground border-border/50"
                            )}
                          >
                            9am - 6pm
                          </button>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-[11px] font-mono text-muted-foreground">From</label>
                          <SelectDropdown
                            value={schedStartTime}
                            onChange={setSchedStartTime}
                            options={getTimeOptions(schedStartTime)}
                            size="sm"
                            className="w-full"
                            triggerClassName="h-8 font-mono text-xs border-border bg-secondary hover:bg-muted text-foreground"
                          />
                        </div>
                        <div className="space-y-1">
                          <label className="text-[11px] font-mono text-muted-foreground">To</label>
                          <SelectDropdown
                            value={schedEndTime}
                            onChange={setSchedEndTime}
                            options={getTimeOptions(schedEndTime)}
                            size="sm"
                            className="w-full"
                            triggerClassName="h-8 font-mono text-xs border-border bg-secondary hover:bg-muted text-foreground"
                          />
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-3 p-3 rounded-geist bg-card border border-border/80">
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-[11px] font-mono text-muted-foreground">Start Date & Time</label>
                          <div className="grid grid-cols-2 gap-2">
                            <ScheduleDatePicker
                              value={schedStartDate}
                              onChange={setSchedStartDate}
                              placeholder="Start date"
                            />
                            <SelectDropdown
                              value={schedStartTime}
                              onChange={setSchedStartTime}
                              options={getTimeOptions(schedStartTime)}
                              size="sm"
                              className="w-full"
                              triggerClassName="h-8 font-mono text-xs border-border bg-secondary hover:bg-muted text-foreground"
                            />
                          </div>
                        </div>
                        <div className="space-y-1">
                          <label className="text-[11px] font-mono text-muted-foreground">End Date & Time</label>
                          <div className="grid grid-cols-2 gap-2">
                            <ScheduleDatePicker
                              value={schedEndDate}
                              onChange={setSchedEndDate}
                              placeholder="End date"
                            />
                            <SelectDropdown
                              value={schedEndTime}
                              onChange={setSchedEndTime}
                              options={getTimeOptions(schedEndTime)}
                              size="sm"
                              className="w-full"
                              triggerClassName="h-8 font-mono text-xs border-border bg-secondary hover:bg-muted text-foreground"
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
                      className="h-8 px-3 rounded-geist bg-primary text-primary-foreground hover:bg-primary/90 text-xs font-sans font-medium flex items-center gap-1.5 cursor-pointer shadow-xs"
                    >
                      <Check className="h-3.5 w-3.5" />
                      <span>{isSavingSchedule ? 'Saving...' : 'Add Schedule Rule'}</span>
                    </Button>
                  </div>
                </form>

                {/* Right Column: Configured Rules List & Overview (50% / 6 cols) */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between pb-1">
                    <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground font-medium flex items-center gap-1.5">
                      <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                      <span>Configured Policies ({data?.schedules?.length || 0})</span>
                    </span>
                    {activeSchedulesCount > 0 && (
                      <span className="text-[11px] font-mono text-[#98c379] flex items-center gap-1">
                        <span className="h-1.5 w-1.5 rounded-full bg-[#98c379] animate-pulse" />
                        Active Now
                      </span>
                    )}
                  </div>

                  {data?.schedules && data.schedules.length > 0 ? (
                    <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
                      {data.schedules.map((sc) => {
                        const isActive = isScheduleActiveNow(sc);
                        return (
                          <div
                            key={sc.id}
                            className="p-3 rounded-geist bg-secondary/40 border border-border flex items-center justify-between gap-3 text-xs font-mono shadow-xs hover:border-[#383838] transition-colors"
                          >
                            <div className="space-y-1 min-w-0 flex-1">
                              <div className="flex items-center gap-2">
                                <span
                                  className={cn(
                                    'h-2 w-2 rounded-full shrink-0',
                                    isActive ? 'bg-[#10b981] animate-pulse' : 'bg-[#666666]'
                                  )}
                                  title={isActive ? 'Active Now' : 'Scheduled'}
                                />
                                <span className="font-bold text-foreground truncate">
                                  {sc.scope === 'global' ? 'Global Policy' : sc.scope}
                                </span>
                                <span
                                  className={cn(
                                    'text-[10px] uppercase font-bold px-1.5 py-0.5 rounded border',
                                    sc.mode === 'auto'
                                      ? 'bg-[#183925] text-[#98c379] border-[#235836]'
                                      : 'bg-[#142334] text-[#52a8ff] border-[#1f4068]'
                                  )}
                                >
                                  {sc.mode === 'auto' ? 'Auto' : 'Request'}
                                </span>
                              </div>
                              <div className="text-[11px] text-muted-foreground truncate pl-4">
                                {sc.start_date ? `${sc.start_date} ` : 'Daily '}{sc.start_time} – {sc.end_date ? `${sc.end_date} ` : ''}{sc.end_time}
                              </div>
                            </div>

                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDeleteSchedulePolicy(sc.id)}
                              className="h-7 w-7 rounded-geist text-muted-foreground hover:text-status-danger hover:bg-status-danger/10 flex items-center justify-center shrink-0 cursor-pointer transition-colors"
                              title="Delete schedule"
                              aria-label="Delete schedule"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <div className="p-6 text-center border border-dashed border-border/80 rounded-geist bg-secondary/20 space-y-2">
                      <div className="h-8 w-8 rounded-full bg-secondary flex items-center justify-center mx-auto text-muted-foreground border border-border/60">
                        <Clock className="h-4 w-4" />
                      </div>
                      <div className="space-y-1">
                        <p className="text-xs font-medium text-foreground font-sans">No Active Automation Rules</p>
                        <p className="text-[11px] text-muted-foreground font-sans max-w-xs mx-auto leading-relaxed">
                          Define off-peak hours (e.g. nights or weekends) to automatically share profiles or auto-approve teammate requests.
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
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
                    className="text-xs font-mono py-0.5 px-1.5 rounded-[4px] border-border/70 bg-secondary/60 text-muted-foreground"
                  >
                    {filteredAuditLogs.length === (data?.audit_logs?.length || 0)
                      ? `${Math.min(visibleAuditCount, filteredAuditLogs.length)} of ${filteredAuditLogs.length} events`
                      : `${Math.min(visibleAuditCount, filteredAuditLogs.length)} of ${filteredAuditLogs.length} filtered (${data?.audit_logs?.length || 0} total)`}
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
                        { value: 'granted', label: 'Leases Granted / Borrowed' },
                        { value: 'requested', label: 'Lease Requests' },
                        { value: 'cancelled', label: 'Leases Cancelled' },
                        { value: 'revoked', label: 'Leases Revoked / Denied' },
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
                  {paginatedAuditLogs.map((log) => {
                    const actionColors: Record<string, string> = {
                      LEASE_GRANTED: 'text-[#98c379] bg-[#183925]/40 border-[#235836]',
                      LEASE_APPROVED: 'text-[#98c379] bg-[#183925]/40 border-[#235836]',
                      LEASE_BORROWED: 'text-[#56b6c2] bg-[#1a2d33]/40 border-[#264e59]',
                      LEASE_REQUESTED: 'text-[#52a8ff] bg-[#142334]/40 border-[#1f4068]',
                      LEASE_CANCELLED: 'text-[#e5c07b] bg-[#332b1a]/40 border-[#594926]',
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

                  {/* Infinite Scroll Sentinel / Status */}
                  <div ref={auditSentinelRef} className="pt-2 pb-1 text-center">
                    {visibleAuditCount < filteredAuditLogs.length ? (
                      <div className="flex items-center justify-center gap-2 text-xs font-mono text-muted-foreground">
                        <RotateCw className="h-3 w-3 animate-spin text-accent-cyan" />
                        <span>Loading more events ({Math.min(visibleAuditCount, filteredAuditLogs.length)} of {filteredAuditLogs.length})…</span>
                      </div>
                    ) : filteredAuditLogs.length > 50 ? (
                      <span className="text-[11px] font-mono text-muted-foreground/60">
                        All {filteredAuditLogs.length} events loaded
                      </span>
                    ) : null}
                  </div>
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
                  <div className="flex items-center gap-1">
                    {selectedAgents.length > 0 ? (
                      selectedAgents.map((ag) => (
                        <AgentBadge key={ag} agent={ag} size="sm" />
                      ))
                    ) : (
                      <AgentBadge agent={selectedAgent} size="sm" />
                    )}
                  </div>
                </div>
              </div>
              {selectedAgents.length > 1 && (
                <div className="flex items-center justify-between text-[#888888]">
                  <span>Preferred Engine</span>
                  <SelectDropdown
                    value={selectedAgent}
                    onChange={(val) => setSelectedAgent(val as string)}
                    size="sm"
                    className="w-36 font-mono text-xs"
                    options={selectedAgents.map((ag) => ({
                      value: ag,
                      label: ag,
                    }))}
                  />
                </div>
              )}
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
                  { value: 24, label: '24 hours (1 Day)' },
                  { value: 120, label: '5 days' },
                  { value: 720, label: '30 days' },
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

      {/* Revoke Lease Confirmation Dialog Modal */}
      <Dialog
        open={!!leaseToRevoke}
        onOpenChange={(open) => {
          if (!open && !isRevoking) {
            setLeaseToRevoke(null);
          }
        }}
      >
        <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
          <DialogHeader className="space-y-1 pb-3 border-b border-[#222222] pr-6">
            <div className="flex items-center gap-2">
              <div className="h-8 w-8 rounded-full bg-status-danger-bg border border-status-danger-border flex items-center justify-center shrink-0">
                <ShieldAlert className="h-4 w-4 text-status-danger" />
              </div>
              <DialogTitle className="text-base font-semibold text-[#ededed] font-sans">
                Revoke Profile Lease
              </DialogTitle>
            </div>
            <DialogDescription className="text-xs text-[#888888] pt-1.5 font-sans">
              Are you sure you want to revoke this lease? Teammate credentials and sandbox access will be disabled immediately.
            </DialogDescription>
          </DialogHeader>

          {leaseToRevoke && (
            <div className="py-2">
              <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-2">
                <div className="flex items-center justify-between text-[#888888]">
                  <span>Profile</span>
                  <span className="font-mono text-[#ededed] font-medium text-accent-cyan">
                    {leaseToRevoke.profileName}
                  </span>
                </div>
                {leaseToRevoke.borrowerId && (
                  <div className="flex items-center justify-between text-[#888888]">
                    <span>Borrower</span>
                    <span className="font-mono text-[#ededed] font-medium" title={leaseToRevoke.borrowerId}>
                      {leaseToRevoke.borrowerId}
                    </span>
                  </div>
                )}
                <div className="pt-2 border-t border-[#1f1f1f] flex items-center justify-between text-[#888888]">
                  <span>Lease ID</span>
                  <span className="font-mono text-[#a1a1a1]">
                    {leaseToRevoke.leaseId}
                  </span>
                </div>
              </div>
            </div>
          )}

          <DialogFooter className="pt-2 border-t border-[#222222] flex flex-col-reverse sm:flex-row gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isRevoking}
              onClick={() => setLeaseToRevoke(null)}
              className="rounded-geist text-xs border-border"
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              disabled={isRevoking}
              onClick={async () => {
                if (!leaseToRevoke) return;
                setIsRevoking(true);
                try {
                  await handleRevokeLease(leaseToRevoke.leaseId);
                  setLeaseToRevoke(null);
                } finally {
                  setIsRevoking(false);
                }
              }}
              className="rounded-geist text-xs font-medium cursor-pointer"
            >
              <ShieldAlert className="h-3.5 w-3.5 mr-1.5" />
              <span>{isRevoking ? 'Revoking...' : 'Revoke Lease'}</span>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Release Lease Confirmation Dialog Modal */}
      <Dialog
        open={!!leaseToRelease}
        onOpenChange={(open) => {
          if (!open && !isReleasing) {
            setLeaseToRelease(null);
          }
        }}
      >
        <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
          <DialogHeader className="space-y-1 pb-3 border-b border-[#222222] pr-6">
            <div className="flex items-center gap-2">
              <div className="h-8 w-8 rounded-full bg-status-danger-bg border border-status-danger-border flex items-center justify-center shrink-0">
                <Trash2 className="h-4 w-4 text-status-danger" />
              </div>
              <DialogTitle className="text-base font-semibold text-[#ededed] font-sans">
                Release Leased Profile
              </DialogTitle>
            </div>
            <DialogDescription className="text-xs text-[#888888] pt-1.5 font-sans">
              Are you sure you want to release this leased profile? Local credentials and sandbox access will be removed immediately.
            </DialogDescription>
          </DialogHeader>

          {leaseToRelease && (
            <div className="py-2">
              <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-2">
                <div className="flex items-center justify-between text-[#888888]">
                  <span>Local Profile</span>
                  <span className="font-mono text-accent-cyan font-medium">
                    {leaseToRelease}
                  </span>
                </div>
              </div>
            </div>
          )}

          <DialogFooter className="pt-2 border-t border-[#222222] flex flex-col-reverse sm:flex-row gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isReleasing}
              onClick={() => setLeaseToRelease(null)}
              className="rounded-geist text-xs border-border"
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              disabled={isReleasing}
              onClick={async () => {
                if (!leaseToRelease) return;
                setIsReleasing(true);
                try {
                  await handleRelease(leaseToRelease);
                  setLeaseToRelease(null);
                } finally {
                  setIsReleasing(false);
                }
              }}
              className="rounded-geist text-xs font-medium cursor-pointer"
            >
              <Trash2 className="h-3.5 w-3.5 mr-1.5" />
              <span>{isReleasing ? 'Releasing...' : 'Release Lease'}</span>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

    </div>
  );
}
