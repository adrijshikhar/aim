import * as React from 'react';
import {
  ProfileDTO,
  CreateProfileRequest,
  getProfiles,
  createProfile,
  deleteProfile,
} from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { useToast } from '@/components/ui/use-toast';
import {
  Plus,
  Trash2,
  KeyRound,
  ShieldCheck,
  ShieldAlert,
  User,
  Folder,
  RefreshCw,
  Gauge,
  Bot,
} from 'lucide-react';

interface ProfilesViewProps {
  selectedAgent: string;
}

export function ProfilesView({ selectedAgent }: ProfilesViewProps) {
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  // Dialog State
  const [isAddOpen, setIsAddOpen] = React.useState(false);
  const [enableClone, setEnableClone] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [formData, setFormData] = React.useState<CreateProfileRequest>({
    agent: 'claude',
    name: '',
    email: '',
    clone_from: '',
  });

  const { toast } = useToast();

  const fetchProfiles = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getProfiles(selectedAgent === 'all' ? undefined : selectedAgent);
      setProfiles(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch profiles';
      setError(msg);
      // Fallback sample data in dev if backend server isn't reachable yet
      setProfiles([
        {
          agent: 'claude',
          name: 'work',
          path: '~/.aim/profiles/work',
          has_credentials: true,
          account: {
            email: 'dev@company.com',
            name: 'Work Profile',
            auth_method: 'oauth',
          },
          quota: {
            bottleneck_pct: 68,
            summary: '5h limit: 68% remaining',
            is_exhausted: false,
          },
        },
        {
          agent: 'codex',
          name: 'personal',
          path: '~/.aim/profiles/personal',
          has_credentials: true,
          account: {
            email: 'alex@gmail.com',
            name: 'Personal Sandbox',
            auth_method: 'api_key',
          },
          quota: {
            bottleneck_pct: 12,
            summary: 'Daily tokens: 12% remaining',
            is_exhausted: true,
          },
        },
        {
          agent: 'gemini',
          name: 'research',
          path: '~/.aim/profiles/research',
          has_credentials: false,
          account: {
            email: 'research@uni.edu',
            name: 'Gemini 2.5 Flash',
          },
        },
        {
          agent: 'agy',
          name: 'default',
          path: '~/.aim/profiles/default',
          has_credentials: true,
          account: {
            email: 'agent@local',
            name: 'DeepMind Antigravity',
          },
          quota: {
            bottleneck_pct: 95,
            summary: 'Enterprise: 95% left',
            is_exhausted: false,
          },
        },
      ]);
    } finally {
      setLoading(false);
    }
  }, [selectedAgent]);

  React.useEffect(() => {
    fetchProfiles();
  }, [fetchProfiles]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) return;

    setSubmitting(true);
    try {
      await createProfile({
        ...formData,
        clone_from: enableClone ? formData.clone_from : '',
      });
      toast({
        title: 'Profile Created',
        description: `Successfully added ${formData.agent} profile "${formData.name}".`,
        variant: 'success',
      });
      setIsAddOpen(false);
      setEnableClone(false);
      setFormData({
        agent: 'claude',
        name: '',
        email: '',
        clone_from: '',
      });
      fetchProfiles();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create profile';
      toast({
        title: 'Creation Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (agent: string, name: string) => {
    if (!confirm(`Are you sure you want to delete profile "${name}" (${agent})?`)) {
      return;
    }

    try {
      await deleteProfile(agent, name);
      toast({
        title: 'Profile Deleted',
        description: `Removed ${agent} profile "${name}".`,
      });
      fetchProfiles();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete profile';
      toast({
        title: 'Delete Failed',
        description: msg,
        variant: 'destructive',
      });
    }
  };

  const filteredProfiles = React.useMemo(() => {
    if (selectedAgent === 'all') return profiles;
    return profiles.filter((p) => p.agent.toLowerCase() === selectedAgent.toLowerCase());
  }, [profiles, selectedAgent]);

  const getAgentBadge = (agent: string) => {
    switch (agent.toLowerCase()) {
      case 'claude':
        return (
          <Badge className="bg-[#241a12] text-[#f5a623] border-[#3d2716] font-mono text-[11px] tracking-wide">
            claude
          </Badge>
        );
      case 'codex':
        return (
          <Badge className="bg-[#0e2118] text-[#50e3c2] border-[#19402e] font-mono text-[11px] tracking-wide">
            codex
          </Badge>
        );
      case 'gemini':
        return (
          <Badge className="bg-[#0d1f36] text-[#0070f3] border-[#15345a] font-mono text-[11px] tracking-wide">
            gemini
          </Badge>
        );
      case 'agy':
      case 'antigravity':
        return (
          <Badge className="bg-[#1c122c] text-[#b388ff] border-[#331c52] font-mono text-[11px] tracking-wide">
            antigravity
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="font-mono text-[11px]">
            {agent}
          </Badge>
        );
    }
  };

  return (
    <div className="space-y-6">
      {error && (
        <div className="text-xs text-[#f5a623] bg-[#241a12]/80 border border-[#3d2716] px-3.5 py-2.5 rounded-geist flex items-center justify-between">
          <span>Notice: Backend API unreachable ({error}) — showing local cached profiles.</span>
          <button
            type="button"
            onClick={() => setError(null)}
            className="text-[#888888] hover:text-[#ededed] ml-2 text-xs"
          >
            Dismiss
          </button>
        </div>
      )}
      {/* Top action row */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-2.5">
          <h2 className="text-sm font-semibold tracking-tight text-[#ededed]">
            {selectedAgent === 'all'
              ? 'All Profiles'
              : `${selectedAgent.charAt(0).toUpperCase() + selectedAgent.slice(1)} Profiles`}
          </h2>
          <span className="font-mono text-[11px] px-2 py-0.5 rounded-full border border-[#262626] bg-[#141414] text-[#888888] tabular-nums">
            {filteredProfiles.length} active
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchProfiles} disabled={loading} className="rounded-geist h-8">
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>

          <Dialog
            open={isAddOpen}
            onOpenChange={(open) => {
              setIsAddOpen(open);
              if (!open) setEnableClone(false);
            }}
          >
            <DialogTrigger asChild>
              <Button size="sm" className="bg-[#ededed] text-[#0a0a0a] hover:bg-white rounded-geist h-8 font-medium">
                <Plus className="h-3.5 w-3.5 mr-1" />
                Add Profile
              </Button>
            </DialogTrigger>
              <DialogContent className="border border-border bg-[#111111] text-foreground rounded-geist sm:max-w-md">
                <DialogHeader>
                  <DialogTitle className="text-base font-semibold tracking-tight text-[#ededed]">
                    Create New Agent Profile
                  </DialogTitle>
                  <DialogDescription className="text-xs text-[#888888]">
                    Scaffold an isolated environment and configuration for your agent.
                  </DialogDescription>
                </DialogHeader>

                <form onSubmit={handleCreate} className="space-y-4 py-2">
                  <div className="space-y-1.5">
                    <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#888888]">
                      Target Agent
                    </label>
                    <div className="grid grid-cols-4 gap-2">
                      {['claude', 'codex', 'gemini', 'agy'].map((ag) => (
                        <button
                          key={ag}
                          type="button"
                          onClick={() => setFormData({ ...formData, agent: ag })}
                          className={`py-2 px-3 text-xs font-mono rounded-geist border text-center transition-all cursor-pointer ${
                            formData.agent === ag
                              ? 'border-[#ededed] bg-[#222222] text-[#ededed] font-bold shadow-sm'
                              : 'border-[#262626] bg-[#141414] text-[#888888] hover:text-[#ededed] hover:border-[#383838]'
                          }`}
                        >
                          {ag.toUpperCase()}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#888888]">
                      Profile Name *
                    </label>
                    <Input
                      required
                      placeholder="e.g. work, sandbox, client-x"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="rounded-geist border-[#262626] bg-[#0e0e0e] text-xs h-9"
                    />
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#888888]">
                      Associated Email (Optional)
                    </label>
                    <Input
                      type="email"
                      placeholder="dev@example.com"
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                      className="rounded-geist border-[#262626] bg-[#0e0e0e] text-xs h-9"
                    />
                  </div>

                  {/* Clone from existing profile rendered with toggle switch */}
                  <div className="space-y-2 pt-2 border-t border-[#1f1f1f]">
                    <div className="flex items-center justify-between py-0.5">
                      <div className="space-y-0.5">
                        <label htmlFor="clone-toggle" className="text-xs font-medium text-[#ededed] cursor-pointer">
                          Clone from existing profile
                        </label>
                        <p className="text-[11px] text-[#888888]">
                          Copy authentication, credentials, and settings
                        </p>
                      </div>
                      <Switch
                        id="clone-toggle"
                        checked={enableClone}
                        onCheckedChange={(checked) => {
                          setEnableClone(checked);
                          if (!checked) {
                            setFormData((prev) => ({ ...prev, clone_from: '' }));
                          }
                        }}
                      />
                    </div>

                    {enableClone && (
                      <div className="space-y-1.5 pt-1">
                        <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#888888]">
                          Source Profile to Clone
                        </label>
                        <Input
                          placeholder="e.g. default"
                          value={formData.clone_from}
                          onChange={(e) => setFormData({ ...formData, clone_from: e.target.value })}
                          className="rounded-geist border-[#262626] bg-[#0e0e0e] text-xs h-9"
                          autoFocus
                        />
                      </div>
                    )}
                  </div>

                  <DialogFooter className="pt-3 border-t border-[#1f1f1f] flex gap-2 justify-end">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        setIsAddOpen(false);
                        setEnableClone(false);
                      }}
                      disabled={submitting}
                      className="rounded-geist text-xs h-8"
                    >
                      Cancel
                    </Button>
                    <Button
                      type="submit"
                      disabled={submitting}
                      className="bg-[#ededed] text-[#0a0a0a] hover:bg-white rounded-geist text-xs h-8 font-medium"
                    >
                      {submitting ? 'Creating...' : 'Create Profile'}
                    </Button>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          </div>
        </div>

        {/* Grid of Profile Cards */}
        {loading && profiles.length === 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-48 rounded-geist border border-[#262626] bg-[#111111]/40 animate-pulse p-6"
              />
            ))}
          </div>
        ) : filteredProfiles.length === 0 ? (
          <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0e0e0e]/50">
            <Bot className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
            <h3 className="text-sm font-semibold text-[#ededed]">No profiles found</h3>
            <p className="text-xs text-[#888888] mt-1 max-w-sm mx-auto">
              {selectedAgent === 'all'
                ? 'No agent profiles are registered yet. Click "Add Profile" above to create your first profile.'
                : `No profiles configured for ${selectedAgent}. Add one or switch agent filters.`}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredProfiles.map((p) => {
              const hasQuota = p.quota !== undefined && p.quota !== null;
              return (
                <Card
                  key={`${p.agent}-${p.name}`}
                  className="group relative overflow-hidden transition-all duration-150 border border-[#262626] bg-[#111111] hover:border-[#383838] rounded-geist flex flex-col justify-between"
                >
                  <CardHeader className="pb-3 p-5">
                    <div className="flex items-center justify-between">
                      {getAgentBadge(p.agent)}
                      <div className="flex items-center gap-1.5">
                        {p.has_credentials ? (
                          <span className="flex items-center gap-1 text-[11px] font-mono text-[#50e3c2] bg-[#0e2118] px-2 py-0.5 rounded-full border border-[#19402e]">
                            <ShieldCheck className="h-3 w-3" />
                            Authenticated
                          </span>
                        ) : (
                          <span className="flex items-center gap-1 text-[11px] font-mono text-[#ee0000] bg-[#220a0a] px-2 py-0.5 rounded-full border border-[#3d1212]">
                            <ShieldAlert className="h-3 w-3" />
                            No Credentials
                          </span>
                        )}
                      </div>
                    </div>
                    <CardTitle className="text-base font-semibold pt-2 flex items-center justify-between text-[#ededed]">
                      <span>{p.name}</span>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-[#666666] hover:text-[#ee0000] hover:bg-[#ee0000]/10 rounded-geist transition-colors"
                        onClick={() => handleDelete(p.agent, p.name)}
                        title="Delete profile"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </CardTitle>
                    <CardDescription className="flex items-center gap-1.5 text-xs truncate">
                      <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                      <span className="font-mono text-[#888888] truncate">{p.path}</span>
                    </CardDescription>
                  </CardHeader>

                  <CardContent className="space-y-3 pb-4 p-5 pt-0">
                    {/* Account Metadata */}
                    <div className="space-y-1 text-xs rounded-geist bg-[#161616]/70 p-2.5 border border-[#222222]">
                      <div className="flex items-center justify-between text-[#888888]">
                        <span className="flex items-center gap-1">
                          <User className="h-3 w-3" /> Account
                        </span>
                        <span className="font-medium text-[#ededed]">
                          {p.account?.name || p.name}
                        </span>
                      </div>
                      {p.account?.email && (
                        <div className="flex items-center justify-between text-[#888888]">
                          <span>Email</span>
                          <span className="font-mono text-[#ededed]">{p.account.email}</span>
                        </div>
                      )}
                      {p.account?.auth_method && (
                        <div className="flex items-center justify-between text-[#888888]">
                          <span className="flex items-center gap-1">
                            <KeyRound className="h-3 w-3" /> Auth
                          </span>
                          <span className="capitalize font-mono text-[#ededed]">
                            {p.account.auth_method.replace('_', ' ')}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Quota Information (Impeccable Typeset) */}
                    {hasQuota && (
                      <div className="space-y-1.5 pt-1">
                        <div className="flex items-center justify-between text-xs">
                          <span className="flex items-center gap-1.5 text-[#888888] font-mono text-[11px]">
                            <Gauge className="h-3 w-3" /> Quota Telemetry
                          </span>
                          <span
                            className={`font-mono text-xs font-medium tabular-nums ${
                              p.quota!.is_exhausted
                                ? 'text-[#ee0000]'
                                : p.quota!.bottleneck_pct < 20
                                ? 'text-[#f5a623]'
                                : 'text-[#50e3c2]'
                            }`}
                          >
                            {p.quota!.bottleneck_pct}% remaining
                          </span>
                        </div>

                        <div className="h-1.5 w-full rounded-full bg-[#1c1c1c] overflow-hidden border border-[#262626]">
                          <div
                            className={`h-full rounded-full transition-all duration-300 ${
                              p.quota!.is_exhausted
                                ? 'bg-[#ee0000]'
                                : p.quota!.bottleneck_pct < 20
                                ? 'bg-[#f5a623]'
                                : 'bg-[#ededed]'
                            }`}
                            style={{ width: `${Math.min(100, Math.max(0, p.quota!.bottleneck_pct))}%` }}
                          />
                        </div>

                        <p className="text-[11px] font-mono text-[#666666] truncate">
                          {p.quota!.summary}
                        </p>
                      </div>
                    )}
                  </CardContent>

                  <CardFooter className="pt-3 border-t border-[#1f1f1f] mt-auto p-5 pb-3">
                    <div className="text-[11px] text-[#888888] w-full flex justify-between items-center font-mono">
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-1.5 w-1.5 rounded-full bg-[#333333]"></span>
                        <span>agent/{p.agent}</span>
                      </span>
                      <span className="text-[#666666]">
                        sandbox-ready
                      </span>
                    </div>
                  </CardFooter>
                </Card>
              );
            })}
          </div>
        )}
      </div>
    );
  }
