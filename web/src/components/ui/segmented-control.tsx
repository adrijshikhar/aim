import * as React from 'react';
import { cn } from '@/lib/utils';

export interface SegmentedControlOption<T extends string | number> {
  value: T;
  label: React.ReactNode;
  icon?: React.ReactNode;
  title?: string;
  dotColor?: string;
  badge?: React.ReactNode;
  activeClassName?: string;
}

export interface SegmentedControlProps<T extends string | number> {
  value: T;
  onChange: (value: T) => void;
  options: SegmentedControlOption<T>[];
  size?: 'sm' | 'default';
  disabled?: boolean;
  className?: string;
  prefix?: React.ReactNode;
}

export function SegmentedControl<T extends string | number>({
  value,
  onChange,
  options,
  size = 'default',
  disabled = false,
  className = '',
  prefix,
}: SegmentedControlProps<T>) {
  const isSm = size === 'sm';
  const containerHeight = isSm ? 'h-7' : 'h-8';
  const itemHeight = isSm ? 'h-6' : 'h-7';
  const itemPadding = isSm ? 'px-2.5' : 'px-3';
  const itemText = isSm ? 'text-[11px]' : 'text-xs';

  return (
    <div
      role="radiogroup"
      className={cn(
        'inline-flex items-center p-0.5 rounded-geist bg-secondary/50 border border-border/70 select-none shadow-xs shrink-0 box-border',
        containerHeight,
        className
      )}
    >
      {prefix && (
        <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground px-2 flex items-center gap-1.5 shrink-0 font-medium">
          {prefix}
        </span>
      )}
      {options.map((opt) => {
        const isSelected = opt.value === value;
        return (
          <button
            key={String(opt.value)}
            type="button"
            role="radio"
            aria-checked={isSelected}
            disabled={disabled}
            onClick={() => onChange(opt.value)}
            title={opt.title}
            className={cn(
              'inline-flex items-center justify-center gap-1.5 font-mono font-medium rounded-[3px] transition-colors cursor-pointer border whitespace-nowrap box-border',
              itemHeight,
              itemPadding,
              itemText,
              isSelected
                ? cn('bg-card text-foreground shadow-xs border-border/80 font-semibold', opt.activeClassName)
                : 'border-transparent text-muted-foreground hover:text-foreground hover:bg-secondary/60',
              disabled && 'opacity-50 cursor-not-allowed pointer-events-none'
            )}
          >
            {opt.dotColor && (
              <span className={cn('h-1.5 w-1.5 rounded-full shrink-0', opt.dotColor)} />
            )}
            {opt.icon && <span className="shrink-0">{opt.icon}</span>}
            <span>{opt.label}</span>
            {opt.badge}
          </button>
        );
      })}
    </div>
  );
}
