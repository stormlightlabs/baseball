import 'package:bigfly_mobile/app/navigation/navigation_cubit.dart';
import 'package:bigfly_mobile/app/shell/root_shell.dart';
import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/theme/theme_cubit.dart';
import 'package:bigfly_mobile/core/data/local/cache_store.dart';
import 'package:bigfly_mobile/features/games/application/games_cubit.dart';
import 'package:bigfly_mobile/features/games/data/repositories/game_repository.dart';
import 'package:bigfly_mobile/features/home/application/home_cubit.dart';
import 'package:bigfly_mobile/features/home/data/repositories/home_repository.dart';
import 'package:bigfly_mobile/features/more/application/more_cubit.dart';
import 'package:bigfly_mobile/features/more/data/repositories/more_repository.dart';
import 'package:bigfly_mobile/features/players/application/player_selection_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/data/repositories/player_repository.dart';
import 'package:bigfly_mobile/features/teams/application/teams_cubit.dart';
import 'package:bigfly_mobile/features/teams/data/repositories/team_repository.dart';
import 'package:dynamic_color/dynamic_color.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class BigFlyApp extends StatelessWidget {
  const BigFlyApp({
    super.key,
    required this.cacheStore,
    required this.homeRepository,
    required this.playerRepository,
    required this.teamRepository,
    required this.gameRepository,
    required this.moreRepository,
    this.useDynamicColor = true,
  });

  final CacheStore cacheStore;
  final HomeRepository homeRepository;
  final PlayerRepository playerRepository;
  final TeamRepository teamRepository;
  final GameRepository gameRepository;
  final MoreRepository moreRepository;
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
        BlocProvider<TeamsCubit>(create: (_) => TeamsCubit(teamRepository)),
        BlocProvider<GamesCubit>(create: (_) => GamesCubit(gameRepository)),
        BlocProvider<MoreCubit>(create: (_) => MoreCubit(moreRepository)),
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
      home: const RootShell(),
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
