import * as React from 'react';
import {
  SessionDTO,
  SessionTranscriptDTO,
  SessionTurnDTO,
  getSessionTranscript,
  resumeSession,
  ResumeRequest,
} from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { useToast } from '@/components/ui/use-toast';
import { AgentBadge } from '@/components/AgentBadge';
import { ResumeModal } from '@/components/ResumeModal';
import { TurnMessageRenderer } from '@/components/TurnMessageRenderer';
import {
  ArrowLeft,
  Terminal,
  Copy,
  Check,
  Folder,
  MessageSquare,
  Clock,
  ChevronDown,
  ChevronRight,
  ChevronUp,
  ShieldAlert,
  Wrench,
  Sparkles,
  Bot,
  User,
  RefreshCw,
  Search,
  X,
  GitFork,
  FileText,
  PenTool,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface SessionDetailViewProps {
  session: SessionDTO;
  onBack: () => void;
}

const INITIAL_BATCH_SIZE = 25;

const getToolConfig = (toolName: string) => {
  const name = toolName.toLowerCase();
  // 1. Exec / Bash / Commands -> Mint (#9fc9a2)
  if (
    name.includes('run') ||
    name.includes('exec') ||
    name.includes('bash') ||
    name.includes('command') ||
    name.includes('terminal') ||
    name.includes('shell')
  ) {
    return {
      category: 'exec',
      color: '#9fc9a2',
      borderClass: 'border-[#9fc9a2]/25',
      bgClass: 'bg-[#9fc9a2]/5',
      headerHoverClass: 'hover:bg-[#9fc9a2]/10',
      iconClass: 'text-[#9fc9a2]',
      tagClass: 'border-[#9fc9a2]/30 text-[#9fc9a2] bg-[#9fc9a2]/10',
      Icon: Terminal,
    };
  }

  // 2. Read / View / Grep / Search -> Pastel Blue (#9fbbe0)
  if (
    name.includes('read') ||
    name.includes('view') ||
    name.includes('cat') ||
    name.includes('list') ||
    name.includes('get') ||
    name.includes('grep') ||
    name.includes('search') ||
    name.includes('find') ||
    name.includes('inspect')
  ) {
    return {
      category: 'read',
      color: '#9fbbe0',
      borderClass: 'border-[#9fbbe0]/25',
      bgClass: 'bg-[#9fbbe0]/5',
      headerHoverClass: 'hover:bg-[#9fbbe0]/10',
      iconClass: 'text-[#9fbbe0]',
      tagClass: 'border-[#9fbbe0]/30 text-[#9fbbe0] bg-[#9fbbe0]/10',
      Icon: FileText,
    };
  }

  // 3. Edit / Write / Patch / Replace -> Lavender (#c0a8dd)
  if (
    name.includes('write') ||
    name.includes('edit') ||
    name.includes('replace') ||
    name.includes('patch') ||
    name.includes('modify') ||
    name.includes('create') ||
    name.includes('delete')
  ) {
    return {
      category: 'edit',
      color: '#c0a8dd',
      borderClass: 'border-[#c0a8dd]/25',
      bgClass: 'bg-[#c0a8dd]/5',
      headerHoverClass: 'hover:bg-[#c0a8dd]/10',
      iconClass: 'text-[#c0a8dd]',
      tagClass: 'border-[#c0a8dd]/30 text-[#c0a8dd] bg-[#c0a8dd]/10',
      Icon: PenTool,
    };
  }

  // 4. Default / General Tool
  return {
    category: 'tool',
    color: '#a1a1a1',
    borderClass: 'border-[#262626]',
    bgClass: 'bg-[#0a0a0a]',
    headerHoverClass: 'hover:bg-[#121212]',
    iconClass: 'text-[#a1a1a1]',
    tagClass: 'border-[#333333] text-[#ededed] bg-[#161616]',
    Icon: Wrench,
  };
};

export const SessionDetailView: React.FC<SessionDetailViewProps> = ({ session, onBack }) => {
  const [transcript, setTranscript] = React.useState<SessionTranscriptDTO | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [skipPermissions, setSkipPermissions] = React.useState(false);
  const [resumeModalOpen, setResumeModalOpen] = React.useState(false);
  const [targetForkTurn, setTargetForkTurn] = React.useState<number | null>(null);
  const [isResuming, setIsResuming] = React.useState(false);
  const [copiedCmd, setCopiedCmd] = React.useState(false);
  const [copiedCwd, setCopiedCwd] = React.useState(false);
  const [copiedTurn, setCopiedTurn] = React.useState<number | null>(null);
  const [expandedThinking, setExpandedThinking] = React.useState<Record<number, boolean>>({});
  const [expandedTools, setExpandedTools] = React.useState<Record<string, boolean>>({});

  // Search & Filter state for Conversation History
  const [turnSearch, setTurnSearch] = React.useState('');
  const [roleFilter, setRoleFilter] = React.useState<'all' | 'user' | 'assistant'>('all');
  const [visibleCount, setVisibleCount] = React.useState(INITIAL_BATCH_SIZE);
  const [showScrollTop, setShowScrollTop] = React.useState(false);

  const { toast } = useToast();
  const topRef = React.useRef<HTMLDivElement>(null);
  const bottomSentinelRef = React.useRef<HTMLDivElement>(null);

  const fetchTranscript = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessionTranscript(session.agent, session.id);
      setTranscript(data);
      setVisibleCount(INITIAL_BATCH_SIZE);
    } catch {
      // Fallback synthetic transcript from session metadata
      const fallbackData: SessionTranscriptDTO = {
        session_id: session.id,
        agent: session.agent,
        profile: session.profile,
        title: session.title,
        cwd: session.cwd,
        goal: session.goal,
        turns: [
          {
            turn_index: 1,
            role: 'user',
            content: session.goal || session.title || 'Session initiated',
          },
          {
            turn_index: 2,
            role: 'assistant',
            content: `Session in ${session.cwd} with ${session.turns} recorded turns.`,
          },
        ],
      };
      setTranscript(fallbackData);
      setVisibleCount(INITIAL_BATCH_SIZE);
    } finally {
      setLoading(false);
    }
  }, [session.agent, session.id, session.goal, session.title, session.cwd, session.turns, session.profile]);

  React.useEffect(() => {
    fetchTranscript();
  }, [fetchTranscript]);

  const toggleThinking = (turnIndex: number) => {
    setExpandedThinking((prev) => ({
      ...prev,
      [turnIndex]: !prev[turnIndex],
    }));
  };

  const toggleTool = (toolId: string) => {
    setExpandedTools((prev) => ({
      ...prev,
      [toolId]: !prev[toolId],
    }));
  };

  const buildCliCommand = () => {
    const prof = session.profile && session.profile !== '<host>' ? session.profile : '';
    let cmd = prof ? `aim resume ${session.agent} ${prof} ${session.id}` : `aim resume ${session.agent} ${session.id}`;
    if (skipPermissions && session.agent.toLowerCase() === 'claude') {
      cmd += ' -- --dangerously-skip-permissions';
    }
    return cmd;
  };

  const copyCliCommand = () => {
    navigator.clipboard.writeText(buildCliCommand());
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
    toast({
      title: 'Command copied',
      description: buildCliCommand(),
    });
  };

  const copyCwd = () => {
    navigator.clipboard.writeText(session.cwd);
    setCopiedCwd(true);
    setTimeout(() => setCopiedCwd(false), 2000);
  };

  const copyTurnContent = (turn: SessionTurnDTO) => {
    const text = turn.content || (turn.tool_calls ? JSON.stringify(turn.tool_calls, null, 2) : '');
    navigator.clipboard.writeText(text);
    setCopiedTurn(turn.turn_index);
    setTimeout(() => setCopiedTurn(null), 2000);
    toast({
      title: 'Turn copied',
      description: `Turn #${turn.turn_index} copied to clipboard`,
    });
  };

  const handleQuickResume = async () => {
    setIsResuming(true);
    try {
      const adapterFlags: string[] = [];
      if (skipPermissions && session.agent.toLowerCase() === 'claude') {
        adapterFlags.push('--dangerously-skip-permissions');
      }
      const res = await resumeSession({
        agent: session.agent,
        profile: session.profile,
        session_id: session.id,
        adapter_flags: adapterFlags,
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
      setIsResuming(false);
    }
  };

  const handleModalLaunch = async (req: ResumeRequest) => {
    setIsResuming(true);
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
      setIsResuming(false);
    }
  };

  // Floating scroll-to-top button detector (scrolled > 350px)
  React.useEffect(() => {
    let ticking = false;
    const mainElem = document.querySelector('main');
    const handleScroll = () => {
      if (!ticking) {
        window.requestAnimationFrame(() => {
          const scrollY = mainElem ? mainElem.scrollTop : window.scrollY;
          setShowScrollTop(scrollY > 350);
          ticking = false;
        });
        ticking = true;
      }
    };

    if (mainElem) {
      mainElem.addEventListener('scroll', handleScroll, { passive: true });
    }
    window.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      if (mainElem) {
        mainElem.removeEventListener('scroll', handleScroll);
      }
      window.removeEventListener('scroll', handleScroll);
    };
  }, []);

  const scrollToTop = () => {
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const behavior: ScrollBehavior = prefersReducedMotion ? 'auto' : 'smooth';
    const mainElem = document.querySelector('main');
    if (mainElem && mainElem.scrollTop > 0) {
      mainElem.scrollTo({ top: 0, behavior });
    } else {
      window.scrollTo({ top: 0, behavior });
    }
  };

  // Synchronized turns count
  const displayTurns = transcript?.turns?.length ?? session.turns;

  // Filtered turns based on search and role filter
  const allTurns = transcript?.turns || [];
  const filteredTurns = React.useMemo(() => {
    return allTurns.filter((turn) => {
      // Role filter
      if (roleFilter !== 'all' && turn.role !== roleFilter) {
        return false;
      }
      // Text search query
      if (turnSearch.trim()) {
        const query = turnSearch.toLowerCase();
        const contentMatch = turn.content?.toLowerCase().includes(query);
        const thinkingMatch = turn.thinking?.toLowerCase().includes(query);
        const toolMatch = turn.tool_calls?.some(
          (t) =>
            t.tool_name.toLowerCase().includes(query) ||
            t.arguments?.toLowerCase().includes(query) ||
            t.output?.toLowerCase().includes(query)
        );
        if (!contentMatch && !thinkingMatch && !toolMatch) {
          return false;
        }
      }
      return true;
    });
  }, [allTurns, turnSearch, roleFilter]);

  // Reverse chronological order: newest turns first at the top, traversing back in time down to Turn #1
  const reversedTurns = React.useMemo(() => {
    return [...filteredTurns].reverse();
  }, [filteredTurns]);

  // Reset visible turns count whenever query or role filter changes
  React.useEffect(() => {
    setVisibleCount(INITIAL_BATCH_SIZE);
  }, [turnSearch, roleFilter]);

  // Infinite scroll slice (top N newest turns)
  const visibleTurns = React.useMemo(() => {
    return reversedTurns.slice(0, visibleCount);
  }, [reversedTurns, visibleCount]);

  // Infinite scroll intersection observer observing bottom sentinel
  React.useEffect(() => {
    const sentinel = bottomSentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setVisibleCount((prev) => {
            if (prev < reversedTurns.length) {
              return Math.min(prev + INITIAL_BATCH_SIZE, reversedTurns.length);
            }
            return prev;
          });
        }
      },
      { root: null, rootMargin: '400px' }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [reversedTurns.length]);

  return (
    <div ref={topRef} className="space-y-4 pb-12">
      {/* Navigation Breadcrumb & Back Action */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={onBack}
            className="h-8 px-2 text-xs font-sans text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist"
          >
            <ArrowLeft className="h-3.5 w-3.5 mr-1" />
            Sessions
          </Button>
          <span className="text-[#333333] font-mono text-xs">/</span>
          <AgentBadge agent={session.agent} />
          <span className="text-[#333333] font-mono text-xs">/</span>
          <span className="font-mono text-xs text-[#a1a1a1] truncate max-w-[200px]" title={session.id}>
            {session.id.slice(0, 12)}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchTranscript}
            disabled={loading}
            className="h-8 px-2.5 text-xs font-sans border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] rounded-geist"
          >
            <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', loading && 'animate-spin')} />
            Reload Replay
          </Button>
        </div>
      </div>

      {/* Sticky Header Card */}
      <div className="rounded-geist border border-border bg-[#111111] p-5 shadow-lg space-y-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="space-y-1.5 min-w-0">
            <div className="flex items-center gap-2.5 flex-wrap">
              <h1 className="text-base font-semibold text-[#ededed] tracking-tight font-sans">
                {session.title || 'Untitled Session'}
              </h1>
              {session.is_active && (
                <Badge className="badge-success text-[10px] font-mono px-2 py-0.5 rounded-[4px] flex items-center gap-1.5">
                  <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                  Active Process
                </Badge>
              )}
            </div>

            <p className="text-xs text-[#a1a1a1] font-sans leading-relaxed max-w-3xl">
              {session.goal || 'No explicit goal specified for this session.'}
            </p>
          </div>

          {/* Action Toolbar — Uniform h-8 sizing, font-sans, aligned */}
          <div className="flex flex-wrap items-center gap-2 shrink-0">
            {/* Quick Skip Permissions Toggle (for Claude Code) */}
            {session.agent.toLowerCase() === 'claude' && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSkipPermissions(!skipPermissions)}
                className={cn(
                  'h-8 px-3 text-xs font-sans rounded-geist border transition-colors flex items-center gap-1.5',
                  skipPermissions
                    ? 'border-amber-500/50 bg-amber-500/15 text-amber-300 hover:bg-amber-500/20'
                    : 'border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]'
                )}
                title="Bypass confirmation prompts for tools and commands"
              >
                <ShieldAlert className={cn('h-3.5 w-3.5', skipPermissions ? 'text-amber-400' : 'text-[#666666]')} />
                <span>{skipPermissions ? 'Skip Perms: ON' : 'Skip Perms: OFF'}</span>
              </Button>
            )}

            {/* Quick Resume Button */}
            <Button
              size="sm"
              onClick={handleQuickResume}
              disabled={isResuming}
              className="h-8 px-3 text-xs font-sans font-medium rounded-geist bg-[#ededed] text-black hover:bg-white flex items-center gap-1.5 cursor-pointer shadow-sm transition-colors"
            >
              {isResuming ? (
                <RefreshCw className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Terminal className="h-3.5 w-3.5" />
              )}
              <span>Resume</span>
            </Button>

            {/* Full Flags Modal Trigger */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setTargetForkTurn(null);
                setResumeModalOpen(true);
              }}
              className="h-8 px-3 text-xs font-sans font-medium border-border bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616] rounded-geist flex items-center gap-1.5 cursor-pointer"
              title="Configure all AIM and adapter flags"
            >
              Flags...
            </Button>

            {/* Copy Command */}
            <Button
              variant="outline"
              size="sm"
              onClick={copyCliCommand}
              className="h-8 px-2.5 text-xs font-sans border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist flex items-center justify-center cursor-pointer"
              title="Copy terminal command"
            >
              {copiedCmd ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            </Button>
          </div>
        </div>

        {/* Metadata Pills — Synchronized Turns Count */}
        <div className="flex flex-wrap items-center gap-2.5 pt-3 border-t border-border text-xs text-[#a1a1a1] font-sans">
          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border">
            <span className="text-[#666666]">Profile:</span>
            <span className="text-[#ededed] font-medium">{session.profile || 'default'}</span>
          </div>

          <button
            type="button"
            onClick={copyCwd}
            className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border cursor-pointer hover:border-[#333333] hover:text-[#ededed] transition-colors focus:outline-none focus:ring-1 focus:ring-ring"
            title="Click to copy path"
            aria-label={`Copy current working directory: ${session.cwd}`}
          >
            <Folder className="h-3 w-3 text-[#666666]" />
            <span className="truncate max-w-[260px] text-[#ededed] font-mono text-[11px]">{session.cwd}</span>
            {copiedCwd ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3 text-[#666666]" />}
          </button>

          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border">
            <MessageSquare className="h-3 w-3 text-[#666666]" />
            <span className="text-[#ededed] font-medium">{displayTurns} turns</span>
          </div>

          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border ml-auto">
            <Clock className="h-3 w-3 text-[#666666]" />
            <span className="font-mono text-[11px]">{new Date(session.updated_at).toLocaleString()}</span>
          </div>
        </div>
      </div>

      {/* Conversation Timeline Explorer Section */}
      <div className="space-y-3">
        {/* Explorer Header & Controls Bar */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[#111111] p-3 rounded-geist border border-border">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs font-semibold uppercase tracking-wider text-[#ededed] flex items-center gap-1.5 font-sans">
              <Sparkles className="h-3.5 w-3.5 text-[#0070f3]" />
              Conversation Explorer
            </span>
            <Badge
              variant="outline"
              className="text-[10px] font-mono bg-[#0a0a0a] border-border text-[#a1a1a1] px-1.5 py-0.5 rounded-[4px]"
            >
              {allTurns.length} turns
            </Badge>

            {/* Role Filter Tabs */}
            <div className="flex items-center bg-[#0a0a0a] border border-border rounded-[4px] p-0.5 text-[11px] font-sans">
              <button
                type="button"
                onClick={() => setRoleFilter('all')}
                className={cn(
                  'px-2 py-0.5 rounded transition-colors cursor-pointer',
                  roleFilter === 'all' ? 'bg-[#222222] text-[#ededed] font-medium' : 'text-[#888888] hover:text-[#ededed]'
                )}
              >
                All
              </button>
              <button
                type="button"
                onClick={() => setRoleFilter('user')}
                className={cn(
                  'px-2 py-0.5 rounded transition-colors cursor-pointer',
                  roleFilter === 'user' ? 'bg-[#222222] text-[#ededed] font-medium' : 'text-[#888888] hover:text-[#ededed]'
                )}
              >
                User Prompts
              </button>
              <button
                type="button"
                onClick={() => setRoleFilter('assistant')}
                className={cn(
                  'px-2 py-0.5 rounded transition-colors cursor-pointer',
                  roleFilter === 'assistant' ? 'bg-[#222222] text-[#ededed] font-medium' : 'text-[#888888] hover:text-[#ededed]'
                )}
              >
                Responses
              </button>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            {/* Search within conversation */}
            <div className="relative min-w-[200px] max-w-xs flex-1">
              <Search className="absolute left-2.5 top-2 h-3.5 w-3.5 text-[#666666]" />
              <Input
                placeholder="Search transcript..."
                value={turnSearch}
                onChange={(e) => setTurnSearch(e.target.value)}
                className="h-7 pl-8 pr-7 text-xs bg-[#0a0a0a] border-border text-[#ededed] placeholder:text-[#666666] focus:border-[#333333] rounded-[4px]"
              />
              {turnSearch && (
                <button
                  type="button"
                  onClick={() => setTurnSearch('')}
                  className="absolute right-2 top-1.5 text-[#666666] hover:text-[#ededed] cursor-pointer"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Loading & Empty States */}
        {loading ? (
          <div className="rounded-geist border border-border bg-[#111111] p-12 text-center space-y-3">
            <RefreshCw className="h-6 w-6 animate-spin mx-auto text-[#666666]" />
            <p className="text-xs font-sans text-[#a1a1a1]">Loading conversation transcript from disk...</p>
          </div>
        ) : filteredTurns.length === 0 ? (
          <div className="rounded-geist border border-border bg-[#111111] p-8 text-center space-y-2">
            <p className="text-xs font-sans text-[#a1a1a1]">
              {turnSearch
                ? `No turns matched the query "${turnSearch}".`
                : 'No transcript records found for this session.'}
            </p>
            {turnSearch && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setTurnSearch('')}
                className="h-7 text-xs font-sans text-[#a1a1a1] hover:text-[#ededed]"
              >
                Clear search filter
              </Button>
            )}
          </div>
        ) : (
          <>
            {/* Turns Stream Header & Status Bar */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 px-2 py-1 text-xs text-[#a1a1a1] font-sans">
              <div className="flex items-center gap-2 flex-wrap">
                <span>
                  Showing{' '}
                  <strong className="text-[#ededed] font-medium font-mono">
                    {Math.min(visibleCount, reversedTurns.length)}
                  </strong>{' '}
                  of{' '}
                  <strong className="text-[#ededed] font-medium font-mono">
                    {reversedTurns.length}
                  </strong>{' '}
                  turns
                  <span className="text-[#666666] ml-1.5 font-normal">(Newest first)</span>
                </span>
                {turnSearch && (
                  <Badge variant="outline" className="text-[10px] font-sans border-[#262626] text-[#ededed] bg-[#141414] px-1.5 py-0">
                    Filtered
                  </Badge>
                )}
              </div>

              {reversedTurns.length > visibleCount && (
                <span className="text-[11px] text-[#666666] font-sans">
                  Scroll down to load older turns ({reversedTurns.length - visibleCount} remaining)
                </span>
              )}
            </div>

            {/* Conversation Stream Items (Reverse Chronological: Latest at Top) */}
            <div className="space-y-3">
              {visibleTurns.map((turn) => {
                const isUser = turn.role === 'user';
                const hasThinking = !!turn.thinking?.trim();
                const hasTools = turn.tool_calls && turn.tool_calls.length > 0;
                const isThinkingOpen = !!expandedThinking[turn.turn_index];
                const isCopied = copiedTurn === turn.turn_index;

                return (
                  <div
                    key={turn.turn_index}
                    className={cn(
                      'rounded-geist transition-colors border p-4',
                      isUser
                        ? 'border-[#262626] bg-[#0c0c0c]'
                        : 'border-[#1f1f1f] bg-[#111111]'
                    )}
                  >
                    {/* Stream Entry Header */}
                    <div className="flex items-center justify-between pb-2 border-b border-[#1c1c1c] text-xs font-sans mb-3">
                      <div className="flex items-center gap-2">
                        {isUser ? (
                          <div className="h-5 w-5 rounded-[4px] bg-[#1a1a1a] border border-[#2a2a2a] flex items-center justify-center text-[#ededed]">
                            <User className="h-3 w-3" />
                          </div>
                        ) : (
                          <div className="h-5 w-5 rounded-[4px] bg-[#1a1a1a] border border-[#2a2a2a] flex items-center justify-center text-[#ededed]">
                            <Bot className="h-3 w-3" />
                          </div>
                        )}
                        <span className="font-semibold text-[#ededed]">
                          {isUser ? 'User Prompt' : `${session.agent} Response`}
                        </span>
                        <span className="text-[10px] font-mono text-[#666666]">
                          Turn #{turn.turn_index}
                        </span>
                      </div>

                      <div className="flex items-center gap-2">
                        {turn.timestamp && (
                          <span className="text-[10px] font-mono text-[#666666]">
                            {new Date(turn.timestamp).toLocaleTimeString()}
                          </span>
                        )}

                        {/* Resume / Fork From Here Button (Warm Gold #c08532) */}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setTargetForkTurn(turn.turn_index);
                            setResumeModalOpen(true);
                          }}
                          className="h-6 px-2 text-[10px] font-sans text-[#c08532] hover:text-[#e09e42] hover:bg-[#c08532]/10 border border-[#c08532]/30 bg-[#c08532]/5 rounded-[3px] flex items-center gap-1 transition-colors cursor-pointer"
                          title={`Resume or fork session starting from Turn #${turn.turn_index}`}
                        >
                          <GitFork className="h-2.5 w-2.5 text-[#c08532]" />
                          <span>Fork from here</span>
                        </Button>

                        {/* Copy Turn Content */}
                        <button
                          type="button"
                          onClick={() => copyTurnContent(turn)}
                          className="p-1 text-[#666666] hover:text-[#ededed] hover:bg-[#1a1a1a] rounded transition-colors"
                          title="Copy turn content"
                        >
                          {isCopied ? (
                            <Check className="h-3 w-3 text-emerald-400" />
                          ) : (
                            <Copy className="h-3 w-3" />
                          )}
                        </button>
                      </div>
                    </div>

                    {/* Entry Body */}
                    <div className="space-y-3">
                      {/* Collapsible Reasoning Block (Cursor Peach #dfa88f) */}
                      {hasThinking && (
                        <div className="rounded-[4px] border border-[#dfa88f]/30 bg-[#dfa88f]/5 overflow-hidden">
                          <button
                            type="button"
                            onClick={() => toggleThinking(turn.turn_index)}
                            className="w-full flex items-center justify-between px-3 py-1.5 text-[11px] font-sans text-[#dfa88f] hover:bg-[#dfa88f]/10 transition-colors cursor-pointer"
                          >
                            <span className="flex items-center gap-1.5">
                              {isThinkingOpen ? (
                                <ChevronDown className="h-3 w-3 text-[#dfa88f]" />
                              ) : (
                                <ChevronRight className="h-3 w-3 text-[#dfa88f]" />
                              )}
                              <Sparkles className="h-3 w-3 text-[#dfa88f]" />
                              <span className="font-medium">Reasoning &amp; Thinking Process</span>
                            </span>
                            <span className="text-[10px] font-mono text-[#dfa88f]/80 px-1.5 py-0.5 rounded border border-[#dfa88f]/30 bg-[#dfa88f]/10">
                              {isThinkingOpen ? 'collapse' : `${turn.thinking!.length} chars`}
                            </span>
                          </button>
                          {isThinkingOpen && (
                            <div className="p-3 border-t border-[#dfa88f]/20 text-xs font-mono text-[#e4cec4] bg-[#0c0908] whitespace-pre-wrap leading-relaxed max-h-96 overflow-y-auto">
                              {turn.thinking}
                            </div>
                          )}
                        </div>
                      )}

                      {/* Tool Calls Execution Cards (Cursor AI Action Timeline) */}
                      {hasTools && (
                        <div className="space-y-1.5">
                          {turn.tool_calls!.map((tool, idx) => {
                            const toolKey = `${turn.turn_index}-${idx}-${tool.tool_name}`;
                            const isToolOpen = expandedTools[toolKey] === true; // Default collapsed for clean stream
                            const cfg = getToolConfig(tool.tool_name);
                            const ToolIcon = cfg.Icon;

                            return (
                              <div
                                key={toolKey}
                                className={cn(
                                  'rounded-[4px] border overflow-hidden transition-colors',
                                  cfg.borderClass,
                                  cfg.bgClass
                                )}
                              >
                                <button
                                  type="button"
                                  onClick={() => toggleTool(toolKey)}
                                  aria-expanded={isToolOpen}
                                  aria-controls={`tool-panel-${toolKey}`}
                                  className={cn(
                                    'w-full flex items-center justify-between px-3 py-1.5 text-xs font-mono cursor-pointer transition-colors select-none text-left focus:outline-none focus:ring-1 focus:ring-ring',
                                    cfg.headerHoverClass
                                  )}
                                >
                                  <div className="flex items-center gap-2 min-w-0">
                                    <ToolIcon className={cn('h-3.5 w-3.5 shrink-0', cfg.iconClass)} />
                                    <span className="text-[#ededed] font-medium">{tool.tool_name}</span>
                                    <Badge
                                      variant="outline"
                                      className={cn('text-[9px] font-mono px-1 py-0 rounded-[2px] uppercase tracking-wider', cfg.tagClass)}
                                    >
                                      {cfg.category}
                                    </Badge>
                                  </div>
                                  <div className="flex items-center gap-2">
                                    <Badge
                                      variant={tool.status === 'error' ? 'destructive' : 'success'}
                                      className="text-[9px] font-mono px-1.5 py-0 rounded-[3px]"
                                    >
                                      {tool.status || 'ok'}
                                    </Badge>
                                    {isToolOpen ? (
                                      <ChevronDown className="h-3 w-3 text-[#666666]" />
                                    ) : (
                                      <ChevronRight className="h-3 w-3 text-[#666666]" />
                                    )}
                                  </div>
                                </button>

                                {isToolOpen && (
                                  <div id={`tool-panel-${toolKey}`} className="p-3 border-t border-[#1c1c1c] space-y-2 text-xs font-mono bg-[#050505]">
                                    {tool.arguments && (
                                      <div className="space-y-1">
                                        <span className="text-[10px] text-[#666666] uppercase tracking-wider">
                                          Arguments:
                                        </span>
                                        <pre className="p-2 rounded-[3px] bg-[#0c0c0c] border border-[#1f1f1f] text-[11px] text-[#a1a1a1] overflow-x-auto whitespace-pre-wrap">
                                          {tool.arguments}
                                        </pre>
                                      </div>
                                    )}
                                    {tool.output && (
                                      <div className="space-y-1">
                                        <span className="text-[10px] text-[#666666] uppercase tracking-wider">
                                          Output:
                                        </span>
                                        <div className="p-2 rounded-[3px] bg-[#0c0c0c] border border-[#1f1f1f] text-[11px] text-[#a1a1a1] overflow-x-auto">
                                          <TurnMessageRenderer content={tool.output} />
                                        </div>
                                      </div>
                                    )}
                                  </div>
                                )}
                              </div>
                            );
                          })}
                        </div>
                      )}

                      {/* Main Prose Content */}
                      {turn.content && (
                        <div className="text-xs text-[#ededed] font-sans leading-relaxed selection:bg-[#333333]">
                          <TurnMessageRenderer content={turn.content} />
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Infinite scroll sentinel & status indicators */}
            <div ref={bottomSentinelRef} className="h-6 w-full pointer-events-none" />

            {visibleCount < reversedTurns.length ? (
              <div className="py-6 text-center text-xs font-sans text-[#777777] flex items-center justify-center gap-2">
                <RefreshCw className="h-3.5 w-3.5 animate-spin text-[#666666]" />
                <span>Loading older conversation turns...</span>
              </div>
            ) : reversedTurns.length > 0 ? (
              <div className="py-8 text-center text-xs font-sans text-[#555555] border-t border-[#1c1c1c] flex items-center justify-center gap-2">
                <span>Reached beginning of session transcript ({reversedTurns.length} total turns)</span>
              </div>
            ) : null}
          </>
        )}
      </div>

      {/* Resume Modal */}
      <ResumeModal
        open={resumeModalOpen}
        onOpenChange={(open) => {
          setResumeModalOpen(open);
          if (!open) setTargetForkTurn(null);
        }}
        session={session}
        onLaunch={handleModalLaunch}
        isResuming={isResuming}
        initialFork={targetForkTurn !== null}
        forkTurnIndex={targetForkTurn}
      />

      {/* Floating Scroll-to-Top Action Button (Appears after scrolling down) */}
      {showScrollTop && (
        <button
          type="button"
          onClick={scrollToTop}
          className="fixed bottom-8 right-8 z-50 px-3.5 py-2.5 rounded-full border border-[#333333] bg-[#111111]/90 backdrop-blur-md text-[#ededed] hover:text-white hover:border-[#555555] shadow-2xl transition-all cursor-pointer flex items-center gap-1.5 text-xs font-sans font-medium group animate-in fade-in slide-in-from-bottom-2 duration-200"
          title="Scroll to top of session"
        >
          <ChevronUp className="h-4 w-4 group-hover:-translate-y-0.5 transition-transform text-[#0070f3]" />
          <span>Top</span>
        </button>
      )}
    </div>
  );
};
