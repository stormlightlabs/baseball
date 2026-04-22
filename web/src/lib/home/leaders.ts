import { toArray, toNumber, toObject, toString } from '$lib/common/converters';

export type LeaderGroup = 'hitting' | 'pitching';

export type LeaderCategory = {
  id: 'HR' | 'AVG' | 'OPS' | 'RBI' | 'SB' | 'ERA' | 'SO' | 'W' | 'SV' | 'WHIP';
  label: string;
  group: LeaderGroup;
  statKey: string;
  localStatKey: string;
  descending: boolean;
  fallbackDisplay: string;
};

export type LeaderRow = {
  rank: number;
  playerName: string;
  teamAbbr: string;
  displayValue: string;
  playerMlbID?: number;
  localPlayerID?: string;
  group: LeaderGroup;
};

export const LEADER_CATEGORIES: LeaderCategory[] = [
  {
    id: 'HR',
    label: 'HR',
    group: 'hitting',
    statKey: 'homeRuns',
    localStatKey: 'hr',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'AVG',
    label: 'AVG',
    group: 'hitting',
    statKey: 'avg',
    localStatKey: 'avg',
    descending: true,
    fallbackDisplay: '.000'
  },
  {
    id: 'OPS',
    label: 'OPS',
    group: 'hitting',
    statKey: 'ops',
    localStatKey: 'ops',
    descending: true,
    fallbackDisplay: '.000'
  },
  {
    id: 'RBI',
    label: 'RBI',
    group: 'hitting',
    statKey: 'rbi',
    localStatKey: 'rbi',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'SB',
    label: 'SB',
    group: 'hitting',
    statKey: 'stolenBases',
    localStatKey: 'sb',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'ERA',
    label: 'ERA',
    group: 'pitching',
    statKey: 'era',
    localStatKey: 'era',
    descending: false,
    fallbackDisplay: '0.00'
  },
  {
    id: 'SO',
    label: 'SO',
    group: 'pitching',
    statKey: 'strikeOuts',
    localStatKey: 'so',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'W',
    label: 'W',
    group: 'pitching',
    statKey: 'wins',
    localStatKey: 'w',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'SV',
    label: 'SV',
    group: 'pitching',
    statKey: 'saves',
    localStatKey: 'sv',
    descending: true,
    fallbackDisplay: '0'
  },
  {
    id: 'WHIP',
    label: 'WHIP',
    group: 'pitching',
    statKey: 'whip',
    localStatKey: 'whip',
    descending: false,
    fallbackDisplay: '0.00'
  }
];

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

export function normalizeMlbTeamsAbbrByIDFromDetails(payload: unknown): Record<number, string> {
  const root = toObject(payload);
  const meta = toObject(root.meta);
  const details = toObject(meta.details);
  const mlbamTeams = toObject(details.mlbam_teams);

  const map: Record<number, string> = {};
  for (const [rawID, rawTeam] of Object.entries(mlbamTeams)) {
    const id = toNumber(rawID);
    const team = toObject(rawTeam);
    const abbr = toString(team.mlb_abbreviation);
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

type LocalLeaderLike = Record<string, unknown> & { player_id?: unknown; team_id?: unknown };

function localStatValueForSort(stat: Record<string, unknown>, key: string): number {
  return toNumber(stat[key]) ?? Number.NaN;
}

function localTeamAbbr(teamID: unknown): string {
  const team = toString(teamID);
  if (!team) return '—';
  return team.toUpperCase();
}

export function buildLocalLeaderBoardByCategory(
  hittingRows: LocalLeaderLike[],
  pitchingRows: LocalLeaderLike[],
  size = 5
): Record<LeaderCategory['id'], LeaderRow[]> {
  const categoryRows = {} as Record<LeaderCategory['id'], LeaderRow[]>;

  for (const category of LEADER_CATEGORIES) {
    const source = category.group === 'hitting' ? hittingRows : pitchingRows;
    const rows = source
      .map((entry) => {
        const stat = entry as Record<string, unknown>;
        const sortValue = localStatValueForSort(stat, category.localStatKey);
        if (Number.isNaN(sortValue)) return null;

        const localPlayerID = toString(entry.player_id);
        if (!localPlayerID) return null;

        return {
          sortValue,
          localPlayerID,
          teamAbbr: localTeamAbbr(entry.team_id),
          displayValue: displayForStat(stat[category.localStatKey], category.fallbackDisplay),
          group: category.group
        };
      })
      .filter((row): row is NonNullable<typeof row> => row != null);

    rows.sort((a, b) => {
      const delta = a.sortValue - b.sortValue;
      if (delta === 0) return a.localPlayerID.localeCompare(b.localPlayerID);
      if (category.descending) return delta > 0 ? -1 : 1;
      return delta < 0 ? -1 : 1;
    });

    categoryRows[category.id] = rows
      .slice(0, size)
      .map((row, index) => ({
        rank: index + 1,
        playerName: row.localPlayerID,
        teamAbbr: row.teamAbbr,
        displayValue: row.displayValue,
        localPlayerID: row.localPlayerID,
        group: row.group
      }));
  }

  return categoryRows;
}
