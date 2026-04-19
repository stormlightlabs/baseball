import 'package:bigfly_mobile/app/app.dart';
import 'package:flutter_test/flutter_test.dart';

import '../support/test_fakes.dart';

void main() {
  testWidgets('teams tab renders team detail and segment panels', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        teamRepository: FakeTeamRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Teams').last);
    await tester.pumpAndSettle();

    expect(find.text('New York Yankees'), findsOneWidget);
    expect(find.textContaining('run differential by month'), findsOneWidget);

    await tester.tap(find.text('Roster'));
    await tester.pumpAndSettle();
    expect(find.textContaining('Current roster'), findsOneWidget);

    await tester.tap(find.text('Schedule'));
    await tester.pumpAndSettle();
    expect(find.text('Recent games'), findsOneWidget);

    await tester.tap(find.text('Daily'));
    await tester.pumpAndSettle();
    expect(find.text('Daily team logs'), findsOneWidget);
  });

  testWidgets('teams tab shows failure panel when detail load fails', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        teamRepository: FakeTeamRepository(detailError: Exception('boom')),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Teams').last);
    await tester.pumpAndSettle();

    expect(find.text('Failed to load team'), findsOneWidget);
  });
}
