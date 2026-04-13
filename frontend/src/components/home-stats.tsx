import { useMemo } from "react";
import type { history } from "../../wailsjs/go/models";

interface HomeStatsProps {
  items: history.Record[];
}

export function HomeStats({ items }: HomeStatsProps) {
  const stats = useMemo(() => computeStats(items), [items]);
  return (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-2 rounded-lg border bg-card px-4 py-3">
      <Stat value={formatNumber(stats.words)} label="total words" />
      <Stat value={`${stats.sessions}`} label="sessions" />
      <Stat value={`${stats.days}`} label={stats.days === 1 ? "day" : "days"} />
    </div>
  );
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="flex items-baseline gap-1.5">
      <span className="text-base font-semibold tabular-nums">{value}</span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  );
}

function computeStats(items: history.Record[]) {
  let words = 0;
  const dayKeys = new Set<string>();
  let sessions = 0;
  for (const item of items) {
    if (item.error) continue;
    const text = item.refinedMessage || item.rawTranscript || "";
    const trimmed = text.trim();
    if (!trimmed) continue;
    sessions++;
    words += trimmed.split(/\s+/).length;
    const ts = new Date(item.timestamp as unknown as string);
    if (!Number.isNaN(ts.getTime())) {
      dayKeys.add(
        `${ts.getFullYear()}-${ts.getMonth()}-${ts.getDate()}`,
      );
    }
  }
  return { words, sessions, days: dayKeys.size };
}

function formatNumber(n: number): string {
  if (n >= 1000) {
    const k = n / 1000;
    return `${k >= 10 ? Math.round(k) : k.toFixed(1)}K`;
  }
  return `${n}`;
}
