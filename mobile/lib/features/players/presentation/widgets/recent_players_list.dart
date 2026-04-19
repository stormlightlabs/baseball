import 'package:bigfly_mobile/app/ui/chips/section_label.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';

class RecentPlayersList extends StatelessWidget {
  const RecentPlayersList({super.key, required this.recentPlayers, required this.onSelectPlayer});

  final List<PlayerSearchResult> recentPlayers;
  final ValueChanged<String> onSelectPlayer;

  @override
  Widget build(BuildContext context) {
    if (recentPlayers.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const SectionLabel('Recently viewed'),
        const SizedBox(height: 8),
        ...recentPlayers.map(
          (player) => Card(
            margin: const EdgeInsets.only(bottom: 8),
            child: ListTile(
              leading: CircleAvatar(child: Text(player.name.isEmpty ? '?' : player.name[0].toUpperCase())),
              title: Text(player.name),
              subtitle: Text(player.subtitle),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => onSelectPlayer(player.id),
            ),
          ),
        ),
      ],
    );
  }
}
