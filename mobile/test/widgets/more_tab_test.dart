import 'package:bigfly_mobile/app/app.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/compare_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/data_sources_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/leaders_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/seasons_screen.dart';
import 'package:bigfly_mobile/features/scorekeeper/presentation/screens/scorecard_hub_screen.dart';
import 'package:flutter_test/flutter_test.dart';

import '../support/test_fakes.dart';

void main() {
  testWidgets('more tab renders seasons, leaders, compare, and data sources screens', (WidgetTester tester) async {
    await tester.pumpWidget(
      BigFlyApp(
        cacheStore: InMemoryCacheStore(),
        homeRepository: FakeHomeRepository(),
        playerRepository: FakePlayerRepository(),
        teamRepository: FakeTeamRepository(),
        gameRepository: FakeGameRepository(),
        moreRepository: FakeMoreRepository(),
        scorecardRepository: FakeScorecardRepository(),
        useDynamicColor: false,
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('More').last);
    await tester.pumpAndSettle();

    expect(find.text('Scorekeeper'), findsOneWidget);
    await tester.tap(find.text('Scorekeeper'));
    await tester.pumpAndSettle();
    expect(find.byType(ScorecardHubScreen), findsOneWidget);
    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(find.byType(SeasonsScreen), findsOneWidget);

    await tester.tap(find.text('Leaders').last);
    await tester.pumpAndSettle();
    expect(find.byType(LeadersScreen), findsOneWidget);

    await tester.tap(find.text('Compare').last);
    await tester.pumpAndSettle();
    expect(find.byType(CompareScreen), findsOneWidget);

    await tester.tap(find.text('Data Sources').last);
    await tester.pumpAndSettle();
    expect(find.byType(DataSourcesScreen), findsOneWidget);
  });
}
