import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:bigfly_mobile/features/home/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/data/repositories/player_repository.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:bigfly_mobile/features/teams/data/repositories/team_repository.dart';

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
  FakeHomeRepository({
    MetaSnapshot? metaSnapshot,
    this.metaError,
    this.searchError,
    List<HomeSearchItem>? searchResults,
  }) : _metaSnapshot = metaSnapshot ?? defaultMetaSnapshot,
       _searchResults = searchResults;

  final MetaSnapshot _metaSnapshot;
  final Object? metaError;
  final Object? searchError;
  final List<HomeSearchItem>? _searchResults;

  HomeEntityType? lastEntity;
  String? lastQuery;

  @override
  Future<MetaSnapshot> fetchMeta() async {
    if (metaError != null) {
      throw metaError!;
    }
    return _metaSnapshot;
  }

  @override
  Future<List<HomeSearchItem>> search(HomeEntityType entity, String query) async {
    lastEntity = entity;
    lastQuery = query;

    if (searchError != null) {
      throw searchError!;
    }

    if (_searchResults != null) {
      return _searchResults;
    }

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
  FakePlayerRepository({
    this.searchResults,
    this.recentPlayers,
    this.searchError,
    this.detailError,
    this.recentError,
    PlayerDetailBundle? detail,
  }) : detail = detail ?? defaultPlayerDetail;

  final List<PlayerSearchResult>? searchResults;
  final List<PlayerSearchResult>? recentPlayers;
  final Object? searchError;
  final Object? detailError;
  final Object? recentError;
  final PlayerDetailBundle detail;

  String? lastRememberedPlayerId;
  String? lastFetchedPlayerId;

  @override
  Future<PlayerDetailBundle> fetchPlayerDetail(String playerId) async {
    lastFetchedPlayerId = playerId;
    if (detailError != null) {
      throw detailError!;
    }
    return detail;
  }

  @override
  Future<List<PlayerSearchResult>> getRecentPlayers() async {
    if (recentError != null) {
      throw recentError!;
    }
    return recentPlayers ??
        const <PlayerSearchResult>[
          PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
        ];
  }

  @override
  Future<void> rememberPlayer(String playerId) async {
    lastRememberedPlayerId = playerId;
  }

  @override
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 12}) async {
    if (searchError != null) {
      throw searchError!;
    }

    return searchResults ??
        const <PlayerSearchResult>[
          PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
        ];
  }
}

class FakeTeamRepository implements TeamRepository {
  FakeTeamRepository({
    this.franchises,
    this.searchResults,
    this.searchError,
    this.detailError,
    TeamDetailBundle? detail,
  }) : detail = detail ?? defaultTeamDetail;

  final List<FranchiseSummary>? franchises;
  final List<TeamSeasonRecord>? searchResults;
  final Object? searchError;
  final Object? detailError;
  final TeamDetailBundle detail;

  TeamSeasonRecord? lastLoadedTeam;
  String? lastSeedCode;

  @override
  Future<TeamDetailBundle> fetchTeamDetail(TeamSeasonRecord team) async {
    lastLoadedTeam = team;
    if (detailError != null) {
      throw detailError!;
    }
    return detail;
  }

  @override
  Future<List<FranchiseSummary>> listFranchises({bool active = true}) async {
    return franchises ??
        const <FranchiseSummary>[
          FranchiseSummary(id: 'NYY', name: 'New York Yankees', active: true, activeFrom: 1901, activeTo: null),
          FranchiseSummary(id: 'BOS', name: 'Boston Red Sox', active: true, activeFrom: 1901, activeTo: null),
        ];
  }

  @override
  Future<List<TeamSeasonRecord>> searchTeams(String query, {int limit = 12}) async {
    if (searchError != null) {
      throw searchError!;
    }

    return searchResults ?? const <TeamSeasonRecord>[defaultTeamSeason];
  }

  @override
  Future<TeamSeasonRecord?> seedTeamForCode(String? teamCode) async {
    lastSeedCode = teamCode;
    final teams = searchResults ?? const <TeamSeasonRecord>[defaultTeamSeason];
    if (teams.isEmpty) {
      return null;
    }
    return teams.first;
  }
}

final MetaSnapshot defaultMetaSnapshot = MetaSnapshot(
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

final PlayerDetailBundle defaultPlayerDetail = PlayerDetailBundle(
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

const TeamSeasonRecord defaultTeamSeason = TeamSeasonRecord(
  teamId: 'NYA',
  year: 2024,
  franchiseId: 'NYY',
  league: 'AL',
  name: 'New York Yankees',
  parkId: 'NYC16',
  games: 162,
  wins: 94,
  losses: 68,
  ties: 0,
  runsScored: 762,
  runsAllowed: 609,
  division: 'East',
);

final TeamDetailBundle defaultTeamDetail = TeamDetailBundle(
  team: defaultTeamSeason,
  franchise: const FranchiseSummary(
    id: 'NYY',
    name: 'New York Yankees',
    active: true,
    activeFrom: 1901,
    activeTo: null,
  ),
  recentSeasons: const <TeamSeasonRecord>[
    defaultTeamSeason,
    TeamSeasonRecord(
      teamId: 'NYA',
      year: 2023,
      franchiseId: 'NYY',
      league: 'AL',
      name: 'New York Yankees',
      parkId: 'NYC16',
      games: 162,
      wins: 82,
      losses: 80,
      ties: 0,
      runsScored: 712,
      runsAllowed: 697,
      division: 'East',
    ),
  ],
  roster: const <TeamRosterPlayer>[
    TeamRosterPlayer(
      playerId: 'judgear01',
      firstName: 'Aaron',
      lastName: 'Judge',
      position: 'OF',
      hr: 58,
      rbi: 144,
      avg: 0.31,
    ),
    TeamRosterPlayer(
      playerId: 'cologe01',
      firstName: 'Gerrit',
      lastName: 'Cole',
      position: 'SP',
      w: 15,
      l: 4,
      era: 2.57,
      so: 222,
    ),
  ],
  schedule: const <TeamGameSummary>[
    TeamGameSummary(
      id: 'NYY202404140',
      season: 2024,
      date: null,
      homeTeam: 'BOS',
      awayTeam: 'NYA',
      homeScore: 4,
      awayScore: 7,
      parkName: 'Fenway Park',
    ),
  ],
  dailyLogs: const <TeamDailyLog>[
    TeamDailyLog(date: null, gamesPlayed: 1, wins: 1, losses: 0, runsScored: 7, runsAllowed: 4, runDiff: 3),
  ],
  runDifferential: const TeamRunDifferentialSeries(
    entityId: 'NYA',
    season: 2024,
    gamesPlayed: 1,
    runsScored: 7,
    runsAllowed: 4,
    runDifferential: 3,
    games: <RunDifferentialGamePoint>[RunDifferentialGamePoint(date: null, differential: 3, cumulativeDiff: 3)],
  ),
  themeTeamCode: 'NYY',
);
