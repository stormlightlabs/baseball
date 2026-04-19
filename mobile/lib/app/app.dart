import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:bigfly_mobile/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/data/repositories/player_repository.dart';
import 'package:bigfly_mobile/features/home/home_cubit.dart';
import 'package:bigfly_mobile/features/home/home_tab.dart';
import 'package:bigfly_mobile/features/players/player_selection_cubit.dart';
import 'package:bigfly_mobile/features/players/players_cubit.dart';
import 'package:bigfly_mobile/features/players/players_tab.dart';
import 'package:bigfly_mobile/features/shared/placeholder_tab.dart';
import 'package:bigfly_mobile/features/teams/teams_tab.dart';
import 'package:dynamic_color/dynamic_color.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class BigFlyApp extends StatelessWidget {
  const BigFlyApp({
    super.key,
    required this.cacheStore,
    required this.homeRepository,
    required this.playerRepository,
    this.useDynamicColor = true,
  });

  final CacheStore cacheStore;
  final HomeRepository homeRepository;
  final PlayerRepository playerRepository;
  final bool useDynamicColor;

  @override
  Widget build(BuildContext context) {
    if (!useDynamicColor) {
      return _buildWithDynamicColors(dynamicLightColor: null, dynamicDarkColor: null);
    }

    return DynamicColorBuilder(
      builder: (lightDynamic, darkDynamic) =>
          _buildWithDynamicColors(dynamicLightColor: lightDynamic, dynamicDarkColor: darkDynamic),
    );
  }

  Widget _buildWithDynamicColors({required ColorScheme? dynamicLightColor, required ColorScheme? dynamicDarkColor}) {
    return MultiBlocProvider(
      providers: <BlocProvider<dynamic>>[
        BlocProvider<NavigationCubit>(create: (_) => NavigationCubit()),
        BlocProvider<ThemeCubit>(create: (_) => ThemeCubit(cacheStore)),
        BlocProvider<PlayerSelectionCubit>(create: (_) => PlayerSelectionCubit()),
        BlocProvider<HomeCubit>(create: (_) => HomeCubit(homeRepository)..initialize()),
        BlocProvider<PlayersCubit>(create: (_) => PlayersCubit(playerRepository)),
      ],
      child: _AppView(dynamicLightColor: dynamicLightColor, dynamicDarkColor: dynamicDarkColor),
    );
  }
}

class _AppView extends StatelessWidget {
  const _AppView({required this.dynamicLightColor, required this.dynamicDarkColor});

  final ColorScheme? dynamicLightColor;
  final ColorScheme? dynamicDarkColor;

  @override
  Widget build(BuildContext context) {
    final themeState = context.watch<ThemeCubit>().state;
    final seedColor = themeState.selectedSeedColor ?? const Color(0xFF3B82F6);

    final darkScheme = themeState.selectedSeedColor != null
        ? ColorScheme.fromSeed(seedColor: seedColor, brightness: Brightness.dark)
        : dynamicDarkColor ?? ColorScheme.fromSeed(seedColor: seedColor, brightness: Brightness.dark);

    final lightScheme = themeState.selectedSeedColor != null
        ? ColorScheme.fromSeed(seedColor: seedColor, brightness: Brightness.light)
        : dynamicLightColor ?? ColorScheme.fromSeed(seedColor: seedColor, brightness: Brightness.light);

    return MaterialApp(
      title: 'Big Fly',
      debugShowCheckedModeBanner: false,
      themeMode: ThemeMode.dark,
      theme: _buildTheme(lightScheme),
      darkTheme: _buildTheme(darkScheme),
      home: const _RootShell(),
    );
  }

  ThemeData _buildTheme(ColorScheme colorScheme) {
    final base = ThemeData(useMaterial3: true, colorScheme: colorScheme, brightness: colorScheme.brightness);

    return base.copyWith(
      textTheme: buildAppTextTheme(base.textTheme),
      extensions: <ThemeExtension<dynamic>>[AppTypography.fallback(colorScheme)],
    );
  }
}

class _RootShell extends StatelessWidget {
  const _RootShell();

  @override
  Widget build(BuildContext context) {
    final tabIndex = context.watch<NavigationCubit>().state;
    final tabs = <Widget>[
      const HomeTab(),
      const PlayersTab(),
      const TeamsTab(),
      const PlaceholderTab(title: 'Games', description: 'Schedules, matchups, and game details.'),
      const PlaceholderTab(title: 'More', description: 'More baseball tools and extras.'),
    ];

    return Scaffold(
      body: IndexedStack(index: tabIndex, children: tabs),
      bottomNavigationBar: NavigationBar(
        selectedIndex: tabIndex,
        destinations: const <NavigationDestination>[
          NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home), label: 'Home'),
          NavigationDestination(icon: Icon(Icons.person_outline), selectedIcon: Icon(Icons.person), label: 'Players'),
          NavigationDestination(icon: Icon(Icons.shield_outlined), selectedIcon: Icon(Icons.shield), label: 'Teams'),
          NavigationDestination(
            icon: Icon(Icons.sports_baseball_outlined),
            selectedIcon: Icon(Icons.sports_baseball),
            label: 'Games',
          ),
          NavigationDestination(icon: Icon(Icons.more_horiz), selectedIcon: Icon(Icons.more), label: 'More'),
        ],
        onDestinationSelected: context.read<NavigationCubit>().setIndex,
      ),
    );
  }
}
