"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import { Topbar } from "@/components/topbar";
import { OverviewPanel } from "@/components/panel-overview";
import { EntityPanel } from "@/components/entity-panel";
import { Twin3DViewer } from "@/components/twin-3d-viewer";
import { Twin2DMap } from "@/components/twin-2d-map";
import { TelemetryPanel } from "@/components/telemetry-panel";
import { RulesPanel } from "@/components/panel-rules";
import { ScenariosPanel } from "@/components/panel-scenarios";
import { HoursPanel } from "@/components/panel-hours";
import { AuditPanel, PermissionsPanel } from "@/components/panel-audit-permissions";
import type { Twin } from "@/lib/types";

type TabKey = "overview" | "entities" | "view3d" | "map2d" | "telemetry" | "rules" | "scenarios" | "hours" | "audit" | "permissions";

export default function TwinWorkspacePage() {
  const params = useParams<{ id: string }>();
  const twinId = params.id;
  const { t } = useLocale();
  const [twin, setTwin] = useState<Twin | null>(null);
  const [tab, setTab] = useState<TabKey>("overview");

  useEffect(() => {
    api.get<Twin>(`/api/twins/${twinId}`).then(setTwin).catch(() => {});
  }, [twinId]);

  const tabs: { key: TabKey; label: string }[] = [
    { key: "overview", label: t.workspace.tabs.overview },
    { key: "entities", label: t.workspace.tabs.entities },
    { key: "view3d", label: t.workspace.tabs.view3d },
    { key: "map2d", label: t.workspace.tabs.map2d },
    { key: "telemetry", label: t.workspace.tabs.telemetry },
    { key: "rules", label: t.workspace.tabs.rules },
    { key: "scenarios", label: t.workspace.tabs.scenarios },
    { key: "hours", label: t.workspace.tabs.hours },
    { key: "audit", label: t.workspace.tabs.audit },
    { key: "permissions", label: t.workspace.tabs.permissions },
  ];

  return (
    <div className="min-h-screen">
      <Topbar breadcrumb={twin?.name} />
      <main className="mx-auto max-w-[1400px] px-5 py-6">
        <div className="mb-5 flex flex-wrap gap-1 overflow-x-auto rounded-xl bg-[var(--surface-3)] p-1">
          {tabs.map((tb) => (
            <button
              key={tb.key}
              onClick={() => setTab(tb.key)}
              className={`whitespace-nowrap rounded-lg px-3 py-2 text-sm font-medium transition ${
                tab === tb.key ? "bg-[var(--surface-2)] shadow-sm" : "text-[var(--ink-muted)] hover:text-[var(--ink)]"
              }`}
            >
              {tb.label}
            </button>
          ))}
        </div>

        {tab === "overview" && <OverviewPanel twinId={twinId} />}
        {tab === "entities" && <EntityPanel twinId={twinId} />}
        {tab === "view3d" && <Twin3DViewer twinId={twinId} />}
        {tab === "map2d" && <Twin2DMap twinId={twinId} />}
        {tab === "telemetry" && <TelemetryPanel twinId={twinId} />}
        {tab === "rules" && <RulesPanel twinId={twinId} />}
        {tab === "scenarios" && <ScenariosPanel twinId={twinId} />}
        {tab === "hours" && <HoursPanel twinId={twinId} />}
        {tab === "audit" && <AuditPanel twinId={twinId} />}
        {tab === "permissions" && <PermissionsPanel twinId={twinId} />}
      </main>
    </div>
  );
}
