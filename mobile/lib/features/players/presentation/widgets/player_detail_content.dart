import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_detail_view.dart';
import 'package:flutter/material.dart';

class PlayerDetailContent extends StatelessWidget {
  const PlayerDetailContent({super.key, required this.state});

  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    return switch (state.detailStatus) {
      PlayerDetailStatus.initial || PlayerDetailStatus.loading => const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator()),
      ),
      PlayerDetailStatus.failure => Card(
        child: ListTile(
          leading: Icon(Icons.error_outline, color: Theme.of(context).colorScheme.error),
          title: const Text('Failed to load player'),
          subtitle: Text(state.detailError ?? 'Unknown error'),
        ),
      ),
      PlayerDetailStatus.ready when state.detail != null => PlayerDetailView(detail: state.detail!, state: state),
      _ => const SizedBox.shrink(),
    };
  }
}
