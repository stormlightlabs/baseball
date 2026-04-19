import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../support/test_fakes.dart';

void main() {
  testWidgets('players tab renders and switches detail tabs', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Players').last);
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.arrow_forward), findsOneWidget);
    await tester.tap(find.text('Awards'));
    await tester.pumpAndSettle();
    expect(find.text('Awards & honors'), findsOneWidget);
  });

  testWidgets('players tab shows failure panel when detail load fails', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(
          detailError: Exception('boom'),
          recentPlayers: const <PlayerSearchResult>[
            PlayerSearchResult(id: 'mayswi01', name: 'Willie Mays', subtitle: 'mayswi01 · OF · 1951–1973'),
          ],
        ),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Players').last);
    await tester.pumpAndSettle();

    expect(find.text('Failed to load player'), findsOneWidget);
  });
}
