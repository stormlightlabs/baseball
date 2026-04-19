import 'package:bigfly_mobile/features/teams/application/teams_cubit.dart';
import 'package:bigfly_mobile/features/teams/application/teams_state.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class TeamSearchBar extends StatelessWidget {
  const TeamSearchBar({super.key, required this.controller, required this.state});

  final TextEditingController controller;
  final TeamsState state;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(
          child: TextField(
            controller: controller,
            onChanged: context.read<TeamsCubit>().setSearchQuery,
            onSubmitted: (_) => context.read<TeamsCubit>().runSearch(),
            decoration: const InputDecoration(
              hintText: 'Franchise, city, abbreviation...',
              border: OutlineInputBorder(),
              isDense: true,
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(
          onPressed: state.searchLoading ? null : () => context.read<TeamsCubit>().runSearch(),
          child: const Icon(Icons.arrow_forward),
        ),
      ],
    );
  }
}
