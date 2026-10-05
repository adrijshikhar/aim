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
import { cn } from '@/lib/utils';

interface SessionsViewProps {
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
  profiles?: ProfileDTO[];
  onViewSession?: (session: SessionDTO) => void;
}

export function SessionsView({
  selectedProfile = 'all',
  onSelectProfile,
  profiles = [],
  onViewSession,
}: SessionsViewProps) {
  const [sessions, setSessions] = React.useState<SessionDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [selectedEngines, setSelectedEngines] = React.useState<Set<string>>(new Set());
  const [selectedProfiles, setSelectedProfiles] = React.useState<Set<string>>(
    selectedProfile && selectedProfile !== 'all' ? new Set([selectedProfile]) : new Set()
  );
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

  const fetchSessions = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessions();
      setSessions(data);
    } catch {
      // Fallback data if server unreachable
      setSessions([
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
      ]);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchSessions();
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

      // Status multi-select
      if (selectedStatuses.size > 0) {
        const st = s.is_active ? 'active' : 'idle';
        if (!selectedStatuses.has(st)) return false;
      }

      return true;
    });
  }, [sessions, searchQuery, selectedEngines, selectedProfiles, selectedStatuses]);

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
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" title="Active Process" />
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
        size: 160,
        cell: ({ row }) => {
          const rawCwd = row.original.cwd || '';
          const parts = rawCwd.split('/').filter(Boolean);
          const basename = parts.length > 0 ? parts[parts.length - 1] : rawCwd;
          const parent = parts.length > 1 ? `.../${parts[parts.length - 2]}/` : '';
          return (
            <div className="flex items-center gap-1.5 font-mono text-xs text-[#a1a1a1] max-w-[155px]" title={rawCwd}>
              <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
              <div className="truncate">
                <span className="text-[#666666]">{parent}</span>
                <span className="text-[#ededed] font-medium">{basename}</span>
              </div>
            </div>
          );
        },
      },
      {
        id: 'goal_title',
        header: 'GOAL / TITLE',
        cell: ({ row }) => (
          <div className="space-y-0.5 min-w-0 pr-2">
            <div className="text-xs font-sans font-medium text-[#ededed] truncate">
              {row.original.title || 'Untitled Session'}
            </div>
            {row.original.goal && (
              <div className="text-[11px] font-sans text-[#888888] truncate">{row.original.goal}</div>
            )}
          </div>
        ),
      },
      {
        accessorKey: 'turns',
        header: ({ column }) => (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            className="h-7 px-1 text-xs font-mono text-[#a1a1a1] hover:text-[#ededed]"
          >
            TURNS
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-1 h-3 w-3" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-1 h-3 w-3" />
            ) : (
              <ArrowUpDown className="ml-1 h-3 w-3 text-[#555555]" />
            )}
          </Button>
        ),
        size: 70,
        cell: ({ row }) => (
          <span className="flex items-center gap-1 font-mono text-xs text-[#a1a1a1]">
            <MessageSquare className="h-3 w-3 text-[#666666]" />
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
            className="h-7 px-1 text-xs font-mono text-[#a1a1a1] hover:text-[#ededed]"
          >
            UPDATED
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="ml-1 h-3 w-3" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="ml-1 h-3 w-3" />
            ) : (
              <ArrowUpDown className="ml-1 h-3 w-3 text-[#555555]" />
            )}
          </Button>
        ),
        size: 90,
        cell: ({ row }) => (
          <span className="flex items-center gap-1 font-mono text-xs text-[#a1a1a1]">
            <Clock className="h-3 w-3 text-[#666666]" />
            {formatRelativeTime(row.original.updated_at)}
          </span>
        ),
      },
      {
        id: 'actions',
        header: () => null,
        size: 110,
        cell: ({ row }) => {
          const sess = row.original;
          const isThisResuming = resumingId === sess.id;
          return (
            <div className="flex items-center justify-end" onClick={(e) => e.stopPropagation()}>
              <div className="inline-flex items-center rounded-geist border border-border bg-[#111111] overflow-hidden hover:border-[#383838] transition-colors">
                <button
                  type="button"
                  onClick={() => handleQuickResume(sess)}
                  disabled={isThisResuming}
                  className="h-7 pl-2.5 pr-2 text-xs font-sans font-medium text-[#ededed] hover:bg-[#1a1a1a] flex items-center gap-1.5 transition-colors cursor-pointer"
                >
                  {isThisResuming ? (
                    <RefreshCw className="h-3 w-3 animate-spin text-[#ededed]" />
                  ) : (
                    <Terminal className="h-3 w-3 text-[#ededed]" />
                  )}
                  <span>Resume</span>
                </button>
                <div className="h-3.5 w-[1px] bg-border" />
                <button
                  type="button"
                  onClick={() => setFlagModalSession(sess)}
                  className="h-7 px-1.5 text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1a1a1a] transition-colors flex items-center justify-center cursor-pointer"
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
    searchQuery !== '' || selectedEngines.size > 0 || selectedProfiles.size > 0 || selectedStatuses.size > 0;

  const resetAllFilters = () => {
    setSearchQuery('');
    setSelectedEngines(new Set());
    setSelectedProfiles(new Set());
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
              className="absolute right-2.5 top-2 text-[#666666] hover:text-[#ededed] cursor-pointer"
              title="Clear search"
              aria-label="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          ) : (
            <div className="absolute right-2.5 top-1.5 pointer-events-none">
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
            onClick={fetchSessions}
            disabled={loading}
            className="rounded-geist h-8 border-border bg-[#111111] text-[#ededed] hover:bg-[#1a1a1a]"
          >
            <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', loading && 'animate-spin')} />
            Refresh
          </Button>
        </div>
      </div>

      {/* TanStack Table Container */}
      <div className="rounded-geist border border-border bg-[#111111] overflow-hidden shadow-sm">
        <Table containerClassName="overflow-x-hidden" className="table-fixed w-full">
          <TableHeader className="bg-[#0a0a0a] border-b border-border">
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
                <TableCell colSpan={columns.length} className="text-center py-12 text-[#a1a1a1] font-mono text-xs">
                  No sessions match the current search or filters.
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
