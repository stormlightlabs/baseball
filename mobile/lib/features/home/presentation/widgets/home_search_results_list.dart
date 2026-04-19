import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:flutter/material.dart';

class HomeSearchResultsList extends StatelessWidget {
  const HomeSearchResultsList({super.key, required this.results, required this.onTapResult});

  final List<HomeSearchItem> results;
  final ValueChanged<HomeSearchItem> onTapResult;

  @override
  Widget build(BuildContext context) {
    if (results.isEmpty) {
      return const SizedBox.shrink();
    }

    return Column(
      children: results
          .map(
            (item) => Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: ListTile(
                title: Text(item.title),
                subtitle: Text(item.subtitle),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => onTapResult(item),
              ),
            ),
          )
          .toList(growable: false),
    );
  }
}
