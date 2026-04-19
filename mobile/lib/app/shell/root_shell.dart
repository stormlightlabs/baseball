import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/shell/app_tabs.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class RootShell extends StatelessWidget {
  const RootShell({super.key});

  @override
  Widget build(BuildContext context) {
    final navState = context.watch<NavigationCubit>().state;
    final tabViews = appTabViews.values.toList(growable: false);

    return Scaffold(
      body: IndexedStack(index: navState.index, children: tabViews),
      bottomNavigationBar: NavigationBar(
        selectedIndex: navState.index,
        destinations: appTabDestinations,
        onDestinationSelected: context.read<NavigationCubit>().setIndex,
      ),
    );
  }
}
