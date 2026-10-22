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
  'aria-label'?: string;
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
  'aria-label': ariaLabel,
}: SelectDropdownProps<T>) {
  const [open, setOpen] = React.useState(false);
  const [highlightedIndex, setHighlightedIndex] = React.useState(-1);
  const listboxId = React.useId();

  const selectedIndex = React.useMemo(
    () => options.findIndex((opt) => opt.value === value),
    [options, value]
  );

  const selectedOption = selectedIndex >= 0 ? options[selectedIndex] : undefined;

  React.useEffect(() => {
    if (open) {
      setHighlightedIndex(selectedIndex >= 0 ? selectedIndex : 0);
    }
  }, [open, selectedIndex]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (disabled) return;

    if (!open) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        setOpen(true);
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setHighlightedIndex((prev) => (prev + 1 < options.length ? prev + 1 : 0));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((prev) => (prev - 1 >= 0 ? prev - 1 : options.length - 1));
        break;
      case 'Home':
        e.preventDefault();
        setHighlightedIndex(0);
        break;
      case 'End':
        e.preventDefault();
        setHighlightedIndex(options.length - 1);
        break;
      case 'Enter':
      case ' ':
        e.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < options.length) {
          onChange(options[highlightedIndex].value);
          setOpen(false);
        }
        break;
      case 'Escape':
        e.preventDefault();
        setOpen(false);
        break;
      case 'Tab':
        setOpen(false);
        break;
    }
  };

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
            role="combobox"
            aria-haspopup="listbox"
            aria-expanded={open}
            aria-controls={listboxId}
            aria-label={ariaLabel || (typeof selectedOption?.label === 'string' ? selectedOption.label : placeholder)}
            disabled={disabled}
            onKeyDown={handleKeyDown}
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
          onKeyDown={handleKeyDown}
          className={cn(
            'w-[var(--radix-popover-trigger-width)] min-w-[140px] max-h-60 overflow-y-auto p-1 bg-[#111111] border border-[#262626] rounded-geist shadow-2xl',
            contentClassName
          )}
        >
          <div
            id={listboxId}
            role="listbox"
            aria-label={ariaLabel || placeholder}
            className="flex flex-col gap-0.5"
          >
            {options.map((opt, index) => {
              const isSelected = opt.value === value;
              const isHighlighted = index === highlightedIndex;
              return (
                <button
                  key={String(opt.value)}
                  type="button"
                  role="option"
                  aria-selected={isSelected}
                  onMouseEnter={() => setHighlightedIndex(index)}
                  onClick={() => {
                    onChange(opt.value);
                    setOpen(false);
                  }}
                  className={cn(
                    'flex w-full items-center justify-between gap-2 px-2.5 py-1.5 text-left text-xs rounded-[4px] transition-colors',
                    isSelected
                      ? 'bg-[#1e1e1e] text-[#ededed] font-medium'
                      : isHighlighted
                        ? 'bg-[#161616] text-[#ededed]'
                        : 'text-[#a1a1a1] hover:bg-[#161616] hover:text-[#ededed]'
                  )}
                >
                  <div className="flex items-center gap-2 min-w-0 truncate">
                    {opt.icon && <span className="shrink-0">{opt.icon}</span>}
                    <div className="flex flex-col truncate">
                      <span className="truncate">{opt.label}</span>
                      {opt.description && (
                        <span className="text-xs text-muted-foreground truncate">
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
