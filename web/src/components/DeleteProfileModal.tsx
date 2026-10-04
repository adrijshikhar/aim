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
import { Badge } from './ui/badge';
import { AlertTriangle, Trash2, RefreshCw } from 'lucide-react';

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
  if (!target) return null;

  return (
    <Dialog open={open} onOpenChange={(val) => !isDeleting && onOpenChange(val)}>
      <DialogContent className="max-w-md bg-[#0c0c0c] border-[#262626] text-[#ededed] p-6 shadow-2xl rounded-geist">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#1f1f1f]">
          <div className="flex items-center justify-between">
            <DialogTitle className="text-base font-semibold text-[#ededed] flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-rose-500" />
              Delete Profile
            </DialogTitle>
            <Badge
              variant="outline"
              className="text-[11px] font-mono bg-[#161616] text-[#ededed] border-[#2e2e2e]"
            >
              {target.agent}
            </Badge>
          </div>
          <DialogDescription className="text-xs text-[#888888]">
            This action cannot be undone. Are you sure you want to delete profile{' '}
            <span className="text-[#ededed] font-medium font-mono">"{target.name}"</span>?
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-md border border-[#222222] bg-[#121212] p-3 text-xs space-y-2 font-mono text-[#888888]">
          <div className="flex justify-between items-center">
            <span>Profile Identifier:</span>
            <span className="text-[#ededed] font-medium">{target.name}</span>
          </div>
          <div className="flex justify-between items-center">
            <span>Agent Engine:</span>
            <span className="text-[#ededed] font-medium">{target.agent}</span>
          </div>
          <p className="text-[11px] text-[#777777] pt-2 border-t border-[#1c1c1c] leading-relaxed">
            This will permanently remove the profile directory, credentials, and configuration from AIM.
          </p>
        </div>

        <DialogFooter className="pt-2 border-t border-[#1f1f1f] flex items-center justify-between sm:justify-between w-full">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={isDeleting}
            onClick={() => onOpenChange(false)}
            className="border-[#262626] bg-[#141414] hover:bg-[#1a1a1a] text-[#888888] hover:text-[#ededed] text-xs font-mono"
          >
            Cancel
          </Button>
          <Button
            type="button"
            size="sm"
            disabled={isDeleting}
            onClick={onConfirm}
            className="bg-rose-600 hover:bg-rose-700 text-white text-xs font-mono font-medium shadow-sm transition-all cursor-pointer"
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
