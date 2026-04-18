export type EntityType = { label: string; path: '/players' | '/teams' | '/games' | '/seasons'; apiEndpoint: string };

export const ENTITY_TYPES: EntityType[] = [
  { label: 'Players', path: '/players', apiEndpoint: '/v1/search/players' },
  { label: 'Teams', path: '/teams', apiEndpoint: '/v1/search/teams' },
  { label: 'Franchises', path: '/teams', apiEndpoint: '/v1/search/teams' },
  { label: 'Games', path: '/games', apiEndpoint: '/v1/search/games' },
  { label: 'Parks', path: '/seasons', apiEndpoint: '/v1/search/parks' }
];

export const QUICK_LINKS = [
  { label: 'Players', path: '/players', hint: '/v1/players', desc: 'Career stats, batting, pitching, awards' },
  { label: 'Teams', path: '/teams', hint: '/v1/franchises', desc: 'Franchise history and team-season records' },
  { label: 'Games', path: '/games', hint: '/v1/games', desc: 'Game finder with advanced filters' },
  {
    label: 'Leaders',
    path: '/leaders',
    hint: '/v1/seasons/{year}/leaders/{type}',
    desc: 'Stat leaderboards by season or career'
  },
  {
    label: 'Seasons',
    path: '/seasons',
    hint: '/v1/seasons/{year}/teams',
    desc: 'Season hub: teams, schedule, date explorer'
  }
] as const;

export type FeaturedQueryGroup = 'standard' | 'derived' | 'historical';

export type FeaturedQuery = { title: string; endpoint: string; group: FeaturedQueryGroup };

export const FEATURED_QUERIES: FeaturedQuery[] = [
  { title: 'HR leaders in 1927', endpoint: '/v1/seasons/1927/leaders/batting?stat=hr', group: 'standard' },
  { title: 'Career HR leaders (min 3000 AB)', endpoint: '/v1/stats/batting?sort_by=hr&min_ab=3000', group: 'standard' },
  {
    title: 'ERA leaders — Year of the Pitcher (1968)',
    endpoint: '/v1/seasons/1968/leaders/pitching?stat=era',
    group: 'standard'
  },
  { title: 'Win expectancy — bases loaded, 2 outs', endpoint: '/v1/win-expectancy?runners=7&outs=2', group: 'derived' },
  { title: 'WAR leaders by season (Statcast era)', endpoint: '/v1/seasons/2019/leaders/war', group: 'derived' },
  {
    title: 'Run differential — 1998 Yankees',
    endpoint: '/v1/teams/NYA/run-differential?season=1998',
    group: 'derived'
  },
  { title: 'Federal League games (1914)', endpoint: '/v1/federalleague/games?season=1914', group: 'historical' },
  { title: 'Negro Leagues teams', endpoint: '/v1/negroleagues/teams', group: 'historical' },
  { title: 'Extra-inning games in 2023', endpoint: '/v1/games?season=2023&min_innings=10', group: 'standard' },
  {
    title: 'Most saves in a season (all-time)',
    endpoint: '/v1/stats/pitching?sort_by=sv&sort_order=desc&min_ip=1',
    group: 'standard'
  },
  { title: 'Batting advanced stats — 2016', endpoint: '/v1/seasons/2016/leaders/batting/advanced', group: 'derived' },
  { title: 'Win expectancy eras (dynamic)', endpoint: '/v1/win-expectancy/eras', group: 'derived' }
];

export const FEATURED_GROUPS: Array<{ key: FeaturedQueryGroup; label: string }> = [
  { key: 'standard', label: 'Standard stats' },
  { key: 'derived', label: 'Derived / computed' },
  { key: 'historical', label: 'Historical leagues' }
];

export const ALL_ENDPOINTS = [
  '/v1/players',
  '/v1/players/{id}/seasons',
  '/v1/players/{id}/awards',
  '/v1/teams',
  '/v1/franchises',
  '/v1/games',
  '/v1/seasons/{year}/teams',
  '/v1/seasons/{year}/leaders/{type}',
  '/v1/stats/batting',
  '/v1/stats/pitching',
  '/v1/win-expectancy',
  '/v1/federalleague/games',
  '/v1/negroleagues/games',
  '/v1/meta'
] as const;

export const SOURCE_COLORS: Record<string, string> = { lahman: '#3b82f6', retrosheet: '#10b981' };

export type DatasetUiHints = { href?: string; tooltip?: string };

export const DATASET_UI_HINTS: Record<string, DatasetUiHints> = {
  fangraphs_constants: {
    href: 'https://www.fangraphs.com/tools/guts',
    tooltip: 'FanGraphs Guts constants used for advanced stat calculations.'
  },
  salary_summary: { tooltip: 'Enriched data: normalized salary rollups derived from multiple salary inputs.' }
};
