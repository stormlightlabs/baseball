import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:bigfly_mobile/features/games/data/models/game_models.dart';
import 'package:bigfly_mobile/features/games/data/repositories/game_repository.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:bigfly_mobile/features/home/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_batting_career.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_pitching_career.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_player_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/data_sources_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';
import 'package:bigfly_mobile/features/more/data/models/leaders_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/postseason_series_record.dart';
import 'package:bigfly_mobile/features/more/data/models/season_award_item.dart';
import 'package:bigfly_mobile/features/more/data/models/season_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_summary_record.dart';
import 'package:bigfly_mobile/features/more/data/repositories/more_repository.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/data/repositories/player_repository.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
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

class FakeGameRepository implements GameRepository {
  FakeGameRepository({
    this.seasons,
    this.teams,
    this.games,
    this.seasonsError,
    this.teamsError,
    this.gamesError,
    this.detailError,
    Map<String, GameCardDetail>? gameDetails,
  }) : gameDetails = gameDetails ?? <String, GameCardDetail>{defaultGames.first.id: defaultGameDetail};

  final List<int>? seasons;
  final List<GameTeamOption>? teams;
  final List<GameSummaryRecord>? games;
  final Object? seasonsError;
  final Object? teamsError;
  final Object? gamesError;
  final Object? detailError;
  final Map<String, GameCardDetail> gameDetails;

  int? lastRequestedSeason;
  String? lastRequestedTeamId;
  String? lastRequestedDetailGameId;

  @override
  Future<List<int>> listAvailableSeasons() async {
    if (seasonsError != null) {
      throw seasonsError!;
    }
    return seasons ?? defaultGameSeasons;
  }

  @override
  Future<List<GameTeamOption>> listTeamsForSeason(int season) async {
    if (teamsError != null) {
      throw teamsError!;
    }
    lastRequestedSeason = season;
    return teams ?? defaultGameTeams;
  }

  @override
  Future<GameListResponse> fetchGames({required int season, String? teamId, int perPage = 120}) async {
    if (gamesError != null) {
      throw gamesError!;
    }
    lastRequestedSeason = season;
    lastRequestedTeamId = teamId;
    final rows = games ?? defaultGames;
    return GameListResponse(games: rows, total: rows.length);
  }

  @override
  Future<GameCardDetail> fetchGameDetail(GameSummaryRecord game) async {
    if (detailError != null) {
      throw detailError!;
    }
    lastRequestedDetailGameId = game.id;
    return gameDetails[game.id] ?? defaultGameDetail;
  }
}

class FakeMoreRepository implements MoreRepository {
  FakeMoreRepository({
    List<SeasonSummaryRecord>? seasons,
    SeasonSnapshot? seasonSnapshot,
    LeadersSnapshot? leadersSnapshot,
    ComparePlayerSnapshot? playerA,
    ComparePlayerSnapshot? playerB,
    DataSourcesSnapshot? dataSourcesSnapshot,
    this.seasonsError,
    this.seasonError,
    this.leadersError,
    this.compareError,
    this.dataSourcesError,
  }) : seasons = seasons ?? defaultSeasonSummaries,
       seasonSnapshot = seasonSnapshot ?? defaultSeasonSnapshot,
       leadersSnapshot = leadersSnapshot ?? defaultLeadersSnapshot,
       playerA = playerA ?? defaultComparePlayerA,
       playerB = playerB ?? defaultComparePlayerB,
       dataSourcesSnapshot = dataSourcesSnapshot ?? defaultDataSourcesSnapshot;

  final List<SeasonSummaryRecord> seasons;
  final SeasonSnapshot seasonSnapshot;
  final LeadersSnapshot leadersSnapshot;
  final ComparePlayerSnapshot playerA;
  final ComparePlayerSnapshot playerB;
  final DataSourcesSnapshot dataSourcesSnapshot;
  final Object? seasonsError;
  final Object? seasonError;
  final Object? leadersError;
  final Object? compareError;
  final Object? dataSourcesError;

  @override
  Future<SeasonSnapshot> fetchSeasonSnapshot({required int year, required String leagueFilter}) async {
    if (seasonError != null) {
      throw seasonError!;
    }
    return seasonSnapshot;
  }

  @override
  Future<List<SeasonSummaryRecord>> listSeasons() async {
    if (seasonsError != null) {
      throw seasonsError!;
    }
    return seasons;
  }

  @override
  Future<LeadersSnapshot> fetchLeaders({
    required LeadersMode mode,
    required LeadersScope scope,
    required int season,
    required String stat,
    String? league,
    int page = 1,
    int perPage = 15,
  }) async {
    if (leadersError != null) {
      throw leadersError!;
    }
    return leadersSnapshot;
  }

  @override
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 8}) async {
    if (compareError != null) {
      throw compareError!;
    }
    return const <PlayerSearchResult>[
      PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
      PlayerSearchResult(id: 'ruthba01', name: 'Babe Ruth', subtitle: 'ruthba01 · OF · 1914–1935'),
    ];
  }

  @override
  Future<ComparePlayerSnapshot> fetchComparePlayer(String playerId) async {
    if (compareError != null) {
      throw compareError!;
    }
    if (playerId == playerA.profile.id) {
      return playerA;
    }
    if (playerId == playerB.profile.id) {
      return playerB;
    }
    return playerA;
  }

  @override
  Future<DataSourcesSnapshot> fetchDataSources() async {
    if (dataSourcesError != null) {
      throw dataSourcesError!;
    }
    return dataSourcesSnapshot;
  }
}

class FakeScorecardRepository implements ScorecardRepository {
  FakeScorecardRepository({List<ScorecardGameSummary>? games}) : _games = games ?? <ScorecardGameSummary>[];

  final List<ScorecardGameSummary> _games;

  @override
  Future<void> appendPlay(ScorecardPlayDraft draft) async {}

  @override
  Future<void> createGame(ScorecardGameDraft draft) async {
    _games.insert(
      0,
      ScorecardGameSummary(
        uuid: draft.uuid,
        awayTeamName: draft.awayTeamName,
        awayTeamAbbreviation: draft.awayTeamAbbreviation,
        homeTeamName: draft.homeTeamName,
        homeTeamAbbreviation: draft.homeTeamAbbreviation,
        venue: draft.venue,
        gameDate: draft.gameDate,
        status: draft.status,
        awayScore: 0,
        homeScore: 0,
        pitchCount: 0,
        lastModifiedAt: DateTime.now(),
      ),
    );
  }

  @override
  Future<ScorecardGameDetail?> getGame(String uuid) async {
    final match = _games.where((game) => game.uuid == uuid).toList(growable: false);
    if (match.isEmpty) {
      return null;
    }
    return ScorecardGameDetail(
      summary: match.first,
      lineups: const <ScorecardLineupSlot>[],
      innings: const <ScorecardInning>[],
    );
  }

  @override
  Future<List<ScorecardGameSummary>> listGames({ScorecardStatusFilter filter = ScorecardStatusFilter.all}) async {
    switch (filter) {
      case ScorecardStatusFilter.all:
        return List<ScorecardGameSummary>.from(_games);
      case ScorecardStatusFilter.inProgress:
        return _games.where((game) => game.status == ScorecardStatus.inProgress).toList(growable: false);
      case ScorecardStatusFilter.finalGame:
        return _games.where((game) => game.status == ScorecardStatus.finalGame).toList(growable: false);
    }
  }

  @override
  Future<bool> undoLastPlay({required String gameUuid, required ScorecardGameRuntimeState restoredState}) async {
    return true;
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

const List<int> defaultGameSeasons = <int>[2025, 2024, 2023];

const List<GameTeamOption> defaultGameTeams = <GameTeamOption>[
  GameTeamOption(teamId: 'CHN', name: 'Chicago Cubs'),
  GameTeamOption(teamId: 'SLN', name: 'St. Louis Cardinals'),
];

final List<GameSummaryRecord> defaultGames = <GameSummaryRecord>[
  GameSummaryRecord(
    id: '199809270SLN',
    season: 1998,
    date: DateTime.parse('1998-09-27'),
    awayTeam: 'CHN',
    homeTeam: 'SLN',
    awayScore: 3,
    homeScore: 6,
    innings: 9,
    attendance: 47994,
    durationMin: 178,
    parkName: 'Busch Stadium',
    isPostseason: false,
  ),
  GameSummaryRecord(
    id: '199807121BOS',
    season: 1998,
    date: DateTime.parse('1998-07-12'),
    awayTeam: 'NYA',
    homeTeam: 'BOS',
    awayScore: 5,
    homeScore: 4,
    innings: 14,
    attendance: 33871,
    durationMin: 205,
    parkName: 'Fenway Park',
    isPostseason: false,
  ),
];

const GameCardDetail defaultGameDetail = GameCardDetail(
  awayWinProbability: 0.35,
  homeWinProbability: 0.65,
  keyPlays: <GameEvent>[
    GameEvent(playNum: 12, inning: 3, topBot: 1, scoreVis: 1, scoreHome: 1, event: 'Sosa homers to left.', runs: 1),
    GameEvent(
      playNum: 24,
      inning: 5,
      topBot: 1,
      scoreVis: 2,
      scoreHome: 2,
      event: 'McGwire homers to center.',
      runs: 1,
    ),
    GameEvent(
      playNum: 33,
      inning: 7,
      topBot: 1,
      scoreVis: 3,
      scoreHome: 5,
      event: 'McGwire doubles home two runs.',
      runs: 2,
    ),
  ],
);

const List<SeasonSummaryRecord> defaultSeasonSummaries = <SeasonSummaryRecord>[
  SeasonSummaryRecord(year: 2025, leagues: <String>['AL', 'NL'], teamCount: 30),
  SeasonSummaryRecord(year: 2024, leagues: <String>['AL', 'NL'], teamCount: 30),
];

const LeadersSnapshot defaultLeadersSnapshot = LeadersSnapshot(
  stat: 'hr',
  page: 1,
  perPage: 15,
  total: 3,
  entries: <LeaderboardEntry>[
    LeaderboardEntry(
      playerId: 'judgear01',
      playerName: 'Aaron Judge',
      teamId: 'NYA',
      league: 'AL',
      year: 2024,
      rawValue: 58,
      displayValue: '58',
    ),
    LeaderboardEntry(
      playerId: 'ohtansh01',
      playerName: 'Shohei Ohtani',
      teamId: 'LAN',
      league: 'NL',
      year: 2024,
      rawValue: 54,
      displayValue: '54',
    ),
    LeaderboardEntry(
      playerId: 'olsonma02',
      playerName: 'Matt Olson',
      teamId: 'ATL',
      league: 'NL',
      year: 2024,
      rawValue: 49,
      displayValue: '49',
    ),
  ],
);

const SeasonSnapshot defaultSeasonSnapshot = SeasonSnapshot(
  year: 2024,
  teams: <TeamSeasonRecord>[defaultTeamSeason],
  totalHomeRuns: 5450,
  leagueAverage: 0.244,
  avgGamesPerTeam: 162,
  hrLeaders: <LeaderboardEntry>[
    LeaderboardEntry(
      playerId: 'judgear01',
      playerName: 'Aaron Judge',
      teamId: 'NYA',
      league: 'AL',
      year: 2024,
      rawValue: 58,
      displayValue: '58',
    ),
  ],
  avgLeaders: <LeaderboardEntry>[
    LeaderboardEntry(
      playerId: 'arrealu01',
      playerName: 'Luis Arraez',
      teamId: 'MIA',
      league: 'NL',
      year: 2024,
      rawValue: 0.331,
      displayValue: '.331',
    ),
  ],
  awards: <SeasonAwardItem>[SeasonAwardItem(awardId: 'MVP', playerId: 'judgear01', year: 2024, league: 'AL')],
  postseasonSeries: <PostseasonSeriesRecord>[
    PostseasonSeriesRecord(round: 'WS', winnerTeam: 'LAN', loserTeam: 'NYA', wins: 4, losses: 2, ties: 0),
  ],
);

const ComparePlayerSnapshot defaultComparePlayerA = ComparePlayerSnapshot(
  profile: PlayerProfile(
    id: 'mayswi01',
    firstName: 'Willie',
    lastName: 'Mays',
    birthYear: 1931,
    birthMonth: null,
    birthDay: null,
    birthCity: null,
    birthState: null,
    bats: 'R',
    throwsHand: 'R',
    debut: null,
    finalGame: null,
    latestSeason: 1973,
    latestTeam: 'SFN',
    positions: 'OF',
  ),
  battingCareer: CompareBattingCareer(hr: 660, avg: 0.302, ops: 0.941, rbi: 1903, sb: 338, hits: 3283, ab: 10881),
  pitchingCareer: ComparePitchingCareer(wins: 0, losses: 0, era: 0, strikeouts: 0, whip: 0, kPer9: 0, ipOuts: 0),
);

const ComparePlayerSnapshot defaultComparePlayerB = ComparePlayerSnapshot(
  profile: PlayerProfile(
    id: 'ruthba01',
    firstName: 'Babe',
    lastName: 'Ruth',
    birthYear: 1895,
    birthMonth: null,
    birthDay: null,
    birthCity: null,
    birthState: null,
    bats: 'L',
    throwsHand: 'L',
    debut: null,
    finalGame: null,
    latestSeason: 1935,
    latestTeam: 'BOS',
    positions: 'OF,P',
  ),
  battingCareer: CompareBattingCareer(hr: 714, avg: 0.342, ops: 1.164, rbi: 2213, sb: 123, hits: 2873, ab: 8399),
  pitchingCareer: ComparePitchingCareer(
    wins: 94,
    losses: 46,
    era: 2.28,
    strikeouts: 488,
    whip: 1.16,
    kPer9: 3.46,
    ipOuts: 3661,
  ),
);

final DataSourcesSnapshot defaultDataSourcesSnapshot = DataSourcesSnapshot(
  meta: defaultMetaSnapshot,
  datasets: defaultMetaSnapshot.datasets,
);
