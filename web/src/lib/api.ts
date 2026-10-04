export interface QuotaDTO {
  bottleneck_pct: number;
  summary: string;
  is_exhausted: boolean;
}

export interface AccountInfo {
  email?: string;
  name?: string;
  auth_method?: string;
  project_id?: string;
}

export interface ProfileDTO {
  agent: string;
  name: string;
  path: string;
  has_credentials: boolean;
  account?: AccountInfo;
  quota?: QuotaDTO;
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
}

export interface SessionFilter {
  agent?: string;
  profile?: string;
  query?: string;
  limit?: number;
}

export interface ResumeRequest {
  agent: string;
  profile: string;
  session_id: string;
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

export async function getProfiles(agent?: string): Promise<ProfileDTO[]> {
  const url = agent ? `${BASE_URL}/profiles?agent=${encodeURIComponent(agent)}` : `${BASE_URL}/profiles`;
  const res = await fetch(url);
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

export async function deleteProfile(agent: string, name: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/profiles/${encodeURIComponent(agent)}/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
  if (!res.ok && res.status === 404) {
    // Fallback if server uses query parameters
    const fallbackRes = await fetch(`${BASE_URL}/profiles?agent=${encodeURIComponent(agent)}&name=${encodeURIComponent(name)}`, {
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
  if (filter?.query) {
    params.set('q', filter.query);
    params.set('query', filter.query);
  }
  if (filter?.limit) params.set('limit', filter.limit.toString());

  const queryString = params.toString();
  const url = queryString ? `${BASE_URL}/sessions?${queryString}` : `${BASE_URL}/sessions`;
  const res = await fetch(url);
  return handleResponse<SessionDTO[]>(res);
}

export async function resumeSession(req: ResumeRequest): Promise<{ success: boolean; message?: string }> {
  const res = await fetch(`${BASE_URL}/sessions/resume`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(req),
  });
  return handleResponse<{ success: boolean; message?: string }>(res);
}

export async function getMcpServers(profile?: string): Promise<MCPServerDTO[]> {
  const url = profile ? `${BASE_URL}/mcp?profile=${encodeURIComponent(profile)}` : `${BASE_URL}/mcp`;
  const res = await fetch(url);
  return handleResponse<MCPServerDTO[]>(res);
}
