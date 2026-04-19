<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import { BATTING_STATS } from '$lib/common/constants';
  import Chart from '$lib/components/Chart.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { createBattingChartConfig } from '$lib/players/charts';
  import { normalizePlayerStatsPage } from '$lib/players/normalizers';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import type { BattingSeason, PlayerStatsPayload } from '$lib/players/types';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));
  let batStat = $derived(page.url.searchParams.get('stat') ?? 'hr');

  const battingResource = new AsyncPaginatedListResource<BattingSeason>();
  let lastKey = '';

  let battingRows = $derived(
    battingResource.items.map((season) => {
      const teamId = season.team_id?.trim() ?? '';
      const teamName = season.team_name?.trim() ?? '';
      const team = season.team?.trim() ?? '';
      const teamDisplay = team || teamId || '?';
      const era = eraForYear(season.year);

      return {
        ...season,
        team: teamDisplay,
        team_display: teamDisplay,
        team_lookup: teamId || teamDisplay,
        team_tooltip: teamName || teamDisplay,
        era_label: era?.label ?? '—'
      };
    })
  );

  const battingColumns = [
    { key: 'year', label: 'Year', sortable: true, href: (value: unknown) => `/seasons?year=${value}` },
    {
      key: 'team_display',
      label: 'Team',
      sortable: true,
      href: (_value: unknown, row: Record<string, unknown>) => {
        const lookup = String(row.team_lookup ?? '').trim();
        if (!lookup) return;
        return `/teams?${new URLSearchParams({ q: lookup }).toString()}`;
      },
      tooltip: (_value: unknown, row: Record<string, unknown>) => String(row.team_tooltip ?? '')
    },
    { key: 'era_label', label: 'Era', sortable: true },
    { key: 'g', label: 'G', sortable: true },
    { key: 'ab', label: 'AB', sortable: true },
    { key: 'h', label: 'H', sortable: true },
    { key: 'hr', label: 'HR', sortable: true },
    { key: 'rbi', label: 'RBI', sortable: true },
    { key: 'avg', label: 'AVG', sortable: true, format: (value: unknown) => fmtAvg(value as number | undefined) },
    { key: 'sb', label: 'SB', sortable: true, format: (value: unknown) => fmtNum(value as number | undefined) },
    { key: 'obp', label: 'OBP', sortable: true, format: (value: unknown) => fmtAvg(value as number | undefined) },
    { key: 'slg', label: 'SLG', sortable: true, format: (value: unknown) => fmtAvg(value as number | undefined) }
  ];

  let battingChartConfig = $derived(createBattingChartConfig(battingRows, batStat));

  const fmtAvg = (v: number | undefined) => (v != null ? Number(v).toFixed(3) : '—');
  const fmtNum = (v: number | undefined) => (v != null ? String(v) : '—');

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    const pageValue = currentPage;
    const perPageValue = perPage;
    const key = `${id}|${pageValue}|${perPageValue}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await battingResource.load(async () => {
      const payload = await apiFetch<PlayerStatsPayload<BattingSeason>>(EP.playerStatsBatting(id));
      return normalizePlayerStatsPage(payload, pageValue, perPageValue);
    });
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/players/${encodeURIComponent(playerId)}/batting`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/players/${string}/batting`), QUERY_NAV_OPTS);
  }
</script>

{#if battingResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if battingResource.items.length === 0}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No batting data found for this player.</p>
{:else}
  <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
    <div class="mb-3 flex items-center gap-3">
      <span class="panel-label">Career batting</span>
      <select
        value={batStat}
        onchange={(event) => updateQuery({ stat: (event.target as HTMLSelectElement).value, page: 1 })}
        class="ml-auto rounded border border-outline bg-surface px-2 py-1 font-mono text-xs text-muted focus:outline-none">
        {#each BATTING_STATS as s (s.value)}
          <option value={s.value}>{s.label}</option>
        {/each}
      </select>
    </div>
    <Chart config={battingChartConfig} height={110} />
  </div>

  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Season log</div>
    <SortableTable columns={battingColumns} rows={battingRows} />

    {#if battingResource.total > perPage}
      <div class="mt-4">
        <Pagination
          page={currentPage}
          {perPage}
          total={battingResource.total}
          onPageChange={(nextPage) => updateQuery({ page: nextPage })}
          onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
      </div>
    {/if}
  </div>
{/if}
