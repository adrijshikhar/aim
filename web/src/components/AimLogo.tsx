import * as React from 'react';
import { cn } from '@/lib/utils';

export interface AimLogoProps extends React.SVGProps<SVGSVGElement> {
  className?: string;
  variant?: 'monochrome' | 'laser';
}

export function AimLogo({ className, variant = 'laser', ...props }: AimLogoProps) {
  const accentColor = variant === 'laser' ? '#00f0ff' : 'currentColor';

  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn('shrink-0 select-none', className)}
      aria-label="AIM Logo"
      {...props}
    >
      {/* 90-degree Navigational Compass Arc */}
      <path
        d="M4 20A16 16 0 0 1 20 4"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
      />

      {/* Trajectory Vector Arrow pointing Northeast */}
      <path
        d="M12 4H20V12"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />

      {/* Trajectory Vector Ray */}
      <line
        x1="8"
        y1="16"
        x2="19"
        y2="5"
        stroke={accentColor}
        strokeWidth="1.8"
        strokeLinecap="round"
      />

      {/* Pivot Origin Dot */}
      <circle
        cx="8"
        cy="16"
        r="1.8"
        fill={accentColor}
      />
    </svg>
  );
}

export default AimLogo;
