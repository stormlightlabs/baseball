export type TeamResult = {
  id: string;
  name: string;
  franchise_id?: string;
  league?: string;
  division?: string;
  year?: number;
  active_from?: number;
  active_to?: number;
};

export type FranchiseProfile = {
  id: string;
  name: string;
  abbr?: string;
  location?: string;
  league?: string;
  division?: string;
  active_from?: number;
  active_to?: number;
};

export type TeamBattingStats = {
  team_id?: string;
  year?: number;
  league?: string;
  g?: number;
  ab?: number;
  h?: number;
  hr?: number;
  r?: number;
  rbi?: number;
  avg?: number;
  obp?: number;
  slg?: number;
  ops?: number;
  players?: Array<Record<string, unknown>>;
};

export type TeamPitchingStats = {
  team_id?: string;
  year?: number;
  league?: string;
  g?: number;
  w?: number;
  l?: number;
  era?: number;
  whip?: number;
  so?: number;
  bb?: number;
  h?: number;
  hr?: number;
  players?: Array<Record<string, unknown>>;
};

export type TeamFieldingStats = {
  team_id?: string;
  year?: number;
  league?: string;
  g?: number;
  e?: number;
  a?: number;
  po?: number;
  fpct?: number;
  dp?: number;
  players?: Array<Record<string, unknown>>;
};

export type RunDifferentialSeries = {
  season?: number;
  games_played?: number;
  runs_scored?: number;
  runs_allowed?: number;
  run_differential?: number;
  games?: Array<Record<string, unknown>>;
};

export type TeamSeasonProfile = {
  id: string;
  name?: string;
  year?: number;
  league?: string;
  division?: string;
  wins?: number;
  losses?: number;
  rank?: number;
  park?: string;
  park_id?: string;
};
