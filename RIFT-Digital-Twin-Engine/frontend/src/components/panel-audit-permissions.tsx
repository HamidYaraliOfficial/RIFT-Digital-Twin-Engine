"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { AuditEntry } from "@/lib/types";

export function AuditPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  useEffect(() => {
    api.get<AuditEntry[]>(`/api/twins/${twinId}/audit`).then(setEntries);
  }, [twinId]);
  return (
    <div className="card overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--border)] text-start text-xs uppercase text-[var(--ink-muted)]">
            <th className="px-4 py-3 text-start">{t.audit.when}</th>
            <th className="px-4 py-3 text-start">{t.audit.actor}</th>
            <th className="px-4 py-3 text-start">{t.audit.action}</th>
            <th className="px-4 py-3 text-start">{t.audit.target}</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e) => (
            <tr key={e.id} className="border-b border-[var(--border)]">
              <td className="px-4 py-2 text-[var(--ink-muted)]">{new Date(e.timestamp).toLocaleString()}</td>
              <td className="px-4 py-2">{e.actor}</td>
              <td className="px-4 py-2">{e.action}</td>
              <td className="px-4 py-2 font-mono text-xs">{e.target}</td>
            </tr>
          ))}
          {entries.length === 0 && (
            <tr><td colSpan={4} className="px-4 py-6 text-center text-[var(--ink-muted)]">{t.common.none}</td></tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

export function PermissionsPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [userId, setUserId] = useState("");
  const [role, setRole] = useState("");
  const [entityId, setEntityId] = useState("");
  const [scopes, setScopes] = useState<string[]>(["read"]);
  const [status, setStatus] = useState<string | null>(null);

  const toggle = (scope: string) => {
    setScopes((s) => (s.includes(scope) ? s.filter((x) => x !== scope) : [...s, scope]));
  };

  const grant = async () => {
    try {
      await api.post("/api/permissions", { twinId, userId, role, entityId, scopes });
      setStatus("✅ " + t.permissions.grant);
    } catch (e) {
      setStatus("⛔ " + (e instanceof Error ? e.message : "failed"));
    }
  };

  const SCOPE_OPTIONS: [string, string][] = [
    ["read", t.permissions.scopeRead],
    ["write", t.permissions.scopeWrite],
    ["scenario", t.permissions.scopeScenario],
    ["actuator_control", t.permissions.scopeActuator],
    ["admin", t.permissions.scopeAdmin],
  ];

  return (
    <div className="card max-w-xl p-5">
      <h3 className="mb-4 text-lg font-semibold">{t.permissions.title}</h3>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.permissions.user} value={userId} onChange={(e) => setUserId(e.target.value)} />
        <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.permissions.role} value={role} onChange={(e) => setRole(e.target.value)} />
        <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm sm:col-span-2" placeholder={t.permissions.entityScope} value={entityId} onChange={(e) => setEntityId(e.target.value)} />
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        {SCOPE_OPTIONS.map(([key, label]) => (
          <button key={key} onClick={() => toggle(key)} className={`badge ${scopes.includes(key) ? "border-[var(--accent)] text-[var(--accent)]" : ""}`}>
            {label}
          </button>
        ))}
      </div>
      <button className="btn btn-accent mt-4" onClick={grant}>{t.permissions.grant}</button>
      {status && <p className="mt-2 text-sm">{status}</p>}
    </div>
  );
}
