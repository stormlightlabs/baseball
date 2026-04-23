<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, type Params } from '$lib/api';
  import { toNumber as toNum, toString as toStr } from '$lib/common/converters';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import LeadersDataView from '$lib/components/LeadersDataView.svelte';
  import { EP } from '$lib/endpoints';
  import { CAREER_BATTING_STATS, CAREER_PITCHING_STATS } from '$lib/leaders/constants';
  import type { LeaderColumn, LeaderRow } from '$lib/leaders/types';
  import {
    endpointWithQuery,
    extractRows,
    formatMetric,
    toErrorMessage,
    toInnings,
    toSampleJson
  } from '$lib/leaders/utils';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';

  type CareerKind = 'batting' | 'pitching';

  let rawKind = $derived(page.url.searchParams.get('kind'));
  let rawBattingStat = $derived(page.url.searchParams.get('bat'));
  let rawPitchingStat = $derived(page.url.searchParams.get('pit'));
  let currentPage = $derived(intParam(page.url.searchParams, 'page', 1));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', 20));

  let kind = $derived(parseCareerKind(rawKind));
  let battingStat = $derived(parseCareerBattingStat(rawBattingStat));
  let pitchingStat = $derived(parseCareerPitchingStat(rawPitchingStat));
  let stat = $derived(kind === 'batting' ? battingStat : pitchingStat);

  let loading = $state(false);
  let error = $state<string | null>(null);

  let rows = $state<LeaderRow[]>([]);
  let columns = $state<LeaderColumn[]>([]);
  let total = $state(0);

  let activeEndpoint = $state<string>(EP.leadersBattingCareer);
  let activeParams = $state<Params>({ stat: 'hr', page: 1, per_page: 20 });
  let sampleJson = $state<string | undefined>();
  let lastKey = '';

  let title = $derived(kind === 'batting' ? 'Career batting leaders' : 'Career pitching leaders');
  let subtitle = $derived(stat.toUpperCase());
  let endpointLabel = $derived(endpointWithQuery(`/v1${activeEndpoint}`, activeParams));
  let endpointUrl = $derived(apiUrl(activeEndpoint, activeParams));

  $effect(() => {
    void refresh();
  });

  async function refresh(): Promise<void> {
    const key = `${kind}|${battingStat}|${pitchingStat}|${currentPage}|${perPage}`;
    if (key === lastKey) return;
    lastKey = key;

    loading = true;
    error = null;

    const endpoint = kind === 'batting' ? EP.leadersBattingCareer : EP.leadersPitchingCareer;
    const params: Params = { stat, page: currentPage, per_page: perPage };
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
      columns = careerColumns(stat);
      total = extracted.total;
      sampleJson = toSampleJson(payload);
    } catch (requestError) {
      rows = [];
      columns = [];
      total = 0;
      sampleJson = undefined;
      error = toErrorMessage(requestError, 'Failed to load career leaders.');
    } finally {
      loading = false;
    }
  }

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery('/leaders/career', page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as '/leaders/career'), QUERY_NAV_OPTS);
  }

  function careerColumns(activeStat: string): LeaderColumn[] {
    return [
      { key: 'rank', label: '#' },
      { key: 'label', label: 'Player' },
      { key: 'team', label: 'Team' },
      { key: 'league', label: 'Lg' },
      { key: 'metricDisplay', label: activeStat.toUpperCase(), align: 'right' }
    ];
  }

  function parseCareerKind(value: string | null): CareerKind {
    return value === 'pitching' ? 'pitching' : 'batting';
  }

  function parseCareerBattingStat(value: string | null): (typeof CAREER_BATTING_STATS)[number] {
    if (value && CAREER_BATTING_STATS.includes(value as (typeof CAREER_BATTING_STATS)[number])) {
      return value as (typeof CAREER_BATTING_STATS)[number];
    }
    return CAREER_BATTING_STATS[0];
  }

  function parseCareerPitchingStat(value: string | null): (typeof CAREER_PITCHING_STATS)[number] {
    if (value && CAREER_PITCHING_STATS.includes(value as (typeof CAREER_PITCHING_STATS)[number])) {
      return value as (typeof CAREER_PITCHING_STATS)[number];
    }
    return CAREER_PITCHING_STATS[0];
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Career leaders</div>

    <div class="space-y-3">
      <div>
        <label for="career-kind" class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">
          Category
        </label>
        <select
          id="career-kind"
          class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
          value={kind}
          onchange={(event) => {
            const nextKind = parseCareerKind((event.target as HTMLSelectElement).value);
            updateQuery({ kind: nextKind, page: 1 });
          }}>
          <option value="batting">Batting</option>
          <option value="pitching">Pitching</option>
        </select>
      </div>

      {#if kind === 'batting'}
        <div>
          <label
            for="career-bat-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="career-bat-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={battingStat}
            onchange={(event) => {
              const nextStat = parseCareerBattingStat((event.target as HTMLSelectElement).value);
              updateQuery({ bat: nextStat, page: 1 });
            }}>
            {#each CAREER_BATTING_STATS as option (option)}
              <option value={option}>{option.toUpperCase()}</option>
            {/each}
          </select>
        </div>
      {:else}
        <div>
          <label
            for="career-pit-stat"
            class="mb-1 block font-mono text-[0.62rem] tracking-[0.08em] text-muted uppercase">Stat</label>
          <select
            id="career-pit-stat"
            class="w-full rounded border border-outline bg-crust px-2 py-1.5 font-mono text-[0.72rem]"
            value={pitchingStat}
            onchange={(event) => {
              const nextStat = parseCareerPitchingStat((event.target as HTMLSelectElement).value);
              updateQuery({ pit: nextStat, page: 1 });
            }}>
            {#each CAREER_PITCHING_STATS as option (option)}
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
      trendEmptyMessage="Career endpoint output is aggregate-heavy; trend buckets depend on year-tagged rows."
      onPageChange={(nextPage) => updateQuery({ page: nextPage })}
      onPerPageChange={(nextPerPage) => updateQuery({ per_page: nextPerPage, page: 1 })} />
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={endpointLabel} url={endpointUrl} {sampleJson} />

    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">Career endpoints</div>
      <div class="space-y-2 font-mono text-[0.68rem] text-muted">
        <div>/v1{EP.leadersBattingCareer}</div>
        <div>/v1{EP.leadersPitchingCareer}</div>
      </div>
    </div>
  {/snippet}
</ThreeColLayout>
