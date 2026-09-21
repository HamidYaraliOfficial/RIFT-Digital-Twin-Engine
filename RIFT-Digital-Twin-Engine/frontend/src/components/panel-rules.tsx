"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Entity, Rule, Sensor } from "@/lib/types";

const SENSOR_TYPES = ["temperature", "humidity", "pressure", "vibration", "energy", "voltage", "current", "cpu", "memory", "network"];
const OPERATORS = ["gt", "lt", "gte", "lte", "eq", "neq"];

export function RulesPanel({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [rules, setRules] = useState<Rule[]>([]);
  const [entities, setEntities] = useState<Entity[]>([]);
  const [form, setForm] = useState({ name: "", entityId: "", trigger: "temperature", operator: "gt", value: 80, cooldownMs: 30000 });

  const load = () => api.get<Rule[]>(`/api/twins/${twinId}/rules`).then(setRules);
  useEffect(() => {
    load();
    api.get<Entity[]>(`/api/twins/${twinId}/entities`).then(setEntities);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [twinId]);

  const create = async () => {
    if (!form.name || !form.entityId) return;
    await api.post(`/api/twins/${twinId}/rules`, {
      name: form.name,
      entityId: form.entityId,
      trigger: form.trigger,
      conditions: [{ field: form.trigger, operator: form.operator, value: Number(form.value) }],
      actions: [{ type: "create_alert", params: { message: `${form.name}: ${form.trigger} ${form.operator} ${form.value}` } }],
      cooldownMs: Number(form.cooldownMs),
      enabled: true,
    });
    setForm({ ...form, name: "" });
    load();
  };

  const remove = async (id: string) => {
    await api.del(`/api/rules/${id}`);
    load();
  };

  return (
    <div className="space-y-4">
      <div className="card p-5">
        <h3 className="mb-4 text-lg font-semibold">{t.rules.newRule}</h3>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3 lg:grid-cols-6">
          <input className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm sm:col-span-2" placeholder={t.common.name} value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm sm:col-span-2" value={form.entityId} onChange={(e) => setForm({ ...form, entityId: e.target.value })}>
            <option value="">{t.entity.parent}</option>
            {entities.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={form.trigger} onChange={(e) => setForm({ ...form, trigger: e.target.value })}>
            {SENSOR_TYPES.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
          <select className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={form.operator} onChange={(e) => setForm({ ...form, operator: e.target.value })}>
            {OPERATORS.map((o) => <option key={o} value={o}>{o}</option>)}
          </select>
          <input type="number" className="rounded-md border border-[var(--border)] bg-[var(--surface)] px-2 py-1.5 text-sm" value={form.value} onChange={(e) => setForm({ ...form, value: Number(e.target.value) })} />
          <button className="btn btn-accent sm:col-span-2" onClick={create}>{t.common.create}</button>
        </div>
      </div>

      <div className="card divide-y divide-[var(--border)]">
        {rules.length === 0 && <p className="p-5 text-[var(--ink-muted)]">{t.rules.empty}</p>}
        {rules.map((r) => (
          <div key={r.id} className="flex items-center justify-between gap-3 p-4">
            <div>
              <p className="font-medium">{r.name}</p>
              <p className="text-xs text-[var(--ink-muted)]">
                {r.trigger} {r.conditions[0]?.operator} {r.conditions[0]?.value} → {r.actions[0]?.type}
              </p>
            </div>
            <button className="btn" onClick={() => remove(r.id)}>{t.common.delete}</button>
          </div>
        ))}
      </div>
    </div>
  );
}
