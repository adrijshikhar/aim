import * as React from 'react';
import { cn } from '@/lib/utils';

export interface SwitchProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}

const Switch = React.forwardRef<HTMLButtonElement, SwitchProps>(
  ({ className, checked = false, onCheckedChange, disabled, onClick, ...props }, ref) => {
    const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
      onClick?.(e);
      if (!disabled && onCheckedChange) {
        onCheckedChange(!checked);
      }
    };

    return (
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        disabled={disabled}
        ref={ref}
        onClick={handleClick}
        className={cn(
          'peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-[#2e2e2e] transition-colors duration-200 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] disabled:cursor-not-allowed disabled:opacity-50',
          checked ? 'bg-[#ededed] border-transparent' : 'bg-[#1c1c1c]',
          className
        )}
        {...props}
      >
        <span
          className={cn(
            'pointer-events-none block h-3.5 w-3.5 rounded-full transition-all duration-200',
            checked ? 'translate-x-[18px] bg-[#0a0a0a]' : 'translate-x-[2px] bg-[#666666]'
          )}
        />
      </button>
    );
  }
);
Switch.displayName = 'Switch';

export { Switch };
