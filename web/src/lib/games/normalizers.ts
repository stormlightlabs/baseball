import type { PaginatedResponse } from '$lib/api';
import type {
  GameBoxscore,
  GamePaginatedResponse,
  GameRecord,
  GameSummary,
  GameWinProbabilitySummary,
  LineupPlayer,
  PitchEvent,
  PlateAppearanceLeverage,
  PlayEvent,
  TeamGameStats,
  WinProbabilityCurve,
  WinProbabilityPoint
} from '$lib/games/types';

function toObject(value: unknown): Record<string, unknown> {
  if (value && typeof value === 'object') {
    return value as Record<string, unknown>;
  }
  return {};
}

function toString(value: unknown): string | undefined {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : undefined;
  }
  if (typeof value === 'number') return String(value);
  return undefined;
}

function toNumber(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (typeof value === 'string' && value.trim().length > 0) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
}

function toBoolean(value: unknown): boolean | undefined {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'string') {
    if (value === 'true') return true;
    if (value === 'false') return false;
  }
  return undefined;
}

function toStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.map((entry) => toString(entry)).filter((entry): entry is string => entry != null);
}

function normalizeLineupPlayer(raw: unknown): LineupPlayer {
  const row = toObject(raw);
  return {
    player_id: toString(row.player_id),
    name: toString(row.name),
    position: toNumber(row.position),
    bats: toString(row.bats),
    throws: toString(row.throws)
  };
}

function normalizeTeamGameStats(raw: unknown): TeamGameStats {
  const row = toObject(raw);
  return {
    ab: toNumber(row.ab),
    h: toNumber(row.h),
    r: toNumber(row.r),
    doubles: toNumber(row.doubles),
    triples: toNumber(row.triples),
    hr: toNumber(row.hr),
    rbi: toNumber(row.rbi),
    sh: toNumber(row.sh),
    sf: toNumber(row.sf),
    hbp: toNumber(row.hbp),
    bb: toNumber(row.bb),
    ibb: toNumber(row.ibb),
    so: toNumber(row.so),
    sb: toNumber(row.sb),
    cs: toNumber(row.cs),
    gdp: toNumber(row.gdp),
    lob: toNumber(row.lob),
    pitchers_used: toNumber(row.pitchers_used),
    er: toNumber(row.er),
    wp: toNumber(row.wp),
    balks: toNumber(row.balks),
    po: toNumber(row.po),
    a: toNumber(row.a),
    e: toNumber(row.e),
    pb: toNumber(row.pb),
    dp: toNumber(row.dp),
    tp: toNumber(row.tp)
  };
}

export function normalizeGameRecord(raw: unknown): GameRecord {
  const row = toObject(raw);
  return {
    id: toString(row.id) ?? toString(row.game_id) ?? 'unknown-game',
    season: toNumber(row.season),
    date: toString(row.date),
    day_of_week: toString(row.day_of_week),
    home_team: toString(row.home_team),
    away_team: toString(row.away_team),
    home_league: toString(row.home_league),
    away_league: toString(row.away_league),
    home_score: toNumber(row.home_score),
    away_score: toNumber(row.away_score),
    innings: toNumber(row.innings),
    attendance: toNumber(row.attendance),
    duration_min: toNumber(row.duration_min),
    park_id: toString(row.park_id),
    park_name: toString(row.park_name),
    park_city: toString(row.park_city),
    park_state: toString(row.park_state),
    day_night: toString(row.day_night),
    is_postseason: toBoolean(row.is_postseason),
    winning_pitcher_name: toString(row.winning_pitcher_name),
    losing_pitcher_name: toString(row.losing_pitcher_name),
    save_pitcher_name: toString(row.save_pitcher_name),
    home_manager_name: toString(row.home_manager_name),
    away_manager_name: toString(row.away_manager_name)
  };
}

export function normalizeGamePage(
  payload: PaginatedResponse<Record<string, unknown>>
): GamePaginatedResponse<GameRecord> {
  return { ...payload, data: payload.data.map((row) => normalizeGameRecord(row)) };
}

export function normalizeGameSummary(raw: unknown): GameSummary {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    date: toString(row.date),
    home_team: toString(row.home_team),
    away_team: toString(row.away_team),
    home_score: toNumber(row.home_score),
    away_score: toNumber(row.away_score),
    innings: toNumber(row.innings),
    winner: toString(row.winner),
    home_lineup: Array.isArray(row.home_lineup) ? row.home_lineup.map((entry) => normalizeLineupPlayer(entry)) : [],
    away_lineup: Array.isArray(row.away_lineup) ? row.away_lineup.map((entry) => normalizeLineupPlayer(entry)) : []
  };
}

export function normalizeGameBoxscore(raw: unknown): GameBoxscore {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    date: toString(row.date),
    home_team: toString(row.home_team),
    away_team: toString(row.away_team),
    home_score: toNumber(row.home_score),
    away_score: toNumber(row.away_score),
    home_stats: normalizeTeamGameStats(row.home_stats),
    away_stats: normalizeTeamGameStats(row.away_stats),
    home_lineup: Array.isArray(row.home_lineup) ? row.home_lineup.map((entry) => normalizeLineupPlayer(entry)) : [],
    away_lineup: Array.isArray(row.away_lineup) ? row.away_lineup.map((entry) => normalizeLineupPlayer(entry)) : []
  };
}

export function normalizePlayEvent(raw: unknown): PlayEvent {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    play_num: toNumber(row.play_num),
    inning: toNumber(row.inning),
    top_bot: toNumber(row.top_bot),
    bat_team: toString(row.bat_team),
    pit_team: toString(row.pit_team),
    date: toString(row.date),
    batter: toString(row.batter),
    batter_name: toString(row.batter_name),
    pitcher: toString(row.pitcher),
    pitcher_name: toString(row.pitcher_name),
    score_vis: toNumber(row.score_vis),
    score_home: toNumber(row.score_home),
    outs_pre: toNumber(row.outs_pre),
    outs_post: toNumber(row.outs_post),
    balls: toNumber(row.balls),
    strikes: toNumber(row.strikes),
    event: toString(row.event),
    runs: toNumber(row.runs),
    rbi: toNumber(row.rbi)
  };
}

export function normalizePlayPage(
  payload: PaginatedResponse<Record<string, unknown>>
): GamePaginatedResponse<PlayEvent> {
  return { ...payload, data: payload.data.map((row) => normalizePlayEvent(row)) };
}

export function normalizePitchEvent(raw: unknown): PitchEvent {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    play_num: toNumber(row.play_num),
    inning: toNumber(row.inning),
    top_bot: toNumber(row.top_bot),
    bat_team: toString(row.bat_team),
    pit_team: toString(row.pit_team),
    date: toString(row.date),
    batter: toString(row.batter),
    batter_name: toString(row.batter_name),
    pitcher: toString(row.pitcher),
    pitcher_name: toString(row.pitcher_name),
    outs_pre: toNumber(row.outs_pre),
    seq_num: toNumber(row.seq_num),
    pitch_type: toString(row.pitch_type),
    ball_count: toNumber(row.ball_count),
    strike_count: toNumber(row.strike_count),
    is_in_play: toBoolean(row.is_in_play),
    is_strike: toBoolean(row.is_strike),
    is_ball: toBoolean(row.is_ball),
    description: toString(row.description),
    event: toString(row.event)
  };
}

export function normalizePitchPage(
  payload: PaginatedResponse<Record<string, unknown>>
): GamePaginatedResponse<PitchEvent> {
  return { ...payload, data: payload.data.map((row) => normalizePitchEvent(row)) };
}

export function normalizePlayPitches(raw: unknown): PitchEvent[] {
  const obj = toObject(raw);
  if (Array.isArray(obj.data)) {
    return obj.data.map((entry) => normalizePitchEvent(entry));
  }
  if (Array.isArray(raw)) {
    return raw.map((entry) => normalizePitchEvent(entry));
  }
  return [];
}

function normalizeWinProbabilityPoint(raw: unknown): WinProbabilityPoint {
  const row = toObject(raw);
  return {
    event_index: toNumber(row.event_index),
    inning: toNumber(row.inning),
    top_of_inning: toBoolean(row.top_of_inning),
    home_score: toNumber(row.home_score),
    away_score: toNumber(row.away_score),
    outs: toNumber(row.outs),
    bases: toString(row.bases),
    home_win_prob: toNumber(row.home_win_prob),
    away_win_prob: toNumber(row.away_win_prob),
    description: toString(row.description)
  };
}

export function normalizeWinProbabilityCurve(raw: unknown): WinProbabilityCurve {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    season: toNumber(row.season),
    home_team: toString(row.home_team),
    away_team: toString(row.away_team),
    points: Array.isArray(row.points) ? row.points.map((point) => normalizeWinProbabilityPoint(point)) : []
  };
}

export function normalizePlateAppearanceLeverage(raw: unknown): PlateAppearanceLeverage {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    event_id: toNumber(row.event_id),
    batter_id: toString(row.batter_id),
    pitcher_id: toString(row.pitcher_id),
    inning: toNumber(row.inning),
    top_of_inning: toBoolean(row.top_of_inning),
    home_score_before: toNumber(row.home_score_before),
    away_score_before: toNumber(row.away_score_before),
    outs_before: toNumber(row.outs_before),
    bases_before: toString(row.bases_before),
    we_before: toNumber(row.we_before),
    we_after: toNumber(row.we_after),
    li: toNumber(row.li),
    we_change: toNumber(row.we_change),
    description: toString(row.description)
  };
}

export function normalizeLeverageList(raw: unknown): PlateAppearanceLeverage[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((entry) => normalizePlateAppearanceLeverage(entry));
}

export function normalizeGameWinProbabilitySummary(raw: unknown): GameWinProbabilitySummary {
  const row = toObject(raw);
  return {
    game_id: toString(row.game_id),
    season: toNumber(row.season),
    home_team: toString(row.home_team),
    away_team: toString(row.away_team),
    home_win_prob_start: toNumber(row.home_win_prob_start),
    home_win_prob_end: toNumber(row.home_win_prob_end),
    biggest_positive_swing: row.biggest_positive_swing
      ? normalizePlateAppearanceLeverage(row.biggest_positive_swing)
      : undefined,
    biggest_negative_swing: row.biggest_negative_swing
      ? normalizePlateAppearanceLeverage(row.biggest_negative_swing)
      : undefined
  };
}

export function toSortedEventIds(events: PlayEvent[]): number[] {
  return events
    .map((event) => event.play_num)
    .filter((value): value is number => value != null)
    .toSorted((a, b) => a - b);
}

export function toBoxscoreStatRows(
  home: TeamGameStats | undefined,
  away: TeamGameStats | undefined
): Array<{ label: string; home: string; away: string }> {
  const keys = [
    { label: 'R', key: 'r' },
    { label: 'H', key: 'h' },
    { label: 'AB', key: 'ab' },
    { label: 'HR', key: 'hr' },
    { label: 'BB', key: 'bb' },
    { label: 'SO', key: 'so' },
    { label: 'LOB', key: 'lob' },
    { label: 'Errors', key: 'e' },
    { label: 'Pitchers', key: 'pitchers_used' }
  ] as const;

  return keys.map((entry) => {
    const homeValue = home?.[entry.key];
    const awayValue = away?.[entry.key];
    return {
      label: entry.label,
      home: homeValue == null ? '—' : String(homeValue),
      away: awayValue == null ? '—' : String(awayValue)
    };
  });
}

export function toCompactLineup(lineup: LineupPlayer[]): string[] {
  return lineup.map((player) => {
    const name = player.name ?? player.player_id ?? 'Unknown';
    const position = player.position != null ? ` (${player.position})` : '';
    return `${name}${position}`;
  });
}

export function parseDateOnly(raw: string | undefined): string {
  if (!raw) return '—';
  if (raw.includes('T')) return raw.slice(0, 10);
  if (raw.length >= 10) return raw.slice(0, 10);
  return raw;
}

export function detectSeason(game: GameRecord | null): number | undefined {
  if (!game) return undefined;
  if (game.season != null) return game.season;
  const date = game.date;
  if (!date) return undefined;
  const year = Number.parseInt(date.slice(0, 4), 10);
  return Number.isNaN(year) ? undefined : year;
}

export function endpointLabelFromFamily(family: 'mlb' | 'fed' | 'nlg'): string {
  switch (family) {
    case 'fed': {
      return 'Federal League route family';
    }
    case 'nlg': {
      return 'Negro Leagues route family';
    }
    default: {
      return 'Standard MLB/AL/NL route family';
    }
  }
}

export function summaryPitcherLine(game: GameRecord | null): string[] {
  if (!game) return [];
  const lines = [
    game.winning_pitcher_name ? `W: ${game.winning_pitcher_name}` : undefined,
    game.losing_pitcher_name ? `L: ${game.losing_pitcher_name}` : undefined,
    game.save_pitcher_name ? `SV: ${game.save_pitcher_name}` : undefined
  ];
  return lines.filter((line): line is string => line != null);
}

export function extractTeamCodes(games: GameRecord[]): string[] {
  const home = games.map((game) => game.home_team);
  const away = games.map((game) => game.away_team);
  return [...home, ...away]
    .filter((team): team is string => team != null)
    .toSorted((a, b) => a.localeCompare(b))
    .filter((team, index, array) => index === 0 || array[index - 1] !== team);
}

export function normalizePaginatedFallback<T>(
  payload: PaginatedResponse<T>,
  mapper: (entry: T) => T
): PaginatedResponse<T> {
  return { ...payload, data: payload.data.map((entry) => mapper(entry)) };
}

export function normalizeStringList(raw: unknown): string[] {
  return toStringArray(raw);
}
