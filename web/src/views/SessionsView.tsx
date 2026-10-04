import * as React from 'react';
import {
  SessionDTO,
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
} from 'lucide-react';
import { ResumeModal } from '@/components/ResumeModal';
import { AgentBadge } from '@/components/AgentBadge';

interface SessionsViewProps {
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
}

export function SessionsView({ selectedProfile = 'all', onSelectProfile }: SessionsViewProps) {
  const [sessions, setSessions] = React.useState<SessionDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [resumingId, setResumingId] = React.useState<string | null>(null);
  const [flagModalSession, setFlagModalSession] = React.useState<SessionDTO | null>(null);

  const { toast } = useToast();

  const fetchSessions = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessions({
        profile: selectedProfile === 'all' ? undefined : selectedProfile,
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
  }, [selectedProfile, searchQuery]);

  React.useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleResume = async (s: SessionDTO, flags?: string[], customFlags?: string) => {
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
  }, [sessions, selectedProfile, searchQuery]);

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
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        <div className="relative w-full sm:w-96">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#666666]" />
          <Input
            placeholder="Search sessions by goal, cwd, title, or profile..."
            className="pl-9 pr-8 h-8 text-xs rounded-geist border-[#262626] bg-[#0e0e0e]"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <div className="absolute right-2.5 top-1.5 pointer-events-none">
            <kbd className="text-[10px] font-mono px-1.5 py-0.5 border border-[#2e2e2e] bg-[#171717] text-[#888888] rounded">
              /
            </kbd>
          </div>
        </div>

        <div className="flex items-center gap-2 self-end sm:self-auto">
          {selectedProfile && selectedProfile !== 'all' && (
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-geist border border-[#2a2a2a] bg-[#141414] text-xs font-mono text-[#888888]">
              <span>Profile: <strong className="text-[#ededed]">{selectedProfile}</strong></span>
              {onSelectProfile && (
                <button
                  type="button"
                  onClick={() => onSelectProfile('all')}
                  className="hover:text-[#ededed] ml-1 text-[10px] text-[#666666] cursor-pointer"
                  title="Clear profile filter"
                >
                  ✕
                </button>
              )}
            </div>
          )}
          <Button variant="outline" size="sm" onClick={fetchSessions} disabled={loading} className="rounded-geist h-8">
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Sessions Table */}
      <div className="rounded-geist border border-[#262626] bg-[#111111] overflow-hidden">
        <Table>
          <TableHeader className="bg-[#141414] border-b border-[#262626]">
            <TableRow className="border-b border-[#262626] hover:bg-transparent">
              <TableHead className="w-[120px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Agent</TableHead>
              <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Profile</TableHead>
              <TableHead className="min-w-[200px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Directory (CWD)</TableHead>
              <TableHead className="min-w-[280px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Goal / Title</TableHead>
              <TableHead className="w-[90px] text-center font-mono text-[11px] uppercase tracking-wider text-[#888888]">Turns</TableHead>
              <TableHead className="w-[110px] font-mono text-[11px] uppercase tracking-wider text-[#888888]">Updated</TableHead>
              <TableHead className="w-[130px] text-right font-mono text-[11px] uppercase tracking-wider text-[#888888]">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && sessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-[#888888]">
                  <div className="flex items-center justify-center gap-2 font-mono text-xs">
                    <RefreshCw className="h-4 w-4 animate-spin text-[#ededed]" />
                    <span>Loading sessions...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : filteredSessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-[#888888]">
                  <div className="flex flex-col items-center justify-center gap-1.5 py-4">
                    <Activity className="h-7 w-7 text-[#555555] mb-1" />
                    <span className="font-medium text-[#ededed] text-xs">No sessions found</span>
                    <span className="text-[11px]">
                      {searchQuery
                        ? 'Try modifying your search criteria.'
                        : 'No active or archived agent sessions detected.'}
                    </span>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              filteredSessions.map((s) => (
                <TableRow key={s.id} className="group hover:bg-[#161616] transition-colors border-b border-[#1f1f1f]">
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <AgentBadge agent={s.agent} />
                      {s.is_active && (
                        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 shrink-0" title="Active session" />
                      )}
                    </div>
                  </TableCell>

                  <TableCell>
                    <span className="font-mono text-xs px-2 py-0.5 rounded border border-[#2e2e2e] bg-[#171717] text-[#ededed]">
                      {s.profile}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1.5 text-xs text-[#888888] max-w-[240px]">
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
                      <div className="text-[11px] text-[#777777] truncate" title={s.goal}>
                        {s.goal}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-xs text-[#888888] font-mono tabular-nums">
                      <MessageSquare className="h-3 w-3 text-[#555555]" />
                      {s.turns}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1 text-xs text-[#888888] font-mono tabular-nums">
                      <Clock className="h-3 w-3 text-[#555555]" />
                      <span>{formatRelativeTime(s.updated_at)}</span>
                    </div>
                  </TableCell>

                  <TableCell className="text-right">
                    <div className="inline-flex items-center rounded-geist border border-[#333333] bg-[#161616] overflow-hidden shadow-sm">
                      <button
                        type="button"
                        className="h-7 px-2.5 text-xs font-mono font-medium hover:bg-[#222222] text-[#ededed] inline-flex items-center gap-1.5 transition-colors cursor-pointer border-r border-[#262626] disabled:opacity-50"
                        title="Resume session in terminal"
                        disabled={resumingId === s.id}
                        onClick={() => handleResume(s)}
                      >
                        {resumingId === s.id ? (
                          <RefreshCw className="h-3 w-3 animate-spin text-[#0070f3]" />
                        ) : (
                          <Terminal className="h-3 w-3 text-[#0070f3]" />
                        )}
                        Resume
                      </button>
                      <button
                        type="button"
                        className="h-7 px-1.5 hover:bg-[#222222] text-[#888888] hover:text-[#ededed] inline-flex items-center transition-colors cursor-pointer"
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

      <div className="flex items-center justify-between text-xs text-[#888888] px-1 font-mono">
        <span className="tabular-nums">
          Showing {filteredSessions.length} session{filteredSessions.length !== 1 ? 's' : ''}
        </span>
        <span className="flex items-center gap-1.5 text-[#888888]">
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
          Terminal launcher ready
        </span>
      </div>
    </div>
  );
}
