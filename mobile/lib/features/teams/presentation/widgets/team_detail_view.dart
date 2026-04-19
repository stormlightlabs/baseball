import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/colors.dart';
import 'package:bigfly_mobile/features/teams/application/teams_types.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:bigfly_mobile/features/teams/presentation/utils/team_detail_formatters.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_run_differential_chart.dart';
import 'package:flutter/material.dart';

class TeamDetailView extends StatelessWidget {
  const TeamDetailView({
    super.key,
    required this.detail,
    required this.activeSegment,
    required this.franchises,
    required this.onSegmentSelected,
    required this.onSelectFranchise,
  });

  final TeamDetailBundle detail;
  final TeamDetailSegment activeSegment;
  final List<FranchiseSummary> franchises;
  final ValueChanged<TeamDetailSegment> onSegmentSelected;
  final ValueChanged<String> onSelectFranchise;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _TeamFranchiseCard(detail: detail),
        const SizedBox(height: 12),
        _TeamSegmentControl(activeSegment: activeSegment, onSegmentSelected: onSegmentSelected),
        const SizedBox(height: 12),
        switch (activeSegment) {
          TeamDetailSegment.overview => _TeamOverview(detail: detail),
          TeamDetailSegment.roster => _TeamRosterPanel(detail: detail),
          TeamDetailSegment.schedule => _TeamSchedulePanel(detail: detail),
          TeamDetailSegment.daily => _TeamDailyPanel(detail: detail),
        },
        const SizedBox(height: 12),
        _OtherFranchisesPanel(detail: detail, franchises: franchises, onSelectFranchise: onSelectFranchise),
      ],
    );
  }
}

class _TeamFranchiseCard extends StatelessWidget {
  const _TeamFranchiseCard({required this.detail});

  final TeamDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    final team = detail.team;
    final franchise = detail.franchise;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          children: <Widget>[
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                CircleAvatar(radius: 26, child: Text((franchise?.id ?? team.franchiseId).substring(0, 2))),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(team.name, style: Theme.of(context).textTheme.titleLarge),
                      const SizedBox(height: 4),
                      Text(
                        '${team.teamId} · ${team.league}${team.division != null ? ' ${team.division}' : ''}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        'Active: ${franchise?.activeRange ?? '—'} · Park: ${team.parkId}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: <Widget>[
                _StatCell(label: 'W', value: team.wins.toString()),
                _StatCell(label: 'L', value: team.losses.toString()),
                _StatCell(label: 'W%', value: formatWinPct(team.winPct)),
                _StatCell(label: 'RS', value: team.runsScored.toString()),
                _StatCell(label: 'RA', value: team.runsAllowed.toString()),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _StatCell extends StatelessWidget {
  const _StatCell({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: Border.all(color: Theme.of(context).colorScheme.outlineVariant),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Column(
            children: <Widget>[
              Text(value, style: Theme.of(context).textTheme.titleSmall),
              const SizedBox(height: 2),
              Text(label, style: Theme.of(context).textTheme.labelSmall),
            ],
          ),
        ),
      ),
    );
  }
}

class _TeamSegmentControl extends StatelessWidget {
  const _TeamSegmentControl({required this.activeSegment, required this.onSegmentSelected});

  final TeamDetailSegment activeSegment;
  final ValueChanged<TeamDetailSegment> onSegmentSelected;

  @override
  Widget build(BuildContext context) {
    return SegmentedButton<TeamDetailSegment>(
      showSelectedIcon: false,
      segments: TeamDetailSegment.values
          .map(
            (segment) => ButtonSegment<TeamDetailSegment>(value: segment, label: Text(teamDetailSegmentLabel(segment))),
          )
          .toList(growable: false),
      selected: <TeamDetailSegment>{activeSegment},
      onSelectionChanged: (selected) {
        if (selected.isEmpty) {
          return;
        }
        onSegmentSelected(selected.first);
      },
    );
  }
}

class _TeamOverview extends StatelessWidget {
  const _TeamOverview({required this.detail});

  final TeamDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        PanelCard(
          title: '${detail.team.year} season · run differential by month',
          child: TeamRunDifferentialChart(series: detail.runDifferential),
        ),
        const SizedBox(height: 12),
        PanelCard(
          title: 'Team seasons · recent',
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: DataTable(
              headingTextStyle: Theme.of(context).textTheme.labelSmall,
              dataTextStyle: Theme.of(context).textTheme.bodySmall,
              columns: const <DataColumn>[
                DataColumn(label: Text('Year')),
                DataColumn(label: Text('W')),
                DataColumn(label: Text('L')),
                DataColumn(label: Text('W%')),
                DataColumn(label: Text('RS')),
                DataColumn(label: Text('RA')),
                DataColumn(label: Text('Finish')),
              ],
              rows: detail.recentSeasons
                  .map(
                    (season) => DataRow(
                      cells: <DataCell>[
                        DataCell(Text(season.year.toString())),
                        DataCell(Text(season.wins.toString())),
                        DataCell(Text(season.losses.toString())),
                        DataCell(Text(formatWinPct(season.winPct))),
                        DataCell(Text(season.runsScored.toString())),
                        DataCell(Text(season.runsAllowed.toString())),
                        DataCell(Text('${season.league} ${season.division ?? '—'}')),
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

class _TeamRosterPanel extends StatelessWidget {
  const _TeamRosterPanel({required this.detail});

  final TeamDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    return PanelCard(
      title: 'Current roster · ${detail.team.year}',
      child: Column(
        children: detail.roster
            .take(18)
            .map(
              (player) => ListTile(
                contentPadding: EdgeInsets.zero,
                leading: CircleAvatar(radius: 16, child: Text(player.initials)),
                title: Text(player.fullName),
                subtitle: Text(player.statLine),
                trailing: player.position == null
                    ? null
                    : Chip(
                        label: Text(player.position!, style: Theme.of(context).textTheme.labelSmall),
                        visualDensity: VisualDensity.compact,
                      ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}

class _TeamSchedulePanel extends StatelessWidget {
  const _TeamSchedulePanel({required this.detail});

  final TeamDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    final teamId = detail.team.teamId;

    return PanelCard(
      title: 'Recent games',
      child: Column(
        children: detail.schedule
            .take(15)
            .map((game) {
              final result = game.hasDecisionFor(teamId) ? (game.didTeamWin(teamId) ? 'W' : 'L') : '—';
              final resultColor = switch (result) {
                'W' => const Color(0xFF10B981),
                'L' => const Color(0xFFEF4444),
                _ => Theme.of(context).colorScheme.onSurfaceVariant,
              };

              final teamScore = game.isTeamHome(teamId) ? game.homeScore : game.awayScore;
              final oppScore = game.isTeamHome(teamId) ? game.awayScore : game.homeScore;

              return ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text('${game.awayTeam} @ ${game.homeTeam}'),
                subtitle: Text(game.parkName ?? '—'),
                leading: Text(formatCompactDate(game.date), style: Theme.of(context).textTheme.labelSmall),
                trailing: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: <Widget>[
                    Text('$teamScore-$oppScore', style: Theme.of(context).textTheme.titleSmall),
                    Text(result, style: Theme.of(context).textTheme.labelSmall?.copyWith(color: resultColor)),
                  ],
                ),
              );
            })
            .toList(growable: false),
      ),
    );
  }
}

class _TeamDailyPanel extends StatelessWidget {
  const _TeamDailyPanel({required this.detail});

  final TeamDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    return PanelCard(
      title: 'Daily team logs',
      child: Column(
        children: detail.dailyLogs
            .take(20)
            .map((log) {
              final diffText = log.runDiff >= 0 ? '+${log.runDiff}' : '${log.runDiff}';
              final diffColor = log.runDiff >= 0 ? const Color(0xFF10B981) : const Color(0xFFEF4444);

              return ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Text(formatIsoDate(log.date), style: Theme.of(context).textTheme.labelSmall),
                title: Text('W-L ${log.wins}-${log.losses} · G ${log.gamesPlayed}'),
                subtitle: Text('RS ${log.runsScored} · RA ${log.runsAllowed}'),
                trailing: Text(diffText, style: Theme.of(context).textTheme.titleSmall?.copyWith(color: diffColor)),
              );
            })
            .toList(growable: false),
      ),
    );
  }
}

class _OtherFranchisesPanel extends StatelessWidget {
  const _OtherFranchisesPanel({required this.detail, required this.franchises, required this.onSelectFranchise});

  final TeamDetailBundle detail;
  final List<FranchiseSummary> franchises;
  final ValueChanged<String> onSelectFranchise;

  @override
  Widget build(BuildContext context) {
    final others = franchises
        .where((franchise) => franchise.id != detail.team.franchiseId)
        .take(8)
        .toList(growable: false);

    if (others.isEmpty) {
      return const SizedBox.shrink();
    }

    return PanelCard(
      title: 'Other franchises',
      child: Column(
        children: others
            .map(
              (franchise) => ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: teamPrimaryColor(franchise.id) ?? Theme.of(context).colorScheme.primary,
                    borderRadius: BorderRadius.circular(99),
                  ),
                ),
                title: Text(franchise.name),
                subtitle: Text(franchise.activeRange),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => onSelectFranchise(franchise.id),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}
