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
import { SelectDropdown } from '@/components/ui/select-dropdown';

const KNOWN_SERVER_TOOLS: Record<string, Array<{ name: string; description: string; params?: string[] }>> = {
  'claude-mem': [
    { name: 'search', description: 'Semantic search across historical workflow runs', params: ['query', 'limit'] },
    { name: 'session_start_context', description: 'Hydrate contextual memory at session start', params: ['session_id'] },
    { name: 'work_state_write', description: 'Persist checkpoint data to long-term memory', params: ['key', 'state'] },
  ],
  'hevo-services': [
    { name: 'hermes_get_integration', description: 'Retrieve integration runtime pipeline status', params: ['integration_id'] },
    { name: 'catalog_get_objects', description: 'Fetch catalog schemas and object configuration', params: ['source_id'] },
    { name: 'fhs_classify_errors', description: 'Classify ingestion and loading failure records', params: ['error_message'] },
  ],
  'cloudflare-docs': [
    { name: 'search_cloudflare_documentation', description: 'Query official Cloudflare developer documentation', params: ['query', 'product'] },
    { name: 'migrate_pages_to_workers_guide', description: 'Fetch migration guide and compatibility matrix', params: ['framework'] },
  ],
  'playwright': [
    { name: 'browser_navigate', description: 'Navigate active page to target URL', params: ['url'] },
    { name: 'browser_take_screenshot', description: 'Capture viewport visual screenshot', params: ['full_page'] },
    { name: 'browser_click', description: 'Click interactive DOM selector element', params: ['selector'] },
  ],
  'snyk': [
    { name: 'snyk_code_scan', description: 'Run static application security testing', params: ['path', 'severity'] },
    { name: 'snyk_sbom_scan', description: 'Generate and audit software bill of materials', params: ['target_file'] },
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
      {/* Header bar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 pb-1">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <Cpu className="h-4 w-4 text-foreground" />
            <h3 className="font-semibold text-sm text-foreground">
              Model Context Protocol (MCP) Catalog
            </h3>
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded border border-border bg-card text-muted-foreground">
              {filteredServers.length} {filteredServers.length === 1 ? 'server' : 'servers'}
            </span>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl leading-relaxed">
            AIM manages MCP server definitions across agent environments. Global servers are shared, while scoped servers attach only to designated profiles.
          </p>
        </div>
        <div className="flex items-center gap-2 self-end sm:self-auto shrink-0">
          <Button
            variant="outline"
            size="sm"
            onClick={fetchServers}
            disabled={loading}
            className="rounded-geist h-8 font-mono text-xs border-border bg-card text-foreground hover:bg-secondary cursor-pointer"
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
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-muted-foreground/60" />
          <Input
            placeholder="Search MCP servers by name, command, or env..."
            className="pl-9 pr-8 h-8 text-xs font-sans rounded-geist border-border bg-card text-foreground placeholder:text-muted-foreground/50 focus:border-border"
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
        <div className="flex items-center gap-2 min-w-[160px]">
          <span className="text-xs text-[#666666] font-mono hidden sm:inline">Scope:</span>
          <SelectDropdown
            value={selectedScope}
            onChange={(val) => onSelectScope && onSelectScope(String(val))}
            size="sm"
            options={[
              { value: 'all', label: 'All Scopes' },
              { value: 'global', label: 'Global Only' },
              ...profileScopeOptions.map((name) => ({
                value: name,
                label: `Profile: ${name}`,
              })),
            ]}
          />
        </div>
      </div>

      {/* Active Filter Chips */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2 pt-0.5 text-xs">
          <span className="flex items-center gap-1 text-muted-foreground text-[11px] font-mono">
            <Filter className="h-3 w-3" /> Filters:
          </span>

          {selectedScope && selectedScope !== 'all' && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-border bg-secondary/50 text-muted-foreground font-mono text-[11px]">
              scope: <strong className="text-foreground">{selectedScope}</strong>
              <button
                type="button"
                onClick={() => onSelectScope && onSelectScope('all')}
                className="hover:text-foreground cursor-pointer"
                title="Clear scope filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          {searchQuery && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-border bg-secondary/50 text-muted-foreground font-mono text-[11px]">
              query: &ldquo;<strong className="text-foreground">{searchQuery}</strong>&rdquo;
              <button
                type="button"
                onClick={() => setSearchQuery('')}
                className="hover:text-foreground cursor-pointer"
                title="Clear query filter"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          )}

          <button
            type="button"
            onClick={resetFilters}
            className="text-[11px] font-mono text-muted-foreground hover:text-foreground underline ml-1 cursor-pointer"
          >
            Reset all
          </button>
        </div>
      )}

      {/* Grid of Servers */}
      {loading && servers.length === 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 rounded-geist border border-border bg-card/40 animate-pulse p-6" />
          ))}
        </div>
      ) : filteredServers.length === 0 ? (
        <div className="rounded-geist border border-dashed border-border p-12 text-center bg-card/50">
          <Server className="h-10 w-10 text-muted-foreground mx-auto mb-3 opacity-60" />
          <h3 className="text-sm font-semibold text-foreground">No MCP servers found</h3>
          <p className="text-xs text-muted-foreground mt-1 max-w-sm mx-auto">
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
                className="group relative overflow-hidden transition-all duration-150 border border-border bg-card hover:border-border/80 rounded-geist flex flex-col justify-between"
              >
                <CardHeader className="pb-3 p-5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <div className="p-1.5 rounded-geist bg-secondary/50 border border-border text-foreground">
                        <Server className="h-3.5 w-3.5" />
                      </div>
                      <CardTitle className="text-sm font-semibold text-foreground font-mono">{srv.name}</CardTitle>
                    </div>

                    <div className="flex items-center gap-1.5 flex-wrap">
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded-[3px] border border-border bg-secondary/40 text-muted-foreground">
                        stdio
                      </span>
                      <span className="badge-success text-[10px] font-mono px-1.5 py-0.5 rounded-[3px] flex items-center gap-1">
                        <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                        Active
                      </span>
                      <Badge
                        variant="outline"
                        className="text-[11px] font-mono flex items-center gap-1.5 rounded-geist px-2 py-0.5 border-border bg-secondary/40 text-muted-foreground"
                      >
                        {isGlobal ? (
                          <>
                            <Globe className="h-3 w-3 text-foreground" />
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
                    <span className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                      <Terminal className="h-3 w-3 text-muted-foreground/70" /> Command
                    </span>
                    <div className="font-mono text-xs p-2 rounded-geist bg-background border border-border text-foreground overflow-x-auto whitespace-pre">
                      <span className="text-foreground font-medium">{srv.command}</span>{' '}
                      <span className="text-muted-foreground">{srv.args.join(' ')}</span>
                    </div>
                  </div>

                  {/* Environment Variables if present */}
                  {srv.env && Object.keys(srv.env).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Layers className="h-3 w-3 text-muted-foreground/70" /> Environment
                      </span>
                      <div className="space-y-1 font-mono text-[11px] bg-background p-2 rounded-geist border border-border">
                        {Object.entries(srv.env).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-muted-foreground">
                            <span className="text-foreground font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-muted-foreground">{v}</span>
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
                      <div className="rounded-[4px] border border-border bg-card overflow-hidden">
                        <button
                          type="button"
                          onClick={() =>
                            setExpandedTools((prev) => ({
                              ...prev,
                              [cardKey]: !prev[cardKey],
                            }))
                          }
                          className="w-full flex items-center justify-between px-2.5 py-1.5 text-[11px] font-mono text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors cursor-pointer"
                        >
                          <span className="flex items-center gap-1.5">
                            <Wrench className="h-3 w-3 text-muted-foreground" />
                            <span>Registered Tools</span>
                            <span className="text-[10px] text-muted-foreground/70">
                              ({toolsList.length})
                            </span>
                          </span>
                          {isToolsOpen ? (
                            <ChevronDown className="h-3 w-3 text-muted-foreground/70" />
                          ) : (
                            <ChevronRight className="h-3 w-3 text-muted-foreground/70" />
                          )}
                        </button>
                        {isToolsOpen && (
                          <div className="p-2 border-t border-border space-y-1.5 bg-background">
                            {toolsList.map((t) => (
                              <div
                                key={t.name}
                                className="p-2 rounded-[3px] bg-secondary/30 border border-border/60 space-y-1"
                              >
                                <div className="flex items-center justify-between gap-2 flex-wrap">
                                  <div className="font-mono text-[11px] font-semibold text-foreground">
                                    {t.name}
                                  </div>
                                  {t.params && t.params.length > 0 && (
                                    <div className="flex items-center gap-1 flex-wrap">
                                      {t.params.map((p) => (
                                        <span
                                          key={p}
                                          className="text-[9px] font-mono px-1 py-0 rounded bg-secondary text-muted-foreground border border-border"
                                        >
                                          {p}
                                        </span>
                                      ))}
                                    </div>
                                  )}
                                </div>
                                <div className="text-[10px] text-muted-foreground font-sans leading-relaxed">
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
                      className="h-7 text-xs font-mono text-muted-foreground hover:text-foreground hover:bg-secondary rounded-geist cursor-pointer"
                      onClick={() => copyConfig(srv)}
                    >
                      {copiedName === srv.name ? (
                        <Check className="h-3.5 w-3.5 mr-1 text-status-success" />
                      ) : (
                        <Copy className="h-3.5 w-3.5 mr-1 text-muted-foreground" />
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
