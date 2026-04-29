import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:bigfly_mobile/features/scorekeeper/presentation/scorekeeper_routes.dart';
import 'package:flutter/material.dart';

class ActiveScoringScreen extends StatefulWidget {
  const ActiveScoringScreen({super.key, required this.repository, required this.gameUuid});

  final ScorecardRepository repository;
  final String gameUuid;

  @override
  State<ActiveScoringScreen> createState() => _ActiveScoringScreenState();
}

class _ActiveScoringScreenState extends State<ActiveScoringScreen> {
  late Future<ScorecardGameDetail?> _future;

  @override
  void initState() {
    super.initState();
    _future = widget.repository.getGame(widget.gameUuid);
  }

  Future<void> _reload() async {
    final refreshed = widget.repository.getGame(widget.gameUuid);
    setState(() => _future = refreshed);
    await refreshed;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Active Scoring'),
        actions: <Widget>[
          IconButton(
            tooltip: 'Grid',
            icon: const Icon(Icons.grid_view),
            onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.grid(widget.gameUuid)),
          ),
          IconButton(
            tooltip: 'Export',
            icon: const Icon(Icons.share),
            onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.export(widget.gameUuid)),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _reload,
        child: FutureBuilder<ScorecardGameDetail?>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState != ConnectionState.done) {
              return ListView(
                physics: const AlwaysScrollableScrollPhysics(),
                children: <Widget>[SizedBox(height: 240, child: Center(child: CircularProgressIndicator()))],
              );
            }
            if (snapshot.hasError) {
              return ListView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(16),
                children: <Widget>[Text('Failed to load game: ${snapshot.error}')],
              );
            }
            final detail = snapshot.data;
            if (detail == null) {
              return ListView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: EdgeInsets.all(16),
                children: <Widget>[Text('Game not found.')],
              );
            }

            final summary = detail.summary;
            return ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.all(16),
              children: <Widget>[
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          '${summary.awayTeamAbbreviation} ${summary.awayScore} @ ${summary.homeTeamAbbreviation} ${summary.homeScore}',
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 6),
                        Text('Inning ${detail.innings.isEmpty ? 1 : detail.innings.last.inningNumber}'),
                        Text('Pitch Count: ${summary.pitchCount}'),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                Text('Recorded Innings: ${detail.innings.length}'),
                const SizedBox(height: 6),
                Text('Recorded Plays: ${detail.innings.fold<int>(0, (sum, inning) => sum + inning.plays.length)}'),
                const SizedBox(height: 14),
                FilledButton.tonalIcon(
                  onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.grid(widget.gameUuid)),
                  icon: const Icon(Icons.swipe_right_alt),
                  label: const Text('Open Scorecard Grid'),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}
