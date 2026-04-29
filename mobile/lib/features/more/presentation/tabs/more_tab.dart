import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/compare_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/data_sources_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/leaders_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/seasons_screen.dart';
import 'package:bigfly_mobile/features/scorekeeper/presentation/scorekeeper_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class MoreTab extends StatefulWidget {
  const MoreTab({super.key});

  @override
  State<MoreTab> createState() => _MoreTabState();
}

class _MoreTabState extends State<MoreTab> with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  static const List<Tab> _tabs = <Tab>[
    Tab(text: 'Seasons'),
    Tab(text: 'Leaders'),
    Tab(text: 'Compare'),
    Tab(text: 'Data Sources'),
  ];

  @override
  void initState() {
    super.initState();
    final selectedSection = context.read<MoreCubit>().state.selectedSection;
    _tabController = TabController(length: MoreSection.values.length, vsync: this, initialIndex: selectedSection.index);
    context.read<MoreCubit>().initialize();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<MoreCubit, MoreState>(
      listenWhen: (previous, next) => previous.selectedSection != next.selectedSection,
      listener: (context, state) {
        if (_tabController.index == state.selectedSection.index) {
          return;
        }
        _tabController.animateTo(state.selectedSection.index);
      },
      builder: (context, state) {
        final cubit = context.read<MoreCubit>();

        return Column(
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
              child: ListTile(
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                tileColor: Theme.of(context).colorScheme.surfaceContainerHighest.withValues(alpha: 0.35),
                leading: const Icon(Icons.edit_note),
                title: const Text('Scorekeeper'),
                subtitle: const Text('Offline scorecard hub, game setup, and exports'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => Navigator.of(context).pushNamed(ScorekeeperRoutes.hub),
              ),
            ),
            Material(
              color: Theme.of(context).colorScheme.surface,
              child: TabBar(
                controller: _tabController,
                isScrollable: true,
                tabs: _tabs,
                onTap: (index) => cubit.setSection(MoreSection.values[index]),
              ),
            ),
            Expanded(
              child: RefreshIndicator(
                onRefresh: cubit.refreshSelectedSection,
                child: switch (state.selectedSection) {
                  MoreSection.seasons => SeasonsScreen(state: state),
                  MoreSection.leaders => LeadersScreen(state: state),
                  MoreSection.compare => CompareScreen(state: state),
                  MoreSection.dataSources => DataSourcesScreen(state: state),
                },
              ),
            ),
          ],
        );
      },
    );
  }
}
