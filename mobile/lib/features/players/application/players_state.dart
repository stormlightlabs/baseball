import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';

class PlayersState {
  const PlayersState({
    required this.initialized,
    required this.searchQuery,
    required this.searchLoading,
    required this.searchResults,
    required this.recentPlayers,
    required this.detailStatus,
    required this.detailTab,
    required this.battingMetric,
    required this.pitchingMetric,
    required this.battingSortColumn,
    required this.battingSortAscending,
    required this.pitchingSortColumn,
    required this.pitchingSortAscending,
    this.detail,
    this.detailError,
    this.searchError,
    this.selectedPlayerId,
  });

  const PlayersState.initial()
    : this(
        initialized: false,
        searchQuery: '',
        searchLoading: false,
        searchResults: const <PlayerSearchResult>[],
        recentPlayers: const <PlayerSearchResult>[],
        detailStatus: PlayerDetailStatus.initial,
        detailTab: PlayerDetailTab.batting,
        battingMetric: BattingChartMetric.hr,
        pitchingMetric: PitchingChartMetric.era,
        battingSortColumn: BattingSortColumn.year,
        battingSortAscending: true,
        pitchingSortColumn: PitchingSortColumn.year,
        pitchingSortAscending: true,
      );

  final bool initialized;
  final String searchQuery;
  final bool searchLoading;
  final List<PlayerSearchResult> searchResults;
  final List<PlayerSearchResult> recentPlayers;
  final PlayerDetailStatus detailStatus;
  final PlayerDetailBundle? detail;
  final String? detailError;
  final String? searchError;
  final String? selectedPlayerId;
  final PlayerDetailTab detailTab;
  final BattingChartMetric battingMetric;
  final PitchingChartMetric pitchingMetric;
  final BattingSortColumn battingSortColumn;
  final bool battingSortAscending;
  final PitchingSortColumn pitchingSortColumn;
  final bool pitchingSortAscending;

  PlayersState copyWith({
    bool? initialized,
    String? searchQuery,
    bool? searchLoading,
    List<PlayerSearchResult>? searchResults,
    List<PlayerSearchResult>? recentPlayers,
    PlayerDetailStatus? detailStatus,
    PlayerDetailBundle? detail,
    String? detailError,
    String? searchError,
    String? selectedPlayerId,
    PlayerDetailTab? detailTab,
    BattingChartMetric? battingMetric,
    PitchingChartMetric? pitchingMetric,
    BattingSortColumn? battingSortColumn,
    bool? battingSortAscending,
    PitchingSortColumn? pitchingSortColumn,
    bool? pitchingSortAscending,
  }) {
    return PlayersState(
      initialized: initialized ?? this.initialized,
      searchQuery: searchQuery ?? this.searchQuery,
      searchLoading: searchLoading ?? this.searchLoading,
      searchResults: searchResults ?? this.searchResults,
      recentPlayers: recentPlayers ?? this.recentPlayers,
      detailStatus: detailStatus ?? this.detailStatus,
      detail: detail ?? this.detail,
      detailError: detailError,
      searchError: searchError,
      selectedPlayerId: selectedPlayerId ?? this.selectedPlayerId,
      detailTab: detailTab ?? this.detailTab,
      battingMetric: battingMetric ?? this.battingMetric,
      pitchingMetric: pitchingMetric ?? this.pitchingMetric,
      battingSortColumn: battingSortColumn ?? this.battingSortColumn,
      battingSortAscending: battingSortAscending ?? this.battingSortAscending,
      pitchingSortColumn: pitchingSortColumn ?? this.pitchingSortColumn,
      pitchingSortAscending: pitchingSortAscending ?? this.pitchingSortAscending,
    );
  }
}
