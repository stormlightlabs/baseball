import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/core/data/local/app_database.dart';
import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/features/games/data/repositories/game_repository.dart';
import 'package:bigfly_mobile/features/home/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/features/more/data/repositories/api_more_repository.dart';
import 'package:bigfly_mobile/features/players/data/repositories/player_repository.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:bigfly_mobile/features/teams/data/repositories/team_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final appDatabase = AppDatabase();
  final cacheBootstrap = await loadCacheBootstrapSnapshot(appDatabase);
  final cacheStore = DriftCacheStore(appDatabase, cacheBootstrap);

  final dio = buildApiDio();
  final apiClient = BaseballApiClient(dio);
  final homeRepository = ApiHomeRepository(apiClient);
  final playerRepository = ApiPlayerRepository(apiClient, cacheStore);
  final teamRepository = ApiTeamRepository(apiClient);
  final gameRepository = ApiGameRepository(apiClient);
  final moreRepository = ApiMoreRepository(apiClient);
  final scorecardRepository = DriftScorecardRepository(appDatabase);

  runApp(
    BigFlyApp(
      cacheStore: cacheStore,
      homeRepository: homeRepository,
      playerRepository: playerRepository,
      teamRepository: teamRepository,
      gameRepository: gameRepository,
      moreRepository: moreRepository,
      scorecardRepository: scorecardRepository,
    ),
  );
}

Dio buildApiDio() {
  final dio = Dio(BaseOptions(connectTimeout: const Duration(seconds: 5), receiveTimeout: const Duration(seconds: 5)));
  dio.options.baseUrl = _resolveBaseUrl();
  dio.options.headers['X-BigFly-Client'] = 'mobile';
  return dio;
}

String _resolveBaseUrl() {
  const fromEnv = String.fromEnvironment('API_BASE_URL');
  if (fromEnv.isNotEmpty) {
    return fromEnv;
  }

  if (defaultTargetPlatform == TargetPlatform.android) {
    return 'http://10.0.2.2:8080';
  }
  return 'http://127.0.0.1:8080';
}
