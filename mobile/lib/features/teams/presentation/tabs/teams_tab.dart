import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/teams/application/teams_cubit.dart';
import 'package:bigfly_mobile/features/teams/application/teams_state.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_detail_content.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_search_bar.dart';
import 'package:bigfly_mobile/features/teams/presentation/widgets/team_search_results_list.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class TeamsTab extends StatefulWidget {
  const TeamsTab({super.key});

  @override
  State<TeamsTab> createState() => _TeamsTabState();
}

class _TeamsTabState extends State<TeamsTab> {
  late final TextEditingController _searchController;

  @override
  void initState() {
    super.initState();
    final cubit = context.read<TeamsCubit>();
    _searchController = TextEditingController(text: cubit.state.searchQuery);
    cubit.initialize(initialTeamCode: context.read<ThemeCubit>().state.selectedTeamCode);
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
        BlocListener<TeamsCubit, TeamsState>(
          listenWhen: (previous, next) => previous.detail?.themeTeamCode != next.detail?.themeTeamCode,
          listener: (context, state) {
            final themeCode = state.detail?.themeTeamCode;
            if (themeCode != null) {
              context.read<ThemeCubit>().selectTeam(themeCode);
            }
          },
        ),
      ],
      child: BlocBuilder<TeamsCubit, TeamsState>(
        builder: (context, state) {
          return ListView(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
            children: <Widget>[
              TeamSearchBar(controller: _searchController, state: state),
              if (state.searchError != null) ...<Widget>[
                const SizedBox(height: 8),
                InlineErrorText(state.searchError!),
              ],
              if (state.searchLoading) ...<Widget>[const SizedBox(height: 8), const LoadingStrip()],
              if (state.searchResults.isNotEmpty) ...<Widget>[
                const SizedBox(height: 8),
                TeamSearchResultsList(results: state.searchResults, onSelectTeam: context.read<TeamsCubit>().loadTeam),
              ],
              const SizedBox(height: 8),
              TeamDetailContent(
                state: state,
                onSegmentSelected: context.read<TeamsCubit>().setActiveSegment,
                onSelectFranchise: context.read<TeamsCubit>().loadTeamForCode,
              ),
            ],
          );
        },
      ),
    );
  }
}
