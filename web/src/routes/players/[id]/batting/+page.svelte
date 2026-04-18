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
  import { createBattingChartConfig } from '$lib/players/charts';
  import { BATTING_STATS } from '$lib/players/constants';
  import { normalizePlayerStatsPage } from '$lib/players/normalizers';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { intParam, QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import type { BattingSeason } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));
  let batStat = $derived(page.url.searchParams.get('stat') ?? 'hr');

  const battingResource = new AsyncPaginatedListResource<BattingSeason>();
  let lastKey = '';

  let battingRows = $derived(battingResource.items.map((s) => ({ ...s, team: s.team ?? s.team_id ?? '?' })));
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
    <div class="overflow-x-auto">
      <table class="w-full border-collapse text-[0.75rem]">
        <thead>
          <tr>
            {#each ['Year', 'Team', 'Era', 'G', 'AB', 'H', 'HR', 'RBI', 'AVG', 'SB', 'OBP', 'SLG'] as col (col)}
              <th
                class="border-b border-outline px-2 py-1.5 text-left font-sans text-xs font-medium whitespace-nowrap text-muted">
                {col}
              </th>
            {/each}
          </tr>
        </thead>
        <tbody>
          {#each battingRows as row (`${row.year}-${row.team}`)}
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
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.ab}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.h}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.hr}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.rbi}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.avg)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtNum(row.sb)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.obp)}</td>
              <td class="px-2 py-1.5 font-mono text-xs text-foreground">{fmtAvg(row.slg)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

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
