<script lang="ts">
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import { apiFetch, apiUrl, fetchPaginated, type PaginatedResponse } from '$lib/api';
  import ApiPanel from '$lib/components/ApiPanel.svelte';
  import EraBadge from '$lib/components/EraBadge.svelte';
  import EraDisclaimer from '$lib/components/EraDisclaimer.svelte';
  import EraRangeChip from '$lib/components/EraRangeChip.svelte';
  import SortableTable from '$lib/components/SortableTable.svelte';
  import { EP } from '$lib/endpoints';
  import { eraForYear, STATIC_ERAS, type Era } from '$lib/eras';
  import ThreeColLayout from '$lib/layouts/ThreeColLayout.svelte';
  import { QUERY_NAV_OPTS, withMergedQuery } from '$lib/players/routing';
  import {
    normalizeAwardsPage,
    normalizeBattingLeadersPage,
    normalizeDateGames,
    normalizeGamesPage,
    normalizeParkFactors,
    normalizePitchingLeadersPage,
    normalizePostseasonSeries,
    normalizeSeasons,
    normalizeSeasonTeamsPage
  } from '$lib/seasons/normalizers';
  import type {
    SeasonAwardResult,
    SeasonBattingLeader,
    SeasonGame,
    SeasonParkFactor,
    SeasonPitchingLeader,
    SeasonPostseasonSeries,
    SeasonSummary,
    SeasonTeam
  } from '$lib/seasons/types';
  import { intParam } from '$lib/url-state.svelte';
  import { onMount } from 'svelte';

  type LeagueFilter = 'both' | 'al' | 'nl';
  type BattingStat = 'hr' | 'avg' | 'rbi' | 'sb' | 'h' | 'r';
  type PitchingStat = 'era' | 'so' | 'w' | 'sv' | 'ip';

  type ScheduleLoadResult = {
    page: PaginatedResponse<SeasonGame>;
    loadedCount: number;
    expectedTotal: number;
    truncated: boolean;
  };

  type CalendarCell = { date: string; label: string; inMonth: boolean; gameCount: number; isSelected: boolean };

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

  const WEEKDAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];
  const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
  const MAX_SCHEDULE_PAGES = 30;

  let queryYear = $derived(intParam(page.url.searchParams, 'year', 0));
  let rawLeague = $derived(page.url.searchParams.get('league'));
  let rawDate = $derived(page.url.searchParams.get('date') ?? '');
  let rawBat = $derived(page.url.searchParams.get('bat'));
  let rawPit = $derived(page.url.searchParams.get('pit'));

  let leagueFilter = $derived(parseLeague(rawLeague));
  let battingStat = $derived(parseBattingStat(rawBat));
  let pitchingStat = $derived(parsePitchingStat(rawPit));

  let seasons = $state<SeasonSummary[]>([]);
  let seasonsLoading = $state(false);
  let seasonsError = $state<string | null>(null);

  let teamsPage = $state<PaginatedResponse<SeasonTeam>>(emptyPage());
  let teamsLoading = $state(false);
  let teamsError = $state<string | null>(null);

  let battingPage = $state<PaginatedResponse<SeasonBattingLeader>>(emptyPage());
  let battingLoading = $state(false);
  let battingError = $state<string | null>(null);

  let pitchingPage = $state<PaginatedResponse<SeasonPitchingLeader>>(emptyPage());
  let pitchingLoading = $state(false);
  let pitchingError = $state<string | null>(null);

  let schedulePage = $state<PaginatedResponse<SeasonGame>>(emptyPage());
  let scheduleLoading = $state(false);
  let scheduleError = $state<string | null>(null);
  let scheduleLoadedCount = $state(0);
  let scheduleExpectedTotal = $state(0);
  let scheduleTruncated = $state(false);

  let dateGames = $state<SeasonGame[]>([]);
  let dateGamesLoading = $state(false);
  let dateGamesError = $state<string | null>(null);

  let awardsPage = $state<PaginatedResponse<SeasonAwardResult>>(emptyPage());
  let awardsLoading = $state(false);
  let awardsError = $state<string | null>(null);

  let postseasonSeries = $state<SeasonPostseasonSeries[]>([]);
  let postseasonGamesPage = $state<PaginatedResponse<SeasonGame>>(emptyPage());
  let postseasonLoading = $state(false);
  let postseasonError = $state<string | null>(null);

  let parkFactors = $state<SeasonParkFactor[]>([]);
  let parkFactorsLoading = $state(false);
  let parkFactorsError = $state<string | null>(null);

  let seasonsRequestVersion = 0;
  let teamsRequestVersion = 0;
  let leadersRequestVersion = 0;
  let scheduleRequestVersion = 0;
  let dateGamesRequestVersion = 0;
  let awardsRequestVersion = 0;
  let postseasonRequestVersion = 0;
  let parkFactorsRequestVersion = 0;

  let lastTeamsKey = '';
  let lastLeadersKey = '';
  let lastScheduleKey = '';
  let lastDateGamesKey = '';
  let lastAwardsKey = '';
  let lastPostseasonKey = '';
  let lastParkFactorsKey = '';

  let latestAvailableYear = $derived.by(() => {
    if (seasons.length === 0) return new Date().getFullYear();
    return Math.max(...seasons.map((season) => season.year));
  });

  const selectedYear = $derived(queryYear > 0 ? queryYear : latestAvailableYear);
  const selectedSeason = $derived(seasons.find((season) => season.year === selectedYear) ?? null);
  const selectedEra = $derived(eraForYear(selectedYear));

  const filteredTeams = $derived.by(() => {
    if (leagueFilter === 'both') return teamsPage.data;
    const league = leagueFilter.toUpperCase();
    return teamsPage.data.filter((team) => (team.league ?? '').toUpperCase() === league);
  });

  const teamRows = $derived(
    filteredTeams
      .map((team) => {
        const pct = winningPct(team.wins, team.losses);
        const runDiff =
          team.runs_scored != null && team.runs_allowed != null ? team.runs_scored - team.runs_allowed : undefined;

        return {
          team_id: team.team_id,
          name: team.name ?? team.team_id,
          league: team.league ?? '—',
          division: team.division ?? '—',
          wins: team.wins,
          losses: team.losses,
          ties: team.ties,
          pct,
          runs_scored: team.runs_scored,
          runs_allowed: team.runs_allowed,
          run_diff: runDiff,
          games: team.games
        };
      })
      .toSorted((a, b) => {
        const pctA = a.pct ?? -1;
        const pctB = b.pct ?? -1;
        if (pctA !== pctB) return pctB - pctA;

        const winsA = a.wins ?? -1;
        const winsB = b.wins ?? -1;
        if (winsA !== winsB) return winsB - winsA;

        return a.team_id.localeCompare(b.team_id);
      })
  );

  const teamColumns = [
    {
      key: 'team_id',
      label: 'Team',
      sortable: true,
      href: (value: unknown) => {
        const teamId = String(value ?? '').trim();
        if (!teamId) return;
        const query = new URLSearchParams({ q: teamId, year: String(selectedYear) }).toString();
        return `/teams?${query}`;
      },
      tooltip: (_value: unknown, row: Record<string, unknown>) => {
        const teamName = String(row.name ?? '').trim();
        return teamName.length > 0 ? teamName : undefined;
      }
    },
    { key: 'league', label: 'Lg', sortable: true },
    { key: 'division', label: 'Div', sortable: true },
    { key: 'wins', label: 'W', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'losses', label: 'L', sortable: true, format: (value: unknown) => fmtInt(value as number | undefined) },
    { key: 'pct', label: 'Pct', sortable: true, format: (value: unknown) => fmtRate(value as number | undefined) },
    {
      key: 'runs_scored',
      label: 'RS',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    {
      key: 'runs_allowed',
      label: 'RA',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    {
      key: 'run_diff',
      label: 'Diff',
      sortable: true,
      format: (value: unknown) => fmtSigned(value as number | undefined)
    }
  ];

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

  let parkFactorRows = $derived.by(() => {
    return parkFactors
      .map((factor) => ({
        park_id: factor.park_id ?? '—',
        runs_factor: factor.runs_factor,
        hr_factor: factor.hr_factor,
        games_sampled: factor.games_sampled,
        provider: factor.provider ?? '—'
      }))
      .toSorted((a, b) => {
        const hrA = a.hr_factor ?? -1;
        const hrB = b.hr_factor ?? -1;
        if (hrA !== hrB) return hrB - hrA;
        return a.park_id.localeCompare(b.park_id);
      });
  });

  const parkFactorColumns = [
    { key: 'park_id', label: 'Park', sortable: true },
    {
      key: 'runs_factor',
      label: 'Runs',
      sortable: true,
      format: (value: unknown) => fmtFloat(value as number | undefined, 1)
    },
    {
      key: 'hr_factor',
      label: 'HR',
      sortable: true,
      format: (value: unknown) => fmtFloat(value as number | undefined, 1)
    },
    {
      key: 'games_sampled',
      label: 'Games',
      sortable: true,
      format: (value: unknown) => fmtInt(value as number | undefined)
    },
    { key: 'provider', label: 'Provider', sortable: true }
  ];

  let selectedEraList = $derived.by(() => {
    const era = selectedEra;
    if (!era) return [] as Era[];
    return [era];
  });

  let highlightedSeasons = $derived.by(() => {
    return seasons.slice(0, 12);
  });

  let dateEndpoint = $derived(EP.seasonDateGames(selectedYear, activeDate));
  let dateEndpointPath = $derived(`/v1${dateEndpoint}`);
  let dateEndpointUrl = $derived(apiUrl(dateEndpoint));

  onMount(() => {
    void loadSeasons(true);
    void refreshTeams(true);
    void refreshLeaders(true);
    void refreshSchedule(true);
    void refreshAwards(true);
    void refreshPostseason(true);
    void refreshParkFactors(true);
    void refreshDateGames(true);
  });

  afterNavigate(() => {
    void loadSeasons();
    void refreshTeams();
    void refreshLeaders();
    void refreshSchedule();
    void refreshAwards();
    void refreshPostseason();
    void refreshParkFactors();
    void refreshDateGames();
  });

  async function loadSeasons(force = false): Promise<void> {
    if (!force && (seasons.length > 0 || seasonsLoading)) return;

    const requestVersion = ++seasonsRequestVersion;
    seasonsLoading = true;
    seasonsError = null;

    try {
      const payload = await apiFetch<unknown>(EP.seasons);
      if (requestVersion !== seasonsRequestVersion) return;
      seasons = normalizeSeasons(payload);

      if (queryYear <= 0) {
        void refreshTeams(true);
        void refreshLeaders(true);
        void refreshSchedule(true);
        void refreshAwards(true);
        void refreshPostseason(true);
        void refreshParkFactors(true);
      }
    } catch (error) {
      if (requestVersion !== seasonsRequestVersion) return;
      seasonsError = toErrorMessage(error, 'Failed to load season list.');
    } finally {
      if (requestVersion === seasonsRequestVersion) {
        seasonsLoading = false;
      }
    }
  }

  async function refreshTeams(force = false): Promise<void> {
    const key = `${selectedYear}|${leagueFilter}`;
    if (!force && key === lastTeamsKey) return;
    lastTeamsKey = key;

    const requestVersion = ++teamsRequestVersion;
    teamsLoading = true;
    teamsError = null;

    try {
      const params: Record<string, string | number> = { page: 1, per_page: 80 };
      if (leagueFilter !== 'both') params.league = leagueFilter.toUpperCase();

      const payload = await fetchPaginated<unknown>(EP.seasonTeams(selectedYear), params);
      if (requestVersion !== teamsRequestVersion) return;
      teamsPage = normalizeSeasonTeamsPage(payload);
    } catch (error) {
      if (requestVersion !== teamsRequestVersion) return;
      teamsError = toErrorMessage(error, 'Failed to load season teams.');
      teamsPage = emptyPage();
    } finally {
      if (requestVersion === teamsRequestVersion) {
        teamsLoading = false;
      }
    }
  }

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
      if (requestVersion === scheduleRequestVersion) {
        scheduleLoading = false;
      }
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
      if (requestVersion === dateGamesRequestVersion) {
        dateGamesLoading = false;
      }
    }
  }

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
      if (requestVersion === awardsRequestVersion) {
        awardsLoading = false;
      }
    }
  }

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
      if (postseasonError == null) postseasonError = null;
    } else {
      postseasonGamesPage = emptyPage();
      const message = toErrorMessage(gamesResult.reason, 'Failed to load postseason games.');
      if (postseasonError) {
        postseasonError = `${postseasonError} ${message}`;
      } else {
        postseasonError = message;
      }
    }

    postseasonLoading = false;
  }

  async function refreshParkFactors(force = false): Promise<void> {
    const key = String(selectedYear);
    if (!force && key === lastParkFactorsKey) return;
    lastParkFactorsKey = key;

    const requestVersion = ++parkFactorsRequestVersion;
    parkFactorsLoading = true;
    parkFactorsError = null;

    try {
      const payload = await apiFetch<unknown>(EP.seasonParkFactors(selectedYear));
      if (requestVersion !== parkFactorsRequestVersion) return;
      parkFactors = normalizeParkFactors(payload);
    } catch (error) {
      if (requestVersion !== parkFactorsRequestVersion) return;
      parkFactorsError = toErrorMessage(error, 'Failed to load park factors.');
      parkFactors = [];
    } finally {
      if (requestVersion === parkFactorsRequestVersion) {
        parkFactorsLoading = false;
      }
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
        const page = normalizeGamesPage(payload);
        for (const game of page.data) {
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

  function parseLeague(value: string | null): LeagueFilter {
    if (!value) return 'both';
    const normalized = value.toLowerCase();
    if (normalized === 'al') return 'al';
    if (normalized === 'nl') return 'nl';
    return 'both';
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
    const href = withMergedQuery('/seasons', page.url.searchParams, overrides, page.url.hash);
    void goto(resolve(href as '/seasons'), QUERY_NAV_OPTS);
  }

  function onYearChange(event: Event): void {
    const nextYear = Number((event.target as HTMLSelectElement).value);
    if (!Number.isFinite(nextYear)) return;
    updateQuery({ year: nextYear, date: null });
  }

  function onLeagueSelect(league: LeagueFilter): void {
    if (league === leagueFilter) return;
    if (league === 'both') {
      updateQuery({ league: null });
      return;
    }
    updateQuery({ league });
  }

  function onDateInput(event: Event): void {
    const next = (event.target as HTMLInputElement).value;
    if (!DATE_PATTERN.test(next)) return;
    updateQuery({ date: next });
  }

  function onBattingStatChange(event: Event): void {
    const stat = parseBattingStat((event.target as HTMLSelectElement).value);
    updateQuery({ bat: stat });
  }

  function onPitchingStatChange(event: Event): void {
    const stat = parsePitchingStat((event.target as HTMLSelectElement).value);
    updateQuery({ pit: stat });
  }

  function onCalendarDateSelect(date: string): void {
    if (!DATE_PATTERN.test(date)) return;
    updateQuery({ date });
  }

  function emptyPage<T>(): PaginatedResponse<T> {
    return { data: [], page: 1, per_page: 1, total: 0 };
  }

  function toErrorMessage(error: unknown, fallback: string): string {
    if (error instanceof Error && error.message.trim().length > 0) return error.message;
    return fallback;
  }

  function winningPct(wins: number | undefined, losses: number | undefined): number | undefined {
    if (wins == null || losses == null) return undefined;
    const decisions = wins + losses;
    if (decisions <= 0) return undefined;
    return wins / decisions;
  }

  function fmtInt(value: number | undefined): string {
    if (value == null) return '—';
    return Math.round(value).toLocaleString();
  }

  function fmtFloat(value: number | undefined, digits = 2): string {
    if (value == null) return '—';
    return value.toFixed(digits);
  }

  function fmtRate(value: number | undefined): string {
    if (value == null) return '—';
    return value.toFixed(3);
  }

  function fmtSigned(value: number | undefined): string {
    if (value == null) return '—';
    if (value > 0) return `+${Math.round(value)}`;
    return String(Math.round(value));
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

  function pad2(value: number): string {
    return String(value).padStart(2, '0');
  }

  function buildMonthCells(monthKey: string, counts: Record<string, number>, selectedDate: string): CalendarCell[] {
    const [yearText, monthText] = monthKey.split('-');
    const year = Number(yearText);
    const month = Number(monthText);

    if (!Number.isInteger(year) || !Number.isInteger(month) || month < 1 || month > 12) {
      return [];
    }

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

    if (!Number.isInteger(year) || !Number.isInteger(month) || month < 1 || month > 12) {
      return monthKey;
    }

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
          class="rounded px-2 py-1 font-display text-[0.78rem] transition-colors {leagueFilter === 'al'
            ? 'bg-crust text-foreground'
            : 'text-muted hover:text-foreground'}"
          onclick={() => onLeagueSelect('al')}>
          AL
        </button>
        <button
          class="rounded px-2 py-1 font-display text-[0.78rem] transition-colors {leagueFilter === 'nl'
            ? 'bg-crust text-foreground'
            : 'text-muted hover:text-foreground'}"
          onclick={() => onLeagueSelect('nl')}>
          NL
        </button>
      </div>
    </div>

    <label class="mt-4 mb-1 block font-mono text-[0.64rem] tracking-[0.08em] text-muted uppercase" for="season-date">
      Calendar date
    </label>
    <input
      id="season-date"
      type="date"
      value={activeDate}
      class="w-full rounded border border-outline bg-surface px-2 py-1.5 font-mono text-[0.78rem] text-foreground"
      onchange={onDateInput} />

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

    <div class="mt-4 panel-label">Season list</div>
    <div class="max-h-52 space-y-1 overflow-y-auto pr-1">
      {#each highlightedSeasons as season (season.year)}
        <button
          class="w-full rounded px-2 py-1 text-left font-mono text-[0.72rem] transition-colors {season.year ===
          selectedYear
            ? 'bg-surface text-foreground'
            : 'text-muted hover:bg-surface hover:text-foreground'}"
          onclick={() => updateQuery({ year: season.year, date: null })}>
          <span class="inline-block w-12">{season.year}</span>
          <span class="text-muted/85">{season.leagues.length > 0 ? season.leagues.join('+') : '—'}</span>
        </button>
      {/each}
    </div>
  {/snippet}

  {#snippet center()}
    <div class="flex flex-col gap-5">
      <div class="rounded-lg border border-outline bg-crust p-4">
        <div class="flex flex-wrap items-center gap-3">
          <div class="font-display text-[1.05rem] text-foreground">Season {selectedYear}</div>
          {#if selectedEra}
            <EraBadge era={selectedEra} size="xs" />
          {/if}
          <div class="font-mono text-[0.7rem] text-muted">
            {#if selectedSeason}
              {selectedSeason.leagues.length > 0 ? selectedSeason.leagues.join(' + ') : '—'}
              · {selectedSeason.team_count ?? '—'} teams · {selectedSeason.game_count != null
                ? selectedSeason.game_count.toLocaleString()
                : '—'} games
            {:else}
              Coverage metadata unavailable for this year.
            {/if}
          </div>
        </div>
      </div>

      <section class="rounded-lg border border-outline bg-crust p-4">
        <div class="mb-3 flex items-center justify-between gap-2">
          <div class="panel-label mb-0 border-0 p-0">Teams</div>
          <div class="font-mono text-xxs tracking-[0.08em] text-muted uppercase">
            {leagueFilter === 'both' ? 'AL + NL' : leagueFilter.toUpperCase()}
          </div>
        </div>

        {#if teamsError}
          <p
            class="mb-3 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
            {teamsError}
          </p>
        {/if}

        {#if teamsLoading}
          <p class="font-mono text-[0.75rem] text-muted">Loading season teams...</p>
        {:else if teamRows.length === 0}
          <p class="font-mono text-[0.75rem] text-muted">No team rows for this season and league filter.</p>
        {:else}
          <SortableTable columns={teamColumns} rows={teamRows} />
        {/if}
      </section>

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
            <p
              class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
              {battingError}
            </p>
          {/if}

          {#if battingLoading}
            <p class="font-mono text-[0.75rem] text-muted">Loading batting leaders...</p>
          {:else if battingRows.length === 0}
            <p class="font-mono text-[0.75rem] text-muted">No batting leader rows found.</p>
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
            <p
              class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
              {pitchingError}
            </p>
          {/if}

          {#if pitchingLoading}
            <p class="font-mono text-[0.75rem] text-muted">Loading pitching leaders...</p>
          {:else if pitchingRows.length === 0}
            <p class="font-mono text-[0.75rem] text-muted">No pitching leader rows found.</p>
          {:else}
            <SortableTable columns={pitchingColumns} rows={pitchingRows} />
          {/if}
        </div>
      </section>

      <section class="rounded-lg border border-outline bg-crust p-4">
        <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
          <div class="panel-label mb-0 border-0 p-0">Schedule calendar</div>
          <div class="font-mono text-[0.7rem] text-muted">{monthTitle}</div>
        </div>

        {#if scheduleError}
          <p
            class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
            {scheduleError}
          </p>
        {/if}

        {#if scheduleLoading}
          <p class="font-mono text-[0.75rem] text-muted">Loading schedule...</p>
        {:else}
          {#if scheduleTruncated}
            <p class="mb-2 font-mono text-[0.68rem] text-warning">
              Showing {scheduleLoadedCount.toLocaleString()} of {scheduleExpectedTotal.toLocaleString()} schedule rows (fetch
              capped at
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
            <p
              class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
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

      <section class="grid gap-5 xl:grid-cols-2">
        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-3">Awards snapshot</div>

          {#if awardsError}
            <p
              class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
              {awardsError}
            </p>
          {/if}

          {#if awardsLoading}
            <p class="font-mono text-[0.75rem] text-muted">Loading season awards...</p>
          {:else if awardsRows.length === 0}
            <p class="font-mono text-[0.75rem] text-muted">No awards were returned for this season.</p>
          {:else}
            <SortableTable columns={awardsColumns} rows={awardsRows} />
          {/if}
        </div>

        <div class="rounded-lg border border-outline bg-crust p-4">
          <div class="panel-label mb-2">Postseason modules</div>

          {#if postseasonError}
            <p
              class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
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
        </div>
      </section>

      <section class="rounded-lg border border-outline bg-crust p-4">
        <div class="panel-label mb-3">Park factors snapshot</div>

        {#if parkFactorsError}
          <p
            class="mb-2 rounded border border-warning/30 bg-warning/10 px-3 py-2 font-mono text-[0.72rem] text-warning">
            {parkFactorsError}
          </p>
        {/if}

        {#if parkFactorsLoading}
          <p class="font-mono text-[0.75rem] text-muted">Loading park factors...</p>
        {:else if parkFactorRows.length === 0}
          <p class="font-mono text-[0.75rem] text-muted">No park factor rows were returned for this season.</p>
        {:else}
          <SortableTable columns={parkFactorColumns} rows={parkFactorRows} />
        {/if}
      </section>
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
