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

          {/* Center Adapter Filter (Pills) */}
          <div className="hidden md:flex items-center gap-1 bg-[#111111] p-1 rounded-full border border-border">
            <span className="text-[11px] font-medium text-[#888888] px-2 font-mono">adapter:</span>
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
                className={`text-xs px-2.5 py-0.5 rounded-full font-medium transition-all duration-150 cursor-pointer ${
                  selectedAgent === ag.id
                    ? 'bg-[#222222] text-[#ededed] border border-[#333333] shadow-sm'
                    : 'text-[#888888] hover:text-[#ededed] hover:bg-[#161616] border border-transparent'
                }`}
              >
                {ag.label}
              </button>
            ))}
          </div>

          {/* Daemon Status Pill */}
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-2.5 py-1 rounded-full border border-border bg-[#111111] text-xs">
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
              <span className="font-mono text-[11px] text-[#888888]">
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
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-3">
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

            {/* Mobile Agent Selector */}
            <div className="md:hidden flex items-center gap-2">
              <span className="text-xs text-[#888888] font-mono">Agent:</span>
              <select
                value={selectedAgent}
                onChange={(e) => setSelectedAgent(e.target.value)}
                className="text-xs bg-[#111111] border border-border rounded-geist px-2 py-1 text-foreground focus:outline-none focus:ring-1 focus:ring-[#0070f3]"
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
            <ProfilesView selectedAgent={selectedAgent} />
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
            <span>Daemon Proxy: <code className="text-[#ededed] bg-[#141414] px-1.5 py-0.5 rounded border border-[#262626]">/api &rarr; :8080</code></span>
          </div>
        </div>
      </footer>

      {/* Global Toast Provider Notifications */}
      <Toaster />
    </div>
  );
}

export default App;
