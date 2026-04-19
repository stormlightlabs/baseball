import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/presentation/utils/player_detail_formatters.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_metric_line_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class PitchingSection extends StatelessWidget {
  const PitchingSection({super.key, required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    if (detail.pitchingSeasons.isEmpty) {
      return const PanelCard(title: 'Pitching', child: Text('No pitching seasons available.'));
    }

    final sortedSeasons = [...detail.pitchingSeasons]..sort((a, b) => a.year.compareTo(b.year));
    final tableRows = [...detail.pitchingSeasons];
    tableRows.sort((a, b) => comparePitchingSeasons(a, b, state.pitchingSortColumn));
    if (!state.pitchingSortAscending) {
      tableRows.setAll(0, tableRows.reversed);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        PanelCard(
          title: 'Career pitching — by season',
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: PitchingChartMetric.values
                    .map(
                      (metric) => ChoiceChip(
                        label: Text(pitchingMetricLabel(metric)),
                        selected: metric == state.pitchingMetric,
                        onSelected: (_) => context.read<PlayersCubit>().setPitchingMetric(metric),
                      ),
                    )
                    .toList(growable: false),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 180,
                child: PlayerMetricLineChart(points: pitchingMetricPoints(sortedSeasons, state.pitchingMetric)),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        PanelCard(
          title: 'Season log — pitching',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              sortColumnIndex: pitchingSortColumnIndex(state.pitchingSortColumn),
              sortAscending: state.pitchingSortAscending,
              columns: <DataColumn>[
                DataColumn(
                  label: const Text('Year'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.year, asc),
                ),
                DataColumn(
                  label: const Text('Team'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.team, asc),
                ),
                DataColumn(
                  label: const Text('G'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.g, asc),
                ),
                DataColumn(
                  label: const Text('W'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.wins, asc),
                ),
                DataColumn(
                  label: const Text('L'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.losses, asc),
                ),
                DataColumn(
                  label: const Text('ERA'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.era, asc),
                ),
                DataColumn(
                  label: const Text('SO'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.so, asc),
                ),
                DataColumn(
                  label: const Text('WHIP'),
                  onSort: (_, asc) => context.read<PlayersCubit>().setPitchingSort(PitchingSortColumn.whip, asc),
                ),
              ],
              rows: tableRows
                  .map(
                    (season) => DataRow(
                      cells: <DataCell>[
                        DataCell(Text(season.year.toString())),
                        DataCell(Text(season.teamId)),
                        DataCell(Text(season.games.toString())),
                        DataCell(Text(season.wins.toString())),
                        DataCell(Text(season.losses.toString())),
                        DataCell(Text(season.era.toStringAsFixed(2))),
                        DataCell(Text(season.so.toString())),
                        DataCell(Text(season.whip.toStringAsFixed(2))),
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
