import 'package:bigfly_mobile/features/home/presentation/constants/home_constants.dart';
import 'package:flutter/material.dart';

class HomeEraChips extends StatelessWidget {
  const HomeEraChips({super.key, required this.chips, required this.onTapChip});

  final List<HomeEraChip> chips;
  final ValueChanged<HomeEraChip> onTapChip;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 6,
      runSpacing: 6,
      children: chips
          .map((chip) => ActionChip(label: Text(chip.label), onPressed: () => onTapChip(chip)))
          .toList(growable: false),
    );
  }
}
