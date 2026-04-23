<script lang="ts">
  import { resolve } from '$app/paths';
  import Chart from '$lib/components/Chart.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import { eraForYear, type Era } from '$lib/eras';
  import type { LeaderColumn, LeaderRow } from '$lib/leaders/types';
  import type { ChartConfiguration } from 'chart.js';
  import { SvelteMap } from 'svelte/reactivity';

  type Props = {
    title: string;
    subtitle?: string;
    rows: LeaderRow[];
    columns: LeaderColumn[];
    loading: boolean;
    error: string | null;
    page: number;
    perPage: number;
    total: number;
    onPageChange: (page: number) => void;
    onPerPageChange: (perPage: number) => void;
    trendEmptyMessage?: string;
  };

  let {
    title,
    subtitle,
    rows,
    columns,
    loading,
    error,
    page,
    perPage,
    total,
    onPageChange,
    onPerPageChange,
    trendEmptyMessage = 'No era trend data available for this query.'
  }: Props = $props();

  const rankedRows = $derived.by(() =>
    rows
      .filter((row) => row.metric != null)
      .toSorted((left, right) => Number(right.metric ?? -Infinity) - Number(left.metric ?? -Infinity))
      .slice(0, 12)
  );

  const barChartConfig = $derived.by(
    (): ChartConfiguration<'bar'> => ({
      type: 'bar',
      data: {
        labels: rankedRows.map((row) => row.label),
        datasets: [
          {
            label: title,
            data: rankedRows.map((row) => row.metric ?? 0),
            backgroundColor: 'rgba(59, 130, 246, 0.75)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 1
          }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280' } } },
        scales: {
          x: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } },
          y: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } }
        }
      }
    })
  );

  const buckets = new SvelteMap<string, { era: Era; values: number[] }>();

  let eraTrend = $derived.by(() => {
    for (const row of rows) {
      if (row.year == null) continue;
      if (row.metric == null) continue;

      const era = eraForYear(row.year);
      if (!era) continue;

      const key = era.code;
      const existing = buckets.get(key);
      if (existing) {
        existing.values.push(row.metric);
        continue;
      }

      buckets.set(key, { era, values: [row.metric] });
    }

    return [...buckets.values()]
      .map((bucket) => {
        const totalValue = bucket.values.reduce((sum, value) => sum + value, 0);
        return {
          era: bucket.era,
          average: bucket.values.length > 0 ? totalValue / bucket.values.length : 0,
          samples: bucket.values.length
        };
      })
      .toSorted((left, right) => left.era.from - right.era.from);
  });

  let eraTrendConfig = $derived.by(
    (): ChartConfiguration<'line'> => ({
      type: 'line',
      data: {
        labels: eraTrend.map((entry) => `${entry.era.code} (${entry.era.from}-${entry.era.to})`),
        datasets: [
          {
            label: 'Era bucket average',
            data: eraTrend.map((entry) => Number(entry.average.toFixed(3))),
            borderColor: 'rgba(16, 185, 129, 1)',
            backgroundColor: 'rgba(16, 185, 129, 0.2)',
            tension: 0.35,
            fill: true
          }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280' } } },
        scales: {
          x: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } },
          y: { ticks: { color: '#6b7280' }, grid: { color: '#252934' } }
        }
      }
    })
  );

  const caveatEras = $derived(eraTrend.map((entry) => entry.era).filter((era) => Boolean(era.caveat)));
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-1">{title}</div>
  {#if subtitle}
    <p class="mb-3 font-mono text-[0.72rem] text-muted">{subtitle}</p>
  {/if}

  {#if error}
    <p class="mb-3 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {error}
    </p>
  {/if}

  {#if loading}
    <p class="font-mono text-[0.72rem] text-muted">Loading rows…</p>
  {:else if rows.length === 0}
    <p class="font-mono text-[0.72rem] text-muted">No rows returned for this query.</p>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full border-collapse">
        <thead>
          <tr>
            {#each columns as column (column.label)}
              <th
                class="border-b border-outline px-2 py-1.5 font-mono text-xxs tracking-[0.08em] text-muted uppercase {column.align ===
                'right'
                  ? 'text-right'
                  : 'text-left'}">
                {column.label}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each rows as row, index (`${row.id}:${index}`)}
            <tr class="border-b border-outline last:border-0 hover:bg-surface/60">
              {#each columns as column (column.label)}
                <td
                  class="px-2 py-1.5 font-mono text-[0.72rem] text-foreground {column.align === 'right'
                    ? 'text-right'
                    : 'text-left'}">
                  {#if column.key === 'rank'}
                    {((page - 1) * perPage + index + 1).toLocaleString()}
                  {:else if column.key === 'label' && row.href}
                    <a href={resolve(row.href as '/')} class="text-primary no-underline hover:underline">{row.label}</a>
                  {:else if column.key === 'team' && row.team}
                    {#if row.kind === 'team'}
                      <a
                        href={resolve(`/teams/${encodeURIComponent(row.team)}` as `/teams/${string}`)}
                        class="text-primary no-underline hover:underline">
                        {row.team}
                      </a>
                    {:else}
                      {row.team}
                    {/if}
                  {:else}
                    {String(row[column.key as keyof LeaderRow] ?? '—')}
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <div class="mt-3">
      <Pagination {page} {perPage} {total} {onPageChange} {onPerPageChange} />
    </div>
  {/if}
</div>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label">Ranked bar chart</div>
  {#if rankedRows.length === 0}
    <p class="font-mono text-[0.72rem] text-muted">No numeric values available for a chart.</p>
  {:else}
    <Chart config={barChartConfig} height={170} />
  {/if}
</div>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label">Era-bucket trend</div>
  {#if eraTrend.length === 0}
    <p class="font-mono text-[0.72rem] text-muted">{trendEmptyMessage}</p>
  {:else}
    <Chart config={eraTrendConfig} height={170} />
    <div class="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
      {#each eraTrend as entry (entry.era.code)}
        <div class="rounded border border-outline bg-surface px-3 py-2 font-mono text-[0.68rem]">
          <div class="text-foreground">{entry.era.label}</div>
          <div class="text-muted">avg: {entry.average.toFixed(3)} · samples: {entry.samples}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if caveatEras.length > 0}
  <EraDisclaimer eras={caveatEras} />
{/if}
