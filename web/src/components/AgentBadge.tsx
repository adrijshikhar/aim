import React from 'react';
import { Badge } from './ui/badge';

interface AgentBadgeProps {
  agent: string;
  className?: string;
  size?: 'sm' | 'default';
}

export const AgentBadge: React.FC<AgentBadgeProps> = ({ agent, className = '', size = 'default' }) => {
  const name = agent?.toLowerCase() || 'agent';
  const sizeClasses = size === 'sm' ? 'text-[10px] px-1.5 py-0' : 'text-[11px] px-2 py-0.5';

  return (
    <Badge
      variant="outline"
      className={`font-mono font-medium tracking-tight rounded-md border border-[#2e2e2e] bg-[#161616] text-[#ededed] hover:border-[#444444] transition-colors ${sizeClasses} ${className}`}
    >
      {name}
    </Badge>
  );
};
