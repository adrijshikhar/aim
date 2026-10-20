import * as React from 'react';
import * as SliderPrimitive from '@radix-ui/react-slider';
import { cn } from '@/lib/utils';

export interface SliderProps
  extends React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root> {
  thumbClassName?: string;
  trackClassName?: string;
  rangeClassName?: string;
  thumbChildren?: React.ReactNode;
}

const Slider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  SliderProps
>(
  (
    {
      className,
      thumbClassName,
      trackClassName,
      rangeClassName,
      thumbChildren,
      ...props
    },
    ref
  ) => (
    <SliderPrimitive.Root
      ref={ref}
      className={cn(
        'relative flex w-full touch-none select-none items-center cursor-pointer',
        className
      )}
      {...props}
    >
      <SliderPrimitive.Track
        className={cn(
          'relative h-2 w-full grow overflow-hidden rounded-full bg-[#1e1e1e] border border-[#2e2e2e]',
          trackClassName
        )}
      >
        <SliderPrimitive.Range
          className={cn('absolute h-full bg-[#3b82f6]', rangeClassName)}
        />
      </SliderPrimitive.Track>
      <SliderPrimitive.Thumb
        className={cn(
          'block h-5 w-5 rounded-full border border-[#444444] bg-white shadow-md transition-shadow hover:scale-105 active:scale-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#61afef] focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 cursor-grab active:cursor-grabbing flex items-center justify-center',
          thumbClassName
        )}
      >
        {thumbChildren}
      </SliderPrimitive.Thumb>
    </SliderPrimitive.Root>
  )
);
Slider.displayName = SliderPrimitive.Root.displayName;

export { Slider };
