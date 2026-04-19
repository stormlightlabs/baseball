import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/data/models/meta_models.dart';
import 'package:bigfly_mobile/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/features/home/home_cubit.dart';
import 'package:bigfly_mobile/features/players/player_selection_cubit.dart';
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
              _buildSearchBar(context, state),
              const SizedBox(height: 10),
              _buildEntityPills(context, state),
              if (state.searchStatus == HomeSearchStatus.loading) ...<Widget>[
                const SizedBox(height: 10),
                const LinearProgressIndicator(minHeight: 2),
              ],
              if (state.searchError != null) ...<Widget>[
                const SizedBox(height: 10),
                Text(
                  state.searchError!,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Theme.of(context).colorScheme.error),
                ),
              ],
              if (state.searchResults.isNotEmpty) ...<Widget>[
                const SizedBox(height: 12),
                ...state.searchResults.map((item) => _buildSearchResultItem(context, item)),
              ],
              const SizedBox(height: 20),
              Text('Quick access', style: Theme.of(context).extension<AppTypography>()?.code),
              const SizedBox(height: 10),
              _buildQuickGrid(context),
              const SizedBox(height: 20),
              Text('Featured queries', style: Theme.of(context).extension<AppTypography>()?.code),
              const SizedBox(height: 10),
              ..._featuredQueries.map((query) => _FeaturedQueryCard(query: query)),
              const SizedBox(height: 14),
              _MetaHealthStrip(status: state.metaStatus, meta: state.meta, error: state.metaError),
              const SizedBox(height: 20),
              Text('Jump to era', style: Theme.of(context).extension<AppTypography>()?.code),
              const SizedBox(height: 10),
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: _eraChips
                    .map(
                      (era) => ActionChip(
                        label: Text(era.label),
                        onPressed: () {
                          _searchController.text = era.query;
                          final cubit = context.read<HomeCubit>();
                          cubit.setEntity(HomeEntityType.seasons);
                          cubit.setQuery(era.query);
                          cubit.runSearch();
                        },
                      ),
                    )
                    .toList(growable: false),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildSearchBar(BuildContext context, HomeState state) {
    return Row(
      children: <Widget>[
        Expanded(
          child: TextField(
            controller: _searchController,
            onChanged: context.read<HomeCubit>().setQuery,
            onSubmitted: (_) => context.read<HomeCubit>().runSearch(),
            decoration: const InputDecoration(
              hintText: '"Babe Ruth", "1927 Yankees"...',
              border: OutlineInputBorder(),
              isDense: true,
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(
          onPressed: state.searchStatus == HomeSearchStatus.loading
              ? null
              : () => context.read<HomeCubit>().runSearch(),
          child: const Text('Go'),
        ),
      ],
    );
  }

  Widget _buildEntityPills(BuildContext context, HomeState state) {
    return SizedBox(
      height: 36,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: HomeEntityType.values
            .map(
              (entity) => Padding(
                padding: const EdgeInsets.only(right: 6),
                child: ChoiceChip(
                  label: Text(entity.label),
                  selected: state.activeEntity == entity,
                  onSelected: (_) {
                    final cubit = context.read<HomeCubit>();
                    cubit.setEntity(entity);
                    if (state.query.trim().isNotEmpty) {
                      cubit.runSearch();
                    }
                  },
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }

  Widget _buildSearchResultItem(BuildContext context, HomeSearchItem item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        title: Text(item.title),
        subtitle: Text(item.subtitle),
        trailing: const Icon(Icons.chevron_right),
        onTap: () => _handleSearchResultTap(context, item),
      ),
    );
  }

  void _handleSearchResultTap(BuildContext context, HomeSearchItem item) {
    switch (item.entity) {
      case HomeEntityType.players:
        if (item.id == null) {
          return;
        }
        context.read<PlayerSelectionCubit>().selectPlayer(item.id!);
        context.read<NavigationCubit>().setIndex(1);
        return;
      case HomeEntityType.teams:
      case HomeEntityType.franchises:
        context.read<NavigationCubit>().setIndex(2);
        return;
      case HomeEntityType.games:
        context.read<NavigationCubit>().setIndex(3);
        return;
      case HomeEntityType.seasons:
        context.read<NavigationCubit>().setIndex(4);
        return;
    }
  }

  Widget _buildQuickGrid(BuildContext context) {
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: _quickLinks.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: 1,
        mainAxisSpacing: 1,
        childAspectRatio: 1.85,
      ),
      itemBuilder: (context, index) {
        final link = _quickLinks[index];
        return InkWell(
          onTap: () {
            if (link.tabIndex != null) {
              context.read<NavigationCubit>().setIndex(link.tabIndex!);
            } else {
              ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(link.snackLabel)));
            }
          },
          child: Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Icon(link.icon, size: 20),
                const SizedBox(height: 6),
                Text(link.title, style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 2),
                Text(link.subtitle, style: Theme.of(context).extension<AppTypography>()?.code),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _MetaHealthStrip extends StatelessWidget {
  const _MetaHealthStrip({required this.status, required this.meta, required this.error});

  final HomeMetaStatus status;
  final MetaSnapshot? meta;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    if (status == HomeMetaStatus.loading || status == HomeMetaStatus.initial) {
      return const Card(
        child: Padding(padding: EdgeInsets.all(16), child: LinearProgressIndicator(minHeight: 2)),
      );
    }

    if (status == HomeMetaStatus.failure || meta == null) {
      return Card(
        child: ListTile(
          leading: Icon(Icons.error_outline, color: colorScheme.error),
          title: const Text('API metadata unavailable'),
          subtitle: Text(error ?? 'Failed to fetch /api/v1/meta'),
        ),
      );
    }

    final online = meta!.allRequiredHealthy;
    final from = meta!.minCoverageYear?.toString() ?? '—';
    final to = meta!.maxCoverageYear?.toString() ?? '—';
    final sources = meta!.sourceCount.toString();

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          children: <Widget>[
            Icon(Icons.circle, size: 10, color: online ? Colors.greenAccent : colorScheme.error),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(online ? 'API Online' : 'API Degraded'),
                  Text('/api/v1/meta · v${meta!.version}', style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
            _HealthStat(value: from, label: 'FROM'),
            const SizedBox(width: 12),
            _HealthStat(value: to, label: 'TO'),
            const SizedBox(width: 12),
            _HealthStat(value: sources, label: 'SOURCES'),
          ],
        ),
      ),
    );
  }
}

class _HealthStat extends StatelessWidget {
  const _HealthStat({required this.value, required this.label});

  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Text(value, style: Theme.of(context).textTheme.titleSmall),
        Text(label, style: Theme.of(context).extension<AppTypography>()?.code),
      ],
    );
  }
}

class _FeaturedQueryCard extends StatelessWidget {
  const _FeaturedQueryCard({required this.query});

  final _FeaturedQuery query;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        title: Text(query.name),
        subtitle: Text(query.endpoint, style: Theme.of(context).extension<AppTypography>()?.code),
      ),
    );
  }
}

class _QuickLink {
  const _QuickLink({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.snackLabel,
    this.tabIndex,
  });

  final String title;
  final String subtitle;
  final IconData icon;
  final int? tabIndex;
  final String snackLabel;
}

class _FeaturedQuery {
  const _FeaturedQuery({required this.name, required this.endpoint});

  final String name;
  final String endpoint;
}

class _EraChipModel {
  const _EraChipModel({required this.label, required this.query});

  final String label;
  final String query;
}

const List<_QuickLink> _quickLinks = <_QuickLink>[
  _QuickLink(
    title: 'Players',
    subtitle: '/players/{id}',
    icon: Icons.person_outline,
    tabIndex: 1,
    snackLabel: 'Opening Players tab',
  ),
  _QuickLink(
    title: 'Teams',
    subtitle: '/franchises/{id}',
    icon: Icons.shield_outlined,
    tabIndex: 2,
    snackLabel: 'Opening Teams tab',
  ),
  _QuickLink(
    title: 'Games',
    subtitle: '/games',
    icon: Icons.sports_baseball_outlined,
    tabIndex: 3,
    snackLabel: 'Opening Games tab',
  ),
  _QuickLink(
    title: 'Seasons',
    subtitle: '/seasons/{year}',
    icon: Icons.calendar_month_outlined,
    tabIndex: 4,
    snackLabel: 'Opening More tab',
  ),
  _QuickLink(
    title: 'Leaders',
    subtitle: '/stats/batting',
    icon: Icons.leaderboard_outlined,
    tabIndex: 4,
    snackLabel: 'Opening More tab',
  ),
  _QuickLink(
    title: 'Compare',
    subtitle: 'side-by-side',
    icon: Icons.compare_arrows_outlined,
    tabIndex: 4,
    snackLabel: 'Opening More tab',
  ),
];

const List<_FeaturedQuery> _featuredQueries = <_FeaturedQuery>[
  _FeaturedQuery(
    name: 'Top HR hitters, 1920–1960',
    endpoint: '/stats/batting?season_from=1920&season_to=1960&sort_by=hr',
  ),
  _FeaturedQuery(name: 'Extra-inning games in 1986', endpoint: '/achievements/extra-inning-games?season=1986'),
  _FeaturedQuery(name: 'Best ERAs: min 200 IP', endpoint: '/stats/pitching?season=2000&min_ip=200&sort_by=era'),
  _FeaturedQuery(
    name: 'Win expectancy — 9th-inning jam',
    endpoint: '/win-expectancy?inning=9&outs=2&runners=12_&score_diff=-1',
  ),
];

const List<_EraChipModel> _eraChips = <_EraChipModel>[
  _EraChipModel(label: 'fed 1914-15', query: '1914-15'),
  _EraChipModel(label: 'nlg 1935-49', query: '1935-49'),
  _EraChipModel(label: 'boomer 1950-62', query: '1950-62'),
  _EraChipModel(label: 'pitcher 1963-68', query: '1963-68'),
  _EraChipModel(label: 'turf 1969-93', query: '1969-93'),
  _EraChipModel(label: 'steroid 1994-04', query: '1994-04'),
  _EraChipModel(label: 'moneyball 2005-12', query: '2005-12'),
  _EraChipModel(label: 'statcast 2013-19', query: '2013-19'),
  _EraChipModel(label: 'modern 2020-25', query: '2020-25'),
];
