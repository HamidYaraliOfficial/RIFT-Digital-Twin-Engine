"use client";
import { useState } from "react";
import { useLocale } from "@/components/providers";
import { useAuth } from "@/components/auth-provider";

export function LoginDialog({ onClose }: { onClose: () => void }) {
  const { t } = useLocale();
  const { login, error } = useAuth();
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    try {
      await login(username, password);
      onClose();
    } catch {
      // error surfaced via context
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <form
        onClick={(e) => e.stopPropagation()}
        onSubmit={submit}
        className="card acrylic shadow-acrylic w-[min(92vw,380px)] p-6 space-y-4"
      >
        <h2 className="text-lg font-semibold">{t.auth.title}</h2>
        <div className="space-y-2">
          <label className="block text-sm text-[var(--ink-muted)]">{t.auth.username}</label>
          <input
            className="w-full rounded-lg border px-3 py-2 bg-[var(--surface)] border-[var(--border)]"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label className="block text-sm text-[var(--ink-muted)]">{t.auth.password}</label>
          <input
            type="password"
            className="w-full rounded-lg border px-3 py-2 bg-[var(--surface)] border-[var(--border)]"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        {error && <p className="text-sm text-[var(--danger)]">{t.auth.error}</p>}
        <p className="text-xs text-[var(--ink-muted)]">{t.auth.defaultHint}</p>
        <div className="flex justify-end gap-2 pt-2">
          <button type="button" className="btn" onClick={onClose}>{t.common.cancel}</button>
          <button type="submit" disabled={busy} className="btn btn-accent">{t.auth.signIn}</button>
        </div>
      </form>
    </div>
  );
}
