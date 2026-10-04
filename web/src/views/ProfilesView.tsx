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
  Sparkles,
  Bot,
  Zap,
} from 'lucide-react';

interface ProfilesViewProps {
  selectedAgent: string;
  onSelectAgent?: (agent: string) => void;
}

const AGENTS = [
  { id: 'all', label: 'All Agents' },
  { id: 'claude', label: 'Claude', color: 'text-amber-400 border-amber-500/30 bg-amber-500/10' },
  { id: 'codex', label: 'Codex', color: 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10' },
  { id: 'gemini', label: 'Gemini', color: 'text-blue-400 border-blue-500/30 bg-blue-500/10' },
  { id: 'agy', label: 'Antigravity', color: 'text-purple-400 border-purple-500/30 bg-purple-500/10' },
];

export function ProfilesView({ selectedAgent, onSelectAgent }: ProfilesViewProps) {
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  // Dialog State
  const [isAddOpen, setIsAddOpen] = React.useState(false);
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
      await createProfile(formData);
      toast({
        title: 'Profile Created',
        description: `Successfully added ${formData.agent} profile "${formData.name}".`,
        variant: 'success',
      });
      setIsAddOpen(false);
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
        return <Badge className="bg-amber-500/15 text-amber-300 border-amber-500/30">Claude</Badge>;
      case 'codex':
        return <Badge className="bg-emerald-500/15 text-emerald-300 border-emerald-500/30">Codex</Badge>;
      case 'gemini':
        return <Badge className="bg-blue-500/15 text-blue-300 border-blue-500/30">Gemini</Badge>;
      case 'agy':
      case 'antigravity':
        return <Badge className="bg-purple-500/15 text-purple-300 border-purple-500/30">Antigravity</Badge>;
      default:
        return <Badge variant="outline">{agent}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      {error && (
        <div className="text-xs text-amber-300 bg-amber-500/10 border border-amber-500/20 px-3.5 py-2.5 rounded-lg flex items-center justify-between">
          <span>Notice: Backend API unreachable ({error}) — showing local cached profiles.</span>
          <button
            type="button"
            onClick={() => setError(null)}
            className="text-muted-foreground hover:text-foreground ml-2 text-xs"
          >
            Dismiss
          </button>
        </div>
      )}
      {/* Top action row */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-2 flex-wrap">
          {AGENTS.map((item) => (
            <Button
              key={item.id}
              variant={selectedAgent === item.id ? 'default' : 'outline'}
              size="sm"
              onClick={() => onSelectAgent?.(item.id)}
              className="text-xs transition-all"
            >
              {item.id === 'all' && <Sparkles className="h-3.5 w-3.5 mr-1" />}
              {item.id === 'claude' && <Bot className="h-3.5 w-3.5 mr-1 text-amber-400" />}
              {item.id === 'codex' && <Zap className="h-3.5 w-3.5 mr-1 text-emerald-400" />}
              {item.id === 'gemini' && <Sparkles className="h-3.5 w-3.5 mr-1 text-blue-400" />}
              {item.id === 'agy' && <Bot className="h-3.5 w-3.5 mr-1 text-purple-400" />}
              {item.label}
            </Button>
          ))}
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={fetchProfiles} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>

          <Dialog open={isAddOpen} onOpenChange={setIsAddOpen}>
            <DialogTrigger asChild>
              <Button size="sm" className="bg-primary text-primary-foreground">
                <Plus className="h-4 w-4 mr-1.5" />
                Add Profile
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Agent Profile</DialogTitle>
                <DialogDescription>
                  Scaffold an isolated environment and configuration for your agent.
                </DialogDescription>
              </DialogHeader>

              <form onSubmit={handleCreate} className="space-y-4 py-2">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Target Agent
                  </label>
                  <div className="grid grid-cols-4 gap-2">
                    {['claude', 'codex', 'gemini', 'agy'].map((ag) => (
                      <button
                        key={ag}
                        type="button"
                        onClick={() => setFormData({ ...formData, agent: ag })}
                        className={`py-2 px-3 text-xs font-medium rounded-md border text-center transition-all cursor-pointer ${
                          formData.agent === ag
                            ? 'border-primary bg-primary/10 text-primary font-bold shadow-sm'
                            : 'border-border bg-card/50 text-muted-foreground hover:bg-accent'
                        }`}
                      >
                        {ag.toUpperCase()}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Profile Name *
                  </label>
                  <Input
                    required
                    placeholder="e.g. work, sandbox, client-x"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Associated Email (Optional)
                  </label>
                  <Input
                    type="email"
                    placeholder="dev@example.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Clone From Existing Profile (Optional)
                  </label>
                  <Input
                    placeholder="e.g. default"
                    value={formData.clone_from}
                    onChange={(e) => setFormData({ ...formData, clone_from: e.target.value })}
                  />
                </div>

                <DialogFooter className="pt-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setIsAddOpen(false)}
                    disabled={submitting}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={submitting}>
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
              className="h-48 rounded-xl border border-border/50 bg-card/40 animate-pulse p-6"
            />
          ))}
        </div>
      ) : filteredProfiles.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border/80 p-12 text-center">
          <Bot className="h-10 w-10 text-muted-foreground mx-auto mb-3 opacity-60" />
          <h3 className="text-base font-semibold">No profiles found</h3>
          <p className="text-sm text-muted-foreground mt-1 max-w-sm mx-auto">
            {selectedAgent === 'all'
              ? 'No agent profiles are registered yet. Click "Add Profile" above to create your first profile.'
              : `No profiles configured for ${selectedAgent}. Add one or switch agent filters.`}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredProfiles.map((p) => {
            const hasQuota = p.quota !== undefined && p.quota !== null;
            return (
              <Card
                key={`${p.agent}-${p.name}`}
                className="group relative overflow-hidden transition-all duration-200 hover:border-primary/50 hover:shadow-md"
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    {getAgentBadge(p.agent)}
                    <div className="flex items-center gap-1.5">
                      {p.has_credentials ? (
                        <span className="flex items-center gap-1 text-[11px] font-medium text-emerald-400 bg-emerald-950/60 px-2 py-0.5 rounded-full border border-emerald-500/20">
                          <ShieldCheck className="h-3 w-3" />
                          Authenticated
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-[11px] font-medium text-rose-400 bg-rose-950/50 px-2 py-0.5 rounded-full border border-rose-500/20">
                          <ShieldAlert className="h-3 w-3" />
                          No Credentials
                        </span>
                      )}
                    </div>
                  </div>
                  <CardTitle className="text-xl pt-2 flex items-center justify-between">
                    <span>{p.name}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                      onClick={() => handleDelete(p.agent, p.name)}
                      title="Delete profile"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </CardTitle>
                  <CardDescription className="flex items-center gap-1.5 text-xs truncate">
                    <Folder className="h-3 w-3 shrink-0 text-muted-foreground/70" />
                    <span className="font-mono text-muted-foreground/80 truncate">{p.path}</span>
                  </CardDescription>
                </CardHeader>

                <CardContent className="space-y-3.5 pb-4">
                  {/* Account Metadata */}
                  <div className="space-y-1.5 text-xs rounded-lg bg-muted/40 p-2.5 border border-border/40">
                    <div className="flex items-center justify-between text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <User className="h-3 w-3" /> Account
                      </span>
                      <span className="font-medium text-foreground">
                        {p.account?.name || p.name}
                      </span>
                    </div>
                    {p.account?.email && (
                      <div className="flex items-center justify-between text-muted-foreground">
                        <span>Email</span>
                        <span className="font-mono text-foreground">{p.account.email}</span>
                      </div>
                    )}
                    {p.account?.auth_method && (
                      <div className="flex items-center justify-between text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <KeyRound className="h-3 w-3" /> Auth
                        </span>
                        <span className="capitalize font-mono text-foreground">
                          {p.account.auth_method.replace('_', ' ')}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Quota Information */}
                  {hasQuota && (
                    <div className="space-y-1.5">
                      <div className="flex items-center justify-between text-xs">
                        <span className="flex items-center gap-1 text-muted-foreground font-medium">
                          <Gauge className="h-3.5 w-3.5" /> Quota Status
                        </span>
                        <span
                          className={`font-semibold ${
                            p.quota!.is_exhausted
                              ? 'text-rose-400'
                              : p.quota!.bottleneck_pct < 20
                              ? 'text-amber-400'
                              : 'text-emerald-400'
                          }`}
                        >
                          {p.quota!.bottleneck_pct}% remaining
                        </span>
                      </div>

                      <div className="h-1.5 w-full rounded-full bg-secondary overflow-hidden">
                        <div
                          className={`h-full rounded-full transition-all ${
                            p.quota!.is_exhausted
                              ? 'bg-rose-500'
                              : p.quota!.bottleneck_pct < 20
                              ? 'bg-amber-500'
                              : 'bg-primary'
                          }`}
                          style={{ width: `${Math.min(100, Math.max(0, p.quota!.bottleneck_pct))}%` }}
                        />
                      </div>

                      <p className="text-[11px] text-muted-foreground truncate">
                        {p.quota!.summary}
                      </p>
                    </div>
                  )}
                </CardContent>

                <CardFooter className="pt-0 border-t border-border/30 mt-auto">
                  <div className="text-[11px] text-muted-foreground pt-3 w-full flex justify-between items-center">
                    <span>AIM Profile Sandbox</span>
                    <span className="font-mono text-[10px] text-muted-foreground/60">
                      agent/{p.agent}
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
