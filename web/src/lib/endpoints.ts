/**
 * Central endpoint map
 *
 * Single source of truth for all /v1/* API paths and their canonical query parameters.
 * Derived from internal/docs/swagger.yaml.
 *
 * New endpoints should be added here rather than hardcoding strings in page components.
 */

export const EP = {
  meta: '/meta' as const,
  metaDatasets: '/meta/datasets' as const,
  metaCrosswalkTeams: '/meta/crosswalk/teams' as const,
  metaCrosswalkPlayers: '/meta/crosswalk/players' as const,
  metaCrosswalkTeamByTeam: (teamId: string) => `/meta/crosswalk/teams/by-team/${teamId}` as const,
  metaCrosswalkTeamByFranchise: (franchiseId: string) => `/meta/crosswalk/teams/by-franchise/${franchiseId}` as const,
  metaCrosswalkTeamByMLBAM: (mlbamTeamId: number | string) => `/meta/crosswalk/teams/by-mlbam/${mlbamTeamId}` as const,
  metaCrosswalkPlayerByPlayer: (playerId: string) => `/meta/crosswalk/players/by-player/${playerId}` as const,
  metaCrosswalkPlayerByRetro: (retroId: string) => `/meta/crosswalk/players/by-retro/${retroId}` as const,
  metaCrosswalkPlayerByMLBAM: (mlbamId: number | string) => `/meta/crosswalk/players/by-mlbam/${mlbamId}` as const,
  health: '/health' as const,
  ready: '/ready' as const,
  metaReadiness: '/meta/readiness' as const,

  searchPlayers: '/search/players' as const,
  searchTeams: '/search/teams' as const,
  searchGames: '/search/games' as const,
  searchParks: '/search/parks' as const,

  players: '/players' as const,
  player: (id: string) => `/players/${id}` as const,
  playerSeasons: (id: string) => `/players/${id}/seasons` as const,
  playerStatsBatting: (id: string) => `/players/${id}/stats/batting` as const,
  playerStatsPitching: (id: string) => `/players/${id}/stats/pitching` as const,
  playerStatsWar: (id: string) => `/players/${id}/stats/war` as const,
  playerStatsBattingAdv: (id: string) => `/players/${id}/stats/batting/advanced` as const,
  playerStatsPitchingAdv: (id: string) => `/players/${id}/stats/pitching/advanced` as const,
  playerGameLogsBatting: (id: string) => `/players/${id}/game-logs/batting` as const,
  playerGameLogsPitching: (id: string) => `/players/${id}/game-logs/pitching` as const,
  playerGameLogsFielding: (id: string) => `/players/${id}/game-logs/fielding` as const,
  playerAwards: (id: string) => `/players/${id}/awards` as const,
  playerHallOfFame: (id: string) => `/players/${id}/hall-of-fame` as const,
  playerTeams: (id: string) => `/players/${id}/teams` as const,
  playerSalaries: (id: string) => `/players/${id}/salaries` as const,
  playerRelatives: (id: string) => `/players/${id}/relatives` as const,
  playerSplits: (id: string) => `/players/${id}/splits` as const,
  playerStreaks: (id: string) => `/players/${id}/streaks` as const,

  franchises: '/franchises' as const,
  franchise: (id: string) => `/franchises/${id}` as const,
  teams: '/teams' as const,
  team: (id: string) => `/teams/${id}` as const,
  internalWebTeamYearRange: (id: string) => `/internal/web/teams/${id}/year-range` as const,
  teamDailyStats: (id: string) => `/teams/${id}/daily-stats` as const,
  teamRunDifferential: (teamId: string) => `/teams/${teamId}/run-differential` as const,

  seasons: '/seasons' as const,
  seasonTeams: (year: number | string) => `/seasons/${year}/teams` as const,
  seasonTeamRoster: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/roster` as const,
  seasonTeamBatting: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/batting` as const,
  seasonTeamPitching: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/pitching` as const,
  seasonTeamFielding: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/fielding` as const,
  seasonTeamGames: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/games` as const,
  seasonTeamSchedule: (year: number | string, teamId: string) => `/seasons/${year}/teams/${teamId}/schedule` as const,
  seasonTeamDailyLogs: (year: number | string, teamId: string) =>
    `/seasons/${year}/teams/${teamId}/daily-logs` as const,
  seasonSchedule: (year: number | string) => `/seasons/${year}/schedule` as const,
  seasonDateGames: (year: number | string, date: string) => `/seasons/${year}/dates/${date}/games` as const,
  seasonParkGames: (year: number | string, parkId: string) => `/seasons/${year}/parks/${parkId}/games` as const,
  seasonAwards: (year: number | string) => `/seasons/${year}/awards` as const,
  seasonPostseasonSeries: (year: number | string) => `/seasons/${year}/postseason/series` as const,
  seasonPostseasonGames: (year: number | string) => `/seasons/${year}/postseason/games` as const,
  seasonParkFactors: (year: number | string) => `/seasons/${year}/park-factors` as const,
  seasonLeadersBatting: (year: number | string) => `/seasons/${year}/leaders/batting` as const,
  seasonLeadersPitching: (year: number | string) => `/seasons/${year}/leaders/pitching` as const,
  seasonLeadersBattingAdv: (season: number | string) => `/seasons/${season}/leaders/batting/advanced` as const,
  seasonLeadersPitchingAdv: (season: number | string) => `/seasons/${season}/leaders/pitching/advanced` as const,
  seasonLeadersWar: (season: number | string) => `/seasons/${season}/leaders/war` as const,
  standings: '/standings' as const,

  games: '/games' as const,
  game: (id: string) => `/games/${id}` as const,
  gameSummary: (id: string) => `/games/${id}/summary` as const,
  gameBoxscore: (id: string) => `/games/${id}/boxscore` as const,
  gameEvents: (id: string) => `/games/${id}/events` as const,
  gameEvent: (id: string, seq: number | string) => `/games/${id}/events/${seq}` as const,
  gamePlays: (id: string) => `/games/${id}/plays` as const,
  gamePitches: (id: string) => `/games/${id}/pitches` as const,
  gamePlayPitches: (gameId: string, playNum: number | string) => `/games/${gameId}/plays/${playNum}/pitches` as const,
  gameWinProb: (gameId: string) => `/games/${gameId}/win-probability` as const,
  gameWinProbSummary: (gameId: string) => `/games/${gameId}/win-probability/summary` as const,
  gameLeverage: (gameId: string) => `/games/${gameId}/plate-appearances/leverage` as const,

  statsBatting: '/stats/batting' as const,
  statsPitching: '/stats/pitching' as const,
  statsFielding: '/stats/fielding' as const,
  statsTeamsBatting: '/stats/teams/batting' as const,
  statsTeamsPitching: '/stats/teams/pitching' as const,
  statsTeamsFielding: '/stats/teams/fielding' as const,
  leadersBattingCareer: '/leaders/batting/career' as const,
  leadersPitchingCareer: '/leaders/pitching/career' as const,

  winExpectancy: '/win-expectancy' as const,
  winExpectancyEras: '/win-expectancy/eras' as const,

  fedGames: '/federalleague/games' as const,
  fedTeams: '/federalleague/teams' as const,
  fedPlays: '/federalleague/plays' as const,
  fedSeasonSchedule: (year: number | string) => `/federalleague/seasons/${year}/schedule` as const,
  fedSeasonTeamGames: (year: number | string, teamId: string) =>
    `/federalleague/seasons/${year}/teams/${teamId}/games` as const,

  nlgGames: '/negroleagues/games' as const,
  nlgTeams: '/negroleagues/teams' as const,
  nlgPlays: '/negroleagues/plays' as const,
  nlgSeasonSchedule: (year: number | string) => `/negroleagues/seasons/${year}/schedule` as const,
  nlgSeasonTeamGames: (year: number | string, teamId: string) =>
    `/negroleagues/seasons/${year}/teams/${teamId}/games` as const,

  mlb: '/mlb' as const,
  mlbTeams: '/mlb/teams' as const,
  mlbSchedule: '/mlb/schedule' as const,
  mlbStats: '/mlb/stats' as const,
  mlbStandings: '/mlb/standings' as const
} as const;

export const PARAMS = {
  /** Shared pagination params (all list endpoints) */
  pagination: { page: 'page', perPage: 'per_page' } as const,

  /** Search endpoints */
  search: { q: 'q' } as const,

  /** /v1/games query params */
  games: {
    season: 'season',
    homeTeam: 'home_team',
    awayTeam: 'away_team',
    parkId: 'park_id',
    dateFrom: 'date_from',
    dateTo: 'date_to',
    minInnings: 'min_innings',
    page: 'page',
    perPage: 'per_page'
  } as const,

  /** /v1/stats/* query params */
  stats: {
    sortBy: 'sort_by',
    sortOrder: 'sort_order',
    minAb: 'min_ab',
    minIp: 'min_ip',
    season: 'season',
    page: 'page',
    perPage: 'per_page'
  } as const,

  /** /v1/seasons/{year}/leaders/* */
  leaders: { stat: 'stat', page: 'page', perPage: 'per_page' } as const,

  /** /v1/teams/{id} */
  team: { year: 'year' } as const,

  /** /v1/players/{id}/stats/* */
  playerStats: { season: 'season', page: 'page', perPage: 'per_page' } as const,

  /** /v1/win-expectancy */
  winExpectancy: { inning: 'inning', outs: 'outs', runners: 'runners', era: 'era' } as const,

  /** /v1/teams/{team_id}/run-differential */
  runDifferential: { season: 'season' } as const
} as const;
