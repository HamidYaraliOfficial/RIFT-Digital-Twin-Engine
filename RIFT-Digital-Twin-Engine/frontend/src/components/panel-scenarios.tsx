"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Entity, FaultSpec, MonteCarloResult, Scenario } from "@/lib/types";

const FAULT_TYPES = ["sensor_failure", "network_failure", "power_failure", "equipment_failure", "communication_delay", "packet_loss", "resource_exhaustion"];

export function ScenariosPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [name, setName] = useState("");
  const [ticks, setTicks] = useState(40);
  const [faults, setFaults] = useState<FaultSpec[]>([]);
  const [mcResult, setMcResult] = useState<MonteCarloResult | null>(null);
  const [busy, setBusy] = useState(false);

  const load = () => api.get<Scenario[]>(`/api/twins/${twinId}/scenarios`).then(setScenarios);
  useEffect(() => {
    load();
    api.get<Entity[]>(`/api/twins/${twinId}/entities`).then(setEntities);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  const addFault = () => {
    if (entities.length === 0) return;
    setFaults((f) => [...f, { type: "equipment_failure", entityId: entities[0].id, atTick: 0, magnitude: 1 }]);
  };
  const updateFault = (i: number, patch: Partial<FaultSpec>) => {
    setFaults((f) => f.map((x, idx) => (idx === i ? { ...x, ...patch } : x)));
  };

  const createAndRun = async () => {
    if (!name || faults.length === 0) return;
    setBusy(true);
    try {
      const scenario = await api.post<Scenario>(`/api/twins/${twinId}/scenarios`, { name, durationTicks: ticks, faults });
      await api.post(`/api/scenarios/${scenario.id}/run`);
      setName("");
      setFaults([]);
      load();
    } finally {
      setBusy(false);
    }
  };

  const runMonteCarlo = async () => {
    if (faults.length === 0) return;
    setBusy(true);
    try {
      const res = await api.post<MonteCarloResult>(`/api/twins/${twinId}/montecarlo`, {
        faults, durationTicks: ticks, runs: 60, magMin: 0.2, magMax: 1,
      });
      setMcResult(res);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="card p-5">
        <h3 className="mb-4 text-lg font-semibold">{t.scenarios.newScenario}</h3>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm sm:col-span-2" placeholder={t.scenarios.scenarioName} value={name} onChange={(e) => setName(e.target.value)} />
          <input type="number" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.scenarios.durationTicks} value={ticks} onChange={(e) => setTicks(Number(e.target.value))} />
        </div>

        <div className="mt-4 space-y-2">
          {faults.map((f, i) => (
            <div key={i} className="grid grid-cols-2 gap-2 sm:grid-cols-4">
              <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={f.type} onChange={(e) => updateFault(i, { type: e.target.value })}>
                {FAULT_TYPES.map((ft) => <option key={ft} value={ft}>{ft}</option>)}
              </select>
              <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={f.entityId} onChange={(e) => updateFault(i, { entityId: e.target.value })}>
                {entities.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
              </select>
              <input type="number" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.scenarios.atTick} value={f.atTick} onChange={(e) => updateFault(i, { atTick: Number(e.target.value) })} />
              <input type="number" step="0.1" min="0" max="1" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.scenarios.magnitude} value={f.magnitude} onChange={(e) => updateFault(i, { magnitude: Number(e.target.value) })} />
            </div>
          ))}
          <button className="btn" onClick={addFault}>+ {t.scenarios.addFault}</button>
        </div>

        <div className="mt-4 flex flex-wrap gap-2">
          <button className="btn btn-accent" disabled={busy} onClick={createAndRun}>{t.scenarios.runScenario}</button>
          <button className="btn" disabled={busy} onClick={runMonteCarlo}>{t.scenarios.montecarlo}</button>
        </div>

        {mcResult && (
          <div className="mt-4 grid grid-cols-3 gap-2 text-center text-sm sm:grid-cols-6">
            {(["mean", "p10", "p50", "p90", "best", "worst"] as const).map((k) => (
              <div key={k} className="rounded-lg bg-[var(--surface-3)] py-2">
                <div className="font-semibold">{mcResult[k]}</div>
                <div className="text-[10px] uppercase text-[var(--ink-muted)]">{k}</div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="card divide-y divide-[var(--border)]">
        {scenarios.length === 0 && <p className="p-5 text-[var(--ink-muted)]">{t.scenarios.empty}</p>}
        {scenarios.map((s) => (
          <div key={s.id} className="p-4">
            <div className="flex items-center justify-between">
              <p className="font-medium">{s.name}</p>
              <span className="badge">{s.status}</span>
            </div>
            {s.result && (
              <div className="mt-2 text-sm text-[var(--ink-muted)]">
                {s.result.summary} — {t.scenarios.affected}: {s.result.affectedEntities.length}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
