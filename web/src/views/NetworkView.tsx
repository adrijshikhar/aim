import * as React from 'react';
import {
  Wifi,
  Radio,
  Clock,
  Laptop,
  RotateCw,
  Trash2,
  Copy,
  Check,
  Search,
  AlertTriangle,
  Plus,
  X,
  Share2,
  ShieldAlert,
  Activity,
} from 'lucide-react';
import {
  getShareStatus,
  setGlobalShareMode,
  toggleShareProfile,
  revokeShareLease,
  addSharePeer,
  requestShareLease,
  releaseShareLease,
  ShareStatusDTO,
  PeerInfoDTO,
  ProfileDTO,
} from '@/lib/api';
import { useToast } from '@/components/ui/use-toast';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { AgentBadge } from '@/components/AgentBadge';
import { SelectDropdown } from '@/components/ui/select-dropdown';
import { Slider } from '@/components/ui/slider';
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

  // Global Sharing Slider state
  const [isShareUpdating, setIsShareUpdating] = React.useState(false);

  // Local Profile Override state (in VPN banner)
  const [selectedLocalProfile, setSelectedLocalProfile] = React.useState<string>('');

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
    loadStatus(false);
    const interval = setInterval(() => loadStatus(false), 5000);
    return () => clearInterval(interval);
  }, [loadStatus]);

  React.useEffect(() => {
    if (!selectedLocalProfile && profiles.length > 0) {
      setSelectedLocalProfile(profiles[0].name);
    }
  }, [profiles, selectedLocalProfile]);

  const currentLocalSharedConfig = React.useMemo(() => {
    if (!data?.shared_profiles || !selectedLocalProfile) return null;
    return data.shared_profiles.find((p) => p.name === selectedLocalProfile) || null;
  }, [data?.shared_profiles, selectedLocalProfile]);

  const currentLocalShareMode: ShareMode = React.useMemo(() => {
    if (!currentLocalSharedConfig || !currentLocalSharedConfig.enabled) return 'off';
    if (currentLocalSharedConfig.auto_approve) return 'auto';
    return 'request';
  }, [currentLocalSharedConfig]);

  const handleSetLocalProfileShareMode = async (profileName: string, mode: ShareMode) => {
    if (!profileName) return;
    setIsShareUpdating(true);
    const enabled = mode !== 'off';
    const auto_approve = mode === 'auto';
    const profObj = profiles.find((p) => p.name === profileName);
    const agent = profObj?.agent || 'claude';
    try {
      await toggleShareProfile({
        name: profileName,
        agent,
        auto_approve,
        ttl: '4h',
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

  const currentGlobalMode: ShareMode = data?.global_mode || 'off';

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
        toast({
          title: 'Lease Granted',
          description: `Installed as ${res.local_profile}. Run with: aim run ${selectedAgent} ${res.local_profile}`,
          variant: 'success',
        });
        setSelectedPeer(null);
        setSelectedProfile('');
        await loadStatus();
      } else {
        toast({
          title: 'Request Pending',
          description: res.message || 'Waiting for approval from host',
        });
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
    const diff = new Date(expiresAtStr).getTime() - Date.now();
    if (diff <= 0) return 'Expired';
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const mins = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    return `${hours}h ${mins}m remaining`;
  };

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
              className="font-mono text-[10px] uppercase tracking-wider py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
            >
              Mesh
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl leading-relaxed">
            Borrow high-tier Claude & AI assistant quotas from teammates on the internal VPN (Pritunl / OpenVPN) or local Wi-Fi. Borrowed profiles are isolated locally as{' '}
            <code className="text-[#56b6c2] font-mono px-1.5 py-0.5 rounded bg-[#111111] border border-[#262626] text-[11px]">
              host:&lt;profile_name&gt;
            </code>
            .
          </p>
        </div>

        {/* TOP RIGHT: Global Sharing Slider (Screenshot 12.28.47 AM) */}
        <div className="rounded-geist bg-[#111111] border border-[#262626] p-2.5 px-3 shadow-md flex flex-col items-center justify-center w-[250px] shrink-0 select-none">
          <div className="flex items-center justify-between w-full mb-1">
            <span className="font-mono text-[10px] uppercase tracking-wider text-[#777777] flex items-center gap-1">
              <Radio className="h-3 w-3 text-accent-cyan" />
              <span>Mesh Policy</span>
            </span>
            <span
              className={cn(
                'text-xs font-mono font-bold tracking-tight transition-colors',
                currentGlobalMode === 'off' && 'text-[#888888]',
                currentGlobalMode === 'request' && 'text-[#52a8ff]',
                currentGlobalMode === 'auto' && 'text-[#46a758]'
              )}
            >
              {currentGlobalMode === 'off' && 'Off'}
              {currentGlobalMode === 'request' && 'Request'}
              {currentGlobalMode === 'auto' && 'Auto Approve'}
            </span>
          </div>

          {/* Shadcn / Radix Draggable 3-Stop Slider */}
          <div className="w-full px-1 py-1.5">
            <Slider
              min={0}
              max={2}
              step={1}
              value={[currentGlobalMode === 'off' ? 0 : currentGlobalMode === 'request' ? 1 : 2]}
              onValueChange={([val]) => {
                const modes: ShareMode[] = ['off', 'request', 'auto'];
                const target = modes[val];
                if (target && target !== currentGlobalMode) {
                  handleGlobalModeChange(target);
                }
              }}
              trackClassName="h-4 bg-[#181818] border border-[#2a2a2a] rounded-full"
              rangeClassName={cn(
                'transition-all duration-300',
                currentGlobalMode === 'off' && 'bg-transparent',
                currentGlobalMode === 'request' && 'bg-[#2563eb]/60',
                currentGlobalMode === 'auto' && 'bg-[#16a34a]/60'
              )}
              thumbClassName="h-6 w-6 rounded-full bg-white shadow-md border-0 ring-1 ring-black/20 cursor-grab active:cursor-grabbing hover:scale-105 active:scale-95 transition-transform"
              thumbChildren={
                <span
                  className={cn(
                    'h-2.5 w-2.5 rounded-full transition-colors',
                    currentGlobalMode === 'off' && 'bg-[#666666]',
                    currentGlobalMode === 'request' && 'bg-[#0070f3]',
                    currentGlobalMode === 'auto' && 'bg-[#10b981]'
                  )}
                />
              }
            />
          </div>

          {/* 3 Stop Interactive Labels */}
          <div className="flex items-center justify-between w-full mt-1 px-1 text-[10px] font-mono">
            <button
              type="button"
              onClick={() => handleGlobalModeChange('off')}
              className={cn(
                'cursor-pointer transition-colors',
                currentGlobalMode === 'off'
                  ? 'text-[#ededed] font-semibold'
                  : 'text-[#666666] hover:text-[#aaaaaa]'
              )}
            >
              Off
            </button>
            <button
              type="button"
              onClick={() => handleGlobalModeChange('request')}
              className={cn(
                'cursor-pointer transition-colors',
                currentGlobalMode === 'request'
                  ? 'text-[#52a8ff] font-semibold'
                  : 'text-[#666666] hover:text-[#aaaaaa]'
              )}
            >
              Request
            </button>
            <button
              type="button"
              onClick={() => handleGlobalModeChange('auto')}
              className={cn(
                'cursor-pointer transition-colors',
                currentGlobalMode === 'auto'
                  ? 'text-[#46a758] font-semibold'
                  : 'text-[#666666] hover:text-[#aaaaaa]'
              )}
            >
              Auto Approve
            </button>
          </div>
        </div>
      </div>

      {/* VPN Tunnel Status Strip (Consistent with AIM Card & Badge Standards) */}
      <div className="p-4 rounded-geist border border-[#222222] bg-[#0a0a0a] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center h-8 w-8 rounded-full border border-border bg-secondary shrink-0">
            {isVpnActive ? (
              <Wifi className="h-4 w-4 text-status-success" />
            ) : (
              <AlertTriangle className="h-4 w-4 text-status-warning" />
            )}
          </div>
          <div className="space-y-0.5">
            <div className="flex items-center gap-2">
              <span className="font-semibold text-xs text-foreground tracking-tight">
                {isVpnActive ? 'Corporate VPN Tunnel Active' : 'VPN Tunnel Offline'}
              </span>
              {isVpnActive ? (
                <Badge
                  variant="outline"
                  className="badge-success text-[11px] font-sans px-2.5 py-0.5 flex items-center gap-1.5"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                  <span>Connected</span>
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="badge-warning text-[11px] font-sans px-2.5 py-0.5 flex items-center gap-1.5"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                  <span>Standby</span>
                </Badge>
              )}
            </div>
            <p className="text-[11px] font-mono text-muted-foreground">
              {isVpnActive
                ? `Interface: ${vpn.interface_name} • IP: ${vpn.ip} • Subnet Broadcast: ${vpn.broadcast_ip}`
                : 'Connect to Pritunl / OpenVPN to discover online teammates across your corporate subnet.'}
            </p>
          </div>
        </div>

        {/* Local Profile Override Sharing Control & Port Badge (Moved here per Screenshot 12.22.04 AM / 12.22.16 AM) */}
        <div className="flex items-center gap-2.5 flex-wrap sm:flex-nowrap shrink-0">
          <div className="inline-flex items-center p-0.5 rounded-geist bg-[#141414] border border-[#262626]">
            {profiles.length > 1 ? (
              <div className="flex items-center gap-1.5 pl-2 pr-1 border-r border-[#262626] mr-1">
                <Share2 className="h-3 w-3 text-accent-cyan shrink-0" />
                <select
                  value={selectedLocalProfile}
                  onChange={(e) => setSelectedLocalProfile(e.target.value)}
                  className="bg-transparent text-xs font-mono text-[#ededed] focus:outline-none cursor-pointer pr-1"
                  title="Select local profile to configure sharing"
                >
                  {profiles.map((p) => (
                    <option key={p.name} value={p.name} className="bg-[#141414] text-[#ededed]">
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>
            ) : profiles.length === 1 ? (
              <span className="text-[10px] font-mono uppercase tracking-wider text-[#666666] px-2 flex items-center gap-1 border-r border-[#262626] mr-1">
                <Share2 className="h-3 w-3 text-accent-cyan" />
                {profiles[0].name}:
              </span>
            ) : (
              <span className="text-[10px] font-mono uppercase tracking-wider text-[#666666] px-2 flex items-center gap-1 border-r border-[#262626] mr-1">
                <Share2 className="h-3 w-3 text-accent-cyan" />
                P2P:
              </span>
            )}

            {/* Off | Request | Auto Approve buttons */}
            <button
              type="button"
              disabled={isShareUpdating || !selectedLocalProfile}
              onClick={() => handleSetLocalProfileShareMode(selectedLocalProfile, 'off')}
              className={cn(
                'px-2 py-0.5 rounded-sm text-xs font-mono transition-all cursor-pointer',
                currentLocalShareMode === 'off'
                  ? 'bg-[#262626] text-[#ededed] shadow-xs font-semibold'
                  : 'text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]'
              )}
              title="Profile is private and not shared on VPN"
            >
              Off
            </button>
            <button
              type="button"
              disabled={isShareUpdating || !selectedLocalProfile}
              onClick={() => handleSetLocalProfileShareMode(selectedLocalProfile, 'request')}
              className={cn(
                'px-2 py-0.5 rounded-sm text-xs font-mono transition-all cursor-pointer flex items-center gap-1',
                currentLocalShareMode === 'request'
                  ? 'bg-[#142334] text-[#52a8ff] border border-[#1f4068] shadow-xs font-semibold'
                  : 'text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]'
              )}
              title="Profile is shared; teammate lease requests require manual approval"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  currentLocalShareMode === 'request' ? 'bg-[#52a8ff]' : 'bg-[#666666]'
                )}
              />
              Request
            </button>
            <button
              type="button"
              disabled={isShareUpdating || !selectedLocalProfile}
              onClick={() => handleSetLocalProfileShareMode(selectedLocalProfile, 'auto')}
              className={cn(
                'px-2 py-0.5 rounded-sm text-xs font-mono transition-all cursor-pointer flex items-center gap-1',
                currentLocalShareMode === 'auto'
                  ? 'bg-[#183925] text-[#46a758] border border-[#235836] shadow-xs font-semibold'
                  : 'text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]'
              )}
              title="Profile is shared; teammate lease requests are auto-approved"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  currentLocalShareMode === 'auto' ? 'bg-[#30a46c]' : 'bg-[#666666]'
                )}
              />
              Auto Approve
            </button>
          </div>

          <span className="inline-flex items-center gap-1.5 text-[11px] font-mono text-muted-foreground px-2.5 py-1 rounded-geist border border-border bg-[#141414]">
            <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
            <span>P2P Port :4319</span>
          </span>
        </div>
      </div>

      {/* Active Borrowed Leases Section */}
      {data && data.active_leases && data.active_leases.length > 0 && (
        <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
          <CardHeader className="p-5 pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-[#61afef]" />
                <CardTitle className="text-sm font-semibold text-foreground">
                  Active Borrowed Profiles
                </CardTitle>
                <Badge
                  variant="outline"
                  className="text-[10px] font-mono py-0 px-1.5 text-muted-foreground"
                >
                  {data.active_leases.length} active
                </Badge>
              </div>
              <span className="text-[11px] text-muted-foreground">
                Credentials running locally in sandbox
              </span>
            </div>
            <CardDescription className="text-xs text-muted-foreground pt-1">
              These profiles run 100% locally on your machine with zero dependency on the host's uptime.
            </CardDescription>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {data.active_leases.map((lease) => {
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
                          className={`text-[10px] font-mono px-2 py-0.5 rounded font-medium ${
                            lease.is_expired
                              ? 'bg-status-danger-bg text-status-danger border border-status-danger-border'
                              : 'bg-status-success-bg text-status-success border border-status-success-border'
                          }`}
                        >
                          {formatExpiresIn(lease.expires_at)}
                        </span>
                      </div>
                      <div className="text-[11px] text-muted-foreground flex items-center gap-1 font-mono">
                        <span>Source:</span>
                        <span className="text-foreground font-semibold">{lease.source_peer}</span>
                        <span className="text-[#666666]">/</span>
                        <span className="text-foreground">{lease.source_profile}</span>
                      </div>
                    </div>

                    <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                      <button
                        onClick={() => copyToClipboard(runCmd, lease.local_profile)}
                        className="inline-flex items-center gap-1.5 text-[11px] font-mono text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
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
                        className="h-6 px-2 text-[11px] text-status-danger hover:bg-status-danger-bg cursor-pointer"
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

      {/* Active Granted Leases Section (Teammates Borrowing Your Profiles) */}
      {data && data.granted_leases && data.granted_leases.length > 0 && (
        <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
          <CardHeader className="p-5 pb-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <Share2 className="h-4 w-4 text-accent-cyan" />
                <CardTitle className="text-sm font-semibold text-foreground">
                  Teammates Borrowing Your Profiles (Active Outbound Leases)
                </CardTitle>
                <Badge
                  variant="outline"
                  className="text-[10px] font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                >
                  {data.granted_leases.filter((l) => !l.revoked && new Date(l.expires_at).getTime() > Date.now()).length} active
                </Badge>
              </div>
              <span className="text-[11px] text-muted-foreground font-mono">
                You can revoke access to your credentials at any time
              </span>
            </div>
            <CardDescription className="text-xs text-muted-foreground pt-1">
              Active sessions on teammate machines authenticated via your shared profile quotas.
            </CardDescription>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {data.granted_leases.map((lease) => {
                const isExpired = new Date(lease.expires_at).getTime() <= Date.now();
                return (
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
                              'text-[10px] font-mono px-1.5 py-0.5 rounded border',
                              lease.auto_approved
                                ? 'bg-[#183925]/50 text-[#46a758] border-[#235836]/60'
                                : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                            )}
                          >
                            {lease.auto_approved ? 'Auto Approved' : 'Manual Approved'}
                          </span>
                        </div>
                        <span
                          className={cn(
                            'text-[10px] font-mono px-2 py-0.5 rounded font-medium',
                            lease.revoked
                              ? 'bg-status-danger-bg text-status-danger border border-status-danger-border'
                              : isExpired
                              ? 'bg-[#1e1e1e] text-[#888888] border border-[#333333]'
                              : 'bg-status-success-bg text-status-success border border-status-success-border'
                          )}
                        >
                          {lease.revoked ? 'Revoked' : isExpired ? 'Expired' : formatExpiresIn(lease.expires_at)}
                        </span>
                      </div>
                      <div className="text-[11px] text-muted-foreground flex items-center gap-1.5 font-mono">
                        <Laptop className="h-3 w-3 text-[#61afef]" />
                        <span>Borrower:</span>
                        <span className="text-foreground font-semibold truncate max-w-[220px]" title={lease.borrower_id}>
                          {lease.borrower_id}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center justify-between pt-2 border-t border-[#1f1f1f]">
                      <span className="text-[10px] font-mono text-[#666666]">
                        Lease ID: {lease.lease_id}
                      </span>
                      {!lease.revoked && !isExpired && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRevokeLease(lease.lease_id)}
                          className="h-6 px-2 text-[11px] text-status-danger hover:bg-status-danger-bg cursor-pointer"
                        >
                          <ShieldAlert className="h-3 w-3 mr-1" />
                          Revoke Lease
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Discovered Teammates & Shared Profiles Card */}
      <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
        <CardHeader className="p-5 pb-3">
          <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <Laptop className="h-4 w-4 text-[#61afef]" />
                <CardTitle className="text-sm font-semibold text-foreground">
                  Discovered Teammates
                </CardTitle>
                <Badge
                  variant="outline"
                  className="text-[10px] font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
                >
                  {filteredPeers.length} online
                </Badge>
              </div>
              <CardDescription className="text-xs text-muted-foreground">
                Teammates actively sharing AI profiles over the VPN subnet or local Wi-Fi.
              </CardDescription>
            </div>

            {/* Filter Input + Connect Peer + Refresh Button Group (Screenshot 12.21.08 AM) */}
            <div className="flex items-center gap-2 flex-wrap sm:flex-nowrap">
              <div className="relative w-full sm:w-64">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60" />
                <Input
                  ref={searchInputRef}
                  type="text"
                  value={filterQuery}
                  onChange={(e) => setFilterQuery(e.target.value)}
                  placeholder="Filter teammates..."
                  className="pl-8 pr-8 h-8 text-xs font-sans bg-[#0c0c0c] border-[#222222] text-[#ededed] placeholder:text-muted-foreground/50 rounded-geist focus:border-[#333333]"
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
                  <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 border border-[#2a2a2a] bg-[#161616] px-1 py-[1px] rounded text-[10px] font-mono text-muted-foreground pointer-events-none">
                    /
                  </kbd>
                )}
              </div>

              {/* Connect Peer Button beside filter */}
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsConnectOpen(true)}
                className="h-8 px-2.5 rounded-geist border-border bg-[#141414] hover:bg-[#1f1f1f] text-xs font-mono flex items-center gap-1.5 shrink-0"
              >
                <Plus className="h-3.5 w-3.5 text-accent-cyan" />
                <span>Connect Peer</span>
              </Button>

              {/* Refresh Button beside filter */}
              <Button
                variant="outline"
                size="sm"
                disabled={isLoading}
                onClick={handleManualRefresh}
                className="h-8 px-2.5 rounded-geist border-border bg-[#141414] hover:bg-[#1f1f1f] text-xs font-mono flex items-center gap-1.5 shrink-0"
                title="Live poll known peers on network"
              >
                <RotateCw className={cn('h-3.5 w-3.5', isLoading && 'animate-spin text-accent-cyan')} />
                <span>Refresh</span>
              </Button>
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-5 pt-0">
          {!data || !data.peers || data.peers.length === 0 ? (
            <div className="p-10 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] space-y-2.5">
              <Radio className="h-7 w-7 text-muted-foreground/30 mx-auto" />
              <div className="text-xs font-semibold text-foreground">No teammates discovered yet</div>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto leading-relaxed">
                Teammates connected to the internal VPN who share a profile will be automatically discovered via mDNS, or you can click "Connect Peer" above to add their IP/hostname directly.
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
                            <CardDescription className="text-[11px] font-mono text-muted-foreground truncate">
                              {peer.vpn_ip}
                            </CardDescription>
                          </div>
                          <Badge
                            variant="outline"
                            className="badge-success text-[10px] font-sans px-2 py-0.5 shrink-0 flex items-center gap-1.5"
                          >
                            <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                            <span>Online</span>
                          </Badge>
                        </div>
                      </CardHeader>

                      <CardContent className="p-4 pt-2 space-y-2">
                        <div className="flex items-center justify-between text-[10px] font-mono uppercase text-muted-foreground tracking-wider font-semibold">
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
                                  className="flex items-center justify-between gap-2 px-3 py-2 rounded-geist bg-[#141414] border border-[#222222] text-xs hover:border-[#333333] transition-colors"
                                >
                                  <div className="flex items-center gap-2 truncate min-w-0">
                                    <AgentBadge agent={agent} size="sm" />
                                    <span className="font-mono text-foreground font-medium truncate">
                                      {prof}
                                    </span>
                                    {sp && (
                                      <span
                                        className={cn(
                                          'text-[10px] font-mono px-1.5 py-0.5 rounded border shrink-0',
                                          isAuto
                                            ? 'bg-[#183925]/50 text-[#46a758] border-[#235836]/60'
                                            : 'bg-[#142334]/50 text-[#52a8ff] border-[#1f4068]/60'
                                        )}
                                      >
                                        {isAuto ? 'Auto' : 'Request'}
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
                                    className="h-6 px-2.5 text-xs font-mono rounded-geist border-[#2a2a2a] bg-[#1a1a1a] hover:bg-[#262626] text-foreground cursor-pointer shrink-0"
                                  >
                                    Borrow
                                  </Button>
                                </div>
                              );
                            })}
                          </div>
                        ) : (
                          <div className="p-3 rounded-geist bg-[#141414] border border-[#1e1e1e] text-[11px] text-muted-foreground text-center italic">
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
        </CardContent>
      </Card>

      {/* P2P Activity & Audit Log Card */}
      <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
        <CardHeader className="p-5 pb-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-accent-cyan" />
              <CardTitle className="text-sm font-semibold text-foreground">
                P2P Sharing Audit Log
              </CardTitle>
              <Badge
                variant="outline"
                className="text-[10px] font-mono py-0 px-2 rounded-full border-border bg-secondary text-muted-foreground"
              >
                {data?.audit_logs?.length || 0} events
              </Badge>
            </div>
            <span className="text-[11px] text-muted-foreground font-mono">
              Tamper-evident mesh audit trail
            </span>
          </div>
          <CardDescription className="text-xs text-muted-foreground pt-1">
            Real-time audit log of all lease requests, approvals, revocations, and sharing policy switches.
          </CardDescription>
        </CardHeader>
        <CardContent className="p-5 pt-0">
          {!data?.audit_logs || data.audit_logs.length === 0 ? (
            <div className="p-6 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground font-mono">
              No sharing events recorded yet.
            </div>
          ) : (
            <div className="max-h-60 overflow-y-auto space-y-1.5 pr-1 font-mono text-xs">
              {data.audit_logs.map((log) => {
                const actionColors: Record<string, string> = {
                  LEASE_GRANTED: 'text-[#46a758] bg-[#183925]/40 border-[#235836]',
                  LEASE_REQUESTED: 'text-[#52a8ff] bg-[#142334]/40 border-[#1f4068]',
                  LEASE_REVOKED: 'text-[#e06c75] bg-[#3a1d20]/40 border-[#5c282d]',
                  LEASE_DENIED: 'text-[#e06c75] bg-[#3a1d20]/40 border-[#5c282d]',
                  GLOBAL_MODE_CHANGED: 'text-[#c678dd] bg-[#2a1b33]/40 border-[#4a2b5a]',
                  PROFILE_SHARING_ON: 'text-[#46a758] bg-[#183925]/40 border-[#235836]',
                  PROFILE_SHARING_OFF: 'text-[#888888] bg-[#1e1e1e] border-[#333333]',
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
                    className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-2 px-2.5 rounded-geist bg-[#111111] border border-[#1f1f1f] text-[11px]"
                  >
                    <div className="flex items-center gap-2 truncate min-w-0">
                      <span className="text-[#666666] shrink-0">{timeStr}</span>
                      <span className={cn('px-1.5 py-0.5 rounded border text-[10px] shrink-0 uppercase font-semibold', color)}>
                        {log.action.replace(/_/g, ' ')}
                      </span>
                      <span className="text-foreground truncate">{log.details}</span>
                    </div>
                    {log.actor && (
                      <span className="text-[#666666] text-[10px] shrink-0 truncate max-w-[150px]">
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
                  <AgentBadge agent={selectedAgent} size="sm" />
                  <span className="font-mono text-[#ededed] font-medium">
                    {selectedProfile}
                  </span>
                </div>
              </div>
              <div className="flex items-center justify-between text-[#888888]">
                <span>Approval Policy</span>
                <span
                  className={cn(
                    'font-mono text-[11px] font-medium px-1.5 py-0.5 rounded border',
                    selectedIsAuto
                      ? 'bg-[#183925]/50 text-[#46a758] border-[#235836]/60'
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
                <span className="text-[10px] text-[#666666] font-mono">
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
