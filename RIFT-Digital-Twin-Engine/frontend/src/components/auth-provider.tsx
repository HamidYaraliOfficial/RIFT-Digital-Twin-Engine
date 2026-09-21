"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { api, setAuthToken, loadStoredToken } from "@/lib/api-client";
import type { LoginResponse } from "@/lib/types";

interface Session { userId: string; displayName: string; role: string; token: string; }
interface AuthCtx {
  session: Session | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  error: string | null;
}

const AuthContext = createContext<AuthCtx | null>(null);

export function useAuth(): AuthCtx {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = loadStoredToken();
    const raw = window.localStorage.getItem("rift_session");
    if (token && raw) {
      try {
        const parsed = JSON.parse(raw) as Session;
        setSession(parsed);
      } catch {
        // ignore corrupt local session
      }
    }
  }, []);

  const login = async (username: string, password: string) => {
    setError(null);
    try {
      const res = await api.post<LoginResponse>("/api/auth/login", { username, password });
      const s: Session = { userId: res.userId, displayName: res.displayName, role: res.role, token: res.token };
      setAuthToken(res.token);
      window.localStorage.setItem("rift_session", JSON.stringify(s));
      setSession(s);
    } catch (e) {
      setError(e instanceof Error ? e.message : "login failed");
      throw e;
    }
  };

  const logout = () => {
    setAuthToken(null);
    window.localStorage.removeItem("rift_session");
    setSession(null);
  };

  return <AuthContext.Provider value={{ session, login, logout, error }}>{children}</AuthContext.Provider>;
}
