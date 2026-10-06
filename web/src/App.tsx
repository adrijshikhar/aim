import * as React from 'react';
import {
  getStatus,
  getDaemonStatus,
  getProfiles,
  getSessions,
  StatusDTO,
  DaemonDTO,
  ProfileDTO,
  SessionDTO,
} from '@/lib/api';
import { ProfilesView } from '@/views/ProfilesView';
import { SessionsView } from '@/views/SessionsView';
import { SessionDetailView } from '@/views/SessionDetailView';
import { McpView } from '@/views/McpView';
import { Sidebar } from '@/components/Sidebar';
import { DaemonModal } from '@/components/DaemonModal';
import { AddProfileModal } from '@/components/AddProfileModal';
import { Toaster } from '@/components/ui/toaster';
import { ChevronRight } from 'lucide-react';

type ViewType = 'overview' | 'sessions' | 'mcp' | 'profile-detail' | 'session-detail';

function parseCurrentRoute(): {
  view: ViewType;
  profile: string;
  tab: string;
  mcpScope: string;
  sessionId: string | null;
} {
  try {
    const rawPath = window.location.pathname.replace(/\/+$/, '') || '/';
    const params = new URLSearchParams(window.location.search);
    const querySessionId = params.get('session') || params.get('id');
    const viewParam = params.get('view');
    const profileParam = params.get('profile');
    const tabParam = params.get('tab') || 'quota';
    const scopeParam = params.get('scope') || 'all';

    const normalizedTab = tabParam === 'telemetry' ? 'quota' : tabParam;

    // 1. Path-based route: /sessions/:id (Session Detail)
    if (rawPath.startsWith('/sessions/') && rawPath.length > '/sessions/'.length) {
      const sid = decodeURIComponent(rawPath.slice('/sessions/'.length));
      return {
        view: 'session-detail',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam,
        sessionId: sid,
      };
    }

    // 2. Path-based route: /sessions (Sessions Explorer)
    if (rawPath === '/sessions') {
      return {
        view: 'sessions',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam,
        sessionId: null,
      };
    }

    // 3. Path-based route: /mcp (MCP Catalog)
    if (rawPath === '/mcp') {
      return {
        view: 'mcp',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam !== 'all' ? scopeParam : (profileParam && profileParam !== 'all' ? profileParam : 'all'),
        sessionId: null,
      };
    }

    // 4. Path-based route: /profiles/:name or /profiles/:name/:tab
    if (rawPath.startsWith('/profiles/') && rawPath.length > '/profiles/'.length) {
      const segments = rawPath.slice('/profiles/'.length).split('/');
      const profileName = decodeURIComponent(segments[0]);
      let routeTab = normalizedTab;
      if (segments[1]) {
        const segTab = decodeURIComponent(segments[1]);
        routeTab = segTab === 'telemetry' ? 'quota' : segTab;
      }
      return {
        view: 'profile-detail',
        profile: profileName,
        tab: routeTab,
        mcpScope: scopeParam,
        sessionId: null,
      };
    }

    // 5. Query param fallback (backward compatibility for existing bookmarks/links)
    if (querySessionId) {
      return {
        view: 'session-detail',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam,
        sessionId: querySessionId,
      };
    }

    if (viewParam === 'sessions') {
      return {
        view: 'sessions',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam,
        sessionId: null,
      };
    }

    if (viewParam === 'mcp') {
      return {
        view: 'mcp',
        profile: profileParam || 'all',
        tab: normalizedTab,
        mcpScope: scopeParam !== 'all' ? scopeParam : (profileParam && profileParam !== 'all' ? profileParam : 'all'),
        sessionId: null,
      };
    }

    if (profileParam && profileParam !== 'all') {
      return {
        view: 'profile-detail',
        profile: profileParam,
        tab: normalizedTab,
        mcpScope: scopeParam,
        sessionId: null,
      };
    }
  } catch {
    // Fallback if URL parsing fails
  }

  return {
    view: 'overview',
    profile: 'all',
    tab: 'quota',
    mcpScope: 'all',
    sessionId: null,
  };
}

export function App() {
  const initialRoute = React.useMemo(() => parseCurrentRoute(), []);
  const [currentView, setCurrentView] = React.useState<ViewType>(initialRoute.view);
  const [selectedProfile, setSelectedProfile] = React.useState<string>(initialRoute.profile);
  const [selectedMcpScope, setSelectedMcpScope] = React.useState<string>(initialRoute.mcpScope);
  const [activeProfileTab, setActiveProfileTab] = React.useState<string>(initialRoute.tab);
  const [viewedSession, setViewedSession] = React.useState<SessionDTO | null>(null);
  const [targetSessionId, setTargetSessionId] = React.useState<string | null>(initialRoute.sessionId);
  const [previousView, setPreviousView] = React.useState<ViewType>('overview');
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>([]);

  const [status, setStatus] = React.useState<StatusDTO>({
    version: 'v0.12.0',
    status: 'healthy',
    active_sessions: 2,
    total_profiles: 4,
    terminals_available: ['tmux', 'wezterm', 'iterm2', 'terminal'],
  });
  const [isOnline, setIsOnline] = React.useState(true);
  const [daemonInfo, setDaemonInfo] = React.useState<DaemonDTO | null>(null);

  // Modals state
  const [isDaemonOpen, setIsDaemonOpen] = React.useState(false);
  const [isAddProfileOpen, setIsAddProfileOpen] = React.useState(false);

  // URL synchronization helper with clean path navigation
  const syncUrl = React.useCallback(
    (
      view: ViewType,
      profile: string,
      tab: string,
      mcpScope: string,
      sessionId?: string | null,
      replace = false
    ) => {
      try {
        let targetPath = '/';
        const params = new URLSearchParams();

        if (view === 'session-detail' && sessionId) {
          targetPath = `/sessions/${encodeURIComponent(sessionId)}`;
          if (profile && profile !== 'all') {
            params.set('profile', profile);
          }
        } else if (view === 'profile-detail' && profile && profile !== 'all') {
          if (tab && tab !== 'quota') {
            targetPath = `/profiles/${encodeURIComponent(profile)}/${encodeURIComponent(tab)}`;
          } else {
            targetPath = `/profiles/${encodeURIComponent(profile)}`;
          }
        } else if (view === 'sessions') {
          targetPath = '/sessions';
          if (profile && profile !== 'all') {
            params.set('profile', profile);
          }
        } else if (view === 'mcp') {
          targetPath = '/mcp';
          if (mcpScope && mcpScope !== 'all') {
            params.set('scope', mcpScope);
          }
        }

        const queryString = params.toString();
        const targetUrl = queryString ? `${targetPath}?${queryString}` : targetPath;
        const currentFull = `${window.location.pathname}${window.location.search}`;

        if (targetUrl !== currentFull) {
          if (replace) {
            window.history.replaceState({ view, profile, tab, mcpScope, sessionId }, '', targetUrl);
          } else {
            window.history.pushState({ view, profile, tab, mcpScope, sessionId }, '', targetUrl);
          }
        }
      } catch {
        // Ignore in environments where window.history is restricted
      }
    },
    []
  );

  // Popstate listener for browser back/forward buttons
  React.useEffect(() => {
    const handlePopState = () => {
      const route = parseCurrentRoute();
      setCurrentView(route.view);
      setSelectedProfile(route.profile);
      setActiveProfileTab(route.tab);
      setSelectedMcpScope(route.mcpScope);
      if (route.sessionId) {
        setTargetSessionId(route.sessionId);
        if (viewedSession?.id !== route.sessionId) {
          setViewedSession(null);
        }
      } else {
        setViewedSession(null);
        setTargetSessionId(null);
      }
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, [viewedSession]);

  // Fetch session details if targeted by URL
  React.useEffect(() => {
    if (targetSessionId && (!viewedSession || viewedSession.id !== targetSessionId)) {
      getSessions()
        .then((allSessions) => {
          const found = allSessions.find((s) => s.id === targetSessionId);
          if (found) {
            setViewedSession(found);
          } else {
            setViewedSession({
              id: targetSessionId,
              agent: 'claude',
              profile: selectedProfile !== 'all' ? selectedProfile : 'work',
              title: `Session ${targetSessionId.slice(0, 8)}`,
              cwd: '',
              goal: '',
              turns: 0,
              updated_at: new Date().toISOString(),
              is_active: false,
            });
          }
        })
        .catch(() => {
          setViewedSession({
            id: targetSessionId,
            agent: 'claude',
            profile: selectedProfile !== 'all' ? selectedProfile : 'work',
            title: `Session ${targetSessionId.slice(0, 8)}`,
            cwd: '',
            goal: '',
            turns: 0,
            updated_at: new Date().toISOString(),
            is_active: false,
          });
        });
    }
  }, [targetSessionId, viewedSession, selectedProfile]);

  const fetchProfiles = React.useCallback(async () => {
    try {
      const data = await getProfiles();
      setProfiles(data);
    } catch {
      // Handled in view
    }
  }, []);

  const fetchDaemonStatus = React.useCallback(async () => {
    try {
      const d = await getDaemonStatus();
      setDaemonInfo(d);
    } catch {
      // Daemon status fetch failed or server down
    }
  }, []);

  React.useEffect(() => {
    async function checkStatus() {
      try {
        const s = await getStatus();
        setStatus(s);
        setIsOnline(true);
      } catch {
        setIsOnline(false);
      }
    }
    checkStatus();
    fetchDaemonStatus();
    fetchProfiles();
    const interval = setInterval(() => {
      checkStatus();
      fetchDaemonStatus();
      fetchProfiles();
    }, 30000);
    return () => clearInterval(interval);
  }, [fetchDaemonStatus, fetchProfiles]);

  // Navigation handlers
  const handleSelectView = (view: 'overview' | 'sessions' | 'mcp') => {
    setViewedSession(null);
    setTargetSessionId(null);
    if (view === 'overview') {
      setSelectedProfile('all');
      setCurrentView('overview');
      syncUrl('overview', 'all', 'quota', 'all', null);
    } else if (view === 'sessions') {
      setSelectedProfile('all');
      setCurrentView('sessions');
      syncUrl('sessions', 'all', 'quota', 'all', null);
    } else if (view === 'mcp') {
      setSelectedMcpScope('all');
      setCurrentView('mcp');
      syncUrl('mcp', 'all', 'quota', 'all', null);
    }
  };

  const handleSelectProfile = (profileName: string) => {
    setViewedSession(null);
    setTargetSessionId(null);
    if (profileName === 'all') {
      setSelectedProfile('all');
      setCurrentView('overview');
      syncUrl('overview', 'all', 'quota', 'all', null);
    } else {
      setSelectedProfile(profileName);
      setCurrentView('profile-detail');
      syncUrl('profile-detail', profileName, activeProfileTab, 'all', null);
    }
  };

  const handleNavigateToSessions = (profileName: string) => {
    setViewedSession(null);
    setTargetSessionId(null);
    setSelectedProfile(profileName);
    setCurrentView('sessions');
    syncUrl('sessions', profileName, 'quota', 'all', null);
  };

  const handleNavigateToMcp = (profileName: string) => {
    setViewedSession(null);
    setTargetSessionId(null);
    setSelectedMcpScope(profileName);
    setCurrentView('mcp');
    syncUrl('mcp', 'all', 'quota', profileName, null);
  };

  const handleViewSession = (session: SessionDTO) => {
    setPreviousView(currentView);
    setViewedSession(session);
    setTargetSessionId(session.id);
    setCurrentView('session-detail');
    syncUrl('session-detail', session.profile || selectedProfile, activeProfileTab, selectedMcpScope, session.id);
  };

  const handleBackFromSession = () => {
    setViewedSession(null);
    setTargetSessionId(null);
    if (previousView === 'profile-detail' && selectedProfile && selectedProfile !== 'all') {
      setCurrentView('profile-detail');
      syncUrl('profile-detail', selectedProfile, activeProfileTab, 'all', null);
    } else {
      setCurrentView('sessions');
      syncUrl('sessions', selectedProfile, 'quota', 'all', null);
    }
  };

  const handleProfileTabChange = (tab: string) => {
    setActiveProfileTab(tab);
    if (currentView === 'profile-detail') {
      syncUrl('profile-detail', selectedProfile, tab, 'all', null, true);
    }
  };

  const getHeaderTitle = () => {
    switch (currentView) {
      case 'overview':
        return (
          <div className="flex items-center gap-1.5 text-sm font-mono">
            <span className="text-[#ededed] font-semibold">All Agent Profiles</span>
          </div>
        );
      case 'profile-detail':
        return (
          <div className="flex items-center gap-1.5 text-sm font-mono">
            <button
              type="button"
              onClick={() => handleSelectView('overview')}
              className="text-[#a1a1a1] hover:text-[#ededed] cursor-pointer"
            >
              Profiles
            </button>
            <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
            <span className="text-[#ededed] font-semibold">{selectedProfile}</span>
          </div>
        );
      case 'sessions':
        return (
          <div className="flex items-center gap-1.5 text-sm font-mono">
            <span className="text-[#ededed] font-semibold">Sessions Explorer</span>
            {selectedProfile !== 'all' && (
              <>
                <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
                <span className="text-[#a1a1a1]">{selectedProfile}</span>
              </>
            )}
          </div>
        );
      case 'session-detail':
        return (
          <div className="flex items-center gap-1.5 text-sm font-mono">
            <button
              type="button"
              onClick={handleBackFromSession}
              className="text-[#a1a1a1] hover:text-[#ededed] cursor-pointer"
            >
              Sessions
            </button>
            <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
            <span className="text-[#a1a1a1]">{viewedSession?.agent}</span>
            <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
            <span className="text-[#ededed] font-semibold">
              {viewedSession ? viewedSession.id.slice(0, 8) : targetSessionId ? targetSessionId.slice(0, 8) : 'Detail'}
            </span>
          </div>
        );
      case 'mcp':
        return (
          <div className="flex items-center gap-1.5 text-sm font-mono">
            <span className="text-[#ededed] font-semibold">MCP Catalog</span>
            {selectedMcpScope !== 'all' && (
              <>
                <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
                <span className="text-[#a1a1a1]">{selectedMcpScope}</span>
              </>
            )}
          </div>
        );
    }
  };

  return (
    <div className="flex h-screen w-full bg-background text-foreground overflow-hidden font-sans">
      {/* Desktop Left Sidebar */}
      <Sidebar
        currentView={currentView}
        selectedProfile={selectedProfile}
        profiles={profiles}
        daemonInfo={daemonInfo}
        isOnline={isOnline}
        version={status.version}
        onSelectView={handleSelectView}
        onSelectProfile={handleSelectProfile}
        onOpenAddProfile={() => setIsAddProfileOpen(true)}
        onOpenDaemonModal={() => setIsDaemonOpen(true)}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden bg-background">
        {/* Top Header Bar */}
        <header className="h-14 border-b border-border bg-background/90 backdrop-blur-md px-6 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3">
            {getHeaderTitle()}
          </div>

          {/* Right Header Status Pills */}
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              onClick={() => setIsDaemonOpen(true)}
              title="Manage OS Background Daemon (15m Quota Pre-Warm)"
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-border bg-card hover:bg-secondary hover:border-border/80 transition-colors cursor-pointer text-xs group"
            >
              <span className="flex h-2 w-2 relative items-center justify-center">
                {daemonInfo?.active ? (
                  <span className="h-2 w-2 rounded-full bg-status-success-solid" />
                ) : daemonInfo?.installed ? (
                  <span className="h-2 w-2 rounded-full bg-status-warning-solid" />
                ) : (
                  <span className="h-2 w-2 rounded-full bg-muted-foreground/40" />
                )}
              </span>
              <span className="font-mono text-[11px] text-muted-foreground group-hover:text-foreground transition-colors">
                {daemonInfo?.active
                  ? 'Daemon: Active'
                  : daemonInfo?.installed
                  ? 'Daemon: Inactive'
                  : 'Daemon: Off'}
              </span>
            </button>

            <div className="hidden sm:flex items-center gap-2 px-2.5 py-1 rounded-full border border-border bg-card text-xs">
              <span className="relative flex h-2 w-2">
                <span
                  className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                    isOnline ? 'bg-status-success-solid' : 'bg-status-warning-solid'
                  }`}
                />
                <span
                  className={`relative inline-flex rounded-full h-2 w-2 ${
                    isOnline ? 'bg-status-success-solid' : 'bg-status-warning-solid'
                  }`}
                />
              </span>
              <span className="font-mono text-[11px] text-foreground">
                {isOnline ? 'Server Connected' : 'Connecting...'}
              </span>
            </div>
          </div>
        </header>

        {/* Stage Content */}
        <main className="flex-1 overflow-y-auto px-6 py-6 min-h-0 bg-background">
          <div className="max-w-7xl mx-auto">
            {currentView === 'overview' && (
              <ProfilesView
                selectedProfile="all"
                profiles={profiles}
                onSelectProfile={handleSelectProfile}
                onNavigateToSessions={handleNavigateToSessions}
                onNavigateToMcp={handleNavigateToMcp}
                onProfilesLoaded={setProfiles}
              />
            )}
            {currentView === 'profile-detail' && (
              <ProfilesView
                selectedProfile={selectedProfile}
                profiles={profiles}
                activeTab={activeProfileTab}
                onTabChange={handleProfileTabChange}
                onSelectProfile={handleSelectProfile}
                onNavigateToSessions={handleNavigateToSessions}
                onNavigateToMcp={handleNavigateToMcp}
                onProfilesLoaded={setProfiles}
                onViewSession={handleViewSession}
              />
            )}
            {currentView === 'sessions' && (
              <SessionsView
                selectedProfile={selectedProfile}
                onSelectProfile={setSelectedProfile}
                profiles={profiles}
                onViewSession={handleViewSession}
              />
            )}
            {currentView === 'session-detail' && (
              viewedSession ? (
                <SessionDetailView
                  key={viewedSession.id}
                  session={viewedSession}
                  onBack={handleBackFromSession}
                />
              ) : (
                <div className="py-24 text-center">
                  <div className="inline-flex items-center gap-2 text-xs font-mono text-[#888888]">
                    <span className="h-3 w-3 rounded-full border-2 border-[#0070f3] border-t-transparent animate-spin" />
                    <span>Loading session {targetSessionId?.slice(0, 8)}...</span>
                  </div>
                </div>
              )
            )}
            {currentView === 'mcp' && (
              <McpView
                selectedScope={selectedMcpScope}
                onSelectScope={(scope) => {
                  setSelectedMcpScope(scope);
                  syncUrl('mcp', 'all', 'quota', scope, null, true);
                }}
                profiles={profiles}
              />
            )}
          </div>
        </main>
      </div>

      {/* Global Modals */}
      <DaemonModal
        open={isDaemonOpen}
        onOpenChange={setIsDaemonOpen}
        daemonInfo={daemonInfo}
        onRefreshDaemon={fetchDaemonStatus}
      />

      <AddProfileModal
        open={isAddProfileOpen}
        onOpenChange={setIsAddProfileOpen}
        onProfileCreated={() => {
          fetchProfiles();
        }}
      />

      <Toaster />
    </div>
  );
}

export default App;
