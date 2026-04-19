import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class PostseasonSeriesRecord {
  const PostseasonSeriesRecord({
    required this.round,
    required this.winnerTeam,
    required this.loserTeam,
    required this.wins,
    required this.losses,
    required this.ties,
  });

  final String round;
  final String? winnerTeam;
  final String? loserTeam;
  final int? wins;
  final int? losses;
  final int? ties;

  factory PostseasonSeriesRecord.fromJson(Map<String, dynamic> json) {
    return PostseasonSeriesRecord(
      round: stringOrEmpty(json['round']),
      winnerTeam: nullableString(json['winner_team']),
      loserTeam: nullableString(json['loser_team']),
      wins: nullableInt(json['wins']),
      losses: nullableInt(json['losses']),
      ties: nullableInt(json['ties']),
    );
  }

  String get matchup {
    final winner = winnerTeam ?? 'TBD';
    final loser = loserTeam ?? 'TBD';
    return '$winner def. $loser';
  }

  String get result {
    if (wins == null || losses == null) {
      return '—';
    }
    return '${wins!}-${losses!}${(ties != null && ties! > 0) ? '-${ties!}' : ''}';
  }
}
