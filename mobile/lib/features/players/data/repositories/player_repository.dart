import 'package:bigfly_mobile/colors.dart';
import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:bigfly_mobile/core/data/json/json_helpers.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';

abstract class PlayerRepository {
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 12});
  Future<PlayerDetailBundle> fetchPlayerDetail(String playerId);
  Future<List<PlayerSearchResult>> getRecentPlayers();
  Future<void> rememberPlayer(String playerId);
}

class ApiPlayerRepository implements PlayerRepository {
  ApiPlayerRepository(this._api, this._cacheStore);

  final BaseballApiClient _api;
  final CacheStore _cacheStore;

  @override
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 12}) async {
    final payload = await _api.searchPlayers(query: query, perPage: limit);
    return asJsonMapList(payload['data']).map(PlayerSearchResult.fromJson).toList(growable: false);
  }

  @override
  Future<PlayerDetailBundle> fetchPlayerDetail(String playerId) async {
    final playerFuture = _api.getPlayer(playerId);
    final seasonsFuture = _api.getPlayerSeasons(playerId);
    final awardsFuture = _api.getPlayerAwards(playerId, perPage: 150);
    final hofFuture = _api.getPlayerHallOfFame(playerId);

    final playerRaw = await playerFuture;
    final seasonsRaw = await seasonsFuture;
    final awardsRaw = await awardsFuture;
    final hofRaw = await hofFuture;

    final player = PlayerProfile.fromJson(playerRaw);
    final battingSeasons = asJsonMapList(
      seasonsRaw['batting'],
    ).map(PlayerBattingSeason.fromJson).toList(growable: false);
    final pitchingSeasons = asJsonMapList(
      seasonsRaw['pitching'],
    ).map(PlayerPitchingSeason.fromJson).toList(growable: false);
    final awards = asJsonMapList(awardsRaw['data']).map(PlayerAward.fromJson).toList(growable: false);
    final hallRecords = asJsonMapList(hofRaw['records']).map(PlayerHallOfFameRecord.fromJson).toList(growable: false);

    final themeTeamCode = await _resolveThemeTeamCode(player, battingSeasons, pitchingSeasons);

    return PlayerDetailBundle(
      player: player,
      battingSeasons: battingSeasons,
      pitchingSeasons: pitchingSeasons,
      awards: awards,
      hallOfFameRecords: hallRecords,
      themeTeamCode: themeTeamCode,
    );
  }

  @override
  Future<List<PlayerSearchResult>> getRecentPlayers() async {
    final recentIds = _cacheStore.recentPlayerIds;
    final players = <PlayerSearchResult>[];
    for (final id in recentIds) {
      try {
        final payload = await _api.getPlayer(id);
        final profile = PlayerProfile.fromJson(payload);
        players.add(PlayerSearchResult(id: profile.id, name: profile.fullName, subtitle: profile.subtitle));
      } catch (_) {
        // Skip stale or invalid cache entries.
      }
    }
    return players;
  }

  @override
  Future<void> rememberPlayer(String playerId) async {
    await _cacheStore.pushRecentPlayerId(playerId);
  }

  Future<String?> _resolveThemeTeamCode(
    PlayerProfile profile,
    List<PlayerBattingSeason> battingSeasons,
    List<PlayerPitchingSeason> pitchingSeasons,
  ) async {
    final direct = normalizeMlbTeamCode(profile.latestTeam);
    if (direct != null) {
      return direct;
    }

    String? teamId = profile.latestTeam;
    int? year = profile.latestSeason;

    if (teamId == null || teamId.isEmpty || year == null) {
      final latestBatting = _latestBattingSeason(battingSeasons);
      final latestPitching = _latestPitchingSeason(pitchingSeasons);
      if (latestBatting != null && latestPitching != null) {
        if (latestBatting.year >= latestPitching.year) {
          teamId = latestBatting.teamId;
          year = latestBatting.year;
        } else {
          teamId = latestPitching.teamId;
          year = latestPitching.year;
        }
      } else if (latestBatting != null) {
        teamId = latestBatting.teamId;
        year = latestBatting.year;
      } else if (latestPitching != null) {
        teamId = latestPitching.teamId;
        year = latestPitching.year;
      }
    }

    if (teamId == null || teamId.isEmpty || year == null) {
      return null;
    }

    try {
      final payload = await _api.getTeamSeason(teamId, year: year);
      final team = TeamSeasonSnapshot.fromJson(payload);
      return normalizeMlbTeamCode(team.franchiseId) ??
          normalizeMlbTeamCode(team.teamId) ??
          normalizeMlbTeamCode(teamId);
    } catch (_) {
      return normalizeMlbTeamCode(teamId);
    }
  }

  PlayerBattingSeason? _latestBattingSeason(List<PlayerBattingSeason> seasons) {
    if (seasons.isEmpty) {
      return null;
    }
    final sorted = [...seasons]..sort((a, b) => b.year.compareTo(a.year));
    return sorted.first;
  }

  PlayerPitchingSeason? _latestPitchingSeason(List<PlayerPitchingSeason> seasons) {
    if (seasons.isEmpty) {
      return null;
    }
    final sorted = [...seasons]..sort((a, b) => b.year.compareTo(a.year));
    return sorted.first;
  }
}
