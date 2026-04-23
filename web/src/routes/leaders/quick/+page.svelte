<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, type Params } from '$lib/api';
  import { toNumber as toNum, toRecordArray, toString as toStr } from '$lib/common/converters';
  import { parseLeague } from '$lib/common/types';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import LeadersDataView from '$lib/components/LeadersDataView.svelte';
  import { EP } from '$lib/endpoints';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { QUICK_BATTING_STATS, QUICK_PITCHING_STATS } from '$lib/leaders/constants';
  import type { LeaderColumn, LeaderRow } from '$lib/leaders/types';
  import {
    endpointWithQuery,
    extractRows,
    formatMetric,
    toErrorMessage,
    toInnings,
    toSampleJson
  } from '$lib/leaders/utils';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  type QuickKind = 'batting' | 'pitching';

  const CURRENT_YEAR = new Date().getFullYear();

  let rawKind = $derived(page.url.searchParams.get('kind'));
  let rawSeason = $derived(page.url.searchParams.get('season'));
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawBattingStat = $derived(page.url.searchParams.get('bat'));
  let rawPitchingStat = $derived(page.url.searchParams.get('pit'));
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  let kind = $derived(parseQuickKind(rawKind));
  let season = $derived(parseSeason(rawSeason));
  let league = $derived(parseLeague(rawLeague));
  let battingStat = $derived(parseQuickBattingStat(rawBattingStat));
  let pitchingStat = $derived(parseQuickPitchingStat(rawPitchingStat));
  let stat = $derived(kind === 'batting' ? battingStat : pitchingStat);

  let seasons = $state<number[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  let rows = $state<LeaderRow[]>([]);
  let columns = $state<LeaderColumn[]>([]);
  let total = $state(0);

  let activeEndpoint = $state<string>(EP.seasonLeadersBatting(CURRENT_YEAR));
  let activeParams = $state<Params>({ stat: 'hr', page: 1, per_page: 20 });
  let sampleJson = $state<string | undefined>();

  let title = $derived(kind === 'batting' ? 'Batting leaders' : 'Pitching leaders');
  let subtitle = $derived(`${season} · ${stat.toUpperCase()}`);
  let endpointLabel = $derived(endpointWithQuery(`/v1${activeEndpoint}`, activeParams));
  let endpointUrl = $derived(apiUrl(activeEndpoint, activeParams));

  let lastKey = '';

  onMount(() => {
    void loadSeasons();
  });

  $effect(() => {
    void refresh();
  });

  async function loadSeasons(): Promise<void> {
    try {
      const payload = await apiFetch<unknown>(EP.seasons);
      const seasonRows = toRecordArray(payload);
      const years = seasonRows
        .map((row) => toNum(row.year))
        .filter((year): year is number => year != null)
        .toSorted((left, right) => right - left);

      seasons = years;

      if (years.length > 0 && !years.includes(season)) {
        updateQuery({ season: years[0], page: 1 });
      }
    } catch {
      seasons = [];
    }
  }

  async function refresh(): Promise<void> {
    const key = `${kind}|${season}|${league}|${battingStat}|${pitchingStat}|${currentPage}|${perPage}`;
    if (key === lastKey) return;
    lastKey = key;

    loading = true;
    error = null;

    const endpoint = kind === 'batting' ? EP.seasonLeadersBatting(season) : EP.seasonLeadersPitching(season);
    const params: Params = { stat, page: currentPage, per_page: perPage };

    if (league !== 'both') {
      params.league = league;
    }

    activeEndpoint = endpoint;
    activeParams = params;

    try {
      const payload = await apiFetch<unknown>(endpoint, params);
      const extracted = extractRows(payload, ['leaders', 'data']);
      rows = extracted.rows.map((row) => {
        const playerID = toStr(row.player_id) ?? 'unknown';
        const metric = stat === 'ip' ? toInnings(row.ip_outs) : toNum(row[stat]);

        return {
          id: playerID,
          label: playerID,
          team: toStr(row.team_id),
          league: toStr(row.league),
          year: toNum(row.year),
          metric,
          metricDisplay: formatMetric(metric, stat),
          kind: 'player' as const,
          href:
            kind === 'batting'
              ? `/players/${encodeURIComponent(playerID)}/batting`
              : `/players/${encodeURIComponent(playerID)}/pitching`
        } satisfies LeaderRow;
      });
      columns = quickColumns(stat);
      total = extracted.total;
      sampleJson = toSampleJson(payload);
    } catch (requestError) {
      rows = [];
      columns = [];
      total = 0;
      sampleJson = undefined;
      error = toErrorMessage(requestError, 'Failed to load quick leaders.');
    } finally {
      loading = false;
    }
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery('/leaders/quick', page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as '/leaders/quick'), QUERY_NAV_OPTS);
  }

  function quickColumns(activeStat: string): LeaderColumn[] {
    return [
      { key: 'rank', label: '#' },
      { key: 'label', label: 'Player' },
      { key: 'team', label: 'Team' },
      { key: 'league', label: 'Lg' },
      { key: 'year', label: 'Year', align: 'right' },
      { key: 'metricDisplay', label: activeStat.toUpperCase(), align: 'right' }
    ];
  }

  function parseQuickKind(value: string | null): QuickKind {
    return value === 'pitching' ? 'pitching' : 'batting';
  }

  function parseSeason(value: string | null): number {
    const parsed = Number.parseInt(value ?? '', 10);
    if (!Number.isFinite(parsed) || parsed <= 0) return CURRENT_YEAR;
    return parsed;
  }

  function parseQuickBattingStat(value: string | null): (typeof QUICK_BATTING_STATS)[number] {
    if (value && QUICK_BATTING_STATS.includes(value as (typeof QUICK_BATTING_STATS)[number])) {
      return value as (typeof QUICK_BATTING_STATS)[number];
    }
    return QUICK_BATTING_STATS[0];
  }

  function parseQuickPitchingStat(value: string | null): (typeof QUICK_PITCHING_STATS)[number] {
    if (value && QUICK_PITCHING_STATS.includes(value as (typeof QUICK_PITCHING_STATS)[number])) {
      return value as (typeof QUICK_PITCHING_STATS)[number];
    }
    return QUICK_PITCHING_STATS[0];
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Quick leaders</div>

    <div class="space-y-3">
      <div>
        <label for="quick-kind" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Category
        </label>
        <select
          id="quick-kind"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={kind}
          onchange={(event) => {
            const nextKind = parseQuickKind((event.target as HTMLSelectElement).value);
            updateQuery({ kind: nextKind, page: 1 });
          }}>
          <option value="batting">Batting</option>
          <option value="pitching">Pitching</option>
        </select>
      </div>

      <div>
        <label for="quick-season" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Season
        </label>
        <select
          id="quick-season"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={season}
          onchange={(event) => {
            const nextSeason = parseSeason((event.target as HTMLSelectElement).value);
            updateQuery({ season: nextSeason, page: 1 });
          }}>
          {#if seasons.length === 0}
            <option value={season}>{season}</option>
          {:else}
            {#each seasons as year (year)}
              <option value={year}>{year}</option>
            {/each}
          {/if}
        </select>
      </div>

      <div>
        <label for="quick-league" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          League
        </label>
        <select
          id="quick-league"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={league}
          onchange={(event) => {
            const nextLeague = parseLeague((event.target as HTMLSelectElement).value);
            updateQuery({ league: nextLeague === 'both' ? null : nextLeague, page: 1 });
          }}>
          <option value="both">Both</option>
          <option value="AL">AL</option>
          <option value="NL">NL</option>
        </select>
      </div>

      {#if kind === 'batting'}
        <div>
          <label
            for="quick-batting-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="quick-batting-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={battingStat}
            onchange={(event) => {
              const nextStat = parseQuickBattingStat((event.target as HTMLSelectElement).value);
              updateQuery({ bat: nextStat, page: 1 });
            }}>
            {#each QUICK_BATTING_STATS as option (option)}
              <option value={option}>{option.toUpperCase()}</option>
            {/each}
          </select>
        </div>
      {:else}
        <div>
          <label
            for="quick-pitching-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="quick-pitching-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={pitchingStat}
            onchange={(event) => {
              const nextStat = parseQuickPitchingStat((event.target as HTMLSelectElement).value);
              updateQuery({ pit: nextStat, page: 1 });
            }}>
            {#each QUICK_PITCHING_STATS as option (option)}
              <option value={option}>{option.toUpperCase()}</option>
            {/each}
          </select>
        </div>
      {/if}
    </div>
  {/snippet}

  {#snippet center()}
    <LeadersDataView
      {title}
      {subtitle}
      {rows}
      {columns}
      {loading}
      {error}
      page={currentPage}
      {perPage}
      {total}
      trendEmptyMessage="Quick mode returns one season at a time; use Query Lab for broader era trends."
      onPageChange={(nextPage) => updateQuery({ page: nextPage })}
      onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={endpointLabel} url={endpointUrl} {sampleJson} />

    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">Quick endpoints</div>
      <div class="space-y-2 font-mono text-[0.68rem] text-muted">
        <div>/v1{EP.seasonLeadersBatting(season)}</div>
        <div>/v1{EP.seasonLeadersPitching(season)}</div>
      </div>
    </div>
  {/snippet}
</ThreeColLayout>
