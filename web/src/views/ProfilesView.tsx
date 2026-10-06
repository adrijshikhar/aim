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
  Boxes,
  Copy,
  Check,
} from 'lucide-react';
import { DeleteProfileModal } from '@/components/DeleteProfileModal';
import { AddProfileModal } from '@/components/AddProfileModal';
import { AgentBadge } from '@/components/AgentBadge';
import { QuotaBreakdown } from '@/components/QuotaBreakdown';

interface ProfilesViewProps {
  selectedProfile: string;
  onSelectProfile: (profile: string) => void;
  onNavigateToSessions?: (profile: string) => void;
  onNavigateToMcp?: (profile: string) => void;
  onProfilesLoaded?: (profiles: ProfileDTO[]) => void;
  onViewSession?: (session: SessionDTO) => void;
}

export function ProfilesView({
  selectedProfile,
  onSelectProfile,
  onNavigateToSessions,
  onNavigateToMcp,
  onProfilesLoaded,
  onViewSession,
}: ProfilesViewProps) {
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [refreshing, setRefreshing] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

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
  if (selectedProfile !== 'all' && currentProfile) {
    const isMcpGlobalOn = currentProfile.mcp_global !== false;
    const isPluginsGlobalOn = currentProfile.plugins_global !== false;

    return (
      <div className="space-y-6">
        {/* Profile Header Card */}
        <div className="rounded-geist border border-border bg-card p-5 space-y-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-2.5 flex-wrap">
              <h1 className="text-xl font-semibold tracking-tight text-foreground font-sans">
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
              <span className="text-xs font-mono text-muted-foreground px-2 py-0.5 rounded-full border border-border bg-secondary">
                {resourcesLoading ? '...' : `${profileSessions.length} sessions`}
              </span>
            </div>

            <div className="flex items-center gap-2 self-end sm:self-auto shrink-0">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefreshCurrentProfile}
                disabled={loading || refreshing}
                className="h-8 rounded-geist text-xs font-sans border-border bg-secondary text-muted-foreground hover:text-foreground hover:bg-accent cursor-pointer px-3"
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
                className="h-8 w-8 rounded-geist border-border bg-secondary text-muted-foreground hover:text-status-danger hover:bg-status-danger-bg hover:border-status-danger-border cursor-pointer"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          {/* Path and identity sub-bar */}
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 pt-3 border-t border-border text-xs font-sans text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <Folder className="h-3.5 w-3.5 text-muted-foreground/70" />
              <span className="text-foreground font-mono select-all">{currentProfile.path}</span>
              <button
                type="button"
                onClick={() => handleCopyPath(currentProfile.path)}
                title="Copy sandbox path"
                className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
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
                <Mail className="h-3.5 w-3.5 text-muted-foreground/70" />
                <span>{currentProfile.account.email}</span>
              </div>
            )}
            {currentProfile.account?.auth_method && (
              <div className="flex items-center gap-1.5">
                <Key className="h-3.5 w-3.5 text-muted-foreground/70" />
                <span>{currentProfile.account.auth_method}</span>
              </div>
            )}
          </div>
        </div>

        {/* 2-Column Bento Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Left Column (7 cols, ~60%): Quota Telemetry & Recent Sessions */}
          <div className="lg:col-span-7 space-y-6">
            {/* Visual Live Quota Breakdown */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-sans font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                  <Gauge className="h-3.5 w-3.5" />
                  Live Quota Telemetry & Rate Limits
                </h3>
                <span className="text-[11px] font-mono text-muted-foreground">
                  {refreshing ? 'Updating live...' : 'Active live telemetry'}
                </span>
              </div>

              <div className="grid grid-cols-1 gap-4">
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
                  <Card key={ad.agent} className="border border-border bg-card rounded-geist">
                    <CardHeader className="pb-3 p-5">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          {getAgentBadge(ad.agent)}
                          <span className="text-xs font-sans text-muted-foreground">Engine</span>
                        </div>
                        {ad.has_credentials ? (
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
                      </div>
                      <CardDescription className="text-xs text-muted-foreground pt-1">
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
                      <div className="p-3 rounded-geist bg-secondary/50 border border-border space-y-1.5 text-xs font-sans">
                        <div className="flex justify-between items-center text-muted-foreground">
                          <span>Account:</span>
                          <span className="text-foreground font-medium truncate max-w-[220px]">
                            {ad.account?.name || currentProfile.name}
                          </span>
                        </div>
                        {ad.account?.auth_method && (
                          <div className="flex justify-between items-center text-muted-foreground">
                            <span>Auth Method:</span>
                            <span className="text-foreground truncate max-w-[220px]">{ad.account.auth_method}</span>
                          </div>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </div>

            {/* Recent Interactive Sessions */}
            <Card className="border border-border bg-card rounded-geist">
              <CardHeader className="p-5 pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold text-foreground flex items-center gap-2">
                    <Terminal className="h-4 w-4 text-muted-foreground" />
                    Recent Sessions
                  </CardTitle>
                  <Badge
                    variant="outline"
                    className="text-xs font-sans font-medium bg-secondary border-border text-foreground px-2 py-0.5"
                  >
                    {resourcesLoading ? '...' : `${profileSessions.length} total`}
                  </Badge>
                </div>
                <CardDescription className="text-xs text-muted-foreground pt-1">
                  Recent interactive terminal sessions for {currentProfile.name}.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                {resourcesLoading ? (
                  <div className="py-8 text-center text-xs text-muted-foreground font-sans">
                    Loading sessions...
                  </div>
                ) : profileSessions.length === 0 ? (
                  <div className="py-8 text-center text-xs text-muted-foreground font-sans border border-dashed border-border rounded-geist">
                    No recorded sessions for this profile yet.
                  </div>
                ) : (
                  <div className="divide-y divide-border border border-border rounded-geist overflow-hidden">
                    {profileSessions.slice(0, 4).map((sess) => (
                      <div
                        key={sess.id}
                        className="p-3.5 hover:bg-secondary/40 transition-colors flex items-center justify-between gap-3 text-xs"
                      >
                        <div className="min-w-0 flex-1 space-y-1">
                          <div className="flex items-center gap-2">
                            <button
                              type="button"
                              onClick={() => onViewSession?.(sess)}
                              className="font-medium text-foreground hover:underline truncate cursor-pointer text-left font-sans"
                            >
                              {sess.title || sess.goal || sess.id.slice(0, 8)}
                            </button>
                            <span className="font-mono text-[10px] text-muted-foreground shrink-0">
                              {sess.turns} turns
                            </span>
                            {sess.is_active && (
                              <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid shrink-0" title="Active" />
                            )}
                          </div>
                          <div className="flex items-center gap-2 text-muted-foreground text-[11px] font-mono truncate">
                            <span className="truncate">{sess.cwd}</span>
                            <span>•</span>
                            <span className="shrink-0">{formatRelativeTime(sess.updated_at)}</span>
                          </div>
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleQuickResume(sess)}
                          disabled={resumingId === sess.id}
                          className="h-7 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-accent text-foreground cursor-pointer px-2.5 shrink-0"
                        >
                          <Terminal className={`h-3 w-3 mr-1 ${resumingId === sess.id ? 'animate-spin' : ''}`} />
                          {resumingId === sess.id ? 'Launching...' : 'Resume'}
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
              <CardFooter className="p-5 pt-0 flex justify-between items-center text-xs">
                <span className="text-muted-foreground font-sans">
                  Showing {Math.min(4, profileSessions.length)} of {profileSessions.length} sessions
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onNavigateToSessions && onNavigateToSessions(currentProfile.name)}
                  className="h-7 text-xs font-sans text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer px-2"
                >
                  <span>View all in Sessions Explorer</span>
                  <ArrowRight className="h-3 w-3 ml-1" />
                </Button>
              </CardFooter>
            </Card>
          </div>

          {/* Right Column (5 cols, ~40%): Sandbox & Environment, Inheritance, MCP Tools */}
          <div className="lg:col-span-5 space-y-6">
            {/* Profile Environment Sandbox Card */}
            <Card className="border border-border bg-card rounded-geist">
              <CardHeader className="pb-3 p-5">
                <CardTitle className="text-sm font-semibold text-foreground flex items-center gap-2">
                  <Folder className="h-4 w-4 text-muted-foreground" />
                  Sandbox & Environment
                </CardTitle>
                <CardDescription className="text-xs text-muted-foreground">
                  Filesystem sandbox isolation and global configuration inheritance.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0 space-y-4 text-xs font-sans">
                <div>
                  <div className="text-muted-foreground mb-1 text-[11px] uppercase tracking-wider font-semibold">
                    Storage Sandbox
                  </div>
                  <div className="flex items-center gap-1.5 p-2 rounded-geist bg-secondary/60 border border-border">
                    <span className="font-mono text-[11px] text-foreground truncate flex-1 select-all">
                      {currentProfile.path}
                    </span>
                    <button
                      type="button"
                      onClick={() => handleCopyPath(currentProfile.path)}
                      className="p-1 rounded text-muted-foreground hover:text-foreground cursor-pointer"
                      title="Copy path"
                    >
                      {copiedPath ? <Check className="h-3 w-3 text-status-success" /> : <Copy className="h-3 w-3" />}
                    </button>
                  </div>
                </div>

                <div>
                  <div className="text-muted-foreground mb-1.5 text-[11px] uppercase tracking-wider font-semibold">
                    Connected Engines
                  </div>
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

                <div className="pt-2 border-t border-border space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5 pr-2">
                      <div className="font-medium text-foreground">Global MCP Inheritance</div>
                      <div className="text-[11px] text-muted-foreground">Inherit host machine MCP servers</div>
                    </div>
                    <Switch
                      checked={isMcpGlobalOn}
                      disabled={updatingConfig}
                      onCheckedChange={() => handleToggleMcpGlobal(currentProfile.name, isMcpGlobalOn)}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5 pr-2">
                      <div className="font-medium text-foreground">Global Plugin Inheritance</div>
                      <div className="text-[11px] text-muted-foreground">Inherit host machine plugins</div>
                    </div>
                    <Switch
                      checked={isPluginsGlobalOn}
                      disabled={updatingConfig}
                      onCheckedChange={() => handleTogglePluginsGlobal(currentProfile.name, isPluginsGlobalOn)}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* MCP Tools & Catalog Preview */}
            <Card className="border border-border bg-card rounded-geist">
              <CardHeader className="pb-3 p-5">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-semibold text-foreground flex items-center gap-2">
                    <Server className="h-4 w-4 text-muted-foreground" />
                    MCP Tools & Catalog
                  </CardTitle>
                  <Badge
                    variant="outline"
                    className="text-xs font-sans font-medium bg-secondary border-border text-foreground px-2 py-0.5"
                  >
                    {resourcesLoading ? '...' : `${profileMcpServers.length} active`}
                  </Badge>
                </div>
                <CardDescription className="text-xs text-muted-foreground pt-1">
                  Configured MCP servers and tool capabilities for this profile.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0">
                {resourcesLoading ? (
                  <div className="py-6 text-center text-xs text-muted-foreground font-sans">
                    Loading MCP servers...
                  </div>
                ) : profileMcpServers.length === 0 ? (
                  <div className="py-6 text-center text-xs text-muted-foreground font-sans border border-dashed border-border rounded-geist">
                    No MCP servers configured for this profile.
                  </div>
                ) : (
                  <div className="space-y-2">
                    <div className="flex flex-wrap gap-1.5">
                      {profileMcpServers.slice(0, 6).map((srv) => (
                        <span
                          key={srv.name}
                          className="inline-flex items-center px-2 py-0.5 rounded-full border border-border bg-secondary font-mono text-[11px] text-foreground"
                        >
                          {srv.name}
                          {srv.scope && (
                            <span className="ml-1 text-muted-foreground text-[10px]">
                              ({srv.scope})
                            </span>
                          )}
                        </span>
                      ))}
                      {profileMcpServers.length > 6 && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full border border-border bg-secondary font-mono text-[11px] text-muted-foreground">
                          +{profileMcpServers.length - 6} more
                        </span>
                      )}
                    </div>
                  </div>
                )}
              </CardContent>
              <CardFooter className="p-5 pt-0">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onNavigateToMcp && onNavigateToMcp(currentProfile.name)}
                  className="w-full justify-between h-8 rounded-geist border-border bg-secondary text-xs font-sans font-medium text-foreground hover:bg-accent cursor-pointer"
                >
                  <span>Open in MCP Catalog</span>
                  <ArrowRight className="h-3.5 w-3.5 ml-1 text-muted-foreground" />
                </Button>
              </CardFooter>
            </Card>
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
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-base font-semibold text-foreground group-hover:text-foreground transition-colors">
                          {p.name}
                        </CardTitle>
                        <Button
                          variant="ghost"
                          size="icon"
                          title={`Delete profile ${p.name}`}
                          aria-label={`Delete profile ${p.name}`}
                          className="h-7 w-7 text-muted-foreground hover:text-status-danger hover:bg-status-danger-bg rounded-geist transition-colors"
                          onClick={(e) => {
                            e.stopPropagation();
                            openDeleteDialog(p.agent, p.name);
                          }}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                      <CardDescription className="text-xs text-muted-foreground truncate pt-0.5 flex items-center gap-1">
                        <Folder className="h-3 w-3 shrink-0 text-muted-foreground/60" />
                        <span className="truncate font-mono">{p.path}</span>
                      </CardDescription>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3 pb-4 p-5 pt-0">
                    {/* Account Info Box */}
                    <div className="rounded-geist border border-border bg-muted p-3 text-xs space-y-1.5 font-sans">
                      <div className="flex items-center justify-between text-muted-foreground">
                        <span className="flex items-center gap-1.5">
                          <User className="h-3 w-3 text-muted-foreground/60" /> Account
                        </span>
                        <span className="text-foreground font-medium truncate max-w-[140px]">
                          {p.account?.name || p.name}
                        </span>
                      </div>
                      {p.account?.email && (
                        <div className="flex items-center justify-between text-muted-foreground">
                          <span className="flex items-center gap-1.5">
                            <Mail className="h-3 w-3 text-muted-foreground/60" /> Email
                          </span>
                          <span className="text-foreground truncate max-w-[160px]">
                            {p.account.email}
                          </span>
                        </div>
                      )}
                      {p.account?.auth_method && (
                        <div className="flex items-center justify-between text-muted-foreground">
                          <span className="flex items-center gap-1.5">
                            <Key className="h-3 w-3 text-muted-foreground/60" /> Auth
                          </span>
                          <span className="text-foreground truncate max-w-[140px]">
                            {p.account.auth_method}
                          </span>
                        </div>
                      )}
                    </div>

                    {/* Standardized Quota Telemetry (Strict h-1.5 across multi- and single-engine) */}
                    {p.adapters && p.adapters.length > 1 ? (
                      <div className="space-y-2 pt-1 border-t border-border">
                        <div className="flex items-center justify-between text-[11px] font-sans font-medium uppercase tracking-wider text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Engine Quotas
                          </span>
                          <span>{p.adapters.length} Engines</span>
                        </div>
                        {p.adapters.map((ad) => {
                          const pct = ad.quota?.bottleneck_pct ?? 100;
                          return (
                            <div key={ad.agent} className="space-y-1">
                              <div className="flex items-center justify-between text-[11px] font-sans">
                                <span className="text-muted-foreground">{ad.agent}</span>
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
                                className="w-full bg-secondary h-1.5 rounded-full overflow-hidden"
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
                          <span className="text-[11px] font-sans font-medium uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                            <Gauge className="h-3 w-3" /> Quota Telemetry
                          </span>
                        {(() => {
                          const pct = p.quota?.bottleneck_pct ?? 100;
                          return (
                            <span
                              className={`font-sans text-xs font-medium tabular-nums ${
                                pct <= 0
                                  ? 'badge-danger'
                                  : pct <= 30
                                  ? 'badge-warning'
                                  : 'text-muted-foreground'
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
                          className="w-full bg-secondary h-1.5 rounded-full overflow-hidden"
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
                          <p className="text-[11px] text-muted-foreground/70 font-sans truncate" title={p.quota.summary}>
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
