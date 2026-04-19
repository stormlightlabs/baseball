import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';
import 'package:flutter/material.dart';

class TeamSearchResultsList extends StatelessWidget {
  const TeamSearchResultsList({super.key, required this.results, required this.onSelectTeam});

  final List<TeamSeasonRecord> results;
  final ValueChanged<TeamSeasonRecord> onSelectTeam;

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
                onTap: () => onSelectTeam(item),
              ),
            ),
          )
          .toList(growable: false),
    );
  }
}
