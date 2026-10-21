export interface QuotaWindowDTO {
  category?: string;
  name: string;
  remaining_pct: number;
  resets_at?: string;
  resets_in_sec?: number;
}

export interface QuotaDTO {
  bottleneck_pct: number;
  summary: string;
  is_exhausted: boolean;
  windows?: QuotaWindowDTO[];
  credits?: string;
  fetched_at?: string;
}

export interface AccountInfo {
  email?: string;
  name?: string;
  auth_method?: string;
  project_id?: string;
}

export interface AdapterInfo {
  agent: string;
  has_credentials: boolean;
  account?: AccountInfo;
  quota?: QuotaDTO;
}

export interface ProfileDTO {
  agent: string;
  name: string;
  path: string;
  has_credentials: boolean;
  account?: AccountInfo;
  quota?: QuotaDTO;
  mcp_global?: boolean;
  plugins_global?: boolean;
  adapters?: AdapterInfo[];
}

export interface CreateProfileRequest {
  agent: string;
  name: string;
  email?: string;
  clone_from?: string;
}

export interface SessionDTO {
  id: string;
  agent: string;
  profile: string;
  title: string;
  cwd: string;
  goal: string;
  turns: number;
  updated_at: string;
  is_active: boolean;
  pid?: number;
}

export interface SessionFilter {
  agent?: string;
  profile?: string;
  cwd?: string;
  query?: string;
  limit?: number;
  force?: boolean;
}

export interface ResumeRequest {
  agent: string;
  profile: string;
  session_id: string;
  terminal?: string;
  flags?: string[];
  aim_flags?: string[];
  adapter_flags?: string[];
  custom_flags?: string;
  custom_args?: string;
}

export interface RunnableFlagDTO {
  flag: string;
  name: string;
  description: string;
  type: 'bool' | 'string' | 'choice';
  default?: string;
  choices?: string[];
  dangerous?: boolean;
}

export interface AdapterDTO {
  name: string;
  display_name: string;
  binary_name: string;
  aliases: string[];
  runnable_flags: RunnableFlagDTO[];
}

export interface ToolCallDTO {
  id?: string;
  tool_name: string;
  arguments?: string;
  output?: string;
  status?: string;
}

export interface SessionTurnDTO {
  turn_index: number;
  role: 'user' | 'assistant' | 'system';
  timestamp?: string;
  content: string;
  thinking?: string;
  tool_calls?: ToolCallDTO[];
}

export interface SessionTranscriptDTO {
  session_id: string;
  agent: string;
  profile: string;
  title: string;
  cwd: string;
  goal: string;
  turns: SessionTurnDTO[];
}

export interface MCPServerDTO {
  name: string;
  command: string;
  args: string[];
  env?: Record<string, string>;
  scope: string;
}

export interface StatusDTO {
  version: string;
  os?: string;
  arch?: string;
  uptime?: string;
  status?: string;
  active_sessions?: number;
  total_profiles?: number;
  terminals_available?: string[];
}

const BASE_URL = '/api';

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorMessage = `HTTP ${res.status}: ${res.statusText}`;
    try {
      const errorData = await res.json();
      if (errorData?.error) {
        errorMessage = errorData.error;
      } else if (errorData?.message) {
        errorMessage = errorData.message;
      }
    } catch {
      // Use fallback error message if JSON parsing fails
    }
    throw new Error(errorMessage);
  }
  return res.json() as Promise<T>;
}

export async function getStatus(): Promise<StatusDTO> {
  const res = await fetch(`${BASE_URL}/status`);
  return handleResponse<StatusDTO>(res);
}

export interface GetProfilesOptions {
  agent?: string;
  profile?: string;
  refresh?: boolean;
}

export async function getProfiles(
  agentOrOptions?: string | GetProfilesOptions,
  legacyRefresh?: boolean
): Promise<ProfileDTO[]> {
  let agent: string | undefined;
  let profile: string | undefined;
  let refresh = false;

  if (typeof agentOrOptions === 'object' && agentOrOptions !== null) {
    agent = agentOrOptions.agent;
    profile = agentOrOptions.profile;
    refresh = !!agentOrOptions.refresh;
  } else {
    agent = agentOrOptions;
    refresh = !!legacyRefresh;
  }

  const params = new URLSearchParams();
  if (agent && agent !== 'all') params.set('agent', agent);
  if (profile && profile !== 'all') params.set('profile', profile);
  if (refresh) params.set('refresh', 'true');

  const queryString = params.toString();
  const url = queryString ? `${BASE_URL}/profiles?${queryString}` : `${BASE_URL}/profiles`;
  const res = await fetch(url);
  return handleResponse<ProfileDTO[]>(res);
}

export async function refreshProfiles(agent?: string, profile?: string): Promise<ProfileDTO[]> {
  const res = await fetch(`${BASE_URL}/profiles/refresh`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ agent, profile }),
  });
  return handleResponse<ProfileDTO[]>(res);
}

export async function refreshSingleProfile(name: string, agent?: string): Promise<ProfileDTO[]> {
  const url = agent
    ? `${BASE_URL}/profiles/${encodeURIComponent(name)}/refresh?agent=${encodeURIComponent(agent)}`
    : `${BASE_URL}/profiles/${encodeURIComponent(name)}/refresh`;
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  return handleResponse<ProfileDTO[]>(res);
}

export async function createProfile(req: CreateProfileRequest): Promise<ProfileDTO> {
  const res = await fetch(`${BASE_URL}/profiles`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(req),
  });
  return handleResponse<ProfileDTO>(res);
}

export async function deleteProfile(agent: string | undefined, name: string): Promise<void> {
  const hasAgent = !!agent && agent.trim() !== '' && agent !== 'all';
  const url = hasAgent
    ? `${BASE_URL}/profiles/${encodeURIComponent(agent)}/${encodeURIComponent(name)}`
    : `${BASE_URL}/profiles/${encodeURIComponent(name)}`;

  const res = await fetch(url, {
    method: 'DELETE',
  });
  if (!res.ok && res.status === 404) {
    // Fallback if server uses query parameters
    const query = hasAgent
      ? `agent=${encodeURIComponent(agent)}&name=${encodeURIComponent(name)}`
      : `name=${encodeURIComponent(name)}`;
    const fallbackRes = await fetch(`${BASE_URL}/profiles?${query}`, {
      method: 'DELETE',
    });
    if (!fallbackRes.ok) {
      return handleResponse<void>(fallbackRes);
    }
    return;
  }
  return handleResponse<void>(res);
}

export async function getSessions(filter?: SessionFilter): Promise<SessionDTO[]> {
  const params = new URLSearchParams();
  if (filter?.agent) params.set('agent', filter.agent);
  if (filter?.profile) params.set('profile', filter.profile);
  if (filter?.cwd) params.set('cwd', filter.cwd);
  if (filter?.query) {
    params.set('q', filter.query);
    params.set('query', filter.query);
  }
  if (filter?.limit) params.set('limit', filter.limit.toString());
  if (filter?.force) params.set('force', 'true');

  const queryString = params.toString();
  const url = queryString ? `${BASE_URL}/sessions?${queryString}` : `${BASE_URL}/sessions`;
  const res = await fetch(url);
  return handleResponse<SessionDTO[]>(res);
}

export async function resumeSession(
  req: ResumeRequest
): Promise<{ success?: boolean; status?: string; terminal?: string; message?: string }> {
  const res = await fetch(`${BASE_URL}/sessions/resume`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(req),
  });
  return handleResponse<{ success?: boolean; status?: string; terminal?: string; message?: string }>(res);
}

export async function getMcpServers(profile?: string): Promise<MCPServerDTO[]> {
  const url = profile ? `${BASE_URL}/mcp?profile=${encodeURIComponent(profile)}` : `${BASE_URL}/mcp`;
  const res = await fetch(url);
  return handleResponse<MCPServerDTO[]>(res);
}

export async function updateProfileConfig(
  name: string,
  config: { mcp_global?: boolean; plugins_global?: boolean }
): Promise<void> {
  const res = await fetch(`${BASE_URL}/profiles/${encodeURIComponent(name)}/config`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(config),
  });
  return handleResponse<void>(res);
}

export interface DaemonDTO {
  installed: boolean;
  active: boolean;
  label: string;
  config_path: string;
  interval_sec: number;
  binary_path: string;
  log_path: string;
  last_run?: string;
  last_run_message?: string;
}

export interface DaemonLogsDTO {
  log_path: string;
  lines: string[];
  total_lines: number;
}

export async function getDaemonStatus(): Promise<DaemonDTO> {
  const res = await fetch(`${BASE_URL}/daemon`);
  return handleResponse<DaemonDTO>(res);
}

export async function getDaemonLogs(tail = 200, offset = 0): Promise<DaemonLogsDTO> {
  const res = await fetch(`${BASE_URL}/daemon/logs?tail=${tail}&offset=${offset}`);
  return handleResponse<DaemonLogsDTO>(res);
}

export async function installDaemon(binaryPath?: string): Promise<DaemonDTO> {
  const res = await fetch(`${BASE_URL}/daemon/install`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ binary_path: binaryPath }),
  });
  return handleResponse<DaemonDTO>(res);
}

export async function uninstallDaemon(): Promise<{ status: string }> {
  const res = await fetch(`${BASE_URL}/daemon/uninstall`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });
  return handleResponse<{ status: string }>(res);
}

export async function runDaemonOnce(): Promise<{ status: string; message: string; info?: DaemonDTO }> {
  const res = await fetch(`${BASE_URL}/daemon/run`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({}),
  });
  return handleResponse<{ status: string; message: string; info?: DaemonDTO }>(res);
}

export async function getAdapters(): Promise<AdapterDTO[]> {
  const res = await fetch(`${BASE_URL}/adapters`);
  return handleResponse<AdapterDTO[]>(res);
}

export async function getSessionTranscript(agent: string, id: string): Promise<SessionTranscriptDTO> {
  const res = await fetch(`${BASE_URL}/sessions/${encodeURIComponent(agent)}/${encodeURIComponent(id)}/transcript`);
  return handleResponse<SessionTranscriptDTO>(res);
}

export interface VPNStatusDTO {
  connected: boolean;
  interface_name: string;
  ip: string;
  broadcast_ip: string;
}

export interface SharedProfileConfigDTO {
  name: string;
  agent: string;
  auto_approve: boolean;
  max_lease_ttl: number;
  enabled: boolean;
}

export interface PeerInfoDTO {
  id: string;
  hostname: string;
  vpn_ip: string;
  port: number;
  last_seen: string;
  profiles: string[];
  shared_profiles?: SharedProfileConfigDTO[];
}

export interface ActiveLeaseDTO {
  local_profile: string;
  source_peer: string;
  source_profile: string;
  agent: string;
  granted_at: string;
  expires_at: string;
  is_expired: boolean;
}

export interface GrantedLeaseDTO {
  lease_id: string;
  profile_name: string;
  agent: string;
  borrower_id: string;
  borrower_ip: string;
  granted_at: string;
  expires_at: string;
  revoked: boolean;
  auto_approved: boolean;
}

export interface PendingRequestDTO {
  request_id: string;
  request: {
    requester_id: string;
    requester_ip: string;
    profile_name: string;
    agent: string;
    requested_hours: number;
  };
  created_at: string;
  expires_at: string;
}

export interface AuditEntryDTO {
  id: string;
  timestamp: string;
  action: string;
  actor: string;
  profile?: string;
  agent?: string;
  details: string;
}

export interface SchedulePolicyDTO {
  id: string;
  enabled: boolean;
  scope: string; // "global" or profile name
  mode: 'auto' | 'request';
  start_time: string; // "HH:MM" e.g. "00:00"
  end_time: string; // "HH:MM" e.g. "08:00"
  start_date?: string; // "YYYY-MM-DD"
  end_date?: string; // "YYYY-MM-DD"
  created_at?: string;
}

export interface ShareStatusDTO {
  vpn: VPNStatusDTO | null;
  global_mode: 'off' | 'request' | 'auto';
  effective_global_mode?: 'off' | 'request' | 'auto';
  shared_profiles: SharedProfileConfigDTO[];
  peers: PeerInfoDTO[];
  active_leases: ActiveLeaseDTO[];
  granted_leases?: GrantedLeaseDTO[];
  pending_requests?: PendingRequestDTO[];
  audit_logs?: AuditEntryDTO[];
  schedules?: SchedulePolicyDTO[];
}

export async function getShareStatus(refresh = false): Promise<ShareStatusDTO> {
  const url = refresh ? `${BASE_URL}/share/status?refresh=true` : `${BASE_URL}/share/status`;
  const res = await fetch(url);
  return handleResponse<ShareStatusDTO>(res);
}

export async function approveShareRequest(requestId: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/share/approve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ request_id: requestId }),
  });
  return handleResponse(res);
}

export async function denyShareRequest(requestId: string): Promise<any> {
  const res = await fetch(`${BASE_URL}/share/deny`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ request_id: requestId }),
  });
  return handleResponse(res);
}

export async function revokeShareLease(leaseId: string): Promise<{ ok: boolean; revoked: GrantedLeaseDTO }> {
  const res = await fetch(`${BASE_URL}/share/revoke`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lease_id: leaseId }),
  });
  return handleResponse<{ ok: boolean; revoked: GrantedLeaseDTO }>(res);
}

export async function getShareAudit(): Promise<AuditEntryDTO[]> {
  const res = await fetch(`${BASE_URL}/share/audit`);
  return handleResponse<AuditEntryDTO[]>(res);
}

export async function setGlobalShareMode(mode: 'off' | 'request' | 'auto'): Promise<{ ok: boolean; mode: string }> {
  const res = await fetch(`${BASE_URL}/share/global`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ mode }),
  });
  return handleResponse<{ ok: boolean; mode: string }>(res);
}

export async function toggleShareProfile(payload: {
  name: string;
  agent: string;
  auto_approve: boolean;
  ttl: string;
  enabled: boolean;
}): Promise<{ ok: boolean; enabled: boolean }> {
  const res = await fetch(`${BASE_URL}/share/toggle`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<{ ok: boolean; enabled: boolean }>(res);
}

export async function addSharePeer(host: string): Promise<PeerInfoDTO> {
  const res = await fetch(`${BASE_URL}/share/peers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ host }),
  });
  return handleResponse<PeerInfoDTO>(res);
}

export async function requestShareLease(payload: {
  peer: string;
  profile: string;
  agent: string;
  hours: number;
}): Promise<{ ok: boolean; local_profile?: string; agent?: string; expires_at?: string; status?: string; message?: string }> {
  const res = await fetch(`${BASE_URL}/share/request`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<{ ok: boolean; local_profile?: string; agent?: string; expires_at?: string; status?: string; message?: string }>(res);
}

export async function releaseShareLease(profile: string): Promise<{ ok: boolean }> {
  const res = await fetch(`${BASE_URL}/share/release`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ profile }),
  });
  return handleResponse<{ ok: boolean }>(res);
}

export async function cleanExpiredLeases(): Promise<{ ok: boolean; cleaned: number }> {
  const res = await fetch(`${BASE_URL}/share/clean-expired`, {
    method: 'POST',
  });
  return handleResponse<{ ok: boolean; cleaned: number }>(res);
}

export async function getSchedules(): Promise<SchedulePolicyDTO[]> {
  const res = await fetch(`${BASE_URL}/share/schedules`);
  return handleResponse<SchedulePolicyDTO[]>(res);
}

export async function saveSchedule(schedule: Partial<SchedulePolicyDTO>): Promise<{ ok: boolean; schedule: SchedulePolicyDTO }> {
  const res = await fetch(`${BASE_URL}/share/schedules`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(schedule),
  });
  return handleResponse<{ ok: boolean; schedule: SchedulePolicyDTO }>(res);
}

export async function deleteSchedule(id: string): Promise<{ ok: boolean }> {
  const res = await fetch(`${BASE_URL}/share/schedules`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id }),
  });
  return handleResponse<{ ok: boolean }>(res);
}
