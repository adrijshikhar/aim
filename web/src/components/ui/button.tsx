import * as React from 'react';
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-geist text-sm font-medium transition-all duration-150 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[#0070f3] focus-visible:ring-offset-1 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 cursor-pointer',
  {
    variants: {
      variant: {
        default:
          'bg-[#ededed] text-[#0a0a0a] hover:bg-white active:scale-[0.98]',
        destructive:
          'bg-[#ee0000] text-white hover:bg-[#c50000] active:scale-[0.98]',
        outline:
          'border border-[#262626] bg-[#0e0e0e] text-[#ededed] hover:bg-[#171717] hover:border-[#383838] active:scale-[0.98]',
        secondary:
          'bg-[#1a1a1a] text-[#ededed] border border-[#2a2a2a] hover:bg-[#222222] active:scale-[0.98]',
        ghost: 'text-[#888888] hover:text-[#ededed] hover:bg-[#171717]',
        link: 'text-[#0070f3] underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-9 px-4 py-2',
        sm: 'h-8 px-3 text-xs',
        lg: 'h-10 px-6 text-sm',
        icon: 'h-8 w-8',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button';
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = 'Button';

export { Button, buttonVariants };
