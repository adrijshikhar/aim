import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const badgeVariants = cva(
  'inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-colors focus:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3]',
  {
    variants: {
      variant: {
        default:
          'border-transparent bg-[#ededed] text-[#0a0a0a]',
        secondary:
          'border-[#262626] bg-[#1a1a1a] text-[#ededed]',
        destructive:
          'border-[#ee0000]/30 bg-[#ee0000]/10 text-[#ee0000]',
        outline: 'border-[#262626] text-foreground',
        success:
          'border-emerald-500/25 bg-emerald-500/10 text-emerald-400',
        warning:
          'border-amber-500/25 bg-amber-500/10 text-amber-400',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
