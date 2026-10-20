import * as React from 'react';
import { Popover, PopoverContent, PopoverTrigger } from './popover';
import { ChevronDown, Check } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface SelectOption<T = string | number> {
  value: T;
  label: React.ReactNode;
  description?: string;
  badge?: React.ReactNode;
  icon?: React.ReactNode;
}

export interface SelectDropdownProps<T = string | number> {
  value: T;
  onChange: (value: T) => void;
  options: SelectOption<T>[];
  placeholder?: string;
  disabled?: boolean;
  className?: string;
  triggerClassName?: string;
  contentClassName?: string;
  size?: 'xs' | 'sm' | 'default';
  id?: string;
}

export function SelectDropdown<T extends string | number>({
  value,
  onChange,
  options,
  placeholder = 'Select an option...',
  disabled = false,
  className,
  triggerClassName,
  contentClassName,
  size = 'default',
  id,
}: SelectDropdownProps<T>) {
  const [open, setOpen] = React.useState(false);

  const selectedOption = React.useMemo(
    () => options.find((opt) => opt.value === value),
    [options, value]
  );

  const sizeClasses = {
    xs: 'h-7 px-2 text-[11px]',
    sm: 'h-8 px-2.5 text-xs',
    default: 'h-9 px-3 text-xs',
  };

  return (
    <div className={cn('relative inline-block w-full', className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild disabled={disabled}>
          <button
            type="button"
            id={id}
            disabled={disabled}
            className={cn(
              'flex w-full items-center justify-between gap-2 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] transition-colors',
              'hover:border-[#383838] hover:bg-[#121212] focus:outline-none focus:ring-1 focus:ring-[#ededed]/40',
              'disabled:cursor-not-allowed disabled:opacity-50',
              sizeClasses[size],
              triggerClassName
            )}
          >
            <div className="flex items-center gap-2 min-w-0 truncate">
              {selectedOption?.icon && (
                <span className="shrink-0 text-muted-foreground">{selectedOption.icon}</span>
              )}
              <span className="truncate">
                {selectedOption ? selectedOption.label : placeholder}
              </span>
              {selectedOption?.badge && (
                <span className="shrink-0">{selectedOption.badge}</span>
              )}
            </div>
            <ChevronDown
              className={cn(
                'h-3.5 w-3.5 shrink-0 text-[#666666] transition-transform duration-200',
                open && 'rotate-180 text-[#ededed]'
              )}
            />
          </button>
        </PopoverTrigger>
        <PopoverContent
          align="start"
          sideOffset={4}
          className={cn(
            'w-[var(--radix-popover-trigger-width)] min-w-[140px] max-h-60 overflow-y-auto p-1 bg-[#111111] border border-[#262626] rounded-geist shadow-2xl',
            contentClassName
          )}
        >
          <div className="flex flex-col gap-0.5">
            {options.map((opt) => {
              const isSelected = opt.value === value;
              return (
                <button
                  key={String(opt.value)}
                  type="button"
                  onClick={() => {
                    onChange(opt.value);
                    setOpen(false);
                  }}
                  className={cn(
                    'flex w-full items-center justify-between gap-2 px-2.5 py-1.5 text-left text-xs rounded-[4px] transition-colors',
                    isSelected
                      ? 'bg-[#1e1e1e] text-[#ededed] font-medium'
                      : 'text-[#a1a1a1] hover:bg-[#161616] hover:text-[#ededed]'
                  )}
                >
                  <div className="flex items-center gap-2 min-w-0 truncate">
                    {opt.icon && <span className="shrink-0">{opt.icon}</span>}
                    <div className="flex flex-col truncate">
                      <span className="truncate">{opt.label}</span>
                      {opt.description && (
                        <span className="text-[10px] text-[#666666] truncate">
                          {opt.description}
                        </span>
                      )}
                    </div>
                    {opt.badge && <span className="shrink-0">{opt.badge}</span>}
                  </div>
                  {isSelected && (
                    <Check className="h-3.5 w-3.5 shrink-0 text-[#ededed]" />
                  )}
                </button>
              );
            })}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
