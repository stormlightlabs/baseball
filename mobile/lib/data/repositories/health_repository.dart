import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:bigfly_mobile/data/models/health_response.dart';
import 'package:bigfly_mobile/data/network/bigfly_api.dart';

abstract class HealthRepository {
  Future<HealthResponse> fetchHealth();
  HealthResponse? getCachedHealth();
}

class ApiHealthRepository implements HealthRepository {
  ApiHealthRepository(this._api, this._cacheStore);

  final BigFlyApi _api;
  final CacheStore _cacheStore;

  @override
  Future<HealthResponse> fetchHealth() async {
    final response = await _api.getHealth();
    await _cacheStore.setCachedHealthStatus(response.status);
    return response;
  }

  @override
  HealthResponse? getCachedHealth() {
    final cached = _cacheStore.cachedHealthStatus;
    if (cached == null || cached.isEmpty) {
      return null;
    }
    return HealthResponse(status: cached);
  }
}
