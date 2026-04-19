import 'package:bigfly_mobile/features/players/application/players_cubit.dart';
import 'package:bigfly_mobile/features/players/application/players_state.dart';
import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/awards_section.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/batting_section.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/hall_of_fame_section.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/pitching_section.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_detail_tab_chips.dart';
import 'package:bigfly_mobile/features/players/presentation/widgets/player_profile_card.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class PlayerDetailView extends StatelessWidget {
  const PlayerDetailView({super.key, required this.detail, required this.state});

  final PlayerDetailBundle detail;
  final PlayersState state;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        PlayerProfileCard(detail: detail),
        const SizedBox(height: 12),
        PlayerDetailTabChips(selectedTab: state.detailTab, onTabSelected: context.read<PlayersCubit>().setDetailTab),
        const SizedBox(height: 12),
        switch (state.detailTab) {
          PlayerDetailTab.batting => BattingSection(detail: detail, state: state),
          PlayerDetailTab.pitching => PitchingSection(detail: detail, state: state),
          PlayerDetailTab.awards => AwardsSection(awards: detail.awards),
          PlayerDetailTab.hallOfFame => HallOfFameSection(records: detail.hallOfFameRecords),
        },
      ],
    );
  }
}
