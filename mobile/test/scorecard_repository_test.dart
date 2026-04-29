import 'package:bigfly_mobile/core/data/local/app_database.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('DriftScorecardRepository', () {
    late AppDatabase database;
    late DriftScorecardRepository repository;

    setUp(() {
      database = AppDatabase(NativeDatabase.memory());
      repository = DriftScorecardRepository(database);
    });

    tearDown(() async {
      await database.close();
    });

    test('creates and lists games sorted by last modified desc', () async {
      await repository.createGame(
        ScorecardGameDraft(
          uuid: 'game_one',
          awayTeamName: 'Chicago Cubs',
          awayTeamAbbreviation: 'CHC',
          homeTeamName: 'St. Louis Cardinals',
          homeTeamAbbreviation: 'STL',
          venue: 'Busch Stadium',
          gameDate: DateTime(2026, 4, 20),
          status: ScorecardStatus.inProgress,
          lineups: const <ScorecardLineupSlot>[],
        ),
      );
      await repository.createGame(
        ScorecardGameDraft(
          uuid: 'game_two',
          awayTeamName: 'New York Yankees',
          awayTeamAbbreviation: 'NYY',
          homeTeamName: 'Boston Red Sox',
          homeTeamAbbreviation: 'BOS',
          venue: 'Fenway Park',
          gameDate: DateTime(2026, 4, 21),
          status: ScorecardStatus.finalGame,
          lineups: const <ScorecardLineupSlot>[],
        ),
      );

      await repository.appendPlay(
        const ScorecardPlayDraft(
          gameUuid: 'game_one',
          inningNumber: 1,
          half: ScorecardHalfInning.top,
          battingTeam: ScorecardTeam.away,
          batterIndex: 0,
          batterName: 'Lead Off',
          batterPosition: 'CF',
          outcomeCode: 'BB',
          putoutSequence: null,
          pitchLog: <String>['B', 'B', 'B', 'B'],
          baseStateBefore: '000',
          baseStateAfter: '100',
          rbi: 0,
          scored: false,
          nextState: ScorecardGameRuntimeState(
            currentInning: 1,
            currentHalf: ScorecardHalfInning.top,
            awayScore: 0,
            homeScore: 0,
            pitchCount: 4,
            outs: 0,
            balls: 0,
            strikes: 0,
            awayBatterIndex: 1,
            homeBatterIndex: 0,
          ),
        ),
      );

      final all = await repository.listGames();
      expect(all.length, 2);
      expect(all.first.uuid, 'game_one');

      final inProgress = await repository.listGames(filter: ScorecardStatusFilter.inProgress);
      expect(inProgress.length, 1);
      expect(inProgress.first.uuid, 'game_one');
    });

    test('appends and undoes a play with persisted inning graph', () async {
      await repository.createGame(
        ScorecardGameDraft(
          uuid: 'game_three',
          awayTeamName: 'Los Angeles Dodgers',
          awayTeamAbbreviation: 'LAD',
          homeTeamName: 'San Diego Padres',
          homeTeamAbbreviation: 'SDP',
          venue: null,
          gameDate: DateTime(2026, 4, 22),
          status: ScorecardStatus.inProgress,
          lineups: const <ScorecardLineupSlot>[
            ScorecardLineupSlot(
              team: ScorecardTeam.away,
              battingOrder: 1,
              playerName: 'Shohei Ohtani',
              positionCode: 'DH',
            ),
          ],
        ),
      );

      await repository.appendPlay(
        const ScorecardPlayDraft(
          gameUuid: 'game_three',
          inningNumber: 1,
          half: ScorecardHalfInning.top,
          battingTeam: ScorecardTeam.away,
          batterIndex: 0,
          batterName: 'Shohei Ohtani',
          batterPosition: 'DH',
          outcomeCode: '1B',
          putoutSequence: null,
          pitchLog: <String>['B', 'S', 'X'],
          baseStateBefore: '000',
          baseStateAfter: '100',
          rbi: 0,
          scored: false,
          nextState: ScorecardGameRuntimeState(
            currentInning: 1,
            currentHalf: ScorecardHalfInning.top,
            awayScore: 0,
            homeScore: 0,
            pitchCount: 3,
            outs: 0,
            balls: 0,
            strikes: 0,
            awayBatterIndex: 1,
            homeBatterIndex: 0,
          ),
        ),
      );

      final detail = await repository.getGame('game_three');
      expect(detail, isNotNull);
      expect(detail!.innings.length, 1);
      expect(detail.innings.first.plays.length, 1);
      expect(detail.summary.pitchCount, 3);

      final undone = await repository.undoLastPlay(
        gameUuid: 'game_three',
        restoredState: const ScorecardGameRuntimeState(
          currentInning: 1,
          currentHalf: ScorecardHalfInning.top,
          awayScore: 0,
          homeScore: 0,
          pitchCount: 0,
          outs: 0,
          balls: 0,
          strikes: 0,
          awayBatterIndex: 0,
          homeBatterIndex: 0,
        ),
      );

      expect(undone, isTrue);
      final afterUndo = await repository.getGame('game_three');
      expect(afterUndo, isNotNull);
      expect(afterUndo!.innings.length, 0);
      expect(afterUndo.summary.pitchCount, 0);
    });

    test('appendPlay fails atomically when game does not exist', () async {
      await expectLater(
        () => repository.appendPlay(
          const ScorecardPlayDraft(
            gameUuid: 'missing',
            inningNumber: 1,
            half: ScorecardHalfInning.top,
            battingTeam: ScorecardTeam.away,
            batterIndex: 0,
            batterName: 'Test Player',
            batterPosition: null,
            outcomeCode: 'K',
            putoutSequence: null,
            pitchLog: <String>['C', 'S', 'S'],
            baseStateBefore: '000',
            baseStateAfter: '000',
            rbi: 0,
            scored: false,
            nextState: ScorecardGameRuntimeState(
              currentInning: 1,
              currentHalf: ScorecardHalfInning.top,
              awayScore: 0,
              homeScore: 0,
              pitchCount: 3,
              outs: 1,
              balls: 0,
              strikes: 0,
              awayBatterIndex: 1,
              homeBatterIndex: 0,
            ),
          ),
        ),
        throwsA(isA<StateError>()),
      );

      final games = await repository.listGames();
      expect(games, isEmpty);
      final plays = await database.select(database.scorecardPlays).get();
      expect(plays, isEmpty);
    });
  });
}
