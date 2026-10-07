import * as React from 'react';
import { cn } from '@/lib/utils';

export interface AimLogoProps extends React.SVGProps<SVGSVGElement> {
  className?: string;
  variant?: 'monochrome' | 'laser' | 'filled';
}

export function AimLogo({ className, variant = 'monochrome', ...props }: AimLogoProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn('shrink-0 select-none', className)}
      aria-label="AIM Logo"
      {...props}
    >
      {/* Precision Reticle Ring */}
      <circle
        cx="12"
        cy="12"
        r="7.5"
        stroke={variant === 'laser' ? '#00f0ff' : 'currentColor'}
        strokeWidth="1.25"
        strokeOpacity={variant === 'laser' ? 0.9 : 0.65}
      />

      {/* Crosshair Alignment Ticks */}
      <path
        d="M12 1.5V4M12 20V22.5M1.5 12H4M20 12H22.5"
        stroke={variant === 'laser' ? '#00f0ff' : 'currentColor'}
        strokeWidth="1.5"
        strokeLinecap="round"
      />

      {/* Precision Reticle Center Pinpoint */}
      <circle
        cx="12"
        cy="11.5"
        r="0.75"
        fill={variant === 'laser' ? '#00f0ff' : 'currentColor'}
      />

      {/* Faceted Geometric 'A' Mark with Transparent Aperture */}
      <path
        d="M12 3.2L19.2 19.8H14.8L13.2 16H10.8L9.2 19.8H4.8L12 3.2ZM12 8.2L13.6 13H10.4L12 8.2Z"
        fill="currentColor"
        fillRule="evenodd"
        clipRule="evenodd"
      />

      {/* Facet Detail (Prism Chevron) */}
      <path
        d="M10.4 14.5L12 17L13.6 14.5L12 15.4L10.4 14.5Z"
        fill={variant === 'laser' ? '#00f0ff' : 'currentColor'}
        fillOpacity={variant === 'laser' ? 1 : 0.45}
      />
    </svg>
  );
}

export default AimLogo;
