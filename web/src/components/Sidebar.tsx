import * as React from 'react';
import {
  Layers,
  Terminal,
  Server,
  Plus,
  Search,
  Cpu,
} from 'lucide-react';
import { ProfileDTO, DaemonDTO } from '@/lib/api';

interface SidebarProps {
  currentView: 'overview' | 'sessions' | 'mcp' | 'profile-detail' | 'session-detail';
  selectedProfile: string;
  profiles: ProfileDTO[];
  daemonInfo: DaemonDTO | null;
  version?: string;
  onSelectView: (view: 'overview' | 'sessions' | 'mcp') => void;
  onSelectProfile: (profileName: string) => void;
  onOpenAddProfile: () => void;
  onOpenDaemonModal: () => void;
}

export function Sidebar({
  currentView,
  selectedProfile,
  profiles,
  daemonInfo,
  version = 'v0.12.0',
  onSelectView,
  onSelectProfile,
  onOpenAddProfile,
  onOpenDaemonModal,
}: SidebarProps) {
  const [findQuery, setFindQuery] = React.useState('');
  const findInputRef = React.useRef<HTMLInputElement>(null);

  // Global hotkey 'f' focuses quick find
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (document.querySelector('[role="dialog"]')) return;
      if (
        (e.key === 'f' || e.key === 'F') &&
        document.activeElement?.tagName !== 'INPUT' &&
        document.activeElement?.tagName !== 'TEXTAREA'
      ) {
        e.preventDefault();
        findInputRef.current?.focus();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const getProfileQuotaStatus = (p: ProfileDTO) => {
    if (!p.has_credentials) {
      return {
        color: 'bg-status-warning-solid',
        text: 'text-status-warning',
        label: 'Auth',
      };
    }
    const pct = p.quota?.bottleneck_pct ?? 100;
    if (p.quota?.is_exhausted || pct <= 0) {
      return {
        color: 'bg-status-danger-solid',
        text: 'text-status-danger',
        label: '0%',
      };
    }
    if (pct <= 30) {
      return {
        color: 'bg-status-warning-solid',
        text: 'text-status-warning',
        label: `${pct}%`,
      };
    }
    return {
      color: 'bg-status-success-solid',
      text: 'text-status-success',
      label: `${pct}%`,
    };
  };

  const filteredProfiles = React.useMemo(() => {
    if (!findQuery.trim()) return profiles;
    const q = findQuery.toLowerCase().trim();
    return profiles.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.agent.toLowerCase().includes(q) ||
        (p.adapters && p.adapters.some((ad) => ad.agent.toLowerCase().includes(q)))
    );
  }, [profiles, findQuery]);

  const isProfilesActive = currentView === 'overview' || currentView === 'profile-detail';
  const isSessionsActive = currentView === 'sessions' || currentView === 'session-detail';
  const isMcpActive = currentView === 'mcp';

  return (
    <aside className="w-64 border-r border-border bg-background flex flex-col h-full shrink-0 select-none">
      {/* AIM Brand Header */}
      <div className="h-14 px-4 border-b border-border flex items-center justify-between">
        <button
          type="button"
          onClick={() => onSelectView('overview')}
          className="flex items-center gap-2.5 text-left hover:opacity-85 transition-opacity cursor-pointer group"
          title="AIM Home"
        >
          <div className="h-7 w-7 rounded-[6px] bg-card border border-border flex items-center justify-center shadow-xs">
            <Cpu className="h-3.5 w-3.5 text-foreground" />
          </div>
          <div className="flex items-center gap-1.5">
            <span className="font-semibold text-sm tracking-tight text-foreground">AIM</span>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-border bg-card text-muted-foreground">
              {version}
            </span>
          </div>
        </button>
      </div>

      {/* Quick Find Row with <kbd>F</kbd> */}
      <div className="px-3 pt-3 pb-1">
        <div className="relative flex items-center">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/60 pointer-events-none" />
          <input
            ref={findInputRef}
            type="text"
            value={findQuery}
            onChange={(e) => setFindQuery(e.target.value)}
            placeholder="Find"
            className="w-full h-7 pl-8 pr-7 text-xs font-sans bg-[#0c0c0c] border border-border/80 focus:border-border text-foreground placeholder:text-muted-foreground/50 rounded-geist focus:outline-none focus:ring-1 focus:ring-ring transition-colors"
          />
          <kbd className="absolute right-2 top-1/2 -translate-y-1/2 border border-border bg-muted/60 px-1 py-[1px] rounded text-[10px] font-mono text-muted-foreground pointer-events-none">
            F
          </kbd>
        </div>
      </div>

      {/* Navigation Sections */}
      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-5">
        {/* Section 1: Navigation Views */}
        <div className="space-y-0.5">
          <button
            type="button"
            onClick={() => onSelectView('overview')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              isProfilesActive
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Layers className={`h-3.5 w-3.5 ${isProfilesActive ? 'text-foreground' : 'text-muted-foreground'}`} />
            <span>Profiles</span>
            <span className="ml-auto text-[10px] font-mono text-muted-foreground/70">
              {profiles.length}
            </span>
          </button>

          <button
            type="button"
            onClick={() => onSelectView('sessions')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              isSessionsActive
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Terminal className={`h-3.5 w-3.5 ${isSessionsActive ? 'text-foreground' : 'text-muted-foreground'}`} />
            <span>Sessions</span>
          </button>

          <button
            type="button"
            onClick={() => onSelectView('mcp')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              isMcpActive
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Server className={`h-3.5 w-3.5 ${isMcpActive ? 'text-foreground' : 'text-muted-foreground'}`} />
            <span>MCP Catalog</span>
          </button>
        </div>

        {/* Section 2: Active Profiles List */}
        <div className="space-y-1">
          <div className="flex items-center justify-between px-2 pb-1">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground/70">
              Active Profiles ({filteredProfiles.length})
            </span>
            <button
              type="button"
              onClick={onOpenAddProfile}
              title="Add new profile"
              aria-label="Add new profile"
              className="h-5 w-5 rounded-[4px] flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer"
            >
              <Plus className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="space-y-0.5">
            {filteredProfiles.length === 0 ? (
              <div className="px-2.5 py-2 text-[11px] font-sans text-muted-foreground/60 italic">
                No matching profiles
              </div>
            ) : (
              filteredProfiles.map((p) => {
                const status = getProfileQuotaStatus(p);
                const isSelected = currentView === 'profile-detail' && selectedProfile === p.name;

                return (
                  <button
                    key={p.name}
                    type="button"
                    onClick={() => onSelectProfile(p.name)}
                    className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left group ${
                      isSelected
                        ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                        : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
                    }`}
                  >
                    <div className="flex items-center gap-2 truncate">
                      <span className={`h-1.5 w-1.5 rounded-full shrink-0 ${status.color}`} />
                      <span className="truncate">{p.name}</span>
                    </div>
                    <span className={`text-[10px] font-mono tabular-nums shrink-0 ${status.text}`}>
                      {status.label}
                    </span>
                  </button>
                );
              })
            )}
          </div>
        </div>
      </div>

      {/* Footer System Status */}
      <div className="p-3 border-t border-border bg-muted/20">
        <button
          type="button"
          onClick={onOpenDaemonModal}
          className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-geist border border-border bg-card hover:bg-secondary hover:border-border/80 transition-colors cursor-pointer text-xs group"
          title="Manage OS Background Daemon"
        >
          <div className="flex items-center gap-2">
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
              Daemon: {daemonInfo?.active ? 'Active' : daemonInfo?.installed ? 'Inactive' : 'Off'}
            </span>
          </div>
          <Cpu className="h-3.5 w-3.5 text-muted-foreground/60 group-hover:text-foreground transition-colors" />
        </button>
      </div>
    </aside>
  );
}
