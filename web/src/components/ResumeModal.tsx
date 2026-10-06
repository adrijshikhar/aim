import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from './ui/dialog';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Badge } from './ui/badge';
import { Checkbox } from './ui/checkbox';
import { Terminal, Copy, Check, RefreshCw, Folder, MessageSquare, ShieldAlert, AlertTriangle } from 'lucide-react';
import { SessionDTO, ResumeRequest, AdapterDTO, getAdapters, RunnableFlagDTO } from '../lib/api';

interface ResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  session: SessionDTO | null;
  onLaunch: (req: ResumeRequest) => Promise<void>;
  isResuming?: boolean;
  adapters?: AdapterDTO[];
  initialFork?: boolean;
  forkTurnIndex?: number | null;
}

interface AimFlagOption {
  key: string;
  flag: string;
  name: string;
  description: string;
}

const AIM_DIRECT_FLAGS: AimFlagOption[] = [
  {
    key: 'exact',
    flag: '--exact',
    name: 'Verbatim Conversation Thread',
    description: 'Resumes native full conversation history without summary truncation.',
  },
  {
    key: 'catalyst',
    flag: '--catalyst',
    name: 'Catalyst Handoff (-c)',
    description: 'Resume via Catalyst checkpoint brief for clean regrounded context.',
  },
  {
    key: 'fork',
    flag: '--fork',
    name: 'Fork Session (-b)',
    description: 'Branch the session into a new conversation ID, preserving original untouched.',
  },
  {
    key: 'force',
    flag: '--force',
    name: 'Force Resume (-f)',
    description: 'Force resume even if marked currently active in another terminal process.',
  },
];

export const ResumeModal: React.FC<ResumeModalProps> = ({
  open,
  onOpenChange,
  session,
  onLaunch,
  isResuming = false,
  adapters: initialAdapters,
  initialFork = false,
  forkTurnIndex = null,
}) => {
  const [adapters, setAdapters] = React.useState<AdapterDTO[]>(initialAdapters || []);
  const [selectedAimFlags, setSelectedAimFlags] = React.useState<Record<string, boolean>>({
    exact: false,
    catalyst: false,
    fork: initialFork,
    force: false,
  });
  const [selectedAdapterFlags, setSelectedAdapterFlags] = React.useState<Record<string, boolean>>({});
  const [customArgs, setCustomArgs] = React.useState('');
  const [copied, setCopied] = React.useState(false);

  React.useEffect(() => {
    if (open) {
      if (!initialAdapters || initialAdapters.length === 0) {
        getAdapters().then(setAdapters).catch(() => {});
      }
      setSelectedAimFlags({
        exact: false,
        catalyst: false,
        fork: !!initialFork,
        force: !!session?.is_active,
      });
      // Default to empty adapter flags, user can toggle --dangerously-skip-permissions
      setSelectedAdapterFlags({});
      setCustomArgs('');
      setCopied(false);
    }
  }, [open, session, initialAdapters, initialFork]);

  if (!session) return null;

  const currentAdapter = adapters.find(
    (a) => a.name.toLowerCase() === session.agent.toLowerCase() || a.aliases?.includes(session.agent.toLowerCase())
  );

  const runnableFlags: RunnableFlagDTO[] = currentAdapter?.runnable_flags || [
    ...(session.agent.toLowerCase() === 'claude'
      ? [
          {
            flag: '--dangerously-skip-permissions',
            name: 'Dangerously Skip Permissions',
            description: 'Bypass all confirmation prompts for tool execution, bash commands, and file edits.',
            type: 'bool' as const,
            dangerous: true,
          },
          {
            flag: '--verbose',
            name: 'Verbose Logging',
            description: 'Enable detailed engine diagnostics and API traces.',
            type: 'bool' as const,
            dangerous: false,
          },
        ]
      : []),
  ];

  const toggleAimFlag = (key: string) => {
    setSelectedAimFlags((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const toggleAdapterFlag = (flag: string) => {
    setSelectedAdapterFlags((prev) => ({
      ...prev,
      [flag]: !prev[flag],
    }));
  };

  const getActiveAimFlagsList = (): string[] => {
    const list: string[] = [];
    AIM_DIRECT_FLAGS.forEach((f) => {
      if (selectedAimFlags[f.key]) {
        list.push(f.flag);
      }
    });
    return list;
  };

  const getActiveAdapterFlagsList = (): string[] => {
    const list: string[] = [];
    Object.entries(selectedAdapterFlags).forEach(([flag, active]) => {
      if (active) {
        list.push(flag);
      }
    });
    return list;
  };

  const buildCommandString = () => {
    const agent = session.agent || 'agent';
    const profile = session.profile && session.profile !== '<host>' ? session.profile : '';
    const sid = session.id;

    const aimFlags = getActiveAimFlagsList();
    const adapterFlags = getActiveAdapterFlagsList();
    const custom = customArgs.trim();

    const parts = ['aim', 'resume'];
    if (aimFlags.length > 0) {
      parts.push(...aimFlags);
    }
    parts.push(agent);
    if (profile) {
      parts.push(profile);
    }
    parts.push(sid);

    const allAdapterArgs = [...adapterFlags];
    if (custom) {
      allAdapterArgs.push(custom);
    }

    if (allAdapterArgs.length > 0) {
      parts.push('--');
      parts.push(...allAdapterArgs);
    }

    return parts.join(' ');
  };

  const handleCopyCommand = () => {
    const cmd = buildCommandString();
    navigator.clipboard.writeText(cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onLaunch({
      agent: session.agent,
      profile: session.profile,
      session_id: session.id,
      aim_flags: getActiveAimFlagsList(),
      adapter_flags: getActiveAdapterFlagsList(),
      custom_args: customArgs.trim() || undefined,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg bg-[#111111] border-[#262626] text-[#ededed] p-5 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1 pb-2 border-b border-[#262626] pr-8">
          <div className="flex items-center gap-2 flex-wrap">
            <DialogTitle className="text-sm font-semibold text-[#ededed] flex items-center gap-2">
              <Terminal className="h-4 w-4 text-[#ededed]" />
              <span>Resume Session with Flags</span>
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[10px] font-mono bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] px-1.5 py-0.5 rounded-geist"
            >
              {session.agent} / {session.profile || 'default'}
            </Badge>
            {forkTurnIndex !== null && forkTurnIndex !== undefined && (
              <Badge
                variant="outline"
                className="text-[10px] font-mono bg-amber-500/10 text-amber-400 border-amber-500/30 px-1.5 py-0.5 rounded-geist"
              >
                Fork at Turn #{forkTurnIndex}
              </Badge>
            )}
          </div>
          <DialogDescription className="text-xs text-[#a1a1a1]">
            Select orchestrator switches and native {session.agent} CLI flags before launching terminal.
          </DialogDescription>
        </DialogHeader>

        {/* Compact Session Metadata Strip */}
        <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-xs flex items-center justify-between text-[#888888]">
          <div className="flex items-center gap-2 min-w-0">
            <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
            <span className="font-mono text-[#ededed] truncate max-w-[240px]" title={session.cwd}>
              {session.cwd}
            </span>
          </div>
          <span className="flex items-center gap-1 font-mono text-[11px] text-[#a1a1a1] shrink-0">
            <MessageSquare className="h-3 w-3" />
            {session.turns} turns
          </span>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3 pt-0.5">
          {/* Section 1: AIM Direct Flags */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold text-[#ededed] uppercase tracking-wider font-mono">
                1. AIM Direct Flags (Orchestrator)
              </span>
              <span className="text-[10px] text-[#666666] font-mono">Profile &amp; Context</span>
            </div>

            <div className="grid grid-cols-2 gap-1.5">
              {AIM_DIRECT_FLAGS.map((f) => {
                const isChecked = selectedAimFlags[f.key];
                const isForkingTurn = f.key === 'fork' && forkTurnIndex != null;
                return (
                  <label
                    key={f.key}
                    htmlFor={`aim-flag-${f.key}`}
                    className={`flex items-start gap-2 p-2 rounded-geist border cursor-pointer transition-colors ${
                      isChecked
                        ? 'border-[#0070f3]/50 bg-[#0070f3]/10'
                        : 'border-[#262626] bg-[#0a0a0a] hover:border-[#333333] hover:bg-[#121212]'
                    }`}
                  >
                    <div className="pt-0.5 shrink-0">
                      <Checkbox
                        id={`aim-flag-${f.key}`}
                        checked={isChecked}
                        onCheckedChange={() => toggleAimFlag(f.key)}
                      />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-1">
                        <span className="font-mono text-[11px] font-bold text-[#ededed] block">{f.flag}</span>
                        {isForkingTurn && (
                          <Badge variant="outline" className="text-[9px] font-mono px-1 py-0 border-amber-500/30 text-amber-300 bg-amber-500/10">
                            Turn #{forkTurnIndex}
                          </Badge>
                        )}
                      </div>
                      <p className="text-[10px] text-[#888888] leading-tight truncate">{f.name}</p>
                    </div>
                  </label>
                );
              })}
            </div>
          </div>

          {/* Section 2: Adapter Runnable Flags */}
          {runnableFlags.length > 0 && (
            <div className="space-y-1.5 pt-1 border-t border-[#262626]">
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-semibold text-[#ededed] uppercase tracking-wider font-mono">
                  2. {session.agent} Flags (Passthrough via --)
                </span>
                <span className="text-[10px] text-[#666666] font-mono">Native CLI</span>
              </div>

              <div className="space-y-1.5">
                {runnableFlags.map((rf) => {
                  const isChecked = !!selectedAdapterFlags[rf.flag];
                  return (
                    <label
                      key={rf.flag}
                      htmlFor={`adapter-flag-${rf.flag}`}
                      className={`flex items-start gap-2.5 p-2 rounded-geist border cursor-pointer transition-colors ${
                        isChecked
                          ? rf.dangerous
                            ? 'border-amber-500/50 bg-amber-500/10'
                            : 'border-[#0070f3]/50 bg-[#0070f3]/10'
                          : 'border-[#262626] bg-[#0a0a0a] hover:border-[#333333] hover:bg-[#121212]'
                      }`}
                    >
                      <div className="pt-0.5 shrink-0">
                        <Checkbox
                          id={`adapter-flag-${rf.flag}`}
                          checked={isChecked}
                          onCheckedChange={() => toggleAdapterFlag(rf.flag)}
                        />
                      </div>
                      <div className="min-w-0 flex-1 space-y-0.5">
                        <div className="flex items-center justify-between gap-1">
                          <span className="font-mono text-[11px] font-bold text-[#ededed]">{rf.flag}</span>
                          {rf.dangerous && (
                            <Badge
                              variant="outline"
                              className="text-[9px] font-mono border-amber-500/30 text-amber-300 bg-amber-500/10 px-1 py-0 flex items-center gap-1"
                            >
                              <ShieldAlert className="h-2.5 w-2.5 text-amber-400" />
                              Skip Prompts
                            </Badge>
                          )}
                        </div>
                        <p className="text-[10px] text-[#888888] leading-tight truncate">{rf.description}</p>
                      </div>
                    </label>
                  );
                })}
              </div>
            </div>
          )}

          {/* Auto-Approve Autonomous Risk Warning Banner */}
          {selectedAdapterFlags['--dangerously-skip-permissions'] && (
            <div className="rounded-geist border border-amber-500/40 bg-amber-500/10 p-2.5 flex items-start gap-2 text-xs">
              <AlertTriangle className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
              <div className="space-y-0.5">
                <p className="font-medium text-amber-200 text-[11px]">Autonomous Tool Execution Active</p>
                <p className="text-[10px] text-amber-300/80 leading-normal">
                  Tool execution, bash commands, and file edits will run automatically without interactive confirmation prompts.
                </p>
              </div>
            </div>
          )}

          {/* Section 3: Custom Adapter Arguments */}
          <div className="space-y-1 pt-1 border-t border-[#262626]">
            <div className="flex items-center justify-between text-[11px] text-[#a1a1a1]">
              <label htmlFor="custom-args" className="font-sans cursor-pointer">Additional Flags:</label>
              <span className="text-[10px] text-[#666666] font-mono">Appended after --</span>
            </div>
            <Input
              id="custom-args"
              placeholder="e.g. --model claude-3-5-sonnet"
              className="h-7 text-xs font-mono rounded-geist border-[#262626] bg-[#0a0a0a] text-[#ededed] placeholder:text-[#666666] focus:border-[#333333]"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
            />
          </div>

          {/* Live Command Preview Box */}
          <div className="space-y-1 pt-1">
            <span className="text-[10px] font-mono text-[#888888]">Terminal Command:</span>
            <div className="relative rounded-geist border border-[#262626] bg-[#0a0a0a] px-2.5 py-1.5 font-mono text-[11px] text-[#ededed] flex items-center justify-between gap-2">
              <div className="select-all text-[11px] text-[#ededed] overflow-x-auto whitespace-pre flex-1 py-0.5">
                {buildCommandString()}
              </div>
              <button
                type="button"
                onClick={handleCopyCommand}
                aria-label="Copy terminal command"
                className="shrink-0 p-1 rounded-geist hover:bg-[#1f1f1f] text-[#a1a1a1] hover:text-[#ededed] transition-colors cursor-pointer"
                title="Copy command to clipboard"
              >
                {copied ? <Check className="h-3 w-3 text-emerald-400" /> : <Copy className="h-3 w-3" />}
              </button>
            </div>
          </div>

          <DialogFooter className="pt-2 flex items-center justify-end gap-2 border-t border-[#262626]">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="rounded-geist h-7 px-3 text-xs font-sans border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isResuming}
              className="rounded-geist h-7 px-3.5 text-xs font-sans bg-[#ffffff] text-black hover:bg-[#eaeaea] font-medium shadow-sm transition-colors cursor-pointer"
            >
              {isResuming ? (
                <>
                  <RefreshCw className="h-3 w-3 mr-1 animate-spin" />
                  Launching...
                </>
              ) : (
                <>
                  <Terminal className="h-3 w-3 mr-1" />
                  Resume in Terminal
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
