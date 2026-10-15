import * as React from 'react';
import { PlusCircle, X } from 'lucide-react';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Checkbox } from './ui/checkbox';
import { Input } from './ui/input';
import { cn } from '@/lib/utils';

export interface FacetedOption {
  label: string;
  value: string;
  icon?: React.ComponentType<{ className?: string }>;
}

interface DataTableFacetedFilterProps {
  title: string;
  options: FacetedOption[];
  selectedValues: Set<string>;
  onSelect: (values: Set<string>) => void;
  counts?: Record<string, number>;
  contentClassName?: string;
}

export const DataTableFacetedFilter: React.FC<DataTableFacetedFilterProps> = ({
  title,
  options,
  selectedValues,
  onSelect,
  counts,
  contentClassName,
}) => {
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState('');

  const filteredOptions = React.useMemo(() => {
    if (!search.trim()) return options;
    const query = search.toLowerCase();
    return options.filter((opt) => opt.label.toLowerCase().includes(query) || opt.value.toLowerCase().includes(query));
  }, [options, search]);

  const toggleOption = (val: string) => {
    const next = new Set(selectedValues);
    if (next.has(val)) {
      next.delete(val);
    } else {
      next.add(val);
    }
    onSelect(next);
  };

  const clearAll = () => {
    onSelect(new Set());
  };

  const selectedCount = selectedValues.size;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            'h-8 border-border bg-[#0a0a0a] text-xs font-mono px-2.5 hover:bg-[#161616] hover:text-foreground hover:border-[#333333]',
            selectedCount > 0 && 'border-[#444444] bg-[#111111]'
          )}
        >
          <PlusCircle className="mr-1.5 h-3.5 w-3.5 text-muted-foreground" />
          <span className="font-sans text-xs text-foreground/90">{title}</span>
          {selectedCount > 0 && (
            <>
              <div className="mx-1.5 h-3.5 w-[1px] bg-border" />
              <Badge
                variant="secondary"
                className="h-4 px-1.5 text-[10px] font-mono bg-[#262626] text-[#ededed] border-0 rounded-[4px]"
              >
                {selectedCount}
              </Badge>
            </>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className={cn('w-56 p-2 bg-[#111111] border-border', contentClassName)} align="start">
        <div className="space-y-2">
          {options.length > 5 && (
            <div className="px-1 pt-1">
              <Input
                placeholder={`Filter ${title.toLowerCase()}...`}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="h-7 text-xs bg-[#0a0a0a] border-border focus-visible:ring-1 focus-visible:ring-[#ededed]"
              />
            </div>
          )}
          <div className="max-h-56 overflow-y-auto space-y-0.5 py-1">
            {filteredOptions.length === 0 ? (
              <div className="py-3 text-center text-xs text-muted-foreground font-mono">No results found.</div>
            ) : (
              filteredOptions.map((opt) => {
                const isSelected = selectedValues.has(opt.value);
                const count = counts?.[opt.value];
                return (
                  <div
                    key={opt.value}
                    title={opt.value === '<none>' ? opt.label : opt.value}
                    role="menuitemcheckbox"
                    aria-checked={isSelected}
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        toggleOption(opt.value);
                      }
                    }}
                    onClick={() => toggleOption(opt.value)}
                    className={cn(
                      'flex items-center justify-between px-2 py-1.5 rounded-[4px] text-xs cursor-pointer select-none transition-colors focus:outline-none focus-visible:ring-1 focus-visible:ring-ring',
                      isSelected
                        ? 'bg-white/[0.08] text-foreground font-medium'
                        : 'text-foreground/80 hover:bg-secondary hover:text-foreground'
                    )}
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Checkbox
                        checked={isSelected}
                        tabIndex={-1}
                        aria-hidden="true"
                        className="pointer-events-none"
                      />
                      {opt.icon && <opt.icon className="h-3.5 w-3.5 text-muted-foreground shrink-0" />}
                      <span className="truncate font-mono text-xs text-foreground/90">{opt.label}</span>
                    </div>
                    {count !== undefined && (
                      <span className="ml-auto pl-2 text-[10px] font-mono text-muted-foreground">{count}</span>
                    )}
                  </div>
                );
              })
            )}
          </div>
          {selectedCount > 0 && (
            <div className="pt-1.5 border-t border-border flex items-center justify-between">
              <span className="text-[10px] font-mono text-muted-foreground pl-1">{selectedCount} selected</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearAll}
                className="h-6 px-2 text-[11px] font-mono text-muted-foreground hover:text-foreground hover:bg-[#1f1f1f]"
              >
                <X className="mr-1 h-3 w-3" /> Clear
              </Button>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
};
