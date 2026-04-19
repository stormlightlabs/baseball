import 'package:dio/dio.dart';

class BaseballApiClient {
  BaseballApiClient(this._dio);

  final Dio _dio;

  Future<Map<String, dynamic>> getMeta() => _getMap('/api/v1/meta');

  Future<Map<String, dynamic>> searchPlayers({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/api/v1/search/players', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> searchTeams({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/api/v1/search/teams', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> searchGames({required String query, int page = 1, int perPage = 12}) =>
      _getMap('/api/v1/search/games', query: <String, dynamic>{'q': query, 'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getFranchises({bool active = true}) =>
      _getMap('/api/v1/franchises', query: <String, dynamic>{'active': active});

  Future<List<dynamic>> getSeasons() => _getList('/api/v1/seasons');

  Future<Map<String, dynamic>> getPlayer(String id) => _getMap('/api/v1/players/$id');

  Future<Map<String, dynamic>> getPlayerSeasons(String id) => _getMap('/api/v1/players/$id/seasons');

  Future<Map<String, dynamic>> getPlayerAwards(String id, {int page = 1, int perPage = 100}) =>
      _getMap('/api/v1/players/$id/awards', query: <String, dynamic>{'page': page, 'per_page': perPage});

  Future<Map<String, dynamic>> getPlayerHallOfFame(String id) => _getMap('/api/v1/players/$id/hall-of-fame');

  Future<Map<String, dynamic>> getTeamSeason(String teamId, {required int year}) =>
      _getMap('/api/v1/teams/$teamId', query: <String, dynamic>{'year': year});

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
