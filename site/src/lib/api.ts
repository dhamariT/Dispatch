const BASE = process.env.NEXT_PUBLIC_DISPATCH_API ?? "http://localhost:8080";

// Every API call from the browser must carry the session cookie, so
// `credentials: "include"` is the default for this module. The backend
// echoes the frontend origin on Access-Control-Allow-Origin (never "*")
// so the cookie is allowed cross-origin during local dev.
async function request<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init.headers ?? {}),
    },
  });
  if (res.status === 204) {
    return undefined as T;
  }
  if (!res.ok) {
    throw new ApiError(res.status, (await res.text()).trim() || res.statusText);
  }
  return (await res.json()) as T;
}

export class ApiError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}

// --- Auth / identity --------------------------------------------------------

export interface Me {
  id: string;
  email: string;
  login: string;
  name: string;
  avatar_url: string;
}

export function githubLoginUrl(): string {
  return `${BASE}/api/auth/github/login`;
}

export async function getMe(): Promise<Me> {
  return request<Me>("/api/auth/me");
}

export async function logout(): Promise<void> {
  await request<void>("/api/auth/logout", { method: "POST" });
}

// --- Orgs -------------------------------------------------------------------

export interface OrgSummary {
  id: string;
  slug: string;
  name: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface Org {
  id: string;
  slug: string;
  name: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export async function listMyOrgs(): Promise<OrgSummary[]> {
  return request<OrgSummary[]>("/api/orgs");
}

export async function createOrg(slug: string, name: string): Promise<Org> {
  return request<Org>("/api/orgs", {
    method: "POST",
    body: JSON.stringify({ slug, name }),
  });
}

export async function getOrg(slug: string): Promise<Org> {
  return request<Org>(`/api/orgs/${encodeURIComponent(slug)}`);
}

export interface Member {
  organization_id: string;
  user_id: string;
  role: string;
  created_at: string;
  email: string;
  login: string;
  name: string;
  avatar_url: string;
}

export async function listMembers(slug: string): Promise<Member[]> {
  return request<Member[]>(`/api/orgs/${encodeURIComponent(slug)}/members`);
}

export async function removeMember(slug: string, userID: string): Promise<void> {
  await request<void>(
    `/api/orgs/${encodeURIComponent(slug)}/members/${encodeURIComponent(userID)}`,
    { method: "DELETE" },
  );
}

// --- Invites ----------------------------------------------------------------

export interface Invite {
  id: string;
  organization_id: string;
  email: string;
  role: string;
  created_by: string;
  created_at: string;
  expires_at: string;
  accepted_at: string | null;
  accepted_by: string | null;
}

export interface CreatedInvite {
  id: string;
  organization_id: string;
  email: string;
  role: string;
  expires_at: string;
  token: string;
}

export async function createInvite(
  slug: string,
  role: string,
  email: string,
): Promise<CreatedInvite> {
  return request<CreatedInvite>(
    `/api/orgs/${encodeURIComponent(slug)}/invites`,
    {
      method: "POST",
      body: JSON.stringify({ role, email }),
    },
  );
}

export async function listInvites(slug: string): Promise<Invite[]> {
  return request<Invite[]>(`/api/orgs/${encodeURIComponent(slug)}/invites`);
}

export async function deleteInvite(slug: string, id: string): Promise<void> {
  await request<void>(
    `/api/orgs/${encodeURIComponent(slug)}/invites/${encodeURIComponent(id)}`,
    { method: "DELETE" },
  );
}

export interface InviteLookup {
  organization_slug: string;
  organization_name: string;
  role: string;
  email: string;
}

export async function getInvite(token: string): Promise<InviteLookup> {
  return request<InviteLookup>(`/api/invites/${encodeURIComponent(token)}`);
}

export async function acceptInvite(token: string): Promise<Org> {
  return request<Org>(`/api/invites/${encodeURIComponent(token)}/accept`, {
    method: "POST",
  });
}

// --- Experiments (org-scoped) ----------------------------------------------

export interface ExperimentMetric {
  metric_name: string;
  direction: number;
  canary_mean: number;
  canary_sd: number;
  canary_n: number;
  control_mean: number;
  control_sd: number;
  control_n: number;
  t_statistic: number;
  p_value: number;
  effect_size: number;
  verdict: "regression" | "improvement" | "no_change" | "insufficient_data";
}

export interface Experiment {
  id: string;
  org_id: string;
  deploy_id: string;
  status: "collecting" | "analyzing" | "decided";
  decision: "promote" | "hold" | "auto_hold" | "";
  hold_reason: string;
  canary_devices: string[];
  control_devices: string[];
  window_minutes: number;
  started_at: string;
  results: ExperimentMetric[] | null;
}

export interface Scenario {
  name: string;
  description: string;
}

export async function listScenarios(slug: string): Promise<Scenario[]> {
  return request<Scenario[]>(
    `/api/orgs/${encodeURIComponent(slug)}/simulation/scenarios`,
  );
}

export async function runScenario(
  slug: string,
  scenario: string,
): Promise<Experiment> {
  return request<Experiment>(
    `/api/orgs/${encodeURIComponent(slug)}/simulation/run`,
    {
      method: "POST",
      body: JSON.stringify({ scenario }),
    },
  );
}

export async function getExperiment(
  slug: string,
  id: string,
): Promise<Experiment> {
  return request<Experiment>(
    `/api/orgs/${encodeURIComponent(slug)}/experiments/${encodeURIComponent(id)}`,
  );
}

export async function listExperiments(slug: string): Promise<Experiment[]> {
  return request<Experiment[]>(
    `/api/orgs/${encodeURIComponent(slug)}/experiments`,
  );
}

export async function promoteExperiment(
  slug: string,
  id: string,
): Promise<Experiment> {
  return request<Experiment>(
    `/api/orgs/${encodeURIComponent(slug)}/experiments/${encodeURIComponent(id)}/promote`,
    { method: "POST" },
  );
}

export async function holdExperiment(
  slug: string,
  id: string,
  reason: string,
): Promise<Experiment> {
  return request<Experiment>(
    `/api/orgs/${encodeURIComponent(slug)}/experiments/${encodeURIComponent(id)}/hold`,
    {
      method: "POST",
      body: JSON.stringify({ reason }),
    },
  );
}
