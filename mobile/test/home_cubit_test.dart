import 'package:bigfly_mobile/features/home/application/home_cubit.dart';
import 'package:bigfly_mobile/features/home/application/home_state.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/test_fakes.dart';

void main() {
  test('initialize loads meta snapshot', () async {
    final cubit = HomeCubit(FakeHomeRepository());

    await cubit.initialize();

    expect(cubit.state.metaStatus, HomeMetaStatus.ready);
    expect(cubit.state.meta, isNotNull);
    expect(cubit.state.metaError, isNull);
  });

  test('loadMeta surfaces failures', () async {
    final cubit = HomeCubit(FakeHomeRepository(metaError: Exception('meta failure')));

    await cubit.loadMeta();

    expect(cubit.state.metaStatus, HomeMetaStatus.failure);
    expect(cubit.state.metaError, contains('meta failure'));
  });

  test('runSearch updates ready state with results', () async {
    final repo = FakeHomeRepository();
    final cubit = HomeCubit(repo);

    cubit.setEntity(HomeEntityType.players);
    cubit.setQuery('mays');
    await cubit.runSearch();

    expect(cubit.state.searchStatus, HomeSearchStatus.ready);
    expect(cubit.state.searchResults, isNotEmpty);
    expect(repo.lastQuery, 'mays');
  });

  test('runSearch handles failures', () async {
    final cubit = HomeCubit(FakeHomeRepository(searchError: Exception('search failed')));

    cubit.setQuery('mays');
    await cubit.runSearch();

    expect(cubit.state.searchStatus, HomeSearchStatus.failure);
    expect(cubit.state.searchError, contains('search failed'));
    expect(cubit.state.searchResults, isEmpty);
  });
}
