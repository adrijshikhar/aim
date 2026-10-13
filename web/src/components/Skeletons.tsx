import { ArrowLeft } from 'lucide-react';
import { GeistSpinner } from '@/components/ui/spinner';
import { Button } from '@/components/ui/button';

export function SessionDetailSkeleton({
  targetSessionId,
  onBack,
}: {
  targetSessionId?: string | null;
  onBack?: () => void;
}) {
  return (
    <div className="space-y-4 pb-12 animate-in fade-in duration-200">
      {/* Back button & top status bar */}
      <div className="flex items-center justify-between gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={onBack}
          className="h-7 px-2.5 text-xs font-sans text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist cursor-pointer"
        >
          <ArrowLeft className="h-3.5 w-3.5 mr-1.5" />
          Back to Sessions
        </Button>

        <div className="flex items-center gap-2 text-xs font-sans text-muted-foreground/80">
          <GeistSpinner size={14} />
          <span>
            Loading session {targetSessionId ? `${targetSessionId.slice(0, 8)}…` : '…'}
          </span>
        </div>
      </div>

      {/* Sticky Header Card Skeleton */}
      <div className="rounded-geist border border-border bg-[#111111]/95 p-5 shadow-lg space-y-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="space-y-2 flex-1 min-w-0">
            <div className="flex items-center gap-2.5">
              <div className="h-5 w-16 rounded-full bg-[#1c1c1c] animate-pulse" />
              <div className="h-6 w-72 max-w-full rounded bg-[#1c1c1c] animate-pulse" />
            </div>
            <div className="h-4 w-4/5 max-w-xl rounded bg-[#161616] animate-pulse" />
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <div className="h-8 w-24 rounded-geist bg-[#181818] border border-[#262626] animate-pulse" />
            <div className="h-8 w-8 rounded-geist bg-[#181818] border border-[#262626] animate-pulse" />
          </div>
        </div>

        {/* Metadata pill strip skeleton */}
        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-[#1c1c1c]">
          <div className="h-6 w-24 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse" />
          <div className="h-6 w-36 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse" />
          <div className="h-6 w-44 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse" />
          <div className="h-6 w-20 rounded-[4px] bg-[#161616] border border-[#222222] animate-pulse" />
        </div>
      </div>

      {/* Conversation Filter Toolbar Skeleton */}
      <div className="rounded-geist border border-border bg-[#0e0e0e] p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <div className="h-5 w-40 rounded bg-[#181818] animate-pulse" />
          <div className="h-6 w-48 rounded bg-[#161616] animate-pulse" />
        </div>
        <div className="h-7 w-64 max-w-full rounded bg-[#161616] border border-[#222222] animate-pulse" />
      </div>

      {/* Conversation Turns Stream Skeleton */}
      <div className="space-y-3">
        {/* User Prompt Skeleton */}
        <div className="rounded-geist border border-[#0070f3]/25 bg-[#0a0d14] p-4 space-y-3">
          <div className="flex items-center justify-between pb-2 border-b border-[#0070f3]/15">
            <div className="flex items-center gap-2">
              <div className="h-5 w-5 rounded-[4px] bg-[#0070f3]/20 animate-pulse" />
              <div className="h-4 w-24 rounded bg-[#0070f3]/20 animate-pulse" />
              <div className="h-3.5 w-12 rounded bg-[#161616] animate-pulse" />
            </div>
            <div className="h-4 w-16 rounded bg-[#161616] animate-pulse" />
          </div>
          <div className="space-y-2">
            <div className="h-4 w-5/6 rounded bg-[#161616] animate-pulse" />
            <div className="h-4 w-2/3 rounded bg-[#161616] animate-pulse" />
          </div>
        </div>

        {/* Assistant Response Skeleton */}
        <div className="rounded-geist border border-[#1f1f1f] bg-[#111111] p-4 space-y-3">
          <div className="flex items-center justify-between pb-2 border-b border-[#1c1c1c]">
            <div className="flex items-center gap-2">
              <div className="h-5 w-5 rounded-[4px] bg-[#222222] animate-pulse" />
              <div className="h-4 w-28 rounded bg-[#222222] animate-pulse" />
              <div className="h-3.5 w-12 rounded bg-[#161616] animate-pulse" />
            </div>
            <div className="h-4 w-16 rounded bg-[#161616] animate-pulse" />
          </div>
          <div className="space-y-2">
            <div className="h-4 w-full rounded bg-[#161616] animate-pulse" />
            <div className="h-4 w-11/12 rounded bg-[#161616] animate-pulse" />
            <div className="h-4 w-3/4 rounded bg-[#161616] animate-pulse" />
          </div>
        </div>
      </div>
    </div>
  );
}

export function ProfileDetailSkeleton({ profileName }: { profileName?: string }) {
  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      {/* Header Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="h-7 w-32 rounded-geist bg-[#181818] animate-pulse" />
          <div className="h-5 w-16 rounded-full bg-[#1c1c1c] animate-pulse" />
          <div className="h-5 w-24 rounded-full bg-[#1c1c1c] animate-pulse" />
        </div>
        <div className="flex items-center gap-2 text-xs font-sans text-muted-foreground/80">
          <GeistSpinner size={14} />
          <span>Loading profile {profileName ? `"${profileName}"` : ''}…</span>
        </div>
      </div>

      {/* Tabs Skeleton */}
      <div className="h-9 w-80 rounded-geist bg-[#141414] border border-[#222222] animate-pulse" />

      {/* Quota Scorecard Cards Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] p-5 space-y-4">
          <div className="h-5 w-32 rounded bg-[#181818] animate-pulse" />
          <div className="h-10 w-full rounded bg-[#141414] animate-pulse" />
          <div className="h-4 w-2/3 rounded bg-[#161616] animate-pulse" />
        </div>
        <div className="rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] p-5 space-y-4">
          <div className="h-5 w-32 rounded bg-[#181818] animate-pulse" />
          <div className="h-10 w-full rounded bg-[#141414] animate-pulse" />
          <div className="h-4 w-2/3 rounded bg-[#161616] animate-pulse" />
        </div>
      </div>
    </div>
  );
}
