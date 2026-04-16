"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import * as api from "@/lib/api";

type AuthStatus = "loading" | "authenticated" | "anonymous";

interface AuthState {
  status: AuthStatus;
  user: api.Me | null;
  refresh: () => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [user, setUser] = useState<api.Me | null>(null);

  const refresh = useCallback(async () => {
    try {
      const me = await api.getMe();
      setUser(me);
      setStatus("authenticated");
    } catch (err) {
      if (err instanceof api.ApiError && err.status === 401) {
        setUser(null);
        setStatus("anonymous");
        return;
      }
      // For any other failure fall back to anonymous so the UI routes
      // the user through login rather than hanging on a spinner.
      setUser(null);
      setStatus("anonymous");
    }
  }, []);

  const signOut = useCallback(async () => {
    try {
      await api.logout();
    } finally {
      setUser(null);
      setStatus("anonymous");
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <AuthContext.Provider value={{ status, user, refresh, signOut }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return ctx;
}
