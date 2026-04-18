import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/features/home/health_cubit.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class HomeTab extends StatelessWidget {
  const HomeTab({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: <Widget>[
        Text('Big Fly Mobile', style: Theme.of(context).textTheme.headlineMedium),
        const SizedBox(height: 8),
        Text('Explore baseball data across players, teams, and games.', style: Theme.of(context).textTheme.bodyMedium),
        const SizedBox(height: 16),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: BlocBuilder<HealthCubit, HealthState>(
              builder: (context, state) {
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text('API Health', style: Theme.of(context).textTheme.titleLarge),
                    const SizedBox(height: 8),
                    Text(
                      _statusLabel(state),
                      style:
                          Theme.of(context).extension<AppTypography>()?.code ?? Theme.of(context).textTheme.bodyMedium,
                    ),
                    if (state.fromCache) ...<Widget>[const SizedBox(height: 4), const Text('Source: cache')],
                    if (state.errorMessage != null) ...<Widget>[
                      const SizedBox(height: 8),
                      Text(
                        'Last error: ${state.errorMessage}',
                        style: Theme.of(
                          context,
                        ).textTheme.bodySmall?.copyWith(color: Theme.of(context).colorScheme.error),
                      ),
                    ],
                    const SizedBox(height: 12),
                    FilledButton.icon(
                      onPressed: state.loadStatus == HealthLoadStatus.loading
                          ? null
                          : () => context.read<HealthCubit>().loadHealth(),
                      icon: const Icon(Icons.refresh),
                      label: const Text('Refresh health'),
                    ),
                  ],
                );
              },
            ),
          ),
        ),
      ],
    );
  }

  String _statusLabel(HealthState state) {
    switch (state.loadStatus) {
      case HealthLoadStatus.initial:
        return 'Status unavailable';
      case HealthLoadStatus.loading:
        return 'Loading /api/v1/health...';
      case HealthLoadStatus.success:
        return 'Status: ${state.statusText ?? 'unknown'}';
      case HealthLoadStatus.failure:
        return 'Request failed';
    }
  }
}
