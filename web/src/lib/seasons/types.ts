export type SeasonSummary = { year: number; leagues: string[]; team_count?: number; game_count?: number };

export type SeasonTeam = {
  team_id: string;
  name?: string;
  franchise_id?: string;
  league?: string;
  division?: string;
  year?: number;
  wins?: number;
  losses?: number;
  ties?: number;
  games?: number;
  runs_scored?: number;
  runs_allowed?: number;
  attendance?: number;
};

export type SeasonBattingLeader = {
  player_id: string;
  year?: number;
  team_id?: string;
  league?: string;
  g?: number;
  pa?: number;
  ab?: number;
  r?: number;
  h?: number;
  hr?: number;
  rbi?: number;
  sb?: number;
  avg?: number;
  obp?: number;
  slg?: number;
  ops?: number;
};

export type SeasonPitchingLeader = {
  player_id: string;
  year?: number;
  team_id?: string;
  league?: string;
  w?: number;
  l?: number;
  sv?: number;
  so?: number;
  era?: number;
  whip?: number;
  ip_outs?: number;
  g?: number;
};

export type SeasonGame = {
  id: string;
  date?: string;
  season?: number;
  home_team?: string;
  away_team?: string;
  home_score?: number;
  away_score?: number;
  innings?: number;
  is_postseason?: boolean;
  park_id?: string;
  park_name?: string;
};

export type SeasonAwardResult = {
  award_id?: string;
  player_id?: string;
  year?: number;
  league?: string;
  votes_first?: number;
  points?: number;
  rank?: number;
};

export type SeasonPostseasonSeries = {
  year?: number;
  round?: string;
  winner_team?: string;
  winner_league?: string;
  loser_team?: string;
  loser_league?: string;
  wins?: number;
  losses?: number;
  ties?: number;
};

export type SeasonParkFactor = {
  season?: number;
  park_id?: string;
  provider?: string;
  runs_factor?: number;
  hr_factor?: number;
  h_factor?: number;
  bb_factor?: number;
  runs_factor_lhb?: number;
  runs_factor_rhb?: number;
  games_sampled?: number;
  multi_year?: boolean;
};
