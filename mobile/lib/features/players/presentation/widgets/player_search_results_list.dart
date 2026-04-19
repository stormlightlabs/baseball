import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';

class PlayerSearchResultsList extends StatelessWidget {
  const PlayerSearchResultsList({super.key, required this.results, required this.onSelectPlayer});

  final List<PlayerSearchResult> results;
  final ValueChanged<String> onSelectPlayer;

  @override
  Widget build(BuildContext context) {
    if (results.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      children: results
          .map(
            (item) => Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text(item.name),
                subtitle: Text(item.subtitle),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => onSelectPlayer(item.id),
              ),
            ),
          )
          .toList(growable: false),
    );
  }
}
