import { type ChartConfiguration } from 'chart.js';
import { BATTING_RATE_STATS, BATTING_STATS, CHART_SCALES } from '$lib/common/constants';
import type { BattingSeason, HofEntry, PitchingSeason } from '$lib/players/types';

export function createBattingChartConfig(rows: BattingSeason[], stat: string): ChartConfiguration {
  const selected = BATTING_STATS.find((candidate) => candidate.value === stat) ?? BATTING_STATS[0];
  const statKey = selected.value as keyof BattingSeason;
  const isRate = BATTING_RATE_STATS.has(selected.value);

  return {
    type: 'line',
    data: {
      labels: rows.map((row) => String(row.year)),
      datasets: [
        {
          label: selected.label,
          data: rows.map((row) => {
            const value = Number(row[statKey] ?? 0);
            return isRate ? Math.round(value * 1000) : value;
          }),
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59,130,246,0.1)',
          pointRadius: 3,
          tension: 0.3,
          fill: true
        }
      ]
    },
    options: { responsive: true, plugins: { legend: { display: false } }, scales: CHART_SCALES }
  };
}

export function createPitchingChartConfig(rows: PitchingSeason[]): ChartConfiguration {
  return {
    type: 'line',
    data: {
      labels: rows.map((row) => String(row.year)),
      datasets: [
        {
          label: 'ERA',
          data: rows.map((row) => Number(row.era ?? 0)),
          borderColor: '#10b981',
          backgroundColor: 'rgba(16,185,129,0.1)',
          pointRadius: 3,
          tension: 0.3,
          fill: true
        }
      ]
    },
    options: { responsive: true, plugins: { legend: { display: false } }, scales: CHART_SCALES }
  };
}

export function createHallOfFameChartConfig(entries: HofEntry[]): ChartConfiguration {
  return {
    type: 'bar',
    data: {
      labels: entries.map((entry) => String(entry.year_inducted ?? '?')),
      datasets: [
        {
          label: 'Vote %',
          data: entries.map((entry) => Number(entry.pct ?? 0)),
          backgroundColor: 'rgba(16,185,129,0.6)'
        }
      ]
    },
    options: {
      responsive: true,
      plugins: { legend: { labels: { color: '#6b7280', font: { size: 10 } } } },
      scales: {
        x: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } },
        y: {
          min: 0,
          max: 100,
          ticks: { color: '#6b7280', callback: (value) => value + '%' },
          grid: { color: '#252934' }
        }
      }
    }
  };
}
