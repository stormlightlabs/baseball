import type { PaginatedResponse } from '$lib/api';
import {
  toBoolean as toBool,
  toNumber as toNum,
  toObject,
  toObjectNullable,
  toRecordArray,
  toString as toStr,
  toStringArray
} from '$lib/common/converters';
import type {
  SeasonAwardResult,
  SeasonBattingLeader,
  SeasonGame,
  SeasonParkFactor,
  SeasonPitchingLeader,
  SeasonPostseasonSeries,
  SeasonSummary,
  SeasonTeam
} from './types';

type ApiRow = Record<string, unknown>;

type RowsPayload = { rows: ApiRow[]; page: number; perPage: number; total: number };

function toStrList(value: unknown): string[] {
  if (Array.isArray(value)) {
    return toStringArray(value);
  }

  const text = toStr(value);
  if (!text) return [];
  return text
    .split(',')
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}

function extractRows(payload: unknown, keys: string[]): RowsPayload {
  if (Array.isArray(payload)) {
    const rows = toRecordArray(payload);
    return { rows, page: 1, perPage: rows.length || 1, total: rows.length };
  }

  const obj = toObject(payload);

  let rows: ApiRow[] = [];
  for (const key of keys) {
    rows = toRecordArray(obj[key]);
    if (rows.length > 0) break;
  }

  const page = toNum(obj.page) ?? 1;
  const perPage = (toNum(obj.per_page) ?? rows.length) || 1;
  const total = toNum(obj.total) ?? rows.length;

  return { rows, page, perPage, total };
}

function toDate(value: unknown): string | undefined {
  const raw = toStr(value);
  if (!raw) return undefined;
  if (raw.length >= 10) return raw.slice(0, 10);
  return raw;
}

export function normalizeSeasons(payload: unknown): SeasonSummary[] {
  const rows = Array.isArray(payload)
    ? toRecordArray(payload)
    : toRecordArray(toObjectNullable(payload)?.seasons ?? toObjectNullable(payload)?.data ?? payload);

  return rows
    .map((row): SeasonSummary | undefined => {
      const year = toNum(row.year);
      if (year == null) return;
      return {
        year,
        leagues: toStrList(row.leagues),
        team_count: toNum(row.team_count),
        game_count: toNum(row.game_count)
      } satisfies SeasonSummary;
    })
    .filter((row): row is SeasonSummary => row !== undefined)
    .toSorted((a, b) => b.year - a.year);
}

export function normalizeSeasonTeamsPage(payload: unknown): PaginatedResponse<SeasonTeam> {
  const page = extractRows(payload, ['data', 'teams']);

  return {
    data: page.rows
      .map((row): SeasonTeam | undefined => {
        const id = toStr(row.team_id) ?? toStr(row.id);
        if (!id) return;

        return {
          team_id: id,
          name: toStr(row.name),
          franchise_id: toStr(row.franchise_id),
          league: toStr(row.league),
          division: toStr(row.division),
          year: toNum(row.year),
          wins: toNum(row.wins) ?? toNum(row.w),
          losses: toNum(row.losses) ?? toNum(row.l),
          ties: toNum(row.ties),
          games: toNum(row.games) ?? toNum(row.g),
          runs_scored: toNum(row.runs_scored) ?? toNum(row.r),
          runs_allowed: toNum(row.runs_allowed) ?? toNum(row.ra),
          attendance: toNum(row.attendance)
        } satisfies SeasonTeam;
      })
      .filter((row): row is SeasonTeam => row !== undefined),
    page: page.page,
    per_page: page.perPage,
    total: page.total
  };
}

export function normalizeBattingLeadersPage(payload: unknown): PaginatedResponse<SeasonBattingLeader> {
  const page = extractRows(payload, ['leaders', 'data']);

  return {
    data: page.rows
      .map((row): SeasonBattingLeader | undefined => {
        const playerId = toStr(row.player_id) ?? toStr(row.id);
        if (!playerId) return;

        return {
          player_id: playerId,
          year: toNum(row.year),
          team_id: toStr(row.team_id),
          league: toStr(row.league),
          g: toNum(row.g),
          pa: toNum(row.pa),
          ab: toNum(row.ab),
          r: toNum(row.r),
          h: toNum(row.h),
          hr: toNum(row.hr),
          rbi: toNum(row.rbi),
          sb: toNum(row.sb),
          avg: toNum(row.avg),
          obp: toNum(row.obp),
          slg: toNum(row.slg),
          ops: toNum(row.ops)
        } satisfies SeasonBattingLeader;
      })
      .filter((row): row is SeasonBattingLeader => row !== undefined),
    page: page.page,
    per_page: page.perPage,
    total: page.total
  };
}

export function normalizePitchingLeadersPage(payload: unknown): PaginatedResponse<SeasonPitchingLeader> {
  const page = extractRows(payload, ['leaders', 'data']);

  return {
    data: page.rows
      .map((row): SeasonPitchingLeader | undefined => {
        const playerId = toStr(row.player_id) ?? toStr(row.id);
        if (!playerId) return;

        return {
          player_id: playerId,
          year: toNum(row.year),
          team_id: toStr(row.team_id),
          league: toStr(row.league),
          w: toNum(row.w),
          l: toNum(row.l),
          sv: toNum(row.sv),
          so: toNum(row.so),
          era: toNum(row.era),
          whip: toNum(row.whip),
          ip_outs: toNum(row.ip_outs),
          g: toNum(row.g)
        } satisfies SeasonPitchingLeader;
      })
      .filter((row): row is SeasonPitchingLeader => row !== undefined),
    page: page.page,
    per_page: page.perPage,
    total: page.total
  };
}

export function normalizeGamesPage(payload: unknown): PaginatedResponse<SeasonGame> {
  const page = extractRows(payload, ['data', 'games']);

  return {
    data: page.rows
      .map((row): SeasonGame | undefined => {
        const id = toStr(row.id) ?? toStr(row.game_id);
        if (!id) return;

        return {
          id,
          date: toDate(row.date),
          season: toNum(row.season),
          home_team: toStr(row.home_team),
          away_team: toStr(row.away_team),
          home_score: toNum(row.home_score),
          away_score: toNum(row.away_score),
          innings: toNum(row.innings),
          is_postseason: toBool(row.is_postseason),
          park_id: toStr(row.park_id),
          park_name: toStr(row.park_name)
        } satisfies SeasonGame;
      })
      .filter((row): row is SeasonGame => row !== undefined),
    page: page.page,
    per_page: page.perPage,
    total: page.total
  };
}

export function normalizeDateGames(payload: unknown): SeasonGame[] {
  return toRecordArray(payload)
    .map((row): SeasonGame | undefined => {
      const id = toStr(row.id) ?? toStr(row.game_id);
      if (!id) return;

      return {
        id,
        date: toDate(row.date),
        season: toNum(row.season),
        home_team: toStr(row.home_team),
        away_team: toStr(row.away_team),
        home_score: toNum(row.home_score),
        away_score: toNum(row.away_score),
        innings: toNum(row.innings),
        is_postseason: toBool(row.is_postseason),
        park_id: toStr(row.park_id),
        park_name: toStr(row.park_name)
      } satisfies SeasonGame;
    })
    .filter((row): row is SeasonGame => row !== undefined);
}

export function normalizeAwardsPage(payload: unknown): PaginatedResponse<SeasonAwardResult> {
  const page = extractRows(payload, ['data', 'awards']);

  return {
    data: page.rows.map((row) => ({
      award_id: toStr(row.award_id),
      player_id: toStr(row.player_id),
      year: toNum(row.year),
      league: toStr(row.league),
      votes_first: toNum(row.votes_first),
      points: toNum(row.points),
      rank: toNum(row.rank)
    })),
    page: page.page,
    per_page: page.perPage,
    total: page.total
  };
}

export function normalizePostseasonSeries(payload: unknown): SeasonPostseasonSeries[] {
  const rows = extractRows(payload, ['series', 'data']).rows;

  return rows.map((row) => ({
    year: toNum(row.year),
    round: toStr(row.round),
    winner_team: toStr(row.winner_team),
    winner_league: toStr(row.winner_league),
    loser_team: toStr(row.loser_team),
    loser_league: toStr(row.loser_league),
    wins: toNum(row.wins),
    losses: toNum(row.losses),
    ties: toNum(row.ties)
  }));
}

export function normalizeParkFactors(payload: unknown): SeasonParkFactor[] {
  return toRecordArray(payload).map((row) => ({
    season: toNum(row.season),
    park_id: toStr(row.park_id),
    provider: toStr(row.provider),
    runs_factor: toNum(row.runs_factor),
    hr_factor: toNum(row.hr_factor),
    h_factor: toNum(row.h_factor),
    bb_factor: toNum(row.bb_factor),
    runs_factor_lhb: toNum(row.runs_factor_lhb),
    runs_factor_rhb: toNum(row.runs_factor_rhb),
    games_sampled: toNum(row.games_sampled),
    multi_year: toBool(row.multi_year)
  }));
}
