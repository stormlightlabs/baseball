<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import Chart from '$lib/components/Chart.svelte';
  import { EP } from '$lib/endpoints';
  import {
    normalizeGameWinProbabilitySummary,
    normalizeLeverageList,
    normalizeWinProbabilityCurve
  } from '$lib/games/normalizers';
  import type { GameWinProbabilitySummary, PlateAppearanceLeverage, WinProbabilityCurve } from '$lib/games/types';
  import { AsyncListResource, AsyncValueResource } from '$lib/players/resources.svelte';
  import type { ChartConfiguration } from 'chart.js';
  import { onMount } from 'svelte';

  let gameId = $derived(page.params.id ?? '');

  const curveResource = new AsyncValueResource<WinProbabilityCurve>();
  const summaryResource = new AsyncValueResource<GameWinProbabilitySummary>();
  const leverageResource = new AsyncListResource<PlateAppearanceLeverage>();

  let lastCurveKey = '';
  let lastSummaryKey = '';
  let lastLeverageKey = '';

  let curvePoints = $derived.by(() =>
    [...(curveResource.value?.points ?? [])].toSorted((a, b) => (a.event_index ?? 0) - (b.event_index ?? 0))
  );

  let chartConfig = $derived.by(
    (): ChartConfiguration => ({
      type: 'line',
      data: {
        labels: curvePoints.map((point, index) => point.event_index ?? index + 1),
        datasets: [
          {
            label: 'Home Win %',
            data: curvePoints.map((point) => pct(point.home_win_prob)),
            borderColor: '#3b82f6',
            backgroundColor: 'rgba(59,130,246,0.18)',
            pointRadius: 0,
            tension: 0.2
          },
          {
            label: 'Away Win %',
            data: curvePoints.map((point) => {
              if (point.away_win_prob != null) return pct(point.away_win_prob);
              if (point.home_win_prob != null) return 100 - pct(point.home_win_prob);
              return null;
            }),
            borderColor: '#f59e0b',
            backgroundColor: 'rgba(245,158,11,0.14)',
            pointRadius: 0,
            tension: 0.2
          }
        ]
      },
      options: {
        responsive: true,
        plugins: {
          legend: { labels: { color: '#6b7280', font: { size: 10 } } },
          tooltip: {
            callbacks: {
              title(items) {
                const idx = items[0]?.dataIndex ?? -1;
                const point = idx >= 0 ? curvePoints[idx] : undefined;
                if (!point) return 'Event';
                return `${halfLabel(point.top_of_inning)} ${point.inning ?? '—'} · Event ${point.event_index ?? '—'}`;
              },
              afterLabel(item) {
                const idx = item.dataIndex;
                const point = idx >= 0 ? curvePoints[idx] : undefined;
                return point?.description ?? '';
              }
            }
          }
        },
        scales: {
          x: { ticks: { color: '#6b7280', font: { size: 9 }, maxTicksLimit: 10 }, grid: { color: '#252934' } },
          y: {
            min: 0,
            max: 100,
            ticks: { color: '#6b7280', font: { size: 9 }, callback: (v) => `${v}%` },
            grid: { color: '#252934' }
          }
        }
      }
    })
  );

  let summaryStart = $derived.by(() => summaryResource.value?.home_win_prob_start ?? curvePoints[0]?.home_win_prob);
  let summaryEnd = $derived.by(() => summaryResource.value?.home_win_prob_end ?? curvePoints.at(-1)?.home_win_prob);
  let summaryDelta = $derived.by(() =>
    summaryStart == null || summaryEnd == null ? null : pct(summaryEnd) - pct(summaryStart)
  );

  let topLeverageRows = $derived.by(() =>
    [...leverageResource.items].toSorted((a, b) => (b.li ?? -Infinity) - (a.li ?? -Infinity)).slice(0, 25)
  );

  onMount(() => {
    void refreshAll(true);
  });

  afterNavigate(() => {
    void refreshAll();
  });

  async function refreshAll(force = false): Promise<void> {
    await Promise.all([refreshCurve(force), refreshSummary(force), refreshLeverage(force)]);
  }

  async function refreshCurve(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      curveResource.clear();
      lastCurveKey = '';
      return;
    }

    if (!force && id === lastCurveKey) return;
    lastCurveKey = id;

    await curveResource.load(async () => {
      const payload = await apiFetch<unknown>(EP.gameWinProb(id));
      return normalizeWinProbabilityCurve(payload);
    });
  }

  async function refreshSummary(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      summaryResource.clear();
      lastSummaryKey = '';
      return;
    }

    if (!force && id === lastSummaryKey) return;
    lastSummaryKey = id;

    await summaryResource.load(async () => {
      const payload = await apiFetch<unknown>(EP.gameWinProbSummary(id));
      return normalizeGameWinProbabilitySummary(payload);
    });
  }

  async function refreshLeverage(force = false): Promise<void> {
    const id = gameId;
    if (!id) {
      leverageResource.clear();
      lastLeverageKey = '';
      return;
    }

    if (!force && id === lastLeverageKey) return;
    lastLeverageKey = id;

    await leverageResource.load(async () => {
      const payload = await apiFetch<unknown>(EP.gameLeverage(id), { min_li: 0 });
      return normalizeLeverageList(payload);
    });
  }

  function pct(value: number | undefined): number {
    if (value == null) return 0;
    return Math.max(0, Math.min(100, value * 100));
  }

  function fmtPct(value: number | undefined): string {
    if (value == null) return '—';
    return `${pct(value).toFixed(1)}%`;
  }

  function fmtSignedPoints(value: number | undefined): string {
    if (value == null) return '—';
    const points = value * 100;
    return `${points >= 0 ? '+' : ''}${points.toFixed(1)} pp`;
  }

  function fmtHalfInning(topOfInning: boolean | undefined, inning: number | undefined): string {
    return `${halfLabel(topOfInning)} ${inning ?? '—'}`;
  }

  function halfLabel(topOfInning: boolean | undefined): 'Top' | 'Bot' | '—' {
    if (topOfInning === true) return 'Top';
    if (topOfInning === false) return 'Bot';
    return '—';
  }
</script>

<div class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-3 flex flex-wrap items-center gap-2">
    <div class="panel-label mb-0 grow border-b-0 pb-0">Win Probability Curve</div>
    <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.64rem] text-muted">{gameId}</span>
  </div>

  {#if curveResource.loading}
    <p class="font-mono text-[0.78rem] text-muted">Loading win probability curve…</p>
  {:else if curveResource.error}
    <p class="font-mono text-[0.78rem] text-warning">{curveResource.error}</p>
  {:else if curvePoints.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No win probability points were returned for this game.</p>
  {:else}
    <Chart config={chartConfig} height={150} />
    <div class="mt-2 flex flex-wrap gap-1.5">
      <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-muted">
        Points: {curvePoints.length.toLocaleString()}
      </span>
      <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-muted">
        Home start: {fmtPct(summaryStart)}
      </span>
      <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-muted">
        Home end: {fmtPct(summaryEnd)}
      </span>
      <span class="rounded border border-outline px-2 py-0.5 font-mono text-[0.66rem] text-muted">
        Net swing: {summaryDelta == null ? '—' : `${summaryDelta >= 0 ? '+' : ''}${summaryDelta.toFixed(1)} pp`}
      </span>
    </div>
  {/if}
</div>

<div class="mt-4 rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Summary + Biggest Swings</div>

  {#if summaryResource.loading}
    <p class="font-mono text-[0.78rem] text-muted">Loading win-probability summary…</p>
  {:else if summaryResource.error}
    <p class="font-mono text-[0.78rem] text-warning">{summaryResource.error}</p>
  {:else if summaryResource.value}
    <div class="mb-3 grid grid-cols-2 gap-2 md:grid-cols-4">
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Home team</div>
        <div class="font-display text-[0.86rem] text-foreground">{summaryResource.value.home_team ?? '—'}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Away team</div>
        <div class="font-display text-[0.86rem] text-foreground">{summaryResource.value.away_team ?? '—'}</div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Home start WE</div>
        <div class="font-display text-[0.86rem] text-foreground">
          {fmtPct(summaryResource.value.home_win_prob_start)}
        </div>
      </div>
      <div class="rounded border border-outline bg-surface px-3 py-2">
        <div class="font-mono text-[0.6rem] text-muted uppercase">Home end WE</div>
        <div class="font-display text-[0.86rem] text-foreground">{fmtPct(summaryResource.value.home_win_prob_end)}</div>
      </div>
    </div>

    <div class="grid gap-3 md:grid-cols-2">
      <div class="rounded border border-outline bg-surface p-3">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Biggest positive swing</div>
        {#if summaryResource.value.biggest_positive_swing}
          {@const swing = summaryResource.value.biggest_positive_swing}
          <div class="font-mono text-[0.72rem] text-foreground">{fmtSignedPoints(swing.we_change)}</div>
          <div class="mt-1 font-mono text-[0.68rem] text-muted">
            {fmtHalfInning(swing.top_of_inning, swing.inning)} · LI {swing.li?.toFixed(2) ?? '—'}
          </div>
          <div class="mt-1 font-mono text-[0.68rem] text-muted">{swing.description ?? 'No description available.'}</div>
        {:else}
          <p class="font-mono text-[0.72rem] text-muted">No positive swing recorded.</p>
        {/if}
      </div>

      <div class="rounded border border-outline bg-surface p-3">
        <div class="mb-1 font-mono text-[0.62rem] tracking-wider text-muted uppercase">Biggest negative swing</div>
        {#if summaryResource.value.biggest_negative_swing}
          {@const swing = summaryResource.value.biggest_negative_swing}
          <div class="font-mono text-[0.72rem] text-foreground">{fmtSignedPoints(swing.we_change)}</div>
          <div class="mt-1 font-mono text-[0.68rem] text-muted">
            {fmtHalfInning(swing.top_of_inning, swing.inning)} · LI {swing.li?.toFixed(2) ?? '—'}
          </div>
          <div class="mt-1 font-mono text-[0.68rem] text-muted">{swing.description ?? 'No description available.'}</div>
        {:else}
          <p class="font-mono text-[0.72rem] text-muted">No negative swing recorded.</p>
        {/if}
      </div>
    </div>
  {:else}
    <p class="font-mono text-[0.78rem] text-muted">No summary payload returned.</p>
  {/if}
</div>

<div class="mt-4 rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-2">Leverage Table</div>
  <p class="mb-3 font-mono text-[0.68rem] text-muted">
    Top leverage plate appearances from <code>/v1/games/{gameId}/plate-appearances/leverage</code>.
  </p>

  {#if leverageResource.loading}
    <p class="font-mono text-[0.78rem] text-muted">Loading leverage rows…</p>
  {:else if leverageResource.error}
    <p class="font-mono text-[0.78rem] text-warning">{leverageResource.error}</p>
  {:else if topLeverageRows.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No leverage rows were returned.</p>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-[0.72rem]">
        <thead>
          <tr>
            {#each ['Event', 'Inning', 'LI', 'WE Before', 'WE After', 'Change', 'Description'] as col (col)}
              <th
                class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.69rem] font-medium whitespace-nowrap text-muted">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each topLeverageRows as row, index (`${row.event_id ?? ''}-${index}`)}
            <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-primary">{row.event_id ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted"
                >{fmtHalfInning(row.top_of_inning, row.inning)}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-foreground">{row.li?.toFixed(2) ?? '—'}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{fmtPct(row.we_before)}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{fmtPct(row.we_after)}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-foreground">{fmtSignedPoints(row.we_change)}</td>
              <td class="px-2 py-1.5 font-mono text-[0.69rem] text-muted">{row.description ?? '—'}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
