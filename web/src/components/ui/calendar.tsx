import * as React from 'react';
import { DayPicker } from 'react-day-picker';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

function Calendar({
  className,
  classNames,
  showOutsideDays = true,
  ...props
}: CalendarProps) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn('p-3 bg-[#111317] rounded-geist border border-border text-foreground', className)}
      classNames={{
        months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
        month: 'space-y-4',
        month_caption: 'flex justify-center pt-1 relative items-center text-xs font-semibold text-foreground',
        caption_label: 'text-xs font-medium text-foreground',
        nav: 'space-x-1 flex items-center justify-between absolute w-full px-1',
        button_previous: cn(
          buttonVariants({ variant: 'outline' }),
          'h-7 w-7 bg-secondary p-0 opacity-70 hover:opacity-100 border-border text-foreground'
        ),
        button_next: cn(
          buttonVariants({ variant: 'outline' }),
          'h-7 w-7 bg-secondary p-0 opacity-70 hover:opacity-100 border-border text-foreground'
        ),
        month_grid: 'w-full border-collapse space-y-1',
        weekdays: 'flex',
        weekday: 'text-muted-foreground rounded-md w-8 font-normal text-[0.8rem] text-center',
        week: 'flex w-full mt-2',
        day: cn(
          buttonVariants({ variant: 'ghost' }),
          'h-8 w-8 p-0 font-normal aria-selected:opacity-100 text-foreground hover:bg-muted rounded-md text-xs'
        ),
        range_start: 'day-range-start bg-primary text-primary-foreground rounded-l-md',
        range_end: 'day-range-end bg-primary text-primary-foreground rounded-r-md',
        selected:
          'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground font-semibold',
        today: 'bg-secondary text-accent-cyan font-bold',
        outside: 'text-muted-foreground/40 aria-selected:bg-accent/50 aria-selected:text-muted-foreground',
        disabled: 'text-muted-foreground/20 cursor-not-allowed',
        range_middle: 'aria-selected:bg-secondary aria-selected:text-foreground rounded-none',
        hidden: 'invisible',
        ...classNames,
      }}
      {...props}
    />
  );
}
Calendar.displayName = 'Calendar';

export { Calendar };
