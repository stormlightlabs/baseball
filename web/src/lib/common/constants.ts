export const MAIN_TEAM_TABS = [
  { id: 'overview', label: 'Overview' },
  { id: 'roster', label: 'Roster' },
  { id: 'batting', label: 'Batting' },
  { id: 'pitching', label: 'Pitching' },
  { id: 'fielding', label: 'Fielding' },
  { id: 'schedule', label: 'Schedule' },
  { id: 'daily-trends', label: 'Daily Trends' },
  { id: 'run-diff', label: 'Run Diff.' }
] as const;

export const ALL_TEAM_TABS = [...MAIN_TEAM_TABS] as const;

export type TeamTabId = (typeof ALL_TEAM_TABS)[number]['id'];

export const DEFAULT_TEAM_TAB: TeamTabId = 'overview';

export function isTeamTabId(value: string): value is TeamTabId {
  return ALL_TEAM_TABS.some((tab) => tab.id === value);
}

export const MAIN_GAME_TABS = [
  { id: 'overview', label: 'Overview' },
  { id: 'events', label: 'Events' },
  { id: 'plays', label: 'Plays + Pitches' },
  { id: 'win-prob', label: 'Win Prob.' }
] as const;

export const ALL_GAME_TABS = [...MAIN_GAME_TABS] as const;

export type GameTabId = (typeof ALL_GAME_TABS)[number]['id'];

export const DEFAULT_GAME_TAB: GameTabId = 'overview';

export function isGameTabId(value: string): value is GameTabId {
  return ALL_GAME_TABS.some((tab) => tab.id === value);
}

export const BATTING_STATS = [
  { value: 'hr', label: 'Home Runs (HR)' },
  { value: 'avg', label: 'Batting Avg (AVG)' },
  { value: 'rbi', label: 'RBI' },
  { value: 'sb', label: 'Stolen Bases (SB)' },
  { value: 'obp', label: 'OBP' },
  { value: 'slg', label: 'SLG' },
  { value: 'ops', label: 'OPS' }
] as const;

export const BATTING_RATE_STATS = new Set(['avg', 'obp', 'slg', 'ops']);

export const CHART_SCALES = {
  x: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } },
  y: { ticks: { color: '#6b7280', font: { size: 9 } }, grid: { color: '#252934' } }
} as const;

export const TEAMS_COLUMNS = [
  { key: 'year', label: 'Year', sortable: true },
  { key: 'team', label: 'Team', sortable: true },
  { key: 'league', label: 'League', sortable: true },
  { key: 'g', label: 'G', sortable: true }
];

export const SALARIES_COLUMNS = [
  { key: 'year', label: 'Year', sortable: true },
  { key: 'team', label: 'Team', sortable: true },
  {
    key: 'salary',
    label: 'Salary',
    sortable: true,
    rank: true,
    format: (v: unknown) => `$${Number(v).toLocaleString()}`
  }
];

export const ADVANCED_PLAYER_TABS = [
  { id: 'batting-adv', label: 'Batting Adv.' },
  { id: 'pitching-adv', label: 'Pitching Adv.' },
  { id: 'war', label: 'WAR' },
  { id: 'splits', label: 'Splits' },
  { id: 'streaks', label: 'Streaks' }
] as const;

export const MAIN_PLAYER_TABS = [
  { id: 'batting', label: 'Batting' },
  { id: 'pitching', label: 'Pitching' },
  { id: 'game-logs', label: 'Game Logs' },
  { id: 'awards', label: 'Awards' },
  { id: 'hof', label: 'Hall of Fame' },
  { id: 'teams', label: 'Teams' },
  { id: 'salaries', label: 'Salaries' },
  { id: 'relatives', label: 'Relatives' }
] as const;

export const ALL_PLAYER_TABS = [...MAIN_PLAYER_TABS, ...ADVANCED_PLAYER_TABS] as const;

export type PlayerTabId = (typeof ALL_PLAYER_TABS)[number]['id'];

export const DEFAULT_PLAYER_TAB: PlayerTabId = 'batting';

export const ADVANCED_PLAYER_TAB_IDS = new Set<string>(ADVANCED_PLAYER_TABS.map((tab) => tab.id));

export function isPlayerTabId(value: string): value is PlayerTabId {
  return ALL_PLAYER_TABS.some((tab) => tab.id === value);
}
