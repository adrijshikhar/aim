import * as React from 'react';
import {
  Clock,
  TrendingUp,
  TrendingDown,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Mail,
  ShieldCheck,
  Sparkles,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { AgentBadge } from '@/components/AgentBadge';
import { Badge } from '@/components/ui/badge';
import type { QuotaDTO, AccountInfo } from '@/lib/api';
import { calculateWindowPacing, WindowPacing } from '@/lib/quota-insights';

export interface QuotaWindowItem {
  key: string;
  label: string;
  pct: number;
  reset: string | null;
  resetsAt?: string | null;
  resetsInSec?: number | null;
}

export interface ModelQuotaGroup {
  name: string;
  windows: QuotaWindowItem[];
}

interface ParsedQuota {
  type: 'multi-model' | 'single' | 'fallback';
  models?: ModelQuotaGroup[];
  single?: QuotaWindowItem;
}

function parseQuotaSummary(summary?: string): ParsedQuota | null {
  if (!summary || !summary.trim()) return null;

  // Pattern A: Model1 (5h: 0% [2h 3m], wk: 6% [4d 8h]), Model2 (...)
  const modelRegex = /([A-Za-z0-9_-]+)\s*\(([^)]+)\)/g;
  const models: ModelQuotaGroup[] = [];
  let match: RegExpExecArray | null;

  while ((match = modelRegex.exec(summary)) !== null) {
    const modelName = match[1];
    const windowsStr = match[2];
    const windows: QuotaWindowItem[] = [];

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
          key: `${modelName}-${winMatch[1]}`,
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
    let label = singleMatch[1].trim();
    if (label.includes('•')) {
      const parts = label.split('•');
      label = parts[parts.length - 1].trim();
    }
    return {
      type: 'single',
      single: {
        key: 'single',
        label,
        pct: parseInt(singleMatch[2], 10),
        reset: singleMatch[3] ? singleMatch[3].trim() : null,
      },
    };
  }

  return { type: 'fallback' };
}

export function cleanAuthMethod(method?: string): string {
  if (!method) return '';
  if (method.includes('self_serve_business_prolite')) return 'ChatGPT · Business Prolite';
  if (method.includes('Google OAuth (consumer)')) return 'Google OAuth (Personal)';
  if (method === 'chatgpt_subscription' || method === 'chatgpt_account') return 'ChatGPT Plus/Team';
  if (method === 'codex_account') return 'Codex Subscription';
  if (method === 'claude_account' || method === 'claude_subscription') return 'Claude Pro/Team';
  if (method === 'api_key') return 'API Key';
  return method.replace(/_/g, ' ');
}

export interface QuotaBreakdownProps {
  agent?: string;
  account?: AccountInfo;
  quota?: QuotaDTO;
  summary?: string;
  bottleneckPct?: number;
  isExhausted?: boolean;
  sessionCount?: number;
}

export function QuotaBreakdown({
  agent,
  account,
  quota,
  summary,
  bottleneckPct = 100,
  isExhausted = false,
  sessionCount,
}: QuotaBreakdownProps) {
  const activeSummary = quota?.summary || summary;
  const effectiveBottleneckPct = isExhausted ? 0 : (quota?.bottleneck_pct ?? bottleneckPct);
  const isEffectiveExhausted = isExhausted || effectiveBottleneckPct <= 0 || (quota?.is_exhausted ?? false);
  const isWarning = !isEffectiveExhausted && effectiveBottleneckPct <= 30;

  // 1. Convert structured backend windows if present, or fallback to parsed summary
  const structuredModels = React.useMemo<ModelQuotaGroup[] | null>(() => {
    if (!quota?.windows || quota.windows.length === 0) return null;

    const groupMap = new Map<string, QuotaWindowItem[]>();
    for (const w of quota.windows) {
      const cat = w.category || (agent ? `${agent.toUpperCase()} Engine` : 'General');
      let label = w.name;
      const lower = label.toLowerCase();
      if (lower.includes('five') || lower.includes('5h')) {
        label = '5-Hour Window';
      } else if (lower.includes('weekly') || lower.includes('week') || lower.includes('7d')) {
        label = 'Weekly Window';
      }

      let resetStr: string | null = null;
      if (w.resets_in_sec && w.resets_in_sec > 0) {
        const d = Math.floor(w.resets_in_sec / 86400);
        const h = Math.floor((w.resets_in_sec % 86400) / 3600);
        const m = Math.floor((w.resets_in_sec % 3600) / 60);
        if (d > 0) resetStr = h > 0 ? `${d}d ${h}h` : `${d}d`;
        else if (h > 0) resetStr = m > 0 ? `${h}h ${m}m` : `${h}h`;
        else resetStr = `${Math.max(1, m)}m`;
      }

      const item: QuotaWindowItem = {
        key: `${cat}-${w.name}`,
        label,
        pct: w.remaining_pct,
        reset: resetStr,
        resetsAt: w.resets_at,
        resetsInSec: w.resets_in_sec,
      };

      if (!groupMap.has(cat)) {
        groupMap.set(cat, []);
      }
      groupMap.get(cat)!.push(item);
    }

    return Array.from(groupMap.entries()).map(([name, windows]) => ({
      name,
      windows,
    }));
  }, [quota?.windows, agent]);

  const parsed = React.useMemo(() => {
    if (structuredModels && structuredModels.length > 0) {
      return { type: 'multi-model' as const, models: structuredModels };
    }
    return parseQuotaSummary(activeSummary);
  }, [structuredModels, activeSummary]);

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

  // Find reset time from parsed summary or windows for card-level alerts
  const cardResetTime = React.useMemo(() => {
    if (parsed?.type === 'single' && parsed.single?.reset) return parsed.single.reset;
    if (parsed?.models) {
      for (const m of parsed.models) {
        for (const w of m.windows) {
          if (w.reset) return w.reset;
        }
      }
    }
    const match = activeSummary?.match(/\[([^\]]+)\]/);
    return match ? match[1].trim() : null;
  }, [parsed, activeSummary]);

  const renderMeter = (win: QuotaWindowItem) => {
    const isWinEx = isEffectiveExhausted && win.pct <= 0;
    const pacing: WindowPacing = calculateWindowPacing({
      remainingPct: win.pct,
      resetTimeStr: win.reset,
      resetsAt: win.resetsAt,
      resetsInSec: win.resetsInSec,
      windowLabel: win.label,
    });

    return (
      <div key={win.key} className="space-y-2 p-3.5 rounded-[6px] bg-[#0f0f0f] border border-[#222222]">
        {/* Header row: Window Label & Remaining % + Reset Countdown */}
        <div className="flex items-center justify-between text-xs font-sans">
          <div className="flex items-center gap-2">
            <span className="text-[#ededed] font-medium font-sans text-xs">
              {win.label}
            </span>
            <span className={cn('font-mono font-semibold tabular-nums text-xs', getTextColor(win.pct))}>
              {win.pct}% remaining
            </span>
          </div>
          {win.reset && (
            <span className="inline-flex items-center gap-1 text-[11px] font-mono text-muted-foreground">
              <Clock className="h-3 w-3 text-muted-foreground/80 shrink-0" />
              <span>Resets in {win.reset}</span>
            </span>
          )}
        </div>

        {/* Segmented-style track with CodexBar Pace Marker */}
        <div
          role="progressbar"
          aria-valuenow={win.pct}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${win.label} quota`}
          className={cn(
            "relative w-full h-2 rounded-full overflow-hidden border",
            isWinEx
              ? "bg-status-danger-bg border-status-danger-border"
              : "bg-[#181818] border-[#262626]/50"
          )}
        >
          {/* Actual Remaining Capacity Fill */}
          <div
            className={cn(
              "h-full transition-all duration-300 rounded-full",
              isWinEx ? "w-full bg-status-danger-solid/40" : getStatusColor(win.pct)
            )}
            style={{ width: isWinEx ? '100%' : `${Math.min(100, Math.max(0, win.pct))}%` }}
          />

          {/* Expected Pace Marker Tick (CodexBar visual parity) */}
          {!isWinEx &&
            pacing.timeUntilResetSeconds &&
            pacing.expectedRemainingPercent > 0 &&
            pacing.expectedRemainingPercent < 100 && (
              <div
                title={`Expected pace at this time: ${Math.round(pacing.expectedRemainingPercent)}% remaining`}
                className={cn(
                  "absolute top-0 bottom-0 w-[2px] z-10 -ml-[1px] transition-all duration-300",
                  pacing.stage === 'deficit'
                    ? "bg-status-danger-solid shadow-[0_0_4px_rgba(239,68,68,0.8)]"
                    : "bg-white/80 shadow-[0_0_3px_rgba(255,255,255,0.6)]"
                )}
                style={{ left: `${Math.min(100, Math.max(0, pacing.expectedRemainingPercent))}%` }}
              />
            )}
        </div>

        {/* Intelligence Line: CodexBar Deficit/Reserve Pacing & Runout Projection */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1 text-[11px] pt-0.5">
          <div className="flex items-center gap-1.5 font-mono">
            {pacing.stage === 'exhausted' ? (
              <AlertCircle className="h-3 w-3 text-status-danger shrink-0" />
            ) : pacing.stage === 'deficit' ? (
              <TrendingDown className="h-3 w-3 text-status-warning shrink-0" />
            ) : pacing.stage === 'reserve' ? (
              <TrendingUp className="h-3 w-3 text-status-success shrink-0" />
            ) : (
              <CheckCircle2 className="h-3 w-3 text-muted-foreground shrink-0" />
            )}

            <span
              className={cn(
                'font-medium whitespace-nowrap',
                pacing.stage === 'exhausted'
                  ? 'text-status-danger'
                  : pacing.stage === 'deficit'
                  ? 'text-status-warning'
                  : pacing.stage === 'reserve'
                  ? 'text-status-success'
                  : 'text-foreground'
              )}
            >
              {pacing.paceLabel}
            </span>
            <span className="text-muted-foreground/40">·</span>
            <span className="text-muted-foreground font-normal whitespace-nowrap">
              {pacing.projectionLabel}
            </span>
          </div>

          {pacing.windowDateRange && (
            <span
              className="text-[10px] font-mono text-muted-foreground/60 truncate whitespace-nowrap"
              title={pacing.windowDateRange}
            >
              {pacing.windowDateRange}
            </span>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="rounded-geist border border-[#2a2a2a] bg-[#0c0c0c] overflow-hidden shadow-sm">
      {/* 1. UNIFIED CARD HEADER: Engine badge, Status Pill, Account info, Refresh time */}
      <div className="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[#0a0a0a]">
        <div className="flex items-center gap-2.5 flex-wrap">
          {agent && <AgentBadge agent={agent} />}
          <Badge
            variant="outline"
            className={cn(
              'text-[11px] font-sans font-medium tracking-tight flex items-center gap-1.5 px-2.5 py-0.5 rounded-full',
              isEffectiveExhausted
                ? 'badge-danger border-status-danger-border bg-status-danger-bg text-status-danger'
                : isWarning
                ? 'badge-warning border-status-warning-border bg-status-warning-bg text-status-warning'
                : 'badge-success border-status-success-border bg-status-success-bg text-status-success'
            )}
          >
            <span
              className={cn(
                'h-1.5 w-1.5 rounded-full',
                isEffectiveExhausted
                  ? 'bg-status-danger-solid animate-pulse'
                  : isWarning
                  ? 'bg-status-warning-solid'
                  : 'bg-status-success-solid'
              )}
            />
            <span>
              {isEffectiveExhausted
                ? 'Exhausted'
                : isWarning
                ? `Throttled (${effectiveBottleneckPct}%)`
                : `Optimal (${effectiveBottleneckPct}%)`}
            </span>
          </Badge>
        </div>

        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs font-sans text-muted-foreground">
          {account?.email && (
            <div className="flex items-center gap-1.5 text-foreground font-medium" title={account.email}>
              <Mail className="h-3.5 w-3.5 text-muted-foreground/70 shrink-0" />
              <span className="truncate max-w-[200px]">{account.email}</span>
            </div>
          )}
          {account?.auth_method && (
            <div className="flex items-center gap-1 text-muted-foreground/80" title={account.auth_method}>
              <ShieldCheck className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
              <span className="truncate max-w-[180px]">{cleanAuthMethod(account.auth_method)}</span>
            </div>
          )}
          {sessionCount !== undefined && (
            <div className="text-[11px] font-mono text-muted-foreground/70">
              {sessionCount} session{sessionCount === 1 ? '' : 's'}
            </div>
          )}
        </div>
      </div>

      <div className="border-t border-[#1f1f1f]" />

      {/* 2. QUOTA METERS & INSIGHTS BODY */}
      <div className="p-4 sm:p-5 space-y-4">
        {(!parsed || parsed.type === 'fallback') && (
          <div className="space-y-3">
            {renderMeter({
              key: 'telemetry',
              label: 'Overall Capacity Window',
              pct: effectiveBottleneckPct,
              reset: cardResetTime,
            })}
            {activeSummary && (
              <p className="text-[11px] font-mono text-muted-foreground/70 pt-1">
                {activeSummary}
              </p>
            )}
          </div>
        )}

        {parsed?.type === 'single' && parsed.single && (
          <div className="space-y-3">
            {renderMeter(parsed.single)}
          </div>
        )}

        {parsed?.type === 'multi-model' && parsed.models && (
          <div className="space-y-4">
            {parsed.models.map((m) => (
              <div key={m.name} className="space-y-2.5">
                <div className="flex items-center justify-between text-xs font-sans border-b border-[#1c1c1c] pb-1.5">
                  <span className="text-[#ededed] font-medium flex items-center gap-1.5">
                    <span className="h-1.5 w-1.5 rounded-full bg-[#555555]" />
                    {m.name}
                  </span>
                  <span className="text-[11px] text-[#666666] font-mono">
                    {m.windows.length} Time Window{m.windows.length === 1 ? '' : 's'} Monitored
                  </span>
                </div>
                <div
                  className={cn(
                    'grid gap-2.5',
                    m.windows.length === 1 ? 'grid-cols-1' : 'grid-cols-1 sm:grid-cols-2'
                  )}
                >
                  {m.windows.map((w) => renderMeter(w))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 3. INTEGRATED BOTTOM STATUS STRIP (No disjointed floating alerts) */}
      <div className="border-t border-[#1a1a1a] bg-[#080808]/70 px-4 sm:px-5 py-2.5 flex items-center justify-between gap-3 text-xs">
        <div className="flex items-center gap-2 min-w-0">
          {isEffectiveExhausted ? (
            <>
              <AlertCircle className="h-3.5 w-3.5 text-status-danger shrink-0" />
              <span className="text-status-danger font-medium font-sans truncate">
                Rate limit window exhausted {cardResetTime ? `· Quota resets in ${cardResetTime}` : ''}
              </span>
            </>
          ) : isWarning ? (
            <>
              <AlertTriangle className="h-3.5 w-3.5 text-status-warning shrink-0" />
              <span className="text-status-warning font-medium font-sans truncate">
                High quota consumption detected · {effectiveBottleneckPct}% headroom remaining
              </span>
            </>
          ) : (
            <>
              <CheckCircle2 className="h-3.5 w-3.5 text-status-success shrink-0" />
              <span className="text-muted-foreground font-sans truncate">
                All quota windows operating within normal limits
              </span>
            </>
          )}
        </div>

        <div className="flex items-center gap-3 shrink-0">
          {quota?.credits && quota.credits !== '0' && quota.credits !== '' && (
            <span className="inline-flex items-center gap-1 text-[11px] font-mono text-blue-400">
              <Sparkles className="h-3 w-3" />
              <span>{quota.credits} reset credits available</span>
            </span>
          )}
          <Badge
            variant="outline"
            className={cn(
              'font-mono text-[10px] font-medium tracking-wider uppercase px-2 py-0.5 rounded-[4px]',
              isEffectiveExhausted
                ? 'border-status-danger-border text-status-danger bg-status-danger-bg'
                : isWarning
                ? 'border-status-warning-border text-status-warning bg-status-warning-bg'
                : 'border-status-success-border text-status-success bg-status-success-bg'
            )}
          >
            {isEffectiveExhausted ? 'Exhausted' : isWarning ? 'Throttled' : 'Optimal'}
          </Badge>
        </div>
      </div>
    </div>
  );
}
