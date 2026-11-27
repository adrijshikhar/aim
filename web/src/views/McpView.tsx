import * as React from 'react';
import {
  MCPServerDTO,
  ProfileDTO,
  getMcpServers,
  addMcpServer,
  setMcpServerState,
  deleteMcpServer,
} from '@/lib/api';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { SegmentedControl } from '@/components/ui/segmented-control';
import {
  Server,
  Terminal,
  Layers,
  RefreshCw,
  Copy,
  Check,
  Globe,
  Lock,
  Search,
  X,
  Filter,
  ChevronDown,
  ChevronRight,
  Wrench,
  Plus,
  Trash2,
  Key,
  Power,
  ShieldCheck,
  AlertCircle,
} from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';
import { SelectDropdown } from '@/components/ui/select-dropdown';
import { McpCatalogSkeleton } from '@/components/Skeletons';

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
  const [error, setError] = React.useState<string | null>(null);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [copiedName, setCopiedName] = React.useState<string | null>(null);
  const [expandedTools, setExpandedTools] = React.useState<Record<string, boolean>>({});

  // Add Server Dialog State
  const [showAddModal, setShowAddModal] = React.useState(false);
  const [transportType, setTransportType] = React.useState<'stdio' | 'remote'>('stdio');
  const [formName, setFormName] = React.useState('');
  const [formCommand, setFormCommand] = React.useState('');
  const [formArgs, setFormArgs] = React.useState('');
  const [formUrl, setFormUrl] = React.useState('');
  const [formEnv, setFormEnv] = React.useState('');
  const [formHeaders, setFormHeaders] = React.useState('');
  const [formScope, setFormScope] = React.useState('global');
  const [formClients, setFormClients] = React.useState('all');
  const [submitting, setSubmitting] = React.useState(false);

  // Mutation in-flight tracking
  const [togglingServer, setTogglingServer] = React.useState<string | null>(null);
  const [deletingServer, setDeletingServer] = React.useState<string | null>(null);

  const { toast } = useToast();

  const fetchServers = React.useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const profileParam =
        selectedScope === 'all' || selectedScope === 'global' ? undefined : selectedScope;
      const data = await getMcpServers(profileParam);
      setServers(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch MCP servers';
      setError(msg);
      setServers([]);
    } finally {
      setLoading(false);
    }
  }, [selectedScope]);

  React.useEffect(() => {
    fetchServers();
  }, [fetchServers]);

  const copyConfig = (srv: MCPServerDTO) => {
    let payload: Record<string, unknown>;
    if (srv.type === 'http' || srv.type === 'sse' || srv.url) {
      payload = {
        [srv.name]: {
          url: srv.url || (srv.args && srv.args[0]) || '',
          headers: srv.headers || {},
          disabled: srv.status === 'disabled',
        },
      };
    } else {
      payload = {
        [srv.name]: {
          command: srv.command,
          args: srv.args || [],
          env: srv.env || {},
          disabled: srv.status === 'disabled',
        },
      };
    }

    navigator.clipboard.writeText(JSON.stringify(payload, null, 2));
    setCopiedName(srv.name);
    toast({
      title: 'Configuration copied',
      description: `JSON definition for ${srv.name} copied to clipboard`,
    });
    setTimeout(() => {
      setCopiedName(null);
    }, 2000);
  };

  const handleToggleState = async (srv: MCPServerDTO) => {
    const isCurrentlyEnabled = srv.status !== 'disabled';
    const nextState = !isCurrentlyEnabled;
    const key = `${srv.scope}-${srv.name}`;
    setTogglingServer(key);

    try {
      await setMcpServerState(srv.name, {
        enabled: nextState,
        profile: srv.scope === 'global' ? '' : srv.scope,
      });

      toast({
        title: nextState ? 'Server enabled' : 'Server disabled',
        description: `MCP server ${srv.name} is now ${nextState ? 'enabled' : 'disabled'}.`,
      });
      await fetchServers();
    } catch (err) {
      toast({
        title: 'Failed to update state',
        description: err instanceof Error ? err.message : String(err),
        variant: 'destructive',
      });
    } finally {
      setTogglingServer(null);
    }
  };

  const handleDelete = async (srv: MCPServerDTO) => {
    const key = `${srv.scope}-${srv.name}`;
    setDeletingServer(key);

    try {
      await deleteMcpServer(srv.name, srv.scope === 'global' ? '' : srv.scope);
      toast({
        title: 'Server removed',
        description: `MCP server ${srv.name} has been unregistered.`,
      });
      await fetchServers();
    } catch (err) {
      toast({
        title: 'Failed to delete server',
        description: err instanceof Error ? err.message : String(err),
        variant: 'destructive',
      });
    } finally {
      setDeletingServer(null);
    }
  };

  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const name = formName.trim();
    if (!name) {
      toast({ title: 'Validation error', description: 'Server name is required', variant: 'destructive' });
      return;
    }

    setSubmitting(true);
    try {
      if (transportType === 'stdio') {
        const cmd = formCommand.trim();
        if (!cmd) {
          toast({ title: 'Validation error', description: 'Command is required for stdio transport', variant: 'destructive' });
          setSubmitting(false);
          return;
        }

        const args = formArgs.trim() ? formArgs.trim().split(/\s+/) : [];
        const envMap: Record<string, string> = {};
        if (formEnv.trim()) {
          formEnv.split('\n').forEach((line) => {
            const idx = line.indexOf('=');
            if (idx > 0) {
              const k = line.substring(0, idx).trim();
              const v = line.substring(idx + 1).trim();
              if (k) envMap[k] = v;
            }
          });
        }

        await addMcpServer({
          name,
          command: cmd,
          args,
          env: envMap,
          profile: formScope === 'global' ? '' : formScope,
          clients: formClients === 'all' ? ['all'] : [formClients],
        });
      } else {
        const url = formUrl.trim();
        if (!url) {
          toast({ title: 'Validation error', description: 'Endpoint URL is required for remote transport', variant: 'destructive' });
          setSubmitting(false);
          return;
        }

        const headersMap: Record<string, string> = {};
        if (formHeaders.trim()) {
          formHeaders.split('\n').forEach((line) => {
            const idx = line.indexOf(':');
            if (idx > 0) {
              const k = line.substring(0, idx).trim();
              const v = line.substring(idx + 1).trim();
              if (k) headersMap[k] = v;
            }
          });
        }

        await addMcpServer({
          name,
          url,
          headers: headersMap,
          profile: formScope === 'global' ? '' : formScope,
          clients: formClients === 'all' ? ['all'] : [formClients],
        });
      }

      toast({
        title: 'MCP server added',
        description: `Successfully registered ${name} across selected targets.`,
      });
      setShowAddModal(false);
      setFormName('');
      setFormCommand('');
      setFormArgs('');
      setFormUrl('');
      setFormEnv('');
      setFormHeaders('');
      await fetchServers();
    } catch (err) {
      toast({
        title: 'Failed to add server',
        description: err instanceof Error ? err.message : String(err),
        variant: 'destructive',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const filteredServers = React.useMemo(() => {
    return servers.filter((srv) => {
      if (selectedScope && selectedScope !== 'all') {
        if (selectedScope === 'global' && srv.scope !== 'global') {
          return false;
        }
        if (selectedScope !== 'global' && srv.scope !== selectedScope) {
          return false;
        }
      }

      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const matchesName = srv.name.toLowerCase().includes(q);
        const matchesCmd = (srv.command || '').toLowerCase().includes(q);
        const matchesUrl = (srv.url || '').toLowerCase().includes(q);
        const matchesArgs = (srv.args || []).some((arg) => arg.toLowerCase().includes(q));
        const matchesEnv =
          srv.env &&
          Object.entries(srv.env).some(
            ([k, v]) => k.toLowerCase().includes(q) || v.toLowerCase().includes(q)
          );
        return matchesName || matchesCmd || matchesUrl || matchesArgs || matchesEnv;
      }

      return true;
    });
  }, [servers, selectedScope, searchQuery]);

  const profileScopeOptions = React.useMemo(() => {
    const fromProfiles = profiles.map((p) => p.name);
    const fromServers = servers
      .map((s) => s.scope)
      .filter((sc) => sc && sc !== 'global');
    return Array.from(new Set([...fromProfiles, ...fromServers])).sort();
  }, [profiles, servers]);

  const resetFilters = () => {
    setSearchQuery('');
    if (onSelectScope) {
      onSelectScope('all');
    }
  };

  const hasActiveFilters = (selectedScope && selectedScope !== 'all') || searchQuery.trim().length > 0;

  return (
    <div className="space-y-6">
      {/* Header and Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-5">
        <div>
          <div className="flex items-center gap-2.5">
            <h3 className="font-semibold text-sm text-foreground">
              Model Context Protocol (MCP) Catalog
            </h3>
            <span
              className={`text-xs font-mono px-1.5 py-0.5 rounded border ${
                filteredServers.length > 0
                  ? 'border-border bg-card text-foreground'
                  : 'border-border/60 bg-secondary/30 text-muted-foreground'
              }`}
            >
              {filteredServers.length} {filteredServers.length === 1 ? 'server' : 'servers'}
            </span>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl leading-relaxed mt-1">
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
          <Button
            size="sm"
            onClick={() => setShowAddModal(true)}
            className="rounded-geist h-8 font-mono text-xs border border-transparent bg-foreground text-background hover:bg-foreground/90 cursor-pointer shadow-xs"
          >
            <Plus className="h-3.5 w-3.5 mr-1.5" />
            Add Server
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search input */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="Search MCP servers by name, command, or endpoint..."
            className="pl-9 pr-8 h-8 text-xs font-sans rounded-geist border-border bg-card text-foreground placeholder:text-muted-foreground focus:border-border"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button
              type="button"
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-2 text-muted-foreground hover:text-foreground cursor-pointer"
              title="Clear search"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* Scope Dropdown Filter */}
        <div className="flex items-center gap-2 min-w-[160px]">
          <span className="text-xs text-muted-foreground font-mono hidden sm:inline">Scope:</span>
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
          <span className="flex items-center gap-1 text-muted-foreground text-xs font-mono">
            <Filter className="h-3 w-3" /> Filters:
          </span>

          {selectedScope && selectedScope !== 'all' && (
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-border bg-secondary/50 text-muted-foreground font-mono text-xs">
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
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-geist border border-border bg-secondary/50 text-muted-foreground font-mono text-xs">
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
            className="text-xs font-mono text-muted-foreground hover:text-foreground underline ml-1 cursor-pointer"
          >
            Reset all
          </button>
        </div>
      )}

      {error && (
        <div className="rounded-geist border border-status-warning-border bg-status-warning-bg px-3.5 py-2 text-xs text-status-warning font-sans flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <AlertCircle className="h-3.5 w-3.5 shrink-0" />
            <span className="truncate">Failed to load MCP catalog: {error}</span>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={fetchServers}
            disabled={loading}
            className="h-6 px-2 text-xs font-mono border-status-warning-border bg-[#161616] text-status-warning hover:bg-[#202020] rounded-geist cursor-pointer shrink-0"
          >
            Retry
          </Button>
        </div>
      )}

      {/* Grid of Servers */}
      {loading && servers.length === 0 ? (
        <McpCatalogSkeleton count={3} />
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 items-start">
          {filteredServers.map((srv) => {
            const isGlobal = srv.scope === 'global';
            const isRemote = srv.type === 'http' || srv.type === 'sse' || !!srv.url;
            const isEnabled = srv.status !== 'disabled';
            const cardKey = `${srv.scope}-${srv.name}`;
            const isToggling = togglingServer === cardKey;
            const isDeleting = deletingServer === cardKey;

            return (
              <Card
                key={cardKey}
                className="group relative overflow-hidden transition-colors duration-150 border border-border bg-card hover:border-border/80 rounded-geist flex flex-col"
              >
                <CardHeader className="pb-3 p-5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <div className="p-1.5 rounded-geist bg-secondary/50 border border-border text-foreground">
                        {isRemote ? <Globe className="h-3.5 w-3.5 text-accent-cyan" /> : <Server className="h-3.5 w-3.5" />}
                      </div>
                      <CardTitle className="text-sm font-semibold text-foreground font-mono truncate max-w-[140px] sm:max-w-[180px]">
                        {srv.name}
                      </CardTitle>
                    </div>

                    <div className="flex items-center gap-1.5 flex-wrap">
                      <span className="text-xs font-mono px-1.5 py-0.5 rounded-[3px] border border-border bg-secondary/40 text-muted-foreground">
                        {srv.type || (isRemote ? 'http' : 'stdio')}
                      </span>

                      {isEnabled ? (
                        <span className="badge-success text-xs font-mono px-1.5 py-0.5 rounded-[3px] flex items-center gap-1">
                          <span className="h-1.5 w-1.5 rounded-full bg-status-success-solid" />
                          Active
                        </span>
                      ) : (
                        <span className="text-xs font-mono px-1.5 py-0.5 rounded-[3px] border border-border bg-secondary/40 text-muted-foreground flex items-center gap-1">
                          <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground" />
                          Disabled
                        </span>
                      )}

                      <Badge
                        variant="outline"
                        className="text-xs font-mono flex items-center gap-1.5 rounded-geist px-2 py-0.5 border-border bg-secondary/40 text-muted-foreground"
                      >
                        {isGlobal ? (
                          <>
                            <Globe className="h-3 w-3 text-foreground" />
                            Global
                          </>
                        ) : (
                          <>
                            <Lock className="h-3 w-3 text-accent-orange" />
                            {srv.scope}
                          </>
                        )}
                      </Badge>
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="space-y-3 pb-4 p-5 pt-0 flex-1 flex flex-col">
                  {/* Remote URL or Stdio Command */}
                  {isRemote ? (
                    <div className="space-y-1">
                      <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Globe className="h-3 w-3 text-muted-foreground/70" /> Endpoint URL
                      </span>
                      <div className="font-mono text-xs p-2 rounded-geist bg-background border border-border text-foreground overflow-x-auto whitespace-pre select-all">
                        {srv.url || (srv.args && srv.args[0]) || '—'}
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-1">
                      <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Terminal className="h-3 w-3 text-muted-foreground/70" /> Command
                      </span>
                      <div className="font-mono text-xs p-2 rounded-geist bg-background border border-border text-foreground overflow-x-auto whitespace-pre">
                        <span className="text-foreground font-medium">{srv.command}</span>{' '}
                        <span className="text-muted-foreground">{(srv.args || []).join(' ')}</span>
                      </div>
                    </div>
                  )}

                  {/* Remote Auth Badge if present */}
                  {isRemote && (
                    <div className="flex items-center justify-between text-xs font-mono bg-background p-2 rounded-geist border border-border">
                      <span className="text-muted-foreground flex items-center gap-1.5">
                        <Key className="h-3 w-3 text-muted-foreground" /> Auth Status
                      </span>
                      {srv.auth === 'connected' ? (
                        <span className="text-status-success font-semibold flex items-center gap-1">
                          <ShieldCheck className="h-3.5 w-3.5" /> Authenticated
                        </span>
                      ) : srv.auth === 'OAuth' ? (
                        <span className="text-[#61afef] font-semibold flex items-center gap-1">
                          <Key className="h-3.5 w-3.5" /> OAuth Protected
                        </span>
                      ) : srv.auth === 'expired' ? (
                        <span className="text-status-warning font-semibold flex items-center gap-1">
                          <AlertCircle className="h-3.5 w-3.5" /> Token Expired
                        </span>
                      ) : (
                        <span className="text-muted-foreground">Standard / Public</span>
                      )}
                    </div>
                  )}

                  {/* Environment Variables if present */}
                  {srv.env && Object.keys(srv.env).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Layers className="h-3 w-3 text-muted-foreground/70" /> Environment
                      </span>
                      <div className="space-y-1 font-mono text-xs bg-background p-2 rounded-geist border border-border max-h-32 overflow-y-auto">
                        {Object.entries(srv.env).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-muted-foreground">
                            <span className="text-foreground font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-muted-foreground">{v}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* HTTP Headers if present */}
                  {srv.headers && Object.keys(srv.headers).length > 0 && (
                    <div className="space-y-1">
                      <span className="text-xs font-mono uppercase tracking-wider text-muted-foreground flex items-center gap-1">
                        <Layers className="h-3 w-3 text-muted-foreground/70" /> Headers
                      </span>
                      <div className="space-y-1 font-mono text-xs bg-background p-2 rounded-geist border border-border">
                        {Object.entries(srv.headers).map(([k, v]) => (
                          <div key={k} className="flex justify-between text-muted-foreground">
                            <span className="text-foreground font-medium">{k}:</span>
                            <span className="truncate max-w-[160px] text-muted-foreground">
                              {k.toLowerCase().includes('auth') ? '••••••••' : v}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Collapsible Registered Tools */}
                  {(() => {
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
                          className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs font-mono text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors duration-150 cursor-pointer"
                        >
                          <span className="flex items-center gap-1.5">
                            <Wrench className="h-3 w-3 text-muted-foreground" />
                            <span>Registered Tools</span>
                            <span className="text-xs text-muted-foreground">
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
                          <div className="p-2 border-t border-border space-y-1.5 bg-background max-h-60 overflow-y-auto">
                            {toolsList.map((t) => (
                              <div
                                key={t.name}
                                className="p-2 rounded-[3px] bg-secondary/30 border border-border/60 space-y-1"
                              >
                                <div className="flex items-center justify-between gap-2 flex-wrap">
                                  <div className="font-mono text-xs font-semibold text-foreground">
                                    {t.name}
                                  </div>
                                  {t.params && t.params.length > 0 && (
                                    <div className="flex items-center gap-1 flex-wrap">
                                      {t.params.map((p) => (
                                        <span
                                          key={p}
                                          className="text-xs font-mono px-1 py-0 rounded bg-secondary text-muted-foreground border border-border"
                                        >
                                          {p}
                                        </span>
                                      ))}
                                    </div>
                                  )}
                                </div>
                                <div className="text-xs text-muted-foreground font-sans leading-relaxed">
                                  {t.description}
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    );
                  })()}

                  {/* Card Action Footer */}
                  <div className="pt-2 flex items-center justify-between mt-auto border-t border-border/40">
                    <div className="flex items-center gap-1.5">
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={isToggling}
                        onClick={() => handleToggleState(srv)}
                        className={`h-7 px-2 text-xs font-mono border transition-colors duration-150 rounded-geist cursor-pointer ${
                          isEnabled
                            ? 'border-border/60 text-muted-foreground hover:text-foreground hover:bg-secondary'
                            : 'border-status-success/30 text-status-success hover:bg-status-success/10'
                        }`}
                      >
                        <Power className={`h-3 w-3 mr-1 ${isToggling ? 'animate-spin' : ''}`} />
                        {isEnabled ? 'Disable' : 'Enable'}
                      </Button>

                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={isDeleting}
                        onClick={() => handleDelete(srv)}
                        className="h-7 px-2 text-xs font-mono text-muted-foreground hover:text-status-error hover:bg-secondary rounded-geist cursor-pointer border border-transparent transition-colors duration-150"
                        title="Remove server"
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>

                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs font-mono text-muted-foreground hover:text-foreground hover:bg-secondary rounded-geist cursor-pointer border border-transparent transition-colors duration-150"
                      onClick={() => copyConfig(srv)}
                    >
                      {copiedName === srv.name ? (
                        <Check className="h-3.5 w-3.5 mr-1 text-status-success" />
                      ) : (
                        <Copy className="h-3.5 w-3.5 mr-1 text-muted-foreground" />
                      )}
                      {copiedName === srv.name ? 'Copied' : 'JSON'}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Add MCP Server Modal */}
      <Dialog open={showAddModal} onOpenChange={setShowAddModal}>
        <DialogContent className="border border-border bg-card text-foreground rounded-geist sm:max-w-lg shadow-2xl">
          <DialogHeader className="space-y-1.5 pb-2 border-b border-border">
            <DialogTitle className="text-base font-semibold tracking-tight text-foreground flex items-center gap-2">
              <Server className="h-4 w-4 text-accent-blue" />
              Register MCP Server
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              Add a new Model Context Protocol tool provider across your agent environment.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleAddSubmit} className="space-y-4 py-2">
            {/* Transport Type Segmented Control */}
            <div className="space-y-1.5">
              <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                Transport Type
              </label>
              <div>
                <SegmentedControl
                  value={transportType}
                  onChange={(val) => setTransportType(val as 'stdio' | 'remote')}
                  size="sm"
                  options={[
                    { value: 'stdio', label: 'Stdio Command', icon: <Terminal className="h-3.5 w-3.5" /> },
                    { value: 'remote', label: 'Remote HTTP / SSE', icon: <Globe className="h-3.5 w-3.5" /> },
                  ]}
                />
              </div>
            </div>

            {/* Server Name */}
            <div className="space-y-1.5">
              <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                Server Name *
              </label>
              <Input
                required
                placeholder="e.g. posthog, fetch, github-tools"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                className="rounded-geist border-border bg-background text-xs h-8 text-foreground placeholder:text-muted-foreground focus:border-border font-mono"
              />
            </div>

            {transportType === 'stdio' ? (
              <>
                {/* Command */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                    Executable Command *
                  </label>
                  <Input
                    required
                    placeholder="e.g. npx, uvx, python3, docker"
                    value={formCommand}
                    onChange={(e) => setFormCommand(e.target.value)}
                    className="rounded-geist border-border bg-background text-xs h-8 text-foreground placeholder:text-muted-foreground focus:border-border font-mono"
                  />
                </div>

                {/* Arguments */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                    Arguments (space separated)
                  </label>
                  <Input
                    placeholder="e.g. -y @modelcontextprotocol/server-postgres"
                    value={formArgs}
                    onChange={(e) => setFormArgs(e.target.value)}
                    className="rounded-geist border-border bg-background text-xs h-8 text-foreground placeholder:text-muted-foreground focus:border-border font-mono"
                  />
                </div>

                {/* Environment Variables */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                    Environment Variables (KEY=value per line)
                  </label>
                  <textarea
                    rows={2}
                    placeholder="API_KEY=secret_123&#10;DEBUG=true"
                    value={formEnv}
                    onChange={(e) => setFormEnv(e.target.value)}
                    className="w-full rounded-geist border border-border bg-background p-2 text-xs text-foreground placeholder:text-muted-foreground focus:border-border font-mono resize-none focus:outline-none"
                  />
                </div>
              </>
            ) : (
              <>
                {/* Endpoint URL */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                    Endpoint URL *
                  </label>
                  <Input
                    required
                    type="url"
                    placeholder="e.g. https://mcp.posthog.com/mcp"
                    value={formUrl}
                    onChange={(e) => setFormUrl(e.target.value)}
                    className="rounded-geist border-border bg-background text-xs h-8 text-foreground placeholder:text-muted-foreground focus:border-border font-mono"
                  />
                </div>

                {/* Custom HTTP Headers */}
                <div className="space-y-1.5">
                  <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                    HTTP Headers (Header-Name: Value per line)
                  </label>
                  <textarea
                    rows={2}
                    placeholder="Authorization: Bearer my-token&#10;X-Custom-Header: value"
                    value={formHeaders}
                    onChange={(e) => setFormHeaders(e.target.value)}
                    className="w-full rounded-geist border border-border bg-background p-2 text-xs text-foreground placeholder:text-muted-foreground focus:border-border font-mono resize-none focus:outline-none"
                  />
                </div>
              </>
            )}

            {/* Scope and Client Target Grid */}
            <div className="grid grid-cols-2 gap-3 pt-1">
              <div className="space-y-1.5">
                <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                  Target Scope
                </label>
                <SelectDropdown
                  value={formScope}
                  onChange={(val) => setFormScope(String(val))}
                  size="sm"
                  options={[
                    { value: 'global', label: 'Global (Host)' },
                    ...profileScopeOptions.map((name) => ({
                      value: name,
                      label: `Profile: ${name}`,
                    })),
                  ]}
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-mono font-medium uppercase tracking-wider text-muted-foreground">
                  Target Clients
                </label>
                <SelectDropdown
                  value={formClients}
                  onChange={(val) => setFormClients(String(val))}
                  size="sm"
                  options={[
                    { value: 'all', label: 'All Installed Clients' },
                    { value: 'claude', label: 'Claude Code Only' },
                    { value: 'codex', label: 'OpenAI Codex Only' },
                    { value: 'agy', label: 'Antigravity Only' },
                  ]}
                />
              </div>
            </div>

            <DialogFooter className="pt-3 border-t border-border flex items-center justify-end gap-2">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setShowAddModal(false)}
                disabled={submitting}
                className="h-8 text-xs font-mono text-muted-foreground hover:text-foreground cursor-pointer rounded-geist"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                size="sm"
                disabled={submitting}
                className="h-8 text-xs font-mono bg-foreground text-background hover:bg-foreground/90 cursor-pointer rounded-geist shadow-xs"
              >
                {submitting ? 'Registering...' : 'Register Server'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
