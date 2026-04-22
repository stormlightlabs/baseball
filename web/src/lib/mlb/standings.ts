import { toArray, toBoolean, toNumber, toObject, toString } from '$lib/common/converters';

export type LeagueCode = 'AL' | 'NL' | 'OTHER';

export type StandingsSortKey =
  | 'rank'
  | 'team'
  | 'wins'
  | 'losses'
  | 'pct'
  | 'gb'
  | 'wcGb'
  | 'streak'
  | 'runDiff'
  | 'last10';

export type StandingsRow = {
  league: LeagueCode;
  division: string;
  rank: number;
  teamName: string;
  teamAbbr: string;
  teamMlbID?: number;
  wins: number;
  losses: number;
  pct: string;
  pctValue: number;
  gamesBack: string;
  gamesBackValue: number;
  wildCardGamesBack: string;
  wildCardGamesBackValue: number;
  streak: string;
  streakValue: number;
  runDiff: number;
  last10: string;
  last10Wins: number;
  divisionLeader: boolean;
  localFranchiseID?: string;
};

export type SeasonStandingsPayload = { season: number; lastUpdated?: string; rows: StandingsRow[] };

function toBool(value: unknown): boolean {
  const parsed = toBoolean(value);
  if (parsed != null) return parsed;
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase();
    return normalized === 'yes';
  }
  return false;
}

function normalizeNameKey(value: string): string {
  return value.replaceAll(/[^A-Za-z0-9]/g, '').toLowerCase();
}

function normalizePct(raw: string | undefined, numeric: number): string {
  if (raw) {
    if (raw.startsWith('.')) return raw;
    if (raw.startsWith('0.')) return raw.slice(1);
    return raw;
  }
  const fixed = numeric.toFixed(3);
  if (fixed.startsWith('0.')) return fixed.slice(1);
  return fixed;
}

function normalizeGamesBack(raw: string | undefined): string {
  if (!raw) return '-';
  const trimmed = raw.trim();
  if (trimmed === 'E' || trimmed === '-' || trimmed === '--') return '-';
  return trimmed;
}

function gamesBackValue(raw: string): number {
  if (raw === '-' || raw === 'E') return 0;
  const parsed = Number(raw);
  if (Number.isFinite(parsed)) return parsed;
  return Number.POSITIVE_INFINITY;
}

function cleanDivisionLabel(raw: string | undefined, league: LeagueCode): string {
  if (raw) {
    return raw.replace('American League ', 'AL ').replace('National League ', 'NL ');
  }
  if (league === 'AL') return 'AL';
  if (league === 'NL') return 'NL';
  return 'League';
}

function leagueCodeForSeasonStanding(leagueRaw: string | undefined, divisionName: string | undefined): LeagueCode {
  const leagueValue = (leagueRaw ?? '').toUpperCase().trim();
  if (leagueValue === 'AL') return 'AL';
  if (leagueValue === 'NL') return 'NL';
  return leagueCodeFor(leagueRaw, divisionName);
}

function leagueCodeFor(leagueLabel: string | undefined, divisionLabel: string | undefined): LeagueCode {
  const label = `${leagueLabel ?? ''} ${divisionLabel ?? ''}`.toUpperCase();
  if (label.includes('AMERICAN') || label.includes(' AL')) return 'AL';
  if (label.includes('NATIONAL') || label.includes(' NL')) return 'NL';
  return 'OTHER';
}

function parseStreakValue(raw: string): number {
  const matched = raw.match(/^([WL])(\d+)$/i);
  if (!matched) return 0;
  const count = Number.parseInt(matched[2] ?? '0', 10);
  if (!Number.isFinite(count)) return 0;
  if ((matched[1] ?? '').toUpperCase() === 'W') return count;
  return -count;
}

function parseLastTenWins(display: string): number {
  const matched = display.match(/^(\d+)-(\d+)$/u);
  if (!matched) return 0;
  const wins = Number.parseInt(matched[1] ?? '0', 10);
  if (!Number.isFinite(wins)) return 0;
  return wins;
}

function parseLastTenRecord(teamRecord: Record<string, unknown>): { display: string; wins: number } {
  const explicitLastTen = toArray(teamRecord.lastTenRecords).map((entry) => toObject(entry))[0] ?? null;
  if (explicitLastTen) {
    const wins = toNumber(explicitLastTen.wins) ?? 0;
    const losses = toNumber(explicitLastTen.losses) ?? 0;
    return { display: `${wins}-${losses}`, wins };
  }

  const records = toObject(teamRecord.records);
  const splitRecords = toArray(records.splitRecords).map((entry) => toObject(entry));
  const lastTen = splitRecords.find((entry) => {
    const type = toString(entry.type)?.toLowerCase();
    return type === 'lastten' || type === 'last ten';
  });

  if (!lastTen) return { display: '—', wins: 0 };
  const wins = toNumber(lastTen.wins) ?? 0;
  const losses = toNumber(lastTen.losses) ?? 0;
  return { display: `${wins}-${losses}`, wins };
}

export function buildMlbTeamAbbrByID(payload: unknown): Record<number, string> {
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

export function buildMlbTeamAbbrByIDFromDetails(payload: unknown): Record<number, string> {
  const root = toObject(payload);
  const meta = toObject(root.meta);
  const details = toObject(meta.details);
  const mlbamTeams = toObject(details.mlbam_teams);

  const map: Record<number, string> = {};
  for (const [key, value] of Object.entries(mlbamTeams)) {
    const id = toNumber(key);
    const row = toObject(value);
    const abbr = toString(row.mlb_abbreviation);
    if (id == null || !abbr) continue;
    map[id] = abbr.toUpperCase();
  }
  return map;
}

export function buildFranchiseIDByName(payload: unknown): Record<string, string> {
  const root = toObject(payload);
  const rows = toArray(root.franchises);

  const map: Record<string, string> = {};
  const register = (candidate: string | undefined, id: string): void => {
    if (!candidate) return;
    const key = normalizeNameKey(candidate);
    if (!key) return;
    if (!map[key]) map[key] = id;
  };

  for (const entry of rows) {
    const row = toObject(entry);
    const id = toString(row.id) ?? toString(row.franchise_id) ?? toString(row.team_id);
    if (!id) continue;

    const name = toString(row.name);
    const location = toString(row.location);
    const nickname = toString(row.nickname);
    const combined = [location, nickname].filter(Boolean).join(' ').trim();

    register(name, id);
    register(combined || undefined, id);
  }

  return map;
}

export function normalizeFranchiseIDByMlbTeamID(payload: unknown): Record<number, string> {
  const root = toObject(payload);
  const rows = toArray(root.rows);
  const meta = toObject(root.meta);
  const details = toObject(meta.details);
  const mlbamTeams = toObject(details.mlbam_teams);
  const crosswalk = toObject(details.crosswalk);
  const mlbamTeamToLocal = toObject(crosswalk.mlbam_team_to_local);
  const teamToFranchise = toObject(crosswalk.team_to_franchise);

  const map: Record<number, string> = {};
  for (const entry of rows) {
    const row = toObject(entry);
    const mlbTeamID = toNumber(row.mlbam_team_id) ?? toNumber(row.mlb_team_id);
    const localFranchiseID = toString(row.franchise_id) ?? toString(row.local_franchise_id);
    if (mlbTeamID != null && localFranchiseID) {
      map[mlbTeamID] = localFranchiseID;
    }
  }

  for (const [key, value] of Object.entries(mlbamTeams)) {
    const mlbamTeamID = toNumber(key);
    const detail = toObject(value);
    const localTeamID = toString(detail.team_id);
    const localFranchiseID = toString(detail.franchise_id);
    if (mlbamTeamID != null && localFranchiseID) {
      map[mlbamTeamID] = localFranchiseID;
      continue;
    }
    if (mlbamTeamID != null && localTeamID) {
      const viaCrosswalk = toString(teamToFranchise[localTeamID]);
      if (viaCrosswalk) {
        map[mlbamTeamID] = viaCrosswalk;
      }
    }
  }

  for (const [key, value] of Object.entries(mlbamTeamToLocal)) {
    const mlbamTeamID = toNumber(key);
    const localTeamID = toString(value);
    if (mlbamTeamID == null || !localTeamID) continue;
    const localFranchiseID = toString(teamToFranchise[localTeamID]);
    if (localFranchiseID) {
      map[mlbamTeamID] = localFranchiseID;
    }
  }

  return map;
}

export function extractSeasonFromStandings(payload: unknown): number | undefined {
  const root = toObject(payload);
  const records = toArray(root.records).map((entry) => toObject(entry));
  for (const record of records) {
    const season = toNumber(record.season);
    if (season != null) return season;

    const teamRecords = toArray(record.teamRecords).map((entry) => toObject(entry));
    for (const teamRecord of teamRecords) {
      const teamSeason = toNumber(teamRecord.season);
      if (teamSeason != null) return teamSeason;
    }
  }

  return;
}

function normalizeTeamAbbrFromSeasonStanding(teamID: string | undefined, teamName: string): string {
  if (teamID && teamID.trim().length > 0) {
    return teamID.trim().toUpperCase();
  }
  const words = teamName
    .trim()
    .split(/\s+/u)
    .filter((word) => word.length > 0);
  if (words.length === 0) return '—';
  if (words.length === 1) return words[0].slice(0, 3).toUpperCase();
  return words
    .slice(-2)
    .map((word) => word[0] ?? '')
    .join('')
    .toUpperCase();
}

export function buildStandingsRowsFromSeasonStandingsPayload(payload: unknown): SeasonStandingsPayload {
  const root = toObject(payload);
  const responseSeason = toNumber(root.season) ?? new Date().getFullYear();
  const lastUpdated = toString(root.last_updated) ?? undefined;
  const rowsInput = toArray(root.standings).map((entry) => toObject(entry));

  const rows: StandingsRow[] = [];
  for (const row of rowsInput) {
    const division = toString(row.division_name) ?? 'Division';
    const league = leagueCodeForSeasonStanding(toString(row.league), division);
    const teamName = toString(row.team_name) ?? toString(row.team_id) ?? 'Unknown Team';
    const teamID = toString(row.team_id) ?? undefined;
    const teamAbbr = normalizeTeamAbbrFromSeasonStanding(teamID, teamName);
    const teamMlbID = toNumber(row.team_mlb_id) ?? undefined;

    const wins = toNumber(row.w) ?? 0;
    const losses = toNumber(row.l) ?? 0;
    const pctValue = toNumber(row.pct) ?? (wins + losses > 0 ? Math.max(0, Math.min(1, wins / (wins + losses))) : 0);
    const pct = normalizePct(toString(row.pct), pctValue);

    const gb = normalizeGamesBack(toString(row.gb));
    const wcGb = normalizeGamesBack(toString(row.wc_gb));
    const streak = toString(row.streak) ?? '—';
    const l10 = toString(row.l10) ?? '—';

    let runDiff = toNumber(row.run_diff);
    if (runDiff == null) {
      const rs = toNumber(row.rs);
      const ra = toNumber(row.ra);
      if (rs != null && ra != null) runDiff = rs - ra;
    }

    rows.push({
      league,
      division,
      rank: Number.MAX_SAFE_INTEGER,
      teamName,
      teamAbbr,
      teamMlbID,
      wins,
      losses,
      pct,
      pctValue,
      gamesBack: gb,
      gamesBackValue: gamesBackValue(gb),
      wildCardGamesBack: wcGb,
      wildCardGamesBackValue: gamesBackValue(wcGb),
      streak,
      streakValue: parseStreakValue(streak),
      runDiff: runDiff ?? 0,
      last10: l10,
      last10Wins: parseLastTenWins(l10),
      divisionLeader: gb === '-',
      localFranchiseID: toString(row.franchise_id) ?? undefined
    });
  }

  const grouped = new Map<string, StandingsRow[]>();
  const groupOrder: string[] = [];
  for (const row of rows) {
    const key = `${row.league}:${row.division}`;
    const existing = grouped.get(key);
    if (existing) {
      existing.push(row);
      continue;
    }
    grouped.set(key, [row]);
    groupOrder.push(key);
  }

  const rankedRows: StandingsRow[] = [];
  for (const key of groupOrder) {
    const divisionRows = grouped.get(key) ?? [];
    const sorted = [...divisionRows].toSorted((left, right) => {
      if (left.wins !== right.wins) return right.wins - left.wins;
      if (left.losses !== right.losses) return left.losses - right.losses;
      if (left.pctValue !== right.pctValue) return right.pctValue - left.pctValue;
      return left.teamName.localeCompare(right.teamName);
    });
    for (const [idx, element] of sorted.entries()) {
      const rank = idx + 1;
      rankedRows.push({ ...element, rank, divisionLeader: rank === 1 });
    }
  }

  return { season: responseSeason, lastUpdated, rows: rankedRows };
}

export function buildStandingsRows(
  standingsPayload: unknown,
  teamAbbrByID: Record<number, string>,
  franchiseIDByName: Record<string, string>,
  franchiseIDByMlbTeamID: Record<number, string>
): StandingsRow[] {
  const root = toObject(standingsPayload);
  const records = toArray(root.records).map((entry) => toObject(entry));
  const rows: StandingsRow[] = [];

  for (const divisionRecord of records) {
    const leagueNode = toObject(divisionRecord.league);
    const divisionNode = toObject(divisionRecord.division);
    const league = leagueCodeFor(
      toString(leagueNode.name) ?? toString(leagueNode.abbreviation),
      toString(divisionNode.name)
    );
    const division = cleanDivisionLabel(toString(divisionNode.name), league);

    const teamRecords = toArray(divisionRecord.teamRecords).map((entry) => toObject(entry));
    for (const teamRecord of teamRecords) {
      const teamNode = toObject(teamRecord.team);
      const teamName = toString(teamNode.name) ?? 'Unknown Team';
      const teamMlbID = toNumber(teamNode.id);
      const teamAbbr =
        (teamMlbID != null ? teamAbbrByID[teamMlbID] : undefined) ??
        toString(teamNode.abbreviation)?.toUpperCase() ??
        '—';

      const wins = toNumber(teamRecord.wins) ?? 0;
      const losses = toNumber(teamRecord.losses) ?? 0;
      const pctValue = toNumber(teamRecord.winningPercentage) ?? 0;
      const pct = normalizePct(toString(teamRecord.winningPercentage), pctValue);

      const gb = normalizeGamesBack(toString(teamRecord.gamesBack));
      const wcGb = normalizeGamesBack(toString(teamRecord.wildCardGamesBack));

      const streakNode = toObject(teamRecord.streak);
      const streak = toString(streakNode.streakCode) ?? '—';
      const streakValue = parseStreakValue(streak);

      const lastTen = parseLastTenRecord(teamRecord);

      let runDiff = toNumber(teamRecord.runDifferential);
      if (runDiff == null) {
        const runsScored = toNumber(teamRecord.runsScored);
        const runsAllowed = toNumber(teamRecord.runsAllowed);
        if (runsScored != null && runsAllowed != null) {
          runDiff = runsScored - runsAllowed;
        }
      }

      const rank = toNumber(teamRecord.divisionRank) ?? Number.MAX_SAFE_INTEGER;
      const localFranchiseID =
        (teamMlbID != null ? franchiseIDByMlbTeamID[teamMlbID] : undefined) ??
        franchiseIDByName[normalizeNameKey(teamName)];

      rows.push({
        league,
        division,
        rank,
        teamName,
        teamAbbr,
        teamMlbID,
        wins,
        losses,
        pct,
        pctValue,
        gamesBack: gb,
        gamesBackValue: gamesBackValue(gb),
        wildCardGamesBack: wcGb,
        wildCardGamesBackValue: gamesBackValue(wcGb),
        streak,
        streakValue,
        runDiff: runDiff ?? 0,
        last10: lastTen.display,
        last10Wins: lastTen.wins,
        divisionLeader: toBool(teamRecord.divisionLeader),
        localFranchiseID
      });
    }
  }

  return rows;
}

export function sortStandingsRows(
  rows: StandingsRow[],
  sortKey: StandingsSortKey,
  sortDir: 'asc' | 'desc'
): StandingsRow[] {
  const sorted = [...rows].toSorted((left, right) => {
    let comparison = 0;

    if (sortKey === 'rank') comparison = left.rank - right.rank;
    if (sortKey === 'team') comparison = left.teamName.localeCompare(right.teamName);
    if (sortKey === 'wins') comparison = left.wins - right.wins;
    if (sortKey === 'losses') comparison = left.losses - right.losses;
    if (sortKey === 'pct') comparison = left.pctValue - right.pctValue;
    if (sortKey === 'gb') comparison = left.gamesBackValue - right.gamesBackValue;
    if (sortKey === 'wcGb') comparison = left.wildCardGamesBackValue - right.wildCardGamesBackValue;
    if (sortKey === 'streak') comparison = left.streakValue - right.streakValue;
    if (sortKey === 'runDiff') comparison = left.runDiff - right.runDiff;
    if (sortKey === 'last10') comparison = left.last10Wins - right.last10Wins;

    if (comparison === 0) {
      comparison = left.rank - right.rank;
    }
    if (comparison === 0) {
      comparison = left.teamName.localeCompare(right.teamName);
    }

    if (sortDir === 'desc') return -comparison;
    return comparison;
  });

  return sorted;
}
