import type { PaginatedResponse } from '$lib/api';
import type {
  FranchiseProfile,
  RunDifferentialSeries,
  TeamBattingStats,
  TeamDailyLog,
  TeamDailyStat,
  TeamFieldingStats,
  TeamPitchingStats,
  TeamResult,
  TeamSeasonProfile
} from './types';

type ApiTeamPayload = Record<string, unknown>;

function toTeamName(team: ApiTeamPayload): string {
  if (typeof team.name === 'string' && team.name.trim()) return team.name;
  const location = typeof team.location === 'string' ? team.location : '';
  const nickname = typeof team.nickname === 'string' ? team.nickname : '';
  const combined = [location, nickname].filter(Boolean).join(' ').trim();
  if (combined) return combined;
  return String(team.id ?? '?');
}

function toStr(v: unknown): string | undefined {
  return v != null && v !== '' ? String(v) : undefined;
}

function toNum(v: unknown): number | undefined {
  if (typeof v === 'number') return v;
  if (typeof v === 'string') {
    const n = Number(v);
    return Number.isNaN(n) ? undefined : n;
  }
  return undefined;
}

export function normalizeTeamResult(team: ApiTeamPayload): TeamResult {
  const teamId = toStr(team.team_id) ?? toStr(team.id);
  const franchiseId = toStr(team.franchise_id) ?? toStr(team.id);
  const id = teamId ?? franchiseId ?? '';
  return {
    id,
    team_id: teamId,
    name: toTeamName(team),
    franchise_id: franchiseId,
    league: toStr(team.league),
    division: toStr(team.division),
    year: toNum(team.year),
    active_from: toNum(team.active_from),
    active_to: toNum(team.active_to)
  };
}

export function normalizeSearchTeamsPage(payload: PaginatedResponse<ApiTeamPayload>): PaginatedResponse<TeamResult> {
  return { ...payload, data: payload.data.map((r) => normalizeTeamResult(r)) };
}

export function normalizeFranchiseProfile(team: ApiTeamPayload): FranchiseProfile {
  return {
    id: toStr(team.id) ?? toStr(team.franchise_id) ?? toStr(team.team_id) ?? '',
    name: toTeamName(team),
    abbr: toStr(team.abbr),
    location: toStr(team.location),
    league: toStr(team.league),
    division: toStr(team.division),
    active_from: toNum(team.active_from),
    active_to: toNum(team.active_to)
  };
}

export function normalizeTeamSeasonProfile(team: ApiTeamPayload): TeamSeasonProfile {
  return {
    id: toStr(team.team_id) ?? toStr(team.id) ?? '',
    franchise_id: toStr(team.franchise_id),
    name: toStr(team.name),
    year: toNum(team.year),
    league: toStr(team.league),
    division: toStr(team.division),
    wins: toNum(team.wins) ?? toNum(team.w),
    losses: toNum(team.losses) ?? toNum(team.l),
    rank: toNum(team.rank),
    park: toStr(team.park) ?? toStr(team.park_name),
    park_id: toStr(team.park_id)
  };
}

export function normalizeTeamBattingStats(payload: ApiTeamPayload): TeamBattingStats {
  return {
    team_id: toStr(payload.team_id),
    year: toNum(payload.year),
    league: toStr(payload.league),
    g: toNum(payload.g),
    ab: toNum(payload.ab),
    h: toNum(payload.h),
    hr: toNum(payload.hr),
    r: toNum(payload.r),
    rbi: toNum(payload.rbi),
    avg: toNum(payload.avg),
    obp: toNum(payload.obp),
    slg: toNum(payload.slg),
    ops: toNum(payload.ops),
    players: Array.isArray(payload.players) ? (payload.players as Array<Record<string, unknown>>) : []
  };
}

export function normalizeTeamPitchingStats(payload: ApiTeamPayload): TeamPitchingStats {
  return {
    team_id: toStr(payload.team_id),
    year: toNum(payload.year),
    league: toStr(payload.league),
    g: toNum(payload.g),
    w: toNum(payload.w),
    l: toNum(payload.l),
    era: toNum(payload.era),
    whip: toNum(payload.whip),
    so: toNum(payload.so),
    bb: toNum(payload.bb),
    h: toNum(payload.h),
    hr: toNum(payload.hr),
    players: Array.isArray(payload.players) ? (payload.players as Array<Record<string, unknown>>) : []
  };
}

export function normalizeTeamFieldingStats(payload: ApiTeamPayload): TeamFieldingStats {
  return {
    team_id: toStr(payload.team_id),
    year: toNum(payload.year),
    league: toStr(payload.league),
    g: toNum(payload.g),
    e: toNum(payload.e),
    a: toNum(payload.a),
    po: toNum(payload.po),
    fpct: toNum(payload.fpct),
    dp: toNum(payload.dp),
    players: Array.isArray(payload.players) ? (payload.players as Array<Record<string, unknown>>) : []
  };
}

export function normalizeRunDifferentialSeries(payload: ApiTeamPayload): RunDifferentialSeries {
  return {
    season: toNum(payload.season),
    games_played: toNum(payload.games_played),
    runs_scored: toNum(payload.runs_scored),
    runs_allowed: toNum(payload.runs_allowed),
    run_differential: toNum(payload.run_differential),
    games: Array.isArray(payload.games) ? (payload.games as Array<Record<string, unknown>>) : []
  };
}

export function normalizeTeamDailyStat(payload: ApiTeamPayload): TeamDailyStat {
  return {
    game_id: toStr(payload.game_id),
    team_id: toStr(payload.team_id),
    date: toStr(payload.date),
    season: toNum(payload.season),
    runs: toNum(payload.runs),
    runs_allowed: toNum(payload.runs_allowed),
    h: toNum(payload.h),
    hr: toNum(payload.hr),
    bb: toNum(payload.bb),
    so: toNum(payload.so),
    avg: toNum(payload.avg),
    obp: toNum(payload.obp),
    slg: toNum(payload.slg),
    result: toStr(payload.result)
  };
}

export function normalizeTeamDailyLog(payload: ApiTeamPayload): TeamDailyLog {
  return {
    date: toStr(payload.date),
    games_played: toNum(payload.games_played),
    wins: toNum(payload.wins),
    losses: toNum(payload.losses),
    runs_scored: toNum(payload.runs_scored),
    runs_allowed: toNum(payload.runs_allowed),
    run_diff: toNum(payload.run_diff)
  };
}

export function normalizeTeamDailyStatsPage(
  payload: PaginatedResponse<ApiTeamPayload>
): PaginatedResponse<TeamDailyStat> {
  return { ...payload, data: payload.data.map((r) => normalizeTeamDailyStat(r)) };
}

export function normalizeTeamDailyLogsPage(
  payload: PaginatedResponse<ApiTeamPayload>
): PaginatedResponse<TeamDailyLog> {
  return { ...payload, data: payload.data.map((r) => normalizeTeamDailyLog(r)) };
}
