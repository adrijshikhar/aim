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
import { Terminal, Copy, Check, RefreshCw, Folder, MessageSquare, ShieldAlert } from 'lucide-react';
import { SessionDTO, ResumeRequest, AdapterDTO, getAdapters, RunnableFlagDTO } from '../lib/api';

interface ResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  session: SessionDTO | null;
  onLaunch: (req: ResumeRequest) => Promise<void>;
  isResuming?: boolean;
  adapters?: AdapterDTO[];
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
}) => {
  const [adapters, setAdapters] = React.useState<AdapterDTO[]>(initialAdapters || []);
  const [selectedAimFlags, setSelectedAimFlags] = React.useState<Record<string, boolean>>({
    exact: false,
    catalyst: false,
    fork: false,
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
        fork: false,
        force: !!session?.is_active,
      });
      // Default to empty adapter flags, user can toggle --dangerously-skip-permissions
      setSelectedAdapterFlags({});
      setCustomArgs('');
      setCopied(false);
    }
  }, [open, session, initialAdapters]);

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
      <DialogContent className="max-w-xl bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist max-h-[90vh] overflow-y-auto">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#262626] pr-8">
          <div className="flex items-center gap-2.5 flex-wrap">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <Terminal className="h-4 w-4 text-[#ededed]" />
              <span>Resume Session with Flags</span>
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[10px] font-mono bg-[#0a0a0a] text-[#a1a1a1] border-[#262626] px-2 py-0.5 rounded-geist"
            >
              {session.agent} / {session.profile || 'default'}
            </Badge>
          </div>
          <DialogDescription className="text-xs text-[#a1a1a1]">
            Configure orchestrator switches and native {session.agent} CLI flags before launching terminal.
          </DialogDescription>
        </DialogHeader>

        {/* Session Metadata Recap */}
        <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-1.5">
          <div className="flex items-center justify-between text-[#a1a1a1]">
            <span className="font-mono text-[11px] truncate max-w-[280px]" title={session.id}>
              ID: <span className="text-[#ededed]">{session.id}</span>
            </span>
            <span className="flex items-center gap-1 font-mono text-[11px] text-[#a1a1a1]">
              <MessageSquare className="h-3 w-3" />
              {session.turns} turns
            </span>
          </div>
          <div className="flex items-center gap-1.5 text-[#a1a1a1] font-mono text-[11px] truncate">
            <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
            <span className="truncate" title={session.cwd}>{session.cwd}</span>
          </div>
          {session.goal && (
            <div className="text-[11px] text-[#a1a1a1] italic truncate pt-1 border-t border-[#262626]">
              &ldquo;{session.goal}&rdquo;
            </div>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 pt-1">
          {/* Section 1: AIM Direct Flags */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-[#ededed] uppercase tracking-wider font-mono">
                1. AIM Direct Flags (Orchestrator)
              </span>
              <span className="text-[10px] text-[#666666] font-mono">Profile &amp; Context</span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {AIM_DIRECT_FLAGS.map((f) => {
                const isChecked = selectedAimFlags[f.key];
                return (
                  <div
                    key={f.key}
                    onClick={() => toggleAimFlag(f.key)}
                    className={`flex items-start gap-2.5 p-2.5 rounded-geist border cursor-pointer transition-colors ${
                      isChecked
                        ? 'border-[#444444] bg-[#161616]'
                        : 'border-[#262626] bg-[#0a0a0a] hover:border-[#333333] hover:bg-[#121212]'
                    }`}
                  >
                    <div className="pt-0.5">
                      <div
                        className={`h-3.5 w-3.5 rounded-[3px] border flex items-center justify-center transition-colors ${
                          isChecked
                            ? 'bg-[#ededed] border-[#ededed] text-[#000000]'
                            : 'border-[#444444] bg-[#0a0a0a]'
                        }`}
                      >
                        {isChecked && <Check className="h-2.5 w-2.5 stroke-[3]" />}
                      </div>
                    </div>
                    <div className="min-w-0 flex-1 space-y-0.5">
                      <div className="flex items-center justify-between gap-1">
                        <span className="font-mono text-[11px] font-bold text-[#ededed]">{f.flag}</span>
                      </div>
                      <p className="text-[10px] text-[#888888] leading-tight line-clamp-2">{f.description}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Section 2: Adapter Runnable Flags */}
          {runnableFlags.length > 0 && (
            <div className="space-y-2 pt-1 border-t border-[#262626]">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-[#ededed] uppercase tracking-wider font-mono">
                  2. {session.agent} Runnable Flags (Passthrough via --)
                </span>
                <span className="text-[10px] text-[#666666] font-mono">Agent CLI Flags</span>
              </div>

              <div className="space-y-2">
                {runnableFlags.map((rf) => {
                  const isChecked = !!selectedAdapterFlags[rf.flag];
                  return (
                    <div
                      key={rf.flag}
                      onClick={() => toggleAdapterFlag(rf.flag)}
                      className={`flex items-start gap-3 p-3 rounded-geist border cursor-pointer transition-colors ${
                        isChecked
                          ? rf.dangerous
                            ? 'border-amber-500/50 bg-amber-500/10'
                            : 'border-[#444444] bg-[#161616]'
                          : 'border-[#262626] bg-[#0a0a0a] hover:border-[#333333] hover:bg-[#121212]'
                      }`}
                    >
                      <div className="pt-0.5">
                        <div
                          className={`h-4 w-4 rounded-[3px] border flex items-center justify-center transition-colors ${
                            isChecked
                              ? rf.dangerous
                                ? 'bg-amber-400 border-amber-400 text-black'
                                : 'bg-[#ededed] border-[#ededed] text-[#000000]'
                              : 'border-[#444444] bg-[#0a0a0a]'
                          }`}
                        >
                          {isChecked && <Check className="h-3 w-3 stroke-[3]" />}
                        </div>
                      </div>
                      <div className="min-w-0 flex-1 space-y-1">
                        <div className="flex items-center justify-between gap-2">
                          <span className="font-mono text-xs font-bold text-[#ededed]">{rf.flag}</span>
                          {rf.dangerous && (
                            <Badge
                              variant="outline"
                              className="text-[10px] font-mono border-amber-500/40 text-amber-300 bg-amber-500/10 px-1.5 py-0.5 flex items-center gap-1"
                            >
                              <ShieldAlert className="h-3 w-3 text-amber-400" />
                              High Velocity / Skip Prompts
                            </Badge>
                          )}
                        </div>
                        <p className="text-[11px] text-[#a1a1a1] leading-tight">{rf.description}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Section 3: Custom Adapter Arguments */}
          <div className="space-y-1.5 pt-1 border-t border-[#262626]">
            <label htmlFor="custom-args" className="text-xs font-mono text-[#a1a1a1] flex items-center justify-between">
              <span>Additional Adapter Arguments:</span>
              <span className="text-[10px] text-[#666666]">Appended after --</span>
            </label>
            <Input
              id="custom-args"
              placeholder="e.g. --model claude-3-5-sonnet"
              className="h-8 text-xs font-mono rounded-geist border-[#262626] bg-[#0a0a0a] text-[#ededed] placeholder:text-[#666666] focus:border-[#333333]"
              value={customArgs}
              onChange={(e) => setCustomArgs(e.target.value)}
            />
          </div>

          {/* Live Command Preview Box */}
          <div className="space-y-1.5 pt-1">
            <span className="text-[11px] font-mono text-[#a1a1a1]">Terminal Launch Command:</span>
            <div className="relative rounded-geist border border-[#262626] bg-[#0a0a0a] p-2.5 font-mono text-xs text-[#ededed] flex items-center justify-between gap-2 overflow-x-auto">
              <span className="select-all text-[11px] text-[#ededed]">{buildCommandString()}</span>
              <button
                type="button"
                onClick={handleCopyCommand}
                className="shrink-0 p-1 rounded-geist hover:bg-[#1f1f1f] text-[#a1a1a1] hover:text-[#ededed] transition-colors"
                title="Copy command to clipboard"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-[#0070f3]" /> : <Copy className="h-3.5 w-3.5" />}
              </button>
            </div>
          </div>

          <DialogFooter className="pt-2 flex items-center justify-end gap-2 border-t border-[#262626]">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="rounded-geist h-8 border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isResuming}
              className="rounded-geist h-8 bg-[#ededed] text-[#000000] hover:bg-[#ffffff] font-medium"
            >
              {isResuming ? (
                <>
                  <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                  Launching...
                </>
              ) : (
                <>
                  <Terminal className="h-3.5 w-3.5 mr-1.5" />
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
