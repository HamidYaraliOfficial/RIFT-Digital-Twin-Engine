// Thin fetch wrapper for the RIFT REST API. Kept dependency-free (no axios)
// so the Control Center has zero required backend-specific tooling beyond
// the browser fetch API.
"use client";

const BASE_URL = process.env.NEXT_PUBLIC_RIFT_API_URL || "http://localhost:8080";

let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
  if (typeof window !== "undefined") {
    if (token) window.localStorage.setItem("rift_token", token);
    else window.localStorage.removeItem("rift_token");
  }
}

export function loadStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  const t = window.localStorage.getItem("rift_token");
  authToken = t;
  return t;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (authToken) headers["Authorization"] = `Bearer ${authToken}`;
  const res = await fetch(`${BASE_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    cache: "no-store",
  });
  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;
  if (!res.ok) {
    const message = (data && (data.error as string)) || res.statusText;
    throw new ApiError(res.status, message);
  }
  return data as T;
}

export const api = {
  base: BASE_URL,
  get: <T,>(path: string) => request<T>("GET", path),
  post: <T,>(path: string, body?: unknown) => request<T>("POST", path, body),
  patch: <T,>(path: string, body?: unknown) => request<T>("PATCH", path, body),
  del: <T,>(path: string) => request<T>("DELETE", path),
};

export function streamUrl(twinId: string): string {
  return `${BASE_URL}/api/twins/${twinId}/stream`;
}
