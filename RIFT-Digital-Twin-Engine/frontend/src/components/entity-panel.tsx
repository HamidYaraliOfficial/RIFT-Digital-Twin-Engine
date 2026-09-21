"use client";
import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import { StatusBadge } from "@/components/status-badge";
import type { Entity, Sensor, ImpactResult, RootCause } from "@/lib/types";

interface TreeNode extends Entity { childNodes: TreeNode[]; }

function buildTree(entities: Entity[]): TreeNode[] {
  const byId = new Map<string, TreeNode>();
  entities.forEach((e) => byId.set(e.id, { ...e, childNodes: [] }));
  const roots: TreeNode[] = [];
  byId.forEach((node) => {
    if (node.parentId && byId.has(node.parentId)) {
      byId.get(node.parentId)!.childNodes.push(node);
    } else {
      roots.push(node);
    }
  });
  return roots;
}

function TreeItem({ node, depth, selected, onSelect }: { node: TreeNode; depth: number; selected: string | null; onSelect: (id: string) => void }) {
  const [open, setOpen] = useState(depth < 1);
  return (
    <div>
      <div
        className={`flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm hover:bg-[var(--surface-3)] ${selected === node.id ? "bg-[var(--surface-3)] font-semibold" : ""}`}
        style={{ paddingInlineStart: depth * 14 + 6 }}
        onClick={() => onSelect(node.id)}
      >
        {node.childNodes.length > 0 ? (
          <span onClick={(e) => { e.stopPropagation(); setOpen((v) => !v); }} className="w-3 text-[var(--ink-muted)]">
            {open ? "▾" : "▸"}
          </span>
        ) : (
          <span className="w-3" />
        )}
        <span className="truncate">{node.name}</span>
      </div>
      {open && node.childNodes.map((c) => (
        <TreeItem key={c.id} node={c} depth={depth + 1} selected={selected} onSelect={onSelect} />
      ))}
    </div>
  );
}

export function EntityPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [impact, setImpact] = useState<ImpactResult[] | null>(null);
  const [rootcause, setRootcause] = useState<RootCause[] | null>(null);
  const [analysisBusy, setAnalysisBusy] = useState(false);

  const load = async () => {
    const list = await api.get<Entity[]>(`/api/twins/${twinId}/entities`);
    setEntities(list);
  };

  useEffect(() => {
    load();
    const iv = setInterval(load, 6000);
    return () => clearInterval(iv);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  useEffect(() => {
    if (!selectedId) return;
    setImpact(null);
    setRootcause(null);
    api.get<Sensor[]>(`/api/twins/${twinId}/sensors`).then((all) => setSensors(all.filter((s) => s.entityId === selectedId)));
  }, [selectedId, twinId]);

  const tree = useMemo(() => buildTree(entities), [entities]);
  const selected = entities.find((e) => e.id === selectedId) || null;

  const runImpact = async () => {
    if (!selectedId) return;
    setAnalysisBusy(true);
    try {
      const res = await api.get<{ impact: ImpactResult[] }>(`/api/entities/${selectedId}/impact`);
      setImpact(res.impact || []);
    } finally {
      setAnalysisBusy(false);
    }
  };

  const runRootCause = async () => {
    if (!selectedId) return;
    setAnalysisBusy(true);
    try {
      const res = await api.get<{ candidates: RootCause[] }>(`/api/entities/${selectedId}/rootcause?windowMinutes=60`);
      setRootcause(res.candidates || []);
    } finally {
      setAnalysisBusy(false);
    }
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px_1fr]">
      <div className="card max-h-[70vh] overflow-y-auto p-3">
        <h3 className="mb-2 px-1 text-sm font-semibold text-[var(--ink-muted)]">{t.entity.tree}</h3>
        {tree.map((n) => (
          <TreeItem key={n.id} node={n} depth={0} selected={selectedId} onSelect={setSelectedId} />
        ))}
      </div>

      <div className="card p-5">
        {!selected ? (
          <p className="text-[var(--ink-muted)]">{t.entity.noSelection}</p>
        ) : (
          <div className="space-y-5">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">{selected.name}</h3>
                <p className="text-xs text-[var(--ink-muted)]">{selected.type} · {selected.id}</p>
              </div>
              <StatusBadge status={selected.status} />
            </div>

            <div className="grid grid-cols-3 gap-3 text-center text-sm">
              <MiniStat label="Health" value={Math.round(selected.health?.health ?? 100)} />
              <MiniStat label="Risk" value={Math.round(selected.health?.risk ?? 0)} />
              <MiniStat label="Criticality" value={Math.round(selected.health?.criticality ?? 0)} />
            </div>

            {selected.properties && Object.keys(selected.properties).length > 0 && (
              <div>
                <h4 className="mb-1 text-xs font-semibold uppercase text-[var(--ink-muted)]">{t.entity.properties}</h4>
                <div className="grid grid-cols-2 gap-2 text-sm sm:grid-cols-3">
                  {Object.entries(selected.properties).map(([k, v]) => (
                    <div key={k} className="rounded-md bg-[var(--surface-3)] px-2 py-1">
                      <span className="text-[var(--ink-muted)]">{k}: </span>
                      <span className="font-medium">{String(v)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {sensors.length > 0 && (
              <div>
                <h4 className="mb-1 text-xs font-semibold uppercase text-[var(--ink-muted)]">{t.entity.sensors}</h4>
                <div className="flex flex-wrap gap-2">
                  {sensors.map((s) => (
                    <span key={s.id} className="badge">
                      {s.type} · {s.status}
                    </span>
                  ))}
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <button className="btn w-full" disabled={analysisBusy} onClick={runImpact}>{t.entity.runImpact}</button>
                {impact && (
                  <ul className="mt-2 max-h-40 overflow-y-auto text-sm">
                    {impact.length === 0 && <li className="text-[var(--ink-muted)]">{t.common.none}</li>}
                    {impact.map((r) => (
                      <li key={r.entityId} className="border-b border-[var(--border)] py-1">
                        {r.entityId} <span className="text-[var(--ink-muted)]">(depth {r.depth})</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
              <div>
                <button className="btn w-full" disabled={analysisBusy} onClick={runRootCause}>{t.entity.runRootCause}</button>
                {rootcause && (
                  <ul className="mt-2 max-h-40 overflow-y-auto text-sm">
                    {rootcause.length === 0 && <li className="text-[var(--ink-muted)]">{t.common.none}</li>}
                    {rootcause.map((r, i) => (
                      <li key={i} className="border-b border-[var(--border)] py-1">
                        {r.reason} <span className="text-[var(--ink-muted)]">({Math.round(r.confidence * 100)}% {t.entity.confidence})</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>

            <ActuatorForm entityId={selected.id} />
          </div>
        )}
      </div>
    </div>
  );
}

function MiniStat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-[var(--surface-3)] py-2">
      <div className="text-base font-semibold">{value}</div>
      <div className="text-[10px] text-[var(--ink-muted)]">{label}</div>
    </div>
  );
}

function ActuatorForm({ entityId }: { entityId: string }) {
  const { t } = useLocale();
  const [type, setType] = useState("start");
  const [value, setValue] = useState<number>(0);
  const [reason, setReason] = useState("");
  const [result, setResult] = useState<string | null>(null);

  const submit = async () => {
    try {
      const res = await api.post<{ accepted: boolean; reason: string }>(`/api/entities/${entityId}/command`, { type, value, reason });
      setResult(res.accepted ? "✅ " + res.reason : "⛔ " + res.reason);
    } catch (e) {
      setResult("⛔ " + (e instanceof Error ? e.message : "failed"));
    }
  };

  return (
    <div className="rounded-lg border border-[var(--border)] p-4">
      <h4 className="mb-2 text-xs font-semibold uppercase text-[var(--ink-muted)]">{t.entity.sendCommand}</h4>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-4">
        <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={type} onChange={(e) => setType(e.target.value)}>
          {["start", "stop", "restart", "open", "close", "lock", "unlock", "set_temperature", "change_speed"].map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
        <input type="number" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" placeholder={t.entity.commandValue} value={value} onChange={(e) => setValue(Number(e.target.value))} />
        <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm sm:col-span-1" placeholder={t.entity.commandReason} value={reason} onChange={(e) => setReason(e.target.value)} />
        <button className="btn btn-accent" onClick={submit}>{t.entity.commandSubmit}</button>
      </div>
      {result && <p className="mt-2 text-sm">{result}</p>}
    </div>
  );
}
