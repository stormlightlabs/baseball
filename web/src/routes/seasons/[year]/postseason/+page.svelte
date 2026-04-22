<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated, type PaginatedResponse } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { normalizeGamesPage, normalizePostseasonSeries } from '$lib/seasons/normalizers';
  import type { SeasonGame, SeasonPostseasonSeries } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : new Date().getFullYear());

  let postseasonSeries = $state<SeasonPostseasonSeries[]>([]);
  let postseasonGamesPage = $state<PaginatedResponse<SeasonGame>>(emptyPage());
  let postseasonLoading = $state(false);
  let postseasonError = $state<string | null>(null);

  let postseasonRequestVersion = 0;
  let lastPostseasonKey = '';

  let postseasonSeriesRows = $derived.by(() => {
    return postseasonSeries
      .map((series) => {
        const winner = series.winner_team ?? 'TBD';
        const loser = series.loser_team ?? 'TBD';
        const wins = series.wins ?? 0;
        const losses = series.losses ?? 0;
        return { round: series.round ?? 'Round', matchup: `${winner} vs ${loser}`, result: `${wins}-${losses}` };
      })
      .toSorted((a, b) => a.round.localeCompare(b.round));
  });

  const postseasonSeriesColumns = [
    { key: 'round', label: 'Round', sortable: true },
    { key: 'matchup', label: 'Series', sortable: true },
    { key: 'result', label: 'Result', sortable: true }
  ];

  onMount(() => {
    void refreshPostseason(true);
  });

  afterNavigate(() => {
    void refreshPostseason();
  });

  async function refreshPostseason(force = false): Promise<void> {
    const key = String(selectedYear);
    if (!force && key === lastPostseasonKey) return;
    lastPostseasonKey = key;

    const requestVersion = ++postseasonRequestVersion;
    postseasonLoading = true;
    postseasonError = null;

    const seriesPromise = apiFetch<unknown>(EP.seasonPostseasonSeries(selectedYear));
    const gamesPromise = fetchPaginated<unknown>(EP.seasonPostseasonGames(selectedYear), { page: 1, per_page: 20 });

    const [seriesResult, gamesResult] = await Promise.allSettled([seriesPromise, gamesPromise]);

    if (requestVersion !== postseasonRequestVersion) return;

    if (seriesResult.status === 'fulfilled') {
      postseasonSeries = normalizePostseasonSeries(seriesResult.value);
    } else {
      postseasonSeries = [];
      postseasonError = toErrorMessage(seriesResult.reason, 'Failed to load postseason series.');
    }

    if (gamesResult.status === 'fulfilled') {
      postseasonGamesPage = normalizeGamesPage(gamesResult.value);
    } else {
      postseasonGamesPage = emptyPage();
      const message = toErrorMessage(gamesResult.reason, 'Failed to load postseason games.');
      postseasonError = postseasonError ? `${postseasonError} ${message}` : message;
    }

    postseasonLoading = false;
  }

  function emptyPage<T>(): PaginatedResponse<T> {
    return { data: [], page: 1, per_page: 1, total: 0 };
  }

  function toErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof Error && error.message.trim().length > 0) return error.message;
    return fallback;
  }
</script>

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-2">Postseason modules</div>

  {#if postseasonError}
    <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {postseasonError}
    </p>
  {/if}

  {#if postseasonLoading}
    <p class="font-mono text-[0.75rem] text-muted">Loading postseason data...</p>
  {:else}
    <div class="mb-2 flex items-center justify-between rounded border border-outline bg-surface px-3 py-2">
      <span class="font-mono text-[0.68rem] text-muted uppercase">Postseason games</span>
      <span class="font-display text-[0.92rem] text-foreground">{postseasonGamesPage.total}</span>
    </div>

    {#if postseasonSeriesRows.length === 0}
      <p class="font-mono text-[0.75rem] text-muted">No postseason series returned for this season.</p>
    {:else}
      <SortableTable columns={postseasonSeriesColumns} rows={postseasonSeriesRows} />
    {/if}
  {/if}
</section>
