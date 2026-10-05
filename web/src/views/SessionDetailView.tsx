import * as React from 'react';
import {
  SessionDTO,
  SessionTranscriptDTO,
  getSessionTranscript,
  resumeSession,
  ResumeRequest,
} from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/components/ui/use-toast';
import { AgentBadge } from '@/components/AgentBadge';
import { ResumeModal } from '@/components/ResumeModal';
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
  ShieldAlert,
  Wrench,
  Sparkles,
  Bot,
  User,
  RefreshCw,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface SessionDetailViewProps {
  session: SessionDTO;
  onBack: () => void;
}

export const SessionDetailView: React.FC<SessionDetailViewProps> = ({ session, onBack }) => {
  const [transcript, setTranscript] = React.useState<SessionTranscriptDTO | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [skipPermissions, setSkipPermissions] = React.useState(false);
  const [resumeModalOpen, setResumeModalOpen] = React.useState(false);
  const [isResuming, setIsResuming] = React.useState(false);
  const [copiedCmd, setCopiedCmd] = React.useState(false);
  const [copiedCwd, setCopiedCwd] = React.useState(false);
  const [expandedThinking, setExpandedThinking] = React.useState<Record<number, boolean>>({});
  const [expandedTools, setExpandedTools] = React.useState<Record<string, boolean>>({});

  const { toast } = useToast();

  const fetchTranscript = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getSessionTranscript(session.agent, session.id);
      setTranscript(data);
    } catch {
      // Fallback synthetic transcript from session metadata
      setTranscript({
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
      });
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

  return (
    <div className="space-y-4 pb-12">
      {/* Navigation Breadcrumb & Back Action */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={onBack}
            className="h-8 px-2 text-xs font-mono text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist"
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
            className="h-8 px-2.5 text-xs font-mono border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] rounded-geist"
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
              <h1 className="text-base font-semibold text-[#ededed] tracking-tight">
                {session.title || 'Untitled Session'}
              </h1>
              {session.is_active && (
                <Badge className="bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 text-[10px] font-mono px-2 py-0.5 rounded-[4px]">
                  ● Active Process
                </Badge>
              )}
            </div>

            <p className="text-xs text-[#a1a1a1] leading-relaxed max-w-3xl">
              {session.goal || 'No explicit goal specified for this session.'}
            </p>
          </div>

          {/* Action Toolbar */}
          <div className="flex flex-wrap items-center gap-2 shrink-0">
            {/* Quick Skip Permissions Toggle (for Claude Code) */}
            {session.agent.toLowerCase() === 'claude' && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSkipPermissions(!skipPermissions)}
                className={cn(
                  'h-8 text-xs font-mono rounded-geist border transition-colors',
                  skipPermissions
                    ? 'border-amber-500/50 bg-amber-500/15 text-amber-300 hover:bg-amber-500/20'
                    : 'border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]'
                )}
                title="Bypass confirmation prompts for tools and commands"
              >
                <ShieldAlert className={cn('h-3.5 w-3.5 mr-1.5', skipPermissions ? 'text-amber-400' : 'text-[#666666]')} />
                {skipPermissions ? 'Skip Perms: ON' : 'Skip Perms: OFF'}
              </Button>
            )}

            {/* Quick Resume Button */}
            <Button
              size="sm"
              onClick={handleQuickResume}
              disabled={isResuming}
              className="h-8 px-3 text-xs font-mono font-medium rounded-geist bg-[#ededed] text-black hover:bg-[#ffffff]"
            >
              {isResuming ? (
                <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
              ) : (
                <Terminal className="h-3.5 w-3.5 mr-1.5" />
              )}
              Resume
            </Button>

            {/* Full Flags Modal Trigger */}
            <Button
              variant="outline"
              size="sm"
              onClick={() => setResumeModalOpen(true)}
              className="h-8 px-2.5 text-xs font-mono border-border bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616] rounded-geist"
              title="Configure all AIM and adapter flags"
            >
              Flags...
            </Button>

            {/* Copy Command */}
            <Button
              variant="outline"
              size="sm"
              onClick={copyCliCommand}
              className="h-8 px-2 text-xs font-mono border-border bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] rounded-geist"
              title="Copy terminal command"
            >
              {copiedCmd ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5" />}
            </Button>
          </div>
        </div>

        {/* Metadata Pills */}
        <div className="flex flex-wrap items-center gap-3 pt-3 border-t border-border text-xs text-[#a1a1a1] font-mono">
          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border">
            <span className="text-[#666666]">Profile:</span>
            <span className="text-[#ededed]">{session.profile || 'default'}</span>
          </div>

          <div
            onClick={copyCwd}
            className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border cursor-pointer hover:border-[#333333] transition-colors"
            title="Click to copy path"
          >
            <Folder className="h-3 w-3 text-[#666666]" />
            <span className="truncate max-w-[260px] text-[#ededed]">{session.cwd}</span>
            {copiedCwd ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3 text-[#666666]" />}
          </div>

          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border">
            <MessageSquare className="h-3 w-3 text-[#666666]" />
            <span>{session.turns} turns</span>
          </div>

          <div className="flex items-center gap-1.5 bg-[#0a0a0a] px-2.5 py-1 rounded-[4px] border border-border ml-auto">
            <Clock className="h-3 w-3 text-[#666666]" />
            <span>{new Date(session.updated_at).toLocaleString()}</span>
          </div>
        </div>
      </div>

      {/* Conversation Timeline Replay (Glyphic-style) */}
      <div className="space-y-4">
        <div className="flex items-center justify-between px-1">
          <span className="text-xs font-mono font-semibold uppercase tracking-wider text-[#a1a1a1] flex items-center gap-2">
            <Sparkles className="h-3.5 w-3.5 text-[#ededed]" />
            Conversation Replay
          </span>
          {transcript && (
            <span className="text-[11px] font-mono text-[#666666]">
              {transcript.turns.length} turns displayed
            </span>
          )}
        </div>

        {loading ? (
          <div className="rounded-geist border border-border bg-[#111111] p-12 text-center space-y-3">
            <RefreshCw className="h-6 w-6 animate-spin mx-auto text-[#666666]" />
            <p className="text-xs font-mono text-[#a1a1a1]">Loading conversation transcript from disk...</p>
          </div>
        ) : !transcript || transcript.turns.length === 0 ? (
          <div className="rounded-geist border border-border bg-[#111111] p-8 text-center space-y-2">
            <p className="text-xs font-mono text-[#a1a1a1]">No transcript records found for this session on disk.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {transcript.turns.map((turn) => {
              const isUser = turn.role === 'user';
              const hasThinking = !!turn.thinking?.trim();
              const hasTools = turn.tool_calls && turn.tool_calls.length > 0;
              const isThinkingOpen = !!expandedThinking[turn.turn_index];

              return (
                <div
                  key={turn.turn_index}
                  className={cn(
                    'rounded-geist border p-4 transition-colors',
                    isUser
                      ? 'border-[#262626] bg-[#0c0c0c]'
                      : 'border-border bg-[#111111]'
                  )}
                >
                  {/* Turn Header */}
                  <div className="flex items-center justify-between pb-2 border-b border-[#1c1c1c] text-xs font-mono mb-3">
                    <div className="flex items-center gap-2">
                      {isUser ? (
                        <div className="h-5 w-5 rounded-[4px] bg-[#1a1a1a] border border-[#262626] flex items-center justify-center text-[#ededed]">
                          <User className="h-3 w-3" />
                        </div>
                      ) : (
                        <div className="h-5 w-5 rounded-[4px] bg-[#1a1a1a] border border-[#262626] flex items-center justify-center text-emerald-400">
                          <Bot className="h-3 w-3" />
                        </div>
                      )}
                      <span className="font-semibold text-[#ededed]">
                        {isUser ? 'User Prompt' : `${session.agent} Response`}
                      </span>
                      <span className="text-[10px] text-[#666666]">Turn #{turn.turn_index}</span>
                    </div>

                    {turn.timestamp && (
                      <span className="text-[10px] text-[#666666]">
                        {new Date(turn.timestamp).toLocaleTimeString()}
                      </span>
                    )}
                  </div>

                  {/* Turn Content */}
                  <div className="space-y-3">
                    {/* Collapsible Thinking Block */}
                    {hasThinking && (
                      <div className="rounded-[4px] border border-[#222222] bg-[#080808] overflow-hidden">
                        <button
                          type="button"
                          onClick={() => toggleThinking(turn.turn_index)}
                          className="w-full flex items-center justify-between px-3 py-2 text-[11px] font-mono text-[#888888] hover:text-[#ededed] hover:bg-[#0f0f0f] transition-colors"
                        >
                          <span className="flex items-center gap-1.5">
                            {isThinkingOpen ? (
                              <ChevronDown className="h-3 w-3 text-[#666666]" />
                            ) : (
                              <ChevronRight className="h-3 w-3 text-[#666666]" />
                            )}
                            Reasoning &amp; Thinking Process
                          </span>
                          <span className="text-[10px] text-[#555555]">
                            {isThinkingOpen ? 'collapse' : 'expand'}
                          </span>
                        </button>
                        {isThinkingOpen && (
                          <div className="p-3 border-t border-[#1c1c1c] text-xs font-mono text-[#888888] whitespace-pre-wrap leading-relaxed max-h-96 overflow-y-auto">
                            {turn.thinking}
                          </div>
                        )}
                      </div>
                    )}

                    {/* Tool Calls Execution Cards */}
                    {hasTools && (
                      <div className="space-y-2">
                        {turn.tool_calls!.map((tool, idx) => {
                          const toolKey = `${turn.turn_index}-${idx}-${tool.tool_name}`;
                          const isToolOpen = expandedTools[toolKey] !== false; // Default open

                          return (
                            <div
                              key={toolKey}
                              className="rounded-[4px] border border-[#222222] bg-[#0a0a0a] overflow-hidden"
                            >
                              <div
                                onClick={() => toggleTool(toolKey)}
                                className="flex items-center justify-between px-3 py-2 text-xs font-mono cursor-pointer hover:bg-[#121212] transition-colors select-none"
                              >
                                <div className="flex items-center gap-2 min-w-0">
                                  <Wrench className="h-3.5 w-3.5 text-blue-400 shrink-0" />
                                  <span className="text-[#ededed] font-semibold">{tool.tool_name}</span>
                                </div>
                                <div className="flex items-center gap-2">
                                  <Badge
                                    variant="outline"
                                    className={cn(
                                      'text-[10px] font-mono px-1.5 py-0.5 rounded-[3px] border',
                                      tool.status === 'error'
                                        ? 'border-red-500/40 text-red-400 bg-red-500/10'
                                        : 'border-emerald-500/40 text-emerald-400 bg-emerald-500/10'
                                    )}
                                  >
                                    {tool.status || 'ok'}
                                  </Badge>
                                  {isToolOpen ? (
                                    <ChevronDown className="h-3.5 w-3.5 text-[#666666]" />
                                  ) : (
                                    <ChevronRight className="h-3.5 w-3.5 text-[#666666]" />
                                  )}
                                </div>
                              </div>

                              {isToolOpen && (
                                <div className="p-3 border-t border-[#1c1c1c] space-y-2 text-xs font-mono bg-[#050505]">
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
                                      <pre className="p-2 rounded-[3px] bg-[#0c0c0c] border border-[#1f1f1f] text-[11px] text-[#a1a1a1] overflow-x-auto whitespace-pre-wrap max-h-60">
                                        {tool.output}
                                      </pre>
                                    </div>
                                  )}
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    )}

                    {/* Main Content Body */}
                    {turn.content && (
                      <div className="text-xs text-[#ededed] whitespace-pre-wrap font-sans leading-relaxed selection:bg-[#333333]">
                        {turn.content}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Resume Modal */}
      <ResumeModal
        open={resumeModalOpen}
        onOpenChange={setResumeModalOpen}
        session={session}
        onLaunch={handleModalLaunch}
        isResuming={isResuming}
      />
    </div>
  );
};
