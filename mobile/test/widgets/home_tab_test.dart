import 'package:bigfly_mobile/app/app.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../support/test_fakes.dart';

void main() {
  testWidgets('home tab renders quick access, featured queries, and meta strip', (WidgetTester tester) async {
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

    expect(find.text('What can I find?'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('Quick access'), 300, scrollable: find.byType(Scrollable).first);
    expect(find.text('Quick access'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('Featured queries'), 300, scrollable: find.byType(Scrollable).first);
    expect(find.text('Featured queries'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('API Online'), 300, scrollable: find.byType(Scrollable).first);
    expect(find.text('API Online'), findsOneWidget);
  });

  testWidgets('home search result tap routes to players tab', (WidgetTester tester) async {
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

    await tester.enterText(find.byType(TextField).first, 'mays');
    await tester.tap(find.text('Go'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Result for mays'));
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.arrow_forward), findsOneWidget);
  });
}
