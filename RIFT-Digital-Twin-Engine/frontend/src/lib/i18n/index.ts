import en, { Dictionary } from "./en";
import fa from "./fa";
import zh from "./zh";

export type Locale = "en" | "fa" | "zh";

export const locales: Locale[] = ["en", "fa", "zh"];

export const dictionaries: Record<Locale, Dictionary> = { en, fa, zh };

export const localeDirection: Record<Locale, "ltr" | "rtl"> = {
  en: "ltr",
  fa: "rtl",
  zh: "ltr",
};

export type { Dictionary };
