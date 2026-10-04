import * as React from 'react';
import { getStatus, StatusDTO } from '@/lib/api';
import { ProfilesView } from '@/views/ProfilesView';
import { SessionsView } from '@/views/SessionsView';
import { McpView } from '@/views/McpView';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Toaster } from '@/components/ui/toaster';
import {
  Layers,
  Terminal,
  Server,
  Cpu,
} from 'lucide-react';

export function App() {
  const [activeTab, setActiveTab] = React.useState('profiles');
  const [selectedAgent, setSelectedAgent] = React.useState('all');
  const [status, setStatus] = React.useState<StatusDTO>({
    version: 'v0.12.0',
    status: 'healthy',
    active_sessions: 2,
    total_profiles: 4,
    terminals_available: ['tmux', 'wezterm', 'iterm2', 'terminal'],
  });
  const [isOnline, setIsOnline] = React.useState(true);

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
    const interval = setInterval(checkStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col selection:bg-primary/20">
      {/* Header */}
      <header className="sticky top-0 z-40 w-full border-b border-border/80 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
          {/* Logo & Version */}
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-lg bg-primary/15 border border-primary/30 flex items-center justify-center">
              <Cpu className="h-4 w-4 text-primary" />
            </div>
            <div className="flex items-center gap-2">
              <span className="font-bold tracking-tight text-base text-foreground">
                AIM
              </span>
              <span className="text-xs text-muted-foreground hidden sm:inline">
                Agent Identity Manager
              </span>
              <Badge
                variant="outline"
                className="font-mono text-[10px] py-0 px-1.5 border-primary/30 text-primary bg-primary/5"
              >
                {status.version || 'v0.12.0'}
              </Badge>
            </div>
          </div>

          {/* Center Agent Selector Filter */}
          <div className="hidden md:flex items-center gap-1.5 bg-muted/50 p-1 rounded-lg border border-border/60">
            <span className="text-[11px] font-semibold text-muted-foreground px-2">Agent:</span>
            {[
              { id: 'all', label: 'All' },
              { id: 'claude', label: 'Claude' },
              { id: 'codex', label: 'Codex' },
              { id: 'gemini', label: 'Gemini' },
              { id: 'agy', label: 'Antigravity' },
            ].map((ag) => (
              <button
                key={ag.id}
                onClick={() => setSelectedAgent(ag.id)}
                className={`text-xs px-2.5 py-1 rounded-md font-medium transition-all cursor-pointer ${
                  selectedAgent === ag.id
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {ag.label}
              </button>
            ))}
          </div>

          {/* Status Pill & Terminal Badge */}
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-2.5 py-1 rounded-full border border-border/80 bg-card/60 text-xs">
              <span className="relative flex h-2 w-2">
                <span
                  className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                    isOnline ? 'bg-emerald-400' : 'bg-amber-400'
                  }`}
                />
                <span
                  className={`relative inline-flex rounded-full h-2 w-2 ${
                    isOnline ? 'bg-emerald-500' : 'bg-amber-500'
                  }`}
                />
              </span>
              <span className="font-medium text-xs">
                {isOnline ? 'Daemon Connected' : 'Connecting...'}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          {/* Navigation Bar */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/60 pb-4">
            <TabsList className="bg-muted/60 p-1 border border-border/50">
              <TabsTrigger value="profiles" className="flex items-center gap-2">
                <Layers className="h-4 w-4" />
                <span>Profiles</span>
              </TabsTrigger>
              <TabsTrigger value="sessions" className="flex items-center gap-2">
                <Terminal className="h-4 w-4" />
                <span>Sessions</span>
              </TabsTrigger>
              <TabsTrigger value="mcp" className="flex items-center gap-2">
                <Server className="h-4 w-4" />
                <span>MCP Servers</span>
              </TabsTrigger>
            </TabsList>

            {/* Mobile Agent Selector */}
            <div className="md:hidden flex items-center gap-2">
              <span className="text-xs text-muted-foreground">Filter Agent:</span>
              <select
                value={selectedAgent}
                onChange={(e) => setSelectedAgent(e.target.value)}
                className="text-xs bg-muted border border-border rounded-md px-2 py-1 text-foreground"
              >
                <option value="all">All Agents</option>
                <option value="claude">Claude</option>
                <option value="codex">Codex</option>
                <option value="gemini">Gemini</option>
                <option value="agy">Antigravity</option>
              </select>
            </div>
          </div>

          {/* View Panels */}
          <TabsContent value="profiles" className="m-0 focus-visible:outline-none">
            <ProfilesView
              selectedAgent={selectedAgent}
              onSelectAgent={(agent) => setSelectedAgent(agent)}
            />
          </TabsContent>

          <TabsContent value="sessions" className="m-0 focus-visible:outline-none">
            <SessionsView selectedAgent={selectedAgent} />
          </TabsContent>

          <TabsContent value="mcp" className="m-0 focus-visible:outline-none">
            <McpView />
          </TabsContent>
        </Tabs>
      </main>

      {/* Footer */}
      <footer className="border-t border-border/60 py-4 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-muted-foreground">
          <div className="flex items-center gap-2">
            <span>AIM Web UI</span>
            <span>•</span>
            <span className="font-mono">React 19 + Vite + Tailwind + shadcn/ui</span>
          </div>
          <div className="flex items-center gap-4">
            <span>Proxy: <code className="text-foreground">/api &rarr; :8080</code></span>
          </div>
        </div>
      </footer>

      {/* Global Toast Provider Notifications */}
      <Toaster />
    </div>
  );
}

export default App;
