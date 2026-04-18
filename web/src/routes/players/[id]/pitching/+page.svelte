<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch } from '$lib/api';
  import Chart from '$lib/components/Chart.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import type { PlayerStatsPayload } from '$lib/players/api-payloads';
  import { createPitchingChartConfig } from '$lib/players/charts';
  import { normalizePlayerStatsPage } from '$lib/players/normalizers';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { intParam, QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import type { PitchingSeason } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  const pitchingResource = new AsyncPaginatedListResource<PitchingSeason>();
  let lastKey = '';

  let pitchingRows = $derived(pitchingResource.items.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));
  let pitchingChartConfig = $derived(createPitchingChartConfig(pitchingRows));

  const fmtNum = (v: number | undefined) => (v != null ? String(v) : '—');

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    const pageValue = currentPage;
    const perPageValue = perPage;
    const key = `${id}|${pageValue}|${perPageValue}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await pitchingResource.load(async () => {
      const payload = await apiFetch<PlayerStatsPayload<PitchingSeason>>(EP.playerStatsPitching(id));
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
    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-[0.75rem]">
        <thead>
          <tr>
            {#each ['Year', 'Team', 'Era', 'G', 'GS', 'W', 'L', 'SV', 'IP', 'SO', 'BB', 'ERA', 'WHIP'] as col (col)}
              <th
                class="border-b border-outline px-2 py-1.5 text-left font-sans text-xs font-medium whitespace-nowrap text-muted">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each pitchingRows as row (`${row.year}-${row.team}`)}
            {@const era = eraForYear(row.year)}
            <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">
                <a href={resolve(`/seasons?year=${row.year}`)} class="text-primary hover:underline">
                  {row.year}
                </a>
              </td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.team}</td>
              <td class="px-2 py-1.5">
                {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
              </td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.g}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.gs)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.w}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.l}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.sv)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.ip}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.so)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.bb)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{Number(row.era).toFixed(2)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">
                {row.whip != null ? Number(row.whip).toFixed(2) : '—'}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

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
