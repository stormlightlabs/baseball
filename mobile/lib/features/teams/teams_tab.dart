import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/colors.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class TeamsTab extends StatelessWidget {
  const TeamsTab({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<ThemeCubit, AppThemeState>(
      builder: (context, state) {
        return ListView(
          padding: const EdgeInsets.all(16),
          children: <Widget>[
            Text('Teams', style: Theme.of(context).textTheme.headlineMedium),
            const SizedBox(height: 8),
            Text('Choose a team color profile for this view.', style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                labelText: 'Seed theme from team primary color',
              ),
              initialValue: state.selectedTeamCode,
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
            if (state.selectedTeam != null) ...<Widget>[
              const SizedBox(height: 16),
              Card(
                child: ListTile(
                  title: Text('Selected Team: ${state.selectedTeam!.displayName}'),
                  subtitle: Text('Primary: ${state.selectedTeam!.primaryHex}'),
                ),
              ),
            ],
          ],
        );
      },
    );
  }
}
