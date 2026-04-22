<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { fetchPaginated } from '$lib/api';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  type GameRow = {
    game_id?: string;
    date?: string;
    home_team?: string;
    away_team?: string;
    home_score?: number;
    away_score?: number;
    park?: string;
    [key: string]: unknown;
  };

  let teamId = $derived(page.params.id ?? '');
  let year = $derived(page.url.searchParams.get('year') ?? '');
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 25));

  const scheduleResource = new AsyncPaginatedListResource<GameRow>();
  let lastKey = '';

  async function refresh(force = false): Promise<void> {
    const id = teamId;
    const y = year;
    if (!y) {
      scheduleResource.clear();
      return;
    }
    const key = `${id}|${y}|${currentPage}|${perPage}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    await scheduleResource.load(() =>
      fetchPaginated<GameRow>(EP.seasonTeamSchedule(y, id), { page: currentPage, per_page: perPage })
    );
  }

  onMount(() => {
    void refresh(true);
  });
  afterNavigate(() => {
    void refresh();
  });

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(
      `/teams/${encodeURIComponent(teamId)}/schedule`,
      page.url.searchParams,
      overrides,
      page.url.hash
    );
    void goto(resolve(href as `/teams/${string}/schedule`), QUERY_NAV_OPTS);
  }

  function fmtScore(row: GameRow): string {
    if (row.home_score == null && row.away_score == null) return '—';
    return `${row.away_score ?? '?'} – ${row.home_score ?? '?'}`;
  }
</script>

{#if !year}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Enter a season year in the sidebar to view the schedule.</p>
{:else if scheduleResource.loading}
  <p class="mt-4 font-mono text-[0.78rem] text-muted">Loading…</p>
{:else if scheduleResource.error}
  <p class="mt-4 font-mono text-[0.78rem] text-warning">{scheduleResource.error}</p>
{:else}
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="panel-label mb-3">{year} Schedule</div>
    {#if scheduleResource.items.length === 0}
      <p class="font-mono text-[0.78rem] text-muted">No schedule data found for {year}.</p>
    {:else}
      <div class="overflow-x-auto">
        <table class="w-full border-collapse text-xs">
          <thead>
            <tr>
              {#each ['Date', 'Era', 'Away', 'Home', 'Score', 'Park'] as col (col)}
                <th
                  class="border-b border-outline px-2 py-1.5 text-left font-sans text-[0.72rem] font-medium whitespace-nowrap text-muted">
                  {col}
                </th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each scheduleResource.items as row, idx (`${row.game_id ?? ''}-${idx}`)}
              {@const era = row.date ? eraForYear(Number(String(row.date).slice(0, 4))) : null}
              <tr class="border-b border-outline last:border-b-0 hover:[&>td]:bg-surface">
                <td class="px-2 py-1.5 font-mono text-xs whitespace-nowrap text-foreground">{row.date ?? '—'}</td>
                <td class="px-2 py-1.5">
                  {#if era}<EraBadge {era} size="xs" />{:else}<span class="text-muted">—</span>{/if}
                </td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.away_team ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-foreground">{row.home_team ?? '—'}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{fmtScore(row)}</td>
                <td class="px-2 py-1.5 font-mono text-xs text-muted">{row.park ?? '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>

      {#if scheduleResource.total > perPage}
        <div class="mt-4">
          <Pagination
            page={currentPage}
            {perPage}
            total={scheduleResource.total}
            onPageChange={(p) => updateQuery({ page: p })}
            onPerPageChange={(pp) => updateQuery({ per_page: pp, page: 1 })} />
        </div>
      {/if}
    {/if}
  </div>
{/if}
