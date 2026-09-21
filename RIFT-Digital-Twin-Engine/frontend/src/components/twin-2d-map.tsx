"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Entity, Relationship } from "@/lib/types";

const STATUS_COLOR: Record<string, string> = {
  operational: "var(--success)",
  recovered: "var(--success)",
  warning: "var(--warning)",
  degraded: "var(--warning)",
  maintenance: "var(--ink-muted)",
  failed: "var(--danger)",
  unknown: "var(--ink-muted)",
};

// A dependency-free, top-down 2D operational map: entities are placed by
// their own Position.x/y (scaled), edges are drawn from Relationships. This
// covers the Building/Factory/Warehouse/City "2D Floor Plan / Map" use case
// without needing an external mapping library.
export function Twin2DMap({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [rels, setRels] = useState<Relationship[]>([]);
  const [hovered, setHovered] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.get<Entity[]>(`/api/twins/${twinId}/entities`),
      api.get<Relationship[]>(`/api/twins/${twinId}/relationships`),
    ]).then(([e, r]) => {
      setEntities(e);
      setRels(r);
    });
  }, [twinId]);

  if (entities.length === 0) {
    return <div className="card p-10 text-center text-[var(--ink-muted)]">{t.common.loading}</div>;
  }

  const xs = entities.map((e) => e.position.x);
  const ys = entities.map((e) => e.position.y);
  const minX = Math.min(...xs, 0), maxX = Math.max(...xs, 1);
  const minY = Math.min(...ys, 0), maxY = Math.max(...ys, 1);
  const W = 900, H = 560, pad = 40;
  const sx = (x: number) => pad + ((x - minX) / (maxX - minX || 1)) * (W - pad * 2);
  const sy = (y: number) => pad + ((y - minY) / (maxY - minY || 1)) * (H - pad * 2);
  const byId = new Map(entities.map((e) => [e.id, e]));

  return (
    <div className="card overflow-hidden p-3">
      <svg viewBox={`0 0 ${W} ${H}`} className="h-[560px] w-full">
        <rect x={0} y={0} width={W} height={H} fill="var(--surface-3)" opacity={0.35} />
        {rels.map((r) => {
          const a = byId.get(r.sourceId);
          const b = byId.get(r.targetId);
          if (!a || !b) return null;
          return (
            <line
              key={r.id}
              x1={sx(a.position.x)} y1={sy(a.position.y)}
              x2={sx(b.position.x)} y2={sy(b.position.y)}
              stroke="var(--border)" strokeWidth={1.5}
            />
          );
        })}
        {entities.map((e) => (
          <g key={e.id} onMouseEnter={() => setHovered(e.id)} onMouseLeave={() => setHovered(null)} className="cursor-pointer">
            <circle
              cx={sx(e.position.x)} cy={sy(e.position.y)} r={hovered === e.id ? 9 : 6.5}
              fill={STATUS_COLOR[e.status] ?? "var(--ink-muted)"}
              stroke="var(--surface-2)" strokeWidth={2}
            />
            {hovered === e.id && (
              <text x={sx(e.position.x) + 12} y={sy(e.position.y) + 4} fontSize={12} fill="var(--ink)">
                {e.name} · {e.type}
              </text>
            )}
          </g>
        ))}
      </svg>
    </div>
  );
}
