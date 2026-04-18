import type { PaginatedResponse } from '$lib/api';

export type GameFinderMode = 'filters' | 'nl';

export type GameFamily = 'mlb' | 'fed' | 'nlg';

export type GameRecord = {
  id: string;
  season?: number;
  date?: string;
  day_of_week?: string;
  home_team?: string;
  away_team?: string;
  home_league?: string;
  away_league?: string;
  home_score?: number;
  away_score?: number;
  innings?: number;
  attendance?: number;
  duration_min?: number;
  park_id?: string;
  park_name?: string;
  park_city?: string;
  park_state?: string;
  day_night?: string;
  is_postseason?: boolean;
  winning_pitcher_name?: string;
  losing_pitcher_name?: string;
  save_pitcher_name?: string;
  home_manager_name?: string;
  away_manager_name?: string;
};

export type TeamGameStats = {
  ab?: number;
  h?: number;
  r?: number;
  doubles?: number;
  triples?: number;
  hr?: number;
  rbi?: number;
  sh?: number;
  sf?: number;
  hbp?: number;
  bb?: number;
  ibb?: number;
  so?: number;
  sb?: number;
  cs?: number;
  gdp?: number;
  lob?: number;
  pitchers_used?: number;
  er?: number;
  wp?: number;
  balks?: number;
  po?: number;
  a?: number;
  e?: number;
  pb?: number;
  dp?: number;
  tp?: number;
};

export type LineupPlayer = { player_id?: string; name?: string; position?: number; bats?: string; throws?: string };

export type GameBoxscore = {
  game_id?: string;
  date?: string;
  home_team?: string;
  away_team?: string;
  home_score?: number;
  away_score?: number;
  home_stats?: TeamGameStats;
  away_stats?: TeamGameStats;
  home_lineup: LineupPlayer[];
  away_lineup: LineupPlayer[];
};

export type GameSummary = {
  game_id?: string;
  date?: string;
  home_team?: string;
  away_team?: string;
  home_score?: number;
  away_score?: number;
  innings?: number;
  winner?: string;
  home_lineup: LineupPlayer[];
  away_lineup: LineupPlayer[];
};

export type PlayEvent = {
  game_id?: string;
  play_num?: number;
  inning?: number;
  top_bot?: number;
  bat_team?: string;
  pit_team?: string;
  date?: string;
  batter?: string;
  batter_name?: string;
  pitcher?: string;
  pitcher_name?: string;
  score_vis?: number;
  score_home?: number;
  outs_pre?: number;
  outs_post?: number;
  balls?: number;
  strikes?: number;
  event?: string;
  runs?: number;
  rbi?: number;
};

export type PitchEvent = {
  game_id?: string;
  play_num?: number;
  inning?: number;
  top_bot?: number;
  bat_team?: string;
  pit_team?: string;
  date?: string;
  batter?: string;
  batter_name?: string;
  pitcher?: string;
  pitcher_name?: string;
  outs_pre?: number;
  seq_num?: number;
  pitch_type?: string;
  ball_count?: number;
  strike_count?: number;
  is_in_play?: boolean;
  is_strike?: boolean;
  is_ball?: boolean;
  description?: string;
  event?: string;
};

export type WinProbabilityPoint = {
  event_index?: number;
  inning?: number;
  top_of_inning?: boolean;
  home_score?: number;
  away_score?: number;
  outs?: number;
  bases?: string;
  home_win_prob?: number;
  away_win_prob?: number;
  description?: string;
};

export type WinProbabilityCurve = {
  game_id?: string;
  season?: number;
  home_team?: string;
  away_team?: string;
  points: WinProbabilityPoint[];
};

export type PlateAppearanceLeverage = {
  game_id?: string;
  event_id?: number;
  batter_id?: string;
  pitcher_id?: string;
  inning?: number;
  top_of_inning?: boolean;
  home_score_before?: number;
  away_score_before?: number;
  outs_before?: number;
  bases_before?: string;
  we_before?: number;
  we_after?: number;
  li?: number;
  we_change?: number;
  description?: string;
};

export type GameWinProbabilitySummary = {
  game_id?: string;
  season?: number;
  home_team?: string;
  away_team?: string;
  home_win_prob_start?: number;
  home_win_prob_end?: number;
  biggest_positive_swing?: PlateAppearanceLeverage;
  biggest_negative_swing?: PlateAppearanceLeverage;
};

export type GamePaginatedResponse<T> = PaginatedResponse<T>;
