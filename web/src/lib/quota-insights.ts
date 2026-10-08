/**
 * Quota Insights and Pacing Intelligence Engine
 * 
 * Ported with mathematical fidelity from CodexBar's UsagePace and
 * RateWindow projection algorithms. Calculates:
 * - Real-time consumption vs elapsed window budget (% in deficit / in reserve)
 * - Projected runout ETA ("Runs out in Xd Yh" / "Lasts until reset")
 * - Headroom speed multiplier (e.g. "1.5× headroom")
 * - Estimated window start/end datetime boundaries
 */

export type PaceStage = 'onTrack' | 'deficit' | 'reserve' | 'exhausted';

export interface WindowPacing {
  stage: PaceStage;
  deltaPercent: number; // raw difference (positive = deficit, negative = reserve)
  deltaValue: number;   // absolute rounded integer
  expectedUsedPercent: number; // 0..100
  actualUsedPercent: number;   // 0..100
  expectedRemainingPercent: number; // 0..100
  paceLabel: string;           // "7% in deficit" | "14% in reserve" | "On pace"
  projectionLabel: string;     // "Runs out in 3d 18h" | "Lasts until reset" | "Resets in 1d 12h"
  combinedSummary: string;     // "7% in deficit · Runs out in 3d 18h"
  willLastToReset: boolean;
  headroomMultiplier?: number;
  expectedMarkerPct: number;   // 0..100 marker position on the remaining-capacity track
  timeUntilResetSeconds?: number;
  windowDateRange?: string;    // "Est. Oct 7 at 9:08 AM – Oct 14 at 9:08 AM"
}

/**
 * Parses human durations like "5d 22h", "1d 12h", "3h 5m", "45m", "14m", "2h", "1d" into total seconds.
 */
export function parseDurationStringToSeconds(str?: string | null): number | null {
  if (!str || !str.trim()) return null;
  const s = str.trim();
  let totalSeconds = 0;
  let matched = false;

  const days = s.match(/(\d+)\s*d/i);
  if (days) {
    totalSeconds += parseInt(days[1], 10) * 86400;
    matched = true;
  }

  const hours = s.match(/(\d+)\s*h/i);
  if (hours) {
    totalSeconds += parseInt(hours[1], 10) * 3600;
    matched = true;
  }

  const mins = s.match(/(\d+)\s*m/i);
  if (mins) {
    totalSeconds += parseInt(mins[1], 10) * 60;
    matched = true;
  }

  const secs = s.match(/(\d+)\s*s/i);
  if (secs) {
    totalSeconds += parseInt(secs[1], 10);
    matched = true;
  }

  return matched && totalSeconds > 0 ? totalSeconds : null;
}

/**
 * Formats seconds into human duration like "3d 18h", "2h 15m", "45m".
 */
export function formatSecondsToDuration(seconds: number): string {
  if (!isFinite(seconds) || seconds <= 0) return 'now';
  const totalMinutes = Math.round(seconds / 60);
  const days = Math.floor(totalMinutes / (24 * 60));
  const hours = Math.floor((totalMinutes % (24 * 60)) / 60);
  const mins = totalMinutes % 60;

  if (days > 0) {
    return hours > 0 ? `${days}d ${hours}h` : `${days}d`;
  }
  if (hours > 0) {
    return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
  }
  return `${Math.max(1, mins)}m`;
}

function formatShortDate(d: Date): string {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const month = months[d.getMonth()];
  const day = d.getDate();
  let hours = d.getHours();
  const ampm = hours >= 12 ? 'PM' : 'AM';
  hours = hours % 12;
  hours = hours ? hours : 12; // 0 -> 12
  const mins = d.getMinutes().toString().padStart(2, '0');
  return `${month} ${day} at ${hours}:${mins} ${ampm}`;
}

/**
 * Calculates CodexBar-grade pacing telemetry for a given rate window.
 */
export function calculateWindowPacing({
  remainingPct,
  resetTimeStr,
  resetsAt,
  resetsInSec,
  windowLabel = '',
  now = new Date(),
}: {
  remainingPct: number;
  resetTimeStr?: string | null;
  resetsAt?: string | Date | null;
  resetsInSec?: number | null;
  windowLabel?: string;
  now?: Date;
}): WindowPacing {
  const clampedRemaining = Math.max(0, Math.min(100, remainingPct));
  const actualUsedPercent = 100 - clampedRemaining;

  // 1. Determine total window duration in seconds:
  const lowerLabel = windowLabel.toLowerCase();
  let defaultDurationSec = 7 * 86400; // Weekly default (604,800s)
  if (lowerLabel.includes('5h') || lowerLabel.includes('five') || lowerLabel.includes('5-hour') || lowerLabel.includes('5 hour')) {
    defaultDurationSec = 5 * 3600; // 18,000s
  } else if (lowerLabel.includes('day') || lowerLabel.includes('daily') || lowerLabel.includes('24h')) {
    defaultDurationSec = 86400;
  }

  // 2. Resolve time until reset in seconds:
  let timeUntilReset: number | null = null;
  let resetDate: Date | null = null;

  if (resetsAt) {
    const d = typeof resetsAt === 'string' ? new Date(resetsAt) : resetsAt;
    if (!isNaN(d.getTime())) {
      resetDate = d;
      timeUntilReset = Math.max(0, Math.round((d.getTime() - now.getTime()) / 1000));
    }
  }

  if (timeUntilReset === null && typeof resetsInSec === 'number' && resetsInSec > 0) {
    timeUntilReset = resetsInSec;
    resetDate = new Date(now.getTime() + resetsInSec * 1000);
  }

  if (timeUntilReset === null && resetTimeStr) {
    const parsed = parseDurationStringToSeconds(resetTimeStr);
    if (parsed !== null) {
      timeUntilReset = parsed;
      resetDate = new Date(now.getTime() + parsed * 1000);
    }
  }

  // If timeUntilReset exceeds default duration, dynamically expand duration
  let durationSec = defaultDurationSec;
  if (timeUntilReset !== null && timeUntilReset > durationSec) {
    durationSec = timeUntilReset <= 86400 ? 86400 : 7 * 86400;
    if (timeUntilReset > durationSec) {
      durationSec = timeUntilReset;
    }
  }

  // Default fallback if no timing data is available
  if (timeUntilReset === null || timeUntilReset <= 0) {
    if (clampedRemaining <= 0) {
      return {
        stage: 'exhausted',
        deltaPercent: 100,
        deltaValue: 100,
        expectedUsedPercent: 100,
        actualUsedPercent: 100,
        expectedRemainingPercent: 0,
        paceLabel: 'Rate limit reached',
        projectionLabel: resetTimeStr ? `Resets in ${resetTimeStr}` : 'Exhausted',
        combinedSummary: resetTimeStr ? `Rate limit reached · Resets in ${resetTimeStr}` : 'Rate limit reached',
        willLastToReset: false,
        expectedMarkerPct: 0,
      };
    }

    return {
      stage: 'onTrack',
      deltaPercent: 0,
      deltaValue: 0,
      expectedUsedPercent: actualUsedPercent,
      actualUsedPercent,
      expectedRemainingPercent: clampedRemaining,
      paceLabel: 'Operating normally',
      projectionLabel: 'Within quota limits',
      combinedSummary: 'Operating within normal limits',
      willLastToReset: true,
      expectedMarkerPct: clampedRemaining,
    };
  }

  // 3. Compute elapsed time & linear expected usage:
  const elapsedSec = Math.max(0, durationSec - timeUntilReset);
  const expectedUsedPercent = Math.min(100, Math.max(0, (elapsedSec / durationSec) * 100));
  const expectedRemainingPercent = 100 - expectedUsedPercent;

  // Delta: actual used - expected used
  // Positive = deficit (used more than elapsed budget)
  // Negative = reserve (used less than elapsed budget)
  const deltaPercent = actualUsedPercent - expectedUsedPercent;
  const deltaValue = Math.round(Math.abs(deltaPercent));

  let stage: PaceStage = 'onTrack';
  let paceLabel = 'On pace';

  if (clampedRemaining <= 0) {
    stage = 'exhausted';
    paceLabel = 'Rate limit reached';
  } else if (deltaValue <= 2) {
    stage = 'onTrack';
    paceLabel = 'On pace';
  } else if (deltaPercent > 0) {
    stage = 'deficit';
    paceLabel = `${deltaValue}% in deficit`;
  } else {
    stage = 'reserve';
    paceLabel = `${deltaValue}% in reserve`;
  }

  // 4. Runout and ETA projection:
  let willLastToReset = false;
  let projectionLabel = '';
  let headroomMultiplier: number | undefined;

  if (clampedRemaining <= 0) {
    willLastToReset = false;
    const dur = formatSecondsToDuration(timeUntilReset);
    projectionLabel = `Resets in ${dur}`;
  } else if (actualUsedPercent <= 0) {
    willLastToReset = true;
    projectionLabel = 'Lasts until reset';
  } else if (elapsedSec > 0) {
    const rate = actualUsedPercent / elapsedSec; // % consumed per second
    if (rate > 0) {
      const candidateSeconds = clampedRemaining / rate;
      if (candidateSeconds >= timeUntilReset) {
        willLastToReset = true;
        const projectedUsageAtReset = actualUsedPercent * (timeUntilReset / elapsedSec);
        if (projectedUsageAtReset > 0) {
          const mult = clampedRemaining / projectedUsageAtReset;
          if (isFinite(mult) && mult >= 1.5) {
            headroomMultiplier = mult;
            projectionLabel = `Lasts until reset (${mult.toFixed(1)}× headroom)`;
          } else {
            projectionLabel = 'Lasts until reset';
          }
        } else {
          projectionLabel = 'Lasts until reset';
        }
      } else {
        willLastToReset = false;
        const dur = formatSecondsToDuration(candidateSeconds);
        projectionLabel = `Runs out in ${dur}`;
      }
    } else {
      willLastToReset = true;
      projectionLabel = 'Lasts until reset';
    }
  } else {
    willLastToReset = true;
    projectionLabel = 'Lasts until reset';
  }

  const combinedSummary = `${paceLabel} · ${projectionLabel}`;

  // 5. Date range:
  let windowDateRange: string | undefined;
  if (resetDate) {
    const startDate = new Date(resetDate.getTime() - durationSec * 1000);
    windowDateRange = `Est. ${formatShortDate(startDate)} – ${formatShortDate(resetDate)}`;
  }

  return {
    stage,
    deltaPercent,
    deltaValue,
    expectedUsedPercent,
    actualUsedPercent,
    expectedRemainingPercent,
    paceLabel,
    projectionLabel,
    combinedSummary,
    willLastToReset,
    headroomMultiplier,
    expectedMarkerPct: expectedRemainingPercent,
    timeUntilResetSeconds: timeUntilReset,
    windowDateRange,
  };
}
