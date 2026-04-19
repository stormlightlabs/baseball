import 'package:bigfly_mobile/features/teams/application/teams_cubit.dart';
import 'package:bigfly_mobile/features/teams/application/teams_types.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/test_fakes.dart';

void main() {
  test('initialize hydrates franchises and loads seeded team detail', () async {
    final repo = FakeTeamRepository();
    final cubit = TeamsCubit(repo);

    await cubit.initialize(initialTeamCode: 'NYY');

    expect(cubit.state.initialized, isTrue);
    expect(cubit.state.franchises, isNotEmpty);
    expect(cubit.state.detailStatus, TeamDetailStatus.ready);
    expect(repo.lastSeedCode, 'NYY');
    expect(repo.lastLoadedTeam?.teamId, defaultTeamSeason.teamId);
  });

  test('runSearch handles failures', () async {
    final cubit = TeamsCubit(FakeTeamRepository(searchError: Exception('search failed')));

    cubit.setSearchQuery('yankees');
    await cubit.runSearch();

    expect(cubit.state.searchLoading, isFalse);
    expect(cubit.state.searchError, contains('search failed'));
    expect(cubit.state.searchResults, isEmpty);
  });

  test('loadTeam handles repository errors', () async {
    final cubit = TeamsCubit(FakeTeamRepository(detailError: Exception('detail failed')));

    await cubit.loadTeam(defaultTeamSeason);

    expect(cubit.state.detailStatus, TeamDetailStatus.failure);
    expect(cubit.state.detailError, contains('detail failed'));
  });

  test('segment mutation updates state', () {
    final cubit = TeamsCubit(FakeTeamRepository());

    cubit.setActiveSegment(TeamDetailSegment.schedule);

    expect(cubit.state.activeSegment, TeamDetailSegment.schedule);
  });
}
