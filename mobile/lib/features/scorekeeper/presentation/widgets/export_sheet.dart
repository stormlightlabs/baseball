import 'package:flutter/material.dart';

class ExportSheet extends StatelessWidget {
  const ExportSheet({super.key, required this.gameUuid});

  final String gameUuid;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 10, 16, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Text('Export Scorecard', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 4),
            Text(gameUuid, style: Theme.of(context).textTheme.bodySmall),
            const SizedBox(height: 12),
            _ExportRow(label: 'PDF', subtitle: 'Traditional two-page scorecard layout'),
            const SizedBox(height: 8),
            _ExportRow(label: 'JSON', subtitle: 'Canonical play-by-play schema'),
            const SizedBox(height: 8),
            _ExportRow(label: 'Markdown', subtitle: 'GFM linescore and batter table'),
          ],
        ),
      ),
    );
  }
}

class _ExportRow extends StatelessWidget {
  const _ExportRow({required this.label, required this.subtitle});

  final String label;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      dense: true,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      tileColor: Theme.of(context).colorScheme.surfaceContainerHighest.withValues(alpha: 0.4),
      leading: const Icon(Icons.share),
      title: Text(label),
      subtitle: Text(subtitle),
      trailing: const Icon(Icons.chevron_right),
      onTap: () {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$label export is stubbed for this phase.')));
      },
    );
  }
}
