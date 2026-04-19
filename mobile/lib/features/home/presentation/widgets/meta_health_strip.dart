import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/home/application/home_state.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:flutter/material.dart';

class MetaHealthStrip extends StatelessWidget {
  const MetaHealthStrip({super.key, required this.status, required this.meta, required this.error});

  final HomeMetaStatus status;
  final MetaSnapshot? meta;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    if (status == HomeMetaStatus.loading || status == HomeMetaStatus.initial) {
      return const Card(
        child: Padding(padding: EdgeInsets.all(16), child: LoadingStrip()),
      );
    }

    if (status == HomeMetaStatus.failure || meta == null) {
      return Card(
        child: ListTile(
          leading: Icon(Icons.error_outline, color: colorScheme.error),
          title: const Text('API metadata unavailable'),
          subtitle: Text(error ?? 'Failed to fetch /v1/meta'),
        ),
      );
    }

    final online = meta!.allRequiredHealthy;
    final from = meta!.minCoverageYear?.toString() ?? '—';
    final to = meta!.maxCoverageYear?.toString() ?? '—';
    final sources = meta!.sourceCount.toString();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          children: <Widget>[
            Icon(Icons.circle, size: 10, color: online ? Colors.greenAccent : colorScheme.error),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(online ? 'API Online' : 'API Degraded'),
                  Text('/v1/meta · v${meta!.version}', style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
            _HealthStat(value: from, label: 'FROM'),
            const SizedBox(width: 12),
            _HealthStat(value: to, label: 'TO'),
            const SizedBox(width: 12),
            _HealthStat(value: sources, label: 'SOURCES'),
          ],
        ),
      ),
    );
  }
}

class _HealthStat extends StatelessWidget {
  const _HealthStat({required this.value, required this.label});

  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Text(value, style: Theme.of(context).textTheme.titleSmall),
        Text(label, style: Theme.of(context).extension<AppTypography>()?.code),
      ],
    );
  }
}
