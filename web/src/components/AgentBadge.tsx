import React from 'react';
import { cn } from '@/lib/utils';
import { Terminal, AppWindow, Code2 } from 'lucide-react';

export interface AgentBadgeProps {
  agent: string;
  className?: string;
  size?: 'sm' | 'default';
  variant?: 'badge' | 'dot';
  showIcon?: boolean;
}

export const getAgentDotColor = (agent: string) => {
  const a = agent?.toLowerCase() || '';
  switch (a) {
    case 'claude':
    case 'claude-desktop':
      return 'bg-[#be9137]'; // Atom One Dark warm bronze
    case 'codex':
    case 'codex-desktop':
      return 'bg-[#47ae8f]'; // Slate-sage / pine
    case 'cursor':
      return 'bg-[#61afef]'; // Atom One Dark accent blue
    case 'gemini':
      return 'bg-[#6b94d1]'; // Slate-blue
    case 'agy':
    case 'antigravity':
      return 'bg-[#8e7dbf]'; // Slate-iris
    default:
      return 'bg-muted-foreground';
  }
};

export const getAgentIcon = (agent: string, className = 'h-3 w-3 shrink-0') => {
  const a = agent?.toLowerCase() || '';
  if (a === 'cursor') {
    return <Code2 className={className} />;
  }
  if (a === 'claude-desktop' || a === 'codex-desktop') {
    return <AppWindow className={className} />;
  }
  return <Terminal className={className} />;
};

export const AgentBadge: React.FC<AgentBadgeProps> = ({
  agent,
  className = '',
  size = 'default',
  variant = 'badge',
  showIcon = false,
}) => {
  const name = agent?.toLowerCase() || 'agent';
  const dotColor = getAgentDotColor(name);

  if (variant === 'dot') {
    return (
      <span className={cn('inline-flex items-center gap-1.5 font-mono text-xs text-foreground/80 select-none', className)}>
        {showIcon ? getAgentIcon(name) : <span className={cn('h-1.5 w-1.5 rounded-full shrink-0', dotColor)} />}
        <span>{name}</span>
      </span>
    );
  }

  const sizeClasses = size === 'sm' ? 'text-[11px] h-5 px-1.5' : 'text-xs h-5.5 px-2';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 font-mono font-medium rounded-[4px] border border-border/70 bg-secondary/50 text-foreground/90 select-none transition-colors hover:border-border',
        sizeClasses,
        className
      )}
    >
      {showIcon ? getAgentIcon(name) : <span className={cn('h-1.5 w-1.5 rounded-full shrink-0', dotColor)} />}
      <span>{name}</span>
    </span>
  );
};
