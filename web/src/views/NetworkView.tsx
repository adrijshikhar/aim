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
  UserCheck,
  Plus,
  X,
  Share2,
} from 'lucide-react';
import {
  getShareStatus,
  addSharePeer,
  requestShareLease,
  releaseShareLease,
  toggleShareProfile,
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

  // Top-right Sharing Slider state
  const [selectedLocalProfile, setSelectedLocalProfile] = React.useState<string>('');
  const [isShareUpdating, setIsShareUpdating] = React.useState(false);

  // Connect Peer Modal state
  const [isConnectOpen, setIsConnectOpen] = React.useState(false);
  const [addHost, setAddHost] = React.useState('');
  const [isAdding, setIsAdding] = React.useState(false);

  // Request Lease Modal state
  const [selectedPeer, setSelectedPeer] = React.useState<PeerInfoDTO | null>(null);
  const [selectedProfile, setSelectedProfile] = React.useState<string>('');
  const [leaseHours, setLeaseHours] = React.useState<number>(4);
  const [isRequesting, setIsRequesting] = React.useState(false);

  const searchInputRef = React.useRef<HTMLInputElement>(null);

  const loadStatus = React.useCallback(async () => {
    try {
      const res = await getShareStatus();
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
    loadStatus();
    const interval = setInterval(loadStatus, 5000);
    return () => clearInterval(interval);
  }, [loadStatus]);

  // Set default selected profile for sharing controls
  React.useEffect(() => {
    if (!selectedLocalProfile && profiles.length > 0) {
      setSelectedLocalProfile(profiles[0].name);
    }
  }, [profiles, selectedLocalProfile]);

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

  // Compute sharing state of selected local profile
  const currentSharedProfileConfig = React.useMemo(() => {
    if (!data?.shared_profiles || !selectedLocalProfile) return null;
    return data.shared_profiles.find((p) => p.name === selectedLocalProfile) || null;
  }, [data?.shared_profiles, selectedLocalProfile]);

  const currentShareMode: ShareMode = React.useMemo(() => {
    if (!currentSharedProfileConfig || !currentSharedProfileConfig.enabled) return 'off';
    if (currentSharedProfileConfig.auto_approve) return 'auto';
    return 'request';
  }, [currentSharedProfileConfig]);

  const handleSetShareMode = async (profileName: string, mode: ShareMode) => {
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
      const titles: Record<ShareMode, string> = {
        off: 'Sharing Disabled',
        request: 'Sharing Active (Request Approval)',
        auto: 'Sharing Active (Auto-Approve)',
      };
      const descs: Record<ShareMode, string> = {
        off: `Profile "${profileName}" is no longer shared on the VPN.`,
        request: `Profile "${profileName}" is shared. Teammate lease requests require your approval.`,
        auto: `Profile "${profileName}" is shared. Teammate lease requests are granted automatically.`,
      };
      toast({
        title: titles[mode],
        description: descs[mode],
        variant: mode === 'off' ? 'default' : 'success',
      });
      await loadStatus();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not update sharing state';
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
        agent: 'claude',
        hours: leaseHours,
      });

      if (res.ok && res.local_profile) {
        toast({
          title: 'Lease Granted',
          description: `Installed as ${res.local_profile}. Run with: aim run claude ${res.local_profile}`,
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

        {/* TOP RIGHT CONTROLS: Profile Picker + 3-Level Slider + Action Buttons */}
        <div className="flex flex-wrap items-center gap-2.5 self-start lg:self-auto shrink-0">
          {/* 3-Level Segmented Slider for local profile sharing */}
          <div className="inline-flex items-center p-0.5 rounded-geist bg-[#141414] border border-[#262626]">
            {profiles.length > 1 ? (
              <div className="flex items-center gap-1.5 pl-2 pr-1 border-r border-[#262626] mr-1">
                <Share2 className="h-3 w-3 text-accent-cyan shrink-0" />
                <select
                  value={selectedLocalProfile}
                  onChange={(e) => setSelectedLocalProfile(e.target.value)}
                  className="bg-transparent text-xs font-mono text-[#ededed] focus:outline-none cursor-pointer pr-1"
                  title="Select local profile to share"
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

            {/* 3 Level Buttons: Off | Request | Auto Approve */}
            <button
              type="button"
              disabled={isShareUpdating || !selectedLocalProfile}
              onClick={() => handleSetShareMode(selectedLocalProfile, 'off')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-all cursor-pointer',
                currentShareMode === 'off'
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
              onClick={() => handleSetShareMode(selectedLocalProfile, 'request')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-all cursor-pointer flex items-center gap-1.5',
                currentShareMode === 'request'
                  ? 'bg-[#142334] text-[#52a8ff] border border-[#1f4068] shadow-xs font-semibold'
                  : 'text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]'
              )}
              title="Profile is shared; teammate lease requests require manual approval"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  currentShareMode === 'request' ? 'bg-[#52a8ff]' : 'bg-[#666666]'
                )}
              />
              Request
            </button>
            <button
              type="button"
              disabled={isShareUpdating || !selectedLocalProfile}
              onClick={() => handleSetShareMode(selectedLocalProfile, 'auto')}
              className={cn(
                'px-2.5 py-1 rounded-sm text-xs font-mono transition-all cursor-pointer flex items-center gap-1.5',
                currentShareMode === 'auto'
                  ? 'bg-[#183925] text-[#46a758] border border-[#235836] shadow-xs font-semibold'
                  : 'text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]'
              )}
              title="Profile is shared; teammate lease requests are auto-approved"
            >
              <span
                className={cn(
                  'h-1.5 w-1.5 rounded-full',
                  currentShareMode === 'auto' ? 'bg-[#30a46c]' : 'bg-[#666666]'
                )}
              />
              Auto Approve
            </button>
          </div>

          {/* Refresh Action */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setIsLoading(true);
              loadStatus();
            }}
            disabled={isLoading}
            className="rounded-geist h-8 font-mono text-xs border-border bg-[#0c0c0c] text-foreground hover:bg-[#161616] cursor-pointer shrink-0"
          >
            <RotateCw className={`h-3.5 w-3.5 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>

          {/* Connect Peer Button (Solid White CTA) */}
          <Button
            size="sm"
            onClick={() => setIsConnectOpen(true)}
            className="bg-[#ffffff] text-[#000000] hover:bg-[#eaeaea] rounded-geist h-8 font-medium cursor-pointer shadow-sm transition-colors text-xs shrink-0"
          >
            <Plus className="h-3.5 w-3.5 mr-1 text-[#000000]" />
            Connect Peer
          </Button>
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

        <div className="flex items-center gap-2 shrink-0">
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

      {/* Discovered Teammates & Shared Profiles Card */}
      <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
        <CardHeader className="p-5 pb-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
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

            {/* Search filter input with '/' keyboard shortcut */}
            <div className="relative flex-1 max-w-sm">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60" />
              <Input
                ref={searchInputRef}
                type="text"
                value={filterQuery}
                onChange={(e) => setFilterQuery(e.target.value)}
                placeholder="Filter teammates by hostname, IP, or profile..."
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
                            {peer.profiles.map((prof) => (
                              <div
                                key={prof}
                                className="flex items-center justify-between gap-2 px-3 py-2 rounded-geist bg-[#141414] border border-[#222222] text-xs"
                              >
                                <div className="flex items-center gap-2 truncate">
                                  <UserCheck className="h-3.5 w-3.5 text-accent-cyan shrink-0" />
                                  <span className="font-mono text-foreground font-medium truncate">
                                    {prof}
                                  </span>
                                </div>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => {
                                    setSelectedPeer(peer);
                                    setSelectedProfile(prof);
                                  }}
                                  className="h-6 px-2.5 text-xs font-mono rounded-geist border-[#2a2a2a] bg-[#1a1a1a] hover:bg-[#262626] text-foreground cursor-pointer shrink-0"
                                >
                                  Borrow
                                </Button>
                              </div>
                            ))}
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

      {/* Connect Peer Dialog Modal */}
      <Dialog open={isConnectOpen} onOpenChange={setIsConnectOpen}>
        <DialogContent className="max-w-md">
          <form onSubmit={handleAddPeer}>
            <DialogHeader>
              <DialogTitle className="text-base font-semibold text-foreground">
                Connect Teammate Peer
              </DialogTitle>
              <DialogDescription className="text-xs text-muted-foreground">
                Enter a teammate's hostname (e.g. <code>Adrijs-MacBook-Pro.local</code>) or their VPN tunnel IP (e.g. <code>10.8.16.9</code>).
              </DialogDescription>
            </DialogHeader>

            <div className="py-4 space-y-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-foreground">
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
              <p className="text-[11px] font-mono text-muted-foreground">
                AIM will verify connectivity over port :4319 and query available shared profiles.
              </p>
            </div>

            <DialogFooter>
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

      {/* Request Lease Dialog Modal */}
      <Dialog
        open={!!(selectedPeer && selectedProfile)}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedPeer(null);
            setSelectedProfile('');
          }
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="text-base font-semibold text-foreground">
              Request Profile Lease
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              Borrow credentials from host{' '}
              <span className="font-semibold text-foreground">{selectedPeer?.hostname}</span>
            </DialogDescription>
          </DialogHeader>

          <div className="py-2 space-y-4">
            <div className="space-y-2.5 bg-[#141414] p-3.5 rounded-geist border border-[#222222] text-xs font-mono">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Remote Host:</span>
                <span className="text-foreground font-semibold">{selectedPeer?.hostname}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Remote Profile:</span>
                <span className="text-foreground font-semibold">{selectedProfile}</span>
              </div>
              <div className="flex justify-between items-center pt-2 border-t border-[#222222]">
                <span className="text-muted-foreground">Local Lease Identifier:</span>
                <span className="text-accent-cyan font-bold">
                  {selectedPeer?.hostname}:{selectedProfile}
                </span>
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-foreground">Lease Duration</label>
              <select
                value={leaseHours}
                onChange={(e) => setLeaseHours(Number(e.target.value))}
                className="w-full h-8 px-2.5 text-xs bg-[#161616] border border-[#262626] rounded-geist text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono"
              >
                <option value={1}>1 hour</option>
                <option value={2}>2 hours</option>
                <option value={4}>4 hours (Recommended)</option>
                <option value={8}>8 hours (Full Workday)</option>
              </select>
              <p className="text-[11px] text-muted-foreground leading-relaxed pt-1">
                Once approved, credentials operate 100% locally in your isolated profile directory. The teammate's laptop can sleep or disconnect without interrupting your active sessions.
              </p>
            </div>
          </div>

          <DialogFooter>
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
              {isRequesting ? 'Requesting...' : 'Confirm Lease'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
