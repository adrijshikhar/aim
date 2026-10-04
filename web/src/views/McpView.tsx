import * as React from 'react';
import { MCPServerDTO, getMcpServers } from '@/lib/api';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Server,
  Terminal,
  Layers,
  RefreshCw,
  Copy,
  Check,
  Cpu,
  Globe,
  Lock,
} from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

export function McpView() {
  const [servers, setServers] = React.useState<MCPServerDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [copiedName, setCopiedName] = React.useState<string | null>(null);

  const { toast } = useToast();

  const fetchServers = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getMcpServers();
      setServers(data);
    } catch {
      // Dev fallback data
      setServers([
        {
          name: 'claude-mem',
          command: 'npx',
          args: ['-y', '@claude-mem/server'],
          env: { MEMORY_DIR: '~/.claude-mem' },
          scope: 'global',
        },
        {
          name: 'hevo-services',
          command: 'python3',
          args: ['-m', 'hevo_services.mcp'],
          env: { ENVIRONMENT: 'nonprod' },
          scope: 'work',
        },
        {
          name: 'cloudflare-docs',
          command: 'npx',
          args: ['-y', '@cloudflare/mcp-server-cloudflare'],
          scope: 'global',
        },
        {
          name: 'playwright',
          command: 'npx',
          args: ['-y', '@playwright/mcp@latest'],
          scope: 'personal',
        },
        {
          name: 'snyk',
          command: 'snyk',
          args: ['mcp'],
          scope: 'global',
        },
      ]);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    fetchServers();
  }, [fetchServers]);

  const copyConfig = (srv: MCPServerDTO) => {
    const jsonStr = JSON.stringify(srv, null, 2);
    navigator.clipboard.writeText(jsonStr);
    setCopiedName(srv.name);
    toast({
      title: 'Copied Config',
      description: `Copied JSON configuration for "${srv.name}" to clipboard.`,
    });
    setTimeout(() => {
      setCopiedName(null);
    }, 2000);
  };

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="rounded-xl border border-primary/20 bg-primary/5 p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Cpu className="h-5 w-5 text-primary" />
            <h3 className="font-semibold text-sm text-foreground">
              Model Context Protocol (MCP) Infrastructure
            </h3>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl">
            AIM isolates and bridges MCP server definitions across your agent environments.
            Global servers are shared across all profiles, while scoped servers attach only to their designated profile.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchServers} disabled={loading}>
          <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Grid of Servers */}
      {loading && servers.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 rounded-xl border border-border/50 bg-card/40 animate-pulse p-6" />
          ))}
        </div>
      ) : servers.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border/80 p-12 text-center">
          <Server className="h-10 w-10 text-muted-foreground mx-auto mb-3 opacity-60" />
          <h3 className="text-base font-semibold">No MCP servers detected</h3>
          <p className="text-sm text-muted-foreground mt-1 max-w-sm mx-auto">
            Configured MCP servers from global settings and profile manifests will appear here.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {servers.map((srv) => {
            const isGlobal = srv.scope === 'global';
            return (
              <Card
                key={srv.name}
                className="group relative overflow-hidden transition-all duration-200 hover:border-primary/50 hover:shadow-md"
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="p-1.5 rounded-lg bg-primary/10 border border-primary/20 text-primary">
                        <Server className="h-4 w-4" />
                      </div>
                      <CardTitle className="text-base font-semibold">{srv.name}</CardTitle>
                    </div>

                    <Badge
                      variant={isGlobal ? 'secondary' : 'outline'}
                      className={`text-[11px] font-medium flex items-center gap-1 ${
                        isGlobal
                          ? 'bg-blue-500/10 text-blue-400 border-blue-500/20'
                          : 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                      }`}
                    >
                      {isGlobal ? (
                        <>
                          <Globe className="h-3 w-3" />
                          Global
                        </>
                      ) : (
                        <>
                          <Lock className="h-3 w-3" />
                          Profile: {srv.scope}
                        </>
                      )}
                    </Badge>
                  </div>
                </CardHeader>

                <CardContent className="space-y-3 pb-4">
                  {/* Command & Args */}
                  <div className="space-y-1">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-1">
                      <Terminal className="h-3 w-3" /> Command
                    </span>
                    <div className="font-mono text-xs p-2 rounded bg-muted/60 border border-border/50 text-foreground overflow-x-auto whitespace-pre">
                      <span className="text-emerald-400 font-bold">{srv.command}</span>{' '}
                      <span className="text-muted-foreground">{srv.args.join(' ')}</span>
                    </div>
                  </div>

                  {/* Environment Variables if present */}
                  {srv.env && Object.keys(srv.env).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-1">
                        <Layers className="h-3 w-3" /> Environment
                      </span>
                      <div className="space-y-1 font-mono text-[11px] bg-muted/30 p-2 rounded border border-border/40">
                        {Object.entries(srv.env).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-muted-foreground">
                            <span className="text-primary font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-foreground">{v}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Copy config helper */}
                  <div className="pt-2 flex justify-end">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs text-muted-foreground hover:text-foreground"
                      onClick={() => copyConfig(srv)}
                    >
                      {copiedName === srv.name ? (
                        <Check className="h-3.5 w-3.5 mr-1 text-emerald-400" />
                      ) : (
                        <Copy className="h-3.5 w-3.5 mr-1" />
                      )}
                      {copiedName === srv.name ? 'Copied' : 'Copy JSON'}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
