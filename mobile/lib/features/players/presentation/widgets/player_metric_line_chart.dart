import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

class PlayerMetricLineChart extends StatelessWidget {
  const PlayerMetricLineChart({super.key, required this.points});

  final List<FlSpot> points;

  @override
  Widget build(BuildContext context) {
    if (points.isEmpty) {
      return const Center(child: Text('No season data available.'));
    }

    final yValues = points.map((point) => point.y).toList(growable: false);
    final minY = yValues.reduce(math.min);
    final maxY = yValues.reduce(math.max);
    final spread = maxY - minY;
    final padding = spread == 0 ? 1.0 : spread * 0.15;

    return LineChart(
      LineChartData(
        minY: minY - padding,
        maxY: maxY + padding,
        lineTouchData: const LineTouchData(enabled: true),
        gridData: FlGridData(show: true, drawVerticalLine: false),
        titlesData: FlTitlesData(
          topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              interval: points.length > 12 ? 3 : 1,
              getTitlesWidget: (value, meta) {
                final index = value.toInt();
                if (index < 0 || index >= points.length) {
                  return const SizedBox.shrink();
                }
                return Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Text('${index + 1}', style: Theme.of(context).textTheme.bodySmall),
                );
              },
            ),
          ),
          leftTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 36,
              interval: spread == 0 ? 1 : spread / 3,
              getTitlesWidget: (value, _) =>
                  Text(value.toStringAsFixed(value.abs() < 10 ? 2 : 0), style: Theme.of(context).textTheme.bodySmall),
            ),
          ),
        ),
        borderData: FlBorderData(show: false),
        lineBarsData: <LineChartBarData>[
          LineChartBarData(
            spots: points,
            isCurved: true,
            barWidth: 2.5,
            dotData: const FlDotData(show: false),
            belowBarData: BarAreaData(show: true, color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.2)),
            color: Theme.of(context).colorScheme.primary,
          ),
        ],
      ),
    );
  }
}
