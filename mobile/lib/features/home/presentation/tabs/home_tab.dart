import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/navigation/navigation_state.dart';
import 'package:bigfly_mobile/app/ui/chips/section_label.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/home/application/home_cubit.dart';
import 'package:bigfly_mobile/features/home/application/home_state.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/presentation/constants/home_constants.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_entity_chips.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_era_chips.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_featured_queries.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_quick_access_grid.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_search_bar.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/home_search_results_list.dart';
import 'package:bigfly_mobile/features/home/presentation/widgets/meta_health_strip.dart';
import 'package:bigfly_mobile/features/players/application/player_selection_cubit.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class HomeTab extends StatefulWidget {
  const HomeTab({super.key});

  @override
  State<HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends State<HomeTab> {
  late final TextEditingController _searchController;

  @override
  void initState() {
    super.initState();
    final query = context.read<HomeCubit>().state.query;
    _searchController = TextEditingController(text: query);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<HomeCubit, HomeState>(
      builder: (context, state) {
        return RefreshIndicator(
          onRefresh: () => context.read<HomeCubit>().loadMeta(),
          child: ListView(
            padding: const EdgeInsets.fromLTRB(16, 20, 16, 24),
            children: <Widget>[
              Text('What can I find?', style: Theme.of(context).textTheme.headlineMedium),
              const SizedBox(height: 6),
              Text(
                'Search 1871–2025 across players, teams, games, franchises, and seasons.',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 14),
              HomeSearchBar(
                controller: _searchController,
                loading: state.searchStatus == HomeSearchStatus.loading,
                onQueryChanged: context.read<HomeCubit>().setQuery,
                onRunSearch: context.read<HomeCubit>().runSearch,
              ),
              const SizedBox(height: 10),
              HomeEntityChips(
                activeEntity: state.activeEntity,
                onSelectEntity: (entity) => _onEntitySelected(state, entity),
              ),
              if (state.searchStatus == HomeSearchStatus.loading) ...<Widget>[
                const SizedBox(height: 10),
                const LoadingStrip(),
              ],
              if (state.searchError != null) ...<Widget>[
                const SizedBox(height: 10),
                InlineErrorText(state.searchError!),
              ],
              if (state.searchResults.isNotEmpty) ...<Widget>[
                const SizedBox(height: 12),
                HomeSearchResultsList(results: state.searchResults, onTapResult: _handleSearchResultTap),
              ],
              const SizedBox(height: 20),
              const SectionLabel('Quick access'),
              const SizedBox(height: 10),
              HomeQuickAccessGrid(links: homeQuickLinks, onTapLink: _handleQuickLinkTap),
              const SizedBox(height: 20),
              const SectionLabel('Featured queries'),
              const SizedBox(height: 10),
              const HomeFeaturedQueries(queries: homeFeaturedQueries),
              const SizedBox(height: 14),
              MetaHealthStrip(status: state.metaStatus, meta: state.meta, error: state.metaError),
              const SizedBox(height: 20),
              const SectionLabel('Jump to era'),
              const SizedBox(height: 10),
              HomeEraChips(chips: homeEraChips, onTapChip: _handleEraChipTap),
            ],
          ),
        );
      },
    );
  }

  void _onEntitySelected(HomeState state, HomeEntityType entity) {
    final cubit = context.read<HomeCubit>();
    cubit.setEntity(entity);
    if (state.query.trim().isNotEmpty) {
      cubit.runSearch();
    }
  }

  void _handleSearchResultTap(HomeSearchItem item) {
    switch (item.entity) {
      case HomeEntityType.players:
        if (item.id == null) {
          return;
        }
        context.read<PlayerSelectionCubit>().selectPlayer(item.id!);
        context.read<NavigationCubit>().setTab(AppTab.players);
        return;
      case HomeEntityType.teams:
      case HomeEntityType.franchises:
        context.read<NavigationCubit>().setTab(AppTab.teams);
        return;
      case HomeEntityType.games:
        context.read<NavigationCubit>().setTab(AppTab.games);
        return;
      case HomeEntityType.seasons:
        context.read<NavigationCubit>().setTab(AppTab.more);
        return;
    }
  }

  void _handleQuickLinkTap(HomeQuickLink link) {
    if (link.tab != null) {
      context.read<NavigationCubit>().setTab(link.tab!);
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(link.snackLabel)));
  }

  void _handleEraChipTap(HomeEraChip era) {
    _searchController.text = era.query;
    final cubit = context.read<HomeCubit>();
    cubit.setEntity(HomeEntityType.seasons);
    cubit.setQuery(era.query);
    cubit.runSearch();
  }
}
