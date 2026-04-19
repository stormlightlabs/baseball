import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/shell/app_tabs.dart';
import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/app/ui/branding/big_fly_wordmark.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class RootShell extends StatelessWidget {
  const RootShell({super.key});

  @override
  Widget build(BuildContext context) {
    final navState = context.watch<NavigationCubit>().state;
    final teamSeedColor = context.select((ThemeCubit cubit) => cubit.state.selectedSeedColor);
    final tabViews = appTabViews.values.toList(growable: false);
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        toolbarHeight: 52,
        titleSpacing: 16,
        centerTitle: false,
        elevation: 0,
        scrolledUnderElevation: 0,
        backgroundColor: colorScheme.surface,
        surfaceTintColor: Colors.transparent,
        title: BigFlyWordmark(flyColor: teamSeedColor ?? colorScheme.primary),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(1),
          child: Divider(height: 1, thickness: 1, color: colorScheme.outlineVariant.withValues(alpha: 0.55)),
        ),
      ),
      body: IndexedStack(index: navState.index, children: tabViews),
      bottomNavigationBar: NavigationBar(
        height: 64,
        selectedIndex: navState.index,
        destinations: appTabDestinations,
        onDestinationSelected: context.read<NavigationCubit>().setIndex,
      ),
    );
  }
}
