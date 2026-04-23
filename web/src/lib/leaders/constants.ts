export const QUICK_BATTING_STATS = ['hr', 'avg', 'rbi', 'sb', 'h', 'r'] as const;
export const QUICK_PITCHING_STATS = ['era', 'so', 'w', 'sv', 'ip'] as const;

export const CAREER_BATTING_STATS = ['hr', 'avg', 'rbi', 'sb', 'h', 'r', 'ops'] as const;
export const CAREER_PITCHING_STATS = ['w', 'era', 'so', 'sv', 'ip'] as const;

export const ADV_BATTING_STATS = ['WRC_PLUS', 'WOBA', 'OPS', 'ISO', 'BABIP', 'HR', 'BB', 'K_RATE'] as const;
export const ADV_PITCHING_STATS = ['FIP', 'ERA', 'WHIP', 'K_PER_9', 'BB_PER_9', 'HR_PER_9', 'SO'] as const;

export type LabDataset =
  | 'stats_batting'
  | 'stats_pitching'
  | 'stats_fielding'
  | 'stats_teams_batting'
  | 'stats_teams_pitching'
  | 'stats_teams_fielding';

export function sortForLabDataset(dataset: LabDataset): string {
  switch (dataset) {
    case 'stats_batting': {
      return 'hr';
    }
    case 'stats_pitching': {
      return 'so';
    }
    case 'stats_fielding': {
      return 'po';
    }
    case 'stats_teams_batting': {
      return 'hr';
    }
    case 'stats_teams_pitching': {
      return 'era';
    }
    default: {
      return 'po';
    }
  }
}

export function labelForLabDataset(dataset: LabDataset): string {
  switch (dataset) {
    case 'stats_batting': {
      return 'Player batting';
    }
    case 'stats_pitching': {
      return 'Player pitching';
    }
    case 'stats_fielding': {
      return 'Player fielding';
    }
    case 'stats_teams_batting': {
      return 'Team batting';
    }
    case 'stats_teams_pitching': {
      return 'Team pitching';
    }
    default: {
      return 'Team fielding';
    }
  }
}
