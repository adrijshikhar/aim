import { cn } from '@/lib/utils';

export interface GeistSpinnerProps {
  size?: number | 'sm' | 'md' | 'lg';
  variant?: 'spoke' | 'ring';
  className?: string;
  label?: string;
}

export function GeistSpinner({
  size = 16,
  variant = 'spoke',
  className,
  label,
}: GeistSpinnerProps) {
  const pixelSize =
    typeof size === 'number'
      ? size
      : size === 'sm'
      ? 14
      : size === 'lg'
      ? 24
      : 16;

  const spinnerNode =
    variant === 'ring' ? (
      <span
        style={{ width: pixelSize, height: pixelSize }}
        className={cn(
          'inline-block rounded-full border-2 border-white/15 border-t-white animate-spin shrink-0',
          className
        )}
        role="status"
        aria-label={label || 'Loading'}
      />
    ) : (
      <div
        className={cn('geist-spinner inline-block relative text-[#ededed] shrink-0', className)}
        style={{ width: pixelSize, height: pixelSize }}
        role="status"
        aria-label={label || 'Loading'}
      >
        {Array.from({ length: 12 }).map((_, i) => (
          <div
            key={i}
            className="geist-spinner-bar"
            style={{
              transform: `rotate(${i * 30}deg) translate(0, -135%)`,
              animationDelay: `${-1.2 + i * 0.1}s`,
            }}
          />
        ))}
      </div>
    );

  if (!label) {
    return spinnerNode;
  }

  return (
    <div className="inline-flex items-center gap-2 text-xs font-sans text-muted-foreground/80 select-none">
      {spinnerNode}
      <span className="truncate">{label}</span>
    </div>
  );
}
