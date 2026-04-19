import 'dart:math' as math;

import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/more/application/types/season_league_filters.dart';
import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';
import 'package:bigfly_mobile/features/more/data/models/postseason_series_record.dart';
import 'package:bigfly_mobile/features/more/data/models/season_award_item.dart';
import 'package:bigfly_mobile/features/more/data/models/season_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_summary_record.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class SeasonsScreen extends StatelessWidget {
  const SeasonsScreen({super.key, required this.state});

  final MoreState state;

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<MoreCubit>();

    if (state.seasonStatus == MoreStatus.loading && state.seasonSnapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: const <Widget>[LoadingStrip()],
      );
    }

    if (state.seasonStatus == MoreStatus.failure && state.seasonSnapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: <Widget>[InlineErrorText(state.seasonError ?? 'Failed to load seasons data')],
      );
    }

    final snapshot = state.seasonSnapshot;
    if (snapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: const <Widget>[
          Card(
            child: Padding(padding: EdgeInsets.all(14), child: Text('No season selected.')),
          ),
        ],
      );
    }

    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
      children: <Widget>[
        _SeasonToolbar(state: state),
        if (state.seasonStatus == MoreStatus.loading) ...<Widget>[const SizedBox(height: 10), const LoadingStrip()],
        if (state.seasonError != null && state.seasonStatus == MoreStatus.failure) ...<Widget>[
          const SizedBox(height: 10),
          InlineErrorText(state.seasonError!),
        ],
        const SizedBox(height: 12),
        _SeasonSummaryStrip(snapshot: snapshot),
        const SizedBox(height: 12),
        ..._buildStandingsPanels(context, snapshot.teams),
        const SizedBox(height: 12),
        _LeadersMiniPanel(title: 'Home run leaders', entries: snapshot.hrLeaders),
        const SizedBox(height: 10),
        _LeadersMiniPanel(title: 'Batting average leaders', entries: snapshot.avgLeaders),
        const SizedBox(height: 10),
        _AwardsPanel(awards: snapshot.awards),
        const SizedBox(height: 10),
        _PostseasonPanel(series: snapshot.postseasonSeries),
        const SizedBox(height: 10),
        if (state.availableSeasons.isNotEmpty)
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Text(
                'Loaded ${state.availableSeasons.length} seasons from /api/v1/seasons',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
          ),
        const SizedBox(height: 8),
        FilledButton.tonal(
          onPressed: () => cubit.setSection(MoreSection.leaders),
          child: const Text('Open full leaders view'),
        ),
      ],
    );
  }

  List<Widget> _buildStandingsPanels(BuildContext context, List<TeamSeasonRecord> teams) {
    if (teams.isEmpty) {
      return <Widget>[
        const Card(
          child: Padding(padding: EdgeInsets.all(12), child: Text('No standings returned for this season.')),
        ),
      ];
    }

    final leagueGroups = <String, List<TeamSeasonRecord>>{};
    for (final team in teams) {
      leagueGroups.putIfAbsent(team.league, () => <TeamSeasonRecord>[]).add(team);
    }

    final widgets = <Widget>[];
    final sortedLeagues = leagueGroups.keys.toList(growable: false)..sort();

    for (final league in sortedLeagues) {
      final leagueTeams = leagueGroups[league]!
        ..sort((a, b) {
          final divisionCmp = (a.division ?? '').compareTo(b.division ?? '');
          if (divisionCmp != 0) {
            return divisionCmp;
          }
          return b.wins.compareTo(a.wins);
        });

      widgets.add(
        PanelCard(
          title: '$league standings',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              columnSpacing: 12,
              headingRowHeight: 28,
              dataRowMinHeight: 30,
              dataRowMaxHeight: 36,
              columns: const <DataColumn>[
                DataColumn(label: Text('Team')),
                DataColumn(label: Text('W')),
                DataColumn(label: Text('L')),
                DataColumn(label: Text('GB')),
                DataColumn(label: Text('RS')),
                DataColumn(label: Text('RA')),
              ],
              rows: _buildStandingRows(context, leagueTeams),
            ),
          ),
        ),
      );
      widgets.add(const SizedBox(height: 10));
    }

    return widgets;
  }

  List<DataRow> _buildStandingRows(BuildContext context, List<TeamSeasonRecord> teams) {
    final rows = <DataRow>[];

    final divisionGroups = <String, List<TeamSeasonRecord>>{};
    for (final team in teams) {
      final division = team.division ?? '—';
      divisionGroups.putIfAbsent(division, () => <TeamSeasonRecord>[]).add(team);
    }

    final divisions = divisionGroups.keys.toList(growable: false)..sort();
    for (final division in divisions) {
      final divisionTeams = divisionGroups[division]!;
      divisionTeams.sort((a, b) => b.wins.compareTo(a.wins));
      final leader = divisionTeams.first;

      rows.add(
        DataRow(
          cells: <DataCell>[
            DataCell(Text(division, style: Theme.of(context).textTheme.labelMedium)),
            const DataCell(Text('')),
            const DataCell(Text('')),
            const DataCell(Text('')),
            const DataCell(Text('')),
            const DataCell(Text('')),
          ],
        ),
      );

      for (final team in divisionTeams) {
        rows.add(
          DataRow(
            cells: <DataCell>[
              DataCell(Text(team.name, maxLines: 1, overflow: TextOverflow.ellipsis)),
              DataCell(Text(team.wins.toString())),
              DataCell(Text(team.losses.toString())),
              DataCell(Text(_gamesBack(team: team, leader: leader))),
              DataCell(Text(team.runsScored.toString())),
              DataCell(Text(team.runsAllowed.toString())),
            ],
          ),
        );
      }
    }

    return rows;
  }

  String _gamesBack({required TeamSeasonRecord team, required TeamSeasonRecord leader}) {
    if (team.teamId == leader.teamId) {
      return '—';
    }

    final gb = ((leader.wins - team.wins) + (team.losses - leader.losses)) / 2;
    if ((gb - gb.round()).abs() < 0.001) {
      return gb.round().toString();
    }
    return gb.toStringAsFixed(1);
  }
}

class _SeasonToolbar extends StatelessWidget {
  const _SeasonToolbar({required this.state});

  final MoreState state;

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<MoreCubit>();

    final seasonYears = state.availableSeasons.map((season) => season.year).toList(growable: false);
    final currentIndex = seasonYears.indexOf(state.selectedSeason);

    final canMoveBack = currentIndex >= 0 && currentIndex < seasonYears.length - 1;
    final canMoveForward = currentIndex > 0;

    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        child: Column(
          children: <Widget>[
            Row(
              children: <Widget>[
                IconButton(
                  onPressed: canMoveBack ? () => cubit.selectSeason(seasonYears[currentIndex + 1]) : null,
                  icon: const Icon(Icons.chevron_left),
                  tooltip: 'Previous season',
                ),
                Expanded(
                  child: TextButton(
                    onPressed: () =>
                        _showSeasonPicker(context, state.availableSeasons, state.selectedSeason, cubit.selectSeason),
                    child: Text(state.selectedSeason > 0 ? state.selectedSeason.toString() : 'Choose season'),
                  ),
                ),
                IconButton(
                  onPressed: canMoveForward ? () => cubit.selectSeason(seasonYears[currentIndex - 1]) : null,
                  icon: const Icon(Icons.chevron_right),
                  tooltip: 'Next season',
                ),
              ],
            ),
            const SizedBox(height: 6),
            Wrap(
              spacing: 8,
              children: seasonLeagueFilters
                  .map(
                    (filter) => ChoiceChip(
                      label: Text(filter),
                      selected: state.seasonLeagueFilter == filter,
                      onSelected: (_) => cubit.setSeasonLeagueFilter(filter),
                    ),
                  )
                  .toList(growable: false),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _showSeasonPicker(
    BuildContext context,
    List<SeasonSummaryRecord> seasons,
    int selectedSeason,
    Future<void> Function(int season) onSelect,
  ) async {
    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (context) {
        return ListView.builder(
          itemCount: seasons.length,
          itemBuilder: (context, index) {
            final season = seasons[index];
            return ListTile(
              title: Text(season.year.toString()),
              subtitle: Text('${season.teamCount} teams · ${season.leagues.join('/')}'),
              trailing: season.year == selectedSeason ? const Icon(Icons.check) : null,
              onTap: () {
                Navigator.of(context).pop();
                onSelect(season.year);
              },
            );
          },
        );
      },
    );
  }
}

class _SeasonSummaryStrip extends StatelessWidget {
  const _SeasonSummaryStrip({required this.snapshot});

  final SeasonSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(
          child: _SummaryTile(value: snapshot.teams.length.toString(), label: 'Teams'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(value: snapshot.avgGamesPerTeam.toString(), label: 'G / team'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(value: snapshot.totalHomeRuns.toString(), label: 'Total HR'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(
            value: snapshot.leagueAverage == 0
                ? '—'
                : snapshot.leagueAverage.toStringAsFixed(3).replaceFirst('0.', '.'),
            label: 'Lg AVG',
          ),
        ),
      ],
    );
  }
}

class _SummaryTile extends StatelessWidget {
  const _SummaryTile({required this.value, required this.label});

  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 8),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(10),
        color: Theme.of(context).colorScheme.surfaceContainerLow,
        border: Border.all(color: Theme.of(context).colorScheme.outlineVariant.withValues(alpha: 0.6)),
      ),
      child: Column(
        children: <Widget>[
          Text(value, style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 2),
          Text(label, style: Theme.of(context).textTheme.labelSmall),
        ],
      ),
    );
  }
}

class _LeadersMiniPanel extends StatelessWidget {
  const _LeadersMiniPanel({required this.title, required this.entries});

  final String title;
  final List<LeaderboardEntry> entries;

  @override
  Widget build(BuildContext context) {
    return PanelCard(
      title: title,
      child: entries.isEmpty
          ? const Text('No data available')
          : Column(
              children: entries
                  .take(3)
                  .toList(growable: false)
                  .asMap()
                  .entries
                  .map(
                    (entry) => Padding(
                      padding: EdgeInsets.only(bottom: entry.key == math.min(2, entries.length - 1) ? 0 : 8),
                      child: Row(
                        children: <Widget>[
                          Text('#${entry.key + 1}', style: Theme.of(context).textTheme.labelMedium),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: <Widget>[
                                Text(entry.value.playerName, maxLines: 1, overflow: TextOverflow.ellipsis),
                                Text(entry.value.subtitle, style: Theme.of(context).textTheme.labelSmall),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          Text(entry.value.displayValue, style: Theme.of(context).textTheme.titleSmall),
                        ],
                      ),
                    ),
                  )
                  .toList(growable: false),
            ),
    );
  }
}

class _AwardsPanel extends StatelessWidget {
  const _AwardsPanel({required this.awards});

  final List<SeasonAwardItem> awards;

  @override
  Widget build(BuildContext context) {
    if (awards.isEmpty) {
      return const PanelCard(
        title: 'Notable events',
        child: Text('No awards or notable events returned for this season.'),
      );
    }

    return PanelCard(
      title: 'Notable events',
      child: Column(
        children: awards
            .take(5)
            .map(
              (item) => Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(8),
                    color: Theme.of(context).colorScheme.surfaceContainer,
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(item.awardId, style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(height: 2),
                      Text(
                        '${item.playerId}${item.league == null ? '' : ' · ${item.league}'}',
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                    ],
                  ),
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}

class _PostseasonPanel extends StatelessWidget {
  const _PostseasonPanel({required this.series});

  final List<PostseasonSeriesRecord> series;

  @override
  Widget build(BuildContext context) {
    return PanelCard(
      title: 'Postseason results',
      child: series.isEmpty
          ? const Text('No postseason series returned for this season.')
          : Column(
              children: series
                  .map(
                    (row) => Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Row(
                        children: <Widget>[
                          SizedBox(width: 52, child: Text(row.round, style: Theme.of(context).textTheme.labelSmall)),
                          Expanded(child: Text(row.matchup, maxLines: 1, overflow: TextOverflow.ellipsis)),
                          const SizedBox(width: 8),
                          Text(row.result, style: Theme.of(context).extension<AppTypography>()?.code),
                        ],
                      ),
                    ),
                  )
                  .toList(growable: false),
            ),
    );
  }
}
