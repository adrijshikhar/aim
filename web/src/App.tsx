import * as React from 'react';
import {
  getStatus,
  getDaemonStatus,
  getProfiles,
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

export function App() {
  const [currentView, setCurrentView] = React.useState<
    'overview' | 'sessions' | 'mcp' | 'profile-detail' | 'session-detail'
  >('overview');
  const [selectedProfile, setSelectedProfile] = React.useState<string>('all');
  const [selectedMcpScope, setSelectedMcpScope] = React.useState<string>('all');
  const [profiles, setProfiles] = React.useState<ProfileDTO[]>([]);
  const [viewedSession, setViewedSession] = React.useState<SessionDTO | null>(null);

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
        setIsOnline(true);
      }
    }
    checkStatus();
    fetchDaemonStatus();
    fetchProfiles();
    const interval = setInterval(() => {
      checkStatus();
      fetchDaemonStatus();
    }, 30000);
    return () => clearInterval(interval);
  }, [fetchDaemonStatus, fetchProfiles]);

  // Navigation handlers
  const handleSelectView = (view: 'overview' | 'sessions' | 'mcp') => {
    setViewedSession(null);
    if (view === 'overview') {
      setSelectedProfile('all');
      setCurrentView('overview');
    } else if (view === 'sessions') {
      setSelectedProfile('all');
      setCurrentView('sessions');
    } else if (view === 'mcp') {
      setSelectedMcpScope('all');
      setCurrentView('mcp');
    }
  };

  const handleSelectProfile = (profileName: string) => {
    setViewedSession(null);
    if (profileName === 'all') {
      setSelectedProfile('all');
      setCurrentView('overview');
    } else {
      setSelectedProfile(profileName);
      setCurrentView('profile-detail');
    }
  };

  const handleNavigateToSessions = (profileName: string) => {
    setViewedSession(null);
    setSelectedProfile(profileName);
    setCurrentView('sessions');
  };

  const handleNavigateToMcp = (profileName: string) => {
    setViewedSession(null);
    setSelectedMcpScope(profileName);
    setCurrentView('mcp');
  };

  const handleViewSession = (session: SessionDTO) => {
    setViewedSession(session);
    setCurrentView('session-detail');
  };

  const handleBackFromSession = () => {
    setViewedSession(null);
    setCurrentView('sessions');
  };

  const getHeaderTitle = () => {
    switch (currentView) {
      case 'overview':
        return (
          <span className="font-semibold text-sm text-[#ededed]">
            All Agent Profiles
          </span>
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
              {viewedSession ? viewedSession.id.slice(0, 8) : 'Detail'}
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
    <div className="flex h-screen w-full bg-[#000000] text-[#ededed] overflow-hidden font-sans">
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
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden bg-[#000000]">
        {/* Top Header Bar */}
        <header className="h-14 border-b border-[#262626] bg-[#0a0a0a]/90 backdrop-blur-md px-6 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3">
            {getHeaderTitle()}
          </div>

          {/* Right Header Status Pills */}
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              onClick={() => setIsDaemonOpen(true)}
              title="Manage OS Background Daemon (15m Quota Pre-Warm)"
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-[#262626] bg-[#111111] hover:bg-[#161616] hover:border-[#333333] transition-colors cursor-pointer text-xs group"
            >
              <span className="flex h-2 w-2 relative items-center justify-center">
                {daemonInfo?.active ? (
                  <span className="h-2 w-2 rounded-full bg-status-success-solid" />
                ) : daemonInfo?.installed ? (
                  <span className="h-2 w-2 rounded-full bg-status-warning-solid" />
                ) : (
                  <span className="h-2 w-2 rounded-full bg-[#666666]" />
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

            <div className="hidden sm:flex items-center gap-2 px-2.5 py-1 rounded-full border border-border bg-[#111111] text-xs">
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
              <span className="font-mono text-[11px] text-[#ededed]">
                {isOnline ? 'Server Connected' : 'Connecting...'}
              </span>
            </div>
          </div>
        </header>

        {/* Stage Content */}
        <main className="flex-1 overflow-y-auto px-6 py-6 min-h-0 bg-[#000000]">
          <div className="max-w-7xl mx-auto">
            {currentView === 'overview' && (
              <ProfilesView
                selectedProfile="all"
                onSelectProfile={handleSelectProfile}
                onNavigateToSessions={handleNavigateToSessions}
                onNavigateToMcp={handleNavigateToMcp}
                onProfilesLoaded={setProfiles}
              />
            )}
            {currentView === 'profile-detail' && (
              <ProfilesView
                selectedProfile={selectedProfile}
                onSelectProfile={handleSelectProfile}
                onNavigateToSessions={handleNavigateToSessions}
                onNavigateToMcp={handleNavigateToMcp}
                onProfilesLoaded={setProfiles}
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
            {currentView === 'session-detail' && viewedSession && (
              <SessionDetailView
                key={viewedSession.id}
                session={viewedSession}
                onBack={handleBackFromSession}
              />
            )}
            {currentView === 'mcp' && (
              <McpView
                selectedScope={selectedMcpScope}
                onSelectScope={setSelectedMcpScope}
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
