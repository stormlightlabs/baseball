import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class PlayerSearchBar extends StatelessWidget {
  const PlayerSearchBar({super.key, required this.controller, required this.state});

  final TextEditingController controller;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(
          child: TextField(
            controller: controller,
            onChanged: context.read<PlayersCubit>().setSearchQuery,
            onSubmitted: (_) => context.read<PlayersCubit>().runSearch(),
            decoration: const InputDecoration(
              hintText: 'Name, player ID...',
              border: OutlineInputBorder(),
              isDense: true,
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(
          onPressed: state.searchLoading ? null : () => context.read<PlayersCubit>().runSearch(),
          child: const Icon(Icons.arrow_forward),
        ),
      ],
    );
  }
}
