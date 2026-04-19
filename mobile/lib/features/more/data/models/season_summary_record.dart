import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class SeasonSummaryRecord {
  const SeasonSummaryRecord({required this.year, required this.leagues, required this.teamCount, this.gameCount});

  final int year;
  final List<String> leagues;
  final int teamCount;
  final int? gameCount;

  factory SeasonSummaryRecord.fromJson(Map<String, dynamic> json) {
    final leaguesRaw = json['leagues'];
    final leagues = leaguesRaw is List
        ? leaguesRaw.map((item) => stringOrEmpty(item)).where((item) => item.isNotEmpty).toList(growable: false)
        : const <String>[];

    return SeasonSummaryRecord(
      year: intOrZero(json['year']),
      leagues: leagues,
      teamCount: intOrZero(json['team_count']),
      gameCount: nullableInt(json['game_count']),
    );
  }
}
