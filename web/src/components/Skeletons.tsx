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

export function ViewLoadingSkeleton() {
  return (
    <div className="flex items-center justify-center p-16 text-xs font-sans text-muted-foreground animate-in fade-in duration-200">
      <div className="flex items-center gap-2.5">
        <GeistSpinner size={16} />
        <span>Loading view…</span>
      </div>
    </div>
  );
}

export function TeammateCardsSkeleton({
  count = 3,
}: {
  count?: number;
  isRefreshing?: boolean;
}) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="rounded-geist border border-border bg-card p-4 space-y-4 shadow-xs"
        >
            {/* Header: Hostname + IP + Status */}
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1.5 flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <div className="h-4 w-4 rounded bg-muted animate-pulse shrink-0" />
                  <div className="h-4 w-28 rounded bg-muted animate-pulse" />
                </div>
                <div className="h-3 w-20 rounded bg-muted/60 animate-pulse" />
              </div>
              <div className="h-5 w-16 rounded-full bg-muted/60 animate-pulse shrink-0" />
            </div>

            {/* Shared profiles list skeleton */}
            <div className="space-y-2 pt-1 border-t border-border/50">
              <div className="h-3 w-24 rounded bg-muted/60 animate-pulse" />
              <div className="p-2.5 rounded-geist border border-border/60 bg-secondary/40 space-y-2">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5">
                    <div className="h-4 w-12 rounded-full bg-muted animate-pulse" />
                    <div className="h-3.5 w-16 rounded bg-muted animate-pulse" />
                  </div>
                  <div className="h-5 w-16 rounded-geist bg-muted/80 animate-pulse" />
                </div>
                <div className="h-1.5 w-full rounded-full bg-muted/40 animate-pulse" />
              </div>
            </div>

            {/* Footer metrics strip */}
            <div className="flex items-center gap-2 pt-2 border-t border-border/40">
              <div className="h-3.5 w-14 rounded bg-muted/40 animate-pulse" />
              <div className="h-3.5 w-12 rounded bg-muted/40 animate-pulse" />
              <div className="h-3.5 w-16 rounded bg-muted/40 animate-pulse" />
            </div>
          </div>
        ))}
      </div>
  );
}

export function NetworkViewSkeleton() {
  return (
    <div className="space-y-6">
      {/* Top Header Bar Skeleton */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="h-4 w-4 rounded-full bg-muted animate-pulse" />
            <div className="h-6 w-48 rounded bg-muted animate-pulse" />
            <div className="h-5 w-14 rounded-full bg-muted/60 animate-pulse" />
          </div>
        </div>

        {/* Global Policy Control Skeleton placeholder in Header */}
        <div className="flex items-center gap-2 shrink-0">
          <div className="h-8 w-72 rounded-geist bg-secondary border border-border animate-pulse shrink-0" />
        </div>
      </div>

      {/* VPN Tunnel Status Strip Skeleton */}
      <div className="p-4 rounded-geist border border-border bg-card flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0">
          <div className="h-8 w-8 rounded-full border border-border bg-secondary shrink-0 animate-pulse" />
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <div className="h-3.5 w-36 rounded bg-muted animate-pulse" />
              <div className="h-2 w-2 rounded-full bg-muted/60 animate-pulse" />
            </div>
            <div className="h-3 w-64 max-w-full rounded bg-muted/50 animate-pulse" />
          </div>
        </div>
        <div className="h-8 w-20 rounded-geist bg-secondary border border-border animate-pulse shrink-0" />
      </div>

      {/* Tab Bar Skeleton */}
      <div className="flex items-center justify-between gap-3 pb-1">
        <div className="h-9 w-96 rounded-geist bg-secondary/40 border border-border animate-pulse" />
        <div className="hidden md:flex items-center gap-3">
          <div className="h-4 w-24 rounded bg-muted/40 animate-pulse" />
        </div>
      </div>

      {/* Teammates Section Header Skeleton */}
      <div className="space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <div className="h-4 w-4 rounded bg-muted animate-pulse" />
              <div className="h-4 w-36 rounded bg-muted animate-pulse" />
              <div className="h-4 w-12 rounded-full bg-muted/60 animate-pulse" />
            </div>
            <div className="h-3 w-60 rounded bg-muted/40 animate-pulse" />
          </div>
          <div className="flex items-center gap-2">
            <div className="h-8 w-48 rounded-geist bg-secondary border border-border animate-pulse shrink-0" />
            <div className="h-8 w-8 rounded-geist bg-secondary border border-border animate-pulse shrink-0" />
            <div className="h-8 w-8 rounded-geist bg-secondary border border-border animate-pulse shrink-0" />
          </div>
        </div>

        {/* Teammate Cards Skeleton */}
        <TeammateCardsSkeleton count={3} />
      </div>
    </div>
  );
}

export function ProfilesOverviewSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-in fade-in duration-200">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="rounded-geist border border-border bg-card p-5 space-y-4 shadow-xs"
        >
          {/* Header */}
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="h-8 w-8 rounded-full bg-muted animate-pulse shrink-0" />
              <div className="space-y-1">
                <div className="h-4 w-28 rounded bg-muted animate-pulse" />
                <div className="h-3 w-20 rounded bg-muted/60 animate-pulse" />
              </div>
            </div>
            <div className="h-5 w-16 rounded-full bg-muted/60 animate-pulse shrink-0" />
          </div>

          {/* Quota Scorecard Pill */}
          <div className="p-3 rounded-geist bg-secondary/50 border border-border/60 space-y-2">
            <div className="flex items-center justify-between">
              <div className="h-3 w-20 rounded bg-muted/80 animate-pulse" />
              <div className="h-3 w-10 rounded bg-muted/80 animate-pulse" />
            </div>
            <div className="h-2 w-full rounded-full bg-muted/40 animate-pulse" />
          </div>

          {/* Footer details */}
          <div className="flex items-center justify-between pt-2 border-t border-border/40">
            <div className="h-3 w-24 rounded bg-muted/40 animate-pulse" />
            <div className="h-6 w-16 rounded-geist bg-muted/60 animate-pulse" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function McpCatalogSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 animate-in fade-in duration-200">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="rounded-geist border border-border bg-card p-5 space-y-3.5 shadow-xs"
        >
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="h-7 w-7 rounded-geist bg-muted animate-pulse shrink-0" />
              <div className="space-y-1">
                <div className="h-4 w-28 rounded bg-muted animate-pulse" />
                <div className="h-3 w-20 rounded bg-muted/60 animate-pulse" />
              </div>
            </div>
            <div className="h-5 w-14 rounded-full bg-muted/60 animate-pulse shrink-0" />
          </div>
          <div className="h-3 w-3/4 rounded bg-muted/40 animate-pulse" />
          <div className="flex items-center gap-2 pt-2 border-t border-border/40">
            <div className="h-5 w-16 rounded-[4px] bg-secondary border border-border animate-pulse" />
            <div className="h-5 w-20 rounded-[4px] bg-secondary border border-border animate-pulse" />
          </div>
        </div>
      ))}
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
