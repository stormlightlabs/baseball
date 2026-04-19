import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/features/home/presentation/constants/home_constants.dart';
import 'package:flutter/material.dart';

class HomeFeaturedQueries extends StatelessWidget {
  const HomeFeaturedQueries({super.key, required this.queries});

  final List<HomeFeaturedQuery> queries;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: queries
          .map(
            (query) => Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text(query.name),
                subtitle: Text(query.endpoint, style: Theme.of(context).extension<AppTypography>()?.code),
              ),
            ),
          )
          .toList(growable: false),
    );
  }
}
