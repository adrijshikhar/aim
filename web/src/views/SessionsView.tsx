import * as React from 'react';
import {
  SessionDTO,
  ProfileDTO,
  getSessions,
  resumeSession,
} from '@/lib/api';
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
import { useToast } from '@/components/ui/use-toast';
import {
  Terminal,
  Search,
  RefreshCw,
  Folder,
  Clock,
  MessageSquare,
  Activity,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  X,
  Filter,
} from 'lucide-react';
import { ResumeModal } from '@/components/ResumeModal';
import { AgentBadge } from '@/components/AgentBadge';

interface SessionsViewProps {
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
  profiles?: ProfileDTO[];
}

export function SessionsView({
  selectedProfile = 'all',
  onSelectProfile,
  profiles = [],
}: SessionsViewProps) {
  const [sessions, setSessions] = React.useState<SessionDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [selectedEngine, setSelectedEngine] = React.useState<string>('all');
  const [resumingId, setResumingId] = React.useState<string | null>(null);
  const [flagModalSession, setFlagModalSession] = React.useState<SessionDTO | null>(null);

  // Pagination state
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(15);

  const searchInputRef = React.useRef<HTMLInputElement>(null);
  const { toast } = useToast();

  const fetchSessions = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessions({
        profile: selectedProfile === 'all' ? undefined : selectedProfile,
        agent: selectedEngine === 'all' ? undefined : selectedEngine,
        query: searchQuery ? searchQuery : undefined,
      });
      setSessions(data);
    } catch {
      // Fallback dev data if server is offline
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
        {
          id: 'sess-55128c',
          agent: 'gemini',
          profile: 'research',
          title: 'Evaluate vector embeddings performance',
          cwd: '/Users/developer/Workspace/ai-evals',
          goal: 'Benchmark latency and recall on 100k test vectors across batch sizes',
          turns: 41,
          updated_at: new Date(Date.now() - 1000 * 60 * 360).toISOString(),
          is_active: false,
        },
        {
          id: 'sess-99014b',
          agent: 'agy',
          profile: 'default',
          title: 'Autonomous multi-file refactor',
          cwd: '/Users/developer/Projects/catalyst',
          goal: 'Decouple session runners and add isolated state management hooks',
          turns: 58,
          updated_at: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
          is_active: true,
        },
      ]);
    } finally {
      setLoading(false);
    }
  }, [selectedProfile, selectedEngine, searchQuery]);

  React.useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  // Reset to page 1 whenever filters change
  React.useEffect(() => {
    setPage(1);
  }, [searchQuery, selectedProfile, selectedEngine, pageSize]);

  // Keyboard shortcut listener: `/` to focus search, `[` for prev page, `]` for next page
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const activeEl = document.activeElement;
      const isInput =
        activeEl instanceof HTMLInputElement ||
        activeEl instanceof HTMLTextAreaElement ||
        activeEl instanceof HTMLSelectElement;

      if (e.key === '/' && !isInput) {
        e.preventDefault();
        searchInputRef.current?.focus();
      } else if (!isInput && e.key === '[') {
        setPage((p) => Math.max(1, p - 1));
      } else if (!isInput && e.key === ']') {
        setPage((p) => p + 1);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Compute profile options
  const profileOptions = React.useMemo(() => {
    const set = new Set<string>();
    if (profiles && profiles.length > 0) {
      profiles.forEach((p) => set.add(p.name));
    }
    sessions.forEach((s) => set.add(s.profile));
    return Array.from(set).sort();
  }, [profiles, sessions]);

  const handleResume = async (s: SessionDTO, flags?: string[], customFlags?: string) => {
    if (s.is_active && (!flags || flags.length === 0) && !customFlags) {
      setFlagModalSession(s);
      return;
    }

    setResumingId(s.id);
    try {
      const res = await resumeSession({
        agent: s.agent,
        profile: s.profile,
        session_id: s.id,
        flags: flags,
        custom_flags: customFlags,
      });

      const flagsDesc = flags && flags.length > 0 ? ` with ${flags.join(' ')}` : '';
      toast({
        title: 'Terminal Session Launched',
        description: res.message || `Resumed session ${s.id} (${s.agent}/${s.profile})${flagsDesc} in terminal.`,
        variant: 'success',
      });
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to launch terminal';
      toast({
        title: 'Resume Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setResumingId(null);
    }
  };

  const filteredSessions = React.useMemo(() => {
    return sessions.filter((s) => {
      const matchProfile =
        !selectedProfile || selectedProfile === 'all' || s.profile.toLowerCase() === selectedProfile.toLowerCase();
      if (!matchProfile) return false;

      const matchEngine =
        !selectedEngine || selectedEngine === 'all' || s.agent.toLowerCase() === selectedEngine.toLowerCase();
      if (!matchEngine) return false;

      if (!searchQuery.trim()) return true;
      const q = searchQuery.toLowerCase();
      return (
        s.id.toLowerCase().includes(q) ||
        s.profile.toLowerCase().includes(q) ||
        s.agent.toLowerCase().includes(q) ||
        s.title.toLowerCase().includes(q) ||
        s.goal.toLowerCase().includes(q) ||
        s.cwd.toLowerCase().includes(q)
      );
    });
  }, [sessions, selectedProfile, selectedEngine, searchQuery]);

  // Pagination calculation
  const totalItems = filteredSessions.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));
  const currentPage = Math.min(page, totalPages);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = Math.min(startIndex + pageSize, totalItems);
  const paginatedSessions = React.useMemo(() => {
    return filteredSessions.slice(startIndex, endIndex);
  }, [filteredSessions, startIndex, endIndex]);

  const hasActiveFilters =
    (selectedProfile && selectedProfile !== 'all') ||
    (selectedEngine && selectedEngine !== 'all') ||
    searchQuery.trim().length > 0;

  const resetAllFilters = () => {
    if (onSelectProfile) onSelectProfile('all');
    setSelectedEngine('all');
    setSearchQuery('');
  };

  const formatRelativeTime = (isoString: string) => {
    try {
      const date = new Date(isoString);
      const diffMs = Date.now() - date.getTime();
      const diffMins = Math.floor(diffMs / (1000 * 60));
      if (diffMins < 1) return 'Just now';
      if (diffMins < 60) return `${diffMins}m ago`;
      const diffHours = Math.floor(diffMins / 60);
      if (diffHours < 24) return `${diffHours}h ago`;
      const diffDays = Math.floor(diffHours / 24);
      return `${diffDays}d ago`;
    } catch {
      return isoString;
    }
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
            placeholder="Search sessions by goal, cwd, title, or profile..."
            className="pl-9 pr-8 h-8 text-xs rounded-geist border-[#262626] bg-[#0a0a0a] text-[#ededed] placeholder:text-[#666666] focus:border-[#333333]"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery ? (
            <button
              type="button"
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-2 text-[#666666] hover:text-[#ededed] cursor-pointer"
              title="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          ) : (
            <div className="absolute right-2.5 top-1.5 pointer-events-none">
              <kbd className="text-[10px] font-mono px-1.5 py-0.5 border border-[#262626] bg-[#111111] text-[#a1a1a1] rounded">
                /
              </kbd>
            </div>
          )}
        </div>

        {/* Dropdown Filters and Actions */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Profile Filter Dropdown */}
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-[#666666] font-mono hidden sm:inline">Profile:</span>
            <select
              value={selectedProfile}
              onChange={(e) => onSelectProfile && onSelectProfile(e.target.value)}
              className="h-8 px-2.5 text-xs font-mono rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:border-[#333333] focus:outline-none focus:border-[#333333] cursor-pointer"
              title="Filter by Profile"
            >
              <option value="all">All Profiles</option>
              {profileOptions.map((name) => (
                <option key={name} value={name}>
                  {name}
                </option>
              ))}
            </select>
          </div>

          {/* Engine Filter Dropdown */}
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-[#666666] font-mono hidden sm:inline">Engine:</span>
            <select
              value={selectedEngine}
              onChange={(e) => setSelectedEngine(e.target.value)}
              className="h-8 px-2.5 text-xs font-mono rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:border-[#333333] focus:outline-none focus:border-[#333333] cursor-pointer"
              title="Filter by Engine"
            >
              <option value="all">All Engines</option>
              <option value="claude">Claude</option>
              <option value="codex">Codex</option>
              <option value="gemini">Gemini</option>
              <option value="agy">Antigravity</option>
            </select>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={fetchSessions}
            disabled={loading}
            className="rounded-geist h-8 border-[#262626] bg-[#111111] text-[#ededed] hover:bg-[#1a1a1a]"
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Active Filter Chips */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2 pt-1 text-xs">
          <span className="flex items-center gap-1 text-[#666666] text-[11px] font-mono">
            <Filter className="h-3 w-3" /> Filters:
          </span>

          {selectedProfile && selectedProfile !== 'all' && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] font-mono text-[11px]">
              profile: <strong className="text-[#ededed]">{selectedProfile}</strong>
              <button
                type="button"
                onClick={() => onSelectProfile && onSelectProfile('all')}
                className="hover:text-[#ededed] cursor-pointer"
                title="Clear profile filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          {selectedEngine && selectedEngine !== 'all' && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] font-mono text-[11px]">
              engine: <strong className="text-[#ededed]">{selectedEngine}</strong>
              <button
                type="button"
                onClick={() => setSelectedEngine('all')}
                className="hover:text-[#ededed] cursor-pointer"
                title="Clear engine filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          {searchQuery && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] font-mono text-[11px]">
              query: &ldquo;<strong className="text-[#ededed]">{searchQuery}</strong>&rdquo;
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="hover:text-[#ededed] cursor-pointer"
                title="Clear query filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          <button
            type="button"
            onClick={resetAllFilters}
            className="text-[11px] font-mono text-[#a1a1a1] hover:text-[#ededed] underline ml-1 cursor-pointer"
          >
            Reset all
          </button>
        </div>
      )}

      {/* Sessions Table */}
      <div className="rounded-geist border border-[#262626] bg-[#111111] overflow-hidden">
        <Table>
          <TableHeader className="bg-[#0a0a0a] border-b border-[#262626]">
            <TableRow className="border-b border-[#262626] hover:bg-transparent">
              <TableHead className="w-[120px] font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Agent</TableHead>
              <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Profile</TableHead>
              <TableHead className="min-w-[200px] font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Directory (CWD)</TableHead>
              <TableHead className="min-w-[280px] font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Goal / Title</TableHead>
              <TableHead className="w-[90px] text-center font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Turns</TableHead>
              <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Updated</TableHead>
              <TableHead className="w-[130px] text-right font-mono text-[11px] uppercase tracking-wider text-[#a1a1a1]">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && sessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-[#a1a1a1]">
                  <div className="flex items-center justify-center gap-2 font-mono text-xs">
                    <RefreshCw className="h-4 w-4 animate-spin text-[#ededed]" />
                    <span>Loading sessions...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : filteredSessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-[#a1a1a1]">
                  <div className="flex flex-col items-center justify-center gap-1.5 py-4">
                    <Activity className="h-7 w-7 text-[#666666] mb-1" />
                    <span className="font-medium text-[#ededed] text-xs">No sessions found</span>
                    <span className="text-[11px] text-[#666666]">
                      {hasActiveFilters
                        ? 'Try modifying or clearing your filter criteria.'
                        : 'No active or archived agent sessions detected.'}
                    </span>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              paginatedSessions.map((s) => (
                <TableRow
                  key={s.id}
                  className="group hover:bg-[#161616] transition-colors border-b border-[#262626]"
                >
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <AgentBadge agent={s.agent} />
                      {s.is_active && (
                        <span className="h-1.5 w-1.5 rounded-full bg-[#00e599] shrink-0" title="Active session" />
                      )}
                    </div>
                  </TableCell>

                  <TableCell>
                    <span className="font-mono text-xs px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed]">
                      {s.profile}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1.5 text-xs text-[#a1a1a1] max-w-[240px]">
                      <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                      <span className="font-mono truncate" title={s.cwd}>
                        {s.cwd}
                      </span>
                    </div>
                  </TableCell>

                  <TableCell>
                    <div className="space-y-0.5 max-w-[340px]">
                      <div className="font-medium text-xs text-[#ededed] truncate" title={s.title}>
                        {s.title}
                      </div>
                      <div className="text-[11px] text-[#a1a1a1] truncate" title={s.goal}>
                        {s.goal}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-xs text-[#a1a1a1] font-mono tabular-nums">
                      <MessageSquare className="h-3 w-3 text-[#666666]" />
                      {s.turns}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1 text-xs text-[#a1a1a1] font-mono tabular-nums">
                      <Clock className="h-3 w-3 text-[#666666]" />
                      <span>{formatRelativeTime(s.updated_at)}</span>
                    </div>
                  </TableCell>

                  <TableCell className="text-right">
                    <div className="inline-flex items-center rounded-geist border border-[#262626] bg-[#0a0a0a] overflow-hidden shadow-sm">
                      <button
                        type="button"
                        className="h-7 px-2.5 text-xs font-mono font-medium hover:bg-[#161616] text-[#ededed] inline-flex items-center gap-1.5 transition-colors cursor-pointer border-r border-[#262626] disabled:opacity-50"
                        title="Resume session in terminal"
                        disabled={resumingId === s.id}
                        onClick={() => handleResume(s)}
                      >
                        {resumingId === s.id ? (
                          <RefreshCw className="h-3 w-3 animate-spin text-[#ededed]" />
                        ) : (
                          <Terminal className="h-3 w-3 text-[#ededed]" />
                        )}
                        Resume
                      </button>
                      <button
                        type="button"
                        className="h-7 px-1.5 hover:bg-[#161616] text-[#a1a1a1] hover:text-[#ededed] inline-flex items-center transition-colors cursor-pointer"
                        title="Configure flags & options (--exact, --catalyst, --fork, etc.)"
                        onClick={() => setFlagModalSession(s)}
                      >
                        <ChevronDown className="h-3 w-3" />
                      </button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination & Status Footer */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-[#a1a1a1] px-1 font-mono">
        <div className="flex items-center gap-4">
          <span className="tabular-nums">
            Showing{' '}
            <strong className="text-[#ededed]">
              {totalItems === 0 ? 0 : startIndex + 1}–{endIndex}
            </strong>{' '}
            of <strong className="text-[#ededed]">{totalItems}</strong> session{totalItems !== 1 ? 's' : ''}
          </span>

          {/* Rows per page selector */}
          <div className="flex items-center gap-1.5">
            <span className="text-[11px] text-[#666666]">Per page:</span>
            <select
              value={pageSize}
              onChange={(e) => setPageSize(Number(e.target.value))}
              className="h-6 px-1.5 text-xs font-mono rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:border-[#333333] focus:outline-none cursor-pointer"
            >
              <option value={15}>15</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
            </select>
          </div>
        </div>

        {/* Page Nav controls */}
        <div className="flex items-center gap-2">
          <span className="text-[11px] text-[#666666]">
            Page <strong className="text-[#ededed]">{currentPage}</strong> of{' '}
            <strong className="text-[#ededed]">{totalPages}</strong>
          </span>

          <div className="inline-flex items-center rounded-geist border border-[#262626] bg-[#0a0a0a] overflow-hidden">
            <button
              type="button"
              disabled={currentPage <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              className="h-7 px-2 hover:bg-[#161616] text-[#ededed] disabled:opacity-30 disabled:hover:bg-transparent cursor-pointer disabled:cursor-not-allowed border-r border-[#262626] flex items-center gap-1"
              title="Previous page ([)"
            >
              <ChevronLeft className="h-3.5 w-3.5" />
              <span className="text-[10px] hidden sm:inline">Prev</span>
            </button>
            <button
              type="button"
              disabled={currentPage >= totalPages}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              className="h-7 px-2 hover:bg-[#161616] text-[#ededed] disabled:opacity-30 disabled:hover:bg-transparent cursor-pointer disabled:cursor-not-allowed flex items-center gap-1"
              title="Next page (])"
            >
              <span className="text-[10px] hidden sm:inline">Next</span>
              <ChevronRight className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>

      <ResumeModal
        open={!!flagModalSession}
        onOpenChange={(open) => !open && setFlagModalSession(null)}
        session={flagModalSession}
        isResuming={!!resumingId}
        onLaunch={async (req) => {
          if (flagModalSession) {
            await handleResume(flagModalSession, req.flags, req.custom_flags);
          }
        }}
      />
    </div>
  );
}
