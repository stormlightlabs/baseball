import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:flutter/material.dart';

class PanelCard extends StatelessWidget {
  const PanelCard({super.key, required this.title, required this.child, this.padding = const EdgeInsets.all(14)});

  final String title;
  final Widget child;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: padding,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title, style: Theme.of(context).extension<AppTypography>()?.code),
            const SizedBox(height: 10),
            child,
          ],
        ),
      ),
    );
  }
}
