import 'package:bigfly_mobile/core/data/local/app_database.dart';
import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('DriftCacheStore', () {
    late AppDatabase database;

    setUp(() {
      database = AppDatabase(NativeDatabase.memory());
    });

    tearDown(() async {
      await database.close();
    });

    test('stores selected team and health status', () async {
      final snapshot = await loadCacheBootstrapSnapshot(database);
      final store = DriftCacheStore(database, snapshot);

      expect(store.selectedTeamCode, isNull);
      expect(store.cachedHealthStatus, isNull);

      await store.setSelectedTeamCode('NYY');
      await store.setCachedHealthStatus('ok');

      expect(store.selectedTeamCode, 'NYY');
      expect(store.cachedHealthStatus, 'ok');

      await store.setSelectedTeamCode(null);
      expect(store.selectedTeamCode, isNull);
    });

    test('maintains recents ordering with dedupe and cap', () async {
      final snapshot = await loadCacheBootstrapSnapshot(database);
      final store = DriftCacheStore(database, snapshot);

      for (var i = 0; i < 10; i++) {
        await store.pushRecentPlayerId('player_$i');
      }
      await store.pushRecentPlayerId('player_5');

      expect(store.recentPlayerIds.length, 8);
      expect(store.recentPlayerIds.first, 'player_5');
      expect(store.recentPlayerIds.toSet().length, 8);

      final reloaded = await loadCacheBootstrapSnapshot(database);
      expect(reloaded.recentPlayerIds.first, 'player_5');
      expect(reloaded.recentPlayerIds.length, 8);
    });
  });
}
