import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/selected_team_card.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_theme_picker_card.dart';
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
            TeamThemePickerCard(selectedTeamCode: state.selectedTeamCode),
            if (state.selectedTeam != null) ...<Widget>[
              const SizedBox(height: 16),
              SelectedTeamCard(team: state.selectedTeam!),
            ],
          ],
        );
      },
    );
  }
}
