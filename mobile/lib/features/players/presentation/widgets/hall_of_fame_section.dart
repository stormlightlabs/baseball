import 'package:bigfly_mobile/app/ui/cards/panel_card.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:flutter/material.dart';

class HallOfFameSection extends StatelessWidget {
  const HallOfFameSection({super.key, required this.records});

  final List<PlayerHallOfFameRecord> records;

  @override
  Widget build(BuildContext context) {
    if (records.isEmpty) {
      return const PanelCard(title: 'Hall of Fame', child: Text('No Hall of Fame records found.'));
    }

    final sorted = [...records]..sort((a, b) => a.year.compareTo(b.year));

    return PanelCard(
      title: 'Hall of Fame',
      child: Column(
        children: sorted
            .map(
              (record) => ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Icon(
                  record.inducted ? Icons.verified : Icons.how_to_vote_outlined,
                  color: record.inducted ? Colors.amber : null,
                ),
                title: Text('${record.year} · ${record.votedBy}'),
                subtitle: Text(
                  record.votePercent == null
                      ? 'Votes: ${record.votes ?? '—'}'
                      : 'Votes: ${record.votes}/${record.ballots} (${record.votePercent!.toStringAsFixed(1)}%)',
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}
