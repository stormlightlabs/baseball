import 'package:bigfly_mobile/features/more/data/models/compare_batting_career.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_pitching_career.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';

class ComparePlayerSnapshot {
  const ComparePlayerSnapshot({required this.profile, required this.battingCareer, required this.pitchingCareer});

  final PlayerProfile profile;
  final CompareBattingCareer battingCareer;
  final ComparePitchingCareer pitchingCareer;

  int? get debutYear => profile.debut?.year;

  int? get finalYear => profile.finalGame?.year ?? profile.latestSeason;
}
