"use client";
import { useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Twin } from "@/lib/types";

const TEMPLATES = ["", "factory", "building", "data_center", "warehouse", "smart_city"];

export function CreateTwinDialog({ onClose, onCreated }: { onClose: () => void; onCreated: (t: Twin) => void }) {
  const { t } = useLocale();
  const [name, setName] = useState("");
  const [template, setTemplate] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setBusy(true);
    setErr(null);
    try {
      const twin = await api.post<Twin>("/api/twins", { name, template });
      if (template) {
        await api.post(`/api/twins/${twin.id}/synthetic/${template}`, {});
      }
      onCreated(twin);
      onClose();
    } catch (e2) {
      setErr(e2 instanceof Error ? e2.message : "failed");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <form onClick={(e) => e.stopPropagation()} onSubmit={submit} className="card acrylic shadow-acrylic w-[min(92vw,440px)] p-6 space-y-4">
        <h2 className="text-lg font-semibold">{t.dashboard.createDialogTitle}</h2>
        <div className="space-y-2">
          <label className="block text-sm text-[var(--ink-muted)]">{t.dashboard.twinName}</label>
          <input autoFocus className="w-full rounded-lg border px-3 py-2 bg-[var(--surface)] border-[var(--border)]" value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div className="space-y-2">
          <label className="block text-sm text-[var(--ink-muted)]">{t.dashboard.startFromTemplate}</label>
          <div className="grid grid-cols-3 gap-2">
            {TEMPLATES.filter(Boolean).map((tpl) => (
              <button
                type="button"
                key={tpl}
                onClick={() => setTemplate(tpl === template ? "" : tpl)}
                className={`rounded-lg border px-2 py-2 text-xs ${template === tpl ? "border-[var(--accent)] bg-[var(--surface-3)] font-semibold" : "border-[var(--border)]"}`}
              >
                {(t.dashboard.templates as Record<string, string>)[tpl]}
              </button>
            ))}
          </div>
        </div>
        {err && <p className="text-sm text-[var(--danger)]">{err}</p>}
        <div className="flex justify-end gap-2 pt-2">
          <button type="button" className="btn" onClick={onClose}>{t.common.cancel}</button>
          <button type="submit" disabled={busy} className="btn btn-accent">{t.common.create}</button>
        </div>
      </form>
    </div>
  );
}
