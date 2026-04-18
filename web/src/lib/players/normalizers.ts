import type { PaginatedResponse } from '$lib/api';
import type {
  ApiPlayerPayload,
  HallOfFamePayload,
  PlayerStatsPayload,
  HofEntry,
  PlayerProfile,
  PlayerResult
} from '$lib/players/types';

function toDisplayName(player: ApiPlayerPayload): string {
  if (player.name && player.name.trim().length > 0) return player.name;
  const full = [player.first_name, player.last_name].filter(Boolean).join(' ').trim();
  return full || player.id;
}

function normalizePositions(raw: string | string[] | undefined): string[] | undefined {
  if (Array.isArray(raw)) return raw.filter((p) => p.trim().length > 0);
  if (!raw) return undefined;
  return raw
    .split(',')
    .map((segment) => segment.trim().split(':')[0]?.trim() ?? '')
    .filter((p) => p.length > 0);
}

function toBirthDate(player: ApiPlayerPayload): string | undefined {
  if (player.birth_date) return player.birth_date;
  if (!player.birth_year || !player.birth_month || !player.birth_day) return undefined;
  return `${player.birth_year}-${String(player.birth_month).padStart(2, '0')}-${String(player.birth_day).padStart(2, '0')}`;
}

function seasonsFromPayload<T>(payload: PlayerStatsPayload<T>): T[] {
  if (!payload || typeof payload !== 'object') return [];
  if (Array.isArray(payload.seasons)) return payload.seasons;
  if (Array.isArray(payload.data)) return payload.data;
  return [];
}

function paginateArray<T>(items: T[], page: number, perPage: number): PaginatedResponse<T> {
  const safePage = Number.isFinite(page) && page > 0 ? Math.floor(page) : 1;
  const safePerPage = Number.isFinite(perPage) && perPage > 0 ? Math.floor(perPage) : 20;
  const start = (safePage - 1) * safePerPage;
  return { data: items.slice(start, start + safePerPage), page: safePage, per_page: safePerPage, total: items.length };
}

export function normalizePlayerResult(player: ApiPlayerPayload): PlayerResult {
  const positions = normalizePositions(player.positions);
  const primary = player.primary_position ?? player.position ?? positions?.[0];
  return {
    id: player.id,
    name: toDisplayName(player),
    position: player.position ?? primary,
    primary_position: primary,
    debut_year: player.debut_year,
    final_year: player.final_year ?? player.latest_season
  };
}

export function normalizePlayerProfile(player: ApiPlayerPayload): PlayerProfile {
  const positions = normalizePositions(player.positions);
  const primary = player.primary_position ?? player.position ?? positions?.[0];
  return {
    id: player.id,
    name: toDisplayName(player),
    bats: player.bats,
    throws: player.throws,
    birth_date: toBirthDate(player),
    birth_city: player.birth_city,
    birth_state: player.birth_state,
    debut_year: player.debut_year,
    final_year: player.final_year ?? player.latest_season,
    primary_position: primary,
    positions
  };
}

export function normalizeSearchPlayersPage(
  payload: PaginatedResponse<ApiPlayerPayload>
): PaginatedResponse<PlayerResult> {
  return { ...payload, data: payload.data.map((r) => normalizePlayerResult(r)) };
}

export function normalizePlayerStatsPage<T>(
  payload: PlayerStatsPayload<T>,
  page: number,
  perPage: number
): PaginatedResponse<T> {
  return paginateArray(seasonsFromPayload(payload), page, perPage);
}

export function normalizeHallOfFamePayload(payload: HallOfFamePayload<HofEntry>): HofEntry[] {
  if (Array.isArray(payload.records)) return payload.records;
  if (Array.isArray(payload.data)) return payload.data;
  return [];
}
