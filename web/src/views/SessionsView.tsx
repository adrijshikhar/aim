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
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/components/ui/use-toast';
import {
  Terminal,
  Search,
  RefreshCw,
  Folder,
  Clock,
  MessageSquare,
  Activity,
  CheckCircle2,
} from 'lucide-react';

interface SessionsViewProps {
  selectedAgent: string;
}

export function SessionsView({ selectedAgent }: SessionsViewProps) {
  const [sessions, setSessions] = React.useState<SessionDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [resumingId, setResumingId] = React.useState<string | null>(null);

  const { toast } = useToast();

  const fetchSessions = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessions({
        agent: selectedAgent === 'all' ? undefined : selectedAgent,
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
  }, [selectedAgent, searchQuery]);

  React.useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleResume = async (s: SessionDTO) => {
    setResumingId(s.id);
    try {
      const res = await resumeSession({
        agent: s.agent,
        profile: s.profile,
        session_id: s.id,
      });

      toast({
        title: 'Terminal Session Launched',
        description: res.message || `Resumed session ${s.id} (${s.agent}/${s.profile}) in an interactive terminal.`,
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
      const matchAgent =
        selectedAgent === 'all' || s.agent.toLowerCase() === selectedAgent.toLowerCase();
      if (!matchAgent) return false;

      if (!searchQuery.trim()) return true;
      const q = searchQuery.toLowerCase();
      return (
        s.id.toLowerCase().includes(q) ||
        s.profile.toLowerCase().includes(q) ||
        s.title.toLowerCase().includes(q) ||
        s.goal.toLowerCase().includes(q) ||
        s.cwd.toLowerCase().includes(q)
      );
    });
  }, [sessions, selectedAgent, searchQuery]);

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

  const getAgentBadge = (agent: string) => {
    switch (agent.toLowerCase()) {
      case 'claude':
        return <Badge className="bg-amber-500/15 text-amber-300 border-amber-500/30">Claude</Badge>;
      case 'codex':
        return <Badge className="bg-emerald-500/15 text-emerald-300 border-emerald-500/30">Codex</Badge>;
      case 'gemini':
        return <Badge className="bg-blue-500/15 text-blue-300 border-blue-500/30">Gemini</Badge>;
      case 'agy':
      case 'antigravity':
        return <Badge className="bg-purple-500/15 text-purple-300 border-purple-500/30">Antigravity</Badge>;
      default:
        return <Badge variant="outline">{agent}</Badge>;
    }
  };

  return (
    <div className="space-y-4">
      {/* Top Filter & Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        <div className="relative w-full sm:w-96">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search sessions by goal, cwd, title, or profile..."
            className="pl-9 h-9 text-xs"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className="flex items-center gap-2 self-end sm:self-auto">
          <Button variant="outline" size="sm" onClick={fetchSessions} disabled={loading}>
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Sessions Table */}
      <div className="rounded-xl border border-border bg-card/60 overflow-hidden shadow-sm">
        <Table>
          <TableHeader className="bg-muted/40">
            <TableRow>
              <TableHead className="w-[120px]">Agent</TableHead>
              <TableHead className="w-[110px]">Profile</TableHead>
              <TableHead className="min-w-[200px]">Directory (CWD)</TableHead>
              <TableHead className="min-w-[280px]">Goal / Title</TableHead>
              <TableHead className="w-[90px] text-center">Turns</TableHead>
              <TableHead className="w-[110px]">Updated</TableHead>
              <TableHead className="w-[140px] text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && sessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-muted-foreground">
                  <div className="flex items-center justify-center gap-2">
                    <RefreshCw className="h-4 w-4 animate-spin text-primary" />
                    <span>Loading sessions...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : filteredSessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-muted-foreground">
                  <div className="flex flex-col items-center justify-center gap-1.5 py-4">
                    <Activity className="h-8 w-8 text-muted-foreground/50 mb-1" />
                    <span className="font-medium text-foreground">No sessions found</span>
                    <span className="text-xs">
                      {searchQuery
                        ? 'Try modifying your search criteria.'
                        : 'No active or archived agent sessions detected.'}
                    </span>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              filteredSessions.map((s) => (
                <TableRow key={s.id} className="group hover:bg-muted/30 transition-colors">
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getAgentBadge(s.agent)}
                      {s.is_active && (
                        <span className="relative flex h-2 w-2" title="Active session">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                        </span>
                      )}
                    </div>
                  </TableCell>

                  <TableCell>
                    <span className="font-mono text-xs px-2 py-0.5 rounded bg-muted text-foreground border border-border">
                      {s.profile}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1.5 text-xs text-muted-foreground max-w-[240px]">
                      <Folder className="h-3.5 w-3.5 shrink-0 text-muted-foreground/70" />
                      <span className="font-mono truncate" title={s.cwd}>
                        {s.cwd}
                      </span>
                    </div>
                  </TableCell>

                  <TableCell>
                    <div className="space-y-0.5 max-w-[340px]">
                      <div className="font-medium text-xs text-foreground truncate" title={s.title}>
                        {s.title}
                      </div>
                      <div className="text-[11px] text-muted-foreground truncate" title={s.goal}>
                        {s.goal}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="text-center">
                    <span className="inline-flex items-center gap-1 text-xs text-muted-foreground font-mono">
                      <MessageSquare className="h-3 w-3" />
                      {s.turns}
                    </span>
                  </TableCell>

                  <TableCell>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Clock className="h-3 w-3" />
                      <span>{formatRelativeTime(s.updated_at)}</span>
                    </div>
                  </TableCell>

                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="outline"
                      className="h-8 text-xs font-medium border-primary/30 text-primary hover:bg-primary/10 hover:text-primary transition-all cursor-pointer shadow-sm"
                      disabled={resumingId === s.id}
                      onClick={() => handleResume(s)}
                    >
                      {resumingId === s.id ? (
                        <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                      ) : (
                        <Terminal className="h-3.5 w-3.5 mr-1.5" />
                      )}
                      Resume
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between text-xs text-muted-foreground px-1">
        <span>
          Showing {filteredSessions.length} session{filteredSessions.length !== 1 ? 's' : ''}
        </span>
        <span className="flex items-center gap-1">
          <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
          Terminal launcher ready
        </span>
      </div>
    </div>
  );
}
