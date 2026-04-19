import 'package:bigfly_mobile/app/navigation/navigation_state.dart';
import 'package:flutter/material.dart';

class HomeQuickLink {
  const HomeQuickLink({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.snackLabel,
    this.tab,
  });

  final String title;
  final String subtitle;
  final IconData icon;
  final AppTab? tab;
  final String snackLabel;
}

class HomeFeaturedQuery {
  const HomeFeaturedQuery({required this.name, required this.endpoint});

  final String name;
  final String endpoint;
}

class HomeEraChip {
  const HomeEraChip({required this.label, required this.query});

  final String label;
  final String query;
}

const List<HomeQuickLink> homeQuickLinks = <HomeQuickLink>[
  HomeQuickLink(
    title: 'Players',
    subtitle: '/players/{id}',
    icon: Icons.person_outline,
    tab: AppTab.players,
    snackLabel: 'Opening Players tab',
  ),
  HomeQuickLink(
    title: 'Teams',
    subtitle: '/franchises/{id}',
    icon: Icons.shield_outlined,
    tab: AppTab.teams,
    snackLabel: 'Opening Teams tab',
  ),
  HomeQuickLink(
    title: 'Games',
    subtitle: '/games',
    icon: Icons.sports_baseball_outlined,
    tab: AppTab.games,
    snackLabel: 'Opening Games tab',
  ),
  HomeQuickLink(
    title: 'Seasons',
    subtitle: '/seasons/{year}',
    icon: Icons.calendar_month_outlined,
    tab: AppTab.more,
    snackLabel: 'Opening More tab',
  ),
  HomeQuickLink(
    title: 'Leaders',
    subtitle: '/stats/batting',
    icon: Icons.leaderboard_outlined,
    tab: AppTab.more,
    snackLabel: 'Opening More tab',
  ),
  HomeQuickLink(
    title: 'Compare',
    subtitle: 'side-by-side',
    icon: Icons.compare_arrows_outlined,
    tab: AppTab.more,
    snackLabel: 'Opening More tab',
  ),
];

const List<HomeFeaturedQuery> homeFeaturedQueries = <HomeFeaturedQuery>[
  HomeFeaturedQuery(
    name: 'Top HR hitters, 1920–1960',
    endpoint: '/stats/batting?season_from=1920&season_to=1960&sort_by=hr',
  ),
  HomeFeaturedQuery(name: 'Extra-inning games in 1986', endpoint: '/achievements/extra-inning-games?season=1986'),
  HomeFeaturedQuery(name: 'Best ERAs: min 200 IP', endpoint: '/stats/pitching?season=2000&min_ip=200&sort_by=era'),
  HomeFeaturedQuery(
    name: 'Win expectancy — 9th-inning jam',
    endpoint: '/win-expectancy?inning=9&outs=2&runners=12_&score_diff=-1',
  ),
];

const List<HomeEraChip> homeEraChips = <HomeEraChip>[
  HomeEraChip(label: 'fed 1914-15', query: '1914-15'),
  HomeEraChip(label: 'nlg 1935-49', query: '1935-49'),
  HomeEraChip(label: 'boomer 1950-62', query: '1950-62'),
  HomeEraChip(label: 'pitcher 1963-68', query: '1963-68'),
  HomeEraChip(label: 'turf 1969-93', query: '1969-93'),
  HomeEraChip(label: 'steroid 1994-04', query: '1994-04'),
  HomeEraChip(label: 'moneyball 2005-12', query: '2005-12'),
  HomeEraChip(label: 'statcast 2013-19', query: '2013-19'),
  HomeEraChip(label: 'modern 2020-25', query: '2020-25'),
];
