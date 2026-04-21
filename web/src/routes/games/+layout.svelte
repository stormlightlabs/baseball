<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiUrl, fetchPaginated, type Params } from '$lib/api';
  import { DEFAULT_GAME_TAB, isGameTabId, type GameTabId } from '$lib/common/constants';
  import { endpointForGameTab } from '$lib/common/tabs';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import Pagination from '$lib/components/Pagination.svelte';
  import DatePicker from '$lib/components/DatePicker.svelte';
  import SearchInput from '$lib/components/SearchInput.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear } from '$lib/eras';
  import { detectSeason, endpointLabelFromFamily, normalizeGamePage, parseDateOnly } from '$lib/games/normalizers';
  import type { GameFamily, GameFinderMode, GameRecord } from '$lib/games/types';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { AsyncPaginatedListResource } from '$lib/players/resources.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery, type QueryOverrides } from '$lib/players/routing';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount, type Snippet } from 'svelte';
  import { SvelteURLSearchParams } from 'svelte/reactivity';

  const DEFAULT_PAGE = 1;
  const DEFAULT_PER_PAGE = 25;

  const DEFAULT_EVENTS_PAGE = 1;
  const DEFAULT_EVENTS_PER_PAGE = 30;

  const DEFAULT_PLAYS_PAGE = 1;
  const DEFAULT_PLAYS_PER_PAGE = 25;

  let { children }: { children: Snippet } = $props();

  let gameId = $derived(page.params.id ?? '');
  let rawTab = $derived(page.url.pathname.split('/')[3] ?? '');
  let activeTab = $derived.by((): GameTabId => {
    if (!gameId) return DEFAULT_GAME_TAB;
    if (isGameTabId(rawTab)) return rawTab;
    return DEFAULT_GAME_TAB;
  });

  let finderMode = $derived(parseFinderMode(page.url.searchParams.get('mode')));
  let family = $derived(parseGameFamily(page.url.searchParams.get('family')));

  let season = $derived(page.url.searchParams.get('season') ?? '');
  let homeTeam = $derived(page.url.searchParams.get('home_team') ?? '');
  let awayTeam = $derived(page.url.searchParams.get('away_team') ?? '');
  let parkId = $derived(page.url.searchParams.get('park_id') ?? '');
  let dateFrom = $derived(page.url.searchParams.get('date_from') ?? '');
  let dateTo = $derived(page.url.searchParams.get('date_to') ?? '');
  let query = $derived(page.url.searchParams.get('q') ?? '');

  let currentPage = $derived(intParam(page.url.searchParams, 'page', DEFAULT_PAGE));
  let perPage = $derived(intParam(page.url.searchParams, 'per_page', DEFAULT_PER_PAGE));

  let eventsPage = $derived(intParam(page.url.searchParams, 'events_page', DEFAULT_EVENTS_PAGE));
  let eventsPerPage = $derived(intParam(page.url.searchParams, 'events_per_page', DEFAULT_EVENTS_PER_PAGE));

  let playsPage = $derived(intParam(page.url.searchParams, 'plays_page', DEFAULT_PLAYS_PAGE));
  let playsPerPage = $derived(intParam(page.url.searchParams, 'plays_per_page', DEFAULT_PLAYS_PER_PAGE));

  let seasonInput = $state('');
  let homeTeamInput = $state('');
  let awayTeamInput = $state('');
  let parkIdInput = $state('');
  let dateFromInput = $state('');
  let dateToInput = $state('');
  let naturalQueryInput = $state('');
  let familyInput = $state<GameFamily>('mlb');
  let perPageInput = $state(DEFAULT_PER_PAGE);

  const finderResource = new AsyncPaginatedListResource<GameRecord>();
  let lastFinderKey = '';

  let finderEndpoint = $derived.by(() => resolveFinderEndpoint(finderMode, family));
  let finderParams = $derived.by(() => buildFinderParams());

  let activeEndpoint = $derived.by(() => {
    if (!gameId) return finderEndpoint;
    return endpointForGameTab(gameId, activeTab);
  });

  let activeUrl = $derived.by(() => {
    if (!gameId) return apiUrl(activeEndpoint, finderParams);

    switch (activeTab) {
      case 'events': {
        return apiUrl(activeEndpoint, { page: eventsPage, per_page: eventsPerPage });
      }
      case 'plays': {
        return apiUrl(activeEndpoint, { page: playsPage, per_page: playsPerPage });
      }
      default: {
        return apiUrl(activeEndpoint);
      }
    }
  });

  let teamCodes = $derived.by(() => {
    const home = finderResource.items.map((game) => game.home_team);
    const away = finderResource.items.map((game) => game.away_team);
    return [...home, ...away]
      .filter((team): team is string => team != null)
      .toSorted((a, b) => a.localeCompare(b))
      .filter((team, index, array) => index === 0 || team !== array[index - 1]);
  });

  onMount(() => {
    syncInputsFromUrl();
    void refreshFinder(true);
  });

  afterNavigate(() => {
    syncInputsFromUrl();
    void refreshFinder();
  });

  function parseFinderMode(raw: string | null): GameFinderMode {
    return raw === 'nl' ? 'nl' : 'filters';
  }

  function parseGameFamily(raw: string | null): GameFamily {
    if (raw === 'fed' || raw === 'nlg') return raw;
    return 'mlb';
  }

  function resolveFinderEndpoint(mode: GameFinderMode, sourceFamily: GameFamily): string {
    if (mode === 'nl') return EP.searchGames;

    switch (sourceFamily) {
      case 'fed': {
        return EP.fedGames;
      }
      case 'nlg': {
        return EP.nlgGames;
      }
      default: {
        return EP.games;
      }
    }
  }

  function buildFinderParams(): Params {
    if (finderMode === 'nl') {
      return { q: query.trim(), page: currentPage, per_page: perPage };
    }

    const params: Params = {
      season: season.trim() || undefined,
      home_team: homeTeam.trim().toUpperCase() || undefined,
      away_team: awayTeam.trim().toUpperCase() || undefined,
      page: currentPage,
      per_page: perPage
    };

    if (family === 'mlb') {
      params.park_id = parkId.trim() || undefined;
      params.date_from = dateFrom.trim() || undefined;
      params.date_to = dateTo.trim() || undefined;
    }

    return params;
  }

  async function refreshFinder(force = false): Promise<void> {
    if (finderMode === 'nl' && !query.trim()) {
      finderResource.clear();
      lastFinderKey = '';
      return;
    }

    const key = JSON.stringify({ endpoint: finderEndpoint, params: finderParams });
    if (!force && key === lastFinderKey) return;
    lastFinderKey = key;

    await finderResource.load(async () => {
      const payload = await fetchPaginated<Record<string, unknown>>(finderEndpoint, finderParams);
      return normalizeGamePage(payload);
    });
  }

  function syncInputsFromUrl(): void {
    familyInput = family;
    naturalQueryInput = query;
    seasonInput = season;
    homeTeamInput = homeTeam;
    awayTeamInput = awayTeam;
    parkIdInput = parkId;
    dateFromInput = dateFrom;
    dateToInput = dateTo;
    perPageInput = perPage;
  }

  function updateGamesQuery(overrides: QueryOverrides): void {
    const href = withMergedQuery(page.url.pathname, page.url.searchParams, overrides, page.url.hash);
    if (href.startsWith('/games/')) {
      void goto(resolve(href as `/games/${string}`), QUERY_NAV_OPTS);
      return;
    }
    void goto(resolve(href as '/games'), QUERY_NAV_OPTS);
  }

  function setFinderMode(nextMode: GameFinderMode): void {
    if (nextMode === finderMode) return;

    if (nextMode === 'nl') {
      updateGamesQuery({
        mode: 'nl',
        page: 1,
        season: null,
        home_team: null,
        away_team: null,
        park_id: null,
        date_from: null,
        date_to: null
      });
      return;
    }

    updateGamesQuery({ mode: 'filters', page: 1, q: null });
  }

  function setFamily(nextFamily: GameFamily): void {
    if (nextFamily === family) return;
    updateGamesQuery({ family: nextFamily, mode: 'filters', page: 1 });
  }

  function applyFilterFinder(): void {
    updateGamesQuery({
      mode: 'filters',
      family: familyInput,
      q: null,
      season: seasonInput.trim() || null,
      home_team: homeTeamInput.trim().toUpperCase() || null,
      away_team: awayTeamInput.trim().toUpperCase() || null,
      park_id: parkIdInput.trim() || null,
      date_from: dateFromInput.trim() || null,
      date_to: dateToInput.trim() || null,
      per_page: perPageInput !== DEFAULT_PER_PAGE ? perPageInput : null,
      page: 1
    });
  }

  function runNaturalSearch(value = naturalQueryInput): void {
    const trimmed = value.trim();
    if (!trimmed) return;
    updateGamesQuery({
      mode: 'nl',
      q: trimmed,
      per_page: perPageInput !== DEFAULT_PER_PAGE ? perPageInput : null,
      page: 1
    });
  }

  function sharedFinderQuerySuffix(): string {
    const entries: Array<[string, string]> = [];

    if (finderMode) entries.push(['mode', finderMode]);
    if (family !== 'mlb') entries.push(['family', family]);

    if (query.trim()) entries.push(['q', query.trim()]);
    if (season.trim()) entries.push(['season', season.trim()]);
    if (homeTeam.trim()) entries.push(['home_team', homeTeam.trim().toUpperCase()]);
    if (awayTeam.trim()) entries.push(['away_team', awayTeam.trim().toUpperCase()]);
    if (parkId.trim()) entries.push(['park_id', parkId.trim()]);
    if (dateFrom.trim()) entries.push(['date_from', dateFrom.trim()]);
    if (dateTo.trim()) entries.push(['date_to', dateTo.trim()]);

    entries.push(['page', String(currentPage)], ['per_page', String(perPage)]);

    const queryString = new SvelteURLSearchParams(entries).toString();
    return queryString ? `?${queryString}` : '';
  }

  function hasFinderRows(): boolean {
    return finderResource.items.length > 0;
  }

  function finderHref(game: GameRecord): string {
    return `/games/${encodeURIComponent(game.id)}/overview${sharedFinderQuerySuffix()}`;
  }
</script>

<ThreeColLayout>
  {#snippet sidebar()}
    <div class="panel-label">Game Finder</div>

    <div class="mb-3 grid grid-cols-2 gap-1.5">
      <button
        class="rounded border px-2 py-1.5 font-display text-xs transition-colors {finderMode === 'filters'
          ? 'border-primary/40 bg-surface text-foreground'
          : 'border-outline text-muted hover:border-primary/40 hover:text-foreground'}"
        onclick={() => setFinderMode('filters')}>
        Filter mode
      </button>
      <button
        class="rounded border px-2 py-1.5 font-display text-xs transition-colors {finderMode === 'nl'
          ? 'border-primary/40 bg-surface text-foreground'
          : 'border-outline text-muted hover:border-primary/40 hover:text-foreground'}"
        onclick={() => setFinderMode('nl')}>
        NL mode
      </button>
    </div>

    {#if finderMode === 'nl'}
      <SearchInput
        mini
        bind:value={naturalQueryInput}
        placeholder="e.g. yankees red sox 2004 alcs game 7"
        onsubmit={runNaturalSearch} />
      <p class="mt-2 font-mono text-xxs text-muted">
        Uses <code>/v1/search/games</code> natural-language parsing.
      </p>
      <div class="mt-2.5">
        <label for="games-per-page-nl" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
          Results per page
        </label>
        <select
          id="games-per-page-nl"
          bind:value={perPageInput}
          onchange={() => runNaturalSearch()}
          class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground">
          {#each [10, 25, 50, 100] as n (n)}
            <option value={n}>{n}</option>
          {/each}
        </select>
      </div>
    {:else}
      <div class="mb-2 grid grid-cols-3 gap-1.5">
        <button
          class="rounded border px-2 py-1.5 font-mono text-xxs transition-colors {family === 'mlb'
            ? 'border-primary/40 bg-surface text-foreground'
            : 'border-outline text-muted hover:border-primary/40 hover:text-foreground'}"
          onclick={() => setFamily('mlb')}>
          MLB
        </button>
        <button
          class="rounded border px-2 py-1.5 font-mono text-xxs transition-colors {family === 'fed'
            ? 'border-primary/40 bg-surface text-foreground'
            : 'border-outline text-muted hover:border-primary/40 hover:text-foreground'}"
          onclick={() => setFamily('fed')}>
          Federal
        </button>
        <button
          class="rounded border px-2 py-1.5 font-mono text-xxs transition-colors {family === 'nlg'
            ? 'border-primary/40 bg-surface text-foreground'
            : 'border-outline text-muted hover:border-primary/40 hover:text-foreground'}"
          onclick={() => setFamily('nlg')}>
          Negro
        </button>
      </div>

      <p class="mb-3 font-mono text-[0.63rem] text-muted">{endpointLabelFromFamily(family)}</p>

      <div class="space-y-2.5">
        <div>
          <label for="games-season" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
            Season
          </label>
          <input
            id="games-season"
            type="number"
            min="1871"
            max="2026"
            bind:value={seasonInput}
            class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground" />
        </div>

        <div>
          <label for="games-home-team" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
            Home team
          </label>
          <input
            id="games-home-team"
            list="games-team-options"
            bind:value={homeTeamInput}
            placeholder="e.g. NYA"
            class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground" />
        </div>

        <div>
          <label for="games-away-team" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
            Away team
          </label>
          <input
            id="games-away-team"
            list="games-team-options"
            bind:value={awayTeamInput}
            placeholder="e.g. BOS"
            class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground" />
        </div>

        {#if family === 'mlb'}
          <div>
            <label for="games-park-id" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
              Park ID
            </label>
            <input
              id="games-park-id"
              bind:value={parkIdInput}
              placeholder="e.g. NYC01"
              class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground" />
          </div>

          <div class="grid grid-cols-1 gap-2">
            <div>
              <label
                for="games-date-from"
                class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
                Date from
              </label>
              <DatePicker id="games-date-from" bind:value={dateFromInput} placeholder="From date" />
            </div>
            <div>
              <label
                for="games-date-to"
                class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
                Date to
              </label>
              <DatePicker id="games-date-to" bind:value={dateToInput} placeholder="To date" />
            </div>
          </div>
        {/if}

        <div>
          <label for="games-per-page" class="mb-1 block font-mono text-[0.64rem] tracking-wider text-muted uppercase">
            Results per page
          </label>
          <select
            id="games-per-page"
            bind:value={perPageInput}
            class="w-full rounded border border-outline bg-surface px-2.5 py-1.5 font-mono text-xs text-foreground">
            {#each [10, 25, 50, 100] as n (n)}
              <option value={n}>{n}</option>
            {/each}
          </select>
        </div>

        <button
          class="w-full rounded border border-primary/40 bg-primary/10 px-2.5 py-1.75 font-display text-[0.78rem] text-foreground transition-colors hover:bg-primary/15"
          onclick={applyFilterFinder}>
          Apply filters
        </button>
      </div>

      <datalist id="games-team-options">
        {#each teamCodes as code (code)}
          <option value={code}></option>
        {/each}
      </datalist>
    {/if}

    {#if finderResource.loading}
      <p class="mt-3 font-mono text-xs text-muted">Loading games…</p>
    {:else if finderResource.error}
      <p class="mt-3 font-mono text-xs text-warning">{finderResource.error}</p>
    {:else if !hasFinderRows()}
      <p class="mt-4 font-mono text-xs text-muted">
        {finderMode === 'nl' ? 'Enter a natural-language game query.' : 'No games match current filters.'}
      </p>
    {:else}
      <div class="mt-4 panel-label">Results</div>
      <div class="flex flex-col gap-0.5">
        {#each finderResource.items as game (game.id)}
          {@const gameYear = detectSeason(game)}
          {@const era = gameYear != null ? eraForYear(gameYear) : undefined}
          <a
            href={resolve(finderHref(game) as `/games/${string}/overview`)}
            class="rounded-md px-3 py-2 text-left no-underline transition-colors hover:bg-surface {gameId === game.id
              ? 'bg-surface'
              : ''}">
            <div class="flex items-center gap-1.5">
              <div class="font-display text-[0.8rem] text-foreground">
                {game.away_team ?? '—'} @ {game.home_team ?? '—'}
              </div>
              {#if era}
                <EraBadge {era} size="xs" />
              {/if}
            </div>
            <div class="font-mono text-[0.68rem] text-muted">
              {parseDateOnly(game.date)} · {game.away_score ?? '?'}-{game.home_score ?? '?'} · {game.id}
            </div>
          </a>
        {/each}
      </div>

      {#if finderResource.total > perPage}
        <div class="mt-3">
          <Pagination
            page={currentPage}
            {perPage}
            total={finderResource.total}
            onPageChange={(nextPage) => updateGamesQuery({ page: nextPage })} />
        </div>
      {/if}
    {/if}
  {/snippet}

  {#snippet center()}
    {@render children()}
  {/snippet}

  {#snippet panel()}
    <ApiPanel endpoint={`/v1${activeEndpoint}`} url={activeUrl} />
  {/snippet}
</ThreeColLayout>
