import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/bio_stats_grid.dart';
import 'package:flutter/material.dart';

class PlayerProfileCard extends StatelessWidget {
  const PlayerProfileCard({super.key, required this.detail});

  final PlayerDetailBundle detail;

  @override
  Widget build(BuildContext context) {
    final player = detail.player;

    return Card(
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
            BioStatsGrid(detail: detail),
          ],
        ),
      ),
    );
  }
}
