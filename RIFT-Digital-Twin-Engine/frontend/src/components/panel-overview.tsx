"use client";
import { useEffect, useState } from "react";
import { api, streamUrl } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import { SeverityBadge } from "@/components/status-badge";
import type { Alert, RiftEvent, ClockState } from "@/lib/types";

export function ClockWidget({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [clock, setClock] = useState<ClockState | null>(null);

  const load = () => api.get<ClockState>(`/api/twins/${twinId}/clock`).then(setClock).catch(() => {});
  useEffect(() => {
    load();
    const iv = setInterval(load, 3000);
    return () => clearInterval(iv);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  const act = async (action: string, body?: Record<string, unknown>) => {
    const updated = await api.post<ClockState>(`/api/twins/${twinId}/clock/${action}`, body);
    setClock(updated);
  };

  if (!clock) return null;
  return (
    <div className="card flex flex-wrap items-center justify-between gap-3 p-4">
      <div>
        <p className="text-xs uppercase text-[var(--ink-muted)]">{t.workspace.clock}</p>
        <p className="font-mono text-sm">{new Date(clock.simTime).toLocaleString()} · ×{clock.speed}</p>
      </div>
      <div className="flex gap-2">
        <button className="btn" onClick={() => act(clock.running ? "pause" : "resume")}>
          {clock.running ? t.workspace.pause : t.workspace.resume}
        </button>
        <button className="btn" onClick={() => act("step", { seconds: 60 })}>{t.workspace.step}</button>
        {[1, 4, 16].map((sp) => (
          <button key={sp} className={`btn ${clock.speed === sp ? "btn-accent" : ""}`} onClick={() => act("speed", { speed: sp })}>
            ×{sp}
          </button>
        ))}
      </div>
    </div>
  );
}

export function OverviewPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [events, setEvents] = useState<RiftEvent[]>([]);
  const [health, setHealth] = useState<number | null>(null);

  const loadAlerts = () => api.get<Alert[]>(`/api/twins/${twinId}/alerts`).then(setAlerts).catch(() => {});
  const loadHealth = () => api.get<{ overallHealth: number }>(`/api/twins/${twinId}/health`).then((r) => setHealth(Math.round(r.overallHealth))).catch(() => {});

  useEffect(() => {
    loadAlerts();
    loadHealth();
    api.get<RiftEvent[]>(`/api/twins/${twinId}/events?limit=30`).then(setEvents).catch(() => {});
    const iv = setInterval(() => {
      loadAlerts();
      loadHealth();
    }, 6000);
    return () => clearInterval(iv);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  useEffect(() => {
    const es = new EventSource(streamUrl(twinId));
    es.addEventListener("event", (ev) => {
      try {
        const evt: RiftEvent = JSON.parse((ev as MessageEvent).data);
        setEvents((prev) => [evt, ...prev].slice(0, 40));
      } catch {
        // ignore malformed frame
      }
    });
    return () => es.close();
  }, [twinId]);

  const ack = async (id: string) => {
    await api.post(`/api/alerts/${id}/ack`);
    loadAlerts();
  };
  const resolve = async (id: string) => {
    await api.post(`/api/alerts/${id}/resolve`);
    loadAlerts();
  };

  return (
    <div className="space-y-4">
      <ClockWidget twinId={twinId} />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="card p-5">
          <h3 className="mb-3 text-sm font-semibold uppercase text-[var(--ink-muted)]">{t.workspace.healthOverview}</h3>
          <div className="grid h-32 place-items-center">
            <div className="text-5xl font-bold" style={{ color: health && health < 60 ? "var(--danger)" : health && health < 85 ? "var(--warning)" : "var(--success)" }}>
              {health ?? "…"}
            </div>
          </div>
        </div>
        <div className="card p-5 lg:col-span-2">
          <h3 className="mb-3 text-sm font-semibold uppercase text-[var(--ink-muted)]">{t.workspace.recentAlerts}</h3>
          <div className="max-h-56 space-y-2 overflow-y-auto">
            {alerts.length === 0 && <p className="text-[var(--ink-muted)]">{t.alerts.empty}</p>}
            {alerts.map((a) => (
              <div key={a.id} className="flex items-center justify-between gap-2 rounded-lg bg-[var(--surface-3)] px-3 py-2 text-sm">
                <div className="min-w-0">
                  <SeverityBadge severity={a.severity} /> <span className="truncate">{a.message}</span>
                  <span className="ms-2 text-xs text-[var(--ink-muted)]">×{a.occurrenceCount}</span>
                </div>
                {a.status === "open" && (
                  <div className="flex shrink-0 gap-1">
                    <button className="btn" onClick={() => ack(a.id)}>{t.alerts.acknowledge}</button>
                    <button className="btn" onClick={() => resolve(a.id)}>{t.alerts.resolve}</button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
      <div className="card p-5">
        <h3 className="mb-3 text-sm font-semibold uppercase text-[var(--ink-muted)]">{t.workspace.recentEvents}</h3>
        <div className="max-h-56 space-y-1 overflow-y-auto text-sm">
          {events.map((e) => (
            <div key={e.id} className="flex justify-between border-b border-[var(--border)] py-1">
              <span>{e.type}</span>
              <span className="text-[var(--ink-muted)]">{new Date(e.timestamp).toLocaleTimeString()}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
