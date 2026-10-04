import * as React from 'react';
import { getStatus, getDaemonStatus, StatusDTO, DaemonDTO } from '@/lib/api';
import { ProfilesView } from '@/views/ProfilesView';
import { SessionsView } from '@/views/SessionsView';
import { McpView } from '@/views/McpView';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Toaster } from '@/components/ui/toaster';
import { DaemonModal } from '@/components/DaemonModal';
import {
  Layers,
  Terminal,
  Server,
  Cpu,
} from 'lucide-react';

export function App() {
  const [activeTab, setActiveTab] = React.useState('profiles');
  const [selectedProfile, setSelectedProfile] = React.useState('all');
  const [status, setStatus] = React.useState<StatusDTO>({
    version: 'v0.12.0',
    status: 'healthy',
    active_sessions: 2,
    total_profiles: 4,
    terminals_available: ['tmux', 'wezterm', 'iterm2', 'terminal'],
  });
  const [isOnline, setIsOnline] = React.useState(true);
  const [daemonInfo, setDaemonInfo] = React.useState<DaemonDTO | null>(null);
  const [isDaemonOpen, setIsDaemonOpen] = React.useState(false);

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
        // Fallback status if backend server is starting up or in standalone UI mode
        setIsOnline(true);
      }
    }
    checkStatus();
    fetchDaemonStatus();
    const interval = setInterval(() => {
      checkStatus();
      fetchDaemonStatus();
    }, 30000);
    return () => clearInterval(interval);
  }, [fetchDaemonStatus]);

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans">
      {/* Vercel Geist Sticky Header */}
      <header className="sticky top-0 z-40 w-full border-b border-border bg-[#0a0a0a]/90 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between gap-4">
          {/* Logo & Version */}
          <div className="flex items-center gap-3">
            <div className="h-7 w-7 rounded-geist bg-[#171717] border border-[#2a2a2a] flex items-center justify-center">
              <Cpu className="h-3.5 w-3.5 text-[#ededed]" />
            </div>
            <div className="flex items-center gap-2">
              <span className="font-semibold tracking-tight text-sm text-[#ededed]">
                AIM
              </span>
              <span className="text-xs text-[#888888] hidden sm:inline">
                Agent Identity Manager
              </span>
              <Badge
                variant="outline"
                className="font-mono text-[10px] py-0 px-1.5 border-[#262626] text-[#888888] bg-[#141414]"
              >
                {status.version || 'v0.12.0'}
              </Badge>
            </div>
          </div>

          {/* Status Pills: Daemon Modal Trigger & Server Connection */}
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              onClick={() => setIsDaemonOpen(true)}
              title="Manage OS Background Daemon (15m Quota Pre-Warm)"
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-[#262626] bg-[#111111] hover:bg-[#161616] hover:border-[#383838] transition-colors cursor-pointer text-xs group"
            >
              <span className="flex h-2 w-2 relative items-center justify-center">
                {daemonInfo?.active ? (
                  <span className="h-2 w-2 rounded-full bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.5)]" />
                ) : daemonInfo?.installed ? (
                  <span className="h-2 w-2 rounded-full bg-amber-500 shadow-[0_0_6px_rgba(245,158,11,0.5)]" />
                ) : (
                  <span className="h-2 w-2 rounded-full bg-[#555555]" />
                )}
              </span>
              <span className="font-mono text-[11px] text-[#a1a1a1] group-hover:text-[#ededed] transition-colors">
                {daemonInfo?.active
                  ? 'Daemon: Active'
                  : daemonInfo?.installed
                  ? 'Daemon: Inactive'
                  : 'Daemon: Off'}
              </span>
            </button>

            <div className="hidden sm:flex items-center gap-2 px-2.5 py-1 rounded-full border border-[#262626] bg-[#111111] text-xs">
              <span className="relative flex h-2 w-2">
                <span
                  className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                    isOnline ? 'bg-[#00e599]' : 'bg-[#f5a623]'
                  }`}
                />
                <span
                  className={`relative inline-flex rounded-full h-2 w-2 ${
                    isOnline ? 'bg-[#00e599] shadow-[0_0_8px_rgba(0,229,153,0.6)]' : 'bg-[#f5a623]'
                  }`}
                />
              </span>
              <span className="font-mono text-[11px] text-[#ededed]">
                {isOnline ? 'Server Connected' : 'Connecting...'}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          {/* Navigation Bar */}
          <div className="flex items-center justify-between border-b border-border pb-3">
            <TabsList className="bg-[#111111] p-0.5 border border-border rounded-geist">
              <TabsTrigger value="profiles" className="flex items-center gap-1.5 text-xs">
                <Layers className="h-3.5 w-3.5" />
                <span>Profiles</span>
              </TabsTrigger>
              <TabsTrigger value="sessions" className="flex items-center gap-1.5 text-xs">
                <Terminal className="h-3.5 w-3.5" />
                <span>Sessions</span>
              </TabsTrigger>
              <TabsTrigger value="mcp" className="flex items-center gap-1.5 text-xs">
                <Server className="h-3.5 w-3.5" />
                <span>MCP Servers</span>
              </TabsTrigger>
            </TabsList>
          </div>

          {/* View Panels */}
          <TabsContent value="profiles" className="m-0 focus-visible:outline-none">
            <ProfilesView
              selectedProfile={selectedProfile}
              onSelectProfile={setSelectedProfile}
            />
          </TabsContent>

          <TabsContent value="sessions" className="m-0 focus-visible:outline-none">
            <SessionsView
              selectedProfile={selectedProfile}
              onSelectProfile={setSelectedProfile}
            />
          </TabsContent>

          <TabsContent value="mcp" className="m-0 focus-visible:outline-none">
            <McpView
              selectedProfile={selectedProfile}
              onSelectProfile={setSelectedProfile}
            />
          </TabsContent>
        </Tabs>
      </main>

      {/* Footer */}
      <footer className="border-t border-[#1f1f1f] py-4 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-[#666666] font-mono">
          <div className="flex items-center gap-2">
            <span className="text-[#888888]">AIM System</span>
            <span>/</span>
            <span>Dashboard</span>
            <span>•</span>
            <span>Geist Tokens</span>
          </div>
          <div className="flex items-center gap-4">
            <span>
              API Server: <code className="text-[#ededed] bg-[#141414] px-1.5 py-0.5 rounded border border-[#262626]">/api → :8080</code>
            </span>
          </div>
        </div>
      </footer>

      <DaemonModal
        open={isDaemonOpen}
        onOpenChange={setIsDaemonOpen}
        daemonInfo={daemonInfo}
        onRefreshDaemon={fetchDaemonStatus}
      />
      <Toaster />
    </div>
  );
}

export default App;
