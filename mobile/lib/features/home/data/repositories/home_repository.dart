import 'package:bigfly_mobile/core/data/json/json_helpers.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';

abstract class HomeRepository {
  Future<MetaSnapshot> fetchMeta();
  Future<List<HomeSearchItem>> search(HomeEntityType entity, String query);
}

class ApiHomeRepository implements HomeRepository {
  ApiHomeRepository(this._api);

  final BaseballApiClient _api;

  @override
  Future<MetaSnapshot> fetchMeta() async {
    final payload = await _api.getMeta();
    return MetaSnapshot.fromJson(payload);
  }

  @override
  Future<List<HomeSearchItem>> search(HomeEntityType entity, String query) async {
    final trimmed = query.trim();
    if (trimmed.isEmpty) {
      return <HomeSearchItem>[];
    }

    return switch (entity) {
      HomeEntityType.players => _searchPlayers(trimmed),
      HomeEntityType.teams => _searchTeams(trimmed),
      HomeEntityType.games => _searchGames(trimmed),
      HomeEntityType.franchises => _searchFranchises(trimmed),
      HomeEntityType.seasons => _searchSeasons(trimmed),
    };
  }

  Future<List<HomeSearchItem>> _searchPlayers(String query) async {
    final payload = await _api.searchPlayers(query: query, perPage: 12);
    final rows = asJsonMapList(payload['data']);
    return rows
        .map(PlayerSearchResult.fromJson)
        .map(
          (item) =>
              HomeSearchItem(entity: HomeEntityType.players, id: item.id, title: item.name, subtitle: item.subtitle),
        )
        .toList(growable: false);
  }

  Future<List<HomeSearchItem>> _searchTeams(String query) async {
    final payload = await _api.searchTeams(query: query, perPage: 12);
    final rows = asJsonMapList(payload['data']);
    return rows
        .map((row) {
          final teamName = stringOrEmpty(row['name']);
          final teamId = stringOrEmpty(row['team_id']);
          final year = intOrZero(row['year']);
          final league = nullableString(row['league']);
          return HomeSearchItem(
            entity: HomeEntityType.teams,
            id: teamId,
            title: teamName.isEmpty ? teamId : teamName,
            subtitle: '$teamId · $year${league != null ? ' · $league' : ''}',
          );
        })
        .toList(growable: false);
  }

  Future<List<HomeSearchItem>> _searchGames(String query) async {
    final payload = await _api.searchGames(query: query, perPage: 12);
    final rows = asJsonMapList(payload['data']);
    return rows
        .map((row) {
          final home = stringOrEmpty(row['home_team']);
          final away = stringOrEmpty(row['away_team']);
          final gameId = stringOrEmpty(row['id']);
          final date = stringOrEmpty(row['date']);
          return HomeSearchItem(
            entity: HomeEntityType.games,
            id: gameId,
            title: away.isNotEmpty && home.isNotEmpty ? '$away @ $home' : gameId,
            subtitle: '$date · $gameId',
          );
        })
        .toList(growable: false);
  }

  Future<List<HomeSearchItem>> _searchFranchises(String query) async {
    final payload = await _api.getFranchises(active: true);
    final rows = asJsonMapList(payload['franchises']);
    final normalized = query.toLowerCase();
    return rows
        .where((row) {
          final id = stringOrEmpty(row['id']).toLowerCase();
          final name = stringOrEmpty(row['name']).toLowerCase();
          return id.contains(normalized) || name.contains(normalized);
        })
        .take(12)
        .map(
          (row) => HomeSearchItem(
            entity: HomeEntityType.franchises,
            id: stringOrEmpty(row['id']),
            title: stringOrEmpty(row['name']),
            subtitle: 'Franchise · ${stringOrEmpty(row['id'])}',
          ),
        )
        .toList(growable: false);
  }

  Future<List<HomeSearchItem>> _searchSeasons(String query) async {
    final seasons = await _api.getSeasons();
    final rows = asJsonMapList(seasons);
    final normalized = query.trim().toLowerCase();
    final range = _parseYearRange(normalized);

    return rows
        .where((row) {
          final year = intOrZero(row['year']);
          if (range != null) {
            return year >= range.$1 && year <= range.$2;
          }
          return year.toString().contains(normalized);
        })
        .take(12)
        .map((row) {
          final year = intOrZero(row['year']);
          final teams = intOrZero(row['team_count']);
          return HomeSearchItem(
            entity: HomeEntityType.seasons,
            id: year.toString(),
            title: 'Season $year',
            subtitle: '$teams teams',
          );
        })
        .toList(growable: false);
  }

  (int, int)? _parseYearRange(String raw) {
    if (!raw.contains('-')) {
      return null;
    }

    final parts = raw.split('-').map((part) => part.trim()).toList(growable: false);
    if (parts.length != 2) {
      return null;
    }

    final from = int.tryParse(parts[0]);
    final toPart = parts[1];
    if (from == null) {
      return null;
    }

    var to = int.tryParse(toPart);
    if (to == null && toPart.length == 2 && parts[0].length == 4) {
      final prefix = parts[0].substring(0, 2);
      to = int.tryParse('$prefix$toPart');
    }
    if (to == null) {
      return null;
    }
    if (to < from) {
      return (to, from);
    }
    return (from, to);
  }
}
