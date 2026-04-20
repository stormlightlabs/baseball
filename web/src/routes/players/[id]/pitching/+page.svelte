<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import Chart from '$lib/components/Chart.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { createPitchingChartConfig } from '$lib/players/charts';
  import { normalizePlayerStatsPage } from '$lib/players/normalizers';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import type { PitchingSeason, PlayerStatsPayload } from '$lib/players/types';
  import { intParam } from '$lib/url-state.svelte';

  let playerId = $derived(page.params.id ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  const pitchingResource = new AsyncPaginatedListResource<PitchingSeason>();
  let lastKey = '';

  let pitchingRows = $derived(
    pitchingResource.items.map((season) => {
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

  const pitchingColumns = [
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
    { key: 'gs', label: 'GS', sortable: true, format: (value: unknown) => fmtNum(value as number | undefined) },
    { key: 'w', label: 'W', sortable: true },
    { key: 'l', label: 'L', sortable: true },
    { key: 'sv', label: 'SV', sortable: true, format: (value: unknown) => fmtNum(value as number | undefined) },
    { key: 'ip', label: 'IP', sortable: true },
    { key: 'so', label: 'SO', sortable: true, format: (value: unknown) => fmtNum(value as number | undefined) },
    { key: 'bb', label: 'BB', sortable: true, format: (value: unknown) => fmtNum(value as number | undefined) },
    {
      key: 'era',
      label: 'ERA',
      sortable: true,
      format: (value: unknown) => (value == null ? '—' : Number(value).toFixed(2))
    },
    {
      key: 'whip',
      label: 'WHIP',
      sortable: true,
      format: (value: unknown) => (value == null ? '—' : Number(value).toFixed(2))
    }
  ];

  let pitchingChartConfig = $derived(createPitchingChartConfig(pitchingRows));

  const fmtNum = (v: number | undefined) => (v != null ? String(v) : '—');

  async function refresh(): Promise<void> {
    const id = playerId;
    const pageValue = currentPage;
    const perPageValue = perPage;
    const key = `${id}|${pageValue}|${perPageValue}`;
    if (key === lastKey) return;
    lastKey = key;

    await pitchingResource.load(async () => {
      const payload = await apiFetch<PlayerStatsPayload<PitchingSeason>>(EP.playerStatsPitching(id));
      return normalizePlayerStatsPage(payload, pageValue, perPageValue);
    });
  }

  $effect(() => {
    void refresh();
  });

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/players/${encodeURIComponent(playerId)}/pitching`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/players/${string}/pitching`), QUERY_NAV_OPTS);
  }
</script>

{#if pitchingResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if pitchingResource.items.length === 0}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">No pitching data found for this player.</p>
{:else}
  <div class="mb-4 rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Career ERA</div>
    <Chart config={pitchingChartConfig} height={110} />
  </div>

  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">Season log</div>
    <SortableTable columns={pitchingColumns} rows={pitchingRows} />

    {#if pitchingResource.total > perPage}
      <div class="mt-4">
        <Pagination
          page={currentPage}
          {perPage}
          total={pitchingResource.total}
          onPageChange={(nextPage) => updateQuery({ page: nextPage })}
          onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
      </div>
    {/if}
  </div>
{/if}
