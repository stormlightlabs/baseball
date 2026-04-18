import { EP } from '$lib/endpoints';
import type { GameTabId, PlayerTabId, TeamTabId } from '$lib/common/constants';

export type GameLogType = 'batting' | 'pitching' | 'fielding';

export function parseGameLogType(raw: string | null | undefined): GameLogType {
  if (raw === 'pitching' || raw === 'fielding') return raw;
  return 'batting';
}

export function gameLogEndpoint(kind: GameLogType, playerId: string): string {
  switch (kind) {
    case 'pitching': {
      return EP.playerGameLogsPitching(playerId);
    }
    case 'fielding': {
      return EP.playerGameLogsFielding(playerId);
    }
    default: {
      return EP.playerGameLogsBatting(playerId);
    }
  }
}

export function endpointForPlayerTab(playerId: string, tabId: PlayerTabId, gameLogType: GameLogType): string {
  switch (tabId) {
    case 'batting': {
      return EP.playerStatsBatting(playerId);
    }
    case 'pitching': {
      return EP.playerStatsPitching(playerId);
    }
    case 'game-logs': {
      return gameLogEndpoint(gameLogType, playerId);
    }
    case 'awards': {
      return EP.playerAwards(playerId);
    }
    case 'hof': {
      return EP.playerHallOfFame(playerId);
    }
    case 'teams': {
      return EP.playerTeams(playerId);
    }
    case 'salaries': {
      return EP.playerSalaries(playerId);
    }
    case 'relatives': {
      return EP.playerRelatives(playerId);
    }
    case 'batting-adv': {
      return EP.playerStatsBattingAdv(playerId);
    }
    case 'pitching-adv': {
      return EP.playerStatsPitchingAdv(playerId);
    }
    case 'war': {
      return EP.playerStatsWar(playerId);
    }
    case 'splits': {
      return EP.playerSplits(playerId);
    }
    case 'streaks': {
      return EP.playerStreaks(playerId);
    }
  }
}

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

export function endpointForGameTab(gameId: string, tabId: GameTabId): string {
  switch (tabId) {
    case 'overview': {
      return EP.game(gameId);
    }
    case 'events': {
      return EP.gameEvents(gameId);
    }
    case 'plays': {
      return EP.gamePlays(gameId);
    }
    case 'win-prob': {
      return EP.gameWinProb(gameId);
    }
  }
}
