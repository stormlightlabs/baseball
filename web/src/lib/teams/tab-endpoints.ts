import { EP } from '$lib/endpoints';
import type { TeamTabId } from './constants';

export function endpointForTeamTab(teamId: string, tabId: TeamTabId, year?: string | number | null): string {
  switch (tabId) {
    case 'overview': {
      return year ? EP.team(teamId) : EP.franchise(teamId);
    }
    case 'roster': {
      return year ? EP.seasonTeamRoster(year, teamId) : EP.team(teamId);
    }
    case 'batting': {
      return year ? EP.seasonTeamBatting(year, teamId) : EP.team(teamId);
    }
    case 'pitching': {
      return year ? EP.seasonTeamPitching(year, teamId) : EP.team(teamId);
    }
    case 'fielding': {
      return year ? EP.seasonTeamFielding(year, teamId) : EP.team(teamId);
    }
    case 'schedule': {
      return year ? EP.seasonTeamSchedule(year, teamId) : EP.team(teamId);
    }
    case 'daily-trends': {
      return year ? EP.teamDailyStats(teamId) : EP.team(teamId);
    }
    case 'run-diff': {
      return year ? EP.teamRunDifferential(teamId) : EP.team(teamId);
    }
  }
}
