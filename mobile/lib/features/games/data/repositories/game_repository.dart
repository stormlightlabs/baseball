import 'package:bigfly_mobile/core/data/json/json_helpers.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/features/games/data/models/game_models.dart';

abstract class GameRepository {
  Future<List<int>> listAvailableSeasons();
  Future<List<GameTeamOption>> listTeamsForSeason(int season);
  Future<GameListResponse> fetchGames({required int season, String? teamId, int perPage = 120});
  Future<GameCardDetail> fetchGameDetail(GameSummaryRecord game);
}

class ApiGameRepository implements GameRepository {
  ApiGameRepository(this._api);

  final BaseballApiClient _api;

  @override
  Future<List<int>> listAvailableSeasons() async {
    final payload = await _api.getSeasons();
    final years =
        payload
            .map(asJsonMap)
            .map((row) => intOrZero(row['year']))
            .where((year) => year > 0)
            .toSet()
            .toList(growable: false)
          ..sort((a, b) => b.compareTo(a));

    if (years.isNotEmpty) {
      return years;
    }

    return <int>[DateTime.now().year];
  }

  @override
  Future<List<GameTeamOption>> listTeamsForSeason(int season) async {
    final payload = await _api.getSeasonTeams(year: season, perPage: 80);
    final options = asJsonMapList(
      payload['data'],
    ).map(GameTeamOption.fromJson).where((team) => team.teamId.isNotEmpty).toList(growable: false);

    final deduped = <String, GameTeamOption>{};
    for (final option in options) {
      deduped.putIfAbsent(option.teamId, () => option);
    }

    final sorted = deduped.values.toList(growable: false)..sort((a, b) => a.name.compareTo(b.name));
    return sorted;
  }

  @override
  Future<GameListResponse> fetchGames({required int season, String? teamId, int perPage = 120}) async {
    final normalizedTeam = nullableString(teamId);

    final payload = normalizedTeam == null
        ? await _api.getGames(season: season, page: 1, perPage: perPage)
        : await _api.getTeamSchedule(year: season, teamId: normalizedTeam, page: 1, perPage: perPage);

    final games = asJsonMapList(payload['data']).map(GameSummaryRecord.fromJson).toList(growable: false)
      ..sort(_compareGameRowsDesc);

    final total = intOrZero(payload['total']);
    return GameListResponse(games: games, total: total <= 0 ? games.length : total);
  }

  @override
  Future<GameCardDetail> fetchGameDetail(GameSummaryRecord game) async {
    final summaryFuture = _safeFetchMap(() => _api.getGameWinProbabilitySummary(game.id));
    final eventsFuture = _safeFetchMap(() => _api.getGameEvents(game.id, page: 1, perPage: 200));

    final summaryRaw = await summaryFuture;
    final eventsRaw = await eventsFuture;

    final summary = summaryRaw == null || summaryRaw.isEmpty ? null : GameWinProbabilitySummary.fromJson(summaryRaw);
    final events = asJsonMapList(eventsRaw?['data']).map(GameEvent.fromJson).toList(growable: false);

    final homeWinProbability = _resolveHomeWinProbability(game, summary);
    return GameCardDetail(
      awayWinProbability: (1 - homeWinProbability).clamp(0.0, 1.0),
      homeWinProbability: homeWinProbability.clamp(0.0, 1.0),
      keyPlays: _selectKeyPlays(events),
    );
  }

  List<GameEvent> _selectKeyPlays(List<GameEvent> events) {
    if (events.isEmpty) {
      return const <GameEvent>[];
    }

    final notable = events.where(_isNotablePlay).toList(growable: false);
    final source = notable.isNotEmpty ? notable : events;

    final start = source.length > 3 ? source.length - 3 : 0;
    final selected = source.sublist(start).toList(growable: false)..sort((a, b) => a.playNum.compareTo(b.playNum));

    return selected;
  }

  bool _isNotablePlay(GameEvent event) {
    final normalized = event.event.toUpperCase();
    return (event.runs ?? 0) > 0 ||
        normalized.contains('HOMER') ||
        normalized.contains('HOME RUN') ||
        normalized.contains('SCORES') ||
        normalized.contains('WALK-OFF');
  }

  double _resolveHomeWinProbability(GameSummaryRecord game, GameWinProbabilitySummary? summary) {
    if (summary != null) {
      return summary.homeWinProbEnd.clamp(0.0, 1.0);
    }

    if (game.homeScore == game.awayScore) {
      return 0.5;
    }

    return game.homeScore > game.awayScore ? 1.0 : 0.0;
  }

  int _compareGameRowsDesc(GameSummaryRecord a, GameSummaryRecord b) {
    final dateA = a.date;
    final dateB = b.date;

    if (dateA != null && dateB != null) {
      final cmp = dateB.compareTo(dateA);
      if (cmp != 0) {
        return cmp;
      }
    } else if (dateA == null && dateB != null) {
      return 1;
    } else if (dateA != null && dateB == null) {
      return -1;
    }

    return b.gameNumber.compareTo(a.gameNumber);
  }

  Future<Map<String, dynamic>?> _safeFetchMap(Future<Map<String, dynamic>> Function() operation) async {
    try {
      return await operation();
    } catch (_) {
      return null;
    }
  }
}
