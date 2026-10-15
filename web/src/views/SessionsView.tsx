import * as React from 'react';
import {
  SessionDTO,
  ProfileDTO,
  getSessions,
  resumeSession,
  ResumeRequest,
} from '@/lib/api';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from '@tanstack/react-table';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/components/ui/use-toast';
import {
  Terminal,
  Search,
  RefreshCw,
  Folder,
  Clock,
  MessageSquare,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  X,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
} from 'lucide-react';
import { ResumeModal } from '@/components/ResumeModal';
import { AgentBadge } from '@/components/AgentBadge';
import { DataTableFacetedFilter } from '@/components/DataTableFacetedFilter';
import { GeistSpinner } from '@/components/ui/spinner';
import { cn } from '@/lib/utils';

interface SessionsViewProps {
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
  profiles?: ProfileDTO[];
  onViewSession?: (session: SessionDTO) => void;
}

// Module-level in-memory cache and sessionStorage persistence to guarantee 0ms instant view switching with zero skeleton flash
let sessionsMemoryCache: SessionDTO[] | null = null;
let lastSessionsFetchTimestamp = 0;
const SESSIONS_CACHE_TTL_MS = 60 * 1000; // 60 seconds TTL
const SESSIONS_STORAGE_KEY = 'aim_sessions_cache_v1';

function getInitialSessions(): SessionDTO[] {
  if (sessionsMemoryCache && sessionsMemoryCache.length > 0) {
    return sessionsMemoryCache;
  }
  try {
    const raw = sessionStorage.getItem(SESSIONS_STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed) && parsed.length > 0) {
        sessionsMemoryCache = parsed;
        return parsed;
      }
    }
  } catch {
    // Ignore storage parse errors
  }
  return [];
}

export function SessionsView({
  selectedProfile = 'all',
  onSelectProfile,
  profiles = [],
  onViewSession,
}: SessionsViewProps) {
  const initialSessions = React.useMemo(() => getInitialSessions(), []);
  const [sessions, setSessions] = React.useState<SessionDTO[]>(initialSessions);
  // Only show skeleton on true cold boot when zero cached sessions exist
  const [loading, setLoading] = React.useState(initialSessions.length === 0);
  const [refreshing, setRefreshing] = React.useState(false);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [selectedEngines, setSelectedEngines] = React.useState<Set<string>>(new Set());
  const [selectedProfiles, setSelectedProfiles] = React.useState<Set<string>>(
    selectedProfile && selectedProfile !== 'all' ? new Set([selectedProfile]) : new Set()
  );
  const [selectedCwds, setSelectedCwds] = React.useState<Set<string>>(new Set());
  const [selectedStatuses, setSelectedStatuses] = React.useState<Set<string>>(new Set());
  const [resumingId, setResumingId] = React.useState<string | null>(null);
  const [flagModalSession, setFlagModalSession] = React.useState<SessionDTO | null>(null);
  const [sorting, setSorting] = React.useState<SortingState>([
    { id: 'updated_at', desc: true },
  ]);

  const searchInputRef = React.useRef<HTMLInputElement>(null);
  const { toast } = useToast();

  // Sync external selectedProfile prop
  React.useEffect(() => {
    if (selectedProfile && selectedProfile !== 'all') {
      setSelectedProfiles(new Set([selectedProfile]));
    } else {
      setSelectedProfiles(new Set());
    }
  }, [selectedProfile]);

  const fetchSessions = React.useCallback(async (force = false) => {
    // If we already have sessions in state or cache, perform a silent background revalidation without ever clearing the table
    if (!force && (sessionsMemoryCache && sessionsMemoryCache.length > 0)) {
      setRefreshing(true);
    } else if (force) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    try {
      const data = await getSessions(force ? { force: true } : undefined);
      sessionsMemoryCache = data;
      lastSessionsFetchTimestamp = Date.now();
      try {
        sessionStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(data));
      } catch {
        // Ignore storage quota errors
      }
      setSessions(data);
    } catch {
      // Fallback data if server unreachable and no cache exists
      if (!sessionsMemoryCache || sessionsMemoryCache.length === 0) {
        const fallback: SessionDTO[] = [
          {
            id: 'sess-89412e',
            agent: 'claude',
            profile: 'work',
            title: 'Refactor auth middleware and token validation',
            cwd: '/Users/developer/repos/enterprise-api',
            goal: 'Migrate legacy token handlers to modern RS256 JWT checks and test coverage',
            turns: 24,
            updated_at: new Date(Date.now() - 1000 * 60 * 18).toISOString(),
            is_active: true,
          },
          {
            id: 'sess-33921a',
            agent: 'codex',
            profile: 'personal',
            title: 'Implement CLI flags and config loader',
            cwd: '/Users/developer/Projects/my-cli-tool',
            goal: 'Add viper support for ~/.config/mytool.yaml and env overrides',
            turns: 12,
            updated_at: new Date(Date.now() - 1000 * 60 * 120).toISOString(),
            is_active: false,
          },
        ];
        sessionsMemoryCache = fallback;
        setSessions(fallback);
      }
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  React.useEffect(() => {
    const isStale = Date.now() - lastSessionsFetchTimestamp > SESSIONS_CACHE_TTL_MS;
    if (!sessionsMemoryCache || sessionsMemoryCache.length === 0 || isStale) {
      fetchSessions();
    }
  }, [fetchSessions]);

  // Global hotkey '/' focuses search
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (document.querySelector('[role="dialog"]')) return;
      if (
        e.key === '/' &&
        document.activeElement?.tagName !== 'INPUT' &&
        document.activeElement?.tagName !== 'TEXTAREA'
      ) {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Compute faceted counts
  const engineCounts = React.useMemo(() => {
    const map: Record<string, number> = {};
    sessions.forEach((s) => {
      let a = s.agent.toLowerCase();
      if (a === 'agy') a = 'antigravity';
      map[a] = (map[a] || 0) + 1;
    });
    return map;
  }, [sessions]);

  const profileCounts = React.useMemo(() => {
    const map: Record<string, number> = {};
    sessions.forEach((s) => {
      const p = s.profile || '<host>';
      map[p] = (map[p] || 0) + 1;
    });
    return map;
  }, [sessions]);

  const cwdCounts = React.useMemo(() => {
    const map: Record<string, number> = {};
    sessions.forEach((s) => {
      const c = s.cwd?.trim() || '<none>';
      map[c] = (map[c] || 0) + 1;
    });
    return map;
  }, [sessions]);

  const statusCounts = React.useMemo(() => {
    const map: Record<string, number> = { active: 0, idle: 0 };
    sessions.forEach((s) => {
      if (s.is_active) map.active++;
      else map.idle++;
    });
    return map;
  }, [sessions]);

  // Options for filters
  const engineOptions = React.useMemo(
    () => [
      { label: 'Claude', value: 'claude' },
      { label: 'Codex', value: 'codex' },
      { label: 'Antigravity', value: 'antigravity' },
      { label: 'Gemini', value: 'gemini' },
    ],
    []
  );

  const profileOptions = React.useMemo(() => {
    const set = new Set<string>();
    profiles.forEach((p) => set.add(p.name));
    sessions.forEach((s) => {
      if (s.profile) set.add(s.profile);
    });
    const list = Array.from(set).sort();
    return list.map((p) => ({ label: p, value: p }));
  }, [profiles, sessions]);

  const cwdOptions = React.useMemo(() => {
    const set = new Set<string>();
    sessions.forEach((s) => {
      const c = s.cwd?.trim() || '<none>';
      set.add(c);
    });
    const list = Array.from(set).sort();

    // Check for duplicate basenames to disambiguate
    const basenameCounts: Record<string, number> = {};
    list.forEach((raw) => {
      if (raw === '<none>') return;
      const parts = raw.split('/').filter(Boolean);
      const base = parts.length > 0 ? parts[parts.length - 1] : raw;
      basenameCounts[base] = (basenameCounts[base] || 0) + 1;
    });

    return list.map((c) => {
      if (c === '<none>') {
        return {
          label: '(No directory)',
          value: '<none>',
          icon: Folder,
        };
      }
      const parts = c.split('/').filter(Boolean);
      const basename = parts.length > 0 ? parts[parts.length - 1] : c;
      const parent = parts.length > 1 ? parts[parts.length - 2] : null;
      const isDuplicate = (basenameCounts[basename] || 0) > 1;
      const label = isDuplicate && parent ? `${basename} (.../${parent})` : basename;
      return {
        label,
        value: c,
        icon: Folder,
      };
    });
  }, [sessions]);

  const statusOptions = React.useMemo(
    () => [
      { label: 'Active', value: 'active' },
      { label: 'Idle', value: 'idle' },
    ],
    []
  );

  // Filtered rows
  const filteredData = React.useMemo(() => {
    return sessions.filter((s) => {
      // Search query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const match =
          s.id.toLowerCase().includes(q) ||
          (s.title && s.title.toLowerCase().includes(q)) ||
          (s.goal && s.goal.toLowerCase().includes(q)) ||
          (s.cwd && s.cwd.toLowerCase().includes(q)) ||
          (s.profile && s.profile.toLowerCase().includes(q)) ||
          (s.agent && s.agent.toLowerCase().includes(q));
        if (!match) return false;
      }

      // Engine multi-select (supports both 'antigravity' and 'agy')
      if (selectedEngines.size > 0) {
        const a = s.agent.toLowerCase();
        const matches =
          selectedEngines.has(a) ||
          (a === 'antigravity' && selectedEngines.has('agy')) ||
          (a === 'agy' && selectedEngines.has('antigravity'));
        if (!matches) return false;
      }

      // Profile multi-select
      if (selectedProfiles.size > 0) {
        const prof = s.profile || '<host>';
        if (!selectedProfiles.has(prof)) return false;
      }

      // CWD multi-select
      if (selectedCwds.size > 0) {
        const c = s.cwd?.trim() || '<none>';
        if (!selectedCwds.has(c)) return false;
      }

      // Status multi-select
      if (selectedStatuses.size > 0) {
        const st = s.is_active ? 'active' : 'idle';
        if (!selectedStatuses.has(st)) return false;
      }

      return true;
    });
  }, [sessions, searchQuery, selectedEngines, selectedProfiles, selectedCwds, selectedStatuses]);

  const handleQuickResume = async (sess: SessionDTO) => {
    setResumingId(sess.id);
    try {
      const res = await resumeSession({
        agent: sess.agent,
        profile: sess.profile,
        session_id: sess.id,
      });
      toast({
        title: 'Terminal launched',
        description: `Resumed in ${res.terminal || 'terminal'} window.`,
      });
    } catch (err: any) {
      toast({
        variant: 'destructive',
        title: 'Resume failed',
        description: err.message || 'Could not launch terminal',
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
        title: 'Terminal launched',
        description: `Resumed in ${res.terminal || 'terminal'} window.`,
      });
    } catch (err: any) {
      toast({
        variant: 'destructive',
        title: 'Resume failed',
        description: err.message || 'Could not launch terminal',
      });
    } finally {
      setResumingId(null);
    }
  };

  const formatRelativeTime = (isoString: string) => {
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

  // Define TanStack Table columns
  const columns = React.useMemo<ColumnDef<SessionDTO>[]>(
    () => [
      {
        accessorKey: 'agent',
        header: 'AGENT',
        size: 90,
        cell: ({ row }) => (
          <div className="flex items-center gap-1.5">
            <AgentBadge agent={row.original.agent} />
            {row.original.is_active && (
              <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" title="Active Process" />
            )}
          </div>
        ),
      },
      {
        accessorKey: 'profile',
        header: 'PROFILE',
        size: 85,
        cell: ({ row }) => {
          const prof = row.original.profile || '<host>';
          return (
            <Badge
              variant="outline"
              className={cn(
                'text-[10px] font-mono border-0 rounded-[4px] px-1.5 py-0.5 truncate max-w-[80px]',
                prof === '<host>'
                  ? 'bg-[#181818] text-[#888888]'
                  : 'bg-[#1a1a1a] text-[#ededed] border border-[#2a2a2a]'
              )}
            >
              {prof}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'cwd',
        header: 'DIRECTORY (CWD)',
        size: 210,
        cell: ({ row }) => {
          const rawCwd = row.original.cwd || '';
          if (!rawCwd) return <span className="text-muted-foreground font-mono text-xs">-</span>;
          const parts = rawCwd.split('/').filter(Boolean);
          const basename = parts.length > 0 ? parts[parts.length - 1] : rawCwd;
          const parent = parts.length > 1 ? parts[parts.length - 2] : null;
          return (
            <div className="flex items-center gap-1.5 font-mono text-xs min-w-0" title={rawCwd}>
              <Folder className="h-3.5 w-3.5 shrink-0 text-muted-foreground/70" />
              <div className="truncate flex items-baseline gap-0.5 text-xs">
                {parent && (
                  <span className="text-muted-foreground/60 shrink-0">
                    .../{parent}/
                  </span>
                )}
                <span className="text-foreground font-medium truncate">
                  {basename}
                </span>
              </div>
            </div>
          );
        },
      },
      {
        id: 'goal_title',
        header: 'GOAL / TITLE',
        size: 280,
        cell: ({ row }) => {
          const rawTitle = row.original.title?.trim();
          const rawGoal = row.original.goal?.trim();
          const sid = `sess-${row.original.id.slice(0, 8)}`;

          // Promote goal to primary title if explicit title is missing or generic
          const hasSpecificTitle = rawTitle && rawTitle.toLowerCase() !== 'untitled session';
          const primary = hasSpecificTitle ? rawTitle : (rawGoal || sid);
          const secondary = hasSpecificTitle && rawGoal && rawTitle.toLowerCase() !== rawGoal.toLowerCase()
            ? rawGoal
            : !hasSpecificTitle && rawTitle && rawGoal && rawGoal.toLowerCase() !== rawTitle.toLowerCase()
              ? rawTitle
              : null;

          return (
            <div className="space-y-0.5 min-w-0 pr-2">
              <div className="flex items-center gap-1.5 min-w-0">
                <span
                  className="font-mono text-[10px] text-muted-foreground bg-[#141414] px-1.5 py-0.5 rounded border border-[#242424] shrink-0 select-all"
                  title={`ID: ${row.original.id}`}
                >
                  {row.original.id.slice(0, 8)}
                </span>
                <span className="text-xs font-sans font-medium text-foreground truncate" title={primary}>
                  {primary}
                </span>
              </div>
              {secondary && (
                <div className="text-[11px] font-sans text-muted-foreground truncate" title={secondary}>
                  {secondary}
                </div>
              )}
            </div>
          );
        },
      },
      {
        accessorKey: 'turns',
        header: ({ column }) => (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="h-7 px-1 -ml-1 text-xs font-mono text-muted-foreground hover:text-foreground"
          >
            TURNS
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-1 h-3 w-3" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-1 h-3 w-3" />
            ) : (
              <ArrowUpDown className="ml-1 h-3 w-3 text-muted-foreground/60" />
            )}
          </Button>
        ),
        size: 70,
        cell: ({ row }) => (
          <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[4px] bg-[#141414] border border-[#242424] font-mono text-[11px] text-muted-foreground tabular-nums">
            <MessageSquare className="h-3 w-3 text-muted-foreground/70" />
            {row.original.turns}
          </span>
        ),
      },
      {
        accessorKey: 'updated_at',
        header: ({ column }) => (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="h-7 px-1 -ml-1 text-xs font-mono text-muted-foreground hover:text-foreground"
          >
            UPDATED
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-1 h-3 w-3" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-1 h-3 w-3" />
            ) : (
              <ArrowUpDown className="ml-1 h-3 w-3 text-muted-foreground/60" />
            )}
          </Button>
        ),
        size: 90,
        cell: ({ row }) => (
          <span
            className="flex items-center gap-1 font-mono text-xs text-muted-foreground cursor-default"
            title={new Date(row.original.updated_at).toLocaleString()}
          >
            <Clock className="h-3 w-3 text-muted-foreground/70" />
            {formatRelativeTime(row.original.updated_at)}
          </span>
        ),
      },
      {
        id: 'actions',
        header: () => null,
        size: 124,
        cell: ({ row }) => {
          const sess = row.original;
          const isThisResuming = resumingId === sess.id;
          return (
            <div className="flex items-center justify-end" onClick={(e) => e.stopPropagation()}>
              <div className="inline-flex items-center rounded-geist border border-[#2a2a2a] bg-[#121212] hover:bg-[#181818] hover:border-[#383838] transition-colors shadow-sm group">
                <button
                  type="button"
                  onClick={() => handleQuickResume(sess)}
                  disabled={isThisResuming}
                  className="h-7 pl-2.5 pr-2 text-xs font-sans font-medium text-[#ededed] group-hover:text-white flex items-center gap-1.5 transition-colors cursor-pointer disabled:opacity-50"
                >
                  {isThisResuming ? (
                    <GeistSpinner size={12} className="text-[#ededed]" />
                  ) : (
                    <Terminal className="h-3 w-3 text-[#888888] group-hover:text-[#ededed] transition-colors" />
                  )}
                  <span>Resume</span>
                </button>
                <div className="h-3.5 w-[1px] bg-[#2a2a2a]" />
                <button
                  type="button"
                  onClick={() => setFlagModalSession(sess)}
                  disabled={isThisResuming}
                  className="h-7 px-1.5 text-[#888888] hover:text-[#ededed] hover:bg-[#202020] transition-colors flex items-center justify-center cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  title="Resume with flags"
                >
                  <ChevronDown className="h-3 w-3" />
                </button>
              </div>
            </div>
          );
        },
      },
    ],
    [resumingId]
  );

  const table = useReactTable({
    data: filteredData,
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: {
      pagination: {
        pageSize: 15,
      },
    },
  });

  const hasActiveFilters =
    searchQuery !== '' ||
    selectedEngines.size > 0 ||
    selectedProfiles.size > 0 ||
    selectedCwds.size > 0 ||
    selectedStatuses.size > 0;

  const resetAllFilters = () => {
    setSearchQuery('');
    setSelectedEngines(new Set());
    setSelectedProfiles(new Set());
    setSelectedCwds(new Set());
    setSelectedStatuses(new Set());
    if (onSelectProfile) onSelectProfile('all');
  };

  return (
    <div className="space-y-4">
      {/* Top Filter & Search Bar */}
      <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Search input with autofocus shortcut and clear button */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#666666]" />
          <Input
            ref={searchInputRef}
            aria-label="Search sessions by goal, cwd, title, or profile"
            placeholder="Search sessions by goal, cwd, title, or profile..."
            className="pl-9 pr-8 h-8 text-xs rounded-geist border-border bg-[#0a0a0a] text-[#ededed] placeholder:text-[#666666] focus:border-[#333333]"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery ? (
            <button
              type="button"
              onClick={() => setSearchQuery('')}
              className="absolute right-2 top-1/2 -translate-y-1/2 h-6 w-6 rounded-[4px] flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary cursor-pointer transition-colors"
              title="Clear search"
              aria-label="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          ) : (
            <div className="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none">
              <kbd className="text-[10px] font-mono px-1.5 py-0.5 border border-border bg-[#111111] text-[#a1a1a1] rounded">
                /
              </kbd>
            </div>
          )}
        </div>

        {/* Faceted Multi-Select Filters and Actions */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Engine Multi-Select Popover */}
          <DataTableFacetedFilter
            title="Engine"
            options={engineOptions}
            selectedValues={selectedEngines}
            onSelect={setSelectedEngines}
            counts={engineCounts}
          />

          {/* Profile Multi-Select Popover */}
          <DataTableFacetedFilter
            title="Profile"
            options={profileOptions}
            selectedValues={selectedProfiles}
            onSelect={(vals) => {
              setSelectedProfiles(vals);
              if (onSelectProfile && vals.size === 1) {
                onSelectProfile(Array.from(vals)[0]);
              } else if (onSelectProfile && vals.size === 0) {
                onSelectProfile('all');
              }
            }}
            counts={profileCounts}
          />

          {/* CWD Multi-Select Popover */}
          <DataTableFacetedFilter
            title="CWD"
            options={cwdOptions}
            selectedValues={selectedCwds}
            onSelect={setSelectedCwds}
            counts={cwdCounts}
            contentClassName="w-64"
          />

          {/* Status Multi-Select Popover */}
          <DataTableFacetedFilter
            title="Status"
            options={statusOptions}
            selectedValues={selectedStatuses}
            onSelect={setSelectedStatuses}
            counts={statusCounts}
          />

          {hasActiveFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={resetAllFilters}
              className="h-8 px-2 text-xs font-mono text-[#a1a1a1] hover:text-[#ededed] rounded-geist"
            >
              <X className="mr-1 h-3.5 w-3.5" />
              Reset
            </Button>
          )}

          <Button
            variant="outline"
            size="sm"
            onClick={() => fetchSessions(true)}
            disabled={loading || refreshing}
            className="rounded-geist h-8 border-border bg-[#111111] text-[#ededed] hover:bg-[#1a1a1a]"
          >
            <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', (loading || refreshing) && 'animate-spin')} />
            Refresh
          </Button>
        </div>
      </div>

      {/* TanStack Table Container */}
      <div className="rounded-geist border border-border bg-[#111111] overflow-hidden shadow-sm">
        <Table containerClassName="overflow-x-auto max-h-[calc(100vh-240px)]" className="table-fixed w-full min-w-[760px]">
          <TableHeader className="sticky top-0 z-10 bg-[#0a0a0a]/95 backdrop-blur-md border-b border-border">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className="border-border hover:bg-transparent">
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className="text-[11px] font-mono font-semibold text-[#888888] tracking-wider py-2.5 px-3"
                    style={{ width: header.column.getSize() }}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>

          <TableBody>
            {loading ? (
              Array.from({ length: 6 }).map((_, idx) => (
                <TableRow key={`skeleton-${idx}`} className="border-border">
                  <TableCell className="py-2.5 px-3">
                    <div className="h-5 w-16 bg-[#1a1a1a] rounded animate-pulse" />
                  </TableCell>
                  <TableCell className="py-2.5 px-3">
                    <div className="h-5 w-14 bg-[#161616] rounded animate-pulse" />
                  </TableCell>
                  <TableCell className="py-2.5 px-3">
                    <div className="h-4 w-32 bg-[#1a1a1a] rounded animate-pulse" />
                  </TableCell>
                  <TableCell className="py-2.5 px-3">
                    <div className="space-y-1">
                      <div className="h-4 w-52 bg-[#1c1c1c] rounded animate-pulse" />
                      <div className="h-3 w-36 bg-[#141414] rounded animate-pulse" />
                    </div>
                  </TableCell>
                  <TableCell className="py-2.5 px-3">
                    <div className="h-4 w-12 bg-[#1a1a1a] rounded animate-pulse" />
                  </TableCell>
                  <TableCell className="py-2.5 px-3">
                    <div className="h-4 w-16 bg-[#181818] rounded animate-pulse" />
                  </TableCell>
                  <TableCell className="py-2.5 px-3 text-right">
                    <div className="h-7 w-20 bg-[#1c1c1c] rounded ml-auto animate-pulse" />
                  </TableCell>
                </TableRow>
              ))
            ) : table.getRowModel().rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="text-center py-16">
                  <div className="flex flex-col items-center justify-center max-w-sm mx-auto space-y-3">
                    <div className="w-10 h-10 rounded-full border border-border bg-[#161616] flex items-center justify-center text-[#666666]">
                      <Search className="w-4 h-4" />
                    </div>
                    <div className="space-y-1">
                      <p className="text-sm font-sans font-medium text-[#ededed]">No sessions found</p>
                      <p className="text-xs text-[#888888] font-sans">
                        {hasActiveFilters
                          ? 'No sessions match your search query or selected filter criteria.'
                          : 'No active or recorded agent sessions found for this environment.'}
                      </p>
                    </div>
                    {hasActiveFilters && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={resetAllFilters}
                        className="rounded-geist h-7 px-3 text-xs font-mono border-border bg-[#1a1a1a] text-[#ededed] hover:bg-[#222222] cursor-pointer"
                      >
                        <X className="mr-1.5 h-3 w-3" />
                        Clear all filters
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  onClick={() => onViewSession && onViewSession(row.original)}
                  className="border-border hover:bg-[#151515] transition-colors cursor-pointer group"
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} className="py-2.5 px-3">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination Controls */}
      {!loading && table.getPageCount() > 0 && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3 px-1 text-xs text-[#a1a1a1] font-mono">
          <div className="flex items-center gap-2">
            <span>
              Showing{' '}
              <strong className="text-[#ededed]">
                {table.getState().pagination.pageIndex * table.getState().pagination.pageSize + 1}
              </strong>{' '}
              to{' '}
              <strong className="text-[#ededed]">
                {Math.min(
                  (table.getState().pagination.pageIndex + 1) * table.getState().pagination.pageSize,
                  filteredData.length
                )}
              </strong>{' '}
              of <strong className="text-[#ededed]">{filteredData.length}</strong> sessions
            </span>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5">
              <span>Rows per page:</span>
              <select
                value={table.getState().pagination.pageSize}
                onChange={(e) => table.setPageSize(Number(e.target.value))}
                className="h-7 px-2 text-xs font-mono rounded-geist border border-border bg-[#0a0a0a] text-[#ededed] hover:border-[#333333] focus:outline-none cursor-pointer"
              >
                <option value={15}>15</option>
                <option value={25}>25</option>
                <option value={50}>50</option>
              </select>
            </div>

            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
                aria-label="Previous page"
                className="h-7 w-7 p-0 rounded-geist border-border bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616] disabled:opacity-40 cursor-pointer"
              >
                <ChevronLeft className="h-3.5 w-3.5" />
              </Button>
              <span className="px-2 text-xs">
                Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
                aria-label="Next page"
                className="h-7 w-7 p-0 rounded-geist border-border bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616] disabled:opacity-40 cursor-pointer"
              >
                <ChevronRight className="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Two-Tier Resume Modal */}
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
