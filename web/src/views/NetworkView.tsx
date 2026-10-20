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
} from 'lucide-react';
import {
  getShareStatus,
  addSharePeer,
  requestShareLease,
  releaseShareLease,
  ShareStatusDTO,
  PeerInfoDTO,
} from '@/lib/api';
import { useToast } from '@/components/ui/use-toast';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export function NetworkView() {
  const { toast } = useToast();
  const [data, setData] = React.useState<ShareStatusDTO | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [addHost, setAddHost] = React.useState('');
  const [isAdding, setIsAdding] = React.useState(false);
  const [filterQuery, setFilterQuery] = React.useState('');
  const [copiedCmd, setCopiedCmd] = React.useState<string | null>(null);

  // Request Lease Modal state
  const [selectedPeer, setSelectedPeer] = React.useState<PeerInfoDTO | null>(null);
  const [selectedProfile, setSelectedProfile] = React.useState<string>('');
  const [leaseHours, setLeaseHours] = React.useState<number>(4);
  const [isRequesting, setIsRequesting] = React.useState(false);

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
      {/* Header bar */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2 border-b border-border">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Radio className="h-4 w-4 text-[#56b6c2]" />
            <h1 className="text-base font-semibold tracking-tight text-foreground">
              P2P Profile Sharing & Credential Leasing
            </h1>
            <Badge variant="outline" className="font-mono text-[10px] uppercase tracking-wider py-0 px-1.5 text-muted-foreground border-border">
              Mesh
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl leading-relaxed">
            Borrow high-tier Claude & AI assistant quotas from teammates on the internal VPN (Pritunl / OpenVPN) or local Wi-Fi. Borrowed profiles are isolated locally as{' '}
            <code className="text-[#56b6c2] font-mono px-1 py-0.5 rounded bg-[#111111] border border-border text-[11px]">
              host:&lt;profile_name&gt;
            </code>
            .
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            setIsLoading(true);
            loadStatus();
          }}
          disabled={isLoading}
          className="rounded-geist h-8 font-mono text-xs border-border bg-card text-foreground hover:bg-secondary cursor-pointer shrink-0"
        >
          <RotateCw className={`h-3.5 w-3.5 mr-1.5 ${isLoading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* VPN Tunnel Status Banner */}
      <div
        className={`p-4 rounded-geist border transition-colors ${
          isVpnActive
            ? 'bg-[#0a0a0a] border-[#222222]'
            : 'bg-[#141009] border-[#38260c]'
        }`}
      >
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div
              className={`h-9 w-9 rounded-md flex items-center justify-center shrink-0 border ${
                isVpnActive
                  ? 'bg-[#122619] border-[#1d4d2d] text-[#30a46c]'
                  : 'bg-[#2b1d0c] border-[#4a3416] text-[#f5a623]'
              }`}
            >
              {isVpnActive ? <Wifi className="h-4 w-4" /> : <AlertTriangle className="h-4 w-4" />}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-semibold text-xs text-foreground tracking-tight">
                  {isVpnActive ? 'Corporate VPN Tunnel Active' : 'VPN Tunnel Offline'}
                </span>
                <span
                  className={`text-[9px] font-mono px-1.5 py-0.2 rounded font-semibold tracking-wider ${
                    isVpnActive
                      ? 'bg-[#183925] text-[#46a758] border border-[#235836]'
                      : 'bg-[#33220c] text-[#ff9900] border border-[#523916]'
                  }`}
                >
                  {isVpnActive ? 'CONNECTED' : 'STANDBY'}
                </span>
              </div>
              <p className="text-[11px] font-mono text-muted-foreground mt-0.5">
                {isVpnActive
                  ? `Interface: ${vpn.interface_name} • IP: ${vpn.ip} • Subnet Broadcast: ${vpn.broadcast_ip}`
                  : 'Connect to corporate VPN (Pritunl / OpenVPN) to discover teammates across your subnet.'}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 text-xs font-mono text-muted-foreground shrink-0">
            <span className="h-2 w-2 rounded-full bg-[#30a46c] animate-pulse" />
            <span>P2P Port :4319</span>
          </div>
        </div>
      </div>

      {/* Active Borrowed Leases Section */}
      {data && data.active_leases && data.active_leases.length > 0 && (
        <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
          <CardHeader className="p-4 pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-[#61afef]" />
                <CardTitle className="text-sm font-semibold text-foreground">
                  Active Borrowed Profiles
                </CardTitle>
                <Badge variant="outline" className="text-[10px] font-mono py-0 px-1.5 text-muted-foreground">
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
          <CardContent className="p-4 pt-0">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {data.active_leases.map((lease) => {
                const runCmd = `aim run ${lease.agent} ${lease.local_profile}`;
                return (
                  <div
                    key={lease.local_profile}
                    className="p-3.5 rounded-geist border border-[#222222] bg-[#111111] flex flex-col justify-between gap-3 shadow-xs"
                  >
                    <div className="space-y-1.5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-mono font-bold text-[#56b6c2] tracking-tight">
                            {lease.local_profile}
                          </span>
                          <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-[#1c1c1c] border border-border text-muted-foreground">
                            {lease.agent}
                          </span>
                        </div>
                        <span
                          className={`text-[10px] font-mono px-1.5 py-0.5 rounded font-medium ${
                            lease.is_expired
                              ? 'bg-[#3b1215] text-[#ff6369] border border-[#591d22]'
                              : 'bg-[#183925] text-[#46a758] border border-[#235836]'
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
                          <Check className="h-3 w-3 text-[#30a46c]" />
                        ) : (
                          <Copy className="h-3 w-3" />
                        )}
                        <code>{runCmd}</code>
                      </button>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRelease(lease.local_profile)}
                        className="h-6 px-2 text-[11px] text-[#e06c75] hover:text-[#e06c75] hover:bg-[#2b1315] cursor-pointer"
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

      {/* Discovered Teammates & Shared Profiles */}
      <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
        <CardHeader className="p-4 pb-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <Laptop className="h-4 w-4 text-[#61afef]" />
                <CardTitle className="text-sm font-semibold text-foreground">
                  Discovered Teammates
                </CardTitle>
                <Badge variant="outline" className="text-[10px] font-mono py-0 px-1.5 text-muted-foreground">
                  {filteredPeers.length} online
                </Badge>
              </div>
              <CardDescription className="text-xs text-muted-foreground">
                Teammates actively sharing AI profiles over the VPN subnet or local Wi-Fi.
              </CardDescription>
            </div>

            {/* Quick Connect & Search */}
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
              <div className="relative">
                <Search className="h-3.5 w-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="text"
                  value={filterQuery}
                  onChange={(e) => setFilterQuery(e.target.value)}
                  placeholder="Filter peers..."
                  className="h-8 pl-8 pr-2.5 text-xs font-mono bg-background border-border rounded-geist w-full sm:w-40"
                />
              </div>

              <form onSubmit={handleAddPeer} className="flex items-center gap-1.5">
                <Input
                  type="text"
                  value={addHost}
                  onChange={(e) => setAddHost(e.target.value)}
                  placeholder="Hostname or VPN IP"
                  className="h-8 px-2.5 text-xs font-mono bg-background border-border rounded-geist w-full sm:w-48"
                />
                <Button
                  type="submit"
                  size="sm"
                  disabled={isAdding || !addHost.trim()}
                  className="h-8 px-3 rounded-geist bg-[#1c1c1c] hover:bg-[#262626] border border-border text-xs font-medium text-foreground cursor-pointer transition-colors shrink-0"
                >
                  {isAdding ? 'Pinging...' : 'Connect'}
                </Button>
              </form>
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-4 pt-0">
          {!data || !data.peers || data.peers.length === 0 ? (
            <div className="p-10 text-center border border-dashed border-[#262626] rounded-geist bg-[#0e0e0e] space-y-2.5">
              <Radio className="h-7 w-7 text-muted-foreground/30 mx-auto" />
              <div className="text-xs font-semibold text-foreground">No teammates discovered yet</div>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto leading-relaxed">
                Teammates connected to the internal VPN who share a profile will be automatically discovered via mDNS, or you can enter their hostname or VPN IP above.
              </p>
            </div>
          ) : filteredPeers.length === 0 ? (
            <div className="p-6 text-center border border-dashed border-border rounded-geist bg-[#0e0e0e] text-xs text-muted-foreground">
              No teammates match filter &ldquo;{filterQuery}&rdquo;.
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {filteredPeers.map((peer) => {
                const profileCount = peer.profiles?.length || 0;
                return (
                  <div
                    key={peer.id || peer.hostname}
                    className="p-4 rounded-geist border border-[#222222] bg-[#111111] space-y-3.5 shadow-xs hover:border-[#333333] transition-colors"
                  >
                    <div className="flex items-start justify-between">
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold text-xs text-foreground tracking-tight">
                            {peer.hostname}
                          </h3>
                        </div>
                        <p className="text-[11px] font-mono text-muted-foreground">
                          {peer.vpn_ip}
                        </p>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <span className="h-2 w-2 rounded-full bg-[#30a46c] ring-2 ring-[#30a46c]/20" />
                        <span className="text-[10px] font-mono text-muted-foreground">online</span>
                      </div>
                    </div>

                    <div className="space-y-2 pt-2 border-t border-[#1f1f1f]">
                      <div className="flex items-center justify-between text-[10px] font-mono uppercase text-muted-foreground tracking-wider font-semibold">
                        <span>Shared Profiles</span>
                        <span>{profileCount} offered</span>
                      </div>

                      {peer.profiles && peer.profiles.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {peer.profiles.map((prof) => (
                            <div
                              key={prof}
                              className="flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-geist bg-[#161616] border border-[#262626] text-xs w-full"
                            >
                              <div className="flex items-center gap-1.5 truncate">
                                <UserCheck className="h-3 w-3 text-[#56b6c2] shrink-0" />
                                <span className="font-mono text-foreground font-medium truncate">{prof}</span>
                              </div>
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => {
                                  setSelectedPeer(peer);
                                  setSelectedProfile(prof);
                                }}
                                className="h-6 px-2 text-[10px] font-medium border-[#30a46c]/40 text-[#46a758] hover:bg-[#183925] hover:text-[#46a758] rounded-geist cursor-pointer shrink-0"
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
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Lease Request Modal */}
      {selectedPeer && selectedProfile && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-[#0e0e0e] border border-[#262626] rounded-geist max-w-md w-full p-5 space-y-4 shadow-2xl animate-in fade-in-50 zoom-in-95">
            <div className="flex items-start justify-between pb-1 border-b border-[#222222]">
              <div className="space-y-0.5">
                <h3 className="text-sm font-bold text-foreground tracking-tight">Request Profile Lease</h3>
                <p className="text-xs text-muted-foreground">
                  Borrow AI credentials from host <span className="font-semibold text-foreground">{selectedPeer.hostname}</span>
                </p>
              </div>
              <button
                onClick={() => {
                  setSelectedPeer(null);
                  setSelectedProfile('');
                }}
                className="text-muted-foreground hover:text-foreground text-xs font-mono p-1 cursor-pointer"
              >
                ✕
              </button>
            </div>

            <div className="space-y-2.5 bg-[#141414] p-3.5 rounded-geist border border-[#222222] text-xs font-mono">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Remote Host:</span>
                <span className="text-foreground font-semibold">{selectedPeer.hostname}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Remote Profile:</span>
                <span className="text-foreground font-semibold">{selectedProfile}</span>
              </div>
              <div className="flex justify-between items-center pt-2 border-t border-[#222222]">
                <span className="text-muted-foreground">Local Lease Identifier:</span>
                <span className="text-[#56b6c2] font-bold">
                  {selectedPeer.hostname}:{selectedProfile}
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

            <div className="flex justify-end gap-2 pt-2 border-t border-[#222222]">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSelectedPeer(null);
                  setSelectedProfile('');
                }}
                className="rounded-geist h-8 text-xs border-border text-foreground hover:bg-secondary cursor-pointer"
              >
                Cancel
              </Button>
              <Button
                size="sm"
                disabled={isRequesting}
                onClick={handleRequestLease}
                className="rounded-geist h-8 text-xs font-medium bg-foreground text-background hover:bg-foreground/90 transition-colors cursor-pointer disabled:opacity-50"
              >
                {isRequesting ? 'Requesting Lease...' : 'Confirm Lease'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
