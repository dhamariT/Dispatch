"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function RootPage() {
  const router = useRouter();
  const { status: authStatus, user, signOut } = useAuth();
  const [orgs, setOrgs] = useState<api.OrgSummary[] | null>(null);
  const [orgsError, setOrgsError] = useState<string | null>(null);

  const [slug, setSlug] = useState("");
  const [name, setName] = useState("");
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  useEffect(() => {
    if (authStatus === "anonymous") {
      router.replace("/login");
    }
  }, [authStatus, router]);

  useEffect(() => {
    if (authStatus !== "authenticated") return;
    let alive = true;
    api
      .listMyOrgs()
      .then((list) => {
        if (!alive) return;
        setOrgs(list);
        if (list.length === 1) {
          router.replace(`/${list[0].slug}`);
        }
      })
      .catch((err) => {
        if (!alive) return;
        setOrgsError(err instanceof Error ? err.message : "Failed to load orgs");
      });
    return () => {
      alive = false;
    };
  }, [authStatus, router]);

  const handleCreate = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setCreating(true);
      setCreateError(null);
      try {
        const org = await api.createOrg(slug.trim(), name.trim());
        router.replace(`/${org.slug}`);
      } catch (err) {
        setCreateError(err instanceof Error ? err.message : "Failed to create org");
      } finally {
        setCreating(false);
      }
    },
    [slug, name, router],
  );

  if (authStatus === "loading" || authStatus === "anonymous") {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="flex items-center justify-between border-b border-border px-8 py-4">
        <div className="flex items-center gap-2">
          <span className="text-lg font-semibold tracking-tight">Dispatch</span>
          <span className="text-xs text-muted-foreground">causal deploy validation</span>
        </div>
        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          <span>
            Signed in as <span className="text-foreground">{user?.login}</span>
          </span>
          <Button variant="outline" size="sm" onClick={() => void signOut().then(() => router.replace("/login"))}>
            Sign out
          </Button>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-[720px] flex-1 flex-col gap-10 px-8 py-12">
        <section className="flex flex-col gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">Your organizations</h1>
          <p className="text-sm text-muted-foreground">
            Pick an organization to open its deploy-validation dashboard, or create a new one for your team.
          </p>
        </section>

        {orgsError && (
          <div className="rounded-lg border border-critical/40 bg-critical-bg/30 px-4 py-3 text-sm text-critical">
            {orgsError}
          </div>
        )}

        {orgs && orgs.length > 0 && (
          <section className="flex flex-col gap-3">
            {orgs.map((o) => (
              <button
                key={o.id}
                type="button"
                onClick={() => router.push(`/${o.slug}`)}
                className="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 text-left transition-colors hover:border-primary/60"
              >
                <div className="flex flex-col">
                  <span className="text-sm font-medium">{o.name}</span>
                  <span className="text-xs text-muted-foreground">/{o.slug}</span>
                </div>
                <span className="text-xs uppercase tracking-wider text-muted-foreground">{o.role}</span>
              </button>
            ))}
          </section>
        )}

        <section className="flex flex-col gap-4 rounded-lg border border-border bg-card p-6">
          <div className="flex flex-col gap-1">
            <h2 className="text-sm font-semibold">Create a new organization</h2>
            <p className="text-xs text-muted-foreground">
              You become the first admin. You can invite teammates from the members page.
            </p>
          </div>
          <form className="flex flex-col gap-3" onSubmit={handleCreate}>
            <label className="flex flex-col gap-1 text-xs text-muted-foreground">
              Name
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
                placeholder="Acme Robotics"
              />
            </label>
            <label className="flex flex-col gap-1 text-xs text-muted-foreground">
              Slug
              <input
                value={slug}
                onChange={(e) => setSlug(e.target.value.toLowerCase())}
                required
                pattern="^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])?$"
                className="rounded-md border border-border bg-background px-3 py-2 font-mono text-sm text-foreground outline-none focus:border-primary"
                placeholder="acme-robotics"
              />
            </label>
            {createError && <p className="text-xs text-critical">{createError}</p>}
            <Button type="submit" disabled={creating} size="sm" className="self-start">
              {creating ? "Creating…" : "Create organization"}
            </Button>
          </form>
        </section>
      </main>
    </div>
  );
}
