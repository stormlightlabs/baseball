import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/test_fakes.dart';

void main() {
  test('initialize hydrates recent players and loads detail', () async {
    final repo = FakePlayerRepository(
      recentPlayers: const <PlayerSearchResult>[
        PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
      ],
    );
    final cubit = PlayersCubit(repo);

    await cubit.initialize();

    expect(cubit.state.initialized, isTrue);
    expect(cubit.state.detailStatus, PlayerDetailStatus.ready);
    expect(cubit.state.selectedPlayerId, 'mayswi01');
    expect(repo.lastFetchedPlayerId, 'mayswi01');
    expect(repo.lastRememberedPlayerId, 'mayswi01');
  });

  test('runSearch handles failures', () async {
    final cubit = PlayersCubit(FakePlayerRepository(searchError: Exception('search failed')));

    cubit.setSearchQuery('mays');
    await cubit.runSearch();

    expect(cubit.state.searchLoading, isFalse);
    expect(cubit.state.searchError, contains('search failed'));
    expect(cubit.state.searchResults, isEmpty);
  });

  test('loadPlayer handles repository errors', () async {
    final cubit = PlayersCubit(FakePlayerRepository(detailError: Exception('detail failed')));

    await cubit.loadPlayer('mayswi01');

    expect(cubit.state.detailStatus, PlayerDetailStatus.failure);
    expect(cubit.state.detailError, contains('detail failed'));
    expect(cubit.state.selectedPlayerId, 'mayswi01');
  });

  test('metric and sort mutation methods update state', () {
    final cubit = PlayersCubit(FakePlayerRepository());

    cubit.setDetailTab(PlayerDetailTab.awards);
    cubit.setBattingMetric(BattingChartMetric.ops);
    cubit.setPitchingMetric(PitchingChartMetric.whip);
    cubit.setBattingSort(BattingSortColumn.hr, false);
    cubit.setPitchingSort(PitchingSortColumn.whip, false);

    expect(cubit.state.detailTab, PlayerDetailTab.awards);
    expect(cubit.state.battingMetric, BattingChartMetric.ops);
    expect(cubit.state.pitchingMetric, PitchingChartMetric.whip);
    expect(cubit.state.battingSortColumn, BattingSortColumn.hr);
    expect(cubit.state.battingSortAscending, isFalse);
    expect(cubit.state.pitchingSortColumn, PitchingSortColumn.whip);
    expect(cubit.state.pitchingSortAscending, isFalse);
  });
}
