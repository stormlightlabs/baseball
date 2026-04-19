import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:bigfly_mobile/features/home/presentation/constants/home_constants.dart';
import 'package:flutter/material.dart';

class HomeQuickAccessGrid extends StatelessWidget {
  const HomeQuickAccessGrid({super.key, required this.links, required this.onTapLink});

  final List<HomeQuickLink> links;
  final ValueChanged<HomeQuickLink> onTapLink;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: links.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: 1,
        mainAxisSpacing: 1,
        childAspectRatio: 1.85,
      ),
      itemBuilder: (context, index) {
        final link = links[index];
        return InkWell(
          onTap: () => onTapLink(link),
          child: Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Icon(link.icon, size: 20),
                const SizedBox(height: 6),
                Text(link.title, style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 2),
                Text(link.subtitle, style: Theme.of(context).extension<AppTypography>()?.code),
              ],
            ),
          ),
        );
      },
    );
  }
}
