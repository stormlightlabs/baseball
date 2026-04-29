import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:bigfly_mobile/features/scorekeeper/presentation/scorekeeper_routes.dart';
import 'package:flutter/material.dart';

class ScorecardHubScreen extends StatefulWidget {
  const ScorecardHubScreen({super.key, required this.repository});

  final ScorecardRepository repository;

  @override
  State<ScorecardHubScreen> createState() => _ScorecardHubScreenState();
}

class _ScorecardHubScreenState extends State<ScorecardHubScreen> {
  ScorecardStatusFilter _filter = ScorecardStatusFilter.all;
  late Future<List<ScorecardGameSummary>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<ScorecardGameSummary>> _load() {
    return widget.repository.listGames(filter: _filter);
  }

  Future<void> _refresh() async {
    final loaded = _load();
    setState(() => _future = loaded);
    await loaded;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Scorekeeper'),
        actions: <Widget>[
          IconButton(
            tooltip: 'Import (stub)',
            icon: const Icon(Icons.upload_file_outlined),
            onPressed: () {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(const SnackBar(content: Text('Import is planned in a later phase.')));
            },
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.newGame).then((_) => _refresh()),
        icon: const Icon(Icons.add),
        label: const Text('New Game'),
      ),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: ListView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
          children: <Widget>[
            Wrap(
              spacing: 8,
              children: <Widget>[
                ChoiceChip(
                  label: const Text('All'),
                  selected: _filter == ScorecardStatusFilter.all,
                  onSelected: (_) => _setFilter(ScorecardStatusFilter.all),
                ),
                ChoiceChip(
                  label: const Text('In Progress'),
                  selected: _filter == ScorecardStatusFilter.inProgress,
                  onSelected: (_) => _setFilter(ScorecardStatusFilter.inProgress),
                ),
                ChoiceChip(
                  label: const Text('Final'),
                  selected: _filter == ScorecardStatusFilter.finalGame,
                  onSelected: (_) => _setFilter(ScorecardStatusFilter.finalGame),
                ),
              ],
            ),
            const SizedBox(height: 10),
            FutureBuilder<List<ScorecardGameSummary>>(
              future: _future,
              builder: (context, snapshot) {
                if (snapshot.connectionState != ConnectionState.done) {
                  return const Padding(
                    padding: EdgeInsets.all(30),
                    child: Center(child: CircularProgressIndicator()),
                  );
                }

                if (snapshot.hasError) {
                  return Card(
                    child: Padding(
                      padding: const EdgeInsets.all(14),
                      child: Text('Failed to load scorecards: ${snapshot.error}'),
                    ),
                  );
                }

                final games = snapshot.data ?? const <ScorecardGameSummary>[];
                if (games.isEmpty) {
                  return Card(
                    child: Padding(
                      padding: const EdgeInsets.all(18),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text('No saved scorecards yet', style: Theme.of(context).textTheme.titleMedium),
                          const SizedBox(height: 6),
                          const Text('Start your first scorecard to track games offline.'),
                        ],
                      ),
                    ),
                  );
                }

                return Column(
                  children: games
                      .map(
                        (game) => Padding(
                          padding: const EdgeInsets.only(bottom: 10),
                          child: _GameCard(game: game),
                        ),
                      )
                      .toList(growable: false),
                );
              },
            ),
          ],
        ),
      ),
    );
  }

  void _setFilter(ScorecardStatusFilter filter) {
    if (_filter == filter) {
      return;
    }

    setState(() {
      _filter = filter;
      _future = _load();
    });
  }
}

class _GameCard extends StatelessWidget {
  const _GameCard({required this.game});

  final ScorecardGameSummary game;

  @override
  Widget build(BuildContext context) {
    final statusLabel = switch (game.status) {
      ScorecardStatus.inProgress => 'In Progress',
      ScorecardStatus.finalGame => 'Final',
    };

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Expanded(
                  child: Text(
                    '${game.awayTeamAbbreviation} ${game.awayScore} @ ${game.homeTeamAbbreviation} ${game.homeScore}',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                Chip(label: Text(statusLabel)),
              ],
            ),
            const SizedBox(height: 4),
            Text('${game.awayTeamName} vs ${game.homeTeamName}'),
            const SizedBox(height: 2),
            Text(
              '${game.venue ?? 'Venue TBD'} · ${_date(game.gameDate)} · ${game.pitchCount} pitches',
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 10),
            Wrap(
              spacing: 8,
              children: <Widget>[
                OutlinedButton(
                  onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.activeGame(game.uuid)),
                  child: const Text('Resume'),
                ),
                OutlinedButton(
                  onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.grid(game.uuid)),
                  child: const Text('Box Score'),
                ),
                OutlinedButton(
                  onPressed: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.export(game.uuid)),
                  child: const Text('Export'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  static String _date(DateTime value) {
    final month = value.month.toString().padLeft(2, '0');
    final day = value.day.toString().padLeft(2, '0');
    return '${value.year}-$month-$day';
  }
}
