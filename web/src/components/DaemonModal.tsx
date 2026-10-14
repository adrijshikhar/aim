import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from './ui/dialog';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import {
  Activity,
  Play,
  Trash2,
  RefreshCw,
  Clock,
  FileText,
  Check,
  Copy,
  Folder,
  Terminal,
} from 'lucide-react';
import { DaemonDTO, installDaemon, uninstallDaemon, runDaemonOnce, getDaemonLogs } from '../lib/api';
import { cn } from '../lib/utils';
import { useToast } from './ui/use-toast';

interface DaemonModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  daemonInfo: DaemonDTO | null;
  onRefreshDaemon: () => Promise<void>;
}

interface ParsedLogLine {
  timestamp: string;
  level: 'INFO' | 'WARN' | 'ERROR';
  message: string;
}

function parseLogLine(rawLine: string): ParsedLogLine {
  const line = rawLine.trim();

  // Pattern 1: Standard AIM format: [2026-10-14 01:05:00 UTC] [INFO] message
  const standardMatch = line.match(/^\[(.*?)\]\s*\[(INFO|WARN|WARNING|ERROR|FATAL|DEBUG)\]\s*(.*)$/i);
  if (standardMatch) {
    const [, ts, lvl, msg] = standardMatch;
    const upper = lvl.toUpperCase();
    const level: 'INFO' | 'WARN' | 'ERROR' = upper.startsWith('WARN')
      ? 'WARN'
      : upper === 'ERROR' || upper === 'FATAL'
      ? 'ERROR'
      : 'INFO';
    return {
      timestamp: ts.endsWith('UTC') ? ts : `${ts} UTC`,
      level,
      message: msg,
    };
  }

  // Pattern 2: Legacy AIM format: [2026-10-14 01:05:00 UTC] message
  const legacyMatch = line.match(/^\[(.*?)\]\s*(.*)$/);
  if (legacyMatch) {
    const [, ts, msg] = legacyMatch;
    const lower = msg.toLowerCase();
    let level: 'INFO' | 'WARN' | 'ERROR' = 'INFO';
    if ((!lower.includes('0 errors') && lower.includes('error')) || lower.includes('fatal') || lower.includes('failed') || lower.includes('panic')) {
      level = 'ERROR';
    } else if (lower.includes('warn') || lower.includes('0 configured') || lower.includes('timeout') || lower.includes('deadline exceeded')) {
      level = 'WARN';
    }
    return {
      timestamp: ts.endsWith('UTC') ? ts : `${ts} UTC`,
      level,
      message: msg,
    };
  }

  // Pattern 3: Go standard log format: 2026/10/13 14:53:31 message
  const goLogMatch = line.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s*(.*)$/);
  if (goLogMatch) {
    const [, rawTs, msg] = goLogMatch;
    const ts = rawTs.replace(/\//g, '-') + ' UTC';
    const lower = msg.toLowerCase();
    let level: 'INFO' | 'WARN' | 'ERROR' = 'INFO';
    if (lower.includes('error') || lower.includes('fatal') || lower.includes('failed')) {
      level = 'ERROR';
    } else if (lower.includes('warn') || lower.includes('deadline exceeded') || lower.includes('timeout')) {
      level = 'WARN';
    }
    return {
      timestamp: ts,
      level,
      message: msg,
    };
  }

  // Fallback: No timestamp in line
  const lower = line.toLowerCase();
  let level: 'INFO' | 'WARN' | 'ERROR' = 'INFO';
  if ((!lower.includes('0 errors') && lower.includes('error')) || lower.includes('failed')) {
    level = 'ERROR';
  } else if (lower.includes('warn') || lower.includes('timeout') || lower.includes('deadline exceeded')) {
    level = 'WARN';
  }
  return {
    timestamp: '',
    level,
    message: line,
  };
}

const PAGE_SIZE = 200;

export const DaemonModal: React.FC<DaemonModalProps> = ({
  open,
  onOpenChange,
  daemonInfo,
  onRefreshDaemon,
}) => {
  const [running, setRunning] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [uninstalling, setUninstalling] = useState(false);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [copiedLogs, setCopiedLogs] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);
  const [totalLogLines, setTotalLogLines] = useState(0);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [loadingMoreLogs, setLoadingMoreLogs] = useState(false);
  const [showLogs, setShowLogs] = useState(false);
  const terminalRef = useRef<HTMLDivElement>(null);
  const isInitialLoadRef = useRef(true);
  const { toast } = useToast();

  const fetchLogs = useCallback(async () => {
    setLoadingLogs(true);
    isInitialLoadRef.current = true;
    try {
      const data = await getDaemonLogs(PAGE_SIZE, 0);
      setLogs(data.lines || []);
      setTotalLogLines(data.total_lines || 0);
    } catch (err) {
      console.error('Failed to load daemon logs', err);
    } finally {
      setLoadingLogs(false);
    }
  }, []);

  const handleToggleLogs = () => {
    const next = !showLogs;
    setShowLogs(next);
    if (next && logs.length === 0) {
      fetchLogs();
    }
  };

  const handleLoadMoreLogs = useCallback(async () => {
    if (loadingMoreLogs || logs.length >= totalLogLines) return;
    const container = terminalRef.current;
    const prevScrollHeight = container ? container.scrollHeight : 0;
    const prevScrollTop = container ? container.scrollTop : 0;

    setLoadingMoreLogs(true);
    try {
      const data = await getDaemonLogs(PAGE_SIZE, logs.length);
      if (data.lines && data.lines.length > 0) {
        setLogs((prev) => [...data.lines, ...prev]);
        // Preserve scroll position when older logs are prepended
        requestAnimationFrame(() => {
          if (terminalRef.current) {
            const newScrollHeight = terminalRef.current.scrollHeight;
            terminalRef.current.scrollTop = prevScrollTop + (newScrollHeight - prevScrollHeight);
          }
        });
      }
      setTotalLogLines(data.total_lines || 0);
    } catch (err) {
      console.error('Failed to load older daemon logs', err);
    } finally {
      setLoadingMoreLogs(false);
    }
  }, [loadingMoreLogs, logs.length, totalLogLines]);

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    if (target.scrollTop <= 20 && !loadingMoreLogs && logs.length < totalLogLines && !loadingLogs) {
      handleLoadMoreLogs();
    }
  };

  useEffect(() => {
    if (open) {
      fetchLogs();
    } else {
      setShowLogs(false);
    }
  }, [open, fetchLogs]);

  useEffect(() => {
    if (logs.length > 0 && isInitialLoadRef.current && showLogs) {
      if (terminalRef.current) {
        terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
      }
      isInitialLoadRef.current = false;
    }
  }, [logs, showLogs]);

  useEffect(() => {
    if (showLogs && logs.length > 0) {
      requestAnimationFrame(() => {
        if (terminalRef.current) {
          terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
        }
      });
    }
  }, [showLogs]);

  const handleCopyLogs = () => {
    navigator.clipboard.writeText(logs.join('\n'));
    setCopiedLogs(true);
    setTimeout(() => setCopiedLogs(false), 2000);
  };

  const handleCopy = (key: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleInstall = async () => {
    setInstalling(true);
    try {
      await installDaemon();
      toast({
        title: 'Daemon Installed',
        description: 'Native OS background service registered and scheduled for 15-minute quota sync.',
        variant: 'success',
      });
      await onRefreshDaemon();
      await fetchLogs();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to install daemon';
      toast({
        title: 'Installation Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setInstalling(false);
    }
  };

  const handleUninstall = async () => {
    setUninstalling(true);
    try {
      await uninstallDaemon();
      toast({
        title: 'Daemon Removed',
        description: 'Background service unregistered from the host OS service manager.',
        variant: 'success',
      });
      await onRefreshDaemon();
      await fetchLogs();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to uninstall daemon';
      toast({
        title: 'Uninstall Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setUninstalling(false);
    }
  };

  const handleRunNow = async () => {
    setRunning(true);
    try {
      const res = await runDaemonOnce();
      toast({
        title: 'Quota Sync Completed',
        description: res.message || 'Triggered quota cache refresh across all configured profiles.',
        variant: 'success',
      });
      await onRefreshDaemon();
      await fetchLogs();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to run daemon task';
      toast({
        title: 'Execution Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setRunning(false);
    }
  };

  const formatTimeAgo = (iso?: string) => {
    if (!iso) return 'Never';
    try {
      const date = new Date(iso);
      if (isNaN(date.getTime()) || date.getTime() <= 0) return 'Never';
      const diffMs = Date.now() - date.getTime();
      const diffMins = Math.floor(diffMs / (1000 * 60));
      if (diffMins < 1) return 'Just now';
      if (diffMins < 60) return `${diffMins}m ago`;
      const diffHours = Math.floor(diffMins / 60);
      if (diffHours < 24) return `${diffHours}h ago`;
      return `${Math.floor(diffHours / 24)}d ago`;
    } catch {
      return 'Never';
    }
  };

  const isInstalled = !!daemonInfo?.installed;
  const isActive = !!daemonInfo?.active;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto bg-card border-border text-foreground p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-3 border-b border-border pr-8">
          <div className="flex items-center gap-2.5 flex-wrap">
            <DialogTitle className="text-base font-semibold text-foreground flex items-center gap-2">
              <Activity className="h-4 w-4 text-foreground" />
              <span>OS Background Daemon</span>
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[10px] font-mono bg-secondary/40 text-muted-foreground border-border px-2 py-0.5 rounded-geist"
            >
              15m Quota Pre-Warm
            </Badge>
          </div>
          <DialogDescription className="text-xs text-muted-foreground leading-relaxed">
            Native scheduled background service (launchd on macOS / systemd on Linux) for pre-warming quota and rate-limit caches without resident memory overhead.
          </DialogDescription>
        </DialogHeader>

        {/* Operational Status Banner */}
        <div className="rounded-geist border border-border bg-background p-3 text-xs space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] uppercase tracking-wider font-mono text-muted-foreground">
              Operational Status
            </span>
            {isActive ? (
              <Badge
                variant="outline"
                className="bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                <span>Active (Every 15m)</span>
              </Badge>
            ) : isInstalled ? (
              <Badge
                variant="outline"
                className="bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-status-warning-solid" />
                <span>Installed (Inactive)</span>
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="bg-secondary/40 text-muted-foreground border-border text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-geist"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/50" />
                <span>Not Installed</span>
              </Badge>
            )}
          </div>

          <div className="grid grid-cols-2 gap-2 font-mono text-[11px] pt-1">
            <div className="p-2 rounded-geist bg-card border border-border">
              <span className="text-muted-foreground/70 block text-[10px] uppercase">Service Label</span>
              <span className="text-foreground font-medium truncate block" title={daemonInfo?.label || 'com.aim-cli.daemon'}>
                {daemonInfo?.label || 'com.aim-cli.daemon'}
              </span>
            </div>
            <div className="p-2 rounded-geist bg-card border border-border">
              <span className="text-muted-foreground/70 block text-[10px] uppercase">Schedule Interval</span>
              <span className="text-foreground font-medium block">
                Every {daemonInfo?.interval_sec ? `${daemonInfo.interval_sec / 60}m` : '15m'}
              </span>
            </div>
          </div>
        </div>

        {/* Execution & Telemetry Recap */}
        <div className="rounded-geist border border-border bg-background p-3 text-xs space-y-2.5 font-mono">
          <div className="flex items-center justify-between text-muted-foreground">
            <span className="flex items-center gap-1.5 text-[11px]">
              <Clock className="h-3.5 w-3.5 text-muted-foreground/70" /> Last Sync
            </span>
            <span className="text-foreground tabular-nums">
              {formatTimeAgo(daemonInfo?.last_run)}
            </span>
          </div>

          {daemonInfo?.last_run_message && (
            <div className="p-2.5 rounded-geist bg-card border border-border text-[11px] text-muted-foreground leading-relaxed break-words" title={daemonInfo.last_run_message}>
              {daemonInfo.last_run_message}
            </div>
          )}

          {daemonInfo?.config_path && (
            <div className="flex items-center justify-between gap-3 text-muted-foreground pt-1.5 border-t border-border">
              <span className="flex items-center gap-1.5 text-[11px] truncate flex-1 min-w-0" title={daemonInfo.config_path}>
                <FileText className="h-3 w-3 shrink-0 text-muted-foreground/70" />
                <span className="truncate">{daemonInfo.config_path}</span>
              </span>
              <button
                type="button"
                onClick={() => handleCopy('config', daemonInfo.config_path)}
                className="text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors shrink-0"
              >
                {copiedKey === 'config' ? <Check className="h-3 w-3 text-status-success" /> : <Copy className="h-3 w-3" />}
                {copiedKey === 'config' ? 'Copied' : 'Copy'}
              </button>
            </div>
          )}

          {daemonInfo?.log_path && (
            <div className="flex items-center justify-between gap-3 text-muted-foreground">
              <span className="flex items-center gap-1.5 text-[11px] truncate flex-1 min-w-0" title={daemonInfo.log_path}>
                <Folder className="h-3 w-3 shrink-0 text-muted-foreground/70" />
                <span className="truncate">{daemonInfo.log_path}</span>
              </span>
              <button
                type="button"
                onClick={() => handleCopy('log', daemonInfo.log_path)}
                className="text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors shrink-0"
              >
                {copiedKey === 'log' ? <Check className="h-3 w-3 text-status-success" /> : <Copy className="h-3 w-3" />}
                {copiedKey === 'log' ? 'Copied' : 'Copy'}
              </button>
            </div>
          )}
        </div>

        {/* Execution Log Stream */}
        <div className="rounded-geist border border-border bg-background p-3 text-xs space-y-2.5 transition-all">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 select-none">
              <Terminal className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-[11px] uppercase tracking-wider font-mono text-muted-foreground font-semibold">
                Execution Logs
              </span>
              {totalLogLines > 0 && (
                <span className="text-[10px] font-mono bg-secondary/70 text-muted-foreground border border-border px-1.5 py-0.5 rounded-geist leading-none inline-flex items-center">
                  {showLogs && logs.length < totalLogLines
                    ? `${logs.length} of ${totalLogLines} lines`
                    : `${totalLogLines} lines`}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1.5">
              {showLogs && (
                <>
                  <button
                    type="button"
                    onClick={fetchLogs}
                    disabled={loadingLogs}
                    title="Refresh logs"
                    className="text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors"
                  >
                    <RefreshCw className={cn("h-3 w-3", loadingLogs && "animate-spin")} />
                    Refresh
                  </button>
                  {logs.length > 0 && (
                    <button
                      type="button"
                      onClick={handleCopyLogs}
                      title="Copy logs"
                      className="text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors"
                    >
                      {copiedLogs ? <Check className="h-3 w-3 text-status-success" /> : <Copy className="h-3 w-3" />}
                      {copiedLogs ? 'Copied' : 'Copy'}
                    </button>
                  )}
                </>
              )}
              <button
                type="button"
                onClick={handleToggleLogs}
                className="text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded-geist bg-card hover:bg-secondary border border-border inline-flex items-center gap-1 cursor-pointer transition-colors"
              >
                {showLogs ? 'Collapse' : 'View Logs'}
              </button>
            </div>
          </div>

          {showLogs && (
            <div
              ref={terminalRef}
              onScroll={handleScroll}
              className="bg-[#050505] border border-[#222222] rounded-geist p-2.5 font-mono text-[11px] max-h-44 overflow-y-auto space-y-1 select-text scrollbar-thin"
            >
              {loadingLogs && logs.length === 0 ? (
                <div className="py-6 text-center text-muted-foreground/60 flex items-center justify-center gap-2">
                  <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                  <span>Loading daemon logs...</span>
                </div>
              ) : logs.length === 0 ? (
                <div className="py-6 text-center text-muted-foreground/60">
                  No daemon logs recorded yet. Trigger a sync below to generate execution records.
                </div>
              ) : (
                <>
                  {loadingMoreLogs && (
                    <div className="py-2 text-center text-muted-foreground/60 flex items-center justify-center gap-1.5 border-b border-[#1c1c1c] mb-2 font-mono text-[10px]">
                      <RefreshCw className="h-3 w-3 animate-spin text-primary" />
                      <span>Loading earlier records...</span>
                    </div>
                  )}
                  {logs.length >= totalLogLines && totalLogLines > 0 && (
                    <div className="py-2 text-center text-muted-foreground/40 select-none border-b border-[#1c1c1c] mb-2 font-mono text-[10px]">
                      — Beginning of execution logs ({totalLogLines} lines) —
                    </div>
                  )}
                  {logs.map((rawLine, i) => {
                    const { timestamp, level, message } = parseLogLine(rawLine);

                    return (
                      <div key={i} className="leading-relaxed break-all flex items-baseline gap-2 font-mono text-[11px] py-0.5">
                        {timestamp && (
                          <span className="text-[#666666] select-none shrink-0 font-mono text-[10px]">
                            [{timestamp}]
                          </span>
                        )}
                        <span
                          className={cn(
                            "font-semibold select-none shrink-0 text-[9.5px] px-1.5 py-0.2 rounded-xs uppercase tracking-wide",
                            level === 'ERROR' && "text-red-400 bg-red-950/50 border border-red-800/50",
                            level === 'WARN' && "text-yellow-400 bg-yellow-950/50 border border-yellow-700/50",
                            level === 'INFO' && "text-sky-400 bg-sky-950/40 border border-sky-800/40"
                          )}
                        >
                          {level}
                        </span>
                        <span
                          className={cn(
                            "flex-1",
                            level === 'ERROR' && "text-red-300 font-medium",
                            level === 'WARN' && "text-yellow-300 font-medium",
                            level === 'INFO' && "text-[#ededed]"
                          )}
                        >
                          {message}
                        </span>
                      </div>
                    );
                  })}
                </>
              )}
            </div>
          )}
        </div>

        <DialogFooter className="pt-2 flex items-center justify-between sm:justify-between w-full">
          {isInstalled ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={uninstalling || running}
              onClick={handleUninstall}
              className="border-border bg-card hover:bg-status-danger-bg hover:border-status-danger-border text-muted-foreground hover:text-status-danger text-xs font-sans rounded-geist"
            >
              {uninstalling ? (
                <>
                  <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                  Uninstalling...
                </>
              ) : (
                <>
                  <Trash2 className="h-3.5 w-3.5 mr-1.5" />
                  Uninstall Service
                </>
              )}
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="border-border bg-card hover:bg-secondary text-muted-foreground hover:text-foreground text-xs font-mono rounded-geist"
            >
              Close
            </Button>
          )}

          <div className="flex items-center gap-2">
            {isInstalled ? (
              <Button
                type="button"
                size="sm"
                disabled={running || uninstalling}
                onClick={handleRunNow}
                className="bg-primary hover:bg-primary/90 text-primary-foreground text-xs font-sans font-medium shadow-sm transition-all cursor-pointer rounded-geist"
              >
                {running ? (
                  <>
                    <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                    Syncing Quotas...
                  </>
                ) : (
                  <>
                    <Play className="h-3.5 w-3.5 mr-1.5" />
                    Sync Quotas Now
                  </>
                )}
              </Button>
            ) : (
              <Button
                type="button"
                size="sm"
                disabled={installing}
                onClick={handleInstall}
                className="bg-primary hover:bg-primary/90 text-primary-foreground text-xs font-sans font-medium shadow-sm transition-all cursor-pointer rounded-geist"
              >
                {installing ? (
                  <>
                    <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                    Installing...
                  </>
                ) : (
                  <>
                    <Activity className="h-3.5 w-3.5 mr-1.5" />
                    Install & Start Daemon
                  </>
                )}
              </Button>
            )}
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
