<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { page } from '$app/state';
  import { fetchPaginated, type PaginatedResponse } from '$lib/api';
  import { parseLeague } from '$lib/common/types';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { emptyPage, fmtInt, toErrorMessage } from '$lib/leaders/utils';
  import { normalizeAwardsPage } from '$lib/seasons/normalizers';
  import type { SeasonAwardResult } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : new Date().getFullYear());
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let leagueFilter = $derived(parseLeague(rawLeague));

  let awardsPage = $state<PaginatedResponse<SeasonAwardResult>>(emptyPage());
  let awardsLoading = $state(false);
  let awardsError = $state<string | null>(null);
  let awardsRequestVersion = 0;
  let lastAwardsKey = '';

  let awardsRows = $derived.by(() => {
    return awardsPage.data.map((entry) => ({
      award_id: entry.award_id ?? '—',
      player_id: entry.player_id ?? '—',
      league: entry.league ?? '—',
      rank: entry.rank,
      points: entry.points
    }));
  });

  const awardsColumns = [
    { key: 'award_id', label: 'Award', sortable: true },
    {
      key: 'player_id',
      label: 'Player',
      sortable: true,
      href: (value: unknown) => {
        const playerId = String(value ?? '').trim();
        if (!playerId || playerId === '—') return;
        return `/players/${encodeURIComponent(playerId)}/awards`;
      }
    },
    { key: 'league', label: 'Lg', sortable: true },
    { key: 'rank', label: 'Rank', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'points', label: 'Points', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) }
  ];

  onMount(() => {
    void refreshAwards(true);
  });

  afterNavigate(() => {
    void refreshAwards();
  });

  async function refreshAwards(force = false): Promise<void> {
    const key = `${selectedYear}|${leagueFilter}`;
    if (!force && key === lastAwardsKey) return;
    lastAwardsKey = key;

    const requestVersion = ++awardsRequestVersion;
    awardsLoading = true;
    awardsError = null;

    try {
      const params: Record<string, string | number> = { page: 1, per_page: 20 };
      if (leagueFilter !== 'both') params.league = leagueFilter.toUpperCase();

      const payload = await fetchPaginated<unknown>(EP.seasonAwards(selectedYear), params);
      if (requestVersion !== awardsRequestVersion) return;
      awardsPage = normalizeAwardsPage(payload);
    } catch (error) {
      if (requestVersion !== awardsRequestVersion) return;
      awardsError = toErrorMessage(error, 'Failed to load season awards.');
      awardsPage = emptyPage();
    } finally {
      if (requestVersion === awardsRequestVersion) awardsLoading = false;
    }
  }
</script>

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="panel-label mb-3">Awards snapshot</div>

  {#if awardsError}
    <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {awardsError}
    </p>
  {/if}

  {#if awardsLoading}
    <p class="font-mono text-xs text-muted">Loading season awards...</p>
  {:else if awardsRows.length === 0}
    <p class="font-mono text-xs text-muted">No awards were returned for this season.</p>
  {:else}
    <SortableTable columns={awardsColumns} rows={awardsRows} />
  {/if}
</section>
