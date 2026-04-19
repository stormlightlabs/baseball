import 'dart:math' as math;

import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_player_snapshot.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class CompareScreen extends StatelessWidget {
  const CompareScreen({super.key, required this.state});

  final MoreState state;

  @override
  Widget build(BuildContext context) {
    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
      children: <Widget>[
        _ComparePickerCard(
          title: 'Player A',
          loading: state.compareSearchLoadingA || state.compareLoadStatusA == MoreStatus.loading,
          error: state.compareSearchErrorA ?? state.compareErrorA,
          selected: state.comparePlayerA,
          results: state.compareSearchResultsA,
          onQueryChanged: context.read<MoreCubit>().searchComparePlayersA,
          onSelectPlayer: context.read<MoreCubit>().selectComparePlayerA,
          accentColor: Theme.of(context).colorScheme.primary,
        ),
        const SizedBox(height: 10),
        _ComparePickerCard(
          title: 'Player B',
          loading: state.compareSearchLoadingB || state.compareLoadStatusB == MoreStatus.loading,
          error: state.compareSearchErrorB ?? state.compareErrorB,
          selected: state.comparePlayerB,
          results: state.compareSearchResultsB,
          onQueryChanged: context.read<MoreCubit>().searchComparePlayersB,
          onSelectPlayer: context.read<MoreCubit>().selectComparePlayerB,
          accentColor: Colors.amber,
        ),
        const SizedBox(height: 10),
        if (state.hasCompareSelection) ...<Widget>[
          _CompareBioPanel(playerA: state.comparePlayerA!, playerB: state.comparePlayerB!),
          const SizedBox(height: 10),
          _CompareStatsPanel(playerA: state.comparePlayerA!, playerB: state.comparePlayerB!),
        ] else
          const Card(
            child: Padding(padding: EdgeInsets.all(12), child: Text('Select two players to compare.')),
          ),
      ],
    );
  }
}

class _ComparePickerCard extends StatelessWidget {
  const _ComparePickerCard({
    required this.title,
    required this.loading,
    required this.error,
    required this.selected,
    required this.results,
    required this.onQueryChanged,
    required this.onSelectPlayer,
    required this.accentColor,
  });

  final String title;
  final bool loading;
  final String? error;
  final ComparePlayerSnapshot? selected;
  final List<PlayerSearchResult> results;
  final ValueChanged<String> onQueryChanged;
  final ValueChanged<String> onSelectPlayer;
  final Color accentColor;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title, style: Theme.of(context).textTheme.titleMedium?.copyWith(color: accentColor)),
            const SizedBox(height: 8),
            TextField(
              decoration: const InputDecoration(prefixIcon: Icon(Icons.search), hintText: 'Search player ID or name'),
              onChanged: onQueryChanged,
            ),
            if (loading) ...<Widget>[const SizedBox(height: 8), const LoadingStrip(minHeight: 2)],
            if (error != null) ...<Widget>[const SizedBox(height: 8), InlineErrorText(error!)],
            if (results.isNotEmpty) ...<Widget>[
              const SizedBox(height: 8),
              ...results
                  .take(5)
                  .map(
                    (item) => ListTile(
                      dense: true,
                      contentPadding: EdgeInsets.zero,
                      title: Text(item.name, maxLines: 1, overflow: TextOverflow.ellipsis),
                      subtitle: Text(item.subtitle, maxLines: 1, overflow: TextOverflow.ellipsis),
                      onTap: () => onSelectPlayer(item.id),
                    ),
                  ),
            ],
            if (selected != null) ...<Widget>[
              const SizedBox(height: 8),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(8),
                  color: Theme.of(context).colorScheme.surfaceContainerLow,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(selected!.profile.fullName, style: Theme.of(context).textTheme.titleSmall),
                    Text(selected!.profile.subtitle, style: Theme.of(context).textTheme.labelSmall),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _CompareBioPanel extends StatelessWidget {
  const _CompareBioPanel({required this.playerA, required this.playerB});

  final ComparePlayerSnapshot playerA;
  final ComparePlayerSnapshot playerB;

  @override
  Widget build(BuildContext context) {
    return PanelCard(
      title: 'Bio overview',
      child: Row(
        children: <Widget>[
          Expanded(
            child: _CompareBioColumn(player: playerA, accent: Theme.of(context).colorScheme.primary),
          ),
          const SizedBox(width: 8),
          Text('VS', style: Theme.of(context).textTheme.labelMedium),
          const SizedBox(width: 8),
          Expanded(
            child: _CompareBioColumn(player: playerB, accent: Colors.amber),
          ),
        ],
      ),
    );
  }
}

class _CompareBioColumn extends StatelessWidget {
  const _CompareBioColumn({required this.player, required this.accent});

  final ComparePlayerSnapshot player;
  final Color accent;

  @override
  Widget build(BuildContext context) {
    final tenure = '${player.debutYear ?? '—'}–${player.finalYear ?? '—'}';
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: accent.withValues(alpha: 0.35)),
        color: Theme.of(context).colorScheme.surfaceContainerLow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            player.profile.fullName,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(color: accent),
          ),
          const SizedBox(height: 4),
          Text(player.profile.id, style: Theme.of(context).textTheme.labelSmall),
          Text(player.profile.positions ?? '—', style: Theme.of(context).textTheme.labelSmall),
          Text(tenure, style: Theme.of(context).textTheme.labelSmall),
        ],
      ),
    );
  }
}

class _CompareStatsPanel extends StatelessWidget {
  const _CompareStatsPanel({required this.playerA, required this.playerB});

  final ComparePlayerSnapshot playerA;
  final ComparePlayerSnapshot playerB;

  @override
  Widget build(BuildContext context) {
    final rows = <_CompareMetricRowData>[
      _CompareMetricRowData(
        label: 'HR',
        valueA: playerA.battingCareer.hr.toDouble(),
        valueB: playerB.battingCareer.hr.toDouble(),
        format: _intFormat,
      ),
      _CompareMetricRowData(
        label: 'AVG',
        valueA: playerA.battingCareer.avg,
        valueB: playerB.battingCareer.avg,
        format: _threeFormat,
      ),
      _CompareMetricRowData(
        label: 'OPS',
        valueA: playerA.battingCareer.ops,
        valueB: playerB.battingCareer.ops,
        format: _threeFormat,
      ),
      _CompareMetricRowData(
        label: 'RBI',
        valueA: playerA.battingCareer.rbi.toDouble(),
        valueB: playerB.battingCareer.rbi.toDouble(),
        format: _intFormat,
      ),
      _CompareMetricRowData(
        label: 'SB',
        valueA: playerA.battingCareer.sb.toDouble(),
        valueB: playerB.battingCareer.sb.toDouble(),
        format: _intFormat,
      ),
      _CompareMetricRowData(
        label: 'W',
        valueA: playerA.pitchingCareer.wins.toDouble(),
        valueB: playerB.pitchingCareer.wins.toDouble(),
        format: _intFormat,
      ),
      _CompareMetricRowData(
        label: 'ERA',
        valueA: playerA.pitchingCareer.era,
        valueB: playerB.pitchingCareer.era,
        format: _twoFormat,
        lowerIsBetter: true,
      ),
      _CompareMetricRowData(
        label: 'SO',
        valueA: playerA.pitchingCareer.strikeouts.toDouble(),
        valueB: playerB.pitchingCareer.strikeouts.toDouble(),
        format: _intFormat,
      ),
      _CompareMetricRowData(
        label: 'WHIP',
        valueA: playerA.pitchingCareer.whip,
        valueB: playerB.pitchingCareer.whip,
        format: _twoFormat,
        lowerIsBetter: true,
      ),
      _CompareMetricRowData(
        label: 'K/9',
        valueA: playerA.pitchingCareer.kPer9,
        valueB: playerB.pitchingCareer.kPer9,
        format: _twoFormat,
      ),
    ];

    return PanelCard(
      title: 'Stat comparison',
      child: Column(children: rows.map((row) => _CompareMetricRow(data: row)).toList(growable: false)),
    );
  }

  static String _intFormat(double value) => value.round().toString();

  static String _twoFormat(double value) => value.toStringAsFixed(2);

  static String _threeFormat(double value) => value.toStringAsFixed(3).replaceFirst('0.', '.');
}

class _CompareMetricRowData {
  const _CompareMetricRowData({
    required this.label,
    required this.valueA,
    required this.valueB,
    required this.format,
    this.lowerIsBetter = false,
  });

  final String label;
  final double valueA;
  final double valueB;
  final String Function(double value) format;
  final bool lowerIsBetter;
}

class _CompareMetricRow extends StatelessWidget {
  const _CompareMetricRow({required this.data});

  final _CompareMetricRowData data;

  @override
  Widget build(BuildContext context) {
    final values = _normalizedValues(data.valueA, data.valueB, data.lowerIsBetter);

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Center(child: Text(data.label, style: Theme.of(context).textTheme.labelSmall)),
          const SizedBox(height: 4),
          Row(
            children: <Widget>[
              SizedBox(
                width: 56,
                child: Text(
                  data.format(data.valueA),
                  textAlign: TextAlign.right,
                  style: TextStyle(color: Theme.of(context).colorScheme.primary),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      flex: math.max(1, (values.$1 * 100).round()),
                      child: Container(height: 8, color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.65)),
                    ),
                    const SizedBox(width: 2),
                    Expanded(
                      flex: math.max(1, (values.$2 * 100).round()),
                      child: Container(height: 8, color: Colors.amber.withValues(alpha: 0.65)),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              SizedBox(
                width: 56,
                child: Text(data.format(data.valueB), style: const TextStyle(color: Colors.amber)),
              ),
            ],
          ),
        ],
      ),
    );
  }

  (double, double) _normalizedValues(double a, double b, bool lowerIsBetter) {
    if (a == 0 && b == 0) {
      return (0.5, 0.5);
    }

    double scoreA = a;
    double scoreB = b;

    if (lowerIsBetter) {
      final maxValue = math.max(a, b);
      scoreA = maxValue - a;
      scoreB = maxValue - b;
      if (scoreA == 0 && scoreB == 0) {
        scoreA = 1;
        scoreB = 1;
      }
    }

    final total = scoreA + scoreB;
    if (total <= 0) {
      return (0.5, 0.5);
    }

    return ((scoreA / total).clamp(0.0, 1.0), (scoreB / total).clamp(0.0, 1.0));
  }
}
