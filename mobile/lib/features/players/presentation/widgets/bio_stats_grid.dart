import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';

class BioStatsGrid extends StatelessWidget {
  const BioStatsGrid({super.key, required this.detail});

  final PlayerDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    final batting = detail.battingSeasons;
    final hr = batting.fold<int>(0, (sum, season) => sum + season.hr);
    final rbi = batting.fold<int>(0, (sum, season) => sum + season.rbi);
    final hits = batting.fold<int>(0, (sum, season) => sum + season.hits);
    final atBats = batting.fold<int>(0, (sum, season) => sum + season.ab);
    final avg = atBats == 0 ? 0.0 : hits / atBats;
    final years = <int>{...batting.map((season) => season.year)}.length;

    final cells = <(String, String)>[
      (hr.toString(), 'HR'),
      (avg.toStringAsFixed(3), 'AVG'),
      (rbi.toString(), 'RBI'),
      (years.toString(), 'YRS'),
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: cells.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 4,
        mainAxisSpacing: 1,
        crossAxisSpacing: 1,
        childAspectRatio: 1.5,
      ),
      itemBuilder: (context, index) {
        final (value, label) = cells[index];
        return Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Theme.of(context).dividerColor),
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Text(value, style: Theme.of(context).textTheme.titleSmall),
              Text(label, style: Theme.of(context).extension<AppTypography>()?.code),
            ],
          ),
        );
      },
    );
  }
}
