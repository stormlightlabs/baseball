import 'package:bigfly_mobile/colors.dart';
import 'package:flutter/material.dart';

class SelectedTeamCard extends StatelessWidget {
  const SelectedTeamCard({super.key, required this.team});

  final MlbTeam team;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(title: Text('Selected Team: ${team.displayName}'), subtitle: Text('Primary: ${team.primaryHex}')),
    );
  }
}
