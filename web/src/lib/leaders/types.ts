export type LeaderRowKind = 'player' | 'team' | 'unknown';

export type LeaderRow = {
  id: string;
  label: string;
  team?: string;
  league?: string;
  year?: number;
  metric?: number;
  metricDisplay?: string;
  kind: LeaderRowKind;
  href?: string;
  detail?: string;
};

export type LeaderColumn = { key: keyof LeaderRow | 'rank'; label: string; align?: 'left' | 'right' };

export type ExtractedRows = { rows: Array<Record<string, unknown>>; page: number; perPage: number; total: number };
