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
        cursor:
          'border-[#f54e00]/30 bg-[#f54e00]/10 text-[#f54e00]',
        thinking:
          'border-[#dfa88f]/30 bg-[#dfa88f]/10 text-[#dfa88f]',
        exec:
          'border-[#9fc9a2]/30 bg-[#9fc9a2]/10 text-[#9fc9a2]',
        read:
          'border-[#9fbbe0]/30 bg-[#9fbbe0]/10 text-[#9fbbe0]',
        edit:
          'border-[#c0a8dd]/30 bg-[#c0a8dd]/10 text-[#c0a8dd]',
        done:
          'border-[#c08532]/30 bg-[#c08532]/10 text-[#c08532]',
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
