"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Twin, Entity, Alert } from "@/lib/types";

interface Summary { entities: number; sensors: number; health: number; alerts: number; }

export function TwinCard({ twin }: { twin: Twin }) {
  const { t } = useLocale();
  const [summary, setSummary] = useState<Summary | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [entities, sensors, healthRes, alerts] = await Promise.all([
          api.get<Entity[]>(`/api/twins/${twin.id}/entities`),
          api.get<unknown[]>(`/api/twins/${twin.id}/sensors`),
          api.get<{ overallHealth: number }>(`/api/twins/${twin.id}/health`),
          api.get<Alert[]>(`/api/twins/${twin.id}/alerts`),
        ]);
        if (!cancelled) {
          setSummary({
            entities: entities.length,
            sensors: sensors.length,
            health: Math.round(healthRes.overallHealth),
            alerts: alerts.length,
          });
        }
      } catch {
        if (!cancelled) setSummary({ entities: 0, sensors: 0, health: 100, alerts: 0 });
      }
    }
    load();
    const interval = setInterval(load, 8000);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [twin.id]);

  const templateLabel =
    (t.dashboard.templates as Record<string, string>)[twin.template] || twin.template || t.dashboard.templates.custom;

  return (
    <Link href={`/twins/${twin.id}`} className="card block p-5 transition hover:shadow-acrylic hover:-translate-y-0.5">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="font-semibold">{twin.name}</h3>
          <p className="text-xs text-[var(--ink-muted)]">{templateLabel}</p>
        </div>
        <span
          className="grid h-9 w-9 place-items-center rounded-lg text-xs font-bold text-white"
          style={{ background: healthColor(summary?.health ?? 100) }}
        >
          {summary ? `${summary.health}` : "…"}
        </span>
      </div>
      <div className="mt-4 grid grid-cols-3 gap-2 text-center text-sm">
        <Stat label={t.dashboard.entities} value={summary?.entities} />
        <Stat label={t.dashboard.sensors} value={summary?.sensors} />
        <Stat label={t.dashboard.alerts} value={summary?.alerts} accent={Boolean(summary && summary.alerts > 0)} />
      </div>
    </Link>
  );
}

function Stat({ label, value, accent }: { label: string; value?: number; accent?: boolean }) {
  return (
    <div className="rounded-lg bg-[var(--surface-3)] py-2">
      <div className={`text-base font-semibold ${accent ? "text-[var(--danger)]" : ""}`}>{value ?? "…"}</div>
      <div className="text-[10px] text-[var(--ink-muted)]">{label}</div>
    </div>
  );
}

function healthColor(h: number): string {
  if (h >= 85) return "var(--success)";
  if (h >= 60) return "var(--warning)";
  return "var(--danger)";
}
