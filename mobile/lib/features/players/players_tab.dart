import 'dart:math' as math;

import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/player_selection_cubit.dart';
import 'package:bigfly_mobile/features/players/players_cubit.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class PlayersTab extends StatefulWidget {
  const PlayersTab({super.key});

  @override
  State<PlayersTab> createState() => _PlayersTabState();
}

class _PlayersTabState extends State<PlayersTab> {
  late final TextEditingController _searchController;

  @override
  void initState() {
    super.initState();
    final cubit = context.read<PlayersCubit>();
    _searchController = TextEditingController(text: cubit.state.searchQuery);
    cubit.initialize(initialPlayerId: context.read<PlayerSelectionCubit>().state);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: <BlocListener<dynamic, dynamic>>[
        BlocListener<PlayerSelectionCubit, String?>(
          listener: (context, playerId) {
            if (playerId == null || playerId.isEmpty) {
              return;
            }
            final cubit = context.read<PlayersCubit>();
            if (cubit.state.selectedPlayerId == playerId && cubit.state.detailStatus == PlayerDetailStatus.ready) {
              return;
            }
            cubit.loadPlayer(playerId);
          },
        ),
        BlocListener<PlayersCubit, PlayersState>(
          listenWhen: (previous, next) => previous.detail?.themeTeamCode != next.detail?.themeTeamCode,
          listener: (context, state) {
            final themeCode = state.detail?.themeTeamCode;
            if (themeCode != null) {
              context.read<ThemeCubit>().selectTeam(themeCode);
            }
          },
        ),
      ],
      child: BlocBuilder<PlayersCubit, PlayersState>(
        builder: (context, state) {
          return ListView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 20),
            children: <Widget>[
              Text('Players', style: Theme.of(context).textTheme.headlineMedium),
              const SizedBox(height: 10),
              _buildSearchBar(context, state),
              if (state.searchError != null) ...<Widget>[
                const SizedBox(height: 8),
                Text(
                  state.searchError!,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Theme.of(context).colorScheme.error),
                ),
              ],
              if (state.searchLoading) ...<Widget>[
                const SizedBox(height: 8),
                const LinearProgressIndicator(minHeight: 2),
              ],
              if (state.searchResults.isNotEmpty) ...<Widget>[
                const SizedBox(height: 8),
                ...state.searchResults.map(
                  (item) => Card(
                    margin: const EdgeInsets.only(bottom: 8),
                    child: ListTile(
                      title: Text(item.name),
                      subtitle: Text(item.subtitle),
                      trailing: const Icon(Icons.chevron_right),
                      onTap: () => _onSelectPlayer(context, item.id),
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 8),
              _buildDetailContent(context, state),
              const SizedBox(height: 16),
              _buildRecentlyViewed(context, state),
            ],
          );
        },
      ),
    );
  }

  Widget _buildSearchBar(BuildContext context, PlayersState state) {
    return Row(
      children: <Widget>[
        Expanded(
          child: TextField(
            controller: _searchController,
            onChanged: context.read<PlayersCubit>().setSearchQuery,
            onSubmitted: (_) => context.read<PlayersCubit>().runSearch(),
            decoration: const InputDecoration(
              hintText: 'Name, player ID...',
              border: OutlineInputBorder(),
              isDense: true,
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(
          onPressed: state.searchLoading ? null : () => context.read<PlayersCubit>().runSearch(),
          child: const Icon(Icons.arrow_forward),
        ),
      ],
    );
  }

  Widget _buildDetailContent(BuildContext context, PlayersState state) => switch (state.detailStatus) {
    PlayerDetailStatus.initial || PlayerDetailStatus.loading => const Padding(
      padding: EdgeInsets.symmetric(vertical: 24),
      child: Center(child: CircularProgressIndicator()),
    ),
    PlayerDetailStatus.failure => Card(
      child: ListTile(
        leading: Icon(Icons.error_outline, color: Theme.of(context).colorScheme.error),
        title: const Text('Failed to load player'),
        subtitle: Text(state.detailError ?? 'Unknown error'),
      ),
    ),
    PlayerDetailStatus.ready when state.detail != null => _PlayerDetailView(detail: state.detail!, state: state),
    _ => const SizedBox.shrink(),
  };

  Widget _buildRecentlyViewed(BuildContext context, PlayersState state) {
    if (state.recentPlayers.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text('Recently viewed', style: Theme.of(context).extension<AppTypography>()?.code),
        const SizedBox(height: 8),
        ...state.recentPlayers.map(
          (player) => Card(
            margin: const EdgeInsets.only(bottom: 8),
            child: ListTile(
              leading: CircleAvatar(child: Text(player.name.isEmpty ? '?' : player.name[0].toUpperCase())),
              title: Text(player.name),
              subtitle: Text(player.subtitle),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => _onSelectPlayer(context, player.id),
            ),
          ),
        ),
      ],
    );
  }

  void _onSelectPlayer(BuildContext context, String playerId) {
    context.read<PlayerSelectionCubit>().selectPlayer(playerId);
    context.read<PlayersCubit>().loadPlayer(playerId);
  }
}

class _PlayerDetailView extends StatelessWidget {
  const _PlayerDetailView({required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    final player = detail.player;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Card(
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              children: <Widget>[
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    CircleAvatar(radius: 28, child: Text(player.initials)),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text(player.fullName, style: Theme.of(context).textTheme.titleLarge),
                          const SizedBox(height: 4),
                          Text(player.birthLine, style: Theme.of(context).textTheme.bodySmall),
                          const SizedBox(height: 2),
                          Text(
                            'Bats: ${player.bats ?? '—'} · Throws: ${player.throwsHand ?? '—'} · Debut: ${player.debut?.year ?? '—'}',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                _BioStatsGrid(detail: detail),
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 36,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: PlayerDetailTab.values
                .map(
                  (tab) => Padding(
                    padding: const EdgeInsets.only(right: 6),
                    child: ChoiceChip(
                      label: Text(_tabLabel(tab)),
                      selected: state.detailTab == tab,
                      onSelected: (_) => context.read<PlayersCubit>().setDetailTab(tab),
                    ),
                  ),
                )
                .toList(growable: false),
          ),
        ),
        const SizedBox(height: 12),
        switch (state.detailTab) {
          PlayerDetailTab.batting => _BattingSection(detail: detail, state: state),
          PlayerDetailTab.pitching => _PitchingSection(detail: detail, state: state),
          PlayerDetailTab.awards => _AwardsSection(awards: detail.awards),
          PlayerDetailTab.hallOfFame => _HallOfFameSection(records: detail.hallOfFameRecords),
        },
      ],
    );
  }

  String _tabLabel(PlayerDetailTab tab) => switch (tab) {
    PlayerDetailTab.batting => 'Batting',
    PlayerDetailTab.pitching => 'Pitching',
    PlayerDetailTab.awards => 'Awards',
    PlayerDetailTab.hallOfFame => 'HOF',
  };
}

class _BioStatsGrid extends StatelessWidget {
  const _BioStatsGrid({required this.detail});

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

class _BattingSection extends StatelessWidget {
  const _BattingSection({required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    final sortedSeasons = [...detail.battingSeasons]..sort((a, b) => a.year.compareTo(b.year));
    final tableRows = [...detail.battingSeasons];
    tableRows.sort((a, b) => _compareBatting(a, b, state.battingSortColumn));
    if (!state.battingSortAscending) {
      tableRows.setAll(0, tableRows.reversed);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _Panel(
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
                        label: Text(_battingMetricLabel(metric)),
                        selected: metric == state.battingMetric,
                        onSelected: (_) => context.read<PlayersCubit>().setBattingMetric(metric),
                      ),
                    )
                    .toList(growable: false),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 180,
                child: _LineChart(points: _battingMetricPoints(sortedSeasons, state.battingMetric)),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        _Panel(
          title: 'Season log — batting',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              sortColumnIndex: _battingSortColumnIndex(state.battingSortColumn),
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

  List<FlSpot> _battingMetricPoints(List<PlayerBattingSeason> seasons, BattingChartMetric metric) => seasons.isEmpty
      ? <FlSpot>[]
      : seasons
            .asMap()
            .entries
            .map((entry) => FlSpot(entry.key.toDouble(), _metricValue(entry.value, metric)))
            .toList(growable: false);

  double _metricValue(PlayerBattingSeason season, BattingChartMetric metric) => switch (metric) {
    BattingChartMetric.hr => season.hr.toDouble(),
    BattingChartMetric.avg => season.avg,
    BattingChartMetric.rbi => season.rbi.toDouble(),
    BattingChartMetric.obp => season.obp,
    BattingChartMetric.slg => season.slg,
    BattingChartMetric.ops => season.ops,
  };

  int _compareBatting(PlayerBattingSeason a, PlayerBattingSeason b, BattingSortColumn column) => switch (column) {
    BattingSortColumn.year => a.year.compareTo(b.year),
    BattingSortColumn.team => a.teamId.compareTo(b.teamId),
    BattingSortColumn.g => a.g.compareTo(b.g),
    BattingSortColumn.ab => a.ab.compareTo(b.ab),
    BattingSortColumn.avg => a.avg.compareTo(b.avg),
    BattingSortColumn.hr => a.hr.compareTo(b.hr),
    BattingSortColumn.rbi => a.rbi.compareTo(b.rbi),
    BattingSortColumn.ops => a.ops.compareTo(b.ops),
  };

  int _battingSortColumnIndex(BattingSortColumn column) => switch (column) {
    BattingSortColumn.year => 0,
    BattingSortColumn.team => 1,
    BattingSortColumn.g => 2,
    BattingSortColumn.ab => 3,
    BattingSortColumn.avg => 4,
    BattingSortColumn.hr => 5,
    BattingSortColumn.rbi => 6,
    BattingSortColumn.ops => 7,
  };

  String _battingMetricLabel(BattingChartMetric metric) => switch (metric) {
    BattingChartMetric.hr => 'HR',
    BattingChartMetric.avg => 'AVG',
    BattingChartMetric.rbi => 'RBI',
    BattingChartMetric.obp => 'OBP',
    BattingChartMetric.slg => 'SLG',
    BattingChartMetric.ops => 'OPS',
  };
}

class _PitchingSection extends StatelessWidget {
  const _PitchingSection({required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    if (detail.pitchingSeasons.isEmpty) {
      return const _Panel(title: 'Pitching', child: Text('No pitching seasons available.'));
    }

    final sortedSeasons = [...detail.pitchingSeasons]..sort((a, b) => a.year.compareTo(b.year));
    final tableRows = [...detail.pitchingSeasons];
    tableRows.sort((a, b) => _comparePitching(a, b, state.pitchingSortColumn));
    if (!state.pitchingSortAscending) {
      tableRows.setAll(0, tableRows.reversed);
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _Panel(
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
                        label: Text(_pitchingMetricLabel(metric)),
                        selected: metric == state.pitchingMetric,
                        onSelected: (_) => context.read<PlayersCubit>().setPitchingMetric(metric),
                      ),
                    )
                    .toList(growable: false),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 180,
                child: _LineChart(points: _pitchingMetricPoints(sortedSeasons, state.pitchingMetric)),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        _Panel(
          title: 'Season log — pitching',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              sortColumnIndex: _pitchingSortColumnIndex(state.pitchingSortColumn),
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

  List<FlSpot> _pitchingMetricPoints(List<PlayerPitchingSeason> seasons, PitchingChartMetric metric) => seasons.isEmpty
      ? <FlSpot>[]
      : seasons
            .asMap()
            .entries
            .map((entry) => FlSpot(entry.key.toDouble(), _metricValue(entry.value, metric)))
            .toList(growable: false);

  double _metricValue(PlayerPitchingSeason season, PitchingChartMetric metric) => switch (metric) {
    PitchingChartMetric.era => season.era,
    PitchingChartMetric.strikeouts => season.so.toDouble(),
    PitchingChartMetric.whip => season.whip,
    PitchingChartMetric.wins => season.wins.toDouble(),
    PitchingChartMetric.kPer9 => season.kPer9,
  };

  int _comparePitching(PlayerPitchingSeason a, PlayerPitchingSeason b, PitchingSortColumn column) => switch (column) {
    PitchingSortColumn.year => a.year.compareTo(b.year),
    PitchingSortColumn.team => a.teamId.compareTo(b.teamId),
    PitchingSortColumn.g => a.games.compareTo(b.games),
    PitchingSortColumn.wins => a.wins.compareTo(b.wins),
    PitchingSortColumn.losses => a.losses.compareTo(b.losses),
    PitchingSortColumn.era => a.era.compareTo(b.era),
    PitchingSortColumn.so => a.so.compareTo(b.so),
    PitchingSortColumn.whip => a.whip.compareTo(b.whip),
  };

  int _pitchingSortColumnIndex(PitchingSortColumn column) => switch (column) {
    PitchingSortColumn.year => 0,
    PitchingSortColumn.team => 1,
    PitchingSortColumn.g => 2,
    PitchingSortColumn.wins => 3,
    PitchingSortColumn.losses => 4,
    PitchingSortColumn.era => 5,
    PitchingSortColumn.so => 6,
    PitchingSortColumn.whip => 7,
  };

  String _pitchingMetricLabel(PitchingChartMetric metric) => switch (metric) {
    PitchingChartMetric.era => 'ERA',
    PitchingChartMetric.strikeouts => 'SO',
    PitchingChartMetric.whip => 'WHIP',
    PitchingChartMetric.wins => 'W',
    PitchingChartMetric.kPer9 => 'K/9',
  };
}

class _AwardsSection extends StatelessWidget {
  const _AwardsSection({required this.awards});

  final List<PlayerAward> awards;

  @override
  Widget build(BuildContext context) {
    if (awards.isEmpty) {
      return const _Panel(title: 'Awards & honors', child: Text('No awards found for this player.'));
    }

    final grouped = <String, List<int>>{};
    for (final award in awards) {
      grouped.putIfAbsent(award.awardId, () => <int>[]).add(award.year);
    }

    final entries = grouped.entries.toList(growable: false)..sort((a, b) => a.key.compareTo(b.key));

    return _Panel(
      title: 'Awards & honors',
      child: Column(
        children: entries
            .map((entry) {
              final years = [...entry.value]..sort();
              return ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const CircleAvatar(radius: 14, child: Icon(Icons.emoji_events_outlined, size: 14)),
                title: Text(entry.key),
                subtitle: Text(years.join(', ')),
              );
            })
            .toList(growable: false),
      ),
    );
  }
}

class _HallOfFameSection extends StatelessWidget {
  const _HallOfFameSection({required this.records});

  final List<PlayerHallOfFameRecord> records;

  @override
  Widget build(BuildContext context) {
    if (records.isEmpty) {
      return const _Panel(title: 'Hall of Fame', child: Text('No Hall of Fame records found.'));
    }

    final sorted = [...records]..sort((a, b) => a.year.compareTo(b.year));

    return _Panel(
      title: 'Hall of Fame',
      child: Column(
        children: sorted
            .map(
              (record) => ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Icon(
                  record.inducted ? Icons.verified : Icons.how_to_vote_outlined,
                  color: record.inducted ? Colors.amber : null,
                ),
                title: Text('${record.year} · ${record.votedBy}'),
                subtitle: Text(
                  record.votePercent == null
                      ? 'Votes: ${record.votes ?? '—'}'
                      : 'Votes: ${record.votes}/${record.ballots} (${record.votePercent!.toStringAsFixed(1)}%)',
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}

class _Panel extends StatelessWidget {
  const _Panel({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(title, style: Theme.of(context).extension<AppTypography>()?.code),
          const SizedBox(height: 10),
          child,
        ],
      ),
    ),
  );
}

class _LineChart extends StatelessWidget {
  const _LineChart({required this.points});

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
