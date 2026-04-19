import 'package:bigfly_mobile/features/teams/application/teams_state.dart';
import 'package:bigfly_mobile/features/teams/application/teams_types.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_detail_view.dart';
import 'package:flutter/material.dart';

class TeamDetailContent extends StatelessWidget {
  const TeamDetailContent({
    super.key,
    required this.state,
    required this.onSegmentSelected,
    required this.onSelectFranchise,
  });

  final TeamsState state;
  final ValueChanged<TeamDetailSegment> onSegmentSelected;
  final ValueChanged<String> onSelectFranchise;

  @override
  Widget build(BuildContext context) {
    return switch (state.detailStatus) {
      TeamDetailStatus.initial || TeamDetailStatus.loading => const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator()),
      ),
      TeamDetailStatus.failure => Card(
        child: ListTile(
          leading: Icon(Icons.error_outline, color: Theme.of(context).colorScheme.error),
          title: const Text('Failed to load team'),
          subtitle: Text(state.detailError ?? 'Unknown error'),
        ),
      ),
      TeamDetailStatus.ready when state.detail != null => TeamDetailView(
        detail: state.detail!,
        activeSegment: state.activeSegment,
        franchises: state.franchises,
        onSegmentSelected: onSegmentSelected,
        onSelectFranchise: onSelectFranchise,
      ),
      _ => const SizedBox.shrink(),
    };
  }
}
