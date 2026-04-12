const BASE = process.env.NEXT_PUBLIC_DISPATCH_API ?? "http://localhost:8080";

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

export async function listScenarios(): Promise<Scenario[]> {
  const res = await fetch(`${BASE}/api/simulation/scenarios`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function runScenario(scenario: string): Promise<Experiment> {
  const res = await fetch(`${BASE}/api/simulation/run`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ scenario }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function getExperiment(id: string): Promise<Experiment> {
  const res = await fetch(`${BASE}/api/experiments/${id}`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function listExperiments(): Promise<Experiment[]> {
  const res = await fetch(`${BASE}/api/experiments`);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function promoteExperiment(id: string): Promise<Experiment> {
  const res = await fetch(`${BASE}/api/experiments/${id}/promote`, {
    method: "POST",
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function holdExperiment(id: string, reason: string): Promise<Experiment> {
  const res = await fetch(`${BASE}/api/experiments/${id}/hold`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ reason }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
