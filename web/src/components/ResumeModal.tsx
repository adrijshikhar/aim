import React, { useState, useEffect } from 'react';
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
import { Terminal, Copy, Check, RefreshCw, Folder, MessageSquare } from 'lucide-react';
import { SessionDTO, ResumeRequest } from '../lib/api';

interface ResumeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  session: SessionDTO | null;
  onLaunch: (req: ResumeRequest) => Promise<void>;
  isResuming?: boolean;
}

interface FlagOption {
  key: string;
  flag: string;
  name: string;
  description: string;
}

const AVAILABLE_FLAGS: FlagOption[] = [
  {
    key: 'exact',
    flag: '--exact',
    name: 'Verbatim Conversation Thread',
    description: 'Resumes the native full conversation history thread without summary truncation.',
  },
  {
    key: 'catalyst',
    flag: '--catalyst',
    name: 'Catalyst Handoff',
    description: 'Resume via Catalyst checkpoint brief (-c) for clean regrounded context.',
  },
  {
    key: 'fork',
    flag: '--fork',
    name: 'Fork Session',
    description: 'Branch the session into a new conversation ID (-b), keeping the original untouched.',
  },
  {
    key: 'force',
    flag: '--force',
    name: 'Force Resume',
    description: 'Force resume (-f) even if the session is marked currently active.',
  },
];

export const ResumeModal: React.FC<ResumeModalProps> = ({
  open,
  onOpenChange,
  session,
  onLaunch,
  isResuming = false,
}) => {
  const [selectedFlags, setSelectedFlags] = useState<Record<string, boolean>>({
    exact: false,
    catalyst: false,
    fork: false,
    force: false,
  });
  const [customFlags, setCustomFlags] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (open) {
      setSelectedFlags({
        exact: false,
        catalyst: false,
        fork: false,
        force: !!session?.is_active,
      });
      setCustomFlags('');
      setCopied(false);
    }
  }, [open, session]);

  if (!session) return null;

  const toggleFlag = (key: string) => {
    setSelectedFlags((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const getActiveFlagsList = (): string[] => {
    const list: string[] = [];
    AVAILABLE_FLAGS.forEach((f) => {
      if (selectedFlags[f.key]) {
        list.push(f.flag);
      }
    });
    return list;
  };

  const buildCommandString = () => {
    const agent = session.agent || 'agent';
    const profile = session.profile && session.profile !== '<host>' ? session.profile : '';
    const sid = session.id;

    let cmd = profile ? `aim resume ${agent} ${profile} ${sid}` : `aim resume ${agent} ${sid}`;
    const flags = getActiveFlagsList();
    if (flags.length > 0) {
      cmd += ' ' + flags.join(' ');
    }
    const custom = customFlags.trim();
    if (custom) {
      cmd += ' ' + custom;
    }
    return cmd;
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
      flags: getActiveFlagsList(),
      custom_flags: customFlags.trim() || undefined,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl bg-[#0c0c0c] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#1f1f1f] pr-8">
          <div className="flex items-center gap-2.5 flex-wrap">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <Terminal className="h-4 w-4 text-[#ededed]" />
              <span>Resume Session with Flags</span>
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[10px] font-mono bg-[#141414] text-[#888888] border-[#262626] px-2 py-0.5"
            >
              {session.agent} / {session.profile || 'default'}
            </Badge>
          </div>
          <DialogDescription className="text-xs text-[#888888]">
            Configure execution flags and launch options before opening an interactive terminal.
          </DialogDescription>
        </DialogHeader>

        {/* Session Metadata Recap */}
        <div className="rounded-md border border-[#222222] bg-[#121212] p-3 text-xs space-y-1.5">
          <div className="flex items-center justify-between text-[#888888]">
            <span className="font-mono text-[11px] truncate max-w-[280px]" title={session.id}>
              ID: <span className="text-[#ededed]">{session.id}</span>
            </span>
            <span className="flex items-center gap-1 font-mono text-[11px] text-[#888888]">
              <MessageSquare className="h-3 w-3" />
              {session.turns} turns
            </span>
          </div>
          <div className="flex items-center gap-1.5 text-[#888888] font-mono text-[11px] truncate">
            <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
            <span className="truncate" title={session.cwd}>{session.cwd}</span>
          </div>
          {session.goal && (
            <div className="text-[11px] text-[#a1a1a1] italic truncate pt-0.5 border-t border-[#1c1c1c]">
              "{session.goal}"
            </div>
          )}
        </div>

        {/* Active Process Warning */}
        {session.is_active && (
          <div className="rounded-md border border-amber-500/30 bg-[#141414] p-3 text-xs space-y-1">
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
              </span>
              <span className="text-[#ededed] font-medium font-mono text-xs">
                Active Session Running {session.pid ? `(PID ${session.pid})` : ''}
              </span>
            </div>
            <p className="text-[11px] text-[#888888] pl-4 leading-relaxed">
              This session is running in another process. <strong className="text-[#ededed] font-mono">--force</strong> is pre-selected to resume directly without CLI prompt interruptions. Or select <strong className="text-[#ededed] font-mono">--fork</strong> to branch safely into a new conversation.
            </p>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4 pt-1">
          {/* Quick Flags Toggles */}
          <div className="space-y-2">
            <label className="text-[11px] uppercase tracking-wider font-mono text-[#888888]">
              Execution Flags
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {AVAILABLE_FLAGS.map((f) => {
                const active = !!selectedFlags[f.key];
                return (
                  <button
                    key={f.key}
                    type="button"
                    onClick={() => toggleFlag(f.key)}
                    className={`flex flex-col text-left p-2.5 rounded-md border transition-all cursor-pointer select-none ${
                      active
                        ? 'bg-[#1a1a1a] border-[#ededed] text-[#ededed]'
                        : 'bg-[#121212] border-[#222222] hover:border-[#333333] text-[#888888]'
                    }`}
                  >
                    <div className="flex items-center justify-between w-full mb-1">
                      <span className="font-mono text-xs font-semibold text-[#ededed]">
                        {f.flag}
                      </span>
                      <div
                        className={`w-3.5 h-3.5 rounded border flex items-center justify-center transition-colors ${
                          active
                            ? 'bg-[#ededed] border-[#ededed] text-[#0a0a0a]'
                            : 'border-[#3a3a3a] bg-[#1a1a1a]'
                        }`}
                      >
                        {active && <Check className="w-2.5 h-2.5 stroke-[3]" />}
                      </div>
                    </div>
                    <span className="text-[11px] font-medium text-[#cccccc] mb-0.5">{f.name}</span>
                    <span className="text-[10px] text-[#777777] leading-tight">{f.description}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Custom Flags Input */}
          <div className="space-y-1.5">
            <label className="text-[11px] uppercase tracking-wider font-mono text-[#888888]">
              Additional Flags / Arguments
            </label>
            <Input
              value={customFlags}
              onChange={(e) => setCustomFlags(e.target.value)}
              placeholder="e.g. --dangerously-skip-permissions --verbose"
              className="bg-[#121212] border-[#262626] font-mono text-xs text-[#ededed] placeholder:text-[#555555] focus-visible:ring-1 focus-visible:ring-[#ededed] focus-visible:border-[#ededed]"
            />
          </div>

          {/* Live Command Preview Box */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="text-[10.5px] uppercase tracking-wider font-mono text-[#888888]">
                Command Preview
              </span>
              <button
                type="button"
                onClick={handleCopyCommand}
                className="inline-flex items-center gap-1 text-[10.5px] font-mono text-[#888888] hover:text-[#ededed] transition-colors cursor-pointer"
              >
                {copied ? <Check className="h-3 w-3 text-[#ededed]" /> : <Copy className="h-3 w-3" />}
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
            <div className="p-2.5 rounded-md bg-[#000000] border border-[#222222] font-mono text-[11.5px] text-[#ededed] overflow-x-auto select-all whitespace-pre">
              {buildCommandString()}
            </div>
          </div>

          <DialogFooter className="pt-2 border-t border-[#1f1f1f] flex items-center justify-between sm:justify-between w-full">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="border-[#262626] bg-[#141414] hover:bg-[#1a1a1a] text-[#888888] hover:text-[#ededed] text-xs font-mono"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isResuming}
              className="bg-[#ededed] hover:bg-white text-[#0a0a0a] text-xs font-mono font-medium shadow-sm transition-all cursor-pointer"
            >
              {isResuming ? (
                <>
                  <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                  Launching...
                </>
              ) : (
                <>
                  <Terminal className="h-3.5 w-3.5 mr-1.5" />
                  Launch in Terminal
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
