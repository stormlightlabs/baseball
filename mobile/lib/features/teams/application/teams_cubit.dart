import 'package:bigfly_mobile/features/teams/application/teams_state.dart';
import 'package:bigfly_mobile/features/teams/application/teams_types.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:bigfly_mobile/features/teams/data/repositories/team_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class TeamsCubit extends Cubit<TeamsState> {
  TeamsCubit(this._repository) : super(const TeamsState.initial());

  final TeamRepository _repository;

  Future<void> initialize({String? initialTeamCode}) async {
    if (state.initialized) {
      return;
    }

    emit(state.copyWith(initialized: true));
    await _hydrateFranchises();

    TeamSeasonRecord? seed;

    try {
      seed = await _repository.seedTeamForCode(initialTeamCode);
    } catch (_) {
      seed = null;
    }

    if (seed == null) {
      for (final franchise in state.franchises) {
        try {
          seed = await _repository.seedTeamForCode(franchise.id);
        } catch (_) {
          seed = null;
        }

        if (seed != null) {
          break;
        }
      }
    }

    if (seed != null) {
      await loadTeam(seed);
    }
  }

  void setSearchQuery(String value) {
    emit(state.copyWith(searchQuery: value));
  }

  Future<void> runSearch() async {
    final query = state.searchQuery.trim();
    if (query.isEmpty) {
      emit(state.copyWith(searchResults: const <TeamSeasonRecord>[], searchError: null, searchLoading: false));
      return;
    }

    emit(state.copyWith(searchLoading: true, searchError: null));

    try {
      final results = await _repository.searchTeams(query, limit: 12);
      emit(state.copyWith(searchLoading: false, searchResults: results, searchError: null));
    } catch (error) {
      emit(
        state.copyWith(searchLoading: false, searchResults: const <TeamSeasonRecord>[], searchError: error.toString()),
      );
    }
  }

  Future<void> loadTeam(TeamSeasonRecord team) async {
    emit(state.copyWith(detailStatus: TeamDetailStatus.loading, detailError: null, selectedTeam: team));

    try {
      final detail = await _repository.fetchTeamDetail(team);
      emit(
        state.copyWith(
          detailStatus: TeamDetailStatus.ready,
          detail: detail,
          detailError: null,
          selectedTeam: detail.team,
        ),
      );
    } catch (error) {
      emit(state.copyWith(detailStatus: TeamDetailStatus.failure, detailError: error.toString(), selectedTeam: team));
    }
  }

  Future<void> loadTeamForCode(String teamCode) async {
    try {
      final seed = await _repository.seedTeamForCode(teamCode);
      if (seed == null) {
        return;
      }
      await loadTeam(seed);
    } catch (error) {
      emit(state.copyWith(detailStatus: TeamDetailStatus.failure, detailError: error.toString()));
    }
  }

  void setActiveSegment(TeamDetailSegment segment) {
    emit(state.copyWith(activeSegment: segment));
  }

  Future<void> _hydrateFranchises() async {
    try {
      final franchises = await _repository.listFranchises(active: true);
      emit(state.copyWith(franchises: franchises));
    } catch (_) {
      // Best-effort list for quick switching.
    }
  }
}
