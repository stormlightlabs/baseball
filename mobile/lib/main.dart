import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:bigfly_mobile/data/network/bigfly_api.dart';
import 'package:bigfly_mobile/data/repositories/health_repository.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:hive_flutter/hive_flutter.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Hive.initFlutter();
  final cacheBox = await Hive.openBox<String>(CacheStore.appBoxName);
  final cacheStore = HiveCacheStore(cacheBox);

  final dio = Dio(BaseOptions(connectTimeout: const Duration(seconds: 5), receiveTimeout: const Duration(seconds: 5)));

  final api = BigFlyApi(dio, baseUrl: _resolveBaseUrl());
  final healthRepository = ApiHealthRepository(api, cacheStore);

  runApp(BigFlyApp(cacheStore: cacheStore, healthRepository: healthRepository));
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
