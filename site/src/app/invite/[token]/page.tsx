"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function InviteAcceptPage() {
  const router = useRouter();
  const params = useParams<{ token: string }>();
  const token = params.token;
  const { status: authStatus, refresh } = useAuth();

  const [info, setInfo] = useState<api.InviteLookup | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [accepting, setAccepting] = useState(false);

  useEffect(() => {
    if (authStatus === "anonymous") {
      router.replace(`/login`);
      return;
    }
    if (authStatus !== "authenticated") return;

    let alive = true;
    api
      .getInvite(token)
      .then((i) => {
        if (alive) setInfo(i);
      })
      .catch((err) => {
        if (!alive) return;
        setError(err instanceof Error ? err.message : "Invite not found");
      });
    return () => {
      alive = false;
    };
  }, [authStatus, router, token]);

  const handleAccept = useCallback(async () => {
    setAccepting(true);
    setError(null);
    try {
      const org = await api.acceptInvite(token);
      // Refresh auth state in case the server issued any updates, then
      // drop the user into the org they just joined.
      await refresh();
      router.replace(`/${org.slug}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to accept invite");
      setAccepting(false);
    }
  }, [token, refresh, router]);

  if (authStatus === "loading" || authStatus === "anonymous") {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Loading…
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-8">
      <div className="flex w-full max-w-[480px] flex-col gap-6 rounded-xl border border-border bg-card p-8">
        <div className="flex flex-col gap-2">
          <h1 className="text-xl font-semibold tracking-tight">You've been invited</h1>
          {info && (
            <p className="text-sm text-muted-foreground">
              Join <span className="text-foreground">{info.organization_name}</span> as{" "}
              <span className="text-foreground">{info.role}</span>.
            </p>
          )}
        </div>
        {error && (
          <div className="rounded-md border border-critical/40 bg-critical-bg/30 px-3 py-2 text-sm text-critical">
            {error}
          </div>
        )}
        {info && (
          <Button size="lg" onClick={handleAccept} disabled={accepting}>
            {accepting ? "Joining…" : "Accept invite"}
          </Button>
        )}
      </div>
    </div>
  );
}
