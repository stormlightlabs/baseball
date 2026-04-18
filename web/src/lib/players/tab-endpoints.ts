import { EP } from '$lib/endpoints';
import type { PlayerTabId } from '$lib/players/constants';

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
