import * as React from 'react';
import { DesktopStatusDTO, launchDesktopApp, quitDesktopApp, focusDesktopApp } from '@/lib/api';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { AppWindow, Code2, Play, Square, ExternalLink, Loader2, RefreshCw } from 'lucide-react';
import { AgentBadge } from './AgentBadge';
import { cn } from '@/lib/utils';
import { useToast } from '@/components/ui/use-toast';

interface DesktopAppCardProps {
  status: DesktopStatusDTO;
  profiles: string[];
  activeProfile?: string;
  onRefresh: () => void;
  onRequestSwitch: (agent: string, targetProfile: string, currentRunningProfile: string) => void;
}

export const DesktopAppCard: React.FC<DesktopAppCardProps> = ({
  status,
  profiles,
  activeProfile,
  onRefresh,
  onRequestSwitch,
}) => {
  const { toast } = useToast();
  const [selectedProfile, setSelectedProfile] = React.useState(activeProfile || profiles[0] || 'default');
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (activeProfile && profiles.includes(activeProfile)) {
      setSelectedProfile(activeProfile);
    }
  }, [activeProfile, profiles]);

  const isRunning = status.running && status.pid > 0;
  const isRunningThisProfile = isRunning && status.profile === selectedProfile;

  const handleLaunchOrFocus = async () => {
    if (isRunningThisProfile) {
      // Focus existing window
      try {
        setLoading(true);
        await focusDesktopApp(status.agent);
        toast({
          title: `${status.display_name} Focused`,
          description: `Brought window to front (PID ${status.pid}).`,
        });
      } catch (err: any) {
        toast({
          title: 'Focus Failed',
          description: err.message || 'Could not bring window to front.',
          variant: 'destructive',
        });
      } finally {
        setLoading(false);
      }
      return;
    }

    if (isRunning && status.profile !== selectedProfile) {
      // Instance collision -> trigger switch modal
      onRequestSwitch(status.agent, selectedProfile, status.profile);
      return;
    }

    // Launch application
    try {
      setLoading(true);
      const res = await launchDesktopApp({
        agent: status.agent,
        profile: selectedProfile,
      });

      if (res.conflict) {
        onRequestSwitch(status.agent, selectedProfile, res.current_profile || status.profile);
        return;
      }

      toast({
        title: `${status.display_name} Launched`,
        description: `Running isolated with profile "${selectedProfile}".`,
      });
      onRefresh();
    } catch (err: any) {
      toast({
        title: 'Launch Failed',
        description: err.message || 'Failed to launch desktop application.',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleQuit = async () => {
    try {
      setLoading(true);
      await quitDesktopApp(status.agent);
      toast({
        title: `${status.display_name} Stopped`,
        description: `Gracefully terminated process (PID ${status.pid}).`,
      });
      onRefresh();
    } catch (err: any) {
      toast({
        title: 'Quit Failed',
        description: err.message || 'Failed to quit application.',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const renderIcon = () => {
    if (status.agent === 'cursor') {
      return <Code2 className="h-5 w-5 text-[#61afef]" />;
    }
    return <AppWindow className="h-5 w-5 text-accent-cyan" />;
  };

  return (
    <Card className="bg-card/70 border-border/70 hover:border-border transition-colors duration-150">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-9 w-9 rounded-md bg-secondary/60 border border-border flex items-center justify-center shrink-0">
              {renderIcon()}
            </div>
            <div>
              <CardTitle className="text-sm font-semibold text-foreground flex items-center gap-2">
                <span>{status.display_name}</span>
                <AgentBadge agent={status.agent} size="sm" />
              </CardTitle>
              <CardDescription className="text-xs text-muted-foreground mt-0.5">
                Sandboxed Desktop Application
              </CardDescription>
            </div>
          </div>

          <div className="flex items-center gap-1.5">
            {isRunning ? (
              <Badge
                variant="outline"
                className="bg-[#183925] text-status-success border-[#235836] font-mono text-[10px] flex items-center gap-1.5"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-status-success animate-pulse" />
                <span>RUNNING (PID {status.pid})</span>
              </Badge>
            ) : (
              <Badge
                variant="outline"
                className="bg-secondary/40 text-muted-foreground border-border font-mono text-[10px] flex items-center gap-1.5"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
                <span>STOPPED</span>
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-3 pt-0">
        <div className="rounded border border-border/60 bg-secondary/30 p-2.5 text-xs font-mono space-y-1.5">
          <div className="flex justify-between items-center text-muted-foreground">
            <span>Isolated Profile:</span>
            {isRunning ? (
              <span className="text-foreground font-medium bg-secondary px-1.5 py-0.5 rounded border border-border/80">
                {status.profile}
              </span>
            ) : (
              <span className="text-muted-foreground italic">None (Standby)</span>
            )}
          </div>

          <div className="flex justify-between items-center text-muted-foreground">
            <span>Target Launch Profile:</span>
            <select
              value={selectedProfile}
              onChange={(e) => setSelectedProfile(e.target.value)}
              disabled={loading}
              className="bg-card text-foreground border border-border/80 rounded px-2 py-0.5 text-xs font-mono focus:outline-none focus:border-[#61afef] transition-colors"
            >
              {profiles.map((p) => (
                <option key={p} value={p}>
                  {p}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex items-center justify-between gap-2 pt-1">
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant={isRunningThisProfile ? 'secondary' : 'default'}
              disabled={loading}
              onClick={handleLaunchOrFocus}
              className={cn(
                'text-xs h-8 px-3 transition-colors duration-150',
                !isRunningThisProfile && 'bg-[#61afef] text-[#1e1e1e] hover:bg-[#4d9bdf]'
              )}
            >
              {loading ? (
                <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
              ) : isRunningThisProfile ? (
                <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
              ) : isRunning ? (
                <RefreshCw className="h-3.5 w-3.5 mr-1.5" />
              ) : (
                <Play className="h-3.5 w-3.5 mr-1.5" />
              )}
              {isRunningThisProfile
                ? 'Bring to Front'
                : isRunning
                ? 'Switch to Profile'
                : 'Launch App'}
            </Button>

            {isRunning && (
              <Button
                size="sm"
                variant="outline"
                disabled={loading}
                onClick={handleQuit}
                className="text-xs h-8 px-2.5 text-status-danger border-border hover:bg-destructive/10 transition-colors duration-150"
              >
                <Square className="h-3 w-3 mr-1" />
                Quit
              </Button>
            )}
          </div>

          <span className="text-[11px] text-muted-foreground font-mono">
            {isRunning ? `Attached to ${status.profile}` : 'Sandboxed GUI'}
          </span>
        </div>
      </CardContent>
    </Card>
  );
};
