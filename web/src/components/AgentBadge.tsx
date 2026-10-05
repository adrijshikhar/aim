import React from 'react';
import { cn } from '@/lib/utils';

interface AgentBadgeProps {
  agent: string;
  className?: string;
  size?: 'sm' | 'default';
}

const getAgentDotColor = (agent: string) => {
  const a = agent?.toLowerCase() || '';
  switch (a) {
    case 'claude':
      return 'bg-[#be9137]'; // Cool muted bronze
    case 'codex':
      return 'bg-[#47ae8f]'; // Cool slate-pine / sage
    case 'gemini':
      return 'bg-[#6b94d1]'; // Cool slate-blue
    case 'agy':
    case 'antigravity':
      return 'bg-[#8e7dbf]'; // Cool slate-iris
    default:
      return 'bg-muted-foreground';
  }
};

export const AgentBadge: React.FC<AgentBadgeProps> = ({ agent, className = '', size = 'default' }) => {
  const name = agent?.toLowerCase() || 'agent';
  const sizeClasses = size === 'sm' ? 'text-[10px] px-2 py-0.5' : 'text-[11px] px-2.5 py-1';
  const dotSize = size === 'sm' ? 'h-1.5 w-1.5' : 'h-1.5 w-1.5';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 font-mono font-medium tracking-tight rounded-geist border border-border bg-secondary text-foreground select-none transition-colors hover:border-border/80',
        sizeClasses,
        className
      )}
    >
      <span className={cn('rounded-full shrink-0', dotSize, getAgentDotColor(name))} />
      <span>{name}</span>
    </span>
  );
};
