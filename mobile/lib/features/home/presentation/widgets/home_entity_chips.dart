import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:flutter/material.dart';

class HomeEntityChips extends StatelessWidget {
  const HomeEntityChips({super.key, required this.activeEntity, required this.onSelectEntity});

  final HomeEntityType activeEntity;
  final ValueChanged<HomeEntityType> onSelectEntity;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 36,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: HomeEntityType.values
            .map(
              (entity) => Padding(
                padding: const EdgeInsets.only(right: 6),
                child: ChoiceChip(
                  label: Text(entity.label),
                  selected: activeEntity == entity,
                  onSelected: (_) => onSelectEntity(entity),
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}
