export type TableRow = Record<string, unknown>;

export type SortableColumn = {
  key: string;
  label: string;
  sortable?: boolean;
  format?: (value: unknown) => string;
  rank?: boolean;
};

export type ApiListPayload<T> = T[] | { data?: T[] } | T | null | undefined;

export function normalizeApiList<T>(payload: ApiListPayload<T>): T[] {
  if (Array.isArray(payload)) return payload;
  if (!payload || typeof payload !== 'object') return [];

  const record = payload as Record<string, unknown>;
  if (Array.isArray(record.data)) {
    return record.data as T[];
  }

  return [payload as T];
}

export function rowColumns(rows: TableRow[]): SortableColumn[] {
  if (rows.length === 0) return [];

  return Object.keys(rows[0]).map((key) => ({ key, label: key.toUpperCase(), sortable: true }));
}

export type PlayerResult = {
  id: string;
  name: string;
  position?: string;
  primary_position?: string;
  debut_year?: number;
  final_year?: number;
};

export type PlayerProfile = {
  id: string;
  name: string;
  bats?: string;
  throws?: string;
  birth_date?: string;
  birth_city?: string;
  birth_state?: string;
  debut_year?: number;
  final_year?: number;
  primary_position?: string;
  positions?: string[];
  career_hr?: number;
  career_avg?: number;
  career_rbi?: number;
};

export type BattingSeason = {
  year: number;
  team?: string;
  team_id?: string;
  g: number;
  ab: number;
  h: number;
  hr: number;
  rbi: number;
  avg: number;
  sb?: number;
  obp?: number;
  slg?: number;
  ops?: number;
};

export type PitchingSeason = {
  year: number;
  team?: string;
  team_id?: string;
  g: number;
  gs?: number;
  w: number;
  l: number;
  era: number;
  ip: number;
  so?: number;
  bb?: number;
  whip?: number;
  sv?: number;
};

export type Award = { year: number; name?: string; award_id?: string; league?: string; notes?: string };

export type HofEntry = {
  year_inducted?: number;
  inducted?: boolean;
  votes?: number;
  ballots?: number;
  pct?: number;
  category?: string;
};

export type PlayerTeam = { year?: number; team_id?: string; team?: string; league?: string; g?: number };

export type Salary = { year: number; team?: string; team_id?: string; salary: number };

export type Relative = { player_id?: string; name?: string; relationship?: string };

export type GameLog = TableRow;
