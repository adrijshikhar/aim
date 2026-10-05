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
  getMcpServers,
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
  Key,
  Mail,
  User,
  Folder,
  RefreshCw,
  Gauge,
  Bot,
  Terminal,
  Server,
  ArrowLeft,
  ArrowRight,
  ChevronRight,
} from 'lucide-react';
import { DeleteProfileModal } from '@/components/DeleteProfileModal';
import { AgentBadge } from '@/components/AgentBadge';
import { QuotaBreakdown } from '@/components/QuotaBreakdown';

interface ProfilesViewProps {
  selectedProfile: string;
  onSelectProfile: (profile: string) => void;
  onNavigateToSessions?: (profile: string) => void;
  onNavigateToMcp?: (profile: string) => void;
  onProfilesLoaded?: (profiles: ProfileDTO[]) => void;
}

export function ProfilesView({
  selectedProfile,
  onSelectProfile,
  onNavigateToSessions,
  onNavigateToMcp,
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

  // Profile-specific counts for direct launcher cards
  const [profileSessions, setProfileSessions] = React.useState<SessionDTO[]>([]);
  const [profileMcpServers, setProfileMcpServers] = React.useState<MCPServerDTO[]>([]);
  const [resourcesLoading, setResourcesLoading] = React.useState(false);

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

  // Fetch count of sessions & MCP servers for the selected profile
  React.useEffect(() => {
    if (selectedProfile && selectedProfile !== 'all') {
      setResourcesLoading(true);
      Promise.allSettled([
        getSessions({ profile: selectedProfile }),
        getMcpServers(selectedProfile),
      ]).then(([sessRes, mcpRes]) => {
        if (sessRes.status === 'fulfilled') {
          setProfileSessions(sessRes.value);
        } else {
          setProfileSessions([]);
        }
        if (mcpRes.status === 'fulfilled') {
          setProfileMcpServers(mcpRes.value);
        } else {
          setProfileMcpServers([]);
        }
        setResourcesLoading(false);
      });
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

  const getAgentBadge = (agent: string) => {
    return <AgentBadge agent={agent} />;
  };

  const getProgressColor = (pct: number) => {
    if (pct <= 0) return 'bg-[#ee0000]';
    if (pct <= 30) return 'bg-[#f5a623]';
    return 'bg-[#00e599]';
  };

  const getTextColor = (pct: number) => {
    if (pct <= 0) return 'text-[#ee0000]';
    if (pct <= 30) return 'text-[#f5a623]';
    return 'text-[#ededed]';
  };

  const currentProfile = React.useMemo(() => {
    if (!selectedProfile || selectedProfile === 'all') return null;
    return profiles.find((p) => p.name === selectedProfile) || null;
  }, [profiles, selectedProfile]);

  // ==========================================
  // VIEW A: SPECIFIC PROFILE WORKSPACE VIEW
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
                className="h-8 rounded-geist text-xs font-mono text-[#a1a1a1] hover:text-[#ededed] border-[#262626] bg-[#0a0a0a] hover:bg-[#161616]"
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
                    className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-[#00e599]" />
                    <span>Authenticated</span>
                  </Badge>
                ) : (
                  <Badge
                    variant="outline"
                    className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                  >
                    <span className="h-1.5 w-1.5 rounded-full bg-[#f5a623]" />
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
                className="h-8 w-8 rounded-geist border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616]"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => openDeleteDialog(currentProfile.agent, currentProfile.name)}
                title="Delete profile"
                aria-label="Delete profile"
                className="h-8 w-8 rounded-geist border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ee0000] hover:bg-[#ee0000]/10 hover:border-[#ee0000]/30"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Path and identity sub-bar */}
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 pt-2 border-t border-[#262626] text-xs font-mono text-[#a1a1a1]">
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

        {/* Deep-link Action Cards: Single-View Navigation to Sessions & MCP */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Card 1: Sessions Deep Link */}
          <Card className="border border-[#262626] bg-[#111111] rounded-geist hover:border-[#333333] transition-colors">
            <CardHeader className="p-5 pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                  <Terminal className="h-4 w-4 text-[#a1a1a1]" />
                  Agent Sessions
                </CardTitle>
                <Badge
                  variant="outline"
                  className="text-xs font-mono bg-[#0a0a0a] border-[#262626] text-[#ededed] px-2 py-0.5"
                >
                  {resourcesLoading ? '...' : `${profileSessions.length} sessions`}
                </Badge>
              </div>
              <CardDescription className="text-xs text-[#a1a1a1] pt-1">
                View, filter, and resume interactive CLI terminal sessions for <code className="text-[#ededed]">{currentProfile.name}</code>.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-5 pt-0">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onNavigateToSessions && onNavigateToSessions(currentProfile.name)}
                className="w-full justify-between h-9 rounded-geist border-[#262626] bg-[#0a0a0a] text-xs font-mono text-[#ededed] hover:bg-[#161616] hover:border-[#333333] cursor-pointer"
              >
                <span>Open in All Sessions View ({currentProfile.name})</span>
                <ArrowRight className="h-3.5 w-3.5 ml-1 text-[#a1a1a1]" />
              </Button>
            </CardContent>
          </Card>

          {/* Card 2: MCP Catalog Deep Link */}
          <Card className="border border-[#262626] bg-[#111111] rounded-geist hover:border-[#333333] transition-colors">
            <CardHeader className="p-5 pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                  <Server className="h-4 w-4 text-[#a1a1a1]" />
                  MCP Catalog & Tools
                </CardTitle>
                <Badge
                  variant="outline"
                  className="text-xs font-mono bg-[#0a0a0a] border-[#262626] text-[#ededed] px-2 py-0.5"
                >
                  {resourcesLoading ? '...' : `${profileMcpServers.length} active`}
                </Badge>
              </div>
              <CardDescription className="text-xs text-[#a1a1a1] pt-1">
                Explore connected MCP servers, global inheritance, and sandbox integrations.
              </CardDescription>
            </CardHeader>
            <CardContent className="p-5 pt-0">
              <Button
                variant="outline"
                size="sm"
                onClick={() => onNavigateToMcp && onNavigateToMcp(currentProfile.name)}
                className="w-full justify-between h-9 rounded-geist border-[#262626] bg-[#0a0a0a] text-xs font-mono text-[#ededed] hover:bg-[#161616] hover:border-[#333333] cursor-pointer"
              >
                <span>Open in MCP Catalog ({currentProfile.name})</span>
                <ArrowRight className="h-3.5 w-3.5 ml-1 text-[#a1a1a1]" />
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Compact Profile Environment Sandbox Card (No scattered layout or 800px void) */}
        <Card className="border border-[#262626] bg-[#111111] rounded-geist">
          <CardHeader className="pb-3 p-5">
            <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
              <Folder className="h-4 w-4 text-[#a1a1a1]" />
              Profile Environment & Storage Sandbox
            </CardTitle>
            <CardDescription className="text-xs text-[#a1a1a1]">
              Filesystem sandbox isolation and global configuration inheritance.
            </CardDescription>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="grid grid-cols-1 sm:grid-cols-[180px_1fr] gap-x-6 gap-y-3.5 text-xs font-mono py-1">
              <span className="text-[#a1a1a1]">Storage Sandbox:</span>
              <span className="text-[#ededed] font-mono break-all select-all">
                {currentProfile.path}
              </span>

              <span className="text-[#a1a1a1]">Connected Engines:</span>
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

              <span className="text-[#a1a1a1]">Global MCP Inheritance:</span>
              <div className="flex items-center gap-3">
                <Switch
                  checked={isMcpGlobalOn}
                  disabled={updatingConfig}
                  onCheckedChange={() => handleToggleMcpGlobal(currentProfile.name, isMcpGlobalOn)}
                />
                <span className="text-[#ededed]">
                  {isMcpGlobalOn ? 'Enabled (Inherit host MCP servers)' : 'Disabled (Strict sandbox isolation)'}
                </span>
              </div>

              <span className="text-[#a1a1a1]">Global Plugin Inheritance:</span>
              <div className="flex items-center gap-3">
                <Switch
                  checked={isPluginsGlobalOn}
                  disabled={updatingConfig}
                  onCheckedChange={() => handleTogglePluginsGlobal(currentProfile.name, isPluginsGlobalOn)}
                />
                <span className="text-[#ededed]">
                  {isPluginsGlobalOn ? 'Enabled (Inherit host plugins)' : 'Disabled (Strict sandbox isolation)'}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Visual Live Quota Breakdown (Replaces raw monospace text) */}
        <div className="space-y-3">
          <h3 className="text-xs font-mono font-medium uppercase tracking-wider text-[#a1a1a1] flex items-center gap-1.5">
            <Gauge className="h-3.5 w-3.5" />
            Live Quota Telemetry & Rate Limits
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
                      <span className="text-xs font-mono text-[#a1a1a1]">Engine</span>
                    </div>
                    {ad.has_credentials ? (
                      <Badge
                        variant="outline"
                        className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                      >
                        <span className="h-1.5 w-1.5 rounded-full bg-[#00e599]" />
                        <span>Authenticated</span>
                      </Badge>
                    ) : (
                      <Badge
                        variant="outline"
                        className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                      >
                        <span className="h-1.5 w-1.5 rounded-full bg-[#f5a623]" />
                        <span>Needs Auth</span>
                      </Badge>
                    )}
                  </div>
                  <CardDescription className="text-xs text-[#a1a1a1] pt-1">
                    {ad.account?.email || (ad.has_credentials ? 'Active CLI credentials detected' : 'No credentials detected')}
                  </CardDescription>
                </CardHeader>

                <CardContent className="space-y-4 p-5 pt-0">
                  {/* Visual Quota Meters */}
                  <QuotaBreakdown
                    summary={ad.quota?.summary}
                    bottleneckPct={ad.quota?.bottleneck_pct ?? 100}
                    isExhausted={ad.quota?.is_exhausted}
                  />

                  {/* Account detail chips */}
                  <div className="p-3 rounded-geist bg-[#0a0a0a] border border-[#262626] space-y-1.5 text-xs font-mono">
                    <div className="flex justify-between items-center text-[#a1a1a1]">
                      <span>Account:</span>
                      <span className="text-[#ededed] font-medium truncate max-w-[180px]">
                        {ad.account?.name || currentProfile.name}
                      </span>
                    </div>
                    {ad.account?.auth_method && (
                      <div className="flex justify-between items-center text-[#a1a1a1]">
                        <span>Auth Method:</span>
                        <span className="text-[#ededed] truncate max-w-[180px]">{ad.account.auth_method}</span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

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
        <div className="rounded-geist border border-[#f5a623]/30 bg-[#f5a623]/10 px-3.5 py-2 text-xs text-[#f5a623] font-mono flex items-center justify-between">
          <span>Notice: {error}</span>
          <button
            type="button"
            onClick={() => setError(null)}
            className="text-[#a1a1a1] hover:text-[#ededed] ml-2 text-xs cursor-pointer"
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
          <span className="font-mono text-[11px] px-2 py-0.5 rounded-full border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] tabular-nums">
            {profiles.length} active
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchProfiles}
            disabled={loading}
            className="rounded-geist h-8 border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
          >
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
              <Button size="sm" className="bg-[#ededed] text-[#000000] hover:bg-white rounded-geist h-8 font-medium">
                <Plus className="h-3.5 w-3.5 mr-1" />
                Add Profile
              </Button>
            </DialogTrigger>
            <DialogContent className="border border-[#262626] bg-[#111111] text-[#ededed] rounded-geist sm:max-w-md">
              <DialogHeader>
                <DialogTitle className="text-base font-semibold tracking-tight text-[#ededed]">
                  Create New Agent Profile
                </DialogTitle>
                <DialogDescription className="text-xs text-[#a1a1a1]">
                  Scaffold an isolated environment and configuration for your agent.
                </DialogDescription>
              </DialogHeader>

              <form onSubmit={handleCreate} className="space-y-4 py-2">
                <div className="space-y-1.5">
                  <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
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
                            : 'border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]'
                        }`}
                      >
                        {ag.toUpperCase()}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="space-y-1.5">
                  <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
                    Profile Name *
                  </label>
                  <Input
                    required
                    placeholder="e.g. work, sandbox, client-x"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
                    Associated Email (Optional)
                  </label>
                  <Input
                    type="email"
                    placeholder="dev@example.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
                  />
                </div>

                {/* Clone from existing profile rendered with toggle switch */}
                <div className="space-y-2 pt-2 border-t border-[#262626]">
                  <div className="flex items-center justify-between py-0.5">
                    <div className="space-y-0.5">
                      <label htmlFor="clone-toggle" className="text-xs font-medium text-[#ededed] cursor-pointer">
                        Clone from existing profile
                      </label>
                      <p className="text-[11px] text-[#a1a1a1]">
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
                      <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
                        Source Profile to Clone
                      </label>
                      <Input
                        placeholder="e.g. default"
                        value={formData.clone_from}
                        onChange={(e) => setFormData({ ...formData, clone_from: e.target.value })}
                        className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
                        autoFocus
                      />
                    </div>
                  )}
                </div>

                <DialogFooter className="pt-3 border-t border-[#262626] flex gap-2 justify-end">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setIsAddOpen(false);
                      setEnableClone(false);
                    }}
                    disabled={submitting}
                    className="rounded-geist text-xs h-8 border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    disabled={submitting}
                    className="bg-[#ededed] text-[#000000] hover:bg-white rounded-geist text-xs h-8 font-medium"
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
        <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0a0a0a]/50">
          <Bot className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-[#ededed]">No profiles found</h3>
          <p className="text-xs text-[#a1a1a1] mt-1 max-w-sm mx-auto">
            Click &ldquo;Add Profile&rdquo; to create an isolated environment for your agent.
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
                className="group relative overflow-hidden transition-all duration-150 border border-[#262626] bg-[#111111] hover:border-[#333333] rounded-geist flex flex-col justify-between cursor-pointer"
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
                          className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-[#00e599]" />
                          <span>Authenticated</span>
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] text-[11px] font-mono tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-[#f5a623]" />
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
                          className="h-7 w-7 text-[#666666] hover:text-[#ee0000] rounded-geist transition-colors"
                          onClick={(e) => {
                            e.stopPropagation();
                            openDeleteDialog(p.agent, p.name);
                          }}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                      <CardDescription className="text-xs text-[#a1a1a1] font-mono truncate pt-0.5 flex items-center gap-1">
                        <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                        <span className="truncate">{p.path}</span>
                      </CardDescription>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3 pb-4 p-5 pt-0">
                    {/* Account Info Box */}
                    <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-1.5 font-mono">
                      <div className="flex items-center justify-between text-[#a1a1a1]">
                        <span className="flex items-center gap-1.5">
                          <User className="h-3 w-3 text-[#666666]" /> Account
                        </span>
                        <span className="text-[#ededed] font-medium truncate max-w-[140px]">
                          {p.account?.name || p.name}
                        </span>
                      </div>
                      {p.account?.email && (
                        <div className="flex items-center justify-between text-[#a1a1a1]">
                          <span className="flex items-center gap-1.5">
                            <Mail className="h-3 w-3 text-[#666666]" /> Email
                          </span>
                          <span className="text-[#ededed] truncate max-w-[160px]">
                            {p.account.email}
                          </span>
                        </div>
                      )}
                      {p.account?.auth_method && (
                        <div className="flex items-center justify-between text-[#a1a1a1]">
                          <span className="flex items-center gap-1.5">
                            <Key className="h-3 w-3 text-[#666666]" /> Auth
                          </span>
                          <span className="text-[#ededed] truncate max-w-[140px]">
                            {p.account.auth_method}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Standardized Quota Telemetry (Strict h-1.5 across multi- and single-engine) */}
                    {p.adapters && p.adapters.length > 1 ? (
                      <div className="space-y-2 pt-1 border-t border-[#262626]">
                        <div className="flex items-center justify-between text-[11px] font-mono uppercase tracking-wider text-[#a1a1a1]">
                          <span className="flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Engine Quotas
                          </span>
                          <span>{p.adapters.length} Engines</span>
                        </div>
                        {p.adapters.map((ad) => {
                          const pct = ad.quota?.bottleneck_pct ?? 100;
                          return (
                            <div key={ad.agent} className="space-y-1">
                              <div className="flex items-center justify-between text-[11px] font-mono">
                                <span className="text-[#a1a1a1]">{ad.agent}</span>
                                <span className={`tabular-nums ${getTextColor(pct)}`}>
                                  {pct}%
                                </span>
                              </div>
                              <div className="w-full bg-[#262626] h-1.5 rounded-full overflow-hidden">
                                <div
                                  className={`h-full transition-all duration-300 ${getProgressColor(pct)}`}
                                  style={{ width: `${pct}%` }}
                                />
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    ) : (
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-[11px] font-mono uppercase tracking-wider text-[#a1a1a1] flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Quota Telemetry
                          </span>
                          <span
                            className={`font-mono text-xs font-medium tabular-nums ${getTextColor(
                              p.quota?.bottleneck_pct ?? 100
                            )}`}
                          >
                            {p.quota ? `${p.quota.bottleneck_pct}% remaining` : '100% remaining'}
                          </span>
                        </div>
                        <div className="w-full bg-[#262626] h-1.5 rounded-full overflow-hidden">
                          <div
                            className={`h-full transition-all duration-300 ${getProgressColor(
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
                <CardFooter className="pt-2 pb-4 px-5 border-t border-[#262626] flex items-center justify-between text-[11px] font-mono text-[#a1a1a1]">
                  <div className="flex items-center gap-1.5">
                    <span className="px-1.5 py-0.5 rounded border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] text-[10px] font-mono">
                      mcp: {isMcpOn ? 'on' : 'off'}
                    </span>
                    <span className="px-1.5 py-0.5 rounded border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] text-[10px] font-mono">
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
