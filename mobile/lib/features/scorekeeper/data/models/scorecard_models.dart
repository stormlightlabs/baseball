enum ScorecardStatus { inProgress, finalGame }

enum ScorecardStatusFilter { all, inProgress, finalGame }

enum ScorecardHalfInning { top, bottom }

enum ScorecardTeam { away, home }

class ScorecardGameSummary {
  const ScorecardGameSummary({
    required this.uuid,
    required this.awayTeamName,
    required this.awayTeamAbbreviation,
    required this.homeTeamName,
    required this.homeTeamAbbreviation,
    required this.venue,
    required this.gameDate,
    required this.status,
    required this.awayScore,
    required this.homeScore,
    required this.pitchCount,
    required this.lastModifiedAt,
  });

  final String uuid;
  final String awayTeamName;
  final String awayTeamAbbreviation;
  final String homeTeamName;
  final String homeTeamAbbreviation;
  final String? venue;
  final DateTime gameDate;
  final ScorecardStatus status;
  final int awayScore;
  final int homeScore;
  final int pitchCount;
  final DateTime lastModifiedAt;
}

class ScorecardLineupSlot {
  const ScorecardLineupSlot({
    required this.team,
    required this.battingOrder,
    required this.playerName,
    required this.positionCode,
  });

  final ScorecardTeam team;
  final int battingOrder;
  final String playerName;
  final String? positionCode;
}

class ScorecardPlayEntry {
  const ScorecardPlayEntry({
    required this.id,
    required this.gameUuid,
    required this.inningId,
    required this.battingTeam,
    required this.sequence,
    required this.batterIndex,
    required this.batterName,
    required this.batterPosition,
    required this.outcomeCode,
    required this.putoutSequence,
    required this.pitchLog,
    required this.baseStateBefore,
    required this.baseStateAfter,
    required this.rbi,
    required this.scored,
    required this.recordedAt,
  });

  final int id;
  final String gameUuid;
  final int inningId;
  final ScorecardTeam battingTeam;
  final int sequence;
  final int batterIndex;
  final String batterName;
  final String? batterPosition;
  final String outcomeCode;
  final String? putoutSequence;
  final List<String> pitchLog;
  final String baseStateBefore;
  final String baseStateAfter;
  final int rbi;
  final bool scored;
  final DateTime recordedAt;
}

class ScorecardInning {
  const ScorecardInning({
    required this.id,
    required this.gameUuid,
    required this.inningNumber,
    required this.half,
    required this.plays,
  });

  final int id;
  final String gameUuid;
  final int inningNumber;
  final ScorecardHalfInning half;
  final List<ScorecardPlayEntry> plays;
}

class ScorecardGameDetail {
  const ScorecardGameDetail({required this.summary, required this.lineups, required this.innings});

  final ScorecardGameSummary summary;
  final List<ScorecardLineupSlot> lineups;
  final List<ScorecardInning> innings;
}

class ScorecardGameDraft {
  const ScorecardGameDraft({
    required this.uuid,
    required this.awayTeamName,
    required this.awayTeamAbbreviation,
    required this.homeTeamName,
    required this.homeTeamAbbreviation,
    required this.venue,
    required this.gameDate,
    required this.status,
    required this.lineups,
  });

  final String uuid;
  final String awayTeamName;
  final String awayTeamAbbreviation;
  final String homeTeamName;
  final String homeTeamAbbreviation;
  final String? venue;
  final DateTime gameDate;
  final ScorecardStatus status;
  final List<ScorecardLineupSlot> lineups;
}

class ScorecardPlayDraft {
  const ScorecardPlayDraft({
    required this.gameUuid,
    required this.inningNumber,
    required this.half,
    required this.battingTeam,
    required this.batterIndex,
    required this.batterName,
    required this.batterPosition,
    required this.outcomeCode,
    required this.putoutSequence,
    required this.pitchLog,
    required this.baseStateBefore,
    required this.baseStateAfter,
    required this.rbi,
    required this.scored,
    required this.nextState,
  });

  final String gameUuid;
  final int inningNumber;
  final ScorecardHalfInning half;
  final ScorecardTeam battingTeam;
  final int batterIndex;
  final String batterName;
  final String? batterPosition;
  final String outcomeCode;
  final String? putoutSequence;
  final List<String> pitchLog;
  final String baseStateBefore;
  final String baseStateAfter;
  final int rbi;
  final bool scored;
  final ScorecardGameRuntimeState nextState;
}

class ScorecardGameRuntimeState {
  const ScorecardGameRuntimeState({
    required this.currentInning,
    required this.currentHalf,
    required this.awayScore,
    required this.homeScore,
    required this.pitchCount,
    required this.outs,
    required this.balls,
    required this.strikes,
    required this.awayBatterIndex,
    required this.homeBatterIndex,
  });

  final int currentInning;
  final ScorecardHalfInning currentHalf;
  final int awayScore;
  final int homeScore;
  final int pitchCount;
  final int outs;
  final int balls;
  final int strikes;
  final int awayBatterIndex;
  final int homeBatterIndex;
}
