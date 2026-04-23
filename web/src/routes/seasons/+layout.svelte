<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl } from '$lib/api';
  import { parseLeague, type LeagueFilter } from '$lib/common/types';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import DatePicker from '$lib/components/DatePicker.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear, STATIC_ERAS, type Era } from '$lib/eras';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { toErrorMessage } from '$lib/leaders/utils';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeSeasons } from '$lib/seasons/normalizers';
  import type { BattingStat, PitchingStat, SeasonSummary } from '$lib/seasons/types';
  import { onMount, type Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();

  const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
  const CURRENT_YEAR = new Date().getFullYear();

  let routeYear = $derived(Number(page.params.year ?? ''));
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawDate = $derived(page.url.searchParams.get('date') ?? '');
  let rawBat = $derived(page.url.searchParams.get('bat'));
  let rawPit = $derived(page.url.searchParams.get('pit'));

  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : CURRENT_YEAR);
  let leagueFilter = $derived(parseLeague(rawLeague));
  let battingStat = $derived(parseBattingStat(rawBat));
  let pitchingStat = $derived(parsePitchingStat(rawPit));
  let activeDate = $derived(DATE_PATTERN.test(rawDate) ? rawDate : `${selectedYear}-04-01`);

  let seasons = $state<SeasonSummary[]>([]);
  let seasonsLoading = $state(false);
  let seasonsError = $state<string | null>(null);
  let seasonsRequestVersion = 0;
  let dateInput = $state('');

  let selectedSeason = $derived(seasons.find((season) => season.year === selectedYear) ?? null);
  let selectedEra = $derived(eraForYear(selectedYear));
  let selectedEraList = $derived.by(() => {
    const era = selectedEra;
    if (!era) return [] as Era[];
    return [era];
  });

  let highlightedSeasons = $derived.by(() => seasons.slice(0, 12));

  let dateEndpoint = $derived(EP.seasonDateGames(selectedYear, activeDate));
  let dateEndpointPath = $derived(`/v1${dateEndpoint}`);
  let dateEndpointUrl = $derived(apiUrl(dateEndpoint));

  onMount(() => {
    void loadSeasons(true);
  });

  afterNavigate(() => {
    void loadSeasons();
  });

  $effect(() => {
    dateInput = activeDate;
  });

  $effect(() => {
    const nextDate = dateInput.trim();
    if (!nextDate) {
      if (rawDate) updateQuery({ date: null });
      return;
    }
    if (!DATE_PATTERN.test(nextDate) || nextDate === activeDate) return;
    updateQuery({ date: nextDate });
  });

  async function loadSeasons(force = false): Promise<void> {
    if (!force && (seasonsLoading || seasons.length > 0)) return;

    const requestVersion = ++seasonsRequestVersion;
    seasonsLoading = true;
    seasonsError = null;

    try {
      const payload = await apiFetch<unknown>(EP.seasons);
      if (requestVersion !== seasonsRequestVersion) return;
      seasons = normalizeSeasons(payload);
    } catch (error) {
      if (requestVersion !== seasonsRequestVersion) return;
      seasonsError = toErrorMessage(error, 'Failed to load season list.');
    } finally {
      if (requestVersion === seasonsRequestVersion) {
        seasonsLoading = false;
      }
    }
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

  function updateQuery(overrides: Record<string, string | number | null>): void {
    const href = withMergedQuery(page.url.pathname, page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as `/seasons/${string}`), QUERY_NAV_OPTS);
  }

  function navigateToYear(nextYear: number): void {
    if (!Number.isFinite(nextYear) || nextYear <= 0) return;
    const parts = page.url.pathname.split('/');
    if (parts.length < 3) return;
    parts[2] = String(nextYear);
    const basePath = parts.join('/');
    const href = withMergedQuery(basePath, page.url.searchParams, { date: null }, page.url.hash);
    void goto(resolve(href as `/seasons/${string}`), QUERY_NAV_OPTS);
  }

  function onYearChange(event: Event): void {
    const nextYear = Number((event.target as HTMLSelectElement).value);
    navigateToYear(nextYear);
  }

  function onLeagueSelect(league: LeagueFilter): void {
    if (league === leagueFilter) return;
    if (league === 'both') {
      updateQuery({ league: null });
      return;
    }
    updateQuery({ league });
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Season hub</div>

    {#if seasonsError}
      <p class="rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
        {seasonsError}
      </p>
    {/if}

    <label class="mt-3 mb-1 block font-mono text-[0.64rem] tracking-[0.08em] text-muted uppercase" for="season-year">
      Season year
    </label>
    <select
      id="season-year"
      value={selectedYear}
      class="w-full rounded border border-outline bg-surface px-2 py-1.5 font-mono text-[0.8rem] text-foreground"
      onchange={onYearChange}
      disabled={seasonsLoading || seasons.length === 0}>
      {#if seasons.length === 0}
        <option value={selectedYear}>{selectedYear}</option>
      {:else}
        {#each seasons as season (season.year)}
          <option value={season.year}>{season.year}</option>
        {/each}
      {/if}
    </select>

    <div class="mt-4 rounded-lg border border-outline bg-surface p-3">
      <div class="mb-1 flex items-center justify-between gap-2">
        <div class="font-display text-[0.95rem] text-foreground">{selectedYear}</div>
        {#if selectedEra}
          <EraBadge era={selectedEra} />
        {/if}
      </div>
      <div class="space-y-1 font-mono text-[0.7rem] text-muted">
        <div>
          Leagues:
          {#if selectedSeason}
            {selectedSeason.leagues.length > 0 ? selectedSeason.leagues.join(' + ') : '—'}
          {:else}
            —
          {/if}
        </div>
        <div>
          Teams:
          {#if selectedSeason?.team_count != null}
            {selectedSeason.team_count}
          {:else}
            —
          {/if}
        </div>
        <div>
          Games:
          {#if selectedSeason?.game_count != null}
            {selectedSeason.game_count.toLocaleString()}
          {:else}
            —
          {/if}
        </div>
      </div>
    </div>

    <div class="mt-4">
      <div class="mb-1 font-mono text-[0.64rem] tracking-[0.08em] text-muted uppercase">League</div>
      <div class="grid grid-cols-3 gap-1 rounded-lg border border-outline bg-surface p-1">
        <button
          class="rounded px-2 py-1 font-display text-[0.78rem] transition-colors {leagueFilter === 'both'
            ? 'bg-crust text-foreground'
            : 'text-muted hover:text-foreground'}"
          onclick={() => onLeagueSelect('both')}>
          Both
        </button>
        <button
          class="rounded px-2 py-1 font-display text-[0.78rem] transition-colors {leagueFilter === 'AL'
            ? 'bg-crust text-foreground'
            : 'text-muted hover:text-foreground'}"
          onclick={() => onLeagueSelect('AL')}>
          AL
        </button>
        <button
          class="rounded px-2 py-1 font-display text-[0.78rem] transition-colors {leagueFilter === 'NL'
            ? 'bg-crust text-foreground'
            : 'text-muted hover:text-foreground'}"
          onclick={() => onLeagueSelect('NL')}>
          NL
        </button>
      </div>
    </div>

    <label class="mt-4 mb-1 block font-mono text-[0.64rem] tracking-[0.08em] text-muted uppercase" for="season-date">
      Calendar date
    </label>
    <DatePicker
      id="season-date"
      bind:value={dateInput}
      min={`${selectedYear}-01-01`}
      max={`${selectedYear}-12-31`}
      placeholder="Jump to date" />

    {#if selectedEra?.caveat}
      <div class="mt-4">
        <EraDisclaimer eras={selectedEraList} />
      </div>
    {/if}

    <div class="mt-4 panel-label">Era quick jump</div>
    <div class="flex flex-wrap gap-1.5">
      {#each STATIC_ERAS as era (era.code)}
        <EraRangeChip {era} year={era.from} />
      {/each}
    </div>

    <div class="mt-4 flex items-center justify-between gap-2">
      <div class="panel-label mb-0" title="Historical seasons only. Use Current year for in-season data.">
        Season list
      </div>
      {#if selectedYear !== CURRENT_YEAR}
        <button
          class="rounded border border-outline px-2 py-1 font-mono text-[0.66rem] text-muted transition-colors hover:border-primary hover:text-foreground"
          onclick={() => navigateToYear(CURRENT_YEAR)}
          title={`Go to ${CURRENT_YEAR}`}>
          Current year
        </button>
      {/if}
    </div>
    <div class="max-h-52 space-y-1 overflow-y-auto pr-1">
      {#each highlightedSeasons as season (season.year)}
        <button
          class="w-full rounded px-2 py-1 text-left font-mono text-[0.72rem] transition-colors {season.year ===
          selectedYear
            ? 'bg-surface text-foreground'
            : 'text-muted hover:bg-surface hover:text-foreground'}"
          title={season.year < CURRENT_YEAR
            ? 'Historical season only. Use Current year for in-season data.'
            : `Season ${season.year}`}
          onclick={() => navigateToYear(season.year)}>
          <span class="inline-block w-12">{season.year}</span>
          <span class="text-muted/85">{season.leagues.length > 0 ? season.leagues.join('+') : '—'}</span>
        </button>
      {/each}
    </div>
  {/snippet}

  {#snippet center()}
    <div class="flex flex-col gap-5">
      {@render children()}
    </div>
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={dateEndpointPath} url={dateEndpointUrl} />

    <div class="mt-4 rounded-lg border border-outline bg-crust p-4">
      <div class="panel-label">Notable queries</div>
      <div class="space-y-2 text-[0.72rem]">
        <div>
          <div class="text-foreground">Season teams</div>
          <div class="font-mono text-muted">/v1{EP.seasonTeams(selectedYear)}</div>
        </div>
        <div>
          <div class="text-foreground">Batting leaders</div>
          <div class="font-mono text-muted">/v1{EP.seasonLeadersBatting(selectedYear)}?stat={battingStat}</div>
        </div>
        <div>
          <div class="text-foreground">Pitching leaders</div>
          <div class="font-mono text-muted">/v1{EP.seasonLeadersPitching(selectedYear)}?stat={pitchingStat}</div>
        </div>
        <div>
          <div class="text-foreground">Schedule</div>
          <div class="font-mono text-muted">/v1{EP.seasonSchedule(selectedYear)}</div>
        </div>
        <div>
          <div class="text-foreground">Date games</div>
          <div class="font-mono text-muted">/v1{EP.seasonDateGames(selectedYear, activeDate)}</div>
        </div>
        <div>
          <div class="text-foreground">Awards</div>
          <div class="font-mono text-muted">/v1{EP.seasonAwards(selectedYear)}</div>
        </div>
        <div>
          <div class="text-foreground">Postseason series</div>
          <div class="font-mono text-muted">/v1{EP.seasonPostseasonSeries(selectedYear)}</div>
        </div>
        <div>
          <div class="text-foreground">Park factors</div>
          <div class="font-mono text-muted">/v1{EP.seasonParkFactors(selectedYear)}</div>
        </div>
      </div>
    </div>
  {/snippet}
</ThreeColLayout>
