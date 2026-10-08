import * as React from 'react';
import { Clock, Gauge } from 'lucide-react';
import { cn } from '@/lib/utils';

interface QuotaWindow {
  key: string;
  label: string;
  pct: number;
  reset: string | null;
}

interface ModelQuota {
  name: string;
  windows: QuotaWindow[];
}

interface ParsedQuota {
  type: 'multi-model' | 'single' | 'fallback';
  models?: ModelQuota[];
  single?: QuotaWindow;
}

function parseQuotaSummary(summary?: string): ParsedQuota | null {
  if (!summary || !summary.trim()) return null;

  // Pattern A: Model1 (5h: 0% [2h 3m], wk: 6% [4d 8h]), Model2 (...)
  const modelRegex = /([A-Za-z0-9_-]+)\s*\(([^)]+)\)/g;
  const models: ModelQuota[] = [];
  let match: RegExpExecArray | null;

  while ((match = modelRegex.exec(summary)) !== null) {
    const modelName = match[1];
    const windowsStr = match[2];
    const windows: QuotaWindow[] = [];

    const winParts = windowsStr.split(',').map((s) => s.trim());
    for (const part of winParts) {
      const winMatch = part.match(/^([a-zA-Z0-9]+)\s*:\s*(\d+)%(?:\s*\[([^\]]+)\])?/);
      if (winMatch) {
        let label = winMatch[1];
        const lower = label.toLowerCase();
        if (lower === '5h') label = '5-Hour Window';
        else if (lower === 'wk' || lower === 'week' || lower === 'weekly') label = 'Weekly Window';
        else if (lower === 'd' || lower === 'day' || lower === 'daily') label = 'Daily Window';

        windows.push({
          key: winMatch[1],
          label,
          pct: parseInt(winMatch[2], 10),
          reset: winMatch[3] ? winMatch[3].trim() : null,
        });
      }
    }
    if (windows.length > 0) {
      models.push({ name: modelName, windows });
    }
  }

  if (models.length > 0) {
    return { type: 'multi-model', models };
  }

  // Pattern B: "Weekly Limit: 100%" or "5h limit: 68% remaining"
  const singleMatch = summary.match(/(.+?):\s*(\d+)%(?:\s*(?:remaining)?)?(?:\s*\[([^\]]+)\])?/i);
  if (singleMatch) {
    return {
      type: 'single',
      single: {
        key: 'single',
        label: singleMatch[1].trim(),
        pct: parseInt(singleMatch[2], 10),
        reset: singleMatch[3] ? singleMatch[3].trim() : null,
      },
    };
  }

  return { type: 'fallback' };
}

interface QuotaBreakdownProps {
  summary?: string;
  bottleneckPct?: number;
  isExhausted?: boolean;
}

export function QuotaBreakdown({
  summary,
  bottleneckPct = 100,
  isExhausted = false,
}: QuotaBreakdownProps) {
  const parsed = React.useMemo(() => parseQuotaSummary(summary), [summary]);

  const effectiveBottleneckPct = isExhausted ? 0 : bottleneckPct;

  const getStatusColor = (pct: number) => {
    if (pct <= 0) return 'bg-status-danger-solid';
    if (pct <= 30) return 'bg-status-warning-solid';
    return 'bg-status-success-solid';
  };

  const getTextColor = (pct: number) => {
    if (pct <= 0) return 'text-status-danger';
    if (pct <= 30) return 'text-status-warning';
    return 'text-[#ededed]';
  };

  const renderMeter = (win: QuotaWindow) => {
    const isEx = win.pct <= 0;
    return (
      <div key={win.key} className="space-y-1.5 p-2.5 rounded-[6px] bg-[#0f0f0f] border border-[#222222]">
        <div className="flex items-center justify-between text-xs font-sans">
          <span className="text-muted-foreground flex items-center gap-1.5 font-normal">
            {win.label}
          </span>
          <div className="flex items-center gap-2">
            {win.reset && (
              <span className="inline-flex items-center gap-1 text-[10px] font-mono text-muted-foreground font-normal">
                <Clock className="h-2.5 w-2.5 text-muted-foreground/80 shrink-0" />
                <span>resets in {win.reset}</span>
              </span>
            )}
            <span className={`font-semibold font-mono tabular-nums text-xs ${getTextColor(win.pct)}`}>
              {win.pct}%
            </span>
          </div>
        </div>

        {/* Standardized h-1.5 progress bar with visible minimum alert indicator */}
        <div
          role="progressbar"
          aria-valuenow={win.pct}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${win.label} quota`}
          className={cn(
            "w-full h-1.5 rounded-full overflow-hidden border",
            isEx
              ? "bg-status-danger-bg border-status-danger-border"
              : "bg-[#181818] border-[#262626]/50"
          )}
        >
          <div
            className={cn(
              "h-full transition-all duration-300",
              isEx ? "w-full bg-status-danger-solid/40" : getStatusColor(win.pct)
            )}
            style={{ width: isEx ? '100%' : `${Math.min(100, win.pct)}%` }}
          />
        </div>
      </div>
    );
  };

  if (!parsed || parsed.type === 'fallback') {
    const isEx = isExhausted || effectiveBottleneckPct <= 0;
    return (
      <div className="space-y-2 p-3 rounded-[6px] bg-[#0a0a0a] border border-[#222222]">
        <div className="flex items-center justify-between text-xs font-sans">
          <span className="text-muted-foreground uppercase tracking-wider flex items-center gap-1.5 font-medium text-[11px]">
            <Gauge className="h-3 w-3 text-muted-foreground/60" />
            Live Quota Telemetry
          </span>
          <span className={`font-semibold font-mono tabular-nums ${getTextColor(effectiveBottleneckPct)}`}>
            {effectiveBottleneckPct}% remaining
          </span>
        </div>
        <div
          role="progressbar"
          aria-valuenow={effectiveBottleneckPct}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label="Live quota telemetry"
          className={cn(
            "w-full h-1.5 rounded-full overflow-hidden border",
            isEx
              ? "bg-status-danger-bg border-status-danger-border"
              : "bg-[#181818] border-[#262626]/50"
          )}
        >
          <div
            className={cn(
              "h-full transition-all duration-300",
              isEx ? "w-full bg-status-danger-solid/40" : getStatusColor(effectiveBottleneckPct)
            )}
            style={{ width: isEx ? '100%' : `${Math.min(100, effectiveBottleneckPct)}%` }}
          />
        </div>
        {summary && (
          <p className="text-[11px] font-sans text-muted-foreground/70 pt-1">
            {summary}
          </p>
        )}
      </div>
    );
  }

  if (parsed.type === 'single' && parsed.single) {
    const isEx = isExhausted || parsed.single.pct <= 0;
    return (
      <div className="space-y-2 p-4 rounded-geist bg-[#0c0c0c] border border-[#2a2a2a]">
        <div className="flex items-center justify-between text-xs font-sans">
          <span className="text-muted-foreground uppercase tracking-wider flex items-center gap-1.5 font-medium text-[11px] font-mono">
            <Gauge className="h-3.5 w-3.5 text-muted-foreground/60" />
            {parsed.single.label}
          </span>
          <div className="flex items-center gap-2">
            {parsed.single.reset && (
              <span className="inline-flex items-center gap-1 text-[10px] font-mono text-muted-foreground font-normal">
                <Clock className="h-2.5 w-2.5 text-muted-foreground/80 shrink-0" />
                <span>resets in {parsed.single.reset}</span>
              </span>
            )}
            <span className={`font-semibold font-mono tabular-nums text-xs ${getTextColor(parsed.single.pct)}`}>
              {parsed.single.pct}% remaining
            </span>
          </div>
        </div>
        <div
          role="progressbar"
          aria-valuenow={parsed.single.pct}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${parsed.single.label} quota`}
          className={cn(
            "w-full h-1.5 rounded-full overflow-hidden border",
            isEx
              ? "bg-status-danger-bg border-status-danger-border"
              : "bg-[#181818] border-[#262626]/50"
          )}
        >
          <div
            className={cn(
              "h-full transition-all duration-300",
              isEx ? "w-full bg-status-danger-solid/40" : getStatusColor(parsed.single.pct)
            )}
            style={{ width: isEx ? '100%' : `${Math.min(100, parsed.single.pct)}%` }}
          />
        </div>
      </div>
    );
  }

  // Multi-model layout (e.g. Gemini + Claude)
  return (
    <div className="space-y-4 p-4 rounded-geist bg-[#0c0c0c] border border-[#2a2a2a]">
      <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-3 text-xs font-sans">
        <span className="text-[#888888] uppercase tracking-wider flex items-center gap-1.5 font-medium text-[11px] font-mono">
          <Gauge className="h-3.5 w-3.5 text-[#666666]" />
          Visual Quota Breakdown
        </span>
        <span className="text-[11px] text-[#666666] font-mono">
          {parsed.models?.length || 1} Engine Models Monitored
        </span>
      </div>

      <div className="space-y-3">
        {parsed.models?.map((m) => (
          <div key={m.name} className="space-y-2">
            <div className="flex items-center justify-between text-xs font-sans">
              <span className="text-[#ededed] font-medium flex items-center gap-1.5">
                <span className="h-1.5 w-1.5 rounded-full bg-[#555555]" />
                {m.name} Engine
              </span>
              <span className="text-[11px] text-[#555555]">
                {m.windows.length} Time Windows
              </span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {m.windows.map((w) => renderMeter(w))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
