import * as React from 'react';
import {
  ProfileDTO,
  CreateProfileRequest,
  SessionDTO,
  MCPServerDTO,
  getProfiles,
  createProfile,
  deleteProfile,
  updateProfileConfig,
  getSessions,
  resumeSession,
  getMcpServers,
} from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
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
  Key,
  Mail,
  User,
  Folder,
  RefreshCw,
  Gauge,
  Bot,
  Terminal,
  Server,
  Layers,
  Search,
  ArrowLeft,
  Check,
  Copy,
  Cpu,
  ChevronRight,
  ChevronDown,
} from 'lucide-react';
import { ResumeModal } from '@/components/ResumeModal';
import { DeleteProfileModal } from '@/components/DeleteProfileModal';
import { AgentBadge } from '@/components/AgentBadge';

interface ProfilesViewProps {
  selectedProfile: string;
  onSelectProfile: (profile: string) => void;
  onProfilesLoaded?: (profiles: ProfileDTO[]) => void;
}

export function ProfilesView({
  selectedProfile,
  onSelectProfile,
  onProfilesLoaded,
}: ProfilesViewProps) {
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

  // Profile Detail Sub-tab
  const [profileSubTab, setProfileSubTab] = React.useState<'overview' | 'sessions' | 'mcp'>('overview');
  const [profileSessions, setProfileSessions] = React.useState<SessionDTO[]>([]);
  const [sessionsLoading, setSessionsLoading] = React.useState(false);
  const [sessionSearch, setSessionSearch] = React.useState('');
  const [sessionEngineFilter, setSessionEngineFilter] = React.useState<string>('all');
  const [resumingId, setResumingId] = React.useState<string | null>(null);
  const [flagModalSession, setFlagModalSession] = React.useState<SessionDTO | null>(null);

  // Profile MCP Servers
  const [profileMcpServers, setProfileMcpServers] = React.useState<MCPServerDTO[]>([]);
  const [mcpLoading, setMcpLoading] = React.useState(false);
  const [copiedMcpName, setCopiedMcpName] = React.useState<string | null>(null);

  // Updating profile config in-flight
  const [updatingConfig, setUpdatingConfig] = React.useState(false);

  const { toast } = useToast();

  const fetchProfiles = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getProfiles();
      setProfiles(data);
      onProfilesLoaded?.(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch profiles';
      setError(msg);
      // Fallback sample data in dev if backend server isn't reachable yet
      const fallbackData: ProfileDTO[] = [
        {
          agent: 'claude',
          name: 'work',
          path: '/Users/nemesis/.aim/profiles/work',
          has_credentials: true,
          mcp_global: true,
          plugins_global: true,
          account: {
            email: 'dev@company.com',
            name: 'Work Profile',
            auth_method: 'Google OAuth (Workspace)',
          },
          quota: {
            bottleneck_pct: 68,
            summary: '5h limit: 68% remaining',
            is_exhausted: false,
          },
        },
        {
          agent: 'codex',
          name: 'office',
          path: '/Users/nemesis/.aim/profiles/office',
          has_credentials: true,
          mcp_global: true,
          plugins_global: true,
          account: {
            email: 'adrij@hevodata.com',
            name: 'Adrij Shikhar',
            auth_method: 'ChatGPT (Self Serve_business_prolite)',
          },
          quota: {
            bottleneck_pct: 100,
            summary: 'Weekly Limit: 100%',
            is_exhausted: false,
          },
        },
        {
          agent: 'antigravity',
          name: 'bby',
          path: '/Users/nemesis/.aim/profiles/bby',
          has_credentials: true,
          mcp_global: true,
          plugins_global: true,
          account: {
            email: 'bhomika1.1.2000@gmail.com',
            name: 'bby',
            auth_method: 'Google OAuth (Consumer)',
          },
          quota: {
            bottleneck_pct: 0,
            summary: 'Gemini (5h: 0% [1h 29m], wk: 22%)',
            is_exhausted: true,
          },
        },
        {
          agent: 'antigravity',
          name: 'bot',
          path: '/Users/nemesis/.aim/profiles/bot',
          has_credentials: true,
          mcp_global: true,
          plugins_global: false,
          account: {
            email: 'adrijbot@gmail.com',
            name: 'bot',
            auth_method: 'Google OAuth (Consumer)',
          },
          quota: {
            bottleneck_pct: 55,
            summary: 'Gemini (5h: 77% [3h 37m], wk: 55%)',
            is_exhausted: false,
          },
        },
      ];
      setProfiles(fallbackData);
      onProfilesLoaded?.(fallbackData);
    } finally {
      setLoading(false);
    }
  }, [onProfilesLoaded]);

  React.useEffect(() => {
    fetchProfiles();
  }, [fetchProfiles]);

  // Fetch sessions & MCP servers for the selected profile
  React.useEffect(() => {
    if (selectedProfile && selectedProfile !== 'all') {
      setSessionsLoading(true);
      getSessions({ profile: selectedProfile })
        .then((data) => setProfileSessions(data))
        .catch(() => {
          // Dev fallback sessions for selected profile
          setProfileSessions([
            {
              id: 'sess-89412e',
              agent: 'antigravity',
              profile: selectedProfile,
              title: 'Refactor auth middleware and token validation',
              cwd: '/Users/developer/repos/enterprise-api',
              goal: 'Migrate legacy token handlers to modern RS256 JWT checks and test coverage',
              turns: 24,
              updated_at: new Date(Date.now() - 1000 * 60 * 18).toISOString(),
              is_active: true,
            },
            {
              id: 'sess-99014b',
              agent: 'codex',
              profile: selectedProfile,
              title: 'Autonomous multi-file refactor',
              cwd: '/Users/developer/Projects/catalyst',
              goal: 'Decouple session runners and add isolated state management hooks',
              turns: 58,
              updated_at: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
              is_active: true,
            },
          ]);
        })
        .finally(() => setSessionsLoading(false));

      setMcpLoading(true);
      getMcpServers(selectedProfile)
        .then((data) => setProfileMcpServers(data))
        .catch(() => {
          setProfileMcpServers([
            {
              name: 'claude-mem',
              command: 'npx',
              args: ['-y', '@claude-mem/server'],
              env: { MEMORY_DIR: '~/.claude-mem' },
              scope: 'global',
            },
            {
              name: 'playwright',
              command: 'npx',
              args: ['-y', '@playwright/mcp@latest'],
              scope: 'global',
            },
            {
              name: 'hevo-services',
              command: 'python3',
              args: ['-m', 'hevo_services.mcp'],
              env: { ENVIRONMENT: 'nonprod' },
              scope: selectedProfile,
            },
          ]);
        })
        .finally(() => setMcpLoading(false));
    }
  }, [selectedProfile]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Profile name cannot be empty.',
        variant: 'destructive',
      });
      return;
    }

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

  // Delete confirmation modal state
  const [deleteTarget, setDeleteTarget] = React.useState<{ agent: string; name: string } | null>(null);
  const [isDeleting, setIsDeleting] = React.useState(false);

  const openDeleteDialog = (agent: string, name: string) => {
    setDeleteTarget({ agent, name });
  };

  const handleConfirmDelete = async () => {
    if (!deleteTarget) return;
    const { agent, name } = deleteTarget;
    setIsDeleting(true);
    try {
      await deleteProfile(agent, name);
      toast({
        title: 'Profile Deleted',
        description: `Removed profile "${name}".`,
        variant: 'success',
      });
      setDeleteTarget(null);
      if (selectedProfile === name) {
        onSelectProfile('all');
      }
      fetchProfiles();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to delete profile';
      toast({
        title: 'Deletion Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setIsDeleting(false);
    }
  };

  const handleToggleMcpGlobal = async (profileName: string, currentVal: boolean) => {
    const nextVal = !currentVal;
    setUpdatingConfig(true);
    try {
      await updateProfileConfig(profileName, { mcp_global: nextVal });
      setProfiles((prev) =>
        prev.map((p) => (p.name === profileName ? { ...p, mcp_global: nextVal } : p))
      );
      toast({
        title: 'MCP Configuration Updated',
        description: `Global MCP inheritance ${nextVal ? 'enabled' : 'disabled'} for "${profileName}".`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to update config';
      toast({
        title: 'Update Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setUpdatingConfig(false);
    }
  };

  const handleTogglePluginsGlobal = async (profileName: string, currentVal: boolean) => {
    const nextVal = !currentVal;
    setUpdatingConfig(true);
    try {
      await updateProfileConfig(profileName, { plugins_global: nextVal });
      setProfiles((prev) =>
        prev.map((p) => (p.name === profileName ? { ...p, plugins_global: nextVal } : p))
      );
      toast({
        title: 'Plugin Configuration Updated',
        description: `Global plugins inheritance ${nextVal ? 'enabled' : 'disabled'} for "${profileName}".`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to update config';
      toast({
        title: 'Update Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setUpdatingConfig(false);
    }
  };

  const handleResumeSession = async (s: SessionDTO, flags?: string[], customFlags?: string) => {
    // If the session is currently active and no explicit flags were provided,
    // open the ResumeModal so the user can choose --force or --fork instead of hanging on CLI stdin
    if (s.is_active && (!flags || flags.length === 0) && !customFlags) {
      setFlagModalSession(s);
      return;
    }

    setResumingId(s.id);
    try {
      const res = await resumeSession({
        agent: s.agent,
        profile: s.profile,
        session_id: s.id,
        flags: flags,
        custom_flags: customFlags,
      });
      const flagsDesc = flags && flags.length > 0 ? ` with ${flags.join(' ')}` : '';
      toast({
        title: 'Terminal Session Launched',
        description: res.message || `Resumed session ${s.id}${flagsDesc} in an interactive terminal.`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to launch terminal';
      toast({
        title: 'Resume Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setResumingId(null);
    }
  };

  const copyMcpConfig = (srv: MCPServerDTO) => {
    const jsonStr = JSON.stringify(srv, null, 2);
    navigator.clipboard.writeText(jsonStr);
    setCopiedMcpName(srv.name);
    toast({
      title: 'Copied Config',
      description: `Copied JSON configuration for "${srv.name}" to clipboard.`,
    });
    setTimeout(() => {
      setCopiedMcpName(null);
    }, 2000);
  };

  const getAgentBadge = (agent: string) => {
    return <AgentBadge agent={agent} />;
  };

  const getQuotaProgressColor = (pct: number) => {
    if (pct <= 0) return 'bg-rose-500';
    return 'bg-[#ededed]';
  };

  const getQuotaTextColor = (pct: number) => {
    if (pct <= 0) return 'text-rose-400';
    return 'text-[#ededed]';
  };

  const currentProfile = React.useMemo(() => {
    if (!selectedProfile || selectedProfile === 'all') return null;
    return profiles.find((p) => p.name === selectedProfile) || null;
  }, [profiles, selectedProfile]);

  const filteredSessions = React.useMemo(() => {
    let list = profileSessions;
    if (sessionEngineFilter !== 'all') {
      list = list.filter((s) => s.agent.toLowerCase() === sessionEngineFilter.toLowerCase());
    }
    if (!sessionSearch.trim()) return list;
    const q = sessionSearch.toLowerCase();
    return list.filter(
      (s) =>
        s.id.toLowerCase().includes(q) ||
        s.title.toLowerCase().includes(q) ||
        s.goal.toLowerCase().includes(q) ||
        s.cwd.toLowerCase().includes(q)
    );
  }, [profileSessions, sessionSearch, sessionEngineFilter]);

  // ==========================================
  // VIEW A: SPECIFIC PROFILE HUB VIEW
  // ==========================================
  if (selectedProfile !== 'all' && currentProfile) {
    const isMcpGlobalOn = currentProfile.mcp_global !== false;
    const isPluginsGlobalOn = currentProfile.plugins_global !== false;

    return (
      <div className="space-y-6">
        {/* Profile Header Banner */}
        <div className="rounded-geist border border-[#262626] bg-[#111111] p-5 space-y-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onSelectProfile('all')}
                className="h-8 rounded-geist text-xs font-mono text-[#888888] hover:text-[#ededed] border-[#2a2a2a] bg-[#141414] hover:bg-[#1f1f1f]"
              >
                <ArrowLeft className="h-3.5 w-3.5 mr-1" />
                All Profiles
              </Button>
              <div className="flex items-center gap-2 flex-wrap">
                <h1 className="text-lg font-semibold tracking-tight text-[#ededed]">
                  {currentProfile.name}
                </h1>
                {currentProfile.adapters && currentProfile.adapters.length > 0 ? (
                  currentProfile.adapters.map((ad) => (
                    <React.Fragment key={ad.agent}>
                      {getAgentBadge(ad.agent)}
                    </React.Fragment>
                  ))
                ) : (
                  getAgentBadge(currentProfile.agent)
                )}
                {currentProfile.has_credentials ? (
                  <Badge
                    variant="outline"
                    className="bg-[#141414] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                    <span>Authenticated</span>
                  </Badge>
                ) : (
                  <Badge
                    variant="outline"
                    className="bg-[#141414] text-[#888888] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                    <span>Needs Auth</span>
                  </Badge>
                )}
              </div>
            </div>

            <div className="flex items-center gap-1.5 self-end sm:self-auto">
              <Button
                variant="outline"
                size="icon"
                onClick={fetchProfiles}
                disabled={loading}
                title="Refresh profile"
                aria-label="Refresh profile"
                className="h-8 w-8 rounded-geist border-[#2a2a2a] text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a]"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => openDeleteDialog(currentProfile.agent, currentProfile.name)}
                title="Delete profile"
                aria-label="Delete profile"
                className="h-8 w-8 rounded-geist border-[#2a2a2a] text-[#888888] hover:text-red-400 hover:bg-red-500/10 hover:border-red-500/30"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Path and identity sub-bar */}
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 pt-2 border-t border-[#1f1f1f] text-xs font-mono text-[#888888]">
            <div className="flex items-center gap-1.5">
              <Folder className="h-3.5 w-3.5 text-[#666666]" />
              <span className="text-[#ededed]">{currentProfile.path}</span>
            </div>
            {currentProfile.account?.email && (
              <div className="flex items-center gap-1.5">
                <Mail className="h-3.5 w-3.5 text-[#666666]" />
                <span>{currentProfile.account.email}</span>
              </div>
            )}
            {currentProfile.account?.auth_method && (
              <div className="flex items-center gap-1.5">
                <Key className="h-3.5 w-3.5 text-[#666666]" />
                <span>{currentProfile.account.auth_method}</span>
              </div>
            )}
          </div>
        </div>

        {/* Profile Internal Sub-Navigation */}
        <div className="flex items-center gap-2 border-b border-[#262626] pb-3">
          <button
            onClick={() => setProfileSubTab('overview')}
            className={`text-xs px-3 py-1.5 rounded-geist font-medium transition-all duration-150 cursor-pointer flex items-center gap-1.5 ${
              profileSubTab === 'overview'
                ? 'bg-[#222222] text-[#ededed] border border-[#333333] shadow-sm'
                : 'text-[#888888] hover:text-[#ededed] hover:bg-[#161616] border border-transparent'
            }`}
          >
            <User className="h-3.5 w-3.5" />
            Overview
          </button>
          <button
            onClick={() => setProfileSubTab('sessions')}
            className={`text-xs px-3 py-1.5 rounded-geist font-medium transition-all duration-150 cursor-pointer flex items-center gap-1.5 ${
              profileSubTab === 'sessions'
                ? 'bg-[#222222] text-[#ededed] border border-[#333333] shadow-sm'
                : 'text-[#888888] hover:text-[#ededed] hover:bg-[#161616] border border-transparent'
            }`}
          >
            <Terminal className="h-3.5 w-3.5" />
            Sessions ({profileSessions.length})
          </button>
          <button
            onClick={() => setProfileSubTab('mcp')}
            className={`text-xs px-3 py-1.5 rounded-geist font-medium transition-all duration-150 cursor-pointer flex items-center gap-1.5 ${
              profileSubTab === 'mcp'
                ? 'bg-[#222222] text-[#ededed] border border-[#333333] shadow-sm'
                : 'text-[#888888] hover:text-[#ededed] hover:bg-[#161616] border border-transparent'
            }`}
          >
            <Server className="h-3.5 w-3.5" />
            MCP & Plugins
          </button>
        </div>

        {/* SUB-PANEL 1: OVERVIEW */}
        {profileSubTab === 'overview' && (
          <div className="space-y-4">
            {/* Storage & Shared Sandbox Card */}
            <Card className="border border-[#262626] bg-[#111111] rounded-geist">
              <CardHeader className="pb-3 p-5">
                <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                  <Folder className="h-4 w-4 text-[#888888]" />
                  Profile Environment & Storage Sandbox
                </CardTitle>
                <CardDescription className="text-xs text-[#888888]">
                  Filesystem sandbox and isolated configuration home directory.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0 text-xs font-mono space-y-2">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1 py-1.5 border-b border-[#1f1f1f]">
                  <span className="text-[#888888]">Storage Sandbox:</span>
                  <span className="text-[#ededed] truncate max-w-md" title={currentProfile.path}>
                    {currentProfile.path}
                  </span>
                </div>
                <div className="flex items-center justify-between py-1.5 border-b border-[#1f1f1f]">
                  <span className="text-[#888888]">Connected Engines:</span>
                  <div className="flex items-center gap-1.5 flex-wrap">
                    {currentProfile.adapters && currentProfile.adapters.length > 0 ? (
                      currentProfile.adapters.map((ad) => (
                        <React.Fragment key={ad.agent}>
                          {getAgentBadge(ad.agent)}
                        </React.Fragment>
                      ))
                    ) : (
                      getAgentBadge(currentProfile.agent)
                    )}
                  </div>
                </div>
                <div className="flex items-center justify-between py-1.5">
                  <span className="text-[#888888]">Sandbox Tool Inheritance:</span>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-[10px] font-mono bg-[#141414] border-[#262626] text-[#888888]">
                      MCP: {isMcpGlobalOn ? 'Global Shared' : 'Strict Isolated'}
                    </Badge>
                    <Badge variant="outline" className="text-[10px] font-mono bg-[#141414] border-[#262626] text-[#888888]">
                      Plugins: {isPluginsGlobalOn ? 'Global Shared' : 'Strict Isolated'}
                    </Badge>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Individual Engine Cards Grid */}
            <div className="space-y-2">
              <h3 className="text-xs font-mono font-medium uppercase tracking-wider text-[#888888] pt-1">
                Connected Agent Adapters ({currentProfile.adapters?.length || 1})
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {(currentProfile.adapters && currentProfile.adapters.length > 0
                  ? currentProfile.adapters
                  : [
                      {
                        agent: currentProfile.agent,
                        has_credentials: currentProfile.has_credentials,
                        account: currentProfile.account,
                        quota: currentProfile.quota,
                      },
                    ]
                ).map((ad) => (
                  <Card key={ad.agent} className="border border-[#262626] bg-[#111111] rounded-geist flex flex-col justify-between">
                    <CardHeader className="pb-3 p-5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          {getAgentBadge(ad.agent)}
                          <span className="text-xs font-mono text-[#888888]">Engine</span>
                        </div>
                        {ad.has_credentials ? (
                          <Badge
                            variant="outline"
                            className="bg-[#141414] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                          >
                            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                            <span>Authenticated</span>
                          </Badge>
                        ) : (
                          <Badge
                            variant="outline"
                            className="bg-[#141414] text-[#888888] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                          >
                            <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                            <span>Needs Auth</span>
                          </Badge>
                        )}
                      </div>
                      <CardDescription className="text-xs text-[#888888] pt-1">
                        {ad.account?.email || (ad.has_credentials ? 'Active CLI credentials detected' : 'No credentials detected')}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4 p-5 pt-0 text-xs font-mono">
                      <div className="p-3 rounded-geist bg-[#0e0e0e] border border-[#1f1f1f] space-y-1.5">
                        <div className="flex justify-between items-center text-[#888888]">
                          <span>Account:</span>
                          <span className="text-[#ededed] font-medium truncate max-w-[180px]">
                            {ad.account?.name || currentProfile.name}
                          </span>
                        </div>
                        {ad.account?.auth_method && (
                          <div className="flex justify-between items-center text-[#888888]">
                            <span>Auth Method:</span>
                            <span className="text-[#ededed] truncate max-w-[180px]">{ad.account.auth_method}</span>
                          </div>
                        )}
                      </div>

                      {/* Quota Telemetry for this Engine */}
                      <div className="space-y-2 pt-2 border-t border-[#1f1f1f]">
                        <div className="flex items-center justify-between">
                          <span className="text-[11px] uppercase tracking-wider text-[#888888] flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Live Quota Telemetry
                          </span>
                          <span
                            className={`font-mono text-xs font-medium tabular-nums ${getQuotaTextColor(
                              ad.quota?.bottleneck_pct ?? 100
                            )}`}
                          >
                            {ad.quota ? `${ad.quota.bottleneck_pct}% remaining` : '100% remaining'}
                          </span>
                        </div>
                        <div className="w-full bg-[#1c1c1c] h-1.5 rounded-full overflow-hidden border border-[#262626]">
                          <div
                            className={`h-full transition-all duration-300 ${getQuotaProgressColor(
                              ad.quota?.bottleneck_pct ?? 100
                            )}`}
                            style={{ width: `${ad.quota?.bottleneck_pct ?? 100}%` }}
                          />
                        </div>
                        <div className="p-2.5 rounded-geist bg-[#0e0e0e] border border-[#262626] text-xs">
                          <div className="text-[#888888] text-[10px] uppercase tracking-wider">Summary</div>
                          <div className="text-[#ededed] text-xs mt-0.5 truncate" title={ad.quota?.summary}>
                            {ad.quota?.summary || 'No active quota bottleneck detected.'}
                          </div>
                        </div>
                        <div className="grid grid-cols-2 gap-2 text-xs font-mono pt-1">
                          <div className="p-2 rounded-geist bg-[#141414] border border-[#222222]">
                            <div className="text-[10px] text-[#888888] uppercase">Status</div>
                            <div className="text-[#ededed] font-medium mt-0.5">
                              {ad.quota?.is_exhausted ? 'Exhausted' : 'Healthy'}
                            </div>
                          </div>
                          <div className="p-2 rounded-geist bg-[#141414] border border-[#222222]">
                            <div className="text-[10px] text-[#888888] uppercase">Reset Policy</div>
                            <div className="text-[#ededed] font-medium mt-0.5">Sliding Window</div>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* SUB-PANEL 2: SESSIONS */}
        {profileSubTab === 'sessions' && (
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
              <div className="flex items-center gap-2 flex-wrap w-full sm:w-auto">
                <div className="relative w-full sm:w-72">
                  <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#666666]" />
                  <Input
                    placeholder="Search sessions in this profile..."
                    className="pl-9 h-8 text-xs rounded-geist border-[#262626] bg-[#0e0e0e]"
                    value={sessionSearch}
                    onChange={(e) => setSessionSearch(e.target.value)}
                  />
                </div>
                {currentProfile.adapters && currentProfile.adapters.length > 1 && (
                  <div className="flex items-center gap-1 bg-[#141414] p-0.5 rounded-geist border border-[#262626]">
                    <button
                      onClick={() => setSessionEngineFilter('all')}
                      className={`text-[11px] font-mono px-2 py-0.5 rounded transition-colors cursor-pointer ${
                        sessionEngineFilter === 'all'
                          ? 'bg-[#222222] text-[#ededed] font-medium'
                          : 'text-[#888888] hover:text-[#ededed]'
                      }`}
                    >
                      All ({profileSessions.length})
                    </button>
                    {currentProfile.adapters.map((ad) => {
                      const count = profileSessions.filter((s) => s.agent.toLowerCase() === ad.agent.toLowerCase()).length;
                      return (
                        <button
                          key={ad.agent}
                          onClick={() => setSessionEngineFilter(ad.agent)}
                          className={`text-[11px] font-mono px-2 py-0.5 rounded transition-colors cursor-pointer ${
                            sessionEngineFilter.toLowerCase() === ad.agent.toLowerCase()
                              ? 'bg-[#222222] text-[#ededed] font-medium'
                              : 'text-[#888888] hover:text-[#ededed]'
                          }`}
                        >
                          {ad.agent} ({count})
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
              <span className="text-xs text-[#888888] font-mono">
                {filteredSessions.length} session{filteredSessions.length !== 1 ? 's' : ''} in {currentProfile.name}
              </span>
            </div>

            <div className="rounded-geist border border-[#262626] bg-[#111111] overflow-hidden">
              <Table>
                <TableHeader className="bg-[#141414] border-b border-[#262626]">
                  <TableRow className="border-b border-[#262626] hover:bg-transparent">
                    <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Engine</TableHead>
                    <TableHead className="min-w-[200px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Directory (CWD)</TableHead>
                    <TableHead className="min-w-[280px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Goal / Title</TableHead>
                    <TableHead className="w-[80px] text-center font-mono text-[11px] uppercase tracking-wider text-[#888888]">Turns</TableHead>
                    <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Updated</TableHead>
                    <TableHead className="w-[120px] text-right font-mono text-[11px] uppercase tracking-wider text-[#888888]">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sessionsLoading ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-32 text-center text-[#888888]">
                        <RefreshCw className="h-4 w-4 animate-spin text-[#ededed] mx-auto mb-1" />
                        <span className="font-mono text-xs">Loading profile sessions...</span>
                      </TableCell>
                    </TableRow>
                  ) : filteredSessions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-32 text-center text-[#888888]">
                        <Terminal className="h-6 w-6 text-[#555555] mx-auto mb-1 opacity-60" />
                        <div className="text-xs font-medium text-[#ededed]">No sessions found</div>
                        <div className="text-[11px] text-[#777777]">Run an agent in this profile to start a session.</div>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredSessions.map((s) => (
                      <TableRow key={s.id} className="hover:bg-[#161616] transition-colors border-b border-[#1f1f1f]">
                        <TableCell>{getAgentBadge(s.agent)}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1.5 text-xs text-[#888888] max-w-[240px]">
                            <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                            <span className="font-mono truncate" title={s.cwd}>
                              {s.cwd}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="space-y-0.5 max-w-[340px]">
                            <div className="font-medium text-xs text-[#ededed] truncate" title={s.title}>
                              {s.title}
                            </div>
                            <div className="text-[11px] text-[#777777] truncate" title={s.goal}>
                              {s.goal}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-center font-mono text-xs text-[#888888] tabular-nums">
                          {s.turns}
                        </TableCell>
                        <TableCell className="font-mono text-xs text-[#888888]">
                          {new Date(s.updated_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="inline-flex items-center rounded-geist border border-[#333333] bg-[#161616] overflow-hidden shadow-sm">
                            <button
                              type="button"
                              className="h-7 px-2.5 text-xs font-mono font-medium hover:bg-[#222222] text-[#ededed] inline-flex items-center gap-1.5 transition-colors cursor-pointer border-r border-[#262626] disabled:opacity-50"
                              title="Resume session in terminal"
                              disabled={resumingId === s.id}
                              onClick={() => handleResumeSession(s)}
                            >
                              {resumingId === s.id ? (
                                <RefreshCw className="h-3 w-3 animate-spin text-[#ededed]" />
                              ) : (
                                <Terminal className="h-3 w-3 text-[#ededed]" />
                              )}
                              Resume
                            </button>
                            <button
                              type="button"
                              className="h-7 px-1.5 hover:bg-[#222222] text-[#888888] hover:text-[#ededed] inline-flex items-center transition-colors cursor-pointer"
                              title="Configure flags & options (--exact, --catalyst, --fork, etc.)"
                              onClick={() => setFlagModalSession(s)}
                            >
                              <ChevronDown className="h-3 w-3" />
                            </button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        )}

        {/* SUB-PANEL 3: MCP & PLUGINS CONFIGURATION */}
        {profileSubTab === 'mcp' && (
          <div className="space-y-6">
            {currentProfile.adapters && currentProfile.adapters.length > 1 && (
              <div className="flex items-center gap-2.5 px-4 py-3 rounded-geist bg-[#141414] border border-[#262626] text-xs font-mono text-[#888888]">
                <Cpu className="h-4 w-4 text-[#ededed] shrink-0" />
                <span>
                  <strong className="text-[#ededed] font-medium">Multi-Engine Shared Sandbox:</strong> Tools and plugins configured below are mounted across all connected engines ({currentProfile.adapters.map((a) => a.agent).join(', ')}) operating inside the <code className="text-[#ededed]">{currentProfile.name}</code> sandbox.
                </span>
              </div>
            )}
            {/* Toggles Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* MCP Configuration Card */}
              <Card className="border border-[#262626] bg-[#111111] rounded-geist">
                <CardHeader className="p-5 pb-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                      <Cpu className="h-4 w-4 text-[#888888]" />
                      Model Context Protocol (MCP)
                    </CardTitle>
                    <Switch
                      checked={isMcpGlobalOn}
                      disabled={updatingConfig}
                      onCheckedChange={() => handleToggleMcpGlobal(currentProfile.name, isMcpGlobalOn)}
                    />
                  </div>
                  <CardDescription className="text-xs text-[#888888] pt-1">
                    Inherit Global MCP Servers (<code className="text-[#ededed]">mcp_global</code>)
                  </CardDescription>
                </CardHeader>
                <CardContent className="p-5 pt-0 space-y-3">
                  <p className="text-xs text-[#888888] leading-relaxed">
                    When enabled, host MCP servers (e.g. <code>claude-mem</code>, <code>playwright</code>, <code>snyk</code>)
                    are automatically bridged into this profile's session environments. Turn off to enforce strict sandbox isolation.
                  </p>
                  <div className="flex items-center gap-2 text-xs font-mono">
                    <span className="text-[#888888]">Status:</span>
                    <Badge
                      variant="outline"
                      className="text-[11px] font-mono bg-[#141414] border-[#262626] text-[#a1a1a1] flex items-center gap-1.5 px-2 py-0.5"
                    >
                      <span className={`h-1.5 w-1.5 rounded-full ${isMcpGlobalOn ? 'bg-emerald-500' : 'bg-neutral-600'}`} />
                      {isMcpGlobalOn ? 'Global Inheritance Active' : 'Isolated Mode'}
                    </Badge>
                  </div>
                </CardContent>
              </Card>

              {/* Plugin Configuration Card */}
              <Card className="border border-[#262626] bg-[#111111] rounded-geist">
                <CardHeader className="p-5 pb-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                      <Layers className="h-4 w-4 text-[#888888]" />
                      Agent Plugins & Skills
                    </CardTitle>
                    <Switch
                      checked={isPluginsGlobalOn}
                      disabled={updatingConfig}
                      onCheckedChange={() => handleTogglePluginsGlobal(currentProfile.name, isPluginsGlobalOn)}
                    />
                  </div>
                  <CardDescription className="text-xs text-[#888888] pt-1">
                    Inherit Global Plugins (<code className="text-[#ededed]">plugins_global</code>)
                  </CardDescription>
                </CardHeader>
                <CardContent className="p-5 pt-0 space-y-3">
                  <p className="text-xs text-[#888888] leading-relaxed">
                    When enabled, host plugins, custom tools, and extension skills are merged into this profile's agent sessions.
                    Turn off if this profile requires custom plugins.
                  </p>
                  <div className="flex items-center gap-2 text-xs font-mono">
                    <span className="text-[#888888]">Status:</span>
                    <Badge
                      variant="outline"
                      className="text-[11px] font-mono bg-[#141414] border-[#262626] text-[#a1a1a1] flex items-center gap-1.5 px-2 py-0.5"
                    >
                      <span className={`h-1.5 w-1.5 rounded-full ${isPluginsGlobalOn ? 'bg-emerald-500' : 'bg-neutral-600'}`} />
                      {isPluginsGlobalOn ? 'Plugins Active' : 'Plugins Disabled'}
                    </Badge>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Active Servers List for Profile */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-mono font-medium uppercase tracking-wider text-[#888888]">
                  Active MCP Servers Available to {currentProfile.name}
                </h3>
                <span className="text-xs font-mono text-[#888888]">
                  {profileMcpServers.length} server{profileMcpServers.length !== 1 ? 's' : ''}
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {mcpLoading ? (
                  <div className="col-span-full h-24 rounded-geist border border-[#262626] bg-[#111111]/40 flex items-center justify-center font-mono text-xs text-[#888888]">
                    Loading MCP servers...
                  </div>
                ) : (
                  profileMcpServers.map((srv) => {
                    const isGlobal = srv.scope === 'global';
                    return (
                      <Card key={srv.name} className="border border-[#262626] bg-[#111111] rounded-geist">
                        <CardHeader className="p-4 pb-2">
                          <div className="flex items-center justify-between">
                            <CardTitle className="text-xs font-semibold text-[#ededed] font-mono">
                              {srv.name}
                            </CardTitle>
                            <Badge
                              variant="outline"
                              className="text-[10px] font-mono rounded border-[#2e2e2e] bg-[#161616] text-[#888888] px-1.5 py-0"
                            >
                              {isGlobal ? 'Global' : 'Profile'}
                            </Badge>
                          </div>
                        </CardHeader>
                        <CardContent className="p-4 pt-0 space-y-2">
                          <div className="font-mono text-[11px] p-1.5 rounded-geist bg-[#0e0e0e] border border-[#262626] text-[#ededed] truncate">
                            <span className="text-[#ededed] font-medium">{srv.command}</span>{' '}
                            <span className="text-[#888888]">{srv.args.join(' ')}</span>
                          </div>
                          <div className="flex justify-end pt-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 text-[11px] font-mono text-[#888888] hover:text-[#ededed]"
                              onClick={() => copyMcpConfig(srv)}
                            >
                              {copiedMcpName === srv.name ? (
                                <Check className="h-3 w-3 mr-1 text-[#ededed]" />
                              ) : (
                                <Copy className="h-3 w-3 mr-1" />
                              )}
                              {copiedMcpName === srv.name ? 'Copied' : 'JSON'}
                            </Button>
                          </div>
                        </CardContent>
                      </Card>
                    );
                  })
                )}
              </div>
            </div>
          </div>
        )}

        <ResumeModal
          open={!!flagModalSession}
          onOpenChange={(open) => !open && setFlagModalSession(null)}
          session={flagModalSession}
          isResuming={!!resumingId}
          onLaunch={async (req) => {
            if (flagModalSession) {
              await handleResumeSession(flagModalSession, req.flags, req.custom_flags);
            }
          }}
        />

        <DeleteProfileModal
          open={!!deleteTarget}
          onOpenChange={(open) => !open && setDeleteTarget(null)}
          target={deleteTarget}
          isDeleting={isDeleting}
          onConfirm={handleConfirmDelete}
        />
      </div>
    );
  }

  // ==========================================
  // VIEW B: ALL PROFILES DIRECTORY (GRID)
  // ==========================================
  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-geist border border-amber-500/30 bg-amber-500/10 px-3.5 py-2 text-xs text-amber-300 font-mono flex items-center justify-between">
          <span>Notice: {error}</span>
          <button
            type="button"
            onClick={() => setError(null)}
            className="text-[#888888] hover:text-white ml-2 text-xs cursor-pointer"
          >
            ✕
          </button>
        </div>
      )}
      {/* Top action row */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-2.5">
          <h2 className="text-sm font-semibold tracking-tight text-[#ededed]">
            All Agent Profiles
          </h2>
          <span className="font-mono text-[11px] px-2 py-0.5 rounded-full border border-[#262626] bg-[#141414] text-[#888888] tabular-nums">
            {profiles.length} active
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

      {/* Profiles Grid */}
      {loading && profiles.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-64 rounded-geist border border-[#262626] bg-[#111111]/40 animate-pulse p-6" />
          ))}
        </div>
      ) : profiles.length === 0 ? (
        <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0e0e0e]/50">
          <Bot className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-[#ededed]">No profiles found</h3>
          <p className="text-xs text-[#888888] mt-1 max-w-sm mx-auto">
            Click "Add Profile" to create an isolated environment for your agent.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {profiles.map((p) => {
            const isMcpOn = p.mcp_global !== false;
            const isPluginsOn = p.plugins_global !== false;
            return (
              <Card
                key={p.name}
                className="group relative overflow-hidden transition-all duration-150 border border-[#262626] bg-[#111111] hover:border-[#383838] rounded-geist flex flex-col justify-between cursor-pointer"
                onClick={() => onSelectProfile(p.name)}
              >
                <div>
                  <CardHeader className="pb-3 p-5">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-1.5 flex-wrap">
                        {p.adapters && p.adapters.length > 0 ? (
                          p.adapters.map((ad) => (
                            <React.Fragment key={ad.agent}>
                              {getAgentBadge(ad.agent)}
                            </React.Fragment>
                          ))
                        ) : (
                          getAgentBadge(p.agent)
                        )}
                      </div>
                      {p.has_credentials ? (
                        <Badge
                          variant="outline"
                          className="bg-[#141414] text-[#a1a1a1] border-[#262626] text-[11px] font-mono tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                          <span>Authenticated</span>
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="bg-[#141414] text-[#888888] border-[#262626] text-[11px] font-mono tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-md"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                          <span>Needs Auth</span>
                        </Badge>
                      )}
                    </div>

                    <div className="pt-2">
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-base font-semibold text-[#ededed] group-hover:text-white transition-colors">
                          {p.name}
                        </CardTitle>
                        <Button
                          variant="ghost"
                          size="icon"
                          title={`Delete profile ${p.name}`}
                          aria-label={`Delete profile ${p.name}`}
                          className="h-7 w-7 text-[#666666] hover:text-red-400 rounded-geist transition-colors"
                          onClick={(e) => {
                            e.stopPropagation();
                            openDeleteDialog(p.agent, p.name);
                          }}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                      <CardDescription className="text-xs text-[#888888] font-mono truncate pt-0.5 flex items-center gap-1">
                        <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                        <span className="truncate">{p.path}</span>
                      </CardDescription>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3 pb-4 p-5 pt-0">
                    {/* Account Info Box */}
                    <div className="rounded-geist border border-[#1f1f1f] bg-[#0e0e0e] p-3 text-xs space-y-1.5 font-mono">
                      <div className="flex items-center justify-between text-[#888888]">
                        <span className="flex items-center gap-1.5">
                          <User className="h-3 w-3 text-[#666666]" /> Account
                        </span>
                        <span className="text-[#ededed] font-medium truncate max-w-[140px]">
                          {p.account?.name || p.name}
                        </span>
                      </div>
                      {p.account?.email && (
                        <div className="flex items-center justify-between text-[#888888]">
                          <span className="flex items-center gap-1.5">
                            <Mail className="h-3 w-3 text-[#666666]" /> Email
                          </span>
                          <span className="text-[#ededed] truncate max-w-[160px]">
                            {p.account.email}
                          </span>
                        </div>
                      )}
                      {p.account?.auth_method && (
                        <div className="flex items-center justify-between text-[#888888]">
                          <span className="flex items-center gap-1.5">
                            <Key className="h-3 w-3 text-[#666666]" /> Auth
                          </span>
                          <span className="text-[#ededed] truncate max-w-[140px]">
                            {p.account.auth_method}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Quota Telemetry */}
                    {p.adapters && p.adapters.length > 1 ? (
                      <div className="space-y-2 pt-1 border-t border-[#1f1f1f]">
                        <div className="flex items-center justify-between text-[11px] font-mono uppercase tracking-wider text-[#888888]">
                          <span className="flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Engine Quotas
                          </span>
                          <span>{p.adapters.length} Engines</span>
                        </div>
                        {p.adapters.map((ad) => (
                          <div key={ad.agent} className="space-y-1">
                            <div className="flex items-center justify-between text-[11px] font-mono">
                              <span className="text-[#888888]">{ad.agent}</span>
                              <span
                                className={`tabular-nums ${getQuotaTextColor(
                                  ad.quota?.bottleneck_pct ?? 100
                                )}`}
                              >
                                {ad.quota ? `${ad.quota.bottleneck_pct}%` : '100%'}
                              </span>
                            </div>
                            <div className="w-full bg-[#1c1c1c] h-1 rounded-full overflow-hidden border border-[#262626]">
                              <div
                                className={`h-full transition-all duration-300 ${getQuotaProgressColor(
                                  ad.quota?.bottleneck_pct ?? 100
                                )}`}
                                style={{ width: `${ad.quota?.bottleneck_pct ?? 100}%` }}
                              />
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-[11px] font-mono uppercase tracking-wider text-[#888888] flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Quota Telemetry
                          </span>
                          <span
                            className={`font-mono text-xs font-medium tabular-nums ${getQuotaTextColor(
                              p.quota?.bottleneck_pct ?? 100
                            )}`}
                          >
                            {p.quota ? `${p.quota.bottleneck_pct}% remaining` : '100% remaining'}
                          </span>
                        </div>
                        <div className="w-full bg-[#1c1c1c] h-1.5 rounded-full overflow-hidden border border-[#262626]">
                          <div
                            className={`h-full transition-all duration-300 ${getQuotaProgressColor(
                              p.quota?.bottleneck_pct ?? 100
                            )}`}
                            style={{ width: `${p.quota?.bottleneck_pct ?? 100}%` }}
                          />
                        </div>
                        {p.quota?.summary && (
                          <p className="text-[11px] text-[#666666] font-mono truncate" title={p.quota.summary}>
                            {p.quota.summary}
                          </p>
                        )}
                      </div>
                    )}
                  </CardContent>
                </div>

                {/* Card Footer: MCP & Plugin configuration chips + Inspect Action */}
                <CardFooter className="pt-2 pb-4 px-5 border-t border-[#1f1f1f] flex items-center justify-between text-[11px] font-mono text-[#888888]">
                  <div className="flex items-center gap-1.5">
                    <span className="px-1.5 py-0.5 rounded border border-[#262626] bg-[#141414] text-[#888888] text-[10px] font-mono">
                      mcp: {isMcpOn ? 'on' : 'off'}
                    </span>
                    <span className="px-1.5 py-0.5 rounded border border-[#262626] bg-[#141414] text-[#888888] text-[10px] font-mono">
                      plugins: {isPluginsOn ? 'on' : 'off'}
                    </span>
                  </div>

                  <span className="flex items-center text-[#ededed] group-hover:text-white transition-colors">
                    Inspect <ChevronRight className="h-3 w-3 ml-0.5 group-hover:translate-x-0.5 transition-transform" />
                  </span>
                </CardFooter>
              </Card>
            );
          })}
        </div>
      )}

      <ResumeModal
        open={!!flagModalSession}
        onOpenChange={(open) => !open && setFlagModalSession(null)}
        session={flagModalSession}
        isResuming={!!resumingId}
        onLaunch={async (req) => {
          if (flagModalSession) {
            await handleResumeSession(flagModalSession, req.flags, req.custom_flags);
          }
        }}
      />

      <DeleteProfileModal
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        target={deleteTarget}
        isDeleting={isDeleting}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
