import * as React from 'react';
import {
  ProfileDTO,
  SessionDTO,
  MCPServerDTO,
  getProfiles,
  deleteProfile,
  updateProfileConfig,
  getSessions,
  getMcpServers,
  resumeSession,
} from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
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
  ArrowRight,
  ChevronRight,
  Boxes,
  Copy,
  Check,
  Settings,
  Search,
  SlidersHorizontal,
  ShieldAlert,
} from 'lucide-react';
import { DeleteProfileModal } from '@/components/DeleteProfileModal';
import { AddProfileModal } from '@/components/AddProfileModal';
import { AgentBadge } from '@/components/AgentBadge';
import { QuotaBreakdown } from '@/components/QuotaBreakdown';
import { cn } from '@/lib/utils';

interface ProfilesViewProps {
  selectedProfile: string;
  onSelectProfile: (profile: string) => void;
  onNavigateToSessions?: (profile: string) => void;
  onNavigateToMcp?: (profile: string) => void;
  onProfilesLoaded?: (profiles: ProfileDTO[]) => void;
  onViewSession?: (session: SessionDTO) => void;
  activeTab?: string;
  onTabChange?: (tab: string) => void;
  profiles?: ProfileDTO[];
}

let profilesMemoryCache: ProfileDTO[] | null = null;

export function ProfilesView({
  selectedProfile,
  onSelectProfile,
  onNavigateToSessions,
  onNavigateToMcp,
  onProfilesLoaded,
  onViewSession,
  activeTab: controlledActiveTab,
  onTabChange,
  profiles: externalProfiles,
}: ProfilesViewProps) {
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>(() =>
    externalProfiles && externalProfiles.length > 0
      ? externalProfiles
      : profilesMemoryCache || []
  );
  const [loading, setLoading] = React.useState(
    () => !profilesMemoryCache && (!externalProfiles || externalProfiles.length === 0)
  );
  const [refreshing, setRefreshing] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (externalProfiles && externalProfiles.length > 0) {
      setProfiles(externalProfiles);
      profilesMemoryCache = externalProfiles;
      setLoading(false);
    }
  }, [externalProfiles]);

  // Dialog State
  const [isAddOpen, setIsAddOpen] = React.useState(false);

  // Profile-specific counts for direct launcher cards
  const [profileSessions, setProfileSessions] = React.useState<SessionDTO[]>([]);
  const [profileMcpServers, setProfileMcpServers] = React.useState<MCPServerDTO[]>([]);
  const [resourcesLoading, setResourcesLoading] = React.useState(false);

  // Updating profile config in-flight
  const [updatingConfig, setUpdatingConfig] = React.useState(false);

  const { toast } = useToast();

  const fetchProfiles = React.useCallback(async (forceRefresh = false) => {
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      const data = await getProfiles({
        refresh: forceRefresh,
      });
      setProfiles(data);
      profilesMemoryCache = data;
      onProfilesLoaded?.(data);
      if (forceRefresh) {
        toast({
          title: 'Quota Telemetry Refreshed',
          description: `Updated live quota telemetry for ${data.length} profile(s).`,
          variant: 'success',
        });
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch profiles';
      setError(msg);
      if (forceRefresh) {
        toast({
          title: 'Refresh Failed',
          description: msg,
          variant: 'destructive',
        });
      }
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
      setRefreshing(false);
    }
  }, [onProfilesLoaded, toast]);

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
    if (pct <= 0) return 'bg-status-danger-solid';
    if (pct <= 30) return 'bg-status-warning-solid';
    return 'bg-status-meter';
  };

  const getTextColor = (pct: number) => {
    if (pct <= 0) return 'text-status-danger';
    if (pct <= 30) return 'text-status-warning';
    return 'text-muted-foreground';
  };

  const currentProfile = React.useMemo(() => {
    if (!selectedProfile || selectedProfile === 'all') return null;
    return profiles.find((p) => p.name === selectedProfile) || null;
  }, [profiles, selectedProfile]);

  const [copiedPath, setCopiedPath] = React.useState(false);
  const [resumingId, setResumingId] = React.useState<string | null>(null);

  // View A multi-tab state
  const [internalActiveTab, setInternalActiveTab] = React.useState('quota');
  const activeTab = (controlledActiveTab === 'telemetry' ? 'quota' : controlledActiveTab) || internalActiveTab;
  const handleTabChange = (newTab: string) => {
    setInternalActiveTab(newTab);
    onTabChange?.(newTab);
  };
  const [selectedEngineTab, setSelectedEngineTab] = React.useState<string>('all');
  const [sessionSearch, setSessionSearch] = React.useState('');
  const [mcpSearch, setMcpSearch] = React.useState('');

  React.useEffect(() => {
    if (!controlledActiveTab) {
      setInternalActiveTab('quota');
    }
    setSelectedEngineTab('all');
    setSessionSearch('');
    setMcpSearch('');
  }, [selectedProfile, controlledActiveTab]);

  // Derived adapters for current profile
  const profileAdapters = React.useMemo(() => {
    if (!currentProfile) return [];
    if (currentProfile.adapters && currentProfile.adapters.length > 0) {
      return currentProfile.adapters;
    }
    return [
      {
        agent: currentProfile.agent,
        has_credentials: currentProfile.has_credentials,
        account: currentProfile.account,
        quota: currentProfile.quota,
      },
    ];
  }, [currentProfile]);

  const displayedAdapters = React.useMemo(() => {
    if (selectedEngineTab === 'all') return profileAdapters;
    return profileAdapters.filter((ad) => ad.agent === selectedEngineTab);
  }, [profileAdapters, selectedEngineTab]);

  const filteredSessions = React.useMemo(() => {
    if (!sessionSearch.trim()) return profileSessions;
    const q = sessionSearch.toLowerCase().trim();
    return profileSessions.filter(
      (s) =>
        s.title?.toLowerCase().includes(q) ||
        s.goal?.toLowerCase().includes(q) ||
        s.cwd?.toLowerCase().includes(q) ||
        s.id?.toLowerCase().includes(q) ||
        s.agent?.toLowerCase().includes(q)
    );
  }, [profileSessions, sessionSearch]);

  const filteredMcpServers = React.useMemo(() => {
    if (!mcpSearch.trim()) return profileMcpServers;
    const q = mcpSearch.toLowerCase().trim();
    return profileMcpServers.filter(
      (srv) =>
        srv.name.toLowerCase().includes(q) ||
        srv.command?.toLowerCase().includes(q) ||
        srv.scope?.toLowerCase().includes(q)
    );
  }, [profileMcpServers, mcpSearch]);

  const handleCopyPath = (path: string) => {
    navigator.clipboard.writeText(path);
    setCopiedPath(true);
    setTimeout(() => setCopiedPath(false), 2000);
  };

  const handleQuickResume = async (sess: SessionDTO) => {
    setResumingId(sess.id);
    try {
      const res = await resumeSession({
        agent: sess.agent,
        profile: sess.profile,
        session_id: sess.id,
      });
      toast({
        title: 'Terminal Launched',
        description: `Resumed in ${res.terminal || 'terminal'} window.`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Could not launch terminal';
      toast({
        title: 'Resume Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setResumingId(null);
    }
  };

  const formatRelativeTime = (isoString?: string) => {
    if (!isoString) return '';
    try {
      const d = new Date(isoString);
      const now = new Date();
      const diffSecs = Math.floor((now.getTime() - d.getTime()) / 1000);
      if (diffSecs < 60) return 'just now';
      const diffMins = Math.floor(diffSecs / 60);
      if (diffMins < 60) return `${diffMins}m ago`;
      const diffHours = Math.floor(diffMins / 60);
      if (diffHours < 24) return `${diffHours}h ago`;
      const diffDays = Math.floor(diffHours / 24);
      return `${diffDays}d ago`;
    } catch {
      return isoString;
    }
  };

  // Multi-dimensional Quota Health badge (prominently answers "what is the status?")
  const quotaHealthBadge = React.useMemo(() => {
    if (!currentProfile) return null;
    const q = currentProfile.quota;
    if (!q) {
      return (
        <Badge variant="outline" className="text-xs font-sans border-border bg-secondary text-muted-foreground">
          <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
          <span>Quota: Uncached</span>
        </Badge>
      );
    }
    if (q.is_exhausted || q.bottleneck_pct <= 0) {
      return (
        <Badge variant="outline" className="badge-danger text-xs font-sans">
          <span className="h-1.5 w-1.5 rounded-full bg-status-danger-solid" />
          <span>Quota Exhausted (0%)</span>
        </Badge>
      );
    }
    if (q.bottleneck_pct <= 30) {
      return (
        <Badge variant="outline" className="badge-warning text-xs font-sans">
          <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
          <span>Rate Limit Warning ({q.bottleneck_pct}%)</span>
        </Badge>
      );
    }
    return (
      <Badge variant="outline" className="badge-success text-xs font-sans">
        <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
        <span>Quota Healthy ({q.bottleneck_pct}%)</span>
      </Badge>
    );
  }, [currentProfile]);

  const handleRefreshCurrentProfile = React.useCallback(async () => {
    if (!currentProfile) return;
    setRefreshing(true);
    setError(null);
    try {
      const data = await getProfiles({
        profile: currentProfile.name,
        refresh: true,
      });
      setProfiles(data);
      onProfilesLoaded?.(data);
      toast({
        title: 'Quota Telemetry Refreshed',
        description: `Updated live telemetry for "${currentProfile.name}".`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to refresh telemetry';
      toast({
        title: 'Refresh Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setRefreshing(false);
    }
  }, [currentProfile, onProfilesLoaded, toast]);

  // ==========================================
  // VIEW A: SPECIFIC PROFILE WORKSPACE VIEW
  // ==========================================
  if (selectedProfile !== 'all') {
    if (loading && !currentProfile) {
      return (
        <div className="py-24 text-center">
          <div className="inline-flex items-center gap-2 text-xs font-mono text-[#888888]">
            <span className="h-3 w-3 rounded-full border-2 border-[#0070f3] border-t-transparent animate-spin" />
            <span>Loading profile "{selectedProfile}"...</span>
          </div>
        </div>
      );
    }

    if (!currentProfile) {
      return (
        <div className="py-16 text-center space-y-4">
          <div className="text-sm font-semibold text-[#ededed]">Profile "{selectedProfile}" Not Found</div>
          <p className="text-xs text-[#888888]">This profile may have been deleted or the name is incorrect.</p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onSelectProfile('all')}
            className="rounded-geist text-xs text-[#ededed] border-[#262626] bg-[#0e0e0e] hover:bg-[#1a1a1a]"
          >
            Return to Profiles Overview
          </Button>
        </div>
      );
    }

    const isMcpGlobalOn = currentProfile.mcp_global !== false;
    const isPluginsGlobalOn = currentProfile.plugins_global !== false;

    return (
      <div className="space-y-6">
        {/* Profile Page Header (Clean, uncarded page header with high/low contrast hierarchy) */}
        <div className="space-y-4 pb-6 border-b border-[#222222]">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-2.5 flex-wrap">
              <h1 className="text-2xl font-semibold tracking-tight text-[#ededed] font-sans">
                {currentProfile.name}
              </h1>
              {profileAdapters.length > 0 ? (
                profileAdapters.map((ad) => (
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
                  className="badge-success text-xs font-sans"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                  <span>Authenticated</span>
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="badge-warning text-xs font-sans"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                  <span>Needs Auth</span>
                </Badge>
              )}
              {quotaHealthBadge}
              <span className="text-xs font-mono text-[#888888] px-2 py-0.5 rounded-full border border-[#222222] bg-[#0c0c0c]">
                {resourcesLoading ? '...' : `${profileSessions.length} sessions`}
              </span>
            </div>

            <div className="flex items-center gap-2 self-end sm:self-auto shrink-0">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefreshCurrentProfile}
                disabled={loading || refreshing}
                className="h-8 rounded-geist text-xs font-sans border-[#262626] bg-[#0e0e0e] text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a] cursor-pointer px-3"
              >
                <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${refreshing || loading ? 'animate-spin' : ''}`} />
                {refreshing ? 'Refreshing...' : 'Refresh Quota'}
              </Button>
              <Button
                variant="outline"
                size="icon"
                onClick={() => openDeleteDialog(currentProfile.agent, currentProfile.name)}
                title="Delete profile"
                aria-label="Delete profile"
                className="h-8 w-8 rounded-geist border-[#262626] bg-[#0e0e0e] text-[#666666] hover:text-status-danger hover:bg-status-danger-bg hover:border-status-danger-border cursor-pointer"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Sandbox & Environment Sub-bar with low-contrast labels & high-contrast values */}
          <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-y-3 gap-x-6 text-xs font-sans">
            <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
              <div className="flex items-center gap-1.5">
                <Folder className="h-3.5 w-3.5 text-[#555555]" />
                <span className="text-[#666666]">Sandbox:</span>
                <span className="text-[#ededed] font-mono select-all">{currentProfile.path}</span>
                <button
                  type="button"
                  onClick={() => handleCopyPath(currentProfile.path)}
                  title="Copy sandbox path"
                  className="p-1 rounded hover:bg-[#1a1a1a] text-[#666666] hover:text-[#ededed] cursor-pointer transition-colors"
                >
                  {copiedPath ? (
                    <Check className="h-3 w-3 text-status-success" />
                  ) : (
                    <Copy className="h-3 w-3" />
                  )}
                </button>
              </div>
              {currentProfile.account?.email && (
                <div className="flex items-center gap-1.5">
                  <Mail className="h-3.5 w-3.5 text-[#555555]" />
                  <span className="text-[#666666]">Email:</span>
                  <span className="text-[#ededed]">{currentProfile.account.email}</span>
                </div>
              )}
              {currentProfile.account?.auth_method && (
                <div className="flex items-center gap-1.5">
                  <Key className="h-3.5 w-3.5 text-[#555555]" />
                  <span className="text-[#666666]">Auth:</span>
                  <span className="text-[#888888]">{currentProfile.account.auth_method}</span>
                </div>
              )}
            </div>

            {/* Global Sandbox Inheritance Toggles */}
            <div className="flex items-center gap-5 shrink-0 pt-1 lg:pt-0">
              <div className="flex items-center gap-2">
                <span className="text-[#666666]">Global MCP:</span>
                <Switch
                  size="sm"
                  checked={isMcpGlobalOn}
                  disabled={updatingConfig}
                  onCheckedChange={() => handleToggleMcpGlobal(currentProfile.name, isMcpGlobalOn)}
                />
              </div>

              <div className="flex items-center gap-2">
                <span className="text-[#666666]">Global Plugins:</span>
                <Switch
                  size="sm"
                  checked={isPluginsGlobalOn}
                  disabled={updatingConfig}
                  onCheckedChange={() => handleTogglePluginsGlobal(currentProfile.name, isPluginsGlobalOn)}
                />
              </div>
            </div>
          </div>
        </div>

        {/* Multi-Tab Workspace Organization */}
        <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full space-y-6">
          <div className="flex items-center justify-between">
            <TabsList className="bg-[#0c0c0c] border border-[#222222] p-1 h-9 rounded-geist gap-1">
              <TabsTrigger value="quota" className="gap-2 text-xs">
                <Gauge className="h-3.5 w-3.5 text-[#666666]" />
                <span>Quota</span>
                {profileAdapters.length > 1 && (
                  <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                    {profileAdapters.length}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger value="sessions" className="gap-2 text-xs">
                <Terminal className="h-3.5 w-3.5 text-[#666666]" />
                <span>Sessions</span>
                <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                  {profileSessions.length}
                </span>
              </TabsTrigger>
              <TabsTrigger value="mcp" className="gap-2 text-xs">
                <Server className="h-3.5 w-3.5 text-[#666666]" />
                <span>MCP Tools</span>
                <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                  {profileMcpServers.length}
                </span>
              </TabsTrigger>
              <TabsTrigger value="settings" className="gap-2 text-xs">
                <Settings className="h-3.5 w-3.5 text-[#666666]" />
                <span>Settings</span>
              </TabsTrigger>
            </TabsList>
          </div>

          {/* TAB 1: QUOTA */}
          <TabsContent value="quota" className="space-y-5 focus-visible:outline-none">
            {/* Multi-Engine Selector (when a profile runs >1 adapter, e.g. office with codex & claude) */}
            {profileAdapters.length > 1 && (
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-geist bg-[#0a0a0a] border border-[#222222]">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[#666666] font-medium">Engine Switcher:</span>
                  <div className="inline-flex items-center rounded-geist border border-[#262626] bg-[#000000] p-0.5 text-xs">
                    <button
                      type="button"
                      onClick={() => setSelectedEngineTab('all')}
                      className={cn(
                        'px-3 py-1 rounded-[4px] font-medium transition-colors cursor-pointer text-xs flex items-center gap-1.5',
                        selectedEngineTab === 'all'
                          ? 'bg-[#1f1f1f] text-[#ededed] border border-[#333333]'
                          : 'text-[#888888] hover:text-[#ededed]'
                      )}
                    >
                      <span>All Engines</span>
                      <span className="text-[10px] font-mono px-1 rounded bg-[#262626] text-[#888888]">
                        {profileAdapters.length}
                      </span>
                    </button>
                    {profileAdapters.map((ad) => (
                      <button
                        key={ad.agent}
                        type="button"
                        onClick={() => setSelectedEngineTab(ad.agent)}
                        className={cn(
                          'px-3 py-1 rounded-[4px] font-medium transition-colors cursor-pointer text-xs capitalize flex items-center gap-1.5',
                          selectedEngineTab === ad.agent
                            ? 'bg-[#1f1f1f] text-[#ededed] border border-[#333333]'
                            : 'text-[#888888] hover:text-[#ededed]'
                        )}
                      >
                        <span>{ad.agent}</span>
                        {ad.quota && (
                          <span
                            className={cn(
                              'text-[10px] tabular-nums font-mono',
                              ad.quota.bottleneck_pct <= 0
                                ? 'text-status-danger'
                                : ad.quota.bottleneck_pct <= 30
                                ? 'text-status-warning'
                                : 'text-[#888888]'
                            )}
                          >
                            {ad.quota.bottleneck_pct}%
                          </span>
                        )}
                      </button>
                    ))}
                  </div>
                </div>
                <span className="text-[11px] text-[#666666]">
                  {selectedEngineTab === 'all'
                    ? `Viewing combined quota for ${profileAdapters.length} configured engines side-by-side.`
                    : `Focused quota view for ${selectedEngineTab}.`}
                </span>
              </div>
            )}

            {/* Side-by-side adapter cards grid (2-column layout prevents vertical stacking of multi-adapters) */}
            <div
              className={cn(
                'grid gap-5 items-start',
                displayedAdapters.length > 1
                  ? 'grid-cols-1 lg:grid-cols-2'
                  : 'grid-cols-1 max-w-3xl'
              )}
            >
              {displayedAdapters.map((ad) => (
                <Card key={ad.agent} className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
                  <CardHeader className="pb-3 p-5">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Gauge className="h-4 w-4 text-[#666666]" />
                        <CardTitle className="text-sm font-semibold text-[#ededed]">
                          Quota
                        </CardTitle>
                        {getAgentBadge(ad.agent)}
                      </div>
                      {ad.has_credentials ? (
                        <Badge variant="outline" className="badge-success text-xs font-sans">
                          <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                          <span>Authenticated</span>
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="badge-warning text-xs font-sans">
                          <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                          <span>Needs Auth</span>
                        </Badge>
                      )}
                    </div>
                    <CardDescription className="text-xs text-[#888888] pt-1">
                      {ad.account?.email ||
                        (ad.has_credentials ? 'Active CLI credentials detected' : 'No credentials detected')}
                    </CardDescription>
                  </CardHeader>

                  <CardContent className="space-y-4 p-5 pt-0">
                    {/* Visual Quota Meters */}
                    <QuotaBreakdown
                      summary={ad.quota?.summary}
                      bottleneckPct={ad.quota?.bottleneck_pct ?? 100}
                      isExhausted={ad.quota?.is_exhausted}
                    />

                    {/* Account detail chips with low-contrast labels and crisp white values */}
                    <div className="p-3 rounded-geist bg-[#0f0f0f] border border-[#222222] space-y-1.5 text-xs font-sans">
                      <div className="flex justify-between items-center">
                        <span className="text-[#666666]">Account:</span>
                        <span className="text-[#ededed] font-medium truncate max-w-[220px]">
                          {ad.account?.name || currentProfile.name}
                        </span>
                      </div>
                      {ad.account?.email && (
                        <div className="flex justify-between items-center">
                          <span className="text-[#666666]">Email:</span>
                          <span className="text-[#ededed] truncate max-w-[220px]">
                            {ad.account.email}
                          </span>
                        </div>
                      )}
                      {ad.account?.auth_method && (
                        <div className="flex justify-between items-center">
                          <span className="text-[#666666]">Auth Method:</span>
                          <span className="text-[#ededed] truncate max-w-[220px]">
                            {ad.account.auth_method}
                          </span>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </TabsContent>

          {/* TAB 2: SESSIONS */}
          <TabsContent value="sessions" className="space-y-4 focus-visible:outline-none">
            {/* Filter toolbar */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-3 rounded-geist bg-[#0a0a0a] border border-[#222222]">
              <div className="relative flex-1 max-w-md w-full">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-[#666666]" />
                <Input
                  placeholder="Filter sessions by title, directory, or ID..."
                  value={sessionSearch}
                  onChange={(e) => setSessionSearch(e.target.value)}
                  className="pl-8 h-8 text-xs font-sans bg-[#000000] border-[#262626] text-[#ededed] placeholder:text-[#555555] rounded-geist"
                />
              </div>

              <div className="flex items-center gap-3">
                <span className="text-xs text-[#666666] font-sans">
                  {filteredSessions.length} session{filteredSessions.length === 1 ? '' : 's'}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onNavigateToSessions && onNavigateToSessions(currentProfile.name)}
                  className="h-8 text-xs font-sans text-[#888888] hover:text-[#ededed] border-[#262626] bg-[#0e0e0e] hover:bg-[#1a1a1a] cursor-pointer px-3"
                >
                  <span>Open Full Explorer</span>
                  <ArrowRight className="h-3 w-3 ml-1 text-[#666666]" />
                </Button>
              </div>
            </div>

            {/* Sessions List */}
            {resourcesLoading ? (
              <div className="py-12 text-center text-xs text-[#666666] font-sans">
                Loading sessions...
              </div>
            ) : filteredSessions.length === 0 ? (
              <div className="py-12 text-center text-xs text-[#666666] font-sans border border-dashed border-[#222222] rounded-geist bg-[#0a0a0a]">
                {sessionSearch ? 'No sessions match your search query.' : 'No recorded sessions for this profile yet.'}
              </div>
            ) : (
              <div className="divide-y divide-[#222222] border border-[#222222] rounded-geist overflow-hidden bg-[#0a0a0a]">
                {filteredSessions.map((sess) => (
                  <div
                    key={sess.id}
                    className="p-3.5 hover:bg-[#111111] transition-colors flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
                  >
                    <div className="min-w-0 flex-1 space-y-1.5">
                      <div className="flex items-center gap-2 flex-wrap">
                        <button
                          type="button"
                          onClick={() => onViewSession?.(sess)}
                          className="font-medium text-[#ededed] hover:underline truncate cursor-pointer text-left font-sans text-sm"
                        >
                          {sess.title || sess.goal || sess.id.slice(0, 8)}
                        </button>
                        {getAgentBadge(sess.agent)}
                        <span className="font-mono text-[10px] text-[#888888] px-1.5 py-0.5 rounded border border-[#222222] bg-[#050505] shrink-0">
                          {sess.turns} turns
                        </span>
                        {sess.is_active && (
                          <span className="inline-flex items-center gap-1 text-[10px] text-status-success font-medium">
                            <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                            active
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-[11px] font-mono truncate">
                        <span className="text-[#666666] shrink-0">CWD:</span>
                        <span className="text-[#888888] truncate select-all">{sess.cwd}</span>
                        <span className="text-[#444444]">•</span>
                        <span className="text-[#666666] shrink-0">Updated:</span>
                        <span className="text-[#888888] shrink-0">{formatRelativeTime(sess.updated_at)}</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 shrink-0">
                      {onViewSession && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onViewSession(sess)}
                          className="h-7 text-xs font-sans text-[#888888] hover:text-[#ededed] hover:bg-[#1f1f1f] cursor-pointer px-2.5"
                        >
                          View Transcript
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleQuickResume(sess)}
                        disabled={resumingId === sess.id}
                        className="h-7 text-xs font-sans rounded-geist border-[#262626] bg-[#0e0e0e] hover:bg-[#1a1a1a] text-[#ededed] cursor-pointer px-2.5 shrink-0"
                      >
                        <Terminal className={`h-3 w-3 mr-1 ${resumingId === sess.id ? 'animate-spin' : ''}`} />
                        {resumingId === sess.id ? 'Launching...' : 'Resume'}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </TabsContent>

          {/* TAB 3: MCP TOOLS */}
          <TabsContent value="mcp" className="space-y-4 focus-visible:outline-none">
            {/* Filter toolbar */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-3 rounded-geist bg-[#0a0a0a] border border-[#222222]">
              <div className="relative flex-1 max-w-md w-full">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-[#666666]" />
                <Input
                  placeholder="Filter MCP servers..."
                  value={mcpSearch}
                  onChange={(e) => setMcpSearch(e.target.value)}
                  className="pl-8 h-8 text-xs font-sans bg-[#000000] border-[#262626] text-[#ededed] placeholder:text-[#555555] rounded-geist"
                />
              </div>

              <div className="flex items-center gap-3">
                <span className="text-xs text-[#666666] font-sans">
                  {filteredMcpServers.length} server{filteredMcpServers.length === 1 ? '' : 's'}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onNavigateToMcp && onNavigateToMcp(currentProfile.name)}
                  className="h-8 text-xs font-sans text-[#888888] hover:text-[#ededed] border-[#262626] bg-[#0e0e0e] hover:bg-[#1a1a1a] cursor-pointer px-3"
                >
                  <span>Open Full MCP Catalog</span>
                  <ArrowRight className="h-3 w-3 ml-1 text-[#666666]" />
                </Button>
              </div>
            </div>

            {/* Servers Grid */}
            {resourcesLoading ? (
              <div className="py-12 text-center text-xs text-[#666666] font-sans">
                Loading MCP servers...
              </div>
            ) : filteredMcpServers.length === 0 ? (
              <div className="py-12 text-center text-xs text-[#666666] font-sans border border-dashed border-[#222222] rounded-geist bg-[#0a0a0a]">
                {mcpSearch ? 'No MCP servers match your search.' : 'No MCP servers configured for this profile.'}
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredMcpServers.map((srv) => (
                  <Card key={srv.name} className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
                    <CardHeader className="p-4 pb-2">
                      <div className="flex items-center justify-between gap-2">
                        <CardTitle className="font-mono text-sm font-semibold text-[#ededed] truncate">
                          {srv.name}
                        </CardTitle>
                        <Badge
                          variant="outline"
                          className={cn(
                            'text-[10px] font-mono px-2 py-0.5 rounded-full',
                            srv.scope === 'global'
                              ? 'border-[#333333] bg-[#1a1a1a] text-[#888888]'
                              : 'border-[#262626] bg-[#0e0e0e] text-[#ededed]'
                          )}
                        >
                          {srv.scope || 'profile'}
                        </Badge>
                      </div>
                    </CardHeader>
                    <CardContent className="p-4 pt-1 space-y-2">
                      <div className="space-y-1">
                        <span className="text-[11px] text-[#666666]">Command:</span>
                        <div className="bg-[#050505] border border-[#222222] p-2 rounded text-xs font-mono text-[#888888] select-all truncate">
                          {srv.command} {srv.args?.join(' ')}
                        </div>
                      </div>
                      {srv.args && srv.args.length > 0 && (
                        <div className="flex items-center gap-1.5 text-[11px]">
                          <span className="text-[#666666]">Args:</span>
                          <span className="text-[#888888] font-mono">{srv.args.length} argument(s)</span>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </TabsContent>

          {/* TAB 4: SETTINGS */}
          <TabsContent value="settings" className="space-y-5 max-w-3xl focus-visible:outline-none">
            {/* Configured Engines Card */}
            <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
              <CardHeader className="p-5 pb-3">
                <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                  <SlidersHorizontal className="h-4 w-4 text-[#666666]" />
                  Configured Agent Engines
                </CardTitle>
                <CardDescription className="text-xs text-[#888888] pt-1">
                  Engines sharing this sandbox workspace profile.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0 space-y-3">
                {profileAdapters.map((ad) => (
                  <div
                    key={ad.agent}
                    className="p-3 rounded-geist bg-[#0f0f0f] border border-[#222222] flex items-center justify-between gap-3 text-xs"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        {getAgentBadge(ad.agent)}
                        <span className="font-semibold text-[#ededed] capitalize">{ad.agent}</span>
                      </div>
                      <div className="text-[#666666] text-[11px]">
                        {ad.account?.email || (ad.has_credentials ? 'Active CLI credentials' : 'No credentials')}
                      </div>
                    </div>
                    {ad.has_credentials ? (
                      <Badge variant="outline" className="badge-success text-[11px]">
                        <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                        <span>Auth OK</span>
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="badge-warning text-[11px]">
                        <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                        <span>Needs Auth</span>
                      </Badge>
                    )}
                  </div>
                ))}
              </CardContent>
            </Card>

            {/* Danger Zone */}
            <Card className="border border-status-danger-border bg-status-danger-bg rounded-geist">
              <CardHeader className="p-5 pb-3">
                <CardTitle className="text-sm font-semibold text-status-danger flex items-center gap-2">
                  <ShieldAlert className="h-4 w-4 text-status-danger" />
                  Danger Zone
                </CardTitle>
                <CardDescription className="text-xs text-[#888888] pt-1">
                  Permanent actions that cannot be undone.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                  <div className="text-xs text-[#888888]">
                    Deleting this profile permanently removes its sandbox directory and configuration.
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => openDeleteDialog(currentProfile.agent, currentProfile.name)}
                    className="h-8 rounded-geist text-xs font-sans border-status-danger-border bg-status-danger-bg text-status-danger hover:bg-status-danger-solid hover:text-white cursor-pointer px-3 shrink-0"
                  >
                    <Trash2 className="h-3.5 w-3.5 mr-1.5" />
                    Delete Profile
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

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
        <div className="rounded-geist border border-status-warning-border bg-status-warning-bg px-3.5 py-2 text-xs text-status-warning font-sans flex items-center justify-between">
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
          <span className="font-sans text-[11px] font-medium px-2 py-0.5 rounded-full border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] tabular-nums">
            {profiles.length} active
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchProfiles(true)}
            disabled={loading || refreshing}
            className="rounded-geist h-8 border-border bg-secondary text-foreground hover:bg-accent cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${refreshing || loading ? 'animate-spin' : ''}`} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </Button>

          <Button
            size="sm"
            onClick={() => setIsAddOpen(true)}
            className="bg-[#ededed] text-[#000000] hover:bg-white rounded-geist h-8 font-medium cursor-pointer"
          >
            <Plus className="h-3.5 w-3.5 mr-1" />
            Add Profile
          </Button>

          <AddProfileModal
            open={isAddOpen}
            onOpenChange={setIsAddOpen}
            onProfileCreated={fetchProfiles}
          />
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
                tabIndex={0}
                role="button"
                aria-label={`View profile ${p.name}`}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    onSelectProfile(p.name);
                  }
                }}
                className="group relative overflow-hidden transition-all duration-150 border border-border bg-card hover:border-border/80 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring rounded-geist flex flex-col justify-between cursor-pointer"
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
                          className="bg-muted text-muted-foreground border-border text-[11px] font-sans font-medium tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                          <span>Authenticated</span>
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="bg-muted text-muted-foreground border-border text-[11px] font-sans font-medium tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                          <span>Needs Auth</span>
                        </Badge>
                      )}
                    </div>

                    <div className="pt-2">
                      <CardTitle className="text-base font-semibold text-foreground group-hover:text-foreground transition-colors">
                        {p.name}
                      </CardTitle>
                      <CardDescription className="text-xs text-muted-foreground truncate pt-0.5 flex items-center gap-1">
                        <Folder className="h-3 w-3 shrink-0 text-muted-foreground/60" />
                        <span className="truncate font-mono">{p.path}</span>
                      </CardDescription>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3 pb-4 p-5 pt-0">
                    {/* Account Info Box */}
                    <div className="rounded-geist border border-[#222222] bg-[#0c0c0c] p-3 text-xs space-y-1.5 font-sans">
                      <div className="flex items-center justify-between">
                        <span className="flex items-center gap-1.5 text-[#666666]">
                          <User className="h-3 w-3 text-[#555555]" /> Account
                        </span>
                        <span className="text-[#ededed] font-medium truncate max-w-[140px]">
                          {p.account?.name || p.name}
                        </span>
                      </div>
                      {p.account?.email && (
                        <div className="flex items-center justify-between">
                          <span className="flex items-center gap-1.5 text-[#666666]">
                            <Mail className="h-3 w-3 text-[#555555]" /> Email
                          </span>
                          <span className="text-[#ededed] truncate max-w-[160px]">
                            {p.account.email}
                          </span>
                        </div>
                      )}
                      {p.account?.auth_method && (
                        <div className="flex items-center justify-between">
                          <span className="flex items-center gap-1.5 text-[#666666]">
                            <Key className="h-3 w-3 text-[#555555]" /> Auth
                          </span>
                          <span className="text-[#888888] truncate max-w-[140px]">
                            {p.account.auth_method}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Standardized Quota Telemetry (Strict h-1.5 across multi- and single-engine) */}
                    {p.adapters && p.adapters.length > 1 ? (
                      <div className="space-y-2 pt-1 border-t border-[#222222]">
                        <div className="flex items-center justify-between text-[11px] font-sans font-medium uppercase tracking-wider text-[#666666]">
                          <span className="flex items-center gap-1">
                            <Gauge className="h-3 w-3 text-[#555555]" /> Engine Quotas
                          </span>
                          <span className="text-[#555555]">{p.adapters.length} Engines</span>
                        </div>
                        {p.adapters.map((ad) => {
                          const pct = ad.quota?.bottleneck_pct ?? 100;
                          return (
                            <div key={ad.agent} className="space-y-1">
                              <div className="flex items-center justify-between text-[11px] font-sans">
                                <span className="text-[#666666] capitalize">{ad.agent}</span>
                                <span className={`tabular-nums ${getTextColor(pct)}`}>
                                  {pct}%
                                </span>
                              </div>
                              <div
                                role="progressbar"
                                aria-valuenow={pct}
                                aria-valuemin={0}
                                aria-valuemax={100}
                                aria-label={`${ad.agent} quota`}
                                className="w-full bg-[#181818] h-1.5 rounded-full overflow-hidden border border-[#262626]/50"
                              >
                                <div
                                  className={`h-full transition-all duration-300 ${getProgressColor(pct)}`}
                                  style={{
                                    width: `${Math.max(
                                      ad.quota?.is_exhausted || pct <= 0 ? 6 : 0,
                                      Math.min(100, pct)
                                    )}%`,
                                  }}
                                />
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    ) : (
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-[11px] font-sans font-medium uppercase tracking-wider text-[#666666] flex items-center gap-1">
                            <Gauge className="h-3 w-3 text-[#555555]" /> Quota
                          </span>
                        {(() => {
                          const pct = p.quota?.bottleneck_pct ?? 100;
                          return (
                            <span
                              className={`font-sans text-xs font-medium tabular-nums ${
                                pct <= 0
                                  ? 'text-status-danger'
                                  : pct <= 30
                                  ? 'text-status-warning'
                                  : 'text-[#888888]'
                              }`}
                            >
                              {p.quota ? `${pct}% remaining` : '100% remaining'}
                            </span>
                          );
                        })()}
                        </div>
                        <div
                          role="progressbar"
                          aria-valuenow={p.quota?.bottleneck_pct ?? 100}
                          aria-valuemin={0}
                          aria-valuemax={100}
                          aria-label="Profile quota"
                          className="w-full bg-[#181818] h-1.5 rounded-full overflow-hidden border border-[#262626]/50"
                        >
                          <div
                            className={`h-full transition-all duration-300 ${getProgressColor(
                              p.quota?.bottleneck_pct ?? 100
                            )}`}
                            style={{
                              width: `${Math.max(
                                p.quota?.is_exhausted || (p.quota?.bottleneck_pct ?? 100) <= 0 ? 6 : 0,
                                Math.min(100, p.quota?.bottleneck_pct ?? 100)
                              )}%`,
                            }}
                          />
                        </div>
                        {p.quota?.summary && (
                          <p className="text-[11px] text-[#666666] font-sans truncate" title={p.quota.summary}>
                            {p.quota.summary}
                          </p>
                        )}
                      </div>
                    )}
                  </CardContent>
                </div>

                {/* Card Footer: MCP & Plugin configuration chips + Inspect Action */}
                <CardFooter className="pt-2 pb-4 px-5 border-t border-border flex items-center justify-between text-[11px] font-sans text-muted-foreground">
                  <div className="flex items-center gap-1.5">
                    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded border border-border bg-muted text-[10px] font-sans font-medium text-muted-foreground">
                      <Server className="h-3 w-3 text-muted-foreground/60" />
                      <span>MCP</span>
                      <span className={isMcpOn ? 'text-foreground font-medium' : 'text-muted-foreground/60'}>
                        {isMcpOn ? 'on' : 'off'}
                      </span>
                    </span>
                    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded border border-border bg-muted text-[10px] font-sans font-medium text-muted-foreground">
                      <Boxes className="h-3 w-3 text-muted-foreground/60" />
                      <span>Plugins</span>
                      <span className={isPluginsOn ? 'text-foreground font-medium' : 'text-muted-foreground/60'}>
                        {isPluginsOn ? 'on' : 'off'}
                      </span>
                    </span>
                  </div>

                  <span className="flex items-center text-foreground group-hover:text-foreground transition-colors">
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
