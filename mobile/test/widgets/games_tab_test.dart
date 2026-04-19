import 'package:bigfly_mobile/app/app.dart';
import 'package:flutter_test/flutter_test.dart';

import '../support/test_fakes.dart';

void main() {
  testWidgets('games tab renders filter strip and expandable game detail', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        teamRepository: FakeTeamRepository(),
        gameRepository: FakeGameRepository(),
        moreRepository: FakeMoreRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Games').last);
    await tester.pumpAndSettle();

    expect(find.text('Game Finder'), findsOneWidget);
    expect(find.text('CHN @ SLN'), findsOneWidget);

    await tester.tap(find.text('CHN @ SLN'));
    await tester.pumpAndSettle();

    expect(find.text('Final win probability'), findsOneWidget);
    expect(find.text('Key plays'), findsOneWidget);
  });

  testWidgets('games tab shows detail fetch errors inside expanded card', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        teamRepository: FakeTeamRepository(),
        gameRepository: FakeGameRepository(detailError: Exception('detail failed')),
        moreRepository: FakeMoreRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Games').last);
    await tester.pumpAndSettle();

    await tester.tap(find.text('CHN @ SLN'));
    await tester.pumpAndSettle();

    expect(find.textContaining('detail failed'), findsOneWidget);
  });
}
