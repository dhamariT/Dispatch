"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const ROLES = ["admin", "operator", "viewer"] as const;

export default function MembersPage() {
  const router = useRouter();
  const params = useParams<{ orgSlug: string }>();
  const slug = params.orgSlug;
  const { status: authStatus, user } = useAuth();

  const [org, setOrg] = useState<api.Org | null>(null);
  const [members, setMembers] = useState<api.Member[]>([]);
  const [invites, setInvites] = useState<api.Invite[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [inviteRole, setInviteRole] = useState<(typeof ROLES)[number]>("operator");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviting, setInviting] = useState(false);
  const [newInvite, setNewInvite] = useState<api.CreatedInvite | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    try {
      const [o, ms] = await Promise.all([api.getOrg(slug), api.listMembers(slug)]);
      setOrg(o);
      setMembers(ms);
      if (o.role === "admin") {
        const iv = await api.listInvites(slug);
        setInvites(iv.filter((i) => !i.accepted_at));
      }
    } catch (err) {
      if (err instanceof api.ApiError && err.status === 404) {
        router.replace("/");
        return;
      }
      setLoadError(err instanceof Error ? err.message : "Failed to load members");
    }
  }, [router, slug]);

  useEffect(() => {
    if (authStatus === "anonymous") {
      router.replace(`/login`);
      return;
    }
    if (authStatus === "authenticated") {
      void reload();
    }
  }, [authStatus, reload, router]);

  const handleCreateInvite = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setInviting(true);
      setInviteError(null);
      setNewInvite(null);
      try {
        const created = await api.createInvite(slug, inviteRole, inviteEmail.trim());
        setNewInvite(created);
        setInviteEmail("");
        void reload();
      } catch (err) {
        setInviteError(err instanceof Error ? err.message : "Failed to create invite");
      } finally {
        setInviting(false);
      }
    },
    [slug, inviteRole, inviteEmail, reload],
  );

  const handleRemove = useCallback(
    async (userID: string) => {
      try {
        await api.removeMember(slug, userID);
        void reload();
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : "Failed to remove member");
      }
    },
    [slug, reload],
  );

  const handleRevoke = useCallback(
    async (id: string) => {
      try {
        await api.deleteInvite(slug, id);
        void reload();
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : "Failed to revoke invite");
      }
    },
    [slug, reload],
  );

  if (authStatus === "loading" || !org) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }

  const isAdmin = org.role === "admin";
  const shareBase = typeof window !== "undefined" ? window.location.origin : "";

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="flex items-center justify-between border-b border-border px-8 py-4">
        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={() => router.push(`/${slug}`)}>
            ← Back
          </Button>
          <div className="flex flex-col">
            <span className="text-sm font-semibold">{org.name}</span>
            <span className="text-xs text-muted-foreground">Members</span>
          </div>
        </div>
        <span className="text-xs text-muted-foreground">Role: {org.role}</span>
      </header>

      <main className="mx-auto flex w-full max-w-[840px] flex-1 flex-col gap-8 px-8 py-8">
        {loadError && (
          <div className="rounded-lg border border-critical/40 bg-critical-bg/30 px-4 py-3 text-sm text-critical">
            {loadError}
          </div>
        )}

        <section className="flex flex-col gap-3">
          <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
            Members ({members.length})
          </h2>
          <div className="flex flex-col divide-y divide-border rounded-lg border border-border bg-card">
            {members.map((m) => (
              <div key={m.user_id} className="flex items-center justify-between px-4 py-3">
                <div className="flex items-center gap-3">
                  {m.avatar_url && (
                    /* eslint-disable-next-line @next/next/no-img-element */
                    <img src={m.avatar_url} alt="" className="h-8 w-8 rounded-full" />
                  )}
                  <div className="flex flex-col">
                    <span className="text-sm font-medium">{m.name || m.login}</span>
                    <span className="text-xs text-muted-foreground">{m.email}</span>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs uppercase tracking-wider text-muted-foreground">{m.role}</span>
                  {isAdmin && m.user_id !== user?.id && (
                    <Button variant="outline" size="sm" onClick={() => handleRemove(m.user_id)}>
                      Remove
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>

        {isAdmin && (
          <>
            <section className="flex flex-col gap-4 rounded-lg border border-border bg-card p-6">
              <div className="flex flex-col gap-1">
                <h2 className="text-sm font-semibold">Invite a teammate</h2>
                <p className="text-xs text-muted-foreground">
                  Invites are single-use links that expire in 7 days. The token is shown once; copy it now.
                </p>
              </div>
              <form className="flex flex-col gap-3" onSubmit={handleCreateInvite}>
                <div className="flex gap-3">
                  <label className="flex flex-1 flex-col gap-1 text-xs text-muted-foreground">
                    Email (optional)
                    <input
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      type="email"
                      className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
                      placeholder="teammate@acme.com"
                    />
                  </label>
                  <label className="flex flex-col gap-1 text-xs text-muted-foreground">
                    Role
                    <select
                      value={inviteRole}
                      onChange={(e) => setInviteRole(e.target.value as (typeof ROLES)[number])}
                      className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
                    >
                      {ROLES.map((r) => (
                        <option key={r} value={r}>
                          {r}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
                {inviteError && <p className="text-xs text-critical">{inviteError}</p>}
                <Button type="submit" size="sm" disabled={inviting} className="self-start">
                  {inviting ? "Creating…" : "Create invite"}
                </Button>
              </form>
              {newInvite && (
                <div className="flex flex-col gap-2 rounded-md border border-primary/40 bg-primary/5 p-3">
                  <p className="text-xs font-medium">Share this link with your teammate:</p>
                  <code className="break-all rounded bg-background px-2 py-1 text-xs">
                    {shareBase}/invite/{newInvite.token}
                  </code>
                  <p className="text-xs text-muted-foreground">
                    This token will not be shown again. It expires {new Date(newInvite.expires_at).toLocaleString()}.
                  </p>
                </div>
              )}
            </section>

            <section className="flex flex-col gap-3">
              <h2 className="text-xs font-medium uppercase tracking-widest text-muted-foreground">
                Pending invites ({invites.length})
              </h2>
              {invites.length === 0 ? (
                <p className="text-xs text-muted-foreground">No pending invites.</p>
              ) : (
                <div className="flex flex-col divide-y divide-border rounded-lg border border-border bg-card">
                  {invites.map((i) => (
                    <div key={i.id} className="flex items-center justify-between px-4 py-3">
                      <div className="flex flex-col">
                        <span className="text-sm">{i.email || "(no email)"}</span>
                        <span className="text-xs text-muted-foreground">
                          {i.role} · expires {new Date(i.expires_at).toLocaleString()}
                        </span>
                      </div>
                      <Button variant="outline" size="sm" onClick={() => handleRevoke(i.id)}>
                        Revoke
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </>
        )}
      </main>
    </div>
  );
}
