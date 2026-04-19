import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class ComparePitchingCareer {
  const ComparePitchingCareer({
    required this.wins,
    required this.losses,
    required this.era,
    required this.strikeouts,
    required this.whip,
    required this.kPer9,
    required this.ipOuts,
  });

  final int wins;
  final int losses;
  final double era;
  final int strikeouts;
  final double whip;
  final double kPer9;
  final int ipOuts;

  factory ComparePitchingCareer.fromJson(Map<String, dynamic> json) {
    final career = asJsonMap(json['career']);
    return ComparePitchingCareer(
      wins: intOrZero(career['w']),
      losses: intOrZero(career['l']),
      era: doubleOrZero(career['era']),
      strikeouts: intOrZero(career['so']),
      whip: doubleOrZero(career['whip']),
      kPer9: doubleOrZero(career['k_per_9']),
      ipOuts: intOrZero(career['ip_outs']),
    );
  }

  double get inningsPitched => ipOuts / 3.0;
}
