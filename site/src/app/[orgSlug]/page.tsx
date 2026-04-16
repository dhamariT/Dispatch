"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { DemoBanner } from "@/components/demo-banner";
import { Nav } from "@/components/nav";
import { ScenarioPicker, type ScenarioOption } from "@/components/scenario-picker";
import { ExperimentVerdict } from "@/components/experiment-verdict";
import { ExperimentMetricsTable } from "@/components/experiment-metrics-table";
import type { ExperimentMetricRowProps } from "@/components/experiment-metric-row";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const scenarios: ScenarioOption[] = [
  {
    name: "lidar_regression",
    description:
      "Canary's LiDAR accuracy drops significantly. CPU spikes. Auto-hold triggers on 3 of 4 metrics.",
    outcome: "auto_hold",
    icon: "shield",
  },
  {
    name: "healthy_deploy",
    description:
      "All metrics stable across canary and control groups. No significant differences detected.",
    outcome: "promote",
    icon: "check",
  },
  {
    name: "noisy_but_safe",
    description:
      "High variance in both groups drowns out small differences. Tests the effect size gate.",
    outcome: "promote",
    icon: "activity",
  },
  {
    name: "slow_regression",
    description:
      "Small CPU regression with moderate effect size. Borderline case that tests the decision threshold.",
    outcome: "auto_hold",
    icon: "gauge",
  },
];

const metricFormats: Record<string, ExperimentMetricRowProps["format"]> = {
  lidar_accuracy: "percent",
  cpu_usage: "percent",
  memory_mb: "bytes",
  fusion_latency_ms: "ms",
};

const metricDisplayNames: Record<string, string> = {
  lidar_accuracy: "LiDAR accuracy",
  cpu_usage: "CPU usage",
  memory_mb: "Memory",
  fusion_latency_ms: "Fusion latency",
};

function toMetricRows(results: api.ExperimentMetric[]): ExperimentMetricRowProps[] {
  return results.map((r) => ({
    name: metricDisplayNames[r.metric_name] ?? r.metric_name,
    format: metricFormats[r.metric_name] ?? "number",
    canaryMean: r.canary_mean,
    canarySD: r.canary_sd,
    canaryN: r.canary_n,
    controlMean: r.control_mean,
    controlSD: r.control_sd,
    controlN: r.control_n,
    pValue: r.p_value,
    effectSize: r.effect_size,
    verdict: r.verdict,
  }));
}

export default function OrgDashboardPage() {
  const router = useRouter();
  const params = useParams<{ orgSlug: string }>();
  const slug = params.orgSlug;
  const { status: authStatus, user, signOut } = useAuth();

  const [org, setOrg] = useState<api.Org | null>(null);
  const [orgError, setOrgError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [running, setRunning] = useState(false);
  const [experiment, setExperiment] = useState<api.Experiment | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (authStatus === "anonymous") {
      router.replace(`/login?next=${encodeURIComponent(`/${slug}`)}`);
    }
  }, [authStatus, router, slug]);

  useEffect(() => {
    if (authStatus !== "authenticated") return;
    let alive = true;
    api
      .getOrg(slug)
      .then((o) => {
        if (alive) setOrg(o);
      })
      .catch((err) => {
        if (!alive) return;
        if (err instanceof api.ApiError && err.status === 404) {
          router.replace("/");
          return;
        }
        setOrgError(err instanceof Error ? err.message : "Failed to load organization");
      });
    return () => {
      alive = false;
    };
  }, [authStatus, router, slug]);

  const handleRun = useCallback(async () => {
    if (!selected) return;
    setRunning(true);
    setError(null);
    setExperiment(null);
    try {
      const exp = await api.runScenario(slug, selected);
      setExperiment(exp);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to run experiment");
    } finally {
      setRunning(false);
    }
  }, [selected, slug]);

  const handlePromote = useCallback(async () => {
    if (!experiment) return;
    try {
      const updated = await api.promoteExperiment(slug, experiment.id);
      setExperiment(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to promote");
    }
  }, [experiment, slug]);

  const handleHold = useCallback(async () => {
    if (!experiment) return;
    try {
      const updated = await api.holdExperiment(slug, experiment.id, "Manual hold from dashboard");
      setExperiment(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to hold");
    }
  }, [experiment, slug]);

  if (authStatus === "loading" || (authStatus === "authenticated" && !org && !orgError)) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }
  if (orgError) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-critical">
        {orgError}
      </div>
    );
  }
  if (!org) return null;

  const results = experiment?.results ?? [];
  const regressionCount = results.filter((r) => r.verdict === "regression").length;
  const canWrite = org.role === "admin" || org.role === "operator";

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Nav fleetName={org.name} />
      <DemoBanner />

      <main className="mx-auto flex w-full max-w-[1200px] flex-1 flex-col gap-8 px-8 py-8">
        <section className="flex items-center justify-between">
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span>
              Signed in as <span className="text-foreground">{user?.login}</span>
            </span>
            <span className="text-border">|</span>
            <span>Role: {org.role}</span>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => router.push(`/${slug}/members`)}>
              Members
            </Button>
            <Button variant="outline" size="sm" onClick={() => router.push("/")}>
              Switch org
            </Button>
            <Button variant="outline" size="sm" onClick={() => void signOut().then(() => router.replace("/login"))}>
              Sign out
            </Button>
          </div>
        </section>

        <section className="flex flex-col gap-3">
          <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
            Simulation scenarios
          </h2>
          <ScenarioPicker
            scenarios={scenarios}
            selected={selected}
            onSelect={setSelected}
            onRun={handleRun}
            running={running || !canWrite}
          />
          {!canWrite && (
            <p className="text-xs text-muted-foreground">
              Your role ({org.role}) is read-only. Ask an admin to promote you to operator to run experiments.
            </p>
          )}
        </section>

        {error && (
          <div className="rounded-lg border border-critical/40 bg-critical-bg/30 px-4 py-3 text-sm text-critical">
            {error}
          </div>
        )}

        {experiment && (
          <>
            <section className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
                  Experiment verdict
                </h2>
                <span className="text-xs tabular-nums text-muted-foreground">
                  {experiment.id}
                </span>
              </div>

              <ExperimentVerdict
                status={experiment.status}
                decision={experiment.decision || null}
                holdReason={experiment.hold_reason || null}
                regressionCount={regressionCount}
                metricCount={results.length}
              />

              {experiment.status === "decided" && (
                <div className="flex items-center gap-3">
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span>Canary: {experiment.canary_devices.join(", ")}</span>
                    <span className="text-border">|</span>
                    <span>Control: {experiment.control_devices.join(", ")}</span>
                  </div>
                  <div className="ml-auto flex items-center gap-2">
                    {canWrite && experiment.decision === "auto_hold" && (
                      <Button variant="outline" size="sm" onClick={handlePromote}>
                        Override &amp; promote
                      </Button>
                    )}
                    {canWrite && experiment.decision === "promote" && (
                      <Button variant="outline" size="sm" onClick={handleHold}>
                        Hold
                      </Button>
                    )}
                  </div>
                </div>
              )}
            </section>

            {results.length > 0 && (
              <section className="flex flex-col gap-3">
                <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
                  Metric comparison
                </h2>
                <ExperimentMetricsTable metrics={toMetricRows(results)} />
              </section>
            )}
          </>
        )}
      </main>
    </div>
  );
}
