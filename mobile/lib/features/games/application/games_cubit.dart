import 'package:bigfly_mobile/features/games/application/games_state.dart';
import 'package:bigfly_mobile/features/games/application/games_types.dart';
import 'package:bigfly_mobile/features/games/data/models/game_models.dart';
import 'package:bigfly_mobile/features/games/data/repositories/game_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class GamesCubit extends Cubit<GamesState> {
  GamesCubit(this._repository) : super(const GamesState.initial());

  final GameRepository _repository;

  Future<void> initialize() async {
    if (state.initialized) {
      return;
    }

    emit(state.copyWith(initialized: true, status: GamesStatus.loading, clearError: true));

    try {
      final seasons = await _repository.listAvailableSeasons();
      final selectedSeason = seasons.first;
      final teamsFuture = _repository.listTeamsForSeason(selectedSeason);
      final gamesFuture = _repository.fetchGames(season: selectedSeason);

      final teams = await teamsFuture;
      final gamesResult = await gamesFuture;
      final visibleGames = _applyQuickFilters(gamesResult.games, state.quickFilters);

      emit(
        state.copyWith(
          status: GamesStatus.ready,
          clearError: true,
          availableSeasons: seasons,
          selectedSeason: selectedSeason,
          availableTeams: teams,
          selectedTeamId: '',
          allGames: gamesResult.games,
          visibleGames: visibleGames,
          clearExpandedGame: true,
          gameDetails: const <String, GameCardDetail>{},
          detailErrors: const <String, String>{},
          detailLoadingGameIds: const <String>{},
        ),
      );
    } catch (error) {
      emit(state.copyWith(status: GamesStatus.failure, error: error.toString()));
    }
  }

  Future<void> selectSeason(int season) async {
    if (season == state.selectedSeason) {
      return;
    }

    emit(
      state.copyWith(
        status: GamesStatus.loading,
        selectedSeason: season,
        selectedTeamId: '',
        availableTeams: const <GameTeamOption>[],
        allGames: const <GameSummaryRecord>[],
        visibleGames: const <GameSummaryRecord>[],
        clearError: true,
        clearExpandedGame: true,
        gameDetails: const <String, GameCardDetail>{},
        detailErrors: const <String, String>{},
        detailLoadingGameIds: const <String>{},
      ),
    );

    await _loadGamesForSelection();
  }

  Future<void> selectTeam(String teamId) async {
    if (teamId == state.selectedTeamId) {
      return;
    }

    emit(
      state.copyWith(
        status: GamesStatus.loading,
        selectedTeamId: teamId,
        clearError: true,
        clearExpandedGame: true,
        gameDetails: const <String, GameCardDetail>{},
        detailErrors: const <String, String>{},
        detailLoadingGameIds: const <String>{},
      ),
    );

    await _loadGamesOnly();
  }

  void toggleQuickFilter(GamesQuickFilter filter) {
    final nextFilters = <GamesQuickFilter>{...state.quickFilters};
    if (nextFilters.contains(filter)) {
      nextFilters.remove(filter);
    } else {
      nextFilters.add(filter);
    }

    emit(
      state.copyWith(
        quickFilters: nextFilters,
        visibleGames: _applyQuickFilters(state.allGames, nextFilters),
        clearExpandedGame: true,
      ),
    );
  }

  Future<void> toggleExpandedGame(GameSummaryRecord game) async {
    if (state.expandedGameId == game.id) {
      emit(state.copyWith(clearExpandedGame: true));
      return;
    }

    emit(state.copyWith(expandedGameId: game.id));

    if (state.gameDetails.containsKey(game.id) || state.detailLoadingGameIds.contains(game.id)) {
      return;
    }

    final loadingIds = <String>{...state.detailLoadingGameIds, game.id};
    final detailErrors = <String, String>{...state.detailErrors}..remove(game.id);
    emit(state.copyWith(detailLoadingGameIds: loadingIds, detailErrors: detailErrors));

    try {
      final detail = await _repository.fetchGameDetail(game);
      final nextDetails = <String, GameCardDetail>{...state.gameDetails, game.id: detail};
      final nextLoading = <String>{...state.detailLoadingGameIds}..remove(game.id);
      emit(state.copyWith(gameDetails: nextDetails, detailLoadingGameIds: nextLoading));
    } catch (error) {
      final nextLoading = <String>{...state.detailLoadingGameIds}..remove(game.id);
      final nextErrors = <String, String>{...state.detailErrors, game.id: error.toString()};
      emit(state.copyWith(detailLoadingGameIds: nextLoading, detailErrors: nextErrors));
    }
  }

  Future<void> _loadGamesForSelection() async {
    try {
      final season = state.selectedSeason;
      final teamsFuture = _repository.listTeamsForSeason(season);
      final gamesFuture = _repository.fetchGames(season: season, teamId: state.selectedTeamId);

      final teams = await teamsFuture;
      final gamesResult = await gamesFuture;

      emit(
        state.copyWith(
          status: GamesStatus.ready,
          clearError: true,
          availableTeams: teams,
          allGames: gamesResult.games,
          visibleGames: _applyQuickFilters(gamesResult.games, state.quickFilters),
        ),
      );
    } catch (error) {
      emit(state.copyWith(status: GamesStatus.failure, error: error.toString()));
    }
  }

  Future<void> _loadGamesOnly() async {
    try {
      final gamesResult = await _repository.fetchGames(season: state.selectedSeason, teamId: state.selectedTeamId);
      emit(
        state.copyWith(
          status: GamesStatus.ready,
          clearError: true,
          allGames: gamesResult.games,
          visibleGames: _applyQuickFilters(gamesResult.games, state.quickFilters),
        ),
      );
    } catch (error) {
      emit(state.copyWith(status: GamesStatus.failure, error: error.toString()));
    }
  }

  List<GameSummaryRecord> _applyQuickFilters(List<GameSummaryRecord> games, Set<GamesQuickFilter> quickFilters) {
    if (quickFilters.isEmpty) {
      return games;
    }

    return games
        .where((game) {
          for (final filter in quickFilters) {
            if (!_matchesFilter(game, filter)) {
              return false;
            }
          }
          return true;
        })
        .toList(growable: false);
  }

  bool _matchesFilter(GameSummaryRecord game, GamesQuickFilter filter) {
    return switch (filter) {
      GamesQuickFilter.extraInnings => game.innings > 9,
      GamesQuickFilter.doubleheaders => game.isDoubleheader,
      GamesQuickFilter.postseason => game.isLikelyPostseason,
    };
  }
}
