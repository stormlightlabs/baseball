import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class GameTeamOption {
  const GameTeamOption({required this.teamId, required this.name});

  final String teamId;
  final String name;

  factory GameTeamOption.fromJson(Map<String, dynamic> json) {
    return GameTeamOption(teamId: stringOrEmpty(json['team_id']), name: stringOrEmpty(json['name']));
  }
}

class GameSummaryRecord {
  const GameSummaryRecord({
    required this.id,
    required this.season,
    required this.date,
    required this.awayTeam,
    required this.homeTeam,
    required this.awayScore,
    required this.homeScore,
    required this.innings,
    required this.attendance,
    required this.durationMin,
    required this.parkName,
    required this.isPostseason,
  });

  final String id;
  final int season;
  final DateTime? date;
  final String awayTeam;
  final String homeTeam;
  final int awayScore;
  final int homeScore;
  final int innings;
  final int? attendance;
  final int? durationMin;
  final String? parkName;
  final bool isPostseason;

  factory GameSummaryRecord.fromJson(Map<String, dynamic> json) {
    return GameSummaryRecord(
      id: stringOrEmpty(json['id']),
      season: intOrZero(json['season']),
      date: DateTime.tryParse(stringOrEmpty(json['date'])),
      awayTeam: stringOrEmpty(json['away_team']),
      homeTeam: stringOrEmpty(json['home_team']),
      awayScore: intOrZero(json['away_score']),
      homeScore: intOrZero(json['home_score']),
      innings: intOrZero(json['innings']),
      attendance: nullableInt(json['attendance']),
      durationMin: nullableInt(json['duration_min']),
      parkName: nullableString(json['park_name']),
      isPostseason: boolOrFalse(json['is_postseason']),
    );
  }

  int get gameNumber => parseGameNumberFromId(id: id, homeTeam: homeTeam);

  bool get isDoubleheader => gameNumber > 0;

  bool get isLikelyPostseason {
    if (isPostseason) {
      return true;
    }
    final month = date?.month;
    return month != null && month >= 10 && gameNumber == 0;
  }
}

class GameListResponse {
  const GameListResponse({required this.games, required this.total});

  final List<GameSummaryRecord> games;
  final int total;
}

class GameWinProbabilitySummary {
  const GameWinProbabilitySummary({
    required this.gameId,
    required this.season,
    required this.homeTeam,
    required this.awayTeam,
    required this.homeWinProbStart,
    required this.homeWinProbEnd,
  });

  final String gameId;
  final int season;
  final String homeTeam;
  final String awayTeam;
  final double homeWinProbStart;
  final double homeWinProbEnd;

  factory GameWinProbabilitySummary.fromJson(Map<String, dynamic> json) {
    return GameWinProbabilitySummary(
      gameId: stringOrEmpty(json['game_id']),
      season: intOrZero(json['season']),
      homeTeam: stringOrEmpty(json['home_team']),
      awayTeam: stringOrEmpty(json['away_team']),
      homeWinProbStart: doubleOrZero(json['home_win_prob_start']),
      homeWinProbEnd: doubleOrZero(json['home_win_prob_end']),
    );
  }
}

class GameEvent {
  const GameEvent({
    required this.playNum,
    required this.inning,
    required this.topBot,
    required this.scoreVis,
    required this.scoreHome,
    required this.event,
    required this.runs,
  });

  final int playNum;
  final int inning;
  final int topBot;
  final int scoreVis;
  final int scoreHome;
  final String event;
  final int? runs;

  factory GameEvent.fromJson(Map<String, dynamic> json) {
    return GameEvent(
      playNum: intOrZero(json['play_num']),
      inning: intOrZero(json['inning']),
      topBot: intOrZero(json['top_bot']),
      scoreVis: intOrZero(json['score_vis']),
      scoreHome: intOrZero(json['score_home']),
      event: stringOrEmpty(json['event']),
      runs: nullableInt(json['runs']),
    );
  }

  String get inningLabel {
    final half = topBot == 0 ? 'T' : 'B';
    return '$half$inning';
  }

  String get scoreLabel => '$scoreVis-$scoreHome';
}

class GameCardDetail {
  const GameCardDetail({required this.awayWinProbability, required this.homeWinProbability, required this.keyPlays});

  final double awayWinProbability;
  final double homeWinProbability;
  final List<GameEvent> keyPlays;
}

int parseGameNumberFromId({required String id, required String homeTeam}) {
  if (id.length <= 8) {
    return 0;
  }

  var end = id.length;
  if (homeTeam.isNotEmpty && id.endsWith(homeTeam)) {
    end = id.length - homeTeam.length;
  }

  if (end <= 8) {
    return 0;
  }

  final rawGameNumber = id.substring(8, end);
  return int.tryParse(rawGameNumber) ?? 0;
}
