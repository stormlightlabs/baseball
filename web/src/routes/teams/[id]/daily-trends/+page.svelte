<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { fetchPaginated } from '$lib/api';
  import Chart from '$lib/components/Chart.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { AsyncListResource, AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeTeamDailyLogsPage, normalizeTeamDailyStatsPage } from '$lib/teams/normalizers';
  import type { TeamDailyLog, TeamDailyStat } from '$lib/teams/types';
  import { intParam } from '$lib/url-state.svelte';
  import type { ChartConfiguration } from 'chart.js';
  import { onMount } from 'svelte';

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 25));

  const dailyStatsResource = new AsyncPaginatedListResource<TeamDailyStat>();
  const dailyLogsResource = new AsyncListResource<TeamDailyLog>();
  let lastKey = '';

  let dailyLogRows = $derived.by(() =>
    [...dailyLogsResource.items].toSorted((a, b) => (a.date ?? '').localeCompare(b.date ?? ''))
  );

  let recentDailyLogRows = $derived.by(() => [...dailyLogRows].toReversed().slice(0, 30));

  let cumulativeDiffSeries = $derived.by(() => {
    let total = 0;
    return dailyLogRows.map((row) => {
      total += row.run_diff ?? 0;
      return total;
    });
  });

  let dailyTrendChartConfig = $derived.by(
    (): ChartConfiguration => ({
      type: 'line',
      data: {
        labels: dailyLogRows.map((row) => row.date ?? ''),
        datasets: [
          {
            label: 'Daily Run Diff',
            data: dailyLogRows.map((row) => row.run_diff ?? 0),
            borderColor: '#3b82f6',
            backgroundColor: 'rgba(59,130,246,0.15)',
            pointRadius: 2,
            tension: 0.25
          },
          {
            label: 'Cumulative Run Diff',
            data: cumulativeDiffSeries,
            borderColor: '#10b981',
            backgroundColor: 'rgba(16,185,129,0.15)',
            pointRadius: 0,
            tension: 0.2
          }
        ]
      },
      options: {
        responsive: true,
        plugins: { legend: { labels: { color: '#6b7280', font: { size: 10 } } } },
        scales: {
          x: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } },
          y: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } }
        }
      }
    })
  );

  let totals = $derived.by(() => {
    let games = 0;
    let wins = 0;
    let losses = 0;
    let rs = 0;
    let ra = 0;

    for (const row of dailyLogRows) {
      games += row.games_played ?? 0;
      wins += row.wins ?? 0;
      losses += row.losses ?? 0;
      rs += row.runs_scored ?? 0;
      ra += row.runs_allowed ?? 0;
    }

    return { games, wins, losses, runDiff: rs - ra };
  });

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!id || !y) {
      dailyStatsResource.clear();
      dailyLogsResource.clear();
      return;
    }

    const key = `${id}|${y}|${currentPage}|${perPage}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await Promise.all([
      dailyStatsResource.load(async () => {
        const payload = await fetchPaginated<Record<string, unknown>>(EP.teamDailyStats(id), {
          season: y,
          page: currentPage,
          per_page: perPage,
          sort_by: 'date',
          sort_order: 'desc'
        });
        return normalizeTeamDailyStatsPage(payload);
      }),
      dailyLogsResource.load(async () => {
        const payload = await fetchPaginated<Record<string, unknown>>(EP.seasonTeamDailyLogs(y, id), {
          page: 1,
          per_page: 400
        });
        return normalizeTeamDailyLogsPage(payload).data;
      })
    ]);
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/teams/${encodeURIComponent(teamId)}/daily-trends`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/teams/${string}/daily-trends`), QUERY_NAV_OPTS);
  }

  function fmtRate(value?: number): string {
    return value != null ? Number(value).toFixed(3) : '—';
  }

  function dayDiff(row: TeamDailyStat): number | null {
    if (row.runs == null || row.runs_allowed == null) return null;
    return row.runs - row.runs_allowed;
  }
</script>

{#if !year}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view daily trends.</p>
{:else}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Daily Trends — {year}</div>
    <div class="mb-4 grid grid-cols-4 gap-2 text-center">
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{totals.games}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Games</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{totals.wins}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Wins</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{totals.losses}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Losses</div>
      </div>
      <div class="rounded bg-surface p-2">
        <div class="font-display text-[0.9rem] text-foreground">{totals.runDiff}</div>
        <div class="font-mono text-[0.62rem] text-muted uppercase">Run Diff</div>
      </div>
    </div>

    {#if dailyLogsResource.loading}
      <p class="font-mono text-[0.78rem] text-muted">Loading trend chart…</p>
    {:else if dailyLogsResource.error}
      <p class="font-mono text-[0.78rem] text-warning">{dailyLogsResource.error}</p>
    {:else if dailyLogRows.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No daily log data found for {year}.</p>
    {:else}
      <Chart config={dailyTrendChartConfig} height={140} />

      <div class="mt-4 overflow-x-auto">
        <table class="w-full border-collapse text-[0.74rem]">
          <thead>
            <tr>
              {#each ['Date', 'Era', 'G', 'W', 'L', 'RS', 'RA', 'Diff'] as col (col)}
                <th
                  class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.72rem] font-medium whitespace-nowrap text-muted">
                  {col}
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each recentDailyLogRows as row, idx (`${row.date ?? ''}-${idx}`)}
              {@const rowYear = row.date ? Number(String(row.date).slice(0, 4)) : null}
              {@const era = rowYear ? eraForYear(rowYear) : null}
              <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                <td class="px-2 py-1.5 font-mono text-xs whitespace-nowrap text-foreground">{row.date ?? '—'}</td>
                <td class="px-2 py-1.5">
                  {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                </td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.games_played ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.wins ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.losses ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.runs_scored ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.runs_allowed ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{row.run_diff ?? '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Per-Game Daily Stats</div>
    {#if dailyStatsResource.loading}
      <p class="font-mono text-[0.78rem] text-muted">Loading daily stat rows…</p>
    {:else if dailyStatsResource.error}
      <p class="font-mono text-[0.78rem] text-warning">{dailyStatsResource.error}</p>
    {:else if dailyStatsResource.items.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No daily stat rows found for {year}.</p>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full border-collapse text-[0.74rem]">
          <thead>
            <tr>
              {#each ['Date', 'Result', 'R', 'RA', 'Diff', 'H', 'HR', 'BB', 'SO', 'AVG', 'OBP', 'SLG'] as col (col)}
                <th
                  class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.72rem] font-medium whitespace-nowrap text-muted">
                  {col}
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each dailyStatsResource.items as row, idx (`${row.game_id ?? row.date ?? ''}-${idx}`)}
              <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                <td class="px-2 py-1.5 font-mono text-xs whitespace-nowrap text-foreground">{row.date ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.result ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.runs ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.runs_allowed ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{dayDiff(row) ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.h ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.hr ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.bb ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.so ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{fmtRate(row.avg)}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{fmtRate(row.obp)}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{fmtRate(row.slg)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if dailyStatsResource.total > perPage}
        <div class="mt-4">
          <Pagination
            page={currentPage}
            {perPage}
            total={dailyStatsResource.total}
            onPageChange={(p) => updateQuery({ page: p })}
            onPerPageChange={(pp) => updateQuery({ per_page: pp, page: 1 })} />
        </div>
      {/if}
    {/if}
  </div>
{/if}
