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

interface McpViewProps {
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
}

export function McpView({ selectedProfile = 'all', onSelectProfile }: McpViewProps) {
  const [servers, setServers] = React.useState<MCPServerDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [copiedName, setCopiedName] = React.useState<string | null>(null);

  const { toast } = useToast();

  const fetchServers = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await getMcpServers(selectedProfile === 'all' ? undefined : selectedProfile);
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
      <div className="rounded-geist border border-[#262626] bg-[#111111] p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-[#ededed]" />
            <h3 className="font-semibold text-sm text-[#ededed]">
              Model Context Protocol (MCP) Infrastructure
            </h3>
          </div>
          <p className="text-xs text-[#888888] max-w-2xl">
            AIM isolates and bridges MCP server definitions across your agent environments.
            Global servers are shared across all profiles, while scoped servers attach only to their designated profile.
          </p>
        </div>
        <div className="flex items-center gap-2 self-end sm:self-auto">
          {selectedProfile && selectedProfile !== 'all' && (
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded-geist border border-[#2a2a2a] bg-[#141414] text-xs font-mono text-[#888888]">
              <span>Scoped to: <strong className="text-[#ededed]">{selectedProfile}</strong></span>
              {onSelectProfile && (
                <button
                  type="button"
                  onClick={() => onSelectProfile('all')}
                  className="hover:text-[#ededed] ml-1 text-[10px] text-[#666666] cursor-pointer"
                  title="Show all MCP servers"
                >
                  ✕
                </button>
              )}
            </div>
          )}
          <Button variant="outline" size="sm" onClick={fetchServers} disabled={loading} className="rounded-geist h-8 font-mono text-xs">
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Grid of Servers */}
      {loading && servers.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 rounded-geist border border-[#262626] bg-[#111111]/40 animate-pulse p-6" />
          ))}
        </div>
      ) : servers.length === 0 ? (
        <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0e0e0e]/50">
          <Server className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-[#ededed]">No MCP servers detected</h3>
          <p className="text-xs text-[#888888] mt-1 max-w-sm mx-auto">
            Configured MCP servers from global settings and profile manifests will appear here.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {servers.map((srv) => {
            const isGlobal = srv.scope === 'global';
            return (
              <Card
                key={srv.name}
                className="group relative overflow-hidden transition-all duration-150 border border-[#262626] bg-[#111111] hover:border-[#383838] rounded-geist"
              >
                <CardHeader className="pb-3 p-5">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="p-1.5 rounded-geist bg-[#171717] border border-[#2a2a2a] text-[#ededed]">
                        <Server className="h-3.5 w-3.5" />
                      </div>
                      <CardTitle className="text-sm font-semibold text-[#ededed]">{srv.name}</CardTitle>
                    </div>

                    <Badge
                      variant="outline"
                      className={`text-[11px] font-mono flex items-center gap-1 rounded-full px-2 py-0.5 ${
                        isGlobal
                          ? 'bg-[#0d1f36] text-[#0070f3] border-[#15345a]'
                          : 'bg-[#1c122c] text-[#b388ff] border-[#331c52]'
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

                <CardContent className="space-y-3 pb-4 p-5 pt-0">
                  {/* Command & Args */}
                  <div className="space-y-1">
                    <span className="text-[11px] font-mono uppercase tracking-wider text-[#888888] flex items-center gap-1">
                      <Terminal className="h-3 w-3" /> Command
                    </span>
                    <div className="font-mono text-xs p-2 rounded-geist bg-[#0e0e0e] border border-[#262626] text-[#ededed] overflow-x-auto whitespace-pre">
                      <span className="text-[#50e3c2] font-semibold">{srv.command}</span>{' '}
                      <span className="text-[#888888]">{srv.args.join(' ')}</span>
                    </div>
                  </div>

                  {/* Environment Variables if present */}
                  {srv.env && Object.keys(srv.env).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-[11px] font-mono uppercase tracking-wider text-[#888888] flex items-center gap-1">
                        <Layers className="h-3 w-3" /> Environment
                      </span>
                      <div className="space-y-1 font-mono text-[11px] bg-[#161616]/70 p-2 rounded-geist border border-[#222222]">
                        {Object.entries(srv.env).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-[#888888]">
                            <span className="text-[#ededed] font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-[#888888]">{v}</span>
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
                      className="h-7 text-xs font-mono text-[#888888] hover:text-[#ededed] hover:bg-[#1a1a1a] rounded-geist"
                      onClick={() => copyConfig(srv)}
                    >
                      {copiedName === srv.name ? (
                        <Check className="h-3.5 w-3.5 mr-1 text-[#50e3c2]" />
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
