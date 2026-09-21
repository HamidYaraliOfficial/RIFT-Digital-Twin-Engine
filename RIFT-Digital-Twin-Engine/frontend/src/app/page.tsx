"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import { Topbar } from "@/components/topbar";
import { TwinCard } from "@/components/twin-card";
import { CreateTwinDialog } from "@/components/create-twin-dialog";
import type { Twin, PipelineStats } from "@/lib/types";

export default function DashboardPage() {
  const { t } = useLocale();
  const [twins, setTwins] = useState<Twin[]>([]);
  const [stats, setStats] = useState<PipelineStats | null>(null);
  const [createOpen, setCreateOpen] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [twinList, pipelineStats] = await Promise.all([
          api.get<Twin[]>("/api/twins"),
          api.get<PipelineStats>("/api/stats"),
        ]);
        if (!cancelled) {
          setTwins(twinList.filter((tw) => !tw.archived));
          setStats(pipelineStats);
        }
      } catch {
        // riftd not reachable yet — the empty state below covers this
      }
    }
    load();
    const interval = setInterval(load, 5000);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, []);

  return (
    <div className="min-h-screen">
      <Topbar />
      <main className="mx-auto max-w-[1400px] px-5 py-8">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{t.dashboard.title}</h1>
            <p className="text-[var(--ink-muted)]">{t.dashboard.subtitle}</p>
          </div>
          <button className="btn btn-accent" onClick={() => setCreateOpen(true)}>
            + {t.dashboard.createTwin}
          </button>
        </div>

        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {twins.map((twin) => (
            <TwinCard key={twin.id} twin={twin} />
          ))}
          {twins.length === 0 && (
            <div className="card col-span-full p-10 text-center text-[var(--ink-muted)]">
              {t.common.loading}
            </div>
          )}
        </div>

        {stats && (
          <div className="card mt-10 p-5">
            <h2 className="mb-3 text-sm font-semibold uppercase text-[var(--ink-muted)]">{t.dashboard.pipeline}</h2>
            <div className="grid grid-cols-3 gap-4 text-center sm:w-96">
              <PipelineStat label={t.dashboard.received} value={stats.received} />
              <PipelineStat label={t.dashboard.processed} value={stats.processed} />
              <PipelineStat label={t.dashboard.dropped} value={stats.dropped} />
            </div>
          </div>
        )}
      </main>
      {createOpen && (
        <CreateTwinDialog onClose={() => setCreateOpen(false)} onCreated={(tw) => setTwins((prev) => [...prev, tw])} />
      )}
    </div>
  );
}

function PipelineStat({ label, value }: { label: string; value: number }) {
  return (
    <div>
      <div className="text-xl font-bold">{value.toLocaleString()}</div>
      <div className="text-xs text-[var(--ink-muted)]">{label}</div>
    </div>
  );
}
