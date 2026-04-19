import 'package:dio/dio.dart';

class BaseballApiClient {
  BaseballApiClient(this._dio);

  final Dio _dio;

  Future<Map<String, dynamic>> getMeta() => _getMap('/v1/meta');

  Future<List<dynamic>> getMetaDatasets() => _getList('/v1/meta/datasets');

  Future<Map<String, dynamic>> searchPlayers({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/v1/search/players', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> searchTeams({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/v1/search/teams', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> searchGames({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/v1/search/games', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getFranchises({bool active = true}) =>
      _getMap('/v1/franchises', query: <String, dynamic>{'active': active});

  Future<Map<String, dynamic>> getFranchise(String id) => _getMap('/v1/franchises/$id');

  Future<List<dynamic>> getSeasons() => _getList('/v1/seasons');

  Future<Map<String, dynamic>> getSeasonBattingLeaders({
    required int year,
    required String stat,
    String? league,
    int page = 1,
    int perPage = 10,
  }) => _getMap(
    '/v1/seasons/$year/leaders/batting',
    query: <String, dynamic>{
      'stat': stat,
      if (league != null && league.isNotEmpty) 'league': league,
      'page': page,
      'per_page': perPage,
    },
  );

  Future<Map<String, dynamic>> getSeasonPitchingLeaders({
    required int year,
    required String stat,
    String? league,
    int page = 1,
    int perPage = 10,
  }) => _getMap(
    '/v1/seasons/$year/leaders/pitching',
    query: <String, dynamic>{
      'stat': stat,
      if (league != null && league.isNotEmpty) 'league': league,
      'page': page,
      'per_page': perPage,
    },
  );

  Future<Map<String, dynamic>> getCareerBattingLeaders({required String stat, int page = 1, int perPage = 10}) =>
      _getMap('/v1/leaders/batting/career', query: <String, dynamic>{'stat': stat, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getCareerPitchingLeaders({required String stat, int page = 1, int perPage = 10}) =>
      _getMap('/v1/leaders/pitching/career', query: <String, dynamic>{'stat': stat, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getSeasonAwards({required int year, int page = 1, int perPage = 20}) =>
      _getMap('/v1/seasons/$year/awards', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getSeasonPostseasonSeries({required int year}) =>
      _getMap('/v1/seasons/$year/postseason/series');

  Future<Map<String, dynamic>> getTeamBattingStats({required int season, int page = 1, int perPage = 60}) =>
      _getMap('/v1/stats/teams/batting', query: <String, dynamic>{'season': season, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getPlayer(String id) => _getMap('/v1/players/$id');

  Future<Map<String, dynamic>> getPlayerSeasons(String id) => _getMap('/v1/players/$id/seasons');

  Future<Map<String, dynamic>> getPlayerAwards(String id, {int page = 1, int perPage = 100}) =>
      _getMap('/v1/players/$id/awards', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getPlayerHallOfFame(String id) => _getMap('/v1/players/$id/hall-of-fame');

  Future<Map<String, dynamic>> getPlayerBattingStats(String id) => _getMap('/v1/players/$id/stats/batting');

  Future<Map<String, dynamic>> getPlayerPitchingStats(String id) => _getMap('/v1/players/$id/stats/pitching');

  Future<Map<String, dynamic>> getTeamSeason(String teamId, {required int year}) =>
      _getMap('/v1/teams/$teamId', query: <String, dynamic>{'year': year});

  Future<List<dynamic>> getTeamRoster({required int year, required String teamId}) =>
      _getList('/v1/seasons/$year/teams/$teamId/roster');

  Future<Map<String, dynamic>> getTeamSchedule({
    required int year,
    required String teamId,
    int page = 1,
    int perPage = 40,
  }) =>
      _getMap('/v1/seasons/$year/teams/$teamId/schedule', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getTeamDailyLogs({
    required int year,
    required String teamId,
    int page = 1,
    int perPage = 40,
  }) => _getMap(
    '/v1/seasons/$year/teams/$teamId/daily-logs',
    query: <String, dynamic>{'page': page, 'per_page': perPage},
  );

  Future<Map<String, dynamic>> getTeamRunDifferential({
    required String teamId,
    required int season,
    List<int> windows = const <int>[10, 20, 30],
  }) => _getMap(
    '/v1/teams/$teamId/run-differential',
    query: <String, dynamic>{'season': season, 'windows': windows.join(',')},
  );

  Future<Map<String, dynamic>> getSeasonTeams({required int year, int page = 1, int perPage = 50}) =>
      _getMap('/v1/seasons/$year/teams', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getGames({
    required int season,
    String? homeTeam,
    String? awayTeam,
    String? parkId,
    String? dateFrom,
    String? dateTo,
    int page = 1,
    int perPage = 120,
  }) => _getMap(
    '/v1/games',
    query: <String, dynamic>{
      'season': season,
      if (homeTeam != null && homeTeam.isNotEmpty) 'home_team': homeTeam,
      if (awayTeam != null && awayTeam.isNotEmpty) 'away_team': awayTeam,
      if (parkId != null && parkId.isNotEmpty) 'park_id': parkId,
      if (dateFrom != null && dateFrom.isNotEmpty) 'date_from': dateFrom,
      if (dateTo != null && dateTo.isNotEmpty) 'date_to': dateTo,
      'page': page,
      'per_page': perPage,
    },
  );

  Future<Map<String, dynamic>> getGameEvents(String gameId, {int page = 1, int perPage = 200}) =>
      _getMap('/v1/games/$gameId/events', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getGameWinProbabilitySummary(String gameId) =>
      _getMap('/v1/games/$gameId/win-probability/summary');

  Future<Map<String, dynamic>> _getMap(String path, {Map<String, dynamic>? query}) async {
    final response = await _dio.get<dynamic>(path, queryParameters: query);
    final data = response.data;
    if (data is Map<String, dynamic>) {
      return data;
    }
    if (data is Map) {
      return data.map((key, value) => MapEntry(key.toString(), value));
    }
    return <String, dynamic>{};
  }

  Future<List<dynamic>> _getList(String path, {Map<String, dynamic>? query}) async {
    final response = await _dio.get<dynamic>(path, queryParameters: query);
    final data = response.data;
    if (data is List<dynamic>) {
      return data;
    }
    return <dynamic>[];
  }
}
