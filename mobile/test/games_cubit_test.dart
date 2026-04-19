import 'package:bigfly_mobile/features/games/application/games_cubit.dart';
import 'package:bigfly_mobile/features/games/application/games_types.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/test_fakes.dart';

void main() {
  test('initialize hydrates seasons, teams, and games', () async {
    final repo = FakeGameRepository();
    final cubit = GamesCubit(repo);

    await cubit.initialize();

    expect(cubit.state.initialized, isTrue);
    expect(cubit.state.status, GamesStatus.ready);
    expect(cubit.state.availableSeasons, isNotEmpty);
    expect(cubit.state.availableTeams, isNotEmpty);
    expect(cubit.state.visibleGames, isNotEmpty);
    expect(repo.lastRequestedSeason, cubit.state.selectedSeason);
  });

  test('quick filter narrows visible games', () async {
    final cubit = GamesCubit(FakeGameRepository());
    await cubit.initialize();

    cubit.toggleQuickFilter(GamesQuickFilter.extraInnings);

    expect(cubit.state.quickFilters, contains(GamesQuickFilter.extraInnings));
    expect(cubit.state.visibleGames.length, 1);
    expect(cubit.state.visibleGames.first.innings, greaterThan(9));
  });

  test('selectTeam reloads games for selected team', () async {
    final repo = FakeGameRepository();
    final cubit = GamesCubit(repo);
    await cubit.initialize();

    await cubit.selectTeam('SLN');

    expect(cubit.state.selectedTeamId, 'SLN');
    expect(repo.lastRequestedTeamId, 'SLN');
    expect(cubit.state.status, GamesStatus.ready);
  });

  test('toggleExpandedGame stores detail errors', () async {
    final cubit = GamesCubit(FakeGameRepository(detailError: Exception('detail failed')));
    await cubit.initialize();

    final game = cubit.state.visibleGames.first;
    await cubit.toggleExpandedGame(game);

    expect(cubit.state.expandedGameId, game.id);
    expect(cubit.state.detailErrors[game.id], contains('detail failed'));
  });
}
