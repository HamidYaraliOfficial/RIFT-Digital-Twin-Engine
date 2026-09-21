"use client";
import { useLocale } from "@/components/providers";
import type { EntityStatus } from "@/lib/types";

const DOT: Record<string, string> = {
  operational: "bg-[var(--success)]",
  recovered: "bg-[var(--success)]",
  warning: "bg-[var(--warning)]",
  degraded: "bg-[var(--warning)]",
  maintenance: "bg-[var(--ink-muted)]",
  failed: "bg-[var(--danger)]",
  unknown: "bg-[var(--ink-muted)]",
};

export function StatusBadge({ status }: { status: EntityStatus | string }) {
  const { t } = useLocale();
  const label = (t.status as Record<string, string>)[status] ?? status;
  return (
    <span className={`badge status-${status}`}>
      <span className={`h-1.5 w-1.5 rounded-full ${DOT[status] ?? "bg-[var(--ink-muted)]"}`} />
      {label}
    </span>
  );
}

export function SeverityBadge({ severity }: { severity: string }) {
  const { t } = useLocale();
  const label = (t.severity as Record<string, string>)[severity] ?? severity;
  const color =
    severity === "critical" ? "text-[var(--danger)] border-[var(--danger)]"
    : severity === "warning" ? "text-[var(--warning)] border-[var(--warning)]"
    : "text-[var(--ink-muted)]";
  return <span className={`badge ${color}`}>{label}</span>;
}
