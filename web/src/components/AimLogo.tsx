import * as React from 'react';
import { cn } from '@/lib/utils';

export type AimLogoVariant = 'core-void' | 'nexus' | 'monogram' | 'aperture' | 'sandboxes' | 'dart';

export interface AimLogoProps extends React.SVGProps<SVGSVGElement> {
  className?: string;
  variant?: AimLogoVariant;
}

export function AimLogo({ className, variant = 'dart', ...props }: AimLogoProps) {
  const maskId = React.useId();

  switch (variant) {
    case 'core-void':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Core Void"
          {...props}
        >
          <mask id={`void-mask-${maskId}`}>
            <rect width="100" height="100" fill="#ffffff" />
            <line x1="50" y1="10" x2="50" y2="28" stroke="#000000" strokeWidth="5" strokeLinecap="round" />
          </mask>
          {/* Weighted outer ring with precision meridian slit mask */}
          <circle cx="50" cy="50" r="34" stroke="currentColor" strokeWidth="7" mask={`url(#void-mask-${maskId})`} />
          {/* Anchoring center core */}
          <circle cx="50" cy="50" r="14" fill="currentColor" />
        </svg>
      );

    case 'nexus':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Tri-Agent Nexus"
          {...props}
        >
          <line x1="50" y1="24" x2="50" y2="50" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
          <line x1="26" y1="68" x2="50" y2="50" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
          <line x1="74" y1="68" x2="50" y2="50" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
          <circle cx="50" cy="24" r="8" fill="currentColor" />
          <circle cx="26" cy="68" r="7" fill="currentColor" />
          <circle cx="74" cy="68" r="7" fill="currentColor" />
          <circle cx="50" cy="50" r="5" stroke="currentColor" strokeWidth="2.5" fill="var(--background, #000)" />
        </svg>
      );

    case 'monogram':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Architectural Monogram"
          {...props}
        >
          <path d="M22 74L50 22L78 74" stroke="currentColor" strokeWidth="7.5" strokeLinecap="round" strokeLinejoin="round" />
          <line x1="32" y1="56" x2="68" y2="56" stroke="currentColor" strokeWidth="5.5" strokeLinecap="round" />
        </svg>
      );

    case 'aperture':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Radial Aperture"
          {...props}
        >
          <g stroke="currentColor" strokeWidth="5" strokeLinecap="round">
            <line x1="50" y1="16" x2="66" y2="34" />
            <line x1="79" y1="33" x2="71" y2="57" />
            <line x1="79" y1="67" x2="55" y2="73" />
            <line x1="50" y1="84" x2="34" y2="66" />
            <line x1="21" y1="67" x2="29" y2="43" />
            <line x1="21" y1="33" x2="45" y2="27" />
          </g>
          <circle cx="50" cy="50" r="6" fill="currentColor" />
        </svg>
      );

    case 'sandboxes':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Dual Sandboxes"
          {...props}
        >
          <circle cx="39" cy="50" r="23" stroke="currentColor" strokeWidth="5.5" />
          <circle cx="61" cy="50" r="23" stroke="currentColor" strokeWidth="5.5" />
          <circle cx="50" cy="50" r="4.5" fill="currentColor" />
        </svg>
      );

    case 'dart':
      return (
        <svg
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn('shrink-0 select-none', className)}
          aria-label="AIM Logo — Forward Dart"
          {...props}
        >
          <path d="M78 22L28 42L48 52L58 72L78 22Z" fill="currentColor" />
          <circle cx="28" cy="72" r="5" fill="currentColor" />
        </svg>
      );
  }
}

export default AimLogo;
