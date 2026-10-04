import React, { useState } from 'react';
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
} from 'lucide-react';
import { DaemonDTO, installDaemon, uninstallDaemon, runDaemonOnce } from '../lib/api';
import { useToast } from './ui/use-toast';

interface DaemonModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  daemonInfo: DaemonDTO | null;
  onRefreshDaemon: () => Promise<void>;
}

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
  const { toast } = useToast();

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
      <DialogContent className="max-w-xl bg-[#0c0c0c] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#1f1f1f]">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <Activity className="h-4 w-4 text-[#ededed]" />
              OS Background Daemon
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[11px] font-mono bg-[#161616] text-[#ededed] border-[#2e2e2e]"
            >
              15m Quota Pre-Warm
            </Badge>
          </div>
          <DialogDescription className="text-xs text-[#888888]">
            Native scheduled background service (launchd on macOS / systemd on Linux) for pre-warming quota and rate-limit caches without resident memory overhead.
          </DialogDescription>
        </DialogHeader>

        {/* Operational Status Banner */}
        <div className="rounded-md border border-[#222222] bg-[#121212] p-3 text-xs space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-[11px] uppercase tracking-wider font-mono text-[#888888]">
              Operational Status
            </span>
            {isActive ? (
              <Badge
                variant="outline"
                className="bg-[#141414] text-[#a1a1a1] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                <span>Active (Every 15m)</span>
              </Badge>
            ) : isInstalled ? (
              <Badge
                variant="outline"
                className="bg-[#141414] text-[#888888] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                <span>Installed (Inactive)</span>
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="bg-[#141414] text-[#888888] border-[#262626] text-[11px] font-mono flex items-center gap-1.5 px-2 py-0.5 rounded-md"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-[#555555]" />
                <span>Not Installed</span>
              </Badge>
            )}
          </div>

          <div className="grid grid-cols-2 gap-2 font-mono text-[11px] pt-1">
            <div className="p-2 rounded bg-[#0e0e0e] border border-[#1f1f1f]">
              <span className="text-[#666666] block text-[10px] uppercase">Service Label</span>
              <span className="text-[#ededed] font-medium truncate block" title={daemonInfo?.label || 'dev.aim-cli.daemon'}>
                {daemonInfo?.label || 'dev.aim-cli.daemon'}
              </span>
            </div>
            <div className="p-2 rounded bg-[#0e0e0e] border border-[#1f1f1f]">
              <span className="text-[#666666] block text-[10px] uppercase">Schedule Interval</span>
              <span className="text-[#ededed] font-medium block">
                Every {daemonInfo?.interval_sec ? `${daemonInfo.interval_sec / 60}m` : '15m'}
              </span>
            </div>
          </div>
        </div>

        {/* Execution & Telemetry Recap */}
        <div className="rounded-md border border-[#222222] bg-[#121212] p-3 text-xs space-y-2 font-mono">
          <div className="flex items-center justify-between text-[#888888]">
            <span className="flex items-center gap-1.5 text-[11px]">
              <Clock className="h-3.5 w-3.5 text-[#666666]" /> Last Sync
            </span>
            <span className="text-[#ededed] tabular-nums">
              {formatTimeAgo(daemonInfo?.last_run)}
            </span>
          </div>

          {daemonInfo?.last_run_message && (
            <div className="p-2 rounded bg-[#0e0e0e] border border-[#1f1f1f] text-[11px] text-[#a1a1a1] truncate" title={daemonInfo.last_run_message}>
              {daemonInfo.last_run_message}
            </div>
          )}

          {daemonInfo?.config_path && (
            <div className="flex items-center justify-between text-[#888888] pt-1 border-t border-[#1c1c1c]">
              <span className="flex items-center gap-1 text-[10.5px] truncate max-w-[340px]" title={daemonInfo.config_path}>
                <FileText className="h-3 w-3 shrink-0 text-[#666666]" />
                <span className="truncate">{daemonInfo.config_path}</span>
              </span>
              <button
                type="button"
                onClick={() => handleCopy('config', daemonInfo.config_path)}
                className="text-[10px] text-[#888888] hover:text-[#ededed] inline-flex items-center gap-1 cursor-pointer"
              >
                {copiedKey === 'config' ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                {copiedKey === 'config' ? 'Copied' : 'Copy'}
              </button>
            </div>
          )}

          {daemonInfo?.log_path && (
            <div className="flex items-center justify-between text-[#888888]">
              <span className="flex items-center gap-1 text-[10.5px] truncate max-w-[340px]" title={daemonInfo.log_path}>
                <Folder className="h-3 w-3 shrink-0 text-[#666666]" />
                <span className="truncate">{daemonInfo.log_path}</span>
              </span>
              <button
                type="button"
                onClick={() => handleCopy('log', daemonInfo.log_path)}
                className="text-[10px] text-[#888888] hover:text-[#ededed] inline-flex items-center gap-1 cursor-pointer"
              >
                {copiedKey === 'log' ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                {copiedKey === 'log' ? 'Copied' : 'Copy'}
              </button>
            </div>
          )}
        </div>

        <DialogFooter className="pt-2 border-t border-[#1f1f1f] flex items-center justify-between sm:justify-between w-full">
          {isInstalled ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={uninstalling || running}
              onClick={handleUninstall}
              className="border-[#262626] bg-[#141414] hover:bg-rose-500/10 hover:border-rose-500/30 text-[#888888] hover:text-rose-400 text-xs font-mono"
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
              className="border-[#262626] bg-[#141414] hover:bg-[#1a1a1a] text-[#888888] hover:text-[#ededed] text-xs font-mono"
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
                className="bg-[#ededed] hover:bg-white text-[#0a0a0a] text-xs font-mono font-medium shadow-sm transition-all cursor-pointer"
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
                className="bg-[#ededed] hover:bg-white text-[#0a0a0a] text-xs font-mono font-medium shadow-sm transition-all cursor-pointer"
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
