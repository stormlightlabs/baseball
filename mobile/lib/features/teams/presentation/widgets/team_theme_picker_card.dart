import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/colors.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class TeamThemePickerCard extends StatelessWidget {
  const TeamThemePickerCard({super.key, required this.selectedTeamCode});

  final String? selectedTeamCode;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        DropdownButtonFormField<String>(
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            labelText: 'Seed theme from team primary color',
          ),
          initialValue: selectedTeamCode,
          items: <DropdownMenuItem<String>>[
            for (final team in MlbTeam.values)
              DropdownMenuItem<String>(value: team.code, child: Text('${team.code} - ${team.displayName}')),
          ],
          onChanged: (teamCode) => context.read<ThemeCubit>().selectTeam(teamCode),
        ),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          onPressed: () => context.read<ThemeCubit>().selectTeam(null),
          icon: const Icon(Icons.colorize),
          label: const Text('Use neutral/system palette'),
        ),
      ],
    );
  }
}
