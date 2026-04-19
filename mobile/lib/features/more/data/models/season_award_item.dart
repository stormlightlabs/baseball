import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class SeasonAwardItem {
  const SeasonAwardItem({required this.awardId, required this.playerId, required this.year, this.league});

  final String awardId;
  final String playerId;
  final int year;
  final String? league;

  factory SeasonAwardItem.fromJson(Map<String, dynamic> json) {
    return SeasonAwardItem(
      awardId: stringOrEmpty(json['award_id']),
      playerId: stringOrEmpty(json['player_id']),
      year: intOrZero(json['year']),
      league: nullableString(json['league']),
    );
  }
}
