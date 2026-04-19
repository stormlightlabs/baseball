import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/data/repositories/more_repository.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_stat_option.dart';
import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class MoreCubit extends Cubit<MoreState> {
  MoreCubit(this._repository) : super(const MoreState.initial());

  final MoreRepository _repository;

  Future<void> initialize() async {
    if (state.availableSeasons.isNotEmpty) {
      return;
    }

    emit(
      state.copyWith(
        seasonStatus: MoreStatus.loading,
        leadersStatus: MoreStatus.loading,
        dataSourcesStatus: MoreStatus.loading,
        seasonError: null,
        leadersError: null,
        dataSourcesError: null,
      ),
    );

    try {
      final seasons = await _repository.listSeasons();
      final selectedYear = seasons.isEmpty ? DateTime.now().year : seasons.first.year;
      emit(state.copyWith(availableSeasons: seasons, selectedSeason: selectedYear));

      await Future.wait(<Future<void>>[loadSeasonSnapshot(), loadLeaders(), loadDataSources(), _seedCompareDefaults()]);
    } catch (error) {
      emit(
        state.copyWith(
          seasonStatus: MoreStatus.failure,
          leadersStatus: MoreStatus.failure,
          dataSourcesStatus: MoreStatus.failure,
          seasonError: error.toString(),
          leadersError: error.toString(),
          dataSourcesError: error.toString(),
        ),
      );
    }
  }

  void setSection(MoreSection section) {
    emit(state.copyWith(selectedSection: section));
  }

  Future<void> selectSeason(int season) async {
    if (season == state.selectedSeason) {
      return;
    }

    emit(state.copyWith(selectedSeason: season));
    await Future.wait(<Future<void>>[
      loadSeasonSnapshot(),
      if (state.leadersScope == LeadersScope.season) loadLeaders(),
    ]);
  }

  Future<void> setSeasonLeagueFilter(String filter) async {
    if (filter == state.seasonLeagueFilter) {
      return;
    }

    emit(state.copyWith(seasonLeagueFilter: filter));
    await Future.wait(<Future<void>>[
      loadSeasonSnapshot(),
      if (state.leadersScope == LeadersScope.season) loadLeaders(),
    ]);
  }

  Future<void> setLeadersMode(LeadersMode mode) async {
    if (mode == state.leadersMode) {
      return;
    }

    final options = mode == LeadersMode.batting ? battingStatOptions : pitchingStatOptions;
    final hasStat = options.any((option) => option.key == state.leadersStat);

    emit(state.copyWith(leadersMode: mode, leadersStat: hasStat ? state.leadersStat : options.first.key));
    await loadLeaders();
  }

  Future<void> setLeadersScope(LeadersScope scope) async {
    if (scope == state.leadersScope) {
      return;
    }

    emit(state.copyWith(leadersScope: scope));
    await loadLeaders();
  }

  Future<void> setLeadersStat(String stat) async {
    if (stat == state.leadersStat) {
      return;
    }

    emit(state.copyWith(leadersStat: stat));
    await loadLeaders();
  }

  Future<void> loadSeasonSnapshot() async {
    if (state.selectedSeason <= 0) {
      return;
    }

    emit(state.copyWith(seasonStatus: MoreStatus.loading, seasonError: null));
    try {
      final snapshot = await _repository.fetchSeasonSnapshot(
        year: state.selectedSeason,
        leagueFilter: state.seasonLeagueFilter,
      );
      emit(state.copyWith(seasonStatus: MoreStatus.ready, seasonSnapshot: snapshot, seasonError: null));
    } catch (error) {
      emit(state.copyWith(seasonStatus: MoreStatus.failure, seasonError: error.toString()));
    }
  }

  Future<void> loadLeaders() async {
    if (state.selectedSeason <= 0) {
      return;
    }

    emit(state.copyWith(leadersStatus: MoreStatus.loading, leadersError: null));

    try {
      final league = state.leadersScope == LeadersScope.season && state.seasonLeagueFilter != 'Both'
          ? state.seasonLeagueFilter
          : null;

      final snapshot = await _repository.fetchLeaders(
        mode: state.leadersMode,
        scope: state.leadersScope,
        season: state.selectedSeason,
        stat: state.leadersStat,
        league: league,
        perPage: 15,
      );

      emit(state.copyWith(leadersStatus: MoreStatus.ready, leadersSnapshot: snapshot, leadersError: null));
    } catch (error) {
      emit(state.copyWith(leadersStatus: MoreStatus.failure, leadersError: error.toString()));
    }
  }

  Future<void> loadDataSources() async {
    emit(state.copyWith(dataSourcesStatus: MoreStatus.loading, dataSourcesError: null));

    try {
      final snapshot = await _repository.fetchDataSources();
      emit(state.copyWith(dataSourcesStatus: MoreStatus.ready, dataSourcesSnapshot: snapshot, dataSourcesError: null));
    } catch (error) {
      emit(state.copyWith(dataSourcesStatus: MoreStatus.failure, dataSourcesError: error.toString()));
    }
  }

  Future<void> refreshSelectedSection() async {
    switch (state.selectedSection) {
      case MoreSection.seasons:
        await loadSeasonSnapshot();
        return;
      case MoreSection.leaders:
        await loadLeaders();
        return;
      case MoreSection.compare:
        await Future.wait(<Future<void>>[
          if (state.comparePlayerA != null) selectComparePlayerA(state.comparePlayerA!.profile.id),
          if (state.comparePlayerB != null) selectComparePlayerB(state.comparePlayerB!.profile.id),
        ]);
        return;
      case MoreSection.dataSources:
        await loadDataSources();
        return;
    }
  }

  Future<void> searchComparePlayersA(String query) async {
    final trimmed = query.trim();
    emit(
      state.copyWith(
        compareSearchA: query,
        compareSearchLoadingA: trimmed.isNotEmpty,
        compareSearchErrorA: null,
        compareSearchResultsA: trimmed.isEmpty ? const <PlayerSearchResult>[] : null,
      ),
    );

    if (trimmed.isEmpty) {
      emit(
        state.copyWith(
          compareSearchResultsA: const <PlayerSearchResult>[],
          compareSearchLoadingA: false,
          compareSearchErrorA: null,
        ),
      );
      return;
    }

    try {
      final results = await _repository.searchPlayers(trimmed);
      emit(state.copyWith(compareSearchResultsA: results, compareSearchLoadingA: false, compareSearchErrorA: null));
    } catch (error) {
      emit(
        state.copyWith(
          compareSearchResultsA: const <PlayerSearchResult>[],
          compareSearchLoadingA: false,
          compareSearchErrorA: error.toString(),
        ),
      );
    }
  }

  Future<void> searchComparePlayersB(String query) async {
    final trimmed = query.trim();
    emit(
      state.copyWith(
        compareSearchB: query,
        compareSearchLoadingB: trimmed.isNotEmpty,
        compareSearchErrorB: null,
        compareSearchResultsB: trimmed.isEmpty ? const <PlayerSearchResult>[] : null,
      ),
    );

    if (trimmed.isEmpty) {
      emit(
        state.copyWith(
          compareSearchResultsB: const <PlayerSearchResult>[],
          compareSearchLoadingB: false,
          compareSearchErrorB: null,
        ),
      );
      return;
    }

    try {
      final results = await _repository.searchPlayers(trimmed);
      emit(state.copyWith(compareSearchResultsB: results, compareSearchLoadingB: false, compareSearchErrorB: null));
    } catch (error) {
      emit(
        state.copyWith(
          compareSearchResultsB: const <PlayerSearchResult>[],
          compareSearchLoadingB: false,
          compareSearchErrorB: error.toString(),
        ),
      );
    }
  }

  Future<void> selectComparePlayerA(String playerId) async {
    emit(
      state.copyWith(
        compareLoadStatusA: MoreStatus.loading,
        compareErrorA: null,
        compareSearchA: '',
        compareSearchResultsA: const <PlayerSearchResult>[],
        compareSearchLoadingA: false,
      ),
    );

    try {
      final player = await _repository.fetchComparePlayer(playerId);
      emit(state.copyWith(comparePlayerA: player, compareLoadStatusA: MoreStatus.ready, compareErrorA: null));
    } catch (error) {
      emit(state.copyWith(compareLoadStatusA: MoreStatus.failure, compareErrorA: error.toString()));
    }
  }

  Future<void> selectComparePlayerB(String playerId) async {
    emit(
      state.copyWith(
        compareLoadStatusB: MoreStatus.loading,
        compareErrorB: null,
        compareSearchB: '',
        compareSearchResultsB: const <PlayerSearchResult>[],
        compareSearchLoadingB: false,
      ),
    );

    try {
      final player = await _repository.fetchComparePlayer(playerId);
      emit(state.copyWith(comparePlayerB: player, compareLoadStatusB: MoreStatus.ready, compareErrorB: null));
    } catch (error) {
      emit(state.copyWith(compareLoadStatusB: MoreStatus.failure, compareErrorB: error.toString()));
    }
  }

  Future<void> _seedCompareDefaults() async {
    try {
      final aResults = await _repository.searchPlayers('willie mays', limit: 1);
      final bResults = await _repository.searchPlayers('babe ruth', limit: 1);
      if (aResults.isNotEmpty) {
        await selectComparePlayerA(aResults.first.id);
      }
      if (bResults.isNotEmpty) {
        await selectComparePlayerB(bResults.first.id);
      }
    } catch (_) {
      // Keep compare blank if defaults are unavailable.
    }
  }
}
