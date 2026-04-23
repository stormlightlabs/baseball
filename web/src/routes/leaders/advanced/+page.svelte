<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, type Params } from '$lib/api';
  import { toNumber as toNum, toObject, toRecordArray, toString as toStr } from '$lib/common/converters';
  import { parseLeague } from '$lib/common/types';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import LeadersDataView from '$lib/components/LeadersDataView.svelte';
  import { EP } from '$lib/endpoints';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { ADV_BATTING_STATS, ADV_PITCHING_STATS } from '$lib/leaders/constants';
  import type { LeaderColumn, LeaderRow } from '$lib/leaders/types';
  import { endpointWithQuery, extractRows, formatMetric, toErrorMessage, toSampleJson } from '$lib/leaders/utils';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  type AdvancedKind = 'batting' | 'pitching' | 'war';

  const CURRENT_YEAR = new Date().getFullYear();

  let rawKind = $derived(page.url.searchParams.get('kind'));
  let rawSeason = $derived(page.url.searchParams.get('season'));
  let rawBattingStat = $derived(page.url.searchParams.get('bat'));
  let rawPitchingStat = $derived(page.url.searchParams.get('pit'));
  let rawMinPA = $derived(page.url.searchParams.get('min_pa'));
  let rawMinIP = $derived(page.url.searchParams.get('min_ip'));
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawTeamId = $derived(page.url.searchParams.get('team_id'));
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  let kind = $derived(parseAdvancedKind(rawKind));
  let season = $derived(parseSeason(rawSeason));
  let battingStat = $derived(parseAdvancedBattingStat(rawBattingStat));
  let pitchingStat = $derived(parseAdvancedPitchingStat(rawPitchingStat));
  let minPA = $derived(parsePositiveInt(rawMinPA, 502));
  let minIP = $derived(parsePositiveInt(rawMinIP, 162));
  let league = $derived(parseLeague(rawLeague));
  let teamId = $derived(normalizeTeamId(rawTeamId));

  let seasons = $state<number[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  let rows = $state<LeaderRow[]>([]);
  let columns = $state<LeaderColumn[]>([]);
  let total = $state(0);

  let activeEndpoint = $state<string>(EP.seasonLeadersBattingAdv(CURRENT_YEAR));
  let activeParams = $state<Params>({ stat: 'WRC_PLUS', page: 1, per_page: 20 });
  let sampleJson = $state<string | undefined>();
  let lastKey = '';

  let statLabel = $derived.by(() => {
    if (kind === 'batting') return battingStat;
    if (kind === 'pitching') return pitchingStat;
    return 'WAR';
  });
  let title = $derived.by(() => {
    if (kind === 'war') return 'WAR leaders';
    return `${kind.toUpperCase()} advanced leaders`;
  });
  let subtitle = $derived(`${season} · ${statLabel}`);
  let endpointLabel = $derived(endpointWithQuery(`/v1${activeEndpoint}`, activeParams));
  let endpointUrl = $derived(apiUrl(activeEndpoint, activeParams));

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
    const key = `${kind}|${season}|${battingStat}|${pitchingStat}|${minPA}|${minIP}|${league}|${teamId}|${currentPage}|${perPage}`;
    if (key === lastKey) return;
    lastKey = key;

    loading = true;
    error = null;

    let endpoint: string = EP.seasonLeadersWar(season);
    const params: Params = { page: currentPage, per_page: perPage };

    if (kind === 'batting') {
      endpoint = EP.seasonLeadersBattingAdv(season);
      params.stat = battingStat;
      params.min_pa = minPA;
      if (league !== 'both') params.league = league;
    } else if (kind === 'pitching') {
      endpoint = EP.seasonLeadersPitchingAdv(season);
      params.stat = pitchingStat;
      params.min_ip = minIP;
    }

    if (teamId) params.team_id = teamId;

    activeEndpoint = endpoint;
    activeParams = params;

    try {
      const payload = await apiFetch<unknown>(endpoint, params);
      const extracted = extractRows(payload, ['data', 'leaders']);
      rows = extracted.rows.map((row) => {
        const playerID = toStr(row.player_id) ?? 'unknown';
        const context = toObject(row.context);
        const metric = kind === 'war' ? toNum(row.war) : advancedMetric(row, statLabel);
        const year = toNum(context.season);
        const leagueValue = toStr(context.league) ?? toStr(row.league);

        let href = `/players/${encodeURIComponent(playerID)}/war`;
        if (kind === 'batting') href = `/players/${encodeURIComponent(playerID)}/batting-adv`;
        if (kind === 'pitching') href = `/players/${encodeURIComponent(playerID)}/pitching-adv`;

        return {
          id: playerID,
          label: playerID,
          team: toStr(row.team_id),
          league: leagueValue,
          year,
          metric,
          metricDisplay: formatMetric(metric, statLabel),
          kind: 'player' as const,
          href
        } satisfies LeaderRow;
      });
      columns = [
        { key: 'rank', label: '#' },
        { key: 'label', label: 'Player' },
        { key: 'team', label: 'Team' },
        { key: 'league', label: 'Lg' },
        { key: 'year', label: 'Season', align: 'right' },
        { key: 'metricDisplay', label: statLabel.toUpperCase(), align: 'right' }
      ];
      total = extracted.total;
      sampleJson = toSampleJson(payload);
    } catch (requestError) {
      rows = [];
      columns = [];
      total = 0;
      sampleJson = undefined;
      error = toErrorMessage(requestError, 'Failed to load advanced leaders.');
    } finally {
      loading = false;
    }
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery('/leaders/advanced', page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as '/leaders/advanced'), QUERY_NAV_OPTS);
  }

  function advancedMetric(row: Record<string, unknown>, stat: string): number | undefined {
    const lowered = stat.toLowerCase();
    if (lowered === 'k_rate') return toNum(row.k_rate);
    if (lowered === 'bb_rate') return toNum(row.bb_rate);
    if (lowered === 'wrc_plus') return toNum(row.wrc_plus);
    if (lowered === 'k_per_9') return toNum(row.k_per_9);
    if (lowered === 'bb_per_9') return toNum(row.bb_per_9);
    if (lowered === 'hr_per_9') return toNum(row.hr_per_9);
    return toNum(row[lowered]);
  }

  function parseAdvancedKind(value: string | null): AdvancedKind {
    if (value === 'pitching') return 'pitching';
    if (value === 'war') return 'war';
    return 'batting';
  }

  function parseSeason(value: string | null): number {
    const parsed = Number.parseInt(value ?? '', 10);
    if (!Number.isFinite(parsed) || parsed <= 0) return CURRENT_YEAR;
    return parsed;
  }

  function parseAdvancedBattingStat(value: string | null): (typeof ADV_BATTING_STATS)[number] {
    if (value && ADV_BATTING_STATS.includes(value as (typeof ADV_BATTING_STATS)[number])) {
      return value as (typeof ADV_BATTING_STATS)[number];
    }
    return ADV_BATTING_STATS[0];
  }

  function parseAdvancedPitchingStat(value: string | null): (typeof ADV_PITCHING_STATS)[number] {
    if (value && ADV_PITCHING_STATS.includes(value as (typeof ADV_PITCHING_STATS)[number])) {
      return value as (typeof ADV_PITCHING_STATS)[number];
    }
    return ADV_PITCHING_STATS[0];
  }

  function parsePositiveInt(value: string | null, fallback: number): number {
    const parsed = Number.parseInt(value ?? '', 10);
    if (!Number.isFinite(parsed) || parsed <= 0) return fallback;
    return parsed;
  }

  function normalizeTeamId(value: string | null): string {
    return value?.trim().toUpperCase() ?? '';
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Advanced leaders</div>

    <div class="space-y-3">
      <div>
        <label for="advanced-season" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Season
        </label>
        <select
          id="advanced-season"
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
        <label for="advanced-kind" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Category
        </label>
        <select
          id="advanced-kind"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={kind}
          onchange={(event) => {
            const nextKind = parseAdvancedKind((event.target as HTMLSelectElement).value);
            updateQuery({ kind: nextKind, page: 1 });
          }}>
          <option value="batting">Batting advanced</option>
          <option value="pitching">Pitching advanced</option>
          <option value="war">WAR</option>
        </select>
      </div>

      {#if kind === 'batting'}
        <div>
          <label
            for="advanced-bat-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="advanced-bat-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={battingStat}
            onchange={(event) => {
              const nextStat = parseAdvancedBattingStat((event.target as HTMLSelectElement).value);
              updateQuery({ bat: nextStat, page: 1 });
            }}>
            {#each ADV_BATTING_STATS as option (option)}
              <option value={option}>{option}</option>
            {/each}
          </select>
        </div>

        <div>
          <label
            for="advanced-min-pa"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">min_pa</label>
          <input
            id="advanced-min-pa"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={minPA}
            onchange={(event) => {
              const nextValue = parsePositiveInt((event.target as HTMLInputElement).value, minPA);
              updateQuery({ min_pa: nextValue, page: 1 });
            }} />
        </div>

        <div>
          <label
            for="advanced-league"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">League</label>
          <select
            id="advanced-league"
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
      {/if}

      {#if kind === 'pitching'}
        <div>
          <label
            for="advanced-pit-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="advanced-pit-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={pitchingStat}
            onchange={(event) => {
              const nextStat = parseAdvancedPitchingStat((event.target as HTMLSelectElement).value);
              updateQuery({ pit: nextStat, page: 1 });
            }}>
            {#each ADV_PITCHING_STATS as option (option)}
              <option value={option}>{option}</option>
            {/each}
          </select>
        </div>

        <div>
          <label
            for="advanced-min-ip"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">min_ip</label>
          <input
            id="advanced-min-ip"
            type="number"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={minIP}
            onchange={(event) => {
              const nextValue = parsePositiveInt((event.target as HTMLInputElement).value, minIP);
              updateQuery({ min_ip: nextValue, page: 1 });
            }} />
        </div>
      {/if}

      <div>
        <label for="advanced-team-id" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase"
          >Team ID</label>
        <input
          id="advanced-team-id"
          type="text"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={teamId}
          placeholder="optional"
          onchange={(event) => {
            const nextTeamId = normalizeTeamId((event.target as HTMLInputElement).value);
            updateQuery({ team_id: nextTeamId || null, page: 1 });
          }} />
      </div>
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
      trendEmptyMessage="Advanced rows include season context when available; use broader filters for richer era trends."
      onPageChange={(nextPage) => updateQuery({ page: nextPage })}
      onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={endpointLabel} url={endpointUrl} {sampleJson} />

    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">Advanced endpoints</div>
      <div class="space-y-2 font-mono text-[0.68rem] text-muted">
        <div>/v1{EP.seasonLeadersBattingAdv(season)}</div>
        <div>/v1{EP.seasonLeadersPitchingAdv(season)}</div>
        <div>/v1{EP.seasonLeadersWar(season)}</div>
      </div>
    </div>
  {/snippet}
</ThreeColLayout>
