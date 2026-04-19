import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:flutter/material.dart';

class SectionPicker extends StatelessWidget {
  const SectionPicker({super.key, required this.selected, required this.onSelect});

  final MoreSection selected;
  final ValueChanged<MoreSection> onSelect;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        border: Border(bottom: BorderSide(color: Theme.of(context).colorScheme.outlineVariant.withValues(alpha: 0.6))),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: MoreSection.values
              .map((section) {
                final label = switch (section) {
                  MoreSection.seasons => 'Seasons',
                  MoreSection.leaders => 'Leaders',
                  MoreSection.compare => 'Compare',
                  MoreSection.dataSources => 'Data Sources',
                };

                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: ChoiceChip(
                    label: Text(label),
                    selected: selected == section,
                    onSelected: (_) => onSelect(section),
                  ),
                );
              })
              .toList(growable: false),
        ),
      ),
    );
  }
}
