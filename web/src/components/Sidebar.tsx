import * as React from 'react';
import {
  Layers,
  Terminal,
  Server,
  Plus,
  Search,
  ChevronsUpDown,
  Bell,
  Cpu,
} from 'lucide-react';
import { ProfileDTO, DaemonDTO } from '@/lib/api';

interface SidebarProps {
  currentView: 'overview' | 'sessions' | 'mcp' | 'profile-detail' | 'session-detail';
  selectedProfile: string;
  profiles: ProfileDTO[];
  daemonInfo: DaemonDTO | null;
  isOnline: boolean;
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
  isOnline,
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
      color: 'bg-status-meter',
      text: 'text-muted-foreground/60',
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

  return (
    <aside className="w-64 border-r border-border bg-background flex flex-col h-full shrink-0 select-none">
      {/* Vercel-style Top Workspace Switcher Header */}
      <div className="h-14 px-3.5 border-b border-border flex items-center justify-between gap-2">
        <button
          type="button"
          onClick={() => onSelectView('overview')}
          className="flex items-center gap-2.5 min-w-0 text-left hover:opacity-85 transition-opacity cursor-pointer group"
          title="Switch workspace / Home"
        >
          {/* Avatar circle */}
          <div className="h-7 w-7 rounded-full bg-gradient-to-tr from-[#7928ca] to-[#ff0080] flex items-center justify-center text-[10px] font-bold text-white shadow-sm shrink-0">
            A
          </div>
          <div className="flex items-center gap-1.5 truncate">
            <span className="font-semibold text-xs tracking-tight text-foreground truncate">
              AIM Workspace
            </span>
            <span className="text-[9px] font-mono uppercase tracking-wider px-1.5 py-0.5 rounded-full border border-border bg-muted text-muted-foreground shrink-0">
              Sandbox
            </span>
          </div>
        </button>
        <ChevronsUpDown className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
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
          <kbd className="absolute right-2 top-1/2 -translate-y-1/2 border border-border bg-muted/60 px-1 py-0.2 rounded text-[10px] font-mono text-muted-foreground pointer-events-none">
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
              currentView === 'overview' || currentView === 'profile-detail'
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Layers className="h-3.5 w-3.5 text-muted-foreground" />
            <span>Profiles</span>
            <span className="ml-auto text-[10px] font-mono text-muted-foreground/70">
              {profiles.length}
            </span>
          </button>

          <button
            type="button"
            onClick={() => onSelectView('sessions')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              currentView === 'sessions' || currentView === 'session-detail'
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Terminal className="h-3.5 w-3.5 text-muted-foreground" />
            <span>Sessions</span>
          </button>

          <button
            type="button"
            onClick={() => onSelectView('mcp')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-geist text-xs font-sans transition-colors cursor-pointer text-left ${
              currentView === 'mcp'
                ? 'bg-secondary text-foreground border border-border shadow-sm font-medium'
                : 'text-muted-foreground hover:text-foreground hover:bg-card border border-transparent'
            }`}
          >
            <Server className="h-3.5 w-3.5 text-muted-foreground" />
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

      {/* Vercel-style Bottom Callout Card: AIM Daemon */}
      <div className="px-3 pb-3 pt-1">
        <div className="rounded-geist border border-[#222222] bg-[#0c0c0c] p-3 text-xs space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <Cpu className="h-3.5 w-3.5 text-[#ededed]" />
              <span className="font-medium text-xs text-[#ededed]">AIM Daemon</span>
            </div>
            <span className="flex h-2 w-2 relative items-center justify-center">
              {daemonInfo?.active ? (
                <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
              ) : daemonInfo?.installed ? (
                <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
              ) : (
                <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
              )}
            </span>
          </div>

          <p className="text-[11px] text-[#888888] font-sans leading-relaxed">
            {daemonInfo?.active
              ? 'Live background telemetry & session indexer running via launchd.'
              : 'Background service ready for automated quota syncing.'}
          </p>

          <button
            type="button"
            onClick={onOpenDaemonModal}
            className="w-full h-7 rounded-geist bg-[#ffffff] hover:bg-[#eaeaea] text-[#000000] font-sans font-medium text-xs flex items-center justify-center transition-colors shadow-sm cursor-pointer"
          >
            Manage Daemon
          </button>
        </div>
      </div>

      {/* Vercel-style Bottom User & Live Status Bar */}
      <div className="px-3 py-2.5 border-t border-border flex items-center justify-between bg-muted/20 text-xs">
        <div className="flex items-center gap-2 min-w-0">
          <div className="h-6 w-6 rounded-full bg-[#1e1e1e] border border-[#2e2e2e] flex items-center justify-center text-[10px] font-medium text-[#ededed] shrink-0">
            D
          </div>
          <div className="truncate font-sans">
            <div className="text-[11px] font-medium text-foreground truncate leading-tight">
              Local Developer
            </div>
            <div className="text-[10px] font-mono text-muted-foreground/70 truncate">
              AIM {version}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-1.5 shrink-0" title={isOnline ? 'Backend server connected' : 'Backend reconnecting'}>
          <div className="relative flex items-center justify-center h-6 w-6 rounded-[4px] hover:bg-secondary text-muted-foreground cursor-default">
            <Bell className="h-3.5 w-3.5 text-muted-foreground" />
            <span
              className={`absolute top-1 right-1 h-1.5 w-1.5 rounded-full ${
                isOnline ? 'bg-status-success-solid' : 'bg-status-warning-solid'
              }`}
            />
          </div>
        </div>
      </div>
    </aside>
  );
}
