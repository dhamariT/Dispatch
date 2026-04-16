"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const router = useRouter();
  const { status } = useAuth();

  useEffect(() => {
    if (status === "authenticated") {
      router.replace("/");
    }
  }, [status, router]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-8">
      <div className="flex w-full max-w-[420px] flex-col gap-6 rounded-xl border border-border bg-card p-8">
        <div className="flex flex-col gap-2">
          <h1 className="text-xl font-semibold tracking-tight">Dispatch</h1>
          <p className="text-sm text-muted-foreground">
            Sign in to validate fleet deploys with concurrent canary-vs-control experiments.
          </p>
        </div>
        <Button
          size="lg"
          onClick={() => {
            // Full navigation so the browser keeps the GitHub redirect
            // chain; client-side router push won't follow 302s.
            window.location.href = api.githubLoginUrl();
          }}
        >
          Sign in with GitHub
        </Button>
        <p className="text-xs text-muted-foreground">
          Dispatch only reads your public profile and primary verified email from GitHub. No repo or org access is
          requested.
        </p>
      </div>
    </div>
  );
}
