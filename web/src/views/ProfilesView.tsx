import * as React from 'react';
import {
  ProfileDTO,
  SessionDTO,
  MCPServerDTO,
  ResumeRequest,
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
  ChevronDown,
  X,
  Boxes,
  Copy,
  Check,
  Settings,
  Search,
  SlidersHorizontal,
  ShieldAlert,
  AlertCircle,
  AlertTriangle,
  Activity,
  LayoutGrid,
  List,
  Clock,
} from 'lucide-react';
import { DeleteProfileModal } from '@/components/DeleteProfileModal';
import { AddProfileModal } from '@/components/AddProfileModal';
import { ResumeModal } from '@/components/ResumeModal';
import { AgentBadge } from '@/components/AgentBadge';
import { QuotaBreakdown } from '@/components/QuotaBreakdown';
import { cn } from '@/lib/utils';

interface ParsedQuotaDetails {
  resetTime: string | null;
  shortReset: string | null;
  weeklyReset: string | null;
  headline: string;
  subtext: string;
}

function parseQuotaDetails(summary?: string, isExhausted?: boolean, pct: number = 100): ParsedQuotaDetails {
  if (!summary || !summary.trim()) {
    return {
      resetTime: null,
      shortReset: null,
      weeklyReset: null,
      headline: isExhausted ? 'Exhausted' : 'Healthy',
      subtext: isExhausted ? 'Usage ceiling reached' : 'Standard capacity',
    };
  }

  // 1. Extract reset times:
  // Check 5h window reset: e.g. "5h: 0% [1h 29m]" or "5h limit: 68% remaining [2h 15m]"
  const fiveHMatch = summary.match(/5h[^\]]*?\[([^\]]+)\]/i);
  // Check weekly reset: e.g. "wk: 0% [1d 19h]" or "Weekly Limit: 100% [4d 2h]"
  const wkMatch = summary.match(/(?:wk|week|weekly)[^\]]*?\[([^\]]+)\]/i);
  // Any bracket fallback:
  const anyBracket = summary.match(/\[([^\]]+)\]/);

  const shortReset = fiveHMatch ? fiveHMatch[1].trim() : null;
  const weeklyReset = wkMatch ? wkMatch[1].trim() : null;
  const resetTime = shortReset || weeklyReset || (anyBracket ? anyBracket[1].trim() : null);

  // 2. Identify bottleneck model if multi-model: e.g. "Gemini (5h: 0%, wk: 0% ...)"
  let bottleneckModel: string | null = null;
  const modelMatch = summary.match(/([A-Za-z0-9_-]+)\s*\([^)]*?(?:0%|1\d%|2\d%)[^)]*?\)/);
  if (modelMatch) {
    bottleneckModel = modelMatch[1];
  }

  let headline = '';
  let subtext = '';

  if (resetTime) {
    headline = `Resets in ${resetTime}`;
    if (bottleneckModel) {
      subtext = `${bottleneckModel} capacity bottleneck`;
    } else if (shortReset) {
      subtext = '5-hour sliding window reset';
    } else if (weeklyReset) {
      subtext = 'Weekly allowance reset';
    } else {
      subtext = 'Next auto-renewal window';
    }
  } else {
    if (isExhausted || pct <= 0) {
      headline = 'Quota Exhausted';
      subtext = bottleneckModel ? `${bottleneckModel} limit reached` : 'Usage ceiling reached';
    } else if (pct <= 30) {
      headline = `${pct}% Headroom`;
      subtext = bottleneckModel ? `${bottleneckModel} throttled` : 'Approaching rolling limit';
    } else {
      headline = 'Standard Capacity';
      subtext = 'All models operating normally';
    }
  }

  return {
    resetTime,
    shortReset,
    weeklyReset,
    headline,
    subtext,
  };
}

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
    return 'bg-status-success-solid';
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
  const [flagModalSession, setFlagModalSession] = React.useState<SessionDTO | null>(null);

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
  const [profileSearchQuery, setProfileSearchQuery] = React.useState('');
  const [profilesViewMode, setProfilesViewMode] = React.useState<'grid' | 'list'>('grid');
  const profileSearchInputRef = React.useRef<HTMLInputElement>(null);

  // Global hotkey '/' focuses profile search when viewing all profiles
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (document.querySelector('[role="dialog"]')) return;
      if (
        e.key === '/' &&
        document.activeElement?.tagName !== 'INPUT' &&
        document.activeElement?.tagName !== 'TEXTAREA'
      ) {
        if (!selectedProfile || selectedProfile === 'all') {
          e.preventDefault();
          profileSearchInputRef.current?.focus();
        }
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedProfile]);

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

  const filteredProfilesList = React.useMemo(() => {
    if (!profileSearchQuery.trim()) return profiles;
    const q = profileSearchQuery.toLowerCase().trim();
    return profiles.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.agent.toLowerCase().includes(q) ||
        p.path.toLowerCase().includes(q) ||
        p.account?.name?.toLowerCase().includes(q) ||
        p.account?.email?.toLowerCase().includes(q) ||
        (p.adapters && p.adapters.some((ad) => ad.agent.toLowerCase().includes(q)))
    );
  }, [profiles, profileSearchQuery]);

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

  const handleModalLaunch = async (req: ResumeRequest) => {
    setResumingId(req.session_id);
    try {
      const res = await resumeSession(req);
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
              <span className="text-xs font-mono text-[#888888] px-2 py-0.5 rounded-geist border border-[#222222] bg-[#0c0c0c]">
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
            </div>
          </div>

          {/* Sandbox & Environment Sub-bar with low-contrast labels & high-contrast values */}
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs font-sans">
            <div className="flex items-center gap-1.5 max-w-md">
              <Folder className="h-3.5 w-3.5 text-[#666666] shrink-0" />
              <span className="text-[#666666] shrink-0">Sandbox:</span>
              <span className="text-[#ededed] font-mono truncate select-all" title={currentProfile.path}>
                {currentProfile.path}
              </span>
              <button
                type="button"
                onClick={() => handleCopyPath(currentProfile.path)}
                title={copiedPath ? 'Copied to clipboard' : 'Copy sandbox path'}
                className="p-1 rounded hover:bg-[#1a1a1a] text-[#666666] hover:text-[#ededed] cursor-pointer transition-colors shrink-0"
              >
                {copiedPath ? (
                  <span className="flex items-center gap-1 text-[10px] text-status-success font-sans font-medium">
                    <Check className="h-3 w-3" /> Copied!
                  </span>
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
        </div>

        {/* Multi-Tab Workspace Organization */}
        <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full space-y-6">
          <div className="flex items-center justify-between">
            <TabsList className="bg-[#0c0c0c] border border-[#222222] p-1 h-9 rounded-geist gap-1">
              <TabsTrigger value="quota" className="gap-2 text-xs">
                <Gauge className="h-3.5 w-3.5 text-[#666666]" />
                <span>Quota</span>
                {profileAdapters.length > 1 && (
                  <span className="text-[10px] px-1.5 py-0.5 rounded-geist bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                    {profileAdapters.length}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger value="sessions" className="gap-2 text-xs">
                <Terminal className="h-3.5 w-3.5 text-[#666666]" />
                <span>Sessions</span>
                <span className="text-[10px] px-1.5 py-0.5 rounded-geist bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                  {profileSessions.length}
                </span>
              </TabsTrigger>
              <TabsTrigger value="mcp" className="gap-2 text-xs">
                <Server className="h-3.5 w-3.5 text-[#666666]" />
                <span>MCP Tools</span>
                <span className="text-[10px] px-1.5 py-0.5 rounded-geist bg-[#1c1c1c] text-[#888888] border border-[#2e2e2e]">
                  {profileMcpServers.length}
                </span>
              </TabsTrigger>
              <TabsTrigger value="settings" className="gap-2 text-xs">
                <Settings className="h-3.5 w-3.5 text-[#666666]" />
                <span>Settings</span>
              </TabsTrigger>
            </TabsList>
          </div>

          {/* TAB 1: QUOTA (Vercel Speed Insights Master-Detail Architecture) */}
          <TabsContent value="quota" className="space-y-6 focus-visible:outline-none">
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
                    ? `Viewing combined quota for ${profileAdapters.length} configured engines.`
                    : `Focused quota view for ${selectedEngineTab}.`}
                </span>
              </div>
            )}

            {/* Vercel Speed Insights 2-Column Master-Detail Layout */}
            <div className="space-y-6">
              {displayedAdapters.map((ad) => {
                const pct = ad.quota?.bottleneck_pct ?? 100;
                const isExhausted = ad.quota?.is_exhausted || pct <= 0;
                const isWarning = !isExhausted && pct <= 30;

                return (
                  <div key={ad.agent} className="space-y-4">
                    {profileAdapters.length > 1 && selectedEngineTab === 'all' && (
                      <div className="flex items-center gap-2 pt-2 border-t border-[#222222]">
                        <span className="text-xs font-semibold text-[#ededed] uppercase tracking-wider font-mono">
                          Engine: {ad.agent}
                        </span>
                        {getAgentBadge(ad.agent)}
                      </div>
                    )}

                    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
                      {/* LEFT COLUMN: Metric Capacity & Status (~320px / 4 cols) */}
                      <div className="lg:col-span-4 space-y-4">
                        {(() => {
                          const qDetails = parseQuotaDetails(ad.quota?.summary, isExhausted, pct);

                          return (
                            <>
                              {/* Hero Capacity Index Card */}
                              <div className="p-4 rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] space-y-3">
                                <div className="flex items-center justify-between">
                                  <span className="text-xs font-medium text-[#888888]">Overall Capacity</span>
                                  <div className="h-6 w-6 rounded-full bg-[#181818] border border-[#2e2e2e] flex items-center justify-center text-[#ededed]">
                                    <Activity className="h-3 w-3" />
                                  </div>
                                </div>

                                <div>
                                  <div className="flex items-baseline justify-between gap-2">
                                    <div className="text-2xl font-bold font-mono tracking-tight text-[#ededed]">
                                      {pct}%
                                    </div>
                                    {qDetails.resetTime && (
                                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-[4px] border border-amber-500/25 bg-amber-500/10 text-amber-400 font-mono text-[11px] font-medium shrink-0">
                                        <Clock className="h-3 w-3 text-amber-400 shrink-0" />
                                        <span>Resets in {qDetails.resetTime}</span>
                                      </span>
                                    )}
                                  </div>
                                  <div className="text-[11px] text-[#888888] font-sans truncate pt-0.5" title={ad.quota?.summary || 'Standard capacity'}>
                                    {qDetails.subtext}
                                  </div>
                                </div>

                                {/* Micro-bar */}
                                <div className="space-y-1">
                                  <div className="w-full bg-[#181818] h-1.5 rounded-full overflow-hidden border border-[#262626]/50">
                                    <div
                                      className={`h-full transition-all duration-300 ${getProgressColor(pct)}`}
                                      style={{ width: `${isExhausted || pct <= 0 ? 0 : Math.min(100, pct)}%` }}
                                    />
                                  </div>
                                </div>

                                <div className="pt-2 border-t border-[#1a1a1a] flex items-center justify-between text-[11px]">
                                  <span className="text-[#666666]">Status</span>
                                  {isExhausted ? (
                                    <span className="font-mono text-status-danger font-medium flex items-center gap-1">
                                      <span className="h-1.5 w-1.5 rounded-full bg-status-danger-solid" />
                                      Exhausted
                                    </span>
                                  ) : isWarning ? (
                                    <span className="font-mono text-status-warning font-medium flex items-center gap-1">
                                      <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                                      Throttled / Low
                                    </span>
                                  ) : (
                                    <span className="font-mono text-status-success font-medium flex items-center gap-1">
                                      <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                                      Healthy
                                    </span>
                                  )}
                                </div>
                              </div>

                              {/* Semantic Alert / Live Status Card */}
                              {isExhausted ? (
                                <div className="rounded-geist border border-status-danger-border bg-status-danger-bg p-3 text-xs flex items-center gap-2.5 text-status-danger font-mono">
                                  <AlertCircle className="h-4 w-4 shrink-0 text-status-danger" />
                                  <span>Rate limit window exhausted. Quota resets in approximately {qDetails.resetTime || '45m'}.</span>
                                </div>
                              ) : isWarning ? (
                                <div className="rounded-geist border border-status-warning-border bg-status-warning-bg p-3 text-xs flex items-center gap-2.5 text-status-warning font-mono">
                                  <AlertTriangle className="h-4 w-4 shrink-0 text-status-warning" />
                                  <span>High quota consumption detected ({pct}% remaining). Approaching window limit.</span>
                                </div>
                              ) : (
                                <div className="rounded-geist border border-status-success-border bg-status-success-bg p-3 text-xs flex items-center justify-between text-status-success font-mono">
                                  <div className="flex items-center gap-2">
                                    <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                                    <span>All quota windows operating within normal limits</span>
                                  </div>
                                </div>
                              )}
                            </>
                          );
                        })()}
                      </div>

                      {/* RIGHT COLUMN: Engine Quotas Breakdown (8 cols) */}
                      <div className="lg:col-span-8">
                        <QuotaBreakdown
                          summary={ad.quota?.summary}
                          bottleneckPct={pct}
                          isExhausted={isExhausted}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
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

                    <div className="flex items-center gap-2 shrink-0" onClick={(e) => e.stopPropagation()}>
                      <div className="inline-flex items-center rounded-geist border border-[#2a2a2a] bg-[#121212] hover:bg-[#181818] hover:border-[#383838] transition-colors shadow-sm">
                        <button
                          type="button"
                          onClick={() => handleQuickResume(sess)}
                          disabled={resumingId === sess.id}
                          className="h-7 pl-2.5 pr-2 text-xs font-sans font-medium text-[#ededed] hover:text-white flex items-center gap-1.5 transition-colors cursor-pointer disabled:opacity-50"
                        >
                          <Terminal className={`h-3 w-3 text-[#888888] ${resumingId === sess.id ? 'animate-spin text-[#ededed]' : ''}`} />
                          <span>{resumingId === sess.id ? 'Launching...' : 'Resume'}</span>
                        </button>
                        <div className="h-3.5 w-[1px] bg-[#2a2a2a]" />
                        <button
                          type="button"
                          onClick={() => setFlagModalSession(sess)}
                          className="h-7 px-1.5 text-[#888888] hover:text-[#ededed] hover:bg-[#202020] transition-colors flex items-center justify-center cursor-pointer"
                          title="Resume with flags"
                        >
                          <ChevronDown className="h-3 w-3" />
                        </button>
                      </div>
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

            {/* Global Workspace Inheritance Card */}
            <Card className="border border-[#222222] bg-[#0a0a0a] rounded-geist">
              <CardHeader className="p-5 pb-3">
                <CardTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
                  <SlidersHorizontal className="h-4 w-4 text-[#666666]" />
                  Global Workspace Inheritance
                </CardTitle>
                <CardDescription className="text-xs text-[#888888] pt-1">
                  Configure whether this profile inherits tools and plugins from global host settings.
                </CardDescription>
              </CardHeader>
              <CardContent className="p-5 pt-0 space-y-3">
                <div className="p-3 rounded-geist bg-[#0f0f0f] border border-[#222222] flex items-center justify-between gap-4">
                  <div className="space-y-0.5">
                    <div className="text-xs font-semibold text-[#ededed]">Global MCP Servers</div>
                    <div className="text-[11px] text-[#666666]">
                      Inherit centrally installed MCP servers from host environment.
                    </div>
                  </div>
                  <Switch
                    size="sm"
                    checked={isMcpGlobalOn}
                    disabled={updatingConfig}
                    onCheckedChange={() => handleToggleMcpGlobal(currentProfile.name, isMcpGlobalOn)}
                  />
                </div>

                <div className="p-3 rounded-geist bg-[#0f0f0f] border border-[#222222] flex items-center justify-between gap-4">
                  <div className="space-y-0.5">
                    <div className="text-xs font-semibold text-[#ededed]">Global Plugins</div>
                    <div className="text-[11px] text-[#666666]">
                      Inherit global plugins from host environment into this profile.
                    </div>
                  </div>
                  <Switch
                    size="sm"
                    checked={isPluginsGlobalOn}
                    disabled={updatingConfig}
                    onCheckedChange={() => handleTogglePluginsGlobal(currentProfile.name, isPluginsGlobalOn)}
                  />
                </div>
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

        <ResumeModal
          open={!!flagModalSession}
          onOpenChange={(open) => {
            if (!open) setFlagModalSession(null);
          }}
          session={flagModalSession}
          onLaunch={handleModalLaunch}
          isResuming={resumingId === flagModalSession?.id}
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
      {/* Top action row with Vercel search bar, view switcher, and solid white CTA */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search input with '/' hotkey shortcut chip */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60" />
          <Input
            ref={profileSearchInputRef}
            placeholder="Search profiles by name, agent, account, or path..."
            value={profileSearchQuery}
            onChange={(e) => setProfileSearchQuery(e.target.value)}
            className="pl-8 pr-8 h-8 text-xs font-sans bg-[#0c0c0c] border-[#222222] text-[#ededed] placeholder:text-muted-foreground/50 rounded-geist focus:border-[#333333]"
          />
          {profileSearchQuery ? (
            <button
              type="button"
              onClick={() => setProfileSearchQuery('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground cursor-pointer"
              title="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          ) : (
            <kbd className="absolute right-2.5 top-1/2 -translate-y-1/2 border border-[#2a2a2a] bg-[#161616] px-1 py-[1px] rounded text-[10px] font-mono text-muted-foreground pointer-events-none">
              /
            </kbd>
          )}
        </div>

        {/* View Mode Switcher, Refresh, and Solid White Add Profile Button */}
        <div className="flex items-center gap-2 justify-end">
          {/* Grid / List view toggle */}
          <div className="inline-flex items-center rounded-geist border border-[#222222] bg-[#0c0c0c] p-0.5">
            <button
              type="button"
              onClick={() => setProfilesViewMode('grid')}
              className={cn(
                'h-7 w-7 rounded-[4px] flex items-center justify-center cursor-pointer transition-colors',
                profilesViewMode === 'grid'
                  ? 'bg-[#1f1f1f] text-[#ededed]'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              title="Grid view"
            >
              <LayoutGrid className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={() => setProfilesViewMode('list')}
              className={cn(
                'h-7 w-7 rounded-[4px] flex items-center justify-center cursor-pointer transition-colors',
                profilesViewMode === 'list'
                  ? 'bg-[#1f1f1f] text-[#ededed]'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              title="List view"
            >
              <List className="h-3.5 w-3.5" />
            </button>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchProfiles(true)}
            disabled={loading || refreshing}
            className="rounded-geist h-8 border-[#262626] bg-[#0c0c0c] text-foreground hover:bg-[#161616] cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${refreshing || loading ? 'animate-spin' : ''}`} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </Button>

          <Button
            size="sm"
            onClick={() => setIsAddOpen(true)}
            className="bg-[#ffffff] text-[#000000] hover:bg-[#eaeaea] rounded-geist h-8 font-medium cursor-pointer shadow-sm transition-colors"
          >
            <Plus className="h-3.5 w-3.5 mr-1 text-[#000000]" />
            Add Profile
          </Button>

          <AddProfileModal
            open={isAddOpen}
            onOpenChange={setIsAddOpen}
            onProfileCreated={fetchProfiles}
          />
        </div>
      </div>

      {/* Profiles Directory Content */}
      {loading && profiles.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-64 rounded-geist border border-[#262626] bg-[#111111]/40 animate-pulse p-6" />
          ))}
        </div>
      ) : filteredProfilesList.length === 0 ? (
        <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0a0a0a]/50">
          <Bot className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-[#ededed]">
            {profileSearchQuery ? `No profiles match "${profileSearchQuery}"` : 'No profiles found'}
          </h3>
          <p className="text-xs text-[#a1a1a1] mt-1 max-w-sm mx-auto">
            {profileSearchQuery
              ? 'Try clearing the search filter to see all active profiles.'
              : 'Click "Add Profile" to create an isolated environment for your agent.'}
          </p>
        </div>
      ) : profilesViewMode === 'list' ? (
        /* Vercel Table/List View */
        <div className="rounded-geist border border-[#222222] bg-[#0a0a0a] overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-sans">
              <thead className="bg-[#0c0c0c] border-b border-[#222222] text-[11px] font-mono text-[#888888] uppercase">
                <tr>
                  <th className="px-4 py-2.5 font-medium">Profile</th>
                  <th className="px-4 py-2.5 font-medium">Engines</th>
                  <th className="px-4 py-2.5 font-medium">Status</th>
                  <th className="px-4 py-2.5 font-medium">Account</th>
                  <th className="px-4 py-2.5 font-medium">Quota Headroom</th>
                  <th className="px-4 py-2.5 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#1c1c1c]">
                {filteredProfilesList.map((p) => {
                  const pct = p.quota?.bottleneck_pct ?? 100;
                  return (
                    <tr
                      key={p.name}
                      onClick={() => onSelectProfile(p.name)}
                      className="hover:bg-[#111111] transition-colors cursor-pointer group"
                    >
                      <td className="px-4 py-3">
                        <div className="font-semibold text-[#ededed] group-hover:underline">
                          {p.name}
                        </div>
                        <div className="text-[11px] font-mono text-[#666666] truncate max-w-[200px]" title={p.path}>
                          {p.path}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1 flex-wrap">
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
                      </td>
                      <td className="px-4 py-3">
                        {p.has_credentials ? (
                          <span className="inline-flex items-center gap-1.5 text-status-success font-medium text-xs">
                            <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                            Authenticated
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1.5 text-status-warning font-medium text-xs">
                            <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                            Needs Auth
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <div className="text-[#ededed] font-medium truncate max-w-[160px]">
                          {p.account?.name || p.name}
                        </div>
                        {p.account?.email && (
                          <div className="text-[11px] text-[#888888] truncate max-w-[160px]">
                            {p.account.email}
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3 min-w-[160px]">
                        <div className="flex items-center justify-between text-[11px] mb-1">
                          <span className="text-[#666666] font-mono">Headroom</span>
                          <span className={cn('font-mono font-medium', getTextColor(pct))}>
                            {pct}%
                          </span>
                        </div>
                        <div className="w-full bg-[#181818] h-1.5 rounded-full overflow-hidden border border-[#262626]/50">
                          <div
                            className={`h-full transition-all duration-300 ${getProgressColor(pct)}`}
                            style={{
                              width: `${
                                p.quota?.is_exhausted || pct <= 0 ? 0 : Math.min(100, pct)
                              }%`,
                            }}
                          />
                        </div>
                      </td>
                      <td className="px-4 py-3 text-right">
                        <span className="inline-flex items-center text-[#ededed] group-hover:translate-x-0.5 transition-transform font-medium">
                          Inspect <ChevronRight className="h-3.5 w-3.5 ml-0.5" />
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        /* Vercel Grid View */
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredProfilesList.map((p) => {
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
                          className="badge-success text-[11px] font-sans font-medium tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                          <span>Authenticated</span>
                        </Badge>
                      ) : (
                        <Badge
                          variant="outline"
                          className="badge-warning text-[11px] font-sans font-medium tracking-tight flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
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
                                <div className="flex items-center gap-1.5">
                                  {(() => {
                                    const adDetails = parseQuotaDetails(ad.quota?.summary, ad.quota?.is_exhausted || pct <= 0, pct);
                                    return adDetails.resetTime ? (
                                      <span className="inline-flex items-center gap-0.5 text-[10px] font-mono text-amber-400">
                                        <Clock className="h-2.5 w-2.5 text-amber-400 shrink-0" />
                                        <span>{adDetails.resetTime}</span>
                                      </span>
                                    ) : null;
                                  })()}
                                  <span className={`tabular-nums ${getTextColor(pct)}`}>
                                    {pct}%
                                  </span>
                                </div>
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
                                    width: `${
                                      ad.quota?.is_exhausted || pct <= 0 ? 0 : Math.min(100, pct)
                                    }%`,
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
                              width: `${
                                p.quota?.is_exhausted || (p.quota?.bottleneck_pct ?? 100) <= 0
                                  ? 0
                                  : Math.min(100, p.quota?.bottleneck_pct ?? 100)
                              }%`,
                            }}
                          />
                        </div>
                        {p.quota && (() => {
                          const qDetails = parseQuotaDetails(
                            p.quota.summary,
                            p.quota.is_exhausted || (p.quota.bottleneck_pct ?? 100) <= 0,
                            p.quota.bottleneck_pct ?? 100
                          );
                          return (
                            <div className="pt-0.5 flex items-center justify-between text-[11px] font-sans" title={p.quota.summary}>
                              {qDetails.resetTime ? (
                                <span className="inline-flex items-center gap-1 font-mono text-amber-400 font-medium text-[10.5px]">
                                  <Clock className="h-3 w-3 text-amber-400 shrink-0" />
                                  <span>Resets in {qDetails.resetTime}</span>
                                </span>
                              ) : (
                                <span className="text-muted-foreground/70 truncate">{qDetails.subtext}</span>
                              )}
                              {qDetails.resetTime && (
                                <span className="text-[10px] text-muted-foreground/60 truncate font-mono max-w-[130px] text-right">
                                  {qDetails.subtext}
                                </span>
                              )}
                            </div>
                          );
                        })()}
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

      <ResumeModal
        open={!!flagModalSession}
        onOpenChange={(open) => {
          if (!open) setFlagModalSession(null);
        }}
        session={flagModalSession}
        onLaunch={handleModalLaunch}
        isResuming={resumingId === flagModalSession?.id}
      />
    </div>
  );
}
