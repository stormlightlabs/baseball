import 'package:bigfly_mobile/features/games/application/games_types.dart';
import 'package:bigfly_mobile/features/games/data/models/game_models.dart';

class GamesState {
  const GamesState({
    required this.initialized,
    required this.status,
    required this.availableSeasons,
    required this.selectedSeason,
    required this.availableTeams,
    required this.selectedTeamId,
    required this.quickFilters,
    required this.allGames,
    required this.visibleGames,
    required this.expandedGameId,
    required this.gameDetails,
    required this.detailLoadingGameIds,
    required this.detailErrors,
    this.error,
  });

  const GamesState.initial()
    : this(
        initialized: false,
        status: GamesStatus.initial,
        availableSeasons: const <int>[],
        selectedSeason: 0,
        availableTeams: const <GameTeamOption>[],
        selectedTeamId: '',
        quickFilters: const <GamesQuickFilter>{},
        allGames: const <GameSummaryRecord>[],
        visibleGames: const <GameSummaryRecord>[],
        expandedGameId: null,
        gameDetails: const <String, GameCardDetail>{},
        detailLoadingGameIds: const <String>{},
        detailErrors: const <String, String>{},
      );

  final bool initialized;
  final GamesStatus status;
  final String? error;
  final List<int> availableSeasons;
  final int selectedSeason;
  final List<GameTeamOption> availableTeams;
  final String selectedTeamId;
  final Set<GamesQuickFilter> quickFilters;
  final List<GameSummaryRecord> allGames;
  final List<GameSummaryRecord> visibleGames;
  final String? expandedGameId;
  final Map<String, GameCardDetail> gameDetails;
  final Set<String> detailLoadingGameIds;
  final Map<String, String> detailErrors;

  bool get isLoading => status == GamesStatus.loading;

  GamesState copyWith({
    bool? initialized,
    GamesStatus? status,
    String? error,
    bool clearError = false,
    List<int>? availableSeasons,
    int? selectedSeason,
    List<GameTeamOption>? availableTeams,
    String? selectedTeamId,
    Set<GamesQuickFilter>? quickFilters,
    List<GameSummaryRecord>? allGames,
    List<GameSummaryRecord>? visibleGames,
    String? expandedGameId,
    bool clearExpandedGame = false,
    Map<String, GameCardDetail>? gameDetails,
    Set<String>? detailLoadingGameIds,
    Map<String, String>? detailErrors,
  }) {
    return GamesState(
      initialized: initialized ?? this.initialized,
      status: status ?? this.status,
      error: clearError ? null : (error ?? this.error),
      availableSeasons: availableSeasons ?? this.availableSeasons,
      selectedSeason: selectedSeason ?? this.selectedSeason,
      availableTeams: availableTeams ?? this.availableTeams,
      selectedTeamId: selectedTeamId ?? this.selectedTeamId,
      quickFilters: quickFilters ?? this.quickFilters,
      allGames: allGames ?? this.allGames,
      visibleGames: visibleGames ?? this.visibleGames,
      expandedGameId: clearExpandedGame ? null : (expandedGameId ?? this.expandedGameId),
      gameDetails: gameDetails ?? this.gameDetails,
      detailLoadingGameIds: detailLoadingGameIds ?? this.detailLoadingGameIds,
      detailErrors: detailErrors ?? this.detailErrors,
    );
  }
}
