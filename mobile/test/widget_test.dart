import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:bigfly_mobile/data/models/meta_models.dart';
import 'package:bigfly_mobile/data/models/player_models.dart';
import 'package:bigfly_mobile/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/data/repositories/player_repository.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('renders five-tab bottom navigation shell', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Home'), findsWidgets);
    expect(find.text('Players'), findsWidgets);
    expect(find.text('Teams'), findsWidgets);
    expect(find.text('Games'), findsWidgets);
    expect(find.text('More'), findsWidgets);
  });

  testWidgets('displays home search hero and api strip', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('What can I find?'), findsOneWidget);
    expect(find.text('Search 1871–2025 across players, teams, games, franchises, and seasons.'), findsOneWidget);
  });
}

class InMemoryCacheStore implements CacheStore {
  final Map<String, String> _data = <String, String>{};

  @override
  String? get cachedHealthStatus => _data[CacheStore.cachedHealthStatusKey];

  @override
  List<String> get recentPlayerIds {
    final raw = _data[CacheStore.recentPlayerIdsKey];
    if (raw == null || raw.isEmpty) {
      return <String>[];
    }
    return raw.split('|');
  }

  @override
  String? get selectedTeamCode => _data[CacheStore.selectedTeamKey];

  @override
  Future<void> pushRecentPlayerId(String playerId) async {
    final current = recentPlayerIds.where((id) => id != playerId).toList(growable: false);
    _data[CacheStore.recentPlayerIdsKey] = <String>[playerId, ...current].take(8).join('|');
  }

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

class FakeHomeRepository implements HomeRepository {
  @override
  Future<MetaSnapshot> fetchMeta() async {
    return MetaSnapshot(
      version: '1.0.0',
      generatedAt: DateTime.parse('2026-04-18T00:00:00Z'),
      coverage: const <String, CoverageRange>{
        'lahman': CoverageRange(from: 1871, to: 2025),
        'retrosheet': CoverageRange(from: 1871, to: 2025),
      },
      datasets: const <DatasetStatusSnapshot>[
        DatasetStatusSnapshot(
          id: 'lahman',
          name: 'Lahman',
          source: 'lahman',
          required: true,
          healthy: true,
          rowCount: 1,
          coverageFrom: 1871,
          coverageTo: 2025,
        ),
      ],
    );
  }

  @override
  Future<List<HomeSearchItem>> search(HomeEntityType entity, String query) async {
    return <HomeSearchItem>[
      HomeSearchItem(
        entity: entity,
        title: 'Result for $query',
        subtitle: 'Sample subtitle',
        id: entity == HomeEntityType.players ? 'mayswi01' : null,
      ),
    ];
  }
}

class FakePlayerRepository implements PlayerRepository {
  @override
  Future<PlayerDetailBundle> fetchPlayerDetail(String playerId) async {
    return PlayerDetailBundle(
      player: const PlayerProfile(
        id: 'mayswi01',
        firstName: 'Willie',
        lastName: 'Mays',
        birthYear: 1931,
        birthMonth: 5,
        birthDay: 6,
        birthCity: 'Westfield',
        birthState: 'AL',
        bats: 'R',
        throwsHand: 'R',
        debut: null,
        finalGame: null,
        latestSeason: 1973,
        latestTeam: 'SFG',
        positions: 'OF',
      ),
      battingSeasons: const <PlayerBattingSeason>[
        PlayerBattingSeason(
          year: 1954,
          teamId: 'NY1',
          g: 151,
          ab: 565,
          avg: 0.345,
          hr: 41,
          rbi: 110,
          obp: 0.41,
          slg: 0.667,
          ops: 1.078,
          hits: 195,
        ),
      ],
      pitchingSeasons: const <PlayerPitchingSeason>[],
      awards: const <PlayerAward>[PlayerAward(awardId: 'MVP', year: 1954, league: 'NL')],
      hallOfFameRecords: const <PlayerHallOfFameRecord>[
        PlayerHallOfFameRecord(year: 1979, votedBy: 'BBWAA', votes: 409, ballots: 432, inducted: true),
      ],
      themeTeamCode: 'SF',
    );
  }

  @override
  Future<List<PlayerSearchResult>> getRecentPlayers() async {
    return const <PlayerSearchResult>[
      PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
    ];
  }

  @override
  Future<void> rememberPlayer(String playerId) async {}

  @override
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 12}) async {
    return const <PlayerSearchResult>[
      PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
    ];
  }
}
