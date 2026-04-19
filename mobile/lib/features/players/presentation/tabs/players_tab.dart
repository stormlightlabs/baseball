import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/players/application/player_selection_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_detail_content.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_search_bar.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_search_results_list.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/recent_players_list.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class PlayersTab extends StatefulWidget {
  const PlayersTab({super.key});

  @override
  State<PlayersTab> createState() => _PlayersTabState();
}

class _PlayersTabState extends State<PlayersTab> {
  late final TextEditingController _searchController;

  @override
  void initState() {
    super.initState();
    final cubit = context.read<PlayersCubit>();
    _searchController = TextEditingController(text: cubit.state.searchQuery);
    cubit.initialize(initialPlayerId: context.read<PlayerSelectionCubit>().state);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: <BlocListener<dynamic, dynamic>>[
        BlocListener<PlayerSelectionCubit, String?>(
          listener: (context, playerId) {
            if (playerId == null || playerId.isEmpty) {
              return;
            }
            final cubit = context.read<PlayersCubit>();
            if (cubit.state.selectedPlayerId == playerId && cubit.state.detailStatus == PlayerDetailStatus.ready) {
              return;
            }
            cubit.loadPlayer(playerId);
          },
        ),
        BlocListener<PlayersCubit, PlayersState>(
          listenWhen: (previous, next) => previous.detail?.themeTeamCode != next.detail?.themeTeamCode,
          listener: (context, state) {
            final themeCode = state.detail?.themeTeamCode;
            if (themeCode != null) {
              context.read<ThemeCubit>().selectTeam(themeCode);
            }
          },
        ),
      ],
      child: BlocBuilder<PlayersCubit, PlayersState>(
        builder: (context, state) {
          return ListView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 20),
            children: <Widget>[
              Text('Players', style: Theme.of(context).textTheme.headlineMedium),
              const SizedBox(height: 10),
              PlayerSearchBar(controller: _searchController, state: state),
              if (state.searchError != null) ...<Widget>[
                const SizedBox(height: 8),
                InlineErrorText(state.searchError!),
              ],
              if (state.searchLoading) ...<Widget>[const SizedBox(height: 8), const LoadingStrip()],
              if (state.searchResults.isNotEmpty) ...<Widget>[
                const SizedBox(height: 8),
                PlayerSearchResultsList(results: state.searchResults, onSelectPlayer: _onSelectPlayer),
              ],
              const SizedBox(height: 8),
              PlayerDetailContent(state: state),
              const SizedBox(height: 16),
              RecentPlayersList(recentPlayers: state.recentPlayers, onSelectPlayer: _onSelectPlayer),
            ],
          );
        },
      ),
    );
  }

  void _onSelectPlayer(String playerId) {
    context.read<PlayerSelectionCubit>().selectPlayer(playerId);
    context.read<PlayersCubit>().loadPlayer(playerId);
  }
}
