export type LeaderGroup = 'hitting' | 'pitching';

export type LeaderCategory = {
  id: 'HR' | 'AVG' | 'OPS' | 'RBI' | 'SB' | 'ERA' | 'SO' | 'W' | 'SV' | 'WHIP';
  label: string;
  group: LeaderGroup;
  statKey: string;
  descending: boolean;
  fallbackDisplay: string;
};

export type LeaderRow = {
  rank: number;
  playerName: string;
  teamAbbr: string;
  displayValue: string;
  playerMlbID?: number;
  group: LeaderGroup;
};

export const LEADER_CATEGORIES: LeaderCategory[] = [
  { id: 'HR', label: 'HR', group: 'hitting', statKey: 'homeRuns', descending: true, fallbackDisplay: '0' },
  { id: 'AVG', label: 'AVG', group: 'hitting', statKey: 'avg', descending: true, fallbackDisplay: '.000' },
  { id: 'OPS', label: 'OPS', group: 'hitting', statKey: 'ops', descending: true, fallbackDisplay: '.000' },
  { id: 'RBI', label: 'RBI', group: 'hitting', statKey: 'rbi', descending: true, fallbackDisplay: '0' },
  { id: 'SB', label: 'SB', group: 'hitting', statKey: 'stolenBases', descending: true, fallbackDisplay: '0' },
  { id: 'ERA', label: 'ERA', group: 'pitching', statKey: 'era', descending: false, fallbackDisplay: '0.00' },
  { id: 'SO', label: 'SO', group: 'pitching', statKey: 'strikeOuts', descending: true, fallbackDisplay: '0' },
  { id: 'W', label: 'W', group: 'pitching', statKey: 'wins', descending: true, fallbackDisplay: '0' },
  { id: 'SV', label: 'SV', group: 'pitching', statKey: 'saves', descending: true, fallbackDisplay: '0' },
  { id: 'WHIP', label: 'WHIP', group: 'pitching', statKey: 'whip', descending: false, fallbackDisplay: '0.00' }
];

function toObject(value: unknown): Record<string, unknown> {
  if (value != null && typeof value === 'object' && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function toArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
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
  if (typeof value !== 'string') return undefined;
  const trimmed = value.trim();
  if (trimmed.length === 0) return undefined;
  const parsed = Number(trimmed);
  if (Number.isFinite(parsed)) return parsed;
  return undefined;
}

function normalizeRateDisplay(value: string): string {
  if (value.startsWith('.')) return value;
  if (value.startsWith('0.')) return value.slice(1);
  return value;
}

function displayForStat(rawValue: unknown, fallback: string): string {
  const asString = toString(rawValue);
  if (!asString) return fallback;
  if (asString.includes('.')) {
    return normalizeRateDisplay(asString);
  }
  return asString;
}

function statValueForSort(stat: Record<string, unknown>, key: string): number {
  return toNumber(stat[key]) ?? Number.NaN;
}

function extractSplits(payload: unknown): Record<string, unknown>[] {
  const root = toObject(payload);
  const stats = toArray(root.stats);
  const first = toObject(stats[0]);
  const splits = toArray(first.splits);
  return splits.map((entry) => toObject(entry));
}

export function normalizeMlbTeamsAbbrByID(payload: unknown): Record<number, string> {
  const root = toObject(payload);
  const teams = toArray(root.teams);
  const map: Record<number, string> = {};
  for (const entry of teams) {
    const row = toObject(entry);
    const id = toNumber(row.id);
    const abbr = toString(row.abbreviation);
    if (id == null || !abbr) continue;
    map[id] = abbr.toUpperCase();
  }
  return map;
}

export function buildLeaderBoardByCategory(
  hittingPayload: unknown,
  pitchingPayload: unknown,
  teamAbbrByID: Record<number, string>,
  size = 5
): Record<LeaderCategory['id'], LeaderRow[]> {
  const hittingSplits = extractSplits(hittingPayload);
  const pitchingSplits = extractSplits(pitchingPayload);

  const categoryRows = {} as Record<LeaderCategory['id'], LeaderRow[]>;

  for (const category of LEADER_CATEGORIES) {
    const source = category.group === 'hitting' ? hittingSplits : pitchingSplits;
    const rows = source
      .map((split) => {
        const stat = toObject(split.stat);
        const player = toObject(split.player);
        const team = toObject(split.team);
        const sortValue = statValueForSort(stat, category.statKey);
        if (Number.isNaN(sortValue)) return null;

        const playerName = toString(player.fullName);
        if (!playerName) return null;

        const teamID = toNumber(team.id);
        const teamAbbr =
          (teamID != null && teamAbbrByID[teamID]) || toString(team.name)?.slice(0, 3).toUpperCase() || '—';
        const playerMlbID = toNumber(player.id);
        const displayValue = displayForStat(stat[category.statKey], category.fallbackDisplay);

        return { sortValue, playerName, teamAbbr, playerMlbID, group: category.group, displayValue };
      })
      .filter((row): row is NonNullable<typeof row> => row != null);

    rows.sort((a, b) => {
      const delta = a.sortValue - b.sortValue;
      if (delta === 0) return a.playerName.localeCompare(b.playerName);
      if (category.descending) return delta > 0 ? -1 : 1;
      return delta < 0 ? -1 : 1;
    });

    categoryRows[category.id] = rows
      .slice(0, size)
      .map((row, index) => ({
        rank: index + 1,
        playerName: row.playerName,
        teamAbbr: row.teamAbbr,
        displayValue: row.displayValue,
        playerMlbID: row.playerMlbID,
        group: row.group
      }));
  }

  return categoryRows;
}

export function normalizeNameForMatch(name: string): string {
  return name.replaceAll(/[^A-Za-z0-9]/g, '').toLowerCase();
}
