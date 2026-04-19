import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/application/more_state.dart';
import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/compare_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/data_sources_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/leaders_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/seasons_screen.dart';
import 'package:bigfly_mobile/features/more/presentation/widgets/section_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class MoreTab extends StatefulWidget {
  const MoreTab({super.key});

  @override
  State<MoreTab> createState() => _MoreTabState();
}

class _MoreTabState extends State<MoreTab> {
  @override
  void initState() {
    super.initState();
    context.read<MoreCubit>().initialize();
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<MoreCubit, MoreState>(
      builder: (context, state) {
        final cubit = context.read<MoreCubit>();

        return Column(
          children: <Widget>[
            SectionPicker(selected: state.selectedSection, onSelect: cubit.setSection),
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
