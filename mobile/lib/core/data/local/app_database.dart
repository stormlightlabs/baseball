import 'package:drift/drift.dart';
import 'package:drift_flutter/drift_flutter.dart';
import 'package:path_provider/path_provider.dart';

part 'app_database.g.dart';

enum ScorecardGameStatus { inProgress, finalGame }

enum ScorecardHalf { top, bottom }

enum ScorecardTeamSide { away, home }

class AppSettings extends Table {
  TextColumn get key => text()();
  TextColumn get value => text()();

  @override
  Set<Column<Object>> get primaryKey => <Column<Object>>{key};
}

class RecentPlayers extends Table {
  TextColumn get playerId => text().named('player_id')();
  IntColumn get rankIndex => integer().named('rank_index')();
  DateTimeColumn get updatedAt => dateTime().named('updated_at').clientDefault(() => DateTime.now())();

  @override
  Set<Column<Object>> get primaryKey => <Column<Object>>{playerId};
}

@TableIndex(name: 'scorecard_games_status_last_modified_idx', columns: <Symbol>{#status, #lastModifiedAt})
class ScorecardGames extends Table {
  TextColumn get uuid => text()();
  TextColumn get awayTeamName => text().named('away_team_name')();
  TextColumn get awayTeamAbbreviation => text().named('away_team_abbreviation')();
  TextColumn get homeTeamName => text().named('home_team_name')();
  TextColumn get homeTeamAbbreviation => text().named('home_team_abbreviation')();
  TextColumn get venue => text().nullable()();
  DateTimeColumn get gameDate => dateTime().named('game_date')();
  TextColumn get status => text()();

  IntColumn get awayScore => integer().named('away_score').withDefault(const Constant(0))();
  IntColumn get homeScore => integer().named('home_score').withDefault(const Constant(0))();
  IntColumn get awayHits => integer().named('away_hits').withDefault(const Constant(0))();
  IntColumn get homeHits => integer().named('home_hits').withDefault(const Constant(0))();
  IntColumn get awayErrors => integer().named('away_errors').withDefault(const Constant(0))();
  IntColumn get homeErrors => integer().named('home_errors').withDefault(const Constant(0))();
  IntColumn get pitchCount => integer().named('pitch_count').withDefault(const Constant(0))();

  IntColumn get currentInning => integer().named('current_inning').withDefault(const Constant(1))();
  TextColumn get currentHalf => text().named('current_half')();
  IntColumn get outs => integer().withDefault(const Constant(0))();
  IntColumn get balls => integer().withDefault(const Constant(0))();
  IntColumn get strikes => integer().withDefault(const Constant(0))();
  IntColumn get awayBatterIndex => integer().named('away_batter_index').withDefault(const Constant(0))();
  IntColumn get homeBatterIndex => integer().named('home_batter_index').withDefault(const Constant(0))();
  IntColumn get lastPlaySequence => integer().named('last_play_sequence').nullable()();

  DateTimeColumn get createdAt => dateTime().named('created_at').clientDefault(() => DateTime.now())();
  DateTimeColumn get lastModifiedAt => dateTime().named('last_modified_at').clientDefault(() => DateTime.now())();

  @override
  Set<Column<Object>> get primaryKey => <Column<Object>>{uuid};
}

@TableIndex(name: 'scorecard_lineups_game_side_order_idx', columns: <Symbol>{#gameUuid, #teamSide, #battingOrder})
class ScorecardLineups extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get gameUuid => text().named('game_uuid').references(ScorecardGames, #uuid)();
  TextColumn get teamSide => text().named('team_side')();
  IntColumn get battingOrder => integer().named('batting_order')();
  TextColumn get playerName => text().named('player_name')();
  TextColumn get positionCode => text().named('position_code').nullable()();
}

@TableIndex(name: 'scorecard_innings_game_inning_half_idx', columns: <Symbol>{#gameUuid, #inningNumber, #half})
@DataClassName('ScorecardInningRow')
class ScorecardInnings extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get gameUuid => text().named('game_uuid').references(ScorecardGames, #uuid)();
  IntColumn get inningNumber => integer().named('inning_number')();
  TextColumn get half => text()();
}

@TableIndex(name: 'scorecard_plays_game_sequence_idx', columns: <Symbol>{#gameUuid, #sequence})
@TableIndex(name: 'scorecard_plays_inning_sequence_idx', columns: <Symbol>{#inningId, #sequence})
@DataClassName('ScorecardPlayRow')
class ScorecardPlays extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get gameUuid => text().named('game_uuid').references(ScorecardGames, #uuid)();
  IntColumn get inningId => integer().named('inning_id').references(ScorecardInnings, #id)();
  TextColumn get battingTeam => text().named('batting_team')();
  IntColumn get sequence => integer()();
  IntColumn get batterIndex => integer().named('batter_index')();
  TextColumn get batterName => text().named('batter_name')();
  TextColumn get batterPosition => text().named('batter_position').nullable()();
  TextColumn get outcomeCode => text().named('outcome_code')();
  TextColumn get putoutSequence => text().named('putout_sequence').nullable()();
  TextColumn get pitchLogJson => text().named('pitch_log_json')();
  TextColumn get baseStateBefore => text().named('base_state_before')();
  TextColumn get baseStateAfter => text().named('base_state_after')();
  IntColumn get rbi => integer()();
  BoolColumn get scored => boolean()();
  DateTimeColumn get recordedAt => dateTime().named('recorded_at').clientDefault(() => DateTime.now())();
}

@DriftDatabase(
  tables: <Type>[AppSettings, RecentPlayers, ScorecardGames, ScorecardLineups, ScorecardInnings, ScorecardPlays],
)
class AppDatabase extends _$AppDatabase {
  AppDatabase([QueryExecutor? executor]) : super(executor ?? _openConnection());

  @override
  int get schemaVersion => 1;

  static QueryExecutor _openConnection() {
    return driftDatabase(
      name: 'bigfly_mobile',
      native: const DriftNativeOptions(databaseDirectory: getApplicationSupportDirectory),
    );
  }

  Future<Map<String, String>> readSettings(Iterable<String> keys) async {
    final normalized = keys.toSet();
    if (normalized.isEmpty) {
      return <String, String>{};
    }

    final rows = await (select(appSettings)..where((tbl) => tbl.key.isIn(normalized))).get();
    return <String, String>{for (final row in rows) row.key: row.value};
  }

  Future<void> writeSetting({required String key, required String? value}) async {
    if (value == null || value.isEmpty) {
      await (delete(appSettings)..where((tbl) => tbl.key.equals(key))).go();
      return;
    }

    await into(appSettings).insertOnConflictUpdate(AppSettingsCompanion.insert(key: key, value: value));
  }

  Future<List<String>> readRecentPlayerIds({int limit = 8}) async {
    final rows =
        await (select(recentPlayers)
              ..orderBy(<OrderingTerm Function($RecentPlayersTable)>[(tbl) => OrderingTerm.asc(tbl.rankIndex)])
              ..limit(limit))
            .get();
    return rows.map((row) => row.playerId).toList(growable: false);
  }

  Future<void> pushRecentPlayerId(String playerId, {int limit = 8}) async {
    await transaction(() async {
      final rows = await (select(
        recentPlayers,
      )..orderBy(<OrderingTerm Function($RecentPlayersTable)>[(tbl) => OrderingTerm.asc(tbl.rankIndex)])).get();
      final next = <String>[playerId, ...rows.map((row) => row.playerId).where((id) => id != playerId)];
      final bounded = next.take(limit).toList(growable: false);

      await delete(recentPlayers).go();
      for (var index = 0; index < bounded.length; index++) {
        await into(recentPlayers).insert(
          RecentPlayersCompanion.insert(playerId: bounded[index], rankIndex: index, updatedAt: Value(DateTime.now())),
        );
      }
    });
  }
}
