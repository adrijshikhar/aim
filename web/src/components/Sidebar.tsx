import * as React from 'react';
import {
  Layers,
  Terminal,
  Server,
  Plus,
  Search,
  Radio,
} from 'lucide-react';
import { AimLogo } from '@/components/AimLogo';
import { ProfileDTO, DaemonDTO } from '@/lib/api';
import { cn } from '@/lib/utils';

interface SidebarProps {
  currentView: 'overview' | 'sessions' | 'mcp' | 'network' | 'profile-detail' | 'session-detail';
  selectedProfile: string;
  profiles: ProfileDTO[];
  daemonInfo: DaemonDTO | null;
  version?: string;
  isOnline?: boolean;
  onSelectView: (view: 'overview' | 'sessions' | 'mcp' | 'network') => void;
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
  isOnline = true,
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
    if (p.quota?.is_exhausted || (p.quota?.bottleneck_pct != null && p.quota.bottleneck_pct <= 0)) {
      return {
        color: 'bg-status-danger-solid',
        text: 'text-status-danger',
        label: '0%',
      };
    }
    if (p.quota?.bottleneck_pct == null) {
      return {
        color: 'bg-muted-foreground/40',
        text: 'text-muted-foreground',
        label: '--%',
      };
    }
    const pct = p.quota.bottleneck_pct;
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

  const isOverviewActive = currentView === 'overview';
  const isSessionsActive = currentView === 'sessions' || currentView === 'session-detail';
  const isMcpActive = currentView === 'mcp';
  const isNetworkActive = currentView === 'network';

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
            <AimLogo className="h-4 w-4 text-foreground" />
          </div>
          <div className="flex items-center gap-1.5">
            <span className="font-semibold text-sm tracking-tight text-foreground">AIM</span>
            <span className="sr-only">Agent Identity Manager</span>
            <span className="text-xs font-mono px-1.5 py-0.5 rounded border border-border bg-card text-muted-foreground">
              {version}
            </span>
          </div>
        </button>
      </div>

      {/* Quick Find Row with <kbd>F</kbd> */}
      <div className="px-3 pt-3 pb-1">
        <div className="relative flex items-center">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/80 pointer-events-none" />
          <input
            ref={findInputRef}
            type="text"
            value={findQuery}
            onChange={(e) => setFindQuery(e.target.value)}
            placeholder="Find"
            aria-label="Filter navigation items and profiles"
            className="w-full h-7 pl-8 pr-7 text-xs font-sans bg-[#0c0c0c] border border-border/80 focus:border-border text-foreground placeholder:text-muted-foreground rounded-geist focus:outline-none focus:ring-1 focus:ring-ring transition-colors"
          />
          <kbd className="absolute right-2 top-1/2 -translate-y-1/2 border border-border bg-muted/60 px-1 py-[1px] rounded text-xs font-mono text-muted-foreground pointer-events-none">
            F
          </kbd>
        </div>
      </div>

      {/* Navigation Sections */}
      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-5">
        {/* Section 1: Navigation Views */}
        <div role="tablist" aria-label="Views" className="space-y-0.5">
          <button
            role="tab"
            aria-selected={isOverviewActive}
            type="button"
            onClick={() => onSelectView('overview')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              isOverviewActive
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Layers className={`h-3.5 w-3.5 ${isOverviewActive ? 'text-foreground' : 'text-muted-foreground'}`} />
            <span>Profiles</span>
            <span className="ml-auto text-xs font-mono text-muted-foreground/80">
              {profiles.length}
            </span>
          </button>

          <button
            role="tab"
            aria-selected={isSessionsActive}
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
            role="tab"
            aria-selected={isMcpActive}
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

          <button
            role="tab"
            aria-selected={isNetworkActive}
            type="button"
            onClick={() => onSelectView('network')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              isNetworkActive
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Radio className={`h-3.5 w-3.5 ${isNetworkActive ? 'text-accent-cyan' : 'text-muted-foreground'}`} />
            <span>P2P Network</span>
          </button>
        </div>

        {/* Section 2: Active Profiles List */}
        <div className="space-y-1">
          <div className="flex items-center justify-between px-2 pb-1">
            <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground/80">
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
              <div className="px-2.5 py-2 text-xs font-sans text-muted-foreground/80 italic">
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
                    <span className={`text-xs font-mono tabular-nums shrink-0 ${status.text}`}>
                      {status.label}
                    </span>
                  </button>
                );
              })
            )}
          </div>
        </div>
      </div>

      {/* Footer System Status & Developer Profile */}
      <div className="p-3 border-t border-border bg-muted/20">
        <button
          type="button"
          onClick={onOpenDaemonModal}
          className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-geist border border-border bg-card hover:bg-secondary hover:border-border/80 transition-colors cursor-pointer text-xs group"
          title="AIM Background Daemon status (click to manage)"
        >
          <div className="flex items-center gap-2">
            <Radio className="h-3.5 w-3.5 text-muted-foreground/60 group-hover:text-foreground transition-colors shrink-0" />
            <span className="font-mono text-xs text-muted-foreground group-hover:text-foreground transition-colors">
              Daemon: {daemonInfo?.active ? 'Active' : daemonInfo?.installed ? 'Inactive' : 'Off'}
            </span>
          </div>

          <div
            className="flex items-center gap-1.5 px-1.5 py-0.5 rounded-[4px] border border-border/60 bg-muted/40 text-xs font-mono text-muted-foreground shrink-0"
            title={isOnline ? 'Backend connected' : 'Backend offline'}
          >
            <span className="relative flex h-1.5 w-1.5 items-center justify-center">
              {isOnline && (
                <span className="motion-reduce:animate-none animate-ping absolute inline-flex h-full w-full rounded-full bg-status-success-solid opacity-75" />
              )}
              <span
                className={cn(
                  'relative inline-flex h-1.5 w-1.5 rounded-full',
                  isOnline
                    ? 'bg-status-success-solid shadow-[0_0_8px_#30a46c]'
                    : 'bg-status-danger-solid'
                )}
              />
            </span>
            <span>{isOnline ? 'live' : 'offline'}</span>
          </div>
        </button>
      </div>
    </aside>
  );
}
