import React from 'react';
import { Badge } from './ui/badge';
import { cn } from '@/lib/utils';

interface AgentBadgeProps {
  agent: string;
  className?: string;
  size?: 'sm' | 'default';
}

const getAgentColorStyle = (agent: string) => {
  const a = agent?.toLowerCase() || '';
  switch (a) {
    case 'claude':
      return 'border-[#c08532]/30 bg-[#c08532]/10 text-[#d49746]';
    case 'codex':
      return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400';
    case 'gemini':
      return 'border-sky-500/30 bg-sky-500/10 text-sky-400';
    case 'agy':
    case 'antigravity':
      return 'border-purple-500/30 bg-purple-500/10 text-purple-400';
    default:
      return 'border-border bg-secondary text-foreground';
  }
};

export const AgentBadge: React.FC<AgentBadgeProps> = ({ agent, className = '', size = 'default' }) => {
  const name = agent?.toLowerCase() || 'agent';
  const sizeClasses = size === 'sm' ? 'text-[10px] px-1.5 py-0' : 'text-[11px] px-2 py-0.5';

  return (
    <Badge
      variant="outline"
      className={cn(
        'font-mono font-medium tracking-tight rounded-md transition-colors',
        getAgentColorStyle(name),
        sizeClasses,
        className
      )}
    >
      {name}
    </Badge>
  );
};
