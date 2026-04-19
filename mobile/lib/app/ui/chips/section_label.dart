import 'package:bigfly_mobile/app/theme/app_typography.dart';
import 'package:flutter/material.dart';

class SectionLabel extends StatelessWidget {
  const SectionLabel(this.text, {super.key});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Text(text, style: Theme.of(context).extension<AppTypography>()?.code);
  }
}
