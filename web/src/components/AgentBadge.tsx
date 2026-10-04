import React from 'react';
import { Badge } from './ui/badge';

interface AgentBadgeProps {
  agent: string;
  className?: string;
  size?: 'sm' | 'default';
}

export const getAgentBadgeClasses = (agent: string) => {
  const a = agent?.toLowerCase() || '';
  if (a.includes('claude')) {
    return 'bg-[#ea580c]/12 text-[#fb923c] border-[#ea580c]/30 hover:border-[#ea580c]/50';
  }
  if (a.includes('codex') || a.includes('openai')) {
    return 'bg-[#10a37f]/12 text-[#34d399] border-[#10a37f]/30 hover:border-[#10a37f]/50';
  }
  if (a.includes('agy') || a.includes('antigravity')) {
    return 'bg-[#6366f1]/12 text-[#a5b4fc] border-[#6366f1]/30 hover:border-[#6366f1]/50';
  }
  if (a.includes('gemini')) {
    return 'bg-[#0284c7]/12 text-[#38bdf8] border-[#0284c7]/30 hover:border-[#0284c7]/50';
  }
  return 'bg-[#222222] text-[#ededed] border-[#333333] hover:border-[#444444]';
};

export const AgentBadge: React.FC<AgentBadgeProps> = ({ agent, className = '', size = 'default' }) => {
  const name = agent?.toLowerCase() || 'agent';
  const colorClasses = getAgentBadgeClasses(name);
  const sizeClasses = size === 'sm' ? 'text-[10px] px-1.5 py-0' : 'text-[11px] px-2 py-0.5';

  return (
    <Badge
      variant="outline"
      className={`font-mono font-medium tracking-tight rounded-md transition-colors ${colorClasses} ${sizeClasses} ${className}`}
    >
      {name}
    </Badge>
  );
};
