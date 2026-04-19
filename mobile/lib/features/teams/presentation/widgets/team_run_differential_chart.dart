import 'dart:math' as math;

import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

const List<String> _monthAbbr = <String>[
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
];

class TeamRunDifferentialChart extends StatelessWidget {
  const TeamRunDifferentialChart({super.key, required this.series});

  final TeamRunDifferentialSeries? series;

  @override
  Widget build(BuildContext context) {
    final monthDiff = <int, int>{};
    for (final point in series?.games ?? const <RunDifferentialGamePoint>[]) {
      final month = point.date?.month;
      if (month == null) {
        continue;
      }
      monthDiff[month] = (monthDiff[month] ?? 0) + point.differential;
    }

    final months = monthDiff.keys.toList(growable: false)..sort();
    if (months.isEmpty) {
      return const SizedBox(height: 160, child: Center(child: Text('No run differential data available.')));
    }

    final values = months.map((month) => monthDiff[month]!).toList(growable: false);
    final maxAbs = values.map((value) => value.abs()).reduce(math.max).toDouble();
    final yRange = maxAbs == 0 ? 1.0 : maxAbs + 2;

    return SizedBox(
      height: 170,
      child: BarChart(
        BarChartData(
          minY: -yRange,
          maxY: yRange,
          gridData: FlGridData(show: true, drawVerticalLine: false),
          borderData: FlBorderData(show: false),
          titlesData: FlTitlesData(
            topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            leftTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                reservedSize: 32,
                interval: yRange > 20 ? 10 : 5,
                getTitlesWidget: (value, _) =>
                    Text(value.toStringAsFixed(0), style: Theme.of(context).textTheme.bodySmall),
              ),
            ),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                getTitlesWidget: (value, _) {
                  final index = value.toInt();
                  if (index < 0 || index >= months.length) {
                    return const SizedBox.shrink();
                  }
                  final month = months[index];
                  return Padding(
                    padding: const EdgeInsets.only(top: 6),
                    child: Text(_monthAbbr[month - 1], style: Theme.of(context).textTheme.bodySmall),
                  );
                },
              ),
            ),
          ),
          barGroups: List<BarChartGroupData>.generate(months.length, (index) {
            final value = monthDiff[months[index]]!;
            final color = value >= 0 ? const Color(0xFF10B981) : const Color(0xFFEF4444);
            return BarChartGroupData(
              x: index,
              barRods: <BarChartRodData>[
                BarChartRodData(toY: value.toDouble(), width: 14, color: color, borderRadius: BorderRadius.circular(3)),
              ],
            );
          }),
        ),
      ),
    );
  }
}
