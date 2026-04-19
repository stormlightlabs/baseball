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
        crossAxisSpacing: 12,
        mainAxisSpacing: 12,
        childAspectRatio: 1.55,
      ),
      itemBuilder: (context, index) {
        final link = links[index];
        return InkWell(
          onTap: () => onTapLink(link),
          child: Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                Icon(link.icon, size: 20),
                const SizedBox(height: 6),
                Text(link.title, style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 2),
                Text(
                  link.subtitle,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).extension<AppTypography>()?.code,
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
