import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from './ui/dialog';
import { Button } from './ui/button';
import { AlertTriangle, Trash2, RefreshCw } from 'lucide-react';
import { AgentBadge } from './AgentBadge';

interface DeleteProfileModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  target: { agent: string; name: string } | null;
  onConfirm: () => Promise<void>;
  isDeleting?: boolean;
}

export const DeleteProfileModal: React.FC<DeleteProfileModalProps> = ({
  open,
  onOpenChange,
  target,
  onConfirm,
  isDeleting = false,
}) => {
  const [confirmInput, setConfirmInput] = React.useState('');

  React.useEffect(() => {
    if (open) {
      setConfirmInput('');
    }
  }, [open, target]);

  if (!target) return null;

  const isMatched = confirmInput.trim() === target.name;

  return (
    <Dialog open={open} onOpenChange={(val) => !isDeleting && onOpenChange(val)}>
      <DialogContent className="max-w-md bg-[#111111] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#262626] pr-8">
          <div className="flex items-center gap-2.5 flex-wrap">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-status-danger" />
              <span>Delete Profile</span>
            </DialogTitle>
            {target.agent && target.agent.trim() !== '' && (
              <AgentBadge agent={target.agent} size="sm" />
            )}
          </div>
          <DialogDescription className="text-xs text-[#a1a1a1]">
            This action cannot be undone. Are you sure you want to delete profile{' '}
            <span className="text-[#ededed] font-medium font-mono">&ldquo;{target.name}&rdquo;</span>?
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-geist border border-[#262626] bg-[#0a0a0a] p-3 text-xs space-y-2 font-mono text-[#a1a1a1]">
          <div className="flex justify-between items-center">
            <span>Profile Identifier:</span>
            <span className="text-[#ededed] font-medium">{target.name}</span>
          </div>
          {target.agent && target.agent.trim() !== '' && (
            <div className="flex justify-between items-center">
              <span>Agent Engine:</span>
              <span className="text-[#ededed] font-medium">{target.agent}</span>
            </div>
          )}
          <p className="text-[11px] text-[#666666] pt-2 border-t border-[#262626] leading-relaxed">
            This will permanently remove the profile directory, credentials, and configuration from AIM.
          </p>
        </div>

        <div className="space-y-1.5 pt-1">
          <label htmlFor="confirm-delete-input" className="text-xs text-[#a1a1a1] font-sans">
            Please type <span className="font-mono text-[#ededed] font-medium">{target.name}</span> to confirm:
          </label>
          <input
            id="confirm-delete-input"
            type="text"
            value={confirmInput}
            onChange={(e) => setConfirmInput(e.target.value)}
            placeholder={target.name}
            className="w-full h-8 px-2.5 text-xs font-mono rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] placeholder:text-[#555555] focus:border-[#444444] focus:outline-none"
            autoComplete="off"
            disabled={isDeleting}
          />
        </div>

        <DialogFooter className="pt-2 border-t border-[#262626] flex items-center justify-end gap-2 w-full">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={isDeleting}
            onClick={() => onOpenChange(false)}
            className="border-[#262626] bg-[#0a0a0a] hover:bg-[#161616] text-[#a1a1a1] hover:text-[#ededed] text-xs font-sans rounded-geist cursor-pointer"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            disabled={!isMatched || isDeleting}
            onClick={onConfirm}
            className="text-xs font-sans font-medium shadow-sm transition-all cursor-pointer rounded-geist disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {isDeleting ? (
              <>
                <RefreshCw className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                Deleting...
              </>
            ) : (
              <>
                <Trash2 className="h-3.5 w-3.5 mr-1.5" />
                Delete Profile
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
