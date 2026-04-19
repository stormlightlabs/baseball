import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/data/repositories/player_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

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
