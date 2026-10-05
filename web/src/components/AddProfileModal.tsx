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
import { Switch } from './ui/switch';
import { useToast } from './ui/use-toast';
import { CreateProfileRequest, createProfile } from '@/lib/api';

interface AddProfileModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onProfileCreated?: () => void;
}

export function AddProfileModal({
  open,
  onOpenChange,
  onProfileCreated,
}: AddProfileModalProps) {
  const [formData, setFormData] = React.useState<CreateProfileRequest>({
    agent: 'claude',
    name: '',
    email: '',
    clone_from: '',
  });
  const [enableClone, setEnableClone] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);

  const { toast } = useToast();

  React.useEffect(() => {
    if (open) {
      setFormData({
        agent: 'claude',
        name: '',
        email: '',
        clone_from: '',
      });
      setEnableClone(false);
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Profile name cannot be empty.',
        variant: 'destructive',
      });
      return;
    }

    setSubmitting(true);
    try {
      await createProfile({
        ...formData,
        clone_from: enableClone ? formData.clone_from : '',
      });
      toast({
        title: 'Profile Created',
        description: `Successfully added ${formData.agent} profile "${formData.name}".`,
        variant: 'success',
      });
      onOpenChange(false);
      onProfileCreated?.();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create profile';
      toast({
        title: 'Creation Failed',
        description: msg,
        variant: 'destructive',
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="border border-[#262626] bg-[#111111] text-[#ededed] rounded-geist sm:max-w-md shadow-2xl">
        <DialogHeader className="space-y-1.5 pb-2 border-b border-[#262626]">
          <DialogTitle className="text-base font-semibold tracking-tight text-[#ededed]">
            Create New Agent Profile
          </DialogTitle>
          <DialogDescription className="text-xs text-[#a1a1a1]">
            Scaffold an isolated environment and configuration for your agent.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div className="space-y-1.5">
            <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
              Target Agent
            </label>
            <div className="grid grid-cols-4 gap-2">
              {['claude', 'codex', 'gemini', 'agy'].map((ag) => (
                <button
                  key={ag}
                  type="button"
                  onClick={() => setFormData({ ...formData, agent: ag })}
                  className={`py-2 px-3 text-xs font-mono rounded-geist border text-center transition-all cursor-pointer ${
                    formData.agent === ag
                      ? 'border-[#ededed] bg-[#222222] text-[#ededed] font-bold shadow-sm'
                      : 'border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] hover:text-[#ededed] hover:border-[#333333]'
                  }`}
                >
                  {ag.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
              Profile Name *
            </label>
            <Input
              required
              placeholder="e.g. work, sandbox, client-x"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
              Associated Email (Optional)
            </label>
            <Input
              type="email"
              placeholder="dev@example.com"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
            />
          </div>

          {/* Clone from existing profile toggle switch */}
          <div className="space-y-2 pt-2 border-t border-[#262626]">
            <div className="flex items-center justify-between py-0.5">
              <div className="space-y-0.5">
                <label htmlFor="modal-clone-toggle" className="text-xs font-medium text-[#ededed] cursor-pointer">
                  Clone from existing profile
                </label>
                <p className="text-[11px] text-[#a1a1a1]">
                  Copy authentication, credentials, and settings
                </p>
              </div>
              <Switch
                id="modal-clone-toggle"
                checked={enableClone}
                onCheckedChange={(checked) => {
                  setEnableClone(checked);
                  if (!checked) {
                    setFormData((prev) => ({ ...prev, clone_from: '' }));
                  }
                }}
              />
            </div>

            {enableClone && (
              <div className="space-y-1.5 pt-1">
                <label className="text-[11px] font-mono font-medium uppercase tracking-wider text-[#a1a1a1]">
                  Source Profile to Clone
                </label>
                <Input
                  placeholder="e.g. default"
                  value={formData.clone_from}
                  onChange={(e) => setFormData({ ...formData, clone_from: e.target.value })}
                  className="rounded-geist border-[#262626] bg-[#0a0a0a] text-xs h-9 text-[#ededed] placeholder:text-[#666666]"
                  autoFocus
                />
              </div>
            )}
          </div>

          <DialogFooter className="pt-3 border-t border-[#262626] flex gap-2 justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
              className="rounded-geist text-xs h-8 border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={submitting}
              className="bg-[#ededed] text-[#000000] hover:bg-white rounded-geist text-xs h-8 font-medium"
            >
              {submitting ? 'Creating...' : 'Create Profile'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
