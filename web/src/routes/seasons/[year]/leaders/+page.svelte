<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, type PaginatedResponse } from '$lib/api';
  import { parseLeague } from '$lib/common/types';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { emptyPage, toErrorMessage } from '$lib/leaders/utils';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeBattingLeadersPage, normalizePitchingLeadersPage } from '$lib/seasons/normalizers';
  import type { BattingStat, PitchingStat, SeasonBattingLeader, SeasonPitchingLeader } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  const BATTING_STAT_OPTIONS = [
    { value: 'hr', label: 'Home Runs' },
    { value: 'avg', label: 'Batting Avg' },
    { value: 'rbi', label: 'RBI' },
    { value: 'sb', label: 'Stolen Bases' },
    { value: 'h', label: 'Hits' },
    { value: 'r', label: 'Runs' }
  ] as const satisfies Array<{ value: BattingStat; label: string }>;

  const PITCHING_STAT_OPTIONS = [
    { value: 'era', label: 'ERA' },
    { value: 'so', label: 'Strikeouts' },
    { value: 'w', label: 'Wins' },
    { value: 'sv', label: 'Saves' },
    { value: 'ip', label: 'Innings Pitched' }
  ] as const satisfies Array<{ value: PitchingStat; label: string }>;

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : new Date().getFullYear());
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawBat = $derived(page.url.searchParams.get('bat'));
  let rawPit = $derived(page.url.searchParams.get('pit'));

  let leagueFilter = $derived(parseLeague(rawLeague));
  let battingStat = $derived(parseBattingStat(rawBat));
  let pitchingStat = $derived(parsePitchingStat(rawPit));

  let battingPage = $state<PaginatedResponse<SeasonBattingLeader>>(emptyPage());
  let pitchingPage = $state<PaginatedResponse<SeasonPitchingLeader>>(emptyPage());
  let battingLoading = $state(false);
  let pitchingLoading = $state(false);
  let battingError = $state<string | null>(null);
  let pitchingError = $state<string | null>(null);

  let leadersRequestVersion = 0;
  let lastLeadersKey = '';

  let battingRows = $derived.by(() => {
    return battingPage.data.map((leader) => ({
      ...leader,
      player_display: leader.player_id,
      stat_value: battingStatValue(leader, battingStat)
    }));
  });

  let pitchingRows = $derived.by(() => {
    return pitchingPage.data.map((leader) => ({
      ...leader,
      player_display: leader.player_id,
      stat_value: pitchingStatValue(leader, pitchingStat)
    }));
  });

  let battingColumns = $derived.by(() => {
    const statLabel = optionLabel(BATTING_STAT_OPTIONS, battingStat);
    return [
      {
        key: 'player_display',
        label: 'Player',
        sortable: true,
        href: (value: unknown) => {
          const playerId = String(value ?? '').trim();
          if (!playerId) return;
          return `/players/${encodeURIComponent(playerId)}/batting`;
        }
      },
      { key: 'team_id', label: 'Team', sortable: true },
      { key: 'league', label: 'Lg', sortable: true },
      {
        key: 'stat_value',
        label: statLabel,
        sortable: true,
        format: (value: unknown) => formatBattingMetric(value as number | undefined, battingStat)
      }
    ];
  });

  let pitchingColumns = $derived.by(() => {
    const statLabel = optionLabel(PITCHING_STAT_OPTIONS, pitchingStat);
    return [
      {
        key: 'player_display',
        label: 'Pitcher',
        sortable: true,
        href: (value: unknown) => {
          const playerId = String(value ?? '').trim();
          if (!playerId) return;
          return `/players/${encodeURIComponent(playerId)}/pitching`;
        }
      },
      { key: 'team_id', label: 'Team', sortable: true },
      { key: 'league', label: 'Lg', sortable: true },
      {
        key: 'stat_value',
        label: statLabel,
        sortable: true,
        format: (value: unknown) => formatPitchingMetric(value as number | undefined, pitchingStat)
      }
    ];
  });

  onMount(() => {
    void refreshLeaders(true);
  });

  afterNavigate(() => {
    void refreshLeaders();
  });

  async function refreshLeaders(force = false): Promise<void> {
    const key = `${selectedYear}|${leagueFilter}|${battingStat}|${pitchingStat}`;
    if (!force && key === lastLeadersKey) return;
    lastLeadersKey = key;

    const requestVersion = ++leadersRequestVersion;
    battingLoading = true;
    battingError = null;
    pitchingLoading = true;
    pitchingError = null;

    const params: Record<string, string | number> = { page: 1, per_page: 10 };
    if (leagueFilter !== 'both') params.league = leagueFilter.toUpperCase();

    const battingPromise = apiFetch<unknown>(EP.seasonLeadersBatting(selectedYear), { ...params, stat: battingStat });
    const pitchingPromise = apiFetch<unknown>(EP.seasonLeadersPitching(selectedYear), {
      ...params,
      stat: pitchingStat
    });

    const [battingResult, pitchingResult] = await Promise.allSettled([battingPromise, pitchingPromise]);

    if (requestVersion !== leadersRequestVersion) return;

    if (battingResult.status === 'fulfilled') {
      battingPage = normalizeBattingLeadersPage(battingResult.value);
      battingError = null;
    } else {
      battingPage = emptyPage();
      battingError = toErrorMessage(battingResult.reason, 'Failed to load batting leaders.');
    }

    if (pitchingResult.status === 'fulfilled') {
      pitchingPage = normalizePitchingLeadersPage(pitchingResult.value);
      pitchingError = null;
    } else {
      pitchingPage = emptyPage();
      pitchingError = toErrorMessage(pitchingResult.reason, 'Failed to load pitching leaders.');
    }

    battingLoading = false;
    pitchingLoading = false;
  }

  function parseBattingStat(value: string | null): BattingStat {
    const normalized = value?.toLowerCase();
    if (normalized === 'hr' || normalized === 'avg' || normalized === 'rbi' || normalized === 'sb') return normalized;
    if (normalized === 'h' || normalized === 'r') return normalized;
    return 'hr';
  }

  function parsePitchingStat(value: string | null): PitchingStat {
    const normalized = value?.toLowerCase();
    if (normalized === 'era' || normalized === 'so' || normalized === 'w') return normalized;
    if (normalized === 'sv' || normalized === 'ip') return normalized;
    return 'era';
  }

  function onBattingStatChange(event: Event): void {
    const stat = parseBattingStat((event.target as HTMLSelectElement).value);
    const href = withMergedQuery(page.url.pathname, page.url.searchParams, { bat: stat }, page.url.hash);
    void goto(resolve(href as `/seasons/${string}`), QUERY_NAV_OPTS);
  }

  function onPitchingStatChange(event: Event): void {
    const stat = parsePitchingStat((event.target as HTMLSelectElement).value);
    const href = withMergedQuery(page.url.pathname, page.url.searchParams, { pit: stat }, page.url.hash);
    void goto(resolve(href as `/seasons/${string}`), QUERY_NAV_OPTS);
  }

  function battingStatValue(leader: SeasonBattingLeader, stat: BattingStat): number | undefined {
    if (stat === 'hr') return leader.hr;
    if (stat === 'avg') return leader.avg;
    if (stat === 'rbi') return leader.rbi;
    if (stat === 'sb') return leader.sb;
    if (stat === 'h') return leader.h;
    return leader.r;
  }

  function pitchingStatValue(leader: SeasonPitchingLeader, stat: PitchingStat): number | undefined {
    if (stat === 'era') return leader.era;
    if (stat === 'so') return leader.so;
    if (stat === 'w') return leader.w;
    if (stat === 'sv') return leader.sv;
    if (leader.ip_outs == null) return undefined;
    return leader.ip_outs / 3;
  }

  function formatBattingMetric(value: number | undefined, stat: BattingStat): string {
    if (value == null) return '—';
    if (stat === 'avg') return value.toFixed(3);
    return Math.round(value).toLocaleString();
  }

  function formatPitchingMetric(value: number | undefined, stat: PitchingStat): string {
    if (value == null) return '—';
    if (stat === 'era') return value.toFixed(2);
    if (stat === 'ip') return value.toFixed(1);
    return Math.round(value).toLocaleString();
  }

  function optionLabel<T extends string>(options: ReadonlyArray<{ value: T; label: string }>, value: T): string {
    const match = options.find((option) => option.value === value);
    if (match) return match.label;
    return value;
  }
</script>

<section class="grid gap-5 xl:grid-cols-2">
  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="mb-3 flex items-center gap-2">
      <div class="panel-label mb-0 border-0 p-0">Batting leaders</div>
      <select
        value={battingStat}
        class="ml-auto rounded border border-outline bg-surface px-2 py-1 font-mono text-[0.72rem] text-muted"
        onchange={onBattingStatChange}>
        {#each BATTING_STAT_OPTIONS as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>

    {#if battingError}
      <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
        {battingError}
      </p>
    {/if}

    {#if battingLoading}
      <p class="font-mono text-xs text-muted">Loading batting leaders...</p>
    {:else if battingRows.length === 0}
      <p class="font-mono text-xs text-muted">No batting leader rows found.</p>
    {:else}
      <SortableTable columns={battingColumns} rows={battingRows} />
    {/if}
  </div>

  <div class="rounded-lg border border-outline bg-crust p-4">
    <div class="mb-3 flex items-center gap-2">
      <div class="panel-label mb-0 border-0 p-0">Pitching leaders</div>
      <select
        value={pitchingStat}
        class="ml-auto rounded border border-outline bg-surface px-2 py-1 font-mono text-[0.72rem] text-muted"
        onchange={onPitchingStatChange}>
        {#each PITCHING_STAT_OPTIONS as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>

    {#if pitchingError}
      <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
        {pitchingError}
      </p>
    {/if}

    {#if pitchingLoading}
      <p class="font-mono text-xs text-muted">Loading pitching leaders...</p>
    {:else if pitchingRows.length === 0}
      <p class="font-mono text-xs text-muted">No pitching leader rows found.</p>
    {:else}
      <SortableTable columns={pitchingColumns} rows={pitchingRows} />
    {/if}
  </div>
</section>
