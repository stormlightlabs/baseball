import 'dart:convert';

import 'package:bigfly_mobile/core/data/local/app_database.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:drift/drift.dart';

abstract class ScorecardRepository {
  Future<void> createGame(ScorecardGameDraft draft);
  Future<List<ScorecardGameSummary>> listGames({ScorecardStatusFilter filter = ScorecardStatusFilter.all});
  Future<ScorecardGameDetail?> getGame(String uuid);
  Future<void> appendPlay(ScorecardPlayDraft draft);
  Future<bool> undoLastPlay({required String gameUuid, required ScorecardGameRuntimeState restoredState});
}

class DriftScorecardRepository implements ScorecardRepository {
  DriftScorecardRepository(this._database);

  final AppDatabase _database;

  @override
  Future<void> createGame(ScorecardGameDraft draft) async {
    await _database.transaction(() async {
      final now = DateTime.now();
      await _database
          .into(_database.scorecardGames)
          .insert(
            ScorecardGamesCompanion.insert(
              uuid: draft.uuid,
              awayTeamName: draft.awayTeamName,
              awayTeamAbbreviation: draft.awayTeamAbbreviation,
              homeTeamName: draft.homeTeamName,
              homeTeamAbbreviation: draft.homeTeamAbbreviation,
              venue: Value(draft.venue),
              gameDate: draft.gameDate,
              status: _toDbStatus(draft.status),
              currentHalf: _toDbHalf(ScorecardHalfInning.top),
              createdAt: Value(now),
              lastModifiedAt: Value(now),
            ),
          );

      for (final slot in draft.lineups) {
        await _database
            .into(_database.scorecardLineups)
            .insert(
              ScorecardLineupsCompanion.insert(
                gameUuid: draft.uuid,
                teamSide: _toDbTeam(slot.team),
                battingOrder: slot.battingOrder,
                playerName: slot.playerName,
                positionCode: Value(slot.positionCode),
              ),
            );
      }
    });
  }

  @override
  Future<List<ScorecardGameSummary>> listGames({ScorecardStatusFilter filter = ScorecardStatusFilter.all}) async {
    final query = _database.select(_database.scorecardGames)
      ..orderBy(<OrderingTerm Function($ScorecardGamesTable)>[(tbl) => OrderingTerm.desc(tbl.lastModifiedAt)]);

    switch (filter) {
      case ScorecardStatusFilter.all:
        break;
      case ScorecardStatusFilter.inProgress:
        query.where((tbl) => tbl.status.equals(_toDbStatus(ScorecardStatus.inProgress)));
      case ScorecardStatusFilter.finalGame:
        query.where((tbl) => tbl.status.equals(_toDbStatus(ScorecardStatus.finalGame)));
    }

    final rows = await query.get();
    return rows.map(_toSummary).toList(growable: false);
  }

  @override
  Future<ScorecardGameDetail?> getGame(String uuid) async {
    final game = await (_database.select(
      _database.scorecardGames,
    )..where((tbl) => tbl.uuid.equals(uuid))).getSingleOrNull();
    if (game == null) {
      return null;
    }

    final lineups =
        await (_database.select(_database.scorecardLineups)
              ..where((tbl) => tbl.gameUuid.equals(uuid))
              ..orderBy(<OrderingTerm Function($ScorecardLineupsTable)>[
                (tbl) => OrderingTerm.asc(tbl.teamSide),
                (tbl) => OrderingTerm.asc(tbl.battingOrder),
              ]))
            .get();

    final innings =
        await (_database.select(_database.scorecardInnings)
              ..where((tbl) => tbl.gameUuid.equals(uuid))
              ..orderBy(<OrderingTerm Function($ScorecardInningsTable)>[
                (tbl) => OrderingTerm.asc(tbl.inningNumber),
                (tbl) => OrderingTerm.asc(tbl.half),
              ]))
            .get();

    final plays =
        await (_database.select(_database.scorecardPlays)
              ..where((tbl) => tbl.gameUuid.equals(uuid))
              ..orderBy(<OrderingTerm Function($ScorecardPlaysTable)>[(tbl) => OrderingTerm.asc(tbl.sequence)]))
            .get();

    final playsByInningId = <int, List<ScorecardPlayEntry>>{};
    for (final row in plays) {
      playsByInningId.putIfAbsent(row.inningId, () => <ScorecardPlayEntry>[]).add(_toPlay(row));
    }

    return ScorecardGameDetail(
      summary: _toSummary(game),
      lineups: lineups
          .map(
            (row) => ScorecardLineupSlot(
              team: _fromDbTeam(row.teamSide),
              battingOrder: row.battingOrder,
              playerName: row.playerName,
              positionCode: row.positionCode,
            ),
          )
          .toList(growable: false),
      innings: innings
          .map(
            (row) => ScorecardInning(
              id: row.id,
              gameUuid: row.gameUuid,
              inningNumber: row.inningNumber,
              half: _fromDbHalf(row.half),
              plays: List<ScorecardPlayEntry>.unmodifiable(playsByInningId[row.id] ?? const <ScorecardPlayEntry>[]),
            ),
          )
          .toList(growable: false),
    );
  }

  @override
  Future<void> appendPlay(ScorecardPlayDraft draft) async {
    await _database.transaction(() async {
      final game = await (_database.select(
        _database.scorecardGames,
      )..where((tbl) => tbl.uuid.equals(draft.gameUuid))).getSingleOrNull();
      if (game == null) {
        throw StateError('Scorecard game ${draft.gameUuid} was not found');
      }

      final inning =
          await (_database.select(_database.scorecardInnings)..where(
                (tbl) =>
                    tbl.gameUuid.equals(draft.gameUuid) &
                    tbl.inningNumber.equals(draft.inningNumber) &
                    tbl.half.equals(_toDbHalf(draft.half)),
              ))
              .getSingleOrNull();

      final inningId =
          inning?.id ??
          await _database
              .into(_database.scorecardInnings)
              .insert(
                ScorecardInningsCompanion.insert(
                  gameUuid: draft.gameUuid,
                  inningNumber: draft.inningNumber,
                  half: _toDbHalf(draft.half),
                ),
              );

      final sequence = (game.lastPlaySequence ?? 0) + 1;
      await _database
          .into(_database.scorecardPlays)
          .insert(
            ScorecardPlaysCompanion.insert(
              gameUuid: draft.gameUuid,
              inningId: inningId,
              battingTeam: _toDbTeam(draft.battingTeam),
              sequence: sequence,
              batterIndex: draft.batterIndex,
              batterName: draft.batterName,
              batterPosition: Value(draft.batterPosition),
              outcomeCode: draft.outcomeCode,
              putoutSequence: Value(draft.putoutSequence),
              pitchLogJson: jsonEncode(draft.pitchLog),
              baseStateBefore: draft.baseStateBefore,
              baseStateAfter: draft.baseStateAfter,
              rbi: draft.rbi,
              scored: draft.scored,
              recordedAt: Value(DateTime.now()),
            ),
          );

      await (_database.update(_database.scorecardGames)..where((tbl) => tbl.uuid.equals(draft.gameUuid))).write(
        ScorecardGamesCompanion(
          awayScore: Value(draft.nextState.awayScore),
          homeScore: Value(draft.nextState.homeScore),
          pitchCount: Value(draft.nextState.pitchCount),
          currentInning: Value(draft.nextState.currentInning),
          currentHalf: Value(_toDbHalf(draft.nextState.currentHalf)),
          outs: Value(draft.nextState.outs),
          balls: Value(draft.nextState.balls),
          strikes: Value(draft.nextState.strikes),
          awayBatterIndex: Value(draft.nextState.awayBatterIndex),
          homeBatterIndex: Value(draft.nextState.homeBatterIndex),
          lastPlaySequence: Value(sequence),
          lastModifiedAt: Value(DateTime.now()),
        ),
      );
    });
  }

  @override
  Future<bool> undoLastPlay({required String gameUuid, required ScorecardGameRuntimeState restoredState}) async {
    return _database.transaction(() async {
      final game = await (_database.select(
        _database.scorecardGames,
      )..where((tbl) => tbl.uuid.equals(gameUuid))).getSingleOrNull();
      if (game == null || game.lastPlaySequence == null) {
        return false;
      }

      final lastSequence = game.lastPlaySequence!;
      final play = await (_database.select(
        _database.scorecardPlays,
      )..where((tbl) => tbl.gameUuid.equals(gameUuid) & tbl.sequence.equals(lastSequence))).getSingleOrNull();
      if (play == null) {
        return false;
      }

      await (_database.delete(_database.scorecardPlays)..where((tbl) => tbl.id.equals(play.id))).go();
      final remaining =
          await (_database.select(_database.scorecardPlays)
                ..where((tbl) => tbl.gameUuid.equals(gameUuid))
                ..orderBy(<OrderingTerm Function($ScorecardPlaysTable)>[(tbl) => OrderingTerm.desc(tbl.sequence)])
                ..limit(1))
              .getSingleOrNull();

      final hasInningPlays = await (_database.select(
        _database.scorecardPlays,
      )..where((tbl) => tbl.inningId.equals(play.inningId))).getSingleOrNull();
      if (hasInningPlays == null) {
        await (_database.delete(_database.scorecardInnings)..where((tbl) => tbl.id.equals(play.inningId))).go();
      }

      await (_database.update(_database.scorecardGames)..where((tbl) => tbl.uuid.equals(gameUuid))).write(
        ScorecardGamesCompanion(
          awayScore: Value(restoredState.awayScore),
          homeScore: Value(restoredState.homeScore),
          pitchCount: Value(restoredState.pitchCount),
          currentInning: Value(restoredState.currentInning),
          currentHalf: Value(_toDbHalf(restoredState.currentHalf)),
          outs: Value(restoredState.outs),
          balls: Value(restoredState.balls),
          strikes: Value(restoredState.strikes),
          awayBatterIndex: Value(restoredState.awayBatterIndex),
          homeBatterIndex: Value(restoredState.homeBatterIndex),
          lastPlaySequence: Value(remaining?.sequence),
          lastModifiedAt: Value(DateTime.now()),
        ),
      );

      return true;
    });
  }

  ScorecardGameSummary _toSummary(ScorecardGame row) {
    return ScorecardGameSummary(
      uuid: row.uuid,
      awayTeamName: row.awayTeamName,
      awayTeamAbbreviation: row.awayTeamAbbreviation,
      homeTeamName: row.homeTeamName,
      homeTeamAbbreviation: row.homeTeamAbbreviation,
      venue: row.venue,
      gameDate: row.gameDate,
      status: _fromDbStatus(row.status),
      awayScore: row.awayScore,
      homeScore: row.homeScore,
      pitchCount: row.pitchCount,
      lastModifiedAt: row.lastModifiedAt,
    );
  }

  ScorecardPlayEntry _toPlay(ScorecardPlayRow row) {
    final decoded = jsonDecode(row.pitchLogJson);
    final pitchLog = decoded is List<dynamic> ? decoded.whereType<String>().toList(growable: false) : <String>[];

    return ScorecardPlayEntry(
      id: row.id,
      gameUuid: row.gameUuid,
      inningId: row.inningId,
      battingTeam: _fromDbTeam(row.battingTeam),
      sequence: row.sequence,
      batterIndex: row.batterIndex,
      batterName: row.batterName,
      batterPosition: row.batterPosition,
      outcomeCode: row.outcomeCode,
      putoutSequence: row.putoutSequence,
      pitchLog: pitchLog,
      baseStateBefore: row.baseStateBefore,
      baseStateAfter: row.baseStateAfter,
      rbi: row.rbi,
      scored: row.scored,
      recordedAt: row.recordedAt,
    );
  }

  String _toDbStatus(ScorecardStatus status) {
    return switch (status) {
      ScorecardStatus.inProgress => ScorecardGameStatus.inProgress.name,
      ScorecardStatus.finalGame => ScorecardGameStatus.finalGame.name,
    };
  }

  ScorecardStatus _fromDbStatus(String status) {
    if (status == ScorecardGameStatus.finalGame.name) {
      return ScorecardStatus.finalGame;
    }
    return ScorecardStatus.inProgress;
  }

  String _toDbHalf(ScorecardHalfInning half) {
    return switch (half) {
      ScorecardHalfInning.top => ScorecardHalf.top.name,
      ScorecardHalfInning.bottom => ScorecardHalf.bottom.name,
    };
  }

  ScorecardHalfInning _fromDbHalf(String half) {
    if (half == ScorecardHalf.bottom.name) {
      return ScorecardHalfInning.bottom;
    }
    return ScorecardHalfInning.top;
  }

  String _toDbTeam(ScorecardTeam team) {
    return switch (team) {
      ScorecardTeam.away => ScorecardTeamSide.away.name,
      ScorecardTeam.home => ScorecardTeamSide.home.name,
    };
  }

  ScorecardTeam _fromDbTeam(String team) {
    if (team == ScorecardTeamSide.home.name) {
      return ScorecardTeam.home;
    }
    return ScorecardTeam.away;
  }
}
