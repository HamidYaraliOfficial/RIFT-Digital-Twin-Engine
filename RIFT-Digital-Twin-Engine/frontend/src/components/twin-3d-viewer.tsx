"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import { Canvas } from "@react-three/fiber";
import { OrbitControls, Grid, Text } from "@react-three/drei";
import * as THREE from "three";
import { api } from "@/lib/api-client";
import { useLocale } from "@/components/providers";
import type { Entity } from "@/lib/types";

const STATUS_COLOR: Record<string, string> = {
  operational: "#1f9d55",
  recovered: "#1f9d55",
  warning: "#d68a00",
  degraded: "#d68a00",
  maintenance: "#8a8f9b",
  failed: "#d13438",
  unknown: "#8a8f9b",
};

function EntityBox({ entity, selected, onSelect }: { entity: Entity; selected: boolean; onSelect: (id: string) => void }) {
  const size: [number, number, number] = [
    entity.dimensions.x || 1,
    entity.dimensions.y || 1,
    entity.dimensions.z || 1,
  ];
  return (
    <mesh
      position={[entity.position.x, entity.position.z, entity.position.y]}
      onClick={(e) => {
        e.stopPropagation();
        onSelect(entity.id);
      }}
    >
      <boxGeometry args={size} />
      <meshStandardMaterial
        color={STATUS_COLOR[entity.status] ?? "#8a8f9b"}
        emissive={selected ? "#ffffff" : "#000000"}
        emissiveIntensity={selected ? 0.25 : 0}
      />
    </mesh>
  );
}

// A working Three.js / React Three Fiber twin viewer: every entity in the
// twin is rendered as a status-colored box positioned by its own
// Position/Dimensions, with orbit camera controls and click-to-inspect. This
// is intentionally geometry-agnostic (no glTF asset required) so every twin
// — including ones with no imported 3D assets yet — is viewable in 3D
// immediately; the Asset System (see README) is where glTF/GLB models are
// attached to an Entity to replace its box with real geometry.
export function Twin3DViewer({ twinId }: { twinId: string }) {
  const { t } = useLocale();
  const [entities, setEntities] = useState<Entity[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    const load = () => api.get<Entity[]>(`/api/twins/${twinId}/entities`).then((e) => mounted.current && setEntities(e));
    load();
    const iv = setInterval(load, 6000);
    return () => {
      mounted.current = false;
      clearInterval(iv);
    };
  }, [twinId]);

  const selected = useMemo(() => entities.find((e) => e.id === selectedId) || null, [entities, selectedId]);

  return (
    <div className="card overflow-hidden">
      <div className="h-[560px] w-full">
        <Canvas camera={{ position: [24, 20, 24], fov: 45 }} onPointerMissed={() => setSelectedId(null)}>
          <color attach="background" args={["#00000000"]} />
          <ambientLight intensity={0.7} />
          <directionalLight position={[10, 20, 10]} intensity={1} />
          <Grid args={[100, 100]} cellColor="#8a8f9b55" sectionColor="#8a8f9b88" fadeDistance={80} />
          {entities.map((e) => (
            <EntityBox key={e.id} entity={e} selected={selectedId === e.id} onSelect={setSelectedId} />
          ))}
          {selected && (
            <Text
              position={[selected.position.x, (selected.position.z || 0) + (selected.dimensions.y || 1) + 1, selected.position.y]}
              fontSize={0.6}
              color="#ffffff"
              outlineWidth={0.02}
              outlineColor="#000000"
            >
              {selected.name}
            </Text>
          )}
          <OrbitControls makeDefault />
        </Canvas>
      </div>
      <div className="border-t border-[var(--border)] p-3 text-sm text-[var(--ink-muted)]">
        {selected ? `${selected.name} · ${selected.type} · ${selected.status}` : t.entity.noSelection}
      </div>
    </div>
  );
}

// keep THREE import referenced for consumers relying on module side effects
void THREE.Vector3;
