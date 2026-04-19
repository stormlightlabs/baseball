import 'package:bigfly_mobile/app/navigation/navigation_state.dart';
import 'package:bigfly_mobile/features/games/presentation/tabs/games_tab.dart';
import 'package:bigfly_mobile/features/home/presentation/tabs/home_tab.dart';
import 'package:bigfly_mobile/features/more/presentation/tabs/more_tab.dart';
import 'package:bigfly_mobile/features/players/presentation/tabs/players_tab.dart';
import 'package:bigfly_mobile/features/teams/presentation/tabs/teams_tab.dart';
import 'package:flutter/material.dart';

const List<NavigationDestination> appTabDestinations = <NavigationDestination>[
  NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home), label: 'Home'),
  NavigationDestination(icon: Icon(Icons.person_outline), selectedIcon: Icon(Icons.person), label: 'Players'),
  NavigationDestination(icon: Icon(Icons.shield_outlined), selectedIcon: Icon(Icons.shield), label: 'Teams'),
  NavigationDestination(
    icon: Icon(Icons.sports_baseball_outlined),
    selectedIcon: Icon(Icons.sports_baseball),
    label: 'Games',
  ),
  NavigationDestination(icon: Icon(Icons.more_horiz), selectedIcon: Icon(Icons.more), label: 'More'),
];

final Map<AppTab, Widget> appTabViews = <AppTab, Widget>{
  AppTab.home: const HomeTab(),
  AppTab.players: const PlayersTab(),
  AppTab.teams: const TeamsTab(),
  AppTab.games: const GamesTab(),
  AppTab.more: const MoreTab(),
};
