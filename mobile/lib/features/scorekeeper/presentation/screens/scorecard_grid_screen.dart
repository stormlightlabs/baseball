import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:flutter/material.dart';

class ScorecardGridScreen extends StatefulWidget {
  const ScorecardGridScreen({super.key, required this.repository, required this.gameUuid});

  final ScorecardRepository repository;
  final String gameUuid;

  @override
  State<ScorecardGridScreen> createState() => _ScorecardGridScreenState();
}

class _ScorecardGridScreenState extends State<ScorecardGridScreen> {
  ScorecardTeam _team = ScorecardTeam.away;
  late Future<ScorecardGameDetail?> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.repository.getGame(widget.gameUuid);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Scorecard Grid')),
      body: FutureBuilder<ScorecardGameDetail?>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          final game = snapshot.data;
          if (game == null) {
            return const Center(child: Text('Game not found.'));
          }

          final teamLabel = _team == ScorecardTeam.away ? game.summary.awayTeamName : game.summary.homeTeamName;
          final lineups = game.lineups.where((slot) => slot.team == _team).toList(growable: false);

          return ListView(
            padding: const EdgeInsets.all(16),
            children: <Widget>[
              SegmentedButton<ScorecardTeam>(
                segments: const <ButtonSegment<ScorecardTeam>>[
                  ButtonSegment<ScorecardTeam>(value: ScorecardTeam.away, label: Text('Away')),
                  ButtonSegment<ScorecardTeam>(value: ScorecardTeam.home, label: Text('Home')),
                ],
                selected: <ScorecardTeam>{_team},
                onSelectionChanged: (value) => setState(() => _team = value.first),
              ),
              const SizedBox(height: 10),
              Text('$teamLabel lineup', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              if (lineups.isEmpty)
                const Text('No lineup entries were captured yet.')
              else
                ...lineups.map(
                  (slot) => ListTile(
                    dense: true,
                    title: Text('${slot.battingOrder}. ${slot.playerName}'),
                    subtitle: Text(slot.positionCode ?? 'Position TBD'),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}
