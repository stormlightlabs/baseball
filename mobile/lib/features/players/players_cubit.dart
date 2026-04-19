import 'package:bigfly_mobile/data/models/player_models.dart';
import 'package:bigfly_mobile/data/repositories/player_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

enum PlayerDetailStatus { initial, loading, ready, failure }

enum PlayerDetailTab { batting, pitching, awards, hallOfFame }

enum BattingChartMetric { hr, avg, rbi, obp, slg, ops }

enum PitchingChartMetric { era, strikeouts, whip, wins, kPer9 }

enum BattingSortColumn { year, team, g, ab, avg, hr, rbi, ops }

enum PitchingSortColumn { year, team, g, wins, losses, era, so, whip }

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

class PlayersCubit extends Cubit<PlayersState> {
  PlayersCubit(this._repository) : super(const PlayersState.initial());

  final PlayerRepository _repository;

  Future<void> initialize({String? initialPlayerId}) async {
    if (state.initialized) {
      if (initialPlayerId != null && initialPlayerId != state.selectedPlayerId) {
        await loadPlayer(initialPlayerId);
      }
      return;
    }

    emit(state.copyWith(initialized: true));
    await _refreshRecentPlayers();

    if (initialPlayerId != null && initialPlayerId.isNotEmpty) {
      await loadPlayer(initialPlayerId);
      return;
    }

    if (state.recentPlayers.isNotEmpty) {
      await loadPlayer(state.recentPlayers.first.id);
      return;
    }

    final seeds = await _repository.searchPlayers('mays', limit: 5);
    if (seeds.isNotEmpty) {
      await loadPlayer(seeds.first.id);
    }
  }

  void setDetailTab(PlayerDetailTab tab) {
    emit(state.copyWith(detailTab: tab));
  }

  void setSearchQuery(String value) {
    emit(state.copyWith(searchQuery: value));
  }

  Future<void> runSearch() async {
    final query = state.searchQuery.trim();
    if (query.isEmpty) {
      emit(state.copyWith(searchResults: const <PlayerSearchResult>[], searchError: null, searchLoading: false));
      return;
    }

    emit(state.copyWith(searchLoading: true, searchError: null));

    try {
      final results = await _repository.searchPlayers(query, limit: 12);
      emit(state.copyWith(searchLoading: false, searchResults: results, searchError: null));
    } catch (error) {
      emit(
        state.copyWith(
          searchLoading: false,
          searchResults: const <PlayerSearchResult>[],
          searchError: error.toString(),
        ),
      );
    }
  }

  Future<void> loadPlayer(String playerId) async {
    emit(state.copyWith(detailStatus: PlayerDetailStatus.loading, selectedPlayerId: playerId, detailError: null));

    try {
      final detail = await _repository.fetchPlayerDetail(playerId);
      await _repository.rememberPlayer(playerId);
      await _refreshRecentPlayers();
      emit(
        state.copyWith(
          detailStatus: PlayerDetailStatus.ready,
          detail: detail,
          detailError: null,
          selectedPlayerId: playerId,
        ),
      );
    } catch (error) {
      emit(
        state.copyWith(
          detailStatus: PlayerDetailStatus.failure,
          detailError: error.toString(),
          selectedPlayerId: playerId,
        ),
      );
    }
  }

  Future<void> _refreshRecentPlayers() async {
    try {
      final recent = await _repository.getRecentPlayers();
      emit(state.copyWith(recentPlayers: recent));
    } catch (_) {
      // Best effort cache hydration.
    }
  }

  void setBattingMetric(BattingChartMetric metric) {
    emit(state.copyWith(battingMetric: metric));
  }

  void setPitchingMetric(PitchingChartMetric metric) {
    emit(state.copyWith(pitchingMetric: metric));
  }

  void setBattingSort(BattingSortColumn column, bool ascending) {
    emit(state.copyWith(battingSortColumn: column, battingSortAscending: ascending));
  }

  void setPitchingSort(PitchingSortColumn column, bool ascending) {
    emit(state.copyWith(pitchingSortColumn: column, pitchingSortAscending: ascending));
  }
}
