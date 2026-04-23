<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, type Params } from '$lib/api';
  import { toNumber as toNum, toString as toStr } from '$lib/common/converters';
  import { parseLeague, type LeagueFilter } from '$lib/common/types';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import LeadersDataView from '$lib/components/LeadersDataView.svelte';
  import { EP } from '$lib/endpoints';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { labelForLabDataset, sortForLabDataset, type LabDataset } from '$lib/leaders/constants';
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

  const CURRENT_YEAR = new Date().getFullYear();

  let rawDataset = $derived(page.url.searchParams.get('dataset'));
  let rawSortBy = $derived(page.url.searchParams.get('sort_by'));
  let rawSortOrder = $derived(page.url.searchParams.get('sort_order'));
  let rawSeasonFrom = $derived(page.url.searchParams.get('season_from'));
  let rawSeasonTo = $derived(page.url.searchParams.get('season_to'));
  let rawSeason = $derived(page.url.searchParams.get('season'));
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawTeamId = $derived(page.url.searchParams.get('team_id'));
  let rawPlayerId = $derived(page.url.searchParams.get('player_id'));
  let rawPosition = $derived(page.url.searchParams.get('position'));
  let rawMinAB = $derived(page.url.searchParams.get('min_ab'));
  let rawMinIP = $derived(page.url.searchParams.get('min_ip'));
  let rawMinG = $derived(page.url.searchParams.get('min_g'));
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  let dataset = $derived(parseDataset(rawDataset));
  let sortBy = $derived(parseSortBy(rawSortBy, dataset));
  let sortOrder = $derived(parseSortOrder(rawSortOrder));
  let seasonFrom = $derived(parsePositiveInt(rawSeasonFrom, 2000));
  let seasonTo = $derived(parsePositiveInt(rawSeasonTo, CURRENT_YEAR));
  let exactSeason = $derived(parseOptionalSeason(rawSeason));
  let league = $derived(parseLeague(rawLeague));
  let teamId = $derived(normalizeTeamId(rawTeamId));
  let playerId = $derived(normalizePlayerId(rawPlayerId));
  let position = $derived(normalizePosition(rawPosition));
  let minAB = $derived(parsePositiveInt(rawMinAB, 250));
  let minIP = $derived(parsePositiveInt(rawMinIP, 100));
  let minG = $derived(parsePositiveInt(rawMinG, 120));

  let draftDataset = $state<LabDataset>('stats_batting');
  let draftSortBy = $state('hr');
  let draftSortOrder = $state<'asc' | 'desc'>('desc');
  let draftSeasonFrom = $state(2000);
  let draftSeasonTo = $state(CURRENT_YEAR);
  let draftSeason = $state('');
  let draftLeague = $state<LeagueFilter>('both');
  let draftTeamId = $state('');
  let draftPlayerId = $state('');
  let draftPosition = $state('');
  let draftMinAB = $state(250);
  let draftMinIP = $state(100);
  let draftMinG = $state(120);

  let loading = $state(false);
  let error = $state<string | null>(null);
  let rows = $state<LeaderRow[]>([]);
  let columns = $state<LeaderColumn[]>([]);
  let total = $state(0);

  let activeEndpoint = $state<string>(EP.statsBatting);
  let activeParams = $state<Params>({ sort_by: 'hr', sort_order: 'desc', page: 1, per_page: 20 });
  let sampleJson = $state<string | undefined>();
  let lastKey = '';

  let title = $derived('Query lab');
  let subtitle = $derived(`${labelForLabDataset(dataset)} · ${sortBy.toUpperCase()}`);
  let endpointLabel = $derived(endpointWithQuery(`/v1${activeEndpoint}`, activeParams));
  let endpointUrl = $derived(apiUrl(activeEndpoint, activeParams));

  $effect(() => {
    draftDataset = dataset;
    draftSortBy = sortBy;
    draftSortOrder = sortOrder;
    draftSeasonFrom = seasonFrom;
    draftSeasonTo = seasonTo;
    draftSeason = exactSeason?.toString() ?? '';
    draftLeague = league;
    draftTeamId = teamId;
    draftPlayerId = playerId;
    draftPosition = position;
    draftMinAB = minAB;
    draftMinIP = minIP;
    draftMinG = minG;
  });

  $effect(() => {
    void refresh();
  });

  async function refresh(): Promise<void> {
    const key = `${dataset}|${sortBy}|${sortOrder}|${seasonFrom}|${seasonTo}|${exactSeason ?? ''}|${league}|${teamId}|${playerId}|${position}|${minAB}|${minIP}|${minG}|${currentPage}|${perPage}`;
    if (key === lastKey) return;
    lastKey = key;

    loading = true;
    error = null;

    const endpoint = endpointForDataset(dataset);
    const params: Params = {
      sort_by: sortBy,
      sort_order: sortOrder,
      page: currentPage,
      per_page: perPage,
      season_from: seasonFrom,
      season_to: seasonTo
    };

    if (exactSeason != null) {
      params.season = exactSeason;
      delete params.season_from;
      delete params.season_to;
    }

    if (league !== 'both') params.league = league;
    if (teamId) params.team_id = teamId;
    if (playerId) params.player_id = playerId;
    if (position) params.position = position;
    if (dataset === 'stats_batting') params.min_ab = minAB;
    if (dataset === 'stats_pitching') params.min_ip = minIP;
    if (dataset === 'stats_fielding') params.min_g = minG;

    activeEndpoint = endpoint;
    activeParams = params;

    try {
      const payload = await apiFetch<unknown>(endpoint, params);
      const extracted = extractRows(payload, ['data']);
      rows = extracted.rows.map((row) => {
        const playerKey = toStr(row.player_id);
        const teamKey = toStr(row.team_id);
        const id = playerKey ?? teamKey ?? toStr(row.id) ?? 'unknown';

        let kind: LeaderRow['kind'] = 'unknown';
        if (playerKey) kind = 'player';
        if (!playerKey && teamKey) kind = 'team';

        const metric = labMetricValue(row, dataset, sortBy);

        let href: string | undefined;
        if (playerKey) href = `/players/${encodeURIComponent(playerKey)}/batting`;
        if (!playerKey && teamKey) href = `/teams/${encodeURIComponent(teamKey)}/overview`;

        return {
          id,
          label: id,
          team: teamKey,
          league: toStr(row.league),
          year: toNum(row.year),
          metric,
          metricDisplay: formatMetric(metric, sortBy),
          kind,
          href,
          detail: dataset.includes('fielding') ? toStr(row.position) : undefined
        } satisfies LeaderRow;
      });
      columns = labColumns(sortBy);
      total = extracted.total;
      sampleJson = toSampleJson(payload);
    } catch (requestError) {
      rows = [];
      columns = [];
      total = 0;
      sampleJson = undefined;
      error = toErrorMessage(requestError, 'Failed to load lab query results.');
    } finally {
      loading = false;
    }
  }

  function endpointForDataset(value: LabDataset): string {
    if (value === 'stats_batting') return EP.statsBatting;
    if (value === 'stats_pitching') return EP.statsPitching;
    if (value === 'stats_fielding') return EP.statsFielding;
    if (value === 'stats_teams_batting') return EP.statsTeamsBatting;
    if (value === 'stats_teams_pitching') return EP.statsTeamsPitching;
    return EP.statsTeamsFielding;
  }

  function labColumns(activeSort: string): LeaderColumn[] {
    return [
      { key: 'rank', label: '#' },
      { key: 'label', label: 'Entity' },
      { key: 'detail', label: 'Detail' },
      { key: 'team', label: 'Team' },
      { key: 'league', label: 'Lg' },
      { key: 'year', label: 'Year', align: 'right' },
      { key: 'metricDisplay', label: activeSort.toUpperCase(), align: 'right' }
    ];
  }

  function labMetricValue(
    row: Record<string, unknown>,
    activeDataset: LabDataset,
    activeSortBy: string
  ): number | undefined {
    const key = activeSortBy.toLowerCase();
    if (key === 'ip') return toInnings(row.ip_outs);

    const direct = toNum(row[key]);
    if (direct != null) return direct;

    if (activeDataset.includes('fielding') && key === 'fpct') {
      return toNum(row.fpct);
    }

    return undefined;
  }

  function applyFilters(): void {
    const nextDataset = parseDataset(draftDataset);
    const nextSortBy = parseSortBy(draftSortBy, nextDataset);
    const nextSeasonFrom = parsePositiveInt(String(draftSeasonFrom), seasonFrom);
    const nextSeasonTo = parsePositiveInt(String(draftSeasonTo), seasonTo);
    const nextExactSeason = parseOptionalSeason(draftSeason);
    const nextLeague = parseLeague(draftLeague);
    const nextTeamId = normalizeTeamId(draftTeamId);
    const nextPlayerId = normalizePlayerId(draftPlayerId);
    const nextPosition = normalizePosition(draftPosition);
    const nextMinAB = parsePositiveInt(String(draftMinAB), minAB);
    const nextMinIP = parsePositiveInt(String(draftMinIP), minIP);
    const nextMinG = parsePositiveInt(String(draftMinG), minG);

    updateQuery({
      dataset: nextDataset,
      sort_by: nextSortBy,
      sort_order: draftSortOrder,
      season_from: nextExactSeason == null ? nextSeasonFrom : null,
      season_to: nextExactSeason == null ? nextSeasonTo : null,
      season: nextExactSeason,
      league: nextLeague === 'both' ? null : nextLeague,
      team_id: nextTeamId || null,
      player_id: nextPlayerId || null,
      position: nextPosition || null,
      min_ab: nextMinAB,
      min_ip: nextMinIP,
      min_g: nextMinG,
      page: 1
    });
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery('/leaders/lab', page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as '/leaders/lab'), QUERY_NAV_OPTS);
  }

  function parseDataset(value: string | null): LabDataset {
    if (value === 'stats_pitching') return 'stats_pitching';
    if (value === 'stats_fielding') return 'stats_fielding';
    if (value === 'stats_teams_batting') return 'stats_teams_batting';
    if (value === 'stats_teams_pitching') return 'stats_teams_pitching';
    if (value === 'stats_teams_fielding') return 'stats_teams_fielding';
    return 'stats_batting';
  }

  function parseSortBy(value: string | null, activeDataset: LabDataset): string {
    const normalized = String(value ?? '')
      .trim()
      .toLowerCase();
    if (normalized.length > 0) return normalized;
    return sortForLabDataset(activeDataset);
  }

  function parseSortOrder(value: string | null): 'asc' | 'desc' {
    return value === 'asc' ? 'asc' : 'desc';
  }

  function parsePositiveInt(value: string | null, fallback: number): number {
    const parsed = Number.parseInt(value ?? '', 10);
    if (!Number.isFinite(parsed) || parsed <= 0) return fallback;
    return parsed;
  }

  function parseOptionalSeason(value: string | null): number | null {
    const parsed = Number.parseInt(value ?? '', 10);
    if (!Number.isFinite(parsed) || parsed <= 0) return null;
    return parsed;
  }

  function normalizeTeamId(value: string | null): string {
    return value?.trim().toUpperCase() ?? '';
  }

  function normalizePlayerId(value: string | null): string {
    return value?.trim().toLowerCase() ?? '';
  }

  function normalizePosition(value: string | null): string {
    return value?.trim().toUpperCase() ?? '';
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Query lab</div>

    <div class="space-y-3">
      <div>
        <label for="lab-dataset" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Dataset
        </label>
        <select
          id="lab-dataset"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          bind:value={draftDataset}
          onchange={(event) => {
            const nextDataset = parseDataset((event.target as HTMLSelectElement).value);
            draftDataset = nextDataset;
            draftSortBy = sortForLabDataset(nextDataset);
          }}>
          <option value="stats_batting">Player batting</option>
          <option value="stats_pitching">Player pitching</option>
          <option value="stats_fielding">Player fielding</option>
          <option value="stats_teams_batting">Team batting</option>
          <option value="stats_teams_pitching">Team pitching</option>
          <option value="stats_teams_fielding">Team fielding</option>
        </select>
      </div>

      <div class="grid grid-cols-2 gap-2">
        <div>
          <label
            for="lab-season-from"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">From</label>
          <input
            id="lab-season-from"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftSeasonFrom} />
        </div>
        <div>
          <label for="lab-season-to" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            To
          </label>
          <input
            id="lab-season-to"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftSeasonTo} />
        </div>
      </div>

      <div>
        <label for="lab-season-exact" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase"
          >Exact season</label>
        <input
          id="lab-season-exact"
          type="number"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          bind:value={draftSeason}
          placeholder="optional" />
      </div>

      <div class="grid grid-cols-2 gap-2">
        <div>
          <label for="lab-sort-by" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            Sort by
          </label>
          <input
            id="lab-sort-by"
            type="text"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftSortBy} />
        </div>
        <div>
          <label for="lab-sort-order" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase"
            >Order</label>
          <select
            id="lab-sort-order"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftSortOrder}>
            <option value="desc">desc</option>
            <option value="asc">asc</option>
          </select>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-2">
        <div>
          <label for="lab-league" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            League
          </label>
          <select
            id="lab-league"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftLeague}>
            <option value="both">Both</option>
            <option value="AL">AL</option>
            <option value="NL">NL</option>
          </select>
        </div>
        <div>
          <label for="lab-position" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase"
            >Position</label>
          <input
            id="lab-position"
            type="text"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftPosition}
            placeholder="fielding" />
        </div>
      </div>

      <div class="grid grid-cols-2 gap-2">
        <div>
          <label for="lab-player-id" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase"
            >Player ID</label>
          <input
            id="lab-player-id"
            type="text"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftPlayerId}
            placeholder="optional" />
        </div>
        <div>
          <label for="lab-team-id" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            Team ID
          </label>
          <input
            id="lab-team-id"
            type="text"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftTeamId}
            placeholder="optional" />
        </div>
      </div>

      <div class="grid grid-cols-3 gap-2">
        <div>
          <label for="lab-min-ab" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            min_ab
          </label>
          <input
            id="lab-min-ab"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftMinAB} />
        </div>
        <div>
          <label for="lab-min-ip" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            min_ip
          </label>
          <input
            id="lab-min-ip"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftMinIP} />
        </div>
        <div>
          <label for="lab-min-g" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
            min_g
          </label>
          <input
            id="lab-min-g"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            bind:value={draftMinG} />
        </div>
      </div>

      <button
        class="w-full rounded border border-outline bg-surface px-2 py-1.5 font-mono text-[0.72rem] text-foreground hover:border-primary"
        onclick={applyFilters}>
        Run query
      </button>
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
      trendEmptyMessage="Provide a wider season range to produce meaningful era buckets."
      onPageChange={(nextPage) => updateQuery({ page: nextPage })}
      onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={endpointLabel} url={endpointUrl} {sampleJson} />

    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">Lab endpoints</div>
      <div class="space-y-2 font-mono text-[0.68rem] text-muted">
        <div>/v1{EP.statsBatting}</div>
        <div>/v1{EP.statsPitching}</div>
        <div>/v1{EP.statsFielding}</div>
        <div>/v1{EP.statsTeamsBatting}</div>
        <div>/v1{EP.statsTeamsPitching}</div>
        <div>/v1{EP.statsTeamsFielding}</div>
      </div>
    </div>
  {/snippet}
</ThreeColLayout>
