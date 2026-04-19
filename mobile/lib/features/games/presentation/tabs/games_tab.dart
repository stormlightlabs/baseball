import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/app/ui/feedback/inline_error_text.dart';
import 'package:bigfly_mobile/app/ui/feedback/loading_strip.dart';
import 'package:bigfly_mobile/features/games/application/games_cubit.dart';
import 'package:bigfly_mobile/features/games/application/games_state.dart';
import 'package:bigfly_mobile/features/games/application/games_types.dart';
import 'package:bigfly_mobile/features/games/data/models/game_models.dart';
import 'package:bigfly_mobile/features/games/presentation/utils/games_formatters.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class GamesTab extends StatefulWidget {
  const GamesTab({super.key});

  @override
  State<GamesTab> createState() => _GamesTabState();
}

class _GamesTabState extends State<GamesTab> {
  @override
  void initState() {
    super.initState();
    context.read<GamesCubit>().initialize();
  }

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<GamesCubit, GamesState>(
      builder: (context, state) {
        return ListView(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
          children: <Widget>[
            _FilterStrip(state: state),
            const SizedBox(height: 8),
            if (state.error != null) ...<Widget>[InlineErrorText(state.error!), const SizedBox(height: 8)],
            if (state.isLoading) ...<Widget>[const LoadingStrip(), const SizedBox(height: 8)],
            _ResultsHeader(state: state),
            const SizedBox(height: 8),
            if (state.visibleGames.isEmpty)
              const _EmptyState()
            else
              ...state.visibleGames.map((game) {
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: _GameCard(game: game, state: state),
                );
              }),
          ],
        );
      },
    );
  }
}

class _FilterStrip extends StatelessWidget {
  const _FilterStrip({required this.state});

  final GamesState state;

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<GamesCubit>();
    final colors = Theme.of(context).colorScheme;

    final hasSelectedSeason = state.availableSeasons.contains(state.selectedSeason);
    final selectedSeason = hasSelectedSeason ? state.selectedSeason : null;

    final availableTeamIds = state.availableTeams.map((team) => team.teamId).toSet();
    final selectedTeamId = availableTeamIds.contains(state.selectedTeamId) ? state.selectedTeamId : '';

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text('Game Finder', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 10),
            Row(
              children: <Widget>[
                Expanded(
                  child: DropdownButtonFormField<int>(
                    initialValue: selectedSeason,
                    isDense: true,
                    decoration: const InputDecoration(labelText: 'Season', border: OutlineInputBorder()),
                    items: state.availableSeasons
                        .map((season) => DropdownMenuItem<int>(value: season, child: Text(season.toString())))
                        .toList(growable: false),
                    onChanged: (value) {
                      if (value != null) {
                        cubit.selectSeason(value);
                      }
                    },
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: DropdownButtonFormField<String>(
                    initialValue: selectedTeamId,
                    isDense: true,
                    decoration: const InputDecoration(labelText: 'Team', border: OutlineInputBorder()),
                    items: <DropdownMenuItem<String>>[
                      const DropdownMenuItem<String>(value: '', child: Text('Any team')),
                      ...state.availableTeams.map(
                        (team) => DropdownMenuItem<String>(value: team.teamId, child: Text(team.teamId)),
                      ),
                    ],
                    onChanged: (value) => cubit.selectTeam(value ?? ''),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: GamesQuickFilter.values
                    .map((filter) {
                      return Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: FilterChip(
                          label: Text(gamesQuickFilterLabel(filter)),
                          selected: state.quickFilters.contains(filter),
                          showCheckmark: false,
                          selectedColor: colors.primary.withValues(alpha: 0.18),
                          onSelected: (_) => cubit.toggleQuickFilter(filter),
                        ),
                      );
                    })
                    .toList(growable: false),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ResultsHeader extends StatelessWidget {
  const _ResultsHeader({required this.state});

  final GamesState state;

  @override
  Widget build(BuildContext context) {
    final codeStyle = Theme.of(context).extension<AppTypography>()?.code;

    return Row(
      children: <Widget>[
        Text(
          '${state.visibleGames.length} games${state.selectedSeason > 0 ? ' · ${state.selectedSeason}' : ''}',
          style: codeStyle,
        ),
        const Spacer(),
        Text('Sort: Date ↓', style: Theme.of(context).textTheme.labelMedium),
      ],
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Text('No games matched these filters.', style: Theme.of(context).textTheme.bodyMedium),
      ),
    );
  }
}

class _GameCard extends StatelessWidget {
  const _GameCard({required this.game, required this.state});

  final GameSummaryRecord game;
  final GamesState state;

  @override
  Widget build(BuildContext context) {
    final cubit = context.read<GamesCubit>();
    final colors = Theme.of(context).colorScheme;
    final isExpanded = state.expandedGameId == game.id;

    final detail = state.gameDetails[game.id];
    final isLoadingDetail = state.detailLoadingGameIds.contains(game.id);
    final detailError = state.detailErrors[game.id];

    final metaParts = <String>[
      if (game.parkName != null && game.parkName!.isNotEmpty) game.parkName!,
      if (game.attendance != null && game.attendance! > 0) '${formatAttendance(game.attendance)} att',
      if (game.isLikelyPostseason) 'Postseason',
    ];

    return AnimatedContainer(
      duration: const Duration(milliseconds: 180),
      decoration: BoxDecoration(
        color: colors.surfaceContainerLow,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: isExpanded ? colors.primary : colors.outlineVariant.withValues(alpha: 0.7)),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => cubit.toggleExpandedGame(game),
        child: Column(
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
              child: Row(
                children: <Widget>[
                  SizedBox(
                    width: 56,
                    child: Text(formatGamesCompactDate(game.date), style: Theme.of(context).textTheme.labelSmall),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          '${game.awayTeam} @ ${game.homeTeam}',
                          style: Theme.of(context).textTheme.titleSmall,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 2),
                        Text(
                          metaParts.isEmpty ? '—' : metaParts.join(' · '),
                          style: Theme.of(context).textTheme.labelSmall,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 10),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: <Widget>[
                      Text('${game.awayScore}–${game.homeScore}', style: Theme.of(context).textTheme.titleMedium),
                      Text('${game.innings} inn', style: Theme.of(context).textTheme.labelSmall),
                    ],
                  ),
                ],
              ),
            ),
            if (isExpanded) ...<Widget>[
              Divider(height: 1, color: colors.outlineVariant.withValues(alpha: 0.7)),
              if (isLoadingDetail)
                const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  child: LoadingStrip(minHeight: 3),
                )
              else if (detailError != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  child: InlineErrorText(detailError),
                )
              else if (detail != null)
                Padding(
                  padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
                  child: _ExpandedGameDetail(game: game, detail: detail),
                ),
            ],
          ],
        ),
      ),
    );
  }
}

class _ExpandedGameDetail extends StatelessWidget {
  const _ExpandedGameDetail({required this.game, required this.detail});

  final GameSummaryRecord game;
  final GameCardDetail detail;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: <Widget>[
            _TeamScore(teamId: game.awayTeam, score: game.awayScore),
            Text('—', style: Theme.of(context).textTheme.titleLarge),
            _TeamScore(teamId: game.homeTeam, score: game.homeScore),
          ],
        ),
        const SizedBox(height: 10),
        GridView.count(
          crossAxisCount: 2,
          childAspectRatio: 3.1,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          mainAxisSpacing: 6,
          crossAxisSpacing: 6,
          children: <Widget>[
            _MetaTile(label: 'Date', value: formatGamesLongDate(game.date)),
            _MetaTile(label: 'Duration', value: formatDuration(game.durationMin)),
            _MetaTile(label: 'Park', value: game.parkName ?? '—'),
            _MetaTile(label: 'Attendance', value: formatAttendance(game.attendance)),
          ],
        ),
        const SizedBox(height: 10),
        Text('Final win probability', style: Theme.of(context).textTheme.labelSmall),
        const SizedBox(height: 6),
        _WinProbabilityBar(
          awayTeam: game.awayTeam,
          homeTeam: game.homeTeam,
          awayWinProbability: detail.awayWinProbability,
          homeWinProbability: detail.homeWinProbability,
        ),
        const SizedBox(height: 10),
        Text('Key plays', style: Theme.of(context).textTheme.labelSmall),
        const SizedBox(height: 6),
        if (detail.keyPlays.isEmpty)
          Text('No key plays available.', style: Theme.of(context).textTheme.bodySmall)
        else
          ...detail.keyPlays.map((play) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  SizedBox(width: 32, child: Text(play.inningLabel, style: Theme.of(context).textTheme.labelSmall)),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      play.event,
                      style: Theme.of(context).textTheme.bodySmall,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(play.scoreLabel, style: Theme.of(context).textTheme.labelSmall),
                ],
              ),
            );
          }),
      ],
    );
  }
}

class _TeamScore extends StatelessWidget {
  const _TeamScore({required this.teamId, required this.score});

  final String teamId;
  final int score;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Text(teamId, style: Theme.of(context).textTheme.labelMedium),
        Text(score.toString(), style: Theme.of(context).textTheme.headlineSmall),
      ],
    );
  }
}

class _MetaTile extends StatelessWidget {
  const _MetaTile({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.surfaceContainerHighest.withValues(alpha: 0.7),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Text(label, style: Theme.of(context).textTheme.labelSmall),
            Text(value, maxLines: 1, overflow: TextOverflow.ellipsis, style: Theme.of(context).textTheme.bodyMedium),
          ],
        ),
      ),
    );
  }
}

class _WinProbabilityBar extends StatelessWidget {
  const _WinProbabilityBar({
    required this.awayTeam,
    required this.homeTeam,
    required this.awayWinProbability,
    required this.homeWinProbability,
  });

  final String awayTeam;
  final String homeTeam;
  final double awayWinProbability;
  final double homeWinProbability;

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return SizedBox(
      height: 32,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(6),
        child: Stack(
          fit: StackFit.expand,
          children: <Widget>[
            ColoredBox(color: colors.surfaceContainerHighest),
            Align(
              alignment: Alignment.centerLeft,
              child: FractionallySizedBox(
                widthFactor: awayWinProbability.clamp(0.0, 1.0),
                child: ColoredBox(color: colors.primary.withValues(alpha: 0.55)),
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Row(
                children: <Widget>[
                  Text(
                    '$awayTeam ${formatProbabilityPercent(awayWinProbability)}',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(color: colors.onPrimaryContainer),
                  ),
                  const Spacer(),
                  Text(
                    '$homeTeam ${formatProbabilityPercent(homeWinProbability)}',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(color: colors.onSurfaceVariant),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
