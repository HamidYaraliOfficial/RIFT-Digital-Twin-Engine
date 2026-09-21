"use client";

import React, { createContext, useContext, useEffect, useMemo, useState } from "react";
import { dictionaries, localeDirection, Locale, Dictionary } from "@/lib/i18n";

export type Theme = "windows" | "light" | "dark" | "red" | "blue";

interface ThemeCtx { theme: Theme; setTheme: (t: Theme) => void; }
interface LocaleCtx { locale: Locale; setLocale: (l: Locale) => void; t: Dictionary; dir: "ltr" | "rtl"; }

const ThemeContext = createContext<ThemeCtx | null>(null);
const LocaleContext = createContext<LocaleCtx | null>(null);

export function useTheme(): ThemeCtx {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within RiftProviders");
  return ctx;
}

export function useLocale(): LocaleCtx {
  const ctx = useContext(LocaleContext);
  if (!ctx) throw new Error("useLocale must be used within RiftProviders");
  return ctx;
}

export function RiftProviders({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("windows");
  const [locale, setLocaleState] = useState<Locale>("en");
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const storedTheme = window.localStorage.getItem("rift_theme") as Theme | null;
    const storedLocale = window.localStorage.getItem("rift_locale") as Locale | null;
    if (storedTheme) setThemeState(storedTheme);
    if (storedLocale) setLocaleState(storedLocale);
    setReady(true);
  }, []);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
  }, [theme]);

  useEffect(() => {
    document.documentElement.setAttribute("lang", locale);
    document.documentElement.setAttribute("dir", localeDirection[locale]);
  }, [locale]);

  const setTheme = (t: Theme) => {
    setThemeState(t);
    window.localStorage.setItem("rift_theme", t);
  };
  const setLocale = (l: Locale) => {
    setLocaleState(l);
    window.localStorage.setItem("rift_locale", l);
  };

  const themeValue = useMemo(() => ({ theme, setTheme }), [theme]);
  const localeValue = useMemo(
    () => ({ locale, setLocale, t: dictionaries[locale], dir: localeDirection[locale] }),
    [locale]
  );

  if (!ready) return null;

  return (
    <ThemeContext.Provider value={themeValue}>
      <LocaleContext.Provider value={localeValue}>{children}</LocaleContext.Provider>
    </ThemeContext.Provider>
  );
}
