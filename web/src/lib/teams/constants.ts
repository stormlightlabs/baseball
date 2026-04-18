export const MAIN_TABS = [
  { id: 'overview', label: 'Overview' },
  { id: 'roster', label: 'Roster' },
  { id: 'batting', label: 'Batting' },
  { id: 'pitching', label: 'Pitching' },
  { id: 'fielding', label: 'Fielding' },
  { id: 'schedule', label: 'Schedule' },
  { id: 'run-diff', label: 'Run Diff.' }
] as const;

export const ALL_TEAM_TABS = [...MAIN_TABS] as const;

export type TeamTabId = (typeof ALL_TEAM_TABS)[number]['id'];

export const DEFAULT_TEAM_TAB: TeamTabId = 'overview';

export function isTeamTabId(value: string): value is TeamTabId {
  return ALL_TEAM_TABS.some((tab) => tab.id === value);
}
