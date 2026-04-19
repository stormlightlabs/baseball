import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class LeadersScreen extends StatelessWidget {
  const LeadersScreen({super.key, required this.state});

  final MoreState state;

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<MoreCubit>();
    final leaders = state.leadersSnapshot?.entries ?? const <LeaderboardEntry>[];

    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
      children: <Widget>[
        Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text('Stat Leaders', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 10),
                SegmentedButton<LeadersMode>(
                  segments: const <ButtonSegment<LeadersMode>>[
                    ButtonSegment<LeadersMode>(value: LeadersMode.batting, label: Text('Batting')),
                    ButtonSegment<LeadersMode>(value: LeadersMode.pitching, label: Text('Pitching')),
                  ],
                  selected: <LeadersMode>{state.leadersMode},
                  onSelectionChanged: (selection) {
                    if (selection.isNotEmpty) {
                      cubit.setLeadersMode(selection.first);
                    }
                  },
                ),
                const SizedBox(height: 10),
                Row(
                  children: <Widget>[
                    ChoiceChip(
                      label: const Text('Season'),
                      selected: state.leadersScope == LeadersScope.season,
                      onSelected: (_) => cubit.setLeadersScope(LeadersScope.season),
                    ),
                    const SizedBox(width: 8),
                    ChoiceChip(
                      label: const Text('Career'),
                      selected: state.leadersScope == LeadersScope.career,
                      onSelected: (_) => cubit.setLeadersScope(LeadersScope.career),
                    ),
                    const Spacer(),
                    if (state.leadersScope == LeadersScope.season)
                      DropdownButton<int>(
                        value: state.selectedSeason > 0 ? state.selectedSeason : null,
                        hint: const Text('Season'),
                        items: state.availableSeasons
                            .take(25)
                            .map(
                              (season) =>
                                  DropdownMenuItem<int>(value: season.year, child: Text(season.year.toString())),
                            )
                            .toList(growable: false),
                        onChanged: (value) {
                          if (value != null) {
                            cubit.selectSeason(value);
                          }
                        },
                      ),
                  ],
                ),
                const SizedBox(height: 10),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: state.activeStatOptions
                        .map(
                          (option) => Padding(
                            padding: const EdgeInsets.only(right: 8),
                            child: ChoiceChip(
                              label: Text(option.label),
                              selected: state.leadersStat == option.key,
                              onSelected: (_) => cubit.setLeadersStat(option.key),
                            ),
                          ),
                        )
                        .toList(growable: false),
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 8),
        if (state.leadersStatus == MoreStatus.loading) const LoadingStrip(),
        if (state.leadersError != null && state.leadersStatus == MoreStatus.failure) ...<Widget>[
          const SizedBox(height: 8),
          InlineErrorText(state.leadersError!),
        ],
        const SizedBox(height: 10),
        _LeadersPodium(entries: leaders),
        const SizedBox(height: 10),
        PanelCard(
          title: '${state.leadersStat.toUpperCase()} leaderboard',
          child: leaders.isEmpty
              ? const Text('No leaders returned for this filter.')
              : Column(
                  children: leaders
                      .asMap()
                      .entries
                      .map(
                        (entry) => _LeaderRow(rank: entry.key + 1, row: entry.value, maxValue: leaders.first.rawValue),
                      )
                      .toList(growable: false),
                ),
        ),
      ],
    );
  }
}

class _LeadersPodium extends StatelessWidget {
  const _LeadersPodium({required this.entries});

  final List<LeaderboardEntry> entries;

  @override
  Widget build(BuildContext context) {
    if (entries.length < 3) {
      return const SizedBox.shrink();
    }

    final top = entries.take(3).toList(growable: false);

    return Card(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: <Widget>[
            Expanded(child: _PodiumTile(rank: 2, entry: top[1], height: 42)),
            const SizedBox(width: 8),
            Expanded(child: _PodiumTile(rank: 1, entry: top[0], height: 62)),
            const SizedBox(width: 8),
            Expanded(child: _PodiumTile(rank: 3, entry: top[2], height: 30)),
          ],
        ),
      ),
    );
  }
}

class _PodiumTile extends StatelessWidget {
  const _PodiumTile({required this.rank, required this.entry, required this.height});

  final int rank;
  final LeaderboardEntry entry;
  final double height;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Text(
          entry.playerName,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: Theme.of(context).textTheme.labelSmall,
        ),
        const SizedBox(height: 4),
        Container(
          height: height,
          decoration: BoxDecoration(
            color: colors.primary.withValues(alpha: rank == 1 ? 0.95 : 0.65),
            borderRadius: const BorderRadius.vertical(top: Radius.circular(6)),
          ),
        ),
        const SizedBox(height: 4),
        Text(entry.displayValue, style: Theme.of(context).textTheme.titleSmall),
        Text('#$rank', style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}

class _LeaderRow extends StatelessWidget {
  const _LeaderRow({required this.rank, required this.row, required this.maxValue});

  final int rank;
  final LeaderboardEntry row;
  final double maxValue;

  @override
  Widget build(BuildContext context) {
    final ratio = maxValue <= 0 ? 0.0 : (row.rawValue / maxValue).clamp(0.0, 1.0);

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: <Widget>[
          SizedBox(width: 22, child: Text(rank.toString(), textAlign: TextAlign.right)),
          const SizedBox(width: 8),
          CircleAvatar(radius: 14, child: Text(row.initials, style: Theme.of(context).textTheme.labelSmall)),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(row.playerName, maxLines: 1, overflow: TextOverflow.ellipsis),
                Text(
                  row.subtitle,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.labelSmall,
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 66,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: <Widget>[
                Text(row.displayValue, style: Theme.of(context).textTheme.titleSmall),
                const SizedBox(height: 2),
                LinearProgressIndicator(value: ratio),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
