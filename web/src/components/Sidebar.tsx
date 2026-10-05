import {
  Layers,
  Terminal,
  Server,
  Plus,
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
      color: 'bg-[#2b3039]',
      text: 'text-muted-foreground/60',
      label: `${pct}%`,
    };
  };

  return (
    <aside className="w-64 border-r border-border bg-background flex flex-col h-full shrink-0 select-none">
      {/* Brand Header */}
      <div className="h-14 px-4 border-b border-border flex items-center justify-between gap-2">
        <div className="flex items-center gap-2.5">
          <div className="h-7 w-7 rounded-[6px] bg-card border border-border flex items-center justify-center">
            <Cpu className="h-3.5 w-3.5 text-foreground" />
          </div>
          <div className="flex items-center gap-1.5">
            <span className="font-semibold text-sm tracking-tight text-foreground">AIM</span>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-border bg-card text-muted-foreground">
              {version}
            </span>
          </div>
        </div>
      </div>

      {/* Navigation Sections */}
      <div className="flex-1 overflow-y-auto px-3 py-4 space-y-6">
        {/* Section 1: Workspaces */}
        <div className="space-y-1">
          <div className="px-2 pb-1.5 text-[10px] font-mono uppercase tracking-wider text-muted-foreground/70">
            Workspaces
          </div>

          <button
            type="button"
            onClick={() => onSelectView('overview')}
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-[6px] text-xs font-mono transition-colors cursor-pointer text-left ${
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
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-[6px] text-xs font-mono transition-colors cursor-pointer text-left ${
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
            className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-[6px] text-xs font-mono transition-colors cursor-pointer text-left ${
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
          <div className="flex items-center justify-between px-2 pb-1.5">
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground/70">
              Active Profiles ({profiles.length})
            </span>
            <button
              type="button"
              onClick={onOpenAddProfile}
              title="Add new profile"
              aria-label="Add new profile"
              className="h-6 w-6 rounded-[4px] flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors cursor-pointer"
            >
              <Plus className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="space-y-0.5">
            {profiles.map((p) => {
              const status = getProfileQuotaStatus(p);
              const isSelected = currentView === 'profile-detail' && selectedProfile === p.name;

              return (
                <button
                  key={p.name}
                  type="button"
                  onClick={() => onSelectProfile(p.name)}
                  className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-[6px] text-xs font-mono transition-colors cursor-pointer text-left group ${
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
            })}
          </div>
        </div>
      </div>

      {/* Footer System Status */}
      <div className="p-3 border-t border-border space-y-2 bg-muted/30">
        <button
          type="button"
          onClick={onOpenDaemonModal}
          className="w-full flex items-center justify-between px-2.5 py-1.5 rounded-[6px] border border-border bg-card hover:bg-secondary hover:border-border/80 transition-colors cursor-pointer text-xs group"
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
            <span className="font-mono text-[11px] text-muted-foreground group-hover:text-foreground">
              {daemonInfo?.active ? 'Daemon: Active' : daemonInfo?.installed ? 'Daemon: Inactive' : 'Daemon: Off'}
            </span>
          </div>
        </button>

        <div className="flex items-center justify-between px-2.5 py-1 rounded-[6px] border border-border/60 bg-muted/60 text-[11px] font-mono text-muted-foreground">
          <div className="flex items-center gap-2">
            <span className="relative flex h-1.5 w-1.5">
              <span
                className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${
                  isOnline ? 'bg-status-success-solid' : 'bg-status-warning-solid'
                }`}
              />
              <span
                className={`relative inline-flex rounded-full h-1.5 w-1.5 ${
                  isOnline ? 'bg-status-success-solid' : 'bg-status-warning-solid'
                }`}
              />
            </span>
            <span>{isOnline ? 'Server: 8080' : 'Reconnecting'}</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
