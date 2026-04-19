import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';

class AwardsSection extends StatelessWidget {
  const AwardsSection({super.key, required this.awards});

  final List<PlayerAward> awards;

  @override
  Widget build(BuildContext context) {
    if (awards.isEmpty) {
      return const PanelCard(title: 'Awards & honors', child: Text('No awards found for this player.'));
    }

    final grouped = <String, List<int>>{};
    for (final award in awards) {
      grouped.putIfAbsent(award.awardId, () => <int>[]).add(award.year);
    }

    final entries = grouped.entries.toList(growable: false)..sort((a, b) => a.key.compareTo(b.key));

    return PanelCard(
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
