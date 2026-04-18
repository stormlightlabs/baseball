import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:bigfly_mobile/data/models/health_response.dart';
import 'package:bigfly_mobile/data/repositories/health_repository.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('renders five-tab bottom navigation shell', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(cacheStore: InMemoryCacheStore(), healthRepository: FakeHealthRepository(), useDynamicColor: false),
    );
    await tester.pumpAndSettle();

    expect(find.text('Home'), findsOneWidget);
    expect(find.text('Players'), findsOneWidget);
    expect(find.text('Teams'), findsOneWidget);
    expect(find.text('Games'), findsOneWidget);
    expect(find.text('More'), findsOneWidget);
  });

  testWidgets('displays health status from API repository', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        healthRepository: FakeHealthRepository(status: 'ok'),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Status: ok'), findsOneWidget);
  });
}

class InMemoryCacheStore implements CacheStore {
  final Map<String, String> _data = <String, String>{};

  @override
  String? get cachedHealthStatus => _data[CacheStore.cachedHealthStatusKey];

  @override
  String? get selectedTeamCode => _data[CacheStore.selectedTeamKey];

  @override
  Future<void> setCachedHealthStatus(String status) async {
    _data[CacheStore.cachedHealthStatusKey] = status;
  }

  @override
  Future<void> setSelectedTeamCode(String? teamCode) async {
    if (teamCode == null) {
      _data.remove(CacheStore.selectedTeamKey);
      return;
    }
    _data[CacheStore.selectedTeamKey] = teamCode;
  }
}

class FakeHealthRepository implements HealthRepository {
  FakeHealthRepository({this.status = 'ok'});

  final String status;

  @override
  Future<HealthResponse> fetchHealth() async => HealthResponse(status: status);

  @override
  HealthResponse? getCachedHealth() => null;
}
