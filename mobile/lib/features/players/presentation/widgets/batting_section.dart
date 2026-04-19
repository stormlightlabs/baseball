import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/presentation/utils/player_detail_formatters.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_metric_line_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class BattingSection extends StatelessWidget {
  const BattingSection({super.key, required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    final sortedSeasons = [...detail.battingSeasons]..sort((a, b) => a.year.compareTo(b.year));
    final tableRows = [...detail.battingSeasons];
    tableRows.sort((a, b) => compareBattingSeasons(a, b, state.battingSortColumn));
    if (!state.battingSortAscending) {
      tableRows.setAll(0, tableRows.reversed);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        PanelCard(
          title: 'Career batting — by season',
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: BattingChartMetric.values
                    .map(
                      (metric) => ChoiceChip(
                        label: Text(battingMetricLabel(metric)),
                        selected: metric == state.battingMetric,
                        onSelected: (_) => context.read<PlayersCubit>().setBattingMetric(metric),
                      ),
                    )
                    .toList(growable: false),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 180,
                child: PlayerMetricLineChart(points: battingMetricPoints(sortedSeasons, state.battingMetric)),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        PanelCard(
          title: 'Season log — batting',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              sortColumnIndex: battingSortColumnIndex(state.battingSortColumn),
              sortAscending: state.battingSortAscending,
              columns: <DataColumn>[
                DataColumn(
                  label: const Text('Year'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.year, asc),
                ),
                DataColumn(
                  label: const Text('Team'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.team, asc),
                ),
                DataColumn(
                  label: const Text('G'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.g, asc),
                ),
                DataColumn(
                  label: const Text('AB'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.ab, asc),
                ),
                DataColumn(
                  label: const Text('AVG'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.avg, asc),
                ),
                DataColumn(
                  label: const Text('HR'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.hr, asc),
                ),
                DataColumn(
                  label: const Text('RBI'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.rbi, asc),
                ),
                DataColumn(
                  label: const Text('OPS'),
                  numeric: true,
                  onSort: (_, asc) => context.read<PlayersCubit>().setBattingSort(BattingSortColumn.ops, asc),
                ),
              ],
              rows: tableRows
                  .map(
                    (season) => DataRow(
                      cells: <DataCell>[
                        DataCell(Text(season.year.toString())),
                        DataCell(Text(season.teamId)),
                        DataCell(Text(season.g.toString())),
                        DataCell(Text(season.ab.toString())),
                        DataCell(Text(season.avg.toStringAsFixed(3))),
                        DataCell(Text(season.hr.toString())),
                        DataCell(Text(season.rbi.toString())),
                        DataCell(Text(season.ops.toStringAsFixed(3))),
                      ],
                    ),
                  )
                  .toList(growable: false),
            ),
          ),
        ),
      ],
    );
  }
}
