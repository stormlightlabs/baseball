import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class CompareBattingCareer {
  const CompareBattingCareer({
    required this.hr,
    required this.avg,
    required this.ops,
    required this.rbi,
    required this.sb,
    required this.hits,
    required this.ab,
  });

  final int hr;
  final double avg;
  final double ops;
  final int rbi;
  final int sb;
  final int hits;
  final int ab;

  factory CompareBattingCareer.fromJson(Map<String, dynamic> json) {
    final career = asJsonMap(json['career']);
    return CompareBattingCareer(
      hr: intOrZero(career['hr']),
      avg: doubleOrZero(career['avg']),
      ops: doubleOrZero(career['ops']),
      rbi: intOrZero(career['rbi']),
      sb: intOrZero(career['sb']),
      hits: intOrZero(career['h']),
      ab: intOrZero(career['ab']),
    );
  }
}
