<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { Pathname } from '$app/types';
  import { fetchPaginated } from '$lib/api';
  import Pagination from '$lib/components/Pagination.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { intParam, QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { gameLogEndpoint, parseGameLogType } from '$lib/players/tab-endpoints';
  import type { GameLog } from '$lib/players/types';
  import { rowColumns } from '$lib/players/types';
  import { onMount } from 'svelte';

  let playerId = $derived(page.params.id ?? '');
  let gameLogType = $derived(parseGameLogType(page.url.searchParams.get('log')));
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  const gameLogsResource = new AsyncPaginatedListResource<GameLog>();
  let lastKey = '';

  let columns = $derived(rowColumns(gameLogsResource.items));

  async function refresh(force = false): Promise<void> {
    const id = playerId;
    const logType = gameLogType;
    const pageValue = currentPage;
    const perPageValue = perPage;
    const key = `${id}|${logType}|${pageValue}|${perPageValue}`;
    if (!force && key === lastKey) return;
    lastKey = key;

    const endpoint = gameLogEndpoint(logType, id);
    await gameLogsResource.load(() => fetchPaginated<GameLog>(endpoint, { page: pageValue, per_page: perPageValue }));
  }

  onMount(() => {
    void refresh(true);
  });

  afterNavigate(() => {
    void refresh();
  });

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const encodedId = encodeURIComponent(playerId);
    const base = resolve(`/players/${encodedId}/game-logs` as Pathname);
    const href = withMergedQuery(base, page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as Pathname), QUERY_NAV_OPTS);
  }
</script>

<div class="mb-4 flex gap-1">
  {#each ['batting', 'pitching', 'fielding'] as type (type)}
    <button
      onclick={() => updateQuery({ log: type, page: 1 })}
      class="rounded border px-3 py-1.25 font-display text-[0.78rem] transition-all {gameLogType === type
        ? 'border-outline bg-surface text-foreground'
        : 'border-transparent text-muted hover:text-foreground'}">
      {type.charAt(0).toUpperCase() + type.slice(1)}
    </button>
  {/each}
</div>

<div class="rounded-lg border border-outline bg-crust p-4">
  {#if gameLogsResource.loading}
    <p class="font-mono text-xs text-muted">Loading…</p>
  {:else if gameLogsResource.items.length === 0}
    <p class="font-mono text-[0.78rem] text-muted">No {gameLogType} game logs found.</p>
  {:else}
    <div class="panel-label mb-3">
      {gameLogType.charAt(0).toUpperCase() + gameLogType.slice(1)} game logs
    </div>
    <SortableTable {columns} rows={gameLogsResource.items} />

    {#if gameLogsResource.total > perPage}
      <div class="mt-4">
        <Pagination
          page={currentPage}
          {perPage}
          total={gameLogsResource.total}
          onPageChange={(nextPage) => updateQuery({ page: nextPage })}
          onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
      </div>
    {/if}
  {/if}
</div>
