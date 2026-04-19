import 'package:bigfly_mobile/app/app.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'support/test_fakes.dart';

void main() {
  testWidgets('renders five-tab bottom navigation shell', (WidgetTester tester) async {
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

    expect(find.text('Home'), findsWidgets);
    expect(find.text('Players'), findsWidgets);
    expect(find.text('Teams'), findsWidgets);
    expect(find.text('Games'), findsWidgets);
    expect(find.text('More'), findsWidgets);
  });

  testWidgets('bottom nav switches to players tab content', (WidgetTester tester) async {
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

    await tester.tap(find.text('Players').last);
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.arrow_forward), findsOneWidget);
  });
}
