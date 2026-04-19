import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/presentation/utils/player_detail_formatters.dart';
import 'package:flutter/material.dart';

class PlayerDetailTabChips extends StatelessWidget {
  const PlayerDetailTabChips({super.key, required this.selectedTab, required this.onTabSelected});

  final PlayerDetailTab selectedTab;
  final ValueChanged<PlayerDetailTab> onTabSelected;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 36,
      child: ListView(
        scrollDirection: Axis.horizontal,
        children: PlayerDetailTab.values
            .map(
              (tab) => Padding(
                padding: const EdgeInsets.only(right: 6),
                child: ChoiceChip(
                  label: Text(playerDetailTabLabel(tab)),
                  selected: selectedTab == tab,
                  onSelected: (_) => onTabSelected(tab),
                ),
              ),
            )
            .toList(growable: false),
      ),
    );
  }
}
