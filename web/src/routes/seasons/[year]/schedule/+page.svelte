<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, fetchPaginated, type PaginatedResponse } from '$lib/api';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import { normalizeDateGames, normalizeGamesPage } from '$lib/seasons/normalizers';
  import type { SeasonGame } from '$lib/seasons/types';
  import { onMount } from 'svelte';

  type LeagueFilter = 'both' | 'al' | 'nl';

  type ScheduleLoadResult = {
    page: PaginatedResponse<SeasonGame>;
    loadedCount: number;
    expectedTotal: number;
    truncated: boolean;
  };

  type CalendarCell = { date: string; label: string; inMonth: boolean; gameCount: number; isSelected: boolean };

  const WEEKDAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];
  const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
  const MAX_SCHEDULE_PAGES = 30;

  let routeYear = $derived(Number(page.params.year ?? ''));
  let selectedYear = $derived(Number.isFinite(routeYear) && routeYear > 0 ? routeYear : new Date().getFullYear());
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawDate = $derived(page.url.searchParams.get('date') ?? '');
  let leagueFilter = $derived(parseLeague(rawLeague));

  let schedulePage = $state<PaginatedResponse<SeasonGame>>(emptyPage());
  let scheduleLoading = $state(false);
  let scheduleError = $state<string | null>(null);
  let scheduleLoadedCount = $state(0);
  let scheduleExpectedTotal = $state(0);
  let scheduleTruncated = $state(false);

  let dateGames = $state<SeasonGame[]>([]);
  let dateGamesLoading = $state(false);
  let dateGamesError = $state<string | null>(null);

  let scheduleRequestVersion = 0;
  let dateGamesRequestVersion = 0;
  let lastScheduleKey = '';
  let lastDateGamesKey = '';

  let scheduleCountsByDate = $derived.by(() => {
    const counts: Record<string, number> = {};
    for (const game of schedulePage.data) {
      if (!game.date) continue;
      counts[game.date] = (counts[game.date] ?? 0) + 1;
    }
    return counts;
  });

  const scheduleDates = $derived(Object.keys(scheduleCountsByDate).toSorted());

  let fallbackDate = $derived.by(() => {
    const firstDate = scheduleDates.at(0);
    if (firstDate) return firstDate;
    return `${selectedYear}-04-01`;
  });

  let activeDate = $derived.by(() => {
    if (DATE_PATTERN.test(rawDate)) return rawDate;
    return fallbackDate;
  });

  let activeMonth = $derived.by(() => {
    if (DATE_PATTERN.test(activeDate)) return activeDate.slice(0, 7);
    return `${selectedYear}-04`;
  });

  let monthCells = $derived(buildMonthCells(activeMonth, scheduleCountsByDate, activeDate));

  let maxGamesInMonth = $derived.by(() => {
    const counts = monthCells.filter((cell) => cell.inMonth).map((cell) => cell.gameCount);
    if (counts.length === 0) return 0;
    return Math.max(...counts);
  });

  let monthTitle = $derived(formatMonthLabel(activeMonth));

  let dateGameRows = $derived(
    dateGames
      .map((game) => {
        const away = game.away_team ?? 'Away';
        const home = game.home_team ?? 'Home';
        const score = gameScore(game);
        return {
          id: game.id,
          matchup: `${away} @ ${home}`,
          score,
          innings: game.innings,
          park: game.park_name ?? game.park_id ?? '—'
        };
      })
      .toSorted((a, b) => a.id.localeCompare(b.id))
  );

  const dateGameColumns = [
    {
      key: 'id',
      label: 'Game ID',
      sortable: true,
      href: (value: unknown) => {
        const gameId = String(value ?? '').trim();
        if (!gameId) return;
        const query = new URLSearchParams({ q: gameId }).toString();
        return `/games?${query}`;
      }
    },
    { key: 'matchup', label: 'Matchup', sortable: true },
    { key: 'score', label: 'Score', sortable: true },
    { key: 'innings', label: 'Inn', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'park', label: 'Park', sortable: true }
  ];

  onMount(() => {
    void refreshSchedule(true);
    void refreshDateGames(true);
  });

  afterNavigate(() => {
    void refreshSchedule();
    void refreshDateGames();
  });

  async function refreshSchedule(force = false): Promise<void> {
    const key = `${selectedYear}|${leagueFilter}`;
    if (!force && key === lastScheduleKey) return;
    lastScheduleKey = key;

    const requestVersion = ++scheduleRequestVersion;
    scheduleLoading = true;
    scheduleError = null;

    try {
      const result = await loadSeasonSchedule(selectedYear, leagueFilter);
      if (requestVersion !== scheduleRequestVersion) return;

      schedulePage = result.page;
      scheduleLoadedCount = result.loadedCount;
      scheduleExpectedTotal = result.expectedTotal;
      scheduleTruncated = result.truncated;

      void refreshDateGames(true);
    } catch (error) {
      if (requestVersion !== scheduleRequestVersion) return;

      scheduleError = toErrorMessage(error, 'Failed to load season schedule.');
      schedulePage = emptyPage();
      scheduleLoadedCount = 0;
      scheduleExpectedTotal = 0;
      scheduleTruncated = false;
    } finally {
      if (requestVersion === scheduleRequestVersion) scheduleLoading = false;
    }
  }

  async function refreshDateGames(force = false): Promise<void> {
    const date = activeDate;
    const key = `${selectedYear}|${date}`;
    if (!force && key === lastDateGamesKey) return;
    lastDateGamesKey = key;

    const requestVersion = ++dateGamesRequestVersion;
    dateGamesLoading = true;
    dateGamesError = null;

    try {
      const payload = await apiFetch<unknown>(EP.seasonDateGames(selectedYear, date));
      if (requestVersion !== dateGamesRequestVersion) return;
      dateGames = normalizeDateGames(payload);
    } catch (error) {
      if (requestVersion !== dateGamesRequestVersion) return;
      dateGamesError = toErrorMessage(error, 'Failed to load games for this date.');
      dateGames = [];
    } finally {
      if (requestVersion === dateGamesRequestVersion) dateGamesLoading = false;
    }
  }

  async function loadSeasonSchedule(year: number, league: LeagueFilter): Promise<ScheduleLoadResult> {
    const params: Record<string, string | number> = { page: 1, per_page: 200 };
    if (league !== 'both') params.league = league.toUpperCase();

    const firstPayload = await fetchPaginated<unknown>(EP.seasonSchedule(year), params);
    const firstPage = normalizeGamesPage(firstPayload);

    const dedupByID: Record<string, SeasonGame> = {};
    for (const game of firstPage.data) {
      dedupByID[game.id] = game;
    }

    const expectedTotal = firstPage.total;
    const totalPages = Math.max(1, Math.ceil(expectedTotal / Math.max(1, firstPage.per_page)));
    const maxPages = Math.min(totalPages, MAX_SCHEDULE_PAGES);

    if (maxPages > 1) {
      const extraPromises: Array<Promise<unknown>> = [];
      for (let nextPage = 2; nextPage <= maxPages; nextPage += 1) {
        const nextParams: Record<string, string | number> = { ...params, page: nextPage, per_page: firstPage.per_page };
        extraPromises.push(fetchPaginated<unknown>(EP.seasonSchedule(year), nextParams));
      }

      const extraPayloads = await Promise.all(extraPromises);
      for (const payload of extraPayloads) {
        const pageResult = normalizeGamesPage(payload);
        for (const game of pageResult.data) {
          dedupByID[game.id] = game;
        }
      }
    }

    return {
      page: { data: Object.values(dedupByID), page: 1, per_page: firstPage.per_page, total: expectedTotal },
      loadedCount: Object.keys(dedupByID).length,
      expectedTotal,
      truncated: maxPages < totalPages
    };
  }

  function onCalendarDateSelect(date: string): void {
    if (!DATE_PATTERN.test(date)) return;
    const href = withMergedQuery(page.url.pathname, page.url.searchParams, { date }, page.url.hash);
    void goto(resolve(href as `/seasons/${string}`), QUERY_NAV_OPTS);
  }

  function parseLeague(value: string | null): LeagueFilter {
    if (!value) return 'both';
    const normalized = value.toLowerCase();
    if (normalized === 'al') return 'al';
    if (normalized === 'nl') return 'nl';
    return 'both';
  }

  function emptyPage<T>(): PaginatedResponse<T> {
    return { data: [], page: 1, per_page: 1, total: 0 };
  }

  function toErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof Error && error.message.trim().length > 0) return error.message;
    return fallback;
  }

  function fmtInt(value: number | undefined): string {
    if (value == null) return '—';
    return Math.round(value).toLocaleString();
  }

  function pad2(value: number): string {
    return String(value).padStart(2, '0');
  }

  function buildMonthCells(monthKey: string, counts: Record<string, number>, selectedDate: string): CalendarCell[] {
    const [yearText, monthText] = monthKey.split('-');
    const year = Number(yearText);
    const month = Number(monthText);

    if (!Number.isInteger(year) || !Number.isInteger(month) || month < 1 || month > 12) return [];

    const firstWeekday = new Date(Date.UTC(year, month - 1, 1)).getUTCDay();
    const daysInMonth = new Date(Date.UTC(year, month, 0)).getUTCDate();
    const cells: CalendarCell[] = [];

    for (let index = 0; index < firstWeekday; index += 1) {
      cells.push({ date: '', label: '', inMonth: false, gameCount: 0, isSelected: false });
    }

    for (let day = 1; day <= daysInMonth; day += 1) {
      const date = `${year}-${pad2(month)}-${pad2(day)}`;
      cells.push({
        date,
        label: String(day),
        inMonth: true,
        gameCount: counts[date] ?? 0,
        isSelected: date === selectedDate
      });
    }

    while (cells.length % 7 !== 0) {
      cells.push({ date: '', label: '', inMonth: false, gameCount: 0, isSelected: false });
    }

    return cells;
  }

  function formatMonthLabel(monthKey: string): string {
    const [yearText, monthText] = monthKey.split('-');
    const year = Number(yearText);
    const month = Number(monthText);

    if (!Number.isInteger(year) || !Number.isInteger(month) || month < 1 || month > 12) return monthKey;

    const date = new Date(Date.UTC(year, month - 1, 1));
    return date.toLocaleDateString(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' });
  }

  function calendarCellClass(cell: CalendarCell, maxCount: number): string {
    if (!cell.inMonth) return 'bg-transparent text-transparent';

    if (cell.gameCount <= 0 || maxCount <= 0) {
      if (cell.isSelected) return 'border-primary bg-surface text-foreground';
      return 'border-outline bg-surface text-muted';
    }

    const ratio = cell.gameCount / maxCount;
    let intensity = 'bg-primary/25 text-primary border-primary/35';
    if (ratio >= 0.75) {
      intensity = 'bg-primary/80 text-white border-primary';
    } else if (ratio >= 0.5) {
      intensity = 'bg-primary/60 text-white border-primary/75';
    } else if (ratio >= 0.25) {
      intensity = 'bg-primary/40 text-primary border-primary/55';
    }

    if (cell.isSelected) return `${intensity} ring-1 ring-primary/80`;
    return intensity;
  }

  function gameScore(game: SeasonGame): string {
    if (game.away_score == null || game.home_score == null) return 'TBD';
    return `${game.away_score}-${game.home_score}`;
  }
</script>

<section class="rounded-lg border border-outline bg-crust p-4">
  <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
    <div class="panel-label mb-0 border-0 p-0">Schedule calendar</div>
    <div class="font-mono text-[0.7rem] text-muted">{monthTitle}</div>
  </div>

  {#if scheduleError}
    <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
      {scheduleError}
    </p>
  {/if}

  {#if scheduleLoading}
    <p class="font-mono text-[0.75rem] text-muted">Loading schedule...</p>
  {:else}
    {#if scheduleTruncated}
      <p class="mb-2 font-mono text-[0.68rem] text-warning">
        Showing {scheduleLoadedCount.toLocaleString()} of {scheduleExpectedTotal.toLocaleString()} schedule rows (fetch capped
        at
        {MAX_SCHEDULE_PAGES} pages).
      </p>
    {/if}

    <div class="grid grid-cols-7 gap-1">
      {#each WEEKDAY_LABELS as day, index (`header-${index}`)}
        <div class="px-1 py-0.5 text-center font-mono text-[0.62rem] text-muted uppercase">{day}</div>
      {/each}

      {#each monthCells as cell, index (`cell-${index}`)}
        {#if cell.inMonth}
          <button
            class="h-11 rounded border px-1 py-1 text-left font-mono text-[0.64rem] transition-colors {calendarCellClass(
              cell,
              maxGamesInMonth
            )}"
            title={`${cell.date}: ${cell.gameCount} games`}
            onclick={() => onCalendarDateSelect(cell.date)}>
            <span class="block">{cell.label}</span>
            <span class="block text-[0.58rem] opacity-90">{cell.gameCount}</span>
          </button>
        {:else}
          <div class="h-11 rounded border border-transparent"></div>
        {/if}
      {/each}
    </div>
  {/if}

  <div class="mt-4">
    <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
      <div class="font-mono text-[0.66rem] tracking-[0.08em] text-muted uppercase">Games on {activeDate}</div>
      <div class="font-mono text-[0.68rem] text-muted">{dateGames.length} games</div>
    </div>

    {#if dateGamesError}
      <p class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
        {dateGamesError}
      </p>
    {/if}

    {#if dateGamesLoading}
      <p class="font-mono text-[0.75rem] text-muted">Loading date games...</p>
    {:else if dateGameRows.length === 0}
      <p class="font-mono text-[0.75rem] text-muted">No games were returned for this date.</p>
    {:else}
      <SortableTable columns={dateGameColumns} rows={dateGameRows} />
    {/if}
  </div>
</section>
