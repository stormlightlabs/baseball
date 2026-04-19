import 'package:hive/hive.dart';

abstract class CacheStore {
  static const String appBoxName = 'bigfly_cache';
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

class HiveCacheStore implements CacheStore {
  HiveCacheStore(this._box);

  final Box<String> _box;

  @override
  String? get selectedTeamCode => _box.get(CacheStore.selectedTeamKey);

  @override
  Future<void> setSelectedTeamCode(String? teamCode) async {
    if (teamCode == null) {
      await _box.delete(CacheStore.selectedTeamKey);
      return;
    }
    await _box.put(CacheStore.selectedTeamKey, teamCode);
  }

  @override
  String? get cachedHealthStatus => _box.get(CacheStore.cachedHealthStatusKey);

  @override
  Future<void> setCachedHealthStatus(String status) async {
    await _box.put(CacheStore.cachedHealthStatusKey, status);
  }

  @override
  List<String> get recentPlayerIds {
    final raw = _box.get(CacheStore.recentPlayerIdsKey);
    if (raw == null || raw.isEmpty) {
      return <String>[];
    }
    return raw.split('|').map((item) => item.trim()).where((item) => item.isNotEmpty).toList(growable: false);
  }

  @override
  Future<void> pushRecentPlayerId(String playerId) async {
    final next = <String>[playerId, ...recentPlayerIds.where((id) => id != playerId)];
    final bounded = next.take(8).toList(growable: false);
    await _box.put(CacheStore.recentPlayerIdsKey, bounded.join('|'));
  }
}
