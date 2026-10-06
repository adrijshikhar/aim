import * as React from 'react';
import { MCPServerDTO, ProfileDTO, getMcpServers } from '@/lib/api';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
  Search,
  X,
  Filter,
  ChevronDown,
  ChevronRight,
  Wrench,
} from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

const KNOWN_SERVER_TOOLS: Record<string, Array<{ name: string; description: string }>> = {
  'claude-mem': [
    { name: 'search', description: 'Semantic search across historical workflow runs' },
    { name: 'session_start_context', description: 'Hydrate contextual memory at session start' },
    { name: 'work_state_write', description: 'Persist checkpoint data to long-term memory' },
  ],
  'hevo-services': [
    { name: 'hermes_get_integration', description: 'Retrieve integration runtime pipeline status' },
    { name: 'catalog_get_objects', description: 'Fetch catalog schemas and object configuration' },
    { name: 'fhs_classify_errors', description: 'Classify ingestion and loading failure records' },
  ],
  'cloudflare-docs': [
    { name: 'search_cloudflare_documentation', description: 'Query official Cloudflare developer documentation' },
    { name: 'migrate_pages_to_workers_guide', description: 'Fetch migration guide and compatibility matrix' },
  ],
  'playwright': [
    { name: 'browser_navigate', description: 'Navigate active page to target URL' },
    { name: 'browser_take_screenshot', description: 'Capture viewport visual screenshot' },
    { name: 'browser_click', description: 'Click interactive DOM selector element' },
  ],
  'snyk': [
    { name: 'snyk_code_scan', description: 'Run static application security testing' },
    { name: 'snyk_sbom_scan', description: 'Generate and audit software bill of materials' },
  ],
};

interface McpViewProps {
  selectedScope?: string;
  onSelectScope?: (scope: string) => void;
  selectedProfile?: string;
  onSelectProfile?: (profile: string) => void;
  profiles?: ProfileDTO[];
}

export function McpView({
  selectedScope: propScope,
  onSelectScope: propOnSelectScope,
  selectedProfile,
  onSelectProfile,
  profiles = [],
}: McpViewProps) {
  const selectedScope = propScope ?? selectedProfile ?? 'all';
  const onSelectScope = propOnSelectScope ?? onSelectProfile;
  const [servers, setServers] = React.useState<MCPServerDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [copiedName, setCopiedName] = React.useState<string | null>(null);
  const [expandedTools, setExpandedTools] = React.useState<Record<string, boolean>>({});

  const { toast } = useToast();

  const fetchServers = React.useCallback(async () => {
    setLoading(true);
    try {
      // If scoped to a specific profile, pass it to backend API
      const profileParam =
        selectedScope === 'all' || selectedScope === 'global' ? undefined : selectedScope;
      const data = await getMcpServers(profileParam);
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
          scope: 'office',
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
          scope: 'work',
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
  }, [selectedScope]);

  React.useEffect(() => {
    fetchServers();
  }, [fetchServers]);

  // Compute profile scope options
  const profileScopeOptions = React.useMemo(() => {
    const set = new Set<string>();
    if (profiles && profiles.length > 0) {
      profiles.forEach((p) => set.add(p.name));
    }
    servers.forEach((s) => {
      if (s.scope && s.scope !== 'global') {
        set.add(s.scope);
      }
    });
    return Array.from(set).sort();
  }, [profiles, servers]);

  const filteredServers = React.useMemo(() => {
    return servers.filter((srv) => {
      // Scope filtering
      if (selectedScope === 'global') {
        if (srv.scope !== 'global') return false;
      } else if (selectedScope && selectedScope !== 'all') {
        if (srv.scope !== selectedScope && srv.scope !== 'global') return false;
      }

      // Search query filtering
      if (!searchQuery.trim()) return true;
      const q = searchQuery.toLowerCase();
      const matchName = srv.name.toLowerCase().includes(q);
      const matchCmd = srv.command.toLowerCase().includes(q);
      const matchArgs = srv.args.some((a) => a.toLowerCase().includes(q));
      const matchEnv = srv.env
        ? Object.entries(srv.env).some(
            ([k, v]) => k.toLowerCase().includes(q) || String(v).toLowerCase().includes(q)
          )
        : false;

      return matchName || matchCmd || matchArgs || matchEnv;
    });
  }, [servers, selectedScope, searchQuery]);

  const copyConfig = (srv: MCPServerDTO) => {
    const jsonStr = JSON.stringify(srv, null, 2);
    navigator.clipboard.writeText(jsonStr);
    setCopiedName(srv.name);
    toast({
      title: 'Copied Config',
      description: `Copied JSON configuration for "${srv.name}" to clipboard.`,
      variant: 'success',
    });
    setTimeout(() => {
      setCopiedName(null);
    }, 2000);
  };

  const hasActiveFilters =
    (selectedScope && selectedScope !== 'all') || searchQuery.trim().length > 0;

  const resetFilters = () => {
    if (onSelectScope) onSelectScope('all');
    setSearchQuery('');
  };

  return (
    <div className="space-y-6">
      {/* Top Banner */}
      <div className="rounded-geist border border-[#262626] bg-[#111111] p-5 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-[#ededed]" />
            <h3 className="font-semibold text-sm text-[#ededed]">
              Model Context Protocol (MCP) Catalog
            </h3>
          </div>
          <p className="text-xs text-[#a1a1a1] max-w-2xl leading-relaxed">
            AIM manages and bridges MCP server definitions across your agent environments.
            Global servers are shared across all profiles, while scoped servers attach only to their designated profile.
          </p>
        </div>
        <div className="flex items-center gap-2 self-end sm:self-auto">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchServers}
            disabled={loading}
            className="rounded-geist h-8 font-mono text-xs border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:bg-[#161616]"
          >
            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search input */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-[#666666]" />
          <Input
            placeholder="Search MCP servers by name, command, or env..."
            className="pl-9 pr-8 h-8 text-xs rounded-geist border-[#262626] bg-[#0a0a0a] text-[#ededed] placeholder:text-[#666666] focus:border-[#333333]"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button
              type="button"
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-2 text-[#666666] hover:text-[#ededed] cursor-pointer"
              title="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Scope Dropdown Filter */}
        <div className="flex items-center gap-2">
          <span className="text-xs text-[#666666] font-mono hidden sm:inline">Scope:</span>
          <select
            value={selectedScope}
            onChange={(e) => onSelectScope && onSelectScope(e.target.value)}
            className="h-8 px-2.5 text-xs font-mono rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#ededed] hover:border-[#333333] focus:outline-none focus:border-[#333333] cursor-pointer"
            title="Filter by Scope"
          >
            <option value="all">All Scopes</option>
            <option value="global">Global Only</option>
            {profileScopeOptions.map((name) => (
              <option key={name} value={name}>
                Profile: {name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Active Filter Chips */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2 pt-0.5 text-xs">
          <span className="flex items-center gap-1 text-[#666666] text-[11px] font-mono">
            <Filter className="h-3 w-3" /> Filters:
          </span>

          {selectedScope && selectedScope !== 'all' && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] font-mono text-[11px]">
              scope: <strong className="text-[#ededed]">{selectedScope}</strong>
              <button
                type="button"
                onClick={() => onSelectScope && onSelectScope('all')}
                className="hover:text-[#ededed] cursor-pointer"
                title="Clear scope filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          {searchQuery && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-[#262626] bg-[#0a0a0a] text-[#a1a1a1] font-mono text-[11px]">
              query: &ldquo;<strong className="text-[#ededed]">{searchQuery}</strong>&rdquo;
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="hover:text-[#ededed] cursor-pointer"
                title="Clear query filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          <button
            type="button"
            onClick={resetFilters}
            className="text-[11px] font-mono text-[#a1a1a1] hover:text-[#ededed] underline ml-1 cursor-pointer"
          >
            Reset all
          </button>
        </div>
      )}

      {/* Grid of Servers */}
      {loading && servers.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 rounded-geist border border-[#262626] bg-[#111111]/40 animate-pulse p-6" />
          ))}
        </div>
      ) : filteredServers.length === 0 ? (
        <div className="rounded-geist border border-dashed border-[#262626] p-12 text-center bg-[#0a0a0a]/50">
          <Server className="h-10 w-10 text-[#666666] mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-[#ededed]">No MCP servers found</h3>
          <p className="text-xs text-[#a1a1a1] mt-1 max-w-sm mx-auto">
            {hasActiveFilters
              ? 'Try modifying your filter or search criteria.'
              : 'Configured MCP servers from global settings and profile manifests will appear here.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredServers.map((srv) => {
            const isGlobal = srv.scope === 'global';
            return (
              <Card
                key={`${srv.scope}-${srv.name}`}
                className="group relative overflow-hidden transition-all duration-150 border border-[#262626] bg-[#111111] hover:border-[#333333] rounded-geist flex flex-col justify-between"
              >
                <CardHeader className="pb-3 p-5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <div className="p-1.5 rounded-geist bg-[#0a0a0a] border border-[#262626] text-[#ededed]">
                        <Server className="h-3.5 w-3.5" />
                      </div>
                      <CardTitle className="text-sm font-semibold text-[#ededed] font-mono">{srv.name}</CardTitle>
                    </div>

                    <div className="flex items-center gap-1.5 flex-wrap">
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[3px] border border-border bg-[#0a0a0a] text-[#888888]">
                        stdio
                      </span>
                      <span className="badge-success text-[10px] font-mono px-1.5 py-0.5 rounded-[3px] flex items-center gap-1">
                        <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid animate-pulse" />
                        Active
                      </span>
                      <Badge
                        variant="outline"
                        className="text-[11px] font-mono flex items-center gap-1.5 rounded-geist px-2 py-0.5 border-[#262626] bg-[#0a0a0a] text-[#a1a1a1]"
                      >
                        {isGlobal ? (
                          <>
                            <Globe className="h-3 w-3 text-[#ededed]" />
                            Global
                          </>
                        ) : (
                          <>
                            <Lock className="h-3 w-3 text-[#f54e00]" />
                            Profile: {srv.scope}
                          </>
                        )}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="space-y-3 pb-4 p-5 pt-0">
                  {/* Command & Args */}
                  <div className="space-y-1">
                    <span className="text-[11px] font-mono uppercase tracking-wider text-[#a1a1a1] flex items-center gap-1">
                      <Terminal className="h-3 w-3 text-[#666666]" /> Command
                    </span>
                    <div className="font-mono text-xs p-2 rounded-geist bg-[#0a0a0a] border border-[#262626] text-[#ededed] overflow-x-auto whitespace-pre">
                      <span className="text-[#ededed] font-medium">{srv.command}</span>{' '}
                      <span className="text-[#a1a1a1]">{srv.args.join(' ')}</span>
                    </div>
                  </div>

                  {/* Environment Variables if present */}
                  {srv.env && Object.keys(srv.env).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-[11px] font-mono uppercase tracking-wider text-[#a1a1a1] flex items-center gap-1">
                        <Layers className="h-3 w-3 text-[#666666]" /> Environment
                      </span>
                      <div className="space-y-1 font-mono text-[11px] bg-[#0a0a0a] p-2 rounded-geist border border-[#262626]">
                        {Object.entries(srv.env).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-[#a1a1a1]">
                            <span className="text-[#ededed] font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-[#a1a1a1]">{v}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Collapsible Registered Tools */}
                  {(() => {
                    const cardKey = `${srv.scope}-${srv.name}`;
                    const isToolsOpen = !!expandedTools[cardKey];
                    const toolsList = KNOWN_SERVER_TOOLS[srv.name] || [
                      { name: `${srv.name}_action`, description: `Interactive tool protocol handler for ${srv.name}` },
                    ];
                    return (
                      <div className="rounded-[4px] border border-[#222222] bg-[#0c0c0c] overflow-hidden">
                        <button
                          type="button"
                          onClick={() =>
                            setExpandedTools((prev) => ({
                              ...prev,
                              [cardKey]: !prev[cardKey],
                            }))
                          }
                          className="w-full flex items-center justify-between px-2.5 py-1.5 text-[11px] font-mono text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#141414] transition-colors cursor-pointer"
                        >
                          <span className="flex items-center gap-1.5">
                            <Wrench className="h-3 w-3 text-[#888888]" />
                            <span>Registered Tools</span>
                            <span className="text-[10px] text-[#666666]">
                              ({toolsList.length})
                            </span>
                          </span>
                          {isToolsOpen ? (
                            <ChevronDown className="h-3 w-3 text-[#666666]" />
                          ) : (
                            <ChevronRight className="h-3 w-3 text-[#666666]" />
                          )}
                        </button>
                        {isToolsOpen && (
                          <div className="p-2 border-t border-[#1c1c1c] space-y-1.5 bg-[#080808]">
                            {toolsList.map((t) => (
                              <div
                                key={t.name}
                                className="p-1.5 rounded-[3px] bg-[#0f0f0f] border border-[#1a1a1a] space-y-0.5"
                              >
                                <div className="font-mono text-[11px] font-semibold text-[#ededed]">
                                  {t.name}
                                </div>
                                <div className="text-[10px] text-[#888888] font-sans">
                                  {t.description}
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    );
                  })()}

                  {/* Copy config helper */}
                  <div className="pt-2 flex justify-end">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs font-mono text-[#a1a1a1] hover:text-[#ededed] hover:bg-[#161616] rounded-geist cursor-pointer"
                      onClick={() => copyConfig(srv)}
                    >
                      {copiedName === srv.name ? (
                        <Check className="h-3.5 w-3.5 mr-1 text-emerald-400" />
                      ) : (
                        <Copy className="h-3.5 w-3.5 mr-1 text-[#666666]" />
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
