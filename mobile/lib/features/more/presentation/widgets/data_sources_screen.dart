import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/more/data/models/data_sources_snapshot.dart';
import 'package:flutter/material.dart';

class DataSourcesScreen extends StatelessWidget {
  const DataSourcesScreen({super.key, required this.state});

  final MoreState state;

  @override
  Widget build(BuildContext context) {
    if (state.dataSourcesStatus == MoreStatus.loading && state.dataSourcesSnapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: const <Widget>[LoadingStrip()],
      );
    }

    if (state.dataSourcesStatus == MoreStatus.failure && state.dataSourcesSnapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: <Widget>[InlineErrorText(state.dataSourcesError ?? 'Failed to load data sources')],
      );
    }

    final snapshot = state.dataSourcesSnapshot;
    if (snapshot == null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        children: const <Widget>[
          Card(
            child: Padding(padding: EdgeInsets.all(12), child: Text('No data source metadata available.')),
          ),
        ],
      );
    }

    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 20),
      children: <Widget>[
        if (state.dataSourcesStatus == MoreStatus.loading) const LoadingStrip(),
        if (state.dataSourcesError != null && state.dataSourcesStatus == MoreStatus.failure) ...<Widget>[
          const SizedBox(height: 8),
          InlineErrorText(state.dataSourcesError!),
        ],
        const SizedBox(height: 8),
        _DataSourceSummaryStrip(snapshot: snapshot),
        const SizedBox(height: 10),
        ...snapshot.datasets.map(
          (dataset) => Padding(
            padding: const EdgeInsets.only(bottom: 10),
            child: _DatasetCard(dataset: dataset),
          ),
        ),
        _CoverageTimeline(snapshot: snapshot),
        const SizedBox(height: 10),
        const PanelCard(
          title: 'Data caveats',
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text('Federal League (1914–1915) and Negro Leagues seasons have partial game-level coverage.'),
              SizedBox(height: 8),
              Text('Statcast-era context metrics are only available for modern seasons.'),
              SizedBox(height: 8),
              Text('Some legacy salary and transaction tables are season-range limited.'),
            ],
          ),
        ),
      ],
    );
  }
}

class _DataSourceSummaryStrip extends StatelessWidget {
  const _DataSourceSummaryStrip({required this.snapshot});

  final DataSourcesSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    final minYear = snapshot.minCoverage?.toString() ?? '—';
    final maxYear = snapshot.maxCoverage?.toString() ?? '—';
    final spanLabel = snapshot.minCoverage != null && snapshot.maxCoverage != null
        ? '${snapshot.maxCoverage! - snapshot.minCoverage! + 1}yr'
        : '—';

    return Row(
      children: <Widget>[
        Expanded(
          child: _SummaryTile(value: snapshot.datasets.length.toString(), label: 'Sources'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(value: spanLabel, label: 'Coverage'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(value: _compactCount(snapshot.totalRows), label: 'Rows'),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _SummaryTile(value: '$minYear–$maxYear', label: 'Range'),
        ),
      ],
    );
  }

  String _compactCount(int value) {
    if (value >= 1000000) {
      return '${(value / 1000000).toStringAsFixed(1)}M';
    }
    if (value >= 1000) {
      return '${(value / 1000).toStringAsFixed(1)}K';
    }
    return value.toString();
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

class _DatasetCard extends StatelessWidget {
  const _DatasetCard({required this.dataset});

  final DatasetStatusSnapshot dataset;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Chip(label: Text(dataset.name)),
                const SizedBox(width: 8),
                Text(dataset.source, style: Theme.of(context).textTheme.labelSmall),
                const Spacer(),
                Icon(
                  dataset.healthy ? Icons.check_circle_outline : Icons.error_outline,
                  color: dataset.healthy ? Colors.green : Theme.of(context).colorScheme.error,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: <Widget>[
                Expanded(
                  child: _DatasetMetaTile(
                    label: 'Coverage',
                    value: '${dataset.coverageFrom ?? '—'}–${dataset.coverageTo ?? '—'}',
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _DatasetMetaTile(label: 'Rows', value: dataset.rowCount.toString()),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _DatasetMetaTile extends StatelessWidget {
  const _DatasetMetaTile({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        color: Theme.of(context).colorScheme.surfaceContainerLow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(value, style: Theme.of(context).textTheme.titleSmall),
          Text(label, style: Theme.of(context).textTheme.labelSmall),
        ],
      ),
    );
  }
}

class _CoverageTimeline extends StatelessWidget {
  const _CoverageTimeline({required this.snapshot});

  final DataSourcesSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    final minYear = snapshot.minCoverage;
    final maxYear = snapshot.maxCoverage;

    if (minYear == null || maxYear == null || minYear >= maxYear) {
      return const SizedBox.shrink();
    }

    final span = (maxYear - minYear).toDouble();

    return PanelCard(
      title: 'Coverage timeline',
      child: Column(
        children: snapshot.datasets
            .map((dataset) {
              final from = dataset.coverageFrom;
              final to = dataset.coverageTo;
              final active = from != null && to != null;

              final start = active ? ((from - minYear) / span).clamp(0.0, 1.0) : 0.0;
              final width = active ? ((to - from) / span).clamp(0.06, 1.0) : 0.0;

              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Row(
                  children: <Widget>[
                    SizedBox(width: 72, child: Text(dataset.name, style: Theme.of(context).textTheme.labelSmall)),
                    Expanded(
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          final maxWidth = constraints.maxWidth;
                          final left = maxWidth * start;
                          final barWidth = maxWidth * width;

                          return Stack(
                            children: <Widget>[
                              Container(
                                height: 18,
                                decoration: BoxDecoration(
                                  color: Theme.of(context).colorScheme.surfaceContainerLow,
                                  borderRadius: BorderRadius.circular(4),
                                ),
                              ),
                              if (active)
                                Positioned(
                                  left: left,
                                  child: Container(
                                    width: barWidth,
                                    height: 18,
                                    decoration: BoxDecoration(
                                      color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.75),
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                  ),
                                ),
                            ],
                          );
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '${dataset.coverageFrom ?? '—'}–${dataset.coverageTo ?? '—'}',
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                  ],
                ),
              );
            })
            .toList(growable: false),
      ),
    );
  }
}
