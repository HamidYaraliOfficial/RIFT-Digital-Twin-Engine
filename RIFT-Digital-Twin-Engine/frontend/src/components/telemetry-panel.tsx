"use client";
import { useEffect, useMemo, useState } from "react";
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts";
import { api, streamUrl } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Entity, Sensor, Reading } from "@/lib/types";

export function TelemetryPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [selected, setSelected] = useState<string>("");
  const [points, setPoints] = useState<{ t: string; v: number }[]>([]);

  useEffect(() => {
    Promise.all([
      api.get<Entity[]>(`/api/twins/${twinId}/entities`),
      api.get<Sensor[]>(`/api/twins/${twinId}/sensors`),
    ]).then(([e, s]) => {
      setEntities(e);
      setSensors(s);
      if (s.length > 0) setSelected(s[0].id);
    });
  }, [twinId]);

  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    api.get<Reading[]>(`/api/sensors/${selected}/telemetry?window=80`).then((readings) => {
      if (!cancelled) setPoints(readings.map((r) => ({ t: new Date(r.timestamp).toLocaleTimeString(), v: r.value })));
    });
    return () => {
      cancelled = true;
    };
  }, [selected]);

  useEffect(() => {
    const es = new EventSource(streamUrl(twinId));
    es.addEventListener("telemetry", (ev) => {
      try {
        const reading: Reading = JSON.parse((ev as MessageEvent).data);
        if (reading.sensorId !== selected) return;
        setPoints((prev) => [...prev.slice(-119), { t: new Date(reading.timestamp).toLocaleTimeString(), v: reading.value }]);
      } catch {
        // ignore malformed frame
      }
    });
    return () => es.close();
  }, [twinId, selected]);

  const entityName = useMemo(() => {
    const s = sensors.find((x) => x.id === selected);
    return entities.find((e) => e.id === s?.entityId)?.name ?? "";
  }, [selected, sensors, entities]);

  const selectedSensor = sensors.find((s) => s.id === selected);

  return (
    <div className="card p-5">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <h3 className="text-lg font-semibold">{t.telemetry.title}</h3>
        <select
          className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-3 py-2 text-sm"
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
        >
          {sensors.map((s) => {
            const owner = entities.find((e) => e.id === s.entityId);
            return (
              <option key={s.id} value={s.id}>
                {owner?.name ?? s.entityId} — {s.type}
              </option>
            );
          })}
        </select>
      </div>
      {selectedSensor && (
        <p className="mb-2 text-sm text-[var(--ink-muted)]">
          {entityName} · {selectedSensor.type} ({selectedSensor.unit}) · <span className="text-[var(--success)]">● {t.telemetry.live}</span>
        </p>
      )}
      {points.length === 0 ? (
        <p className="py-10 text-center text-[var(--ink-muted)]">{t.telemetry.noData}</p>
      ) : (
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={points}>
              <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" />
              <XAxis dataKey="t" tick={{ fontSize: 10 }} minTickGap={30} stroke="var(--ink-muted)" />
              <YAxis tick={{ fontSize: 10 }} stroke="var(--ink-muted)" domain={["auto", "auto"]} />
              <Tooltip contentStyle={{ background: "var(--surface-2)", border: "1px solid var(--border)", borderRadius: 8 }} />
              <Line type="monotone" dataKey="v" stroke="var(--accent)" strokeWidth={2} dot={false} isAnimationActive={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
