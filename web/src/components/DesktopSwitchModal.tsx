import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { ArrowRight, Loader2, AppWindow } from 'lucide-react';
import { AgentBadge } from './AgentBadge';

interface DesktopSwitchModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  agent: string;
  targetProfile: string;
  currentProfile: string;
  onConfirmSwitch: () => Promise<void>;
}

export const DesktopSwitchModal: React.FC<DesktopSwitchModalProps> = ({
  open,
  onOpenChange,
  agent,
  targetProfile,
  currentProfile,
  onConfirmSwitch,
}) => {
  const [loading, setLoading] = React.useState(false);

  const handleConfirm = async () => {
    try {
      setLoading(true);
      await onConfirmSwitch();
      onOpenChange(false);
    } catch {
      // Error handling is managed by parent or toast
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(val) => !loading && onOpenChange(val)}>
      <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#262626]">
          <div className="flex items-center gap-2.5">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <AppWindow className="h-4 w-4 text-status-warning" />
              <span>Single-Instance Desktop App</span>
            </DialogTitle>
            {agent && <AgentBadge agent={agent} size="sm" />}
          </div>
          <DialogDescription className="text-xs text-[#a1a1a1]">
            Desktop GUI applications only support a single running instance per host. To launch with a different profile, the existing instance must be gracefully closed.
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-3 font-mono">
          <div className="flex items-center justify-between text-muted-foreground">
            <span>Current Running Profile:</span>
            <span className="text-[#ededed] font-medium bg-[#1e2025] px-2 py-0.5 rounded border border-[#262626]">
              {currentProfile || 'active'}
            </span>
          </div>

          <div className="flex items-center justify-center gap-2 py-1 text-muted-foreground">
            <span className="text-[11px]">Gracefully terminate</span>
            <ArrowRight className="h-3.5 w-3.5 text-[#61afef]" />
            <span className="text-[11px] text-[#ededed]">Relaunch isolated</span>
          </div>

          <div className="flex items-center justify-between text-muted-foreground">
            <span>Target Launch Profile:</span>
            <span className="text-status-success font-medium bg-[#183925] px-2 py-0.5 rounded border border-[#235836]">
              {targetProfile}
            </span>
          </div>

          <div className="text-[11px] text-[#666666] pt-2 border-t border-[#262626] leading-relaxed font-sans">
            AIM will send a graceful quit signal to the running application to save open tabs and workspace state, then launch the new instance in profile <span className="font-mono text-[#ededed] font-medium">&ldquo;{targetProfile}&rdquo;</span>.
          </div>
        </div>

        <DialogFooter className="flex items-center justify-end gap-2 pt-2 border-t border-[#262626]">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={loading}
            onClick={() => onOpenChange(false)}
            className="text-xs text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#1f1f1f]"
          >
            Cancel
          </Button>
          <Button
            type="button"
            size="sm"
            disabled={loading}
            onClick={handleConfirm}
            className="text-xs bg-[#61afef] text-[#1e1e1e] hover:bg-[#4d9bdf] font-medium transition-colors"
          >
            {loading ? (
              <>
                <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                Switching Profile...
              </>
            ) : (
              'Close & Switch Profile'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
