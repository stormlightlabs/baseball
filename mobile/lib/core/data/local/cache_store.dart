import 'package:bigfly_mobile/core/data/local/app_database.dart';

class CacheBootstrapSnapshot {
  const CacheBootstrapSnapshot({
    required this.selectedTeamCode,
    required this.cachedHealthStatus,
    required this.recentPlayerIds,
  });

  final String? selectedTeamCode;
  final String? cachedHealthStatus;
  final List<String> recentPlayerIds;
}

abstract class CacheStore {
  static const String selectedTeamKey = 'selected_team';
  static const String cachedHealthStatusKey = 'cached_health_status';
  static const String recentPlayerIdsKey = 'recent_player_ids';

  String? get selectedTeamCode;
  Future<void> setSelectedTeamCode(String? teamCode);

  String? get cachedHealthStatus;
  Future<void> setCachedHealthStatus(String status);

  List<String> get recentPlayerIds;
  Future<void> pushRecentPlayerId(String playerId);
}

Future<CacheBootstrapSnapshot> loadCacheBootstrapSnapshot(AppDatabase database) async {
  final settings = await database.readSettings(<String>{CacheStore.selectedTeamKey, CacheStore.cachedHealthStatusKey});
  final recent = await database.readRecentPlayerIds();
  return CacheBootstrapSnapshot(
    selectedTeamCode: settings[CacheStore.selectedTeamKey],
    cachedHealthStatus: settings[CacheStore.cachedHealthStatusKey],
    recentPlayerIds: recent,
  );
}

class DriftCacheStore implements CacheStore {
  DriftCacheStore(this._database, CacheBootstrapSnapshot snapshot)
    : _selectedTeamCode = snapshot.selectedTeamCode,
      _cachedHealthStatus = snapshot.cachedHealthStatus,
      _recentPlayerIds = List<String>.from(snapshot.recentPlayerIds);

  final AppDatabase _database;

  String? _selectedTeamCode;
  String? _cachedHealthStatus;
  List<String> _recentPlayerIds;

  @override
  String? get selectedTeamCode => _selectedTeamCode;

  @override
  Future<void> setSelectedTeamCode(String? teamCode) async {
    _selectedTeamCode = teamCode;
    await _database.writeSetting(key: CacheStore.selectedTeamKey, value: teamCode);
  }

  @override
  String? get cachedHealthStatus => _cachedHealthStatus;

  @override
  Future<void> setCachedHealthStatus(String status) async {
    _cachedHealthStatus = status;
    await _database.writeSetting(key: CacheStore.cachedHealthStatusKey, value: status);
  }

  @override
  List<String> get recentPlayerIds => List<String>.unmodifiable(_recentPlayerIds);

  @override
  Future<void> pushRecentPlayerId(String playerId) async {
    final next = <String>[playerId, ..._recentPlayerIds.where((id) => id != playerId)];
    _recentPlayerIds = next.take(8).toList(growable: false);
    await _database.pushRecentPlayerId(playerId);
  }
}
