"use client";
import { useState } from "react";
import Link from "next/link";
import { useLocale, useTheme, Theme } from "@/components/providers";
import { useAuth } from "@/components/auth-provider";
import { LoginDialog } from "@/components/login-dialog";
import type { Locale } from "@/lib/i18n";

const THEME_SWATCH: Record<Theme, string> = {
  windows: "#0067c0",
  light: "#2f6fed",
  dark: "#5eb6ff",
  red: "#c4292e",
  blue: "#0b5fc4",
};

export function Topbar({ breadcrumb }: { breadcrumb?: string }) {
  const { t, locale, setLocale } = useLocale();
  const { theme, setTheme } = useTheme();
  const { session, logout } = useAuth();
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [loginOpen, setLoginOpen] = useState(false);

  return (
    <header className="sticky top-0 z-40 acrylic border-b border-[var(--border)]">
      <div className="mx-auto flex max-w-[1400px] items-center gap-4 px-5 py-3">
        <Link href="/" className="flex items-center gap-2 font-semibold tracking-tight">
          <span
            className="grid h-8 w-8 place-items-center rounded-lg text-sm font-bold text-white"
            style={{ background: "var(--accent)" }}
          >
            R
          </span>
          <span>RIFT</span>
        </Link>
        {breadcrumb && (
          <>
            <span className="text-[var(--ink-muted)]">/</span>
            <span className="text-[var(--ink-muted)]">{breadcrumb}</span>
          </>
        )}
        <div className="flex-1" />

        <div className="relative">
          <button className="btn" onClick={() => setSettingsOpen((v) => !v)}>
            ⚙️ {t.nav.settings}
          </button>
          {settingsOpen && (
            <div className="card acrylic shadow-acrylic absolute end-0 mt-2 w-72 p-4 space-y-4">
              <div>
                <p className="mb-2 text-xs font-semibold uppercase text-[var(--ink-muted)]">{t.settings.theme}</p>
                <div className="flex gap-2">
                  {(Object.keys(THEME_SWATCH) as Theme[]).map((th) => (
                    <button
                      key={th}
                      title={t.settings.themes[th]}
                      onClick={() => setTheme(th)}
                      className="h-8 w-8 rounded-full border-2"
                      style={{
                        background: THEME_SWATCH[th],
                        borderColor: theme === th ? "var(--ink)" : "transparent",
                      }}
                    />
                  ))}
                </div>
              </div>
              <div>
                <p className="mb-2 text-xs font-semibold uppercase text-[var(--ink-muted)]">{t.settings.language}</p>
                <div className="flex flex-col gap-1">
                  {(["en", "fa", "zh"] as Locale[]).map((l) => (
                    <button
                      key={l}
                      onClick={() => setLocale(l)}
                      className={`rounded-md px-2 py-1.5 text-start text-sm ${
                        locale === l ? "bg-[var(--surface-3)] font-semibold" : "hover:bg-[var(--surface-3)]"
                      }`}
                    >
                      {t.settings.languages[l]}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        {session ? (
          <button className="btn" onClick={logout}>
            {session.displayName} · {t.nav.logout}
          </button>
        ) : (
          <button className="btn btn-accent" onClick={() => setLoginOpen(true)}>
            {t.nav.login}
          </button>
        )}
      </div>
      {loginOpen && <LoginDialog onClose={() => setLoginOpen(false)} />}
    </header>
  );
}
