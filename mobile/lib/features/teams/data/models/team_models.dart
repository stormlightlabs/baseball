import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class TeamSeasonRecord {
  const TeamSeasonRecord({
    required this.teamId,
    required this.year,
    required this.franchiseId,
    required this.league,
    required this.name,
    required this.parkId,
    required this.games,
    required this.wins,
    required this.losses,
    required this.ties,
    required this.runsScored,
    required this.runsAllowed,
    this.division,
  });

  final String teamId;
  final int year;
  final String franchiseId;
  final String league;
  final String name;
  final String parkId;
  final int games;
  final int wins;
  final int losses;
  final int ties;
  final int runsScored;
  final int runsAllowed;
  final String? division;

  factory TeamSeasonRecord.fromJson(Map<String, dynamic> json) {
    return TeamSeasonRecord(
      teamId: stringOrEmpty(json['team_id']),
      year: intOrZero(json['year']),
      franchiseId: stringOrEmpty(json['franchise_id']),
      league: stringOrEmpty(json['league']),
      name: stringOrEmpty(json['name']),
      parkId: stringOrEmpty(json['park_id']),
      games: intOrZero(json['games']),
      wins: intOrZero(json['wins']),
      losses: intOrZero(json['losses']),
      ties: intOrZero(json['ties']),
      runsScored: intOrZero(json['runs_scored']),
      runsAllowed: intOrZero(json['runs_allowed']),
      division: nullableString(json['division']),
    );
  }

  double get winPct {
    final totalDecisions = wins + losses;
    if (totalDecisions <= 0) {
      return 0;
    }
    return wins / totalDecisions;
  }

  String get subtitle {
    final parts = <String>[
      teamId,
      if (league.isNotEmpty) league,
      if (division != null && division!.isNotEmpty) division!,
      year.toString(),
    ];
    return parts.join(' · ');
  }
}

class FranchiseSummary {
  const FranchiseSummary({required this.id, required this.name, required this.active, this.activeFrom, this.activeTo});

  final String id;
  final String name;
  final bool active;
  final int? activeFrom;
  final int? activeTo;

  factory FranchiseSummary.fromJson(Map<String, dynamic> json) {
    return FranchiseSummary(
      id: stringOrEmpty(json['id']),
      name: stringOrEmpty(json['name']),
      active: boolOrFalse(json['active']),
      activeFrom: nullableInt(json['active_from']),
      activeTo: nullableInt(json['active_to']),
    );
  }

  String get activeRange {
    final from = activeFrom?.toString() ?? '—';
    final to = activeTo?.toString() ?? (active ? 'present' : '—');
    return '$from–$to';
  }
}

class TeamRosterPlayer {
  const TeamRosterPlayer({
    required this.playerId,
    required this.firstName,
    required this.lastName,
    this.position,
    this.battingG,
    this.ab,
    this.h,
    this.hr,
    this.rbi,
    this.avg,
    this.pitchingG,
    this.w,
    this.l,
    this.era,
    this.so,
  });

  final String playerId;
  final String firstName;
  final String lastName;
  final String? position;
  final int? battingG;
  final int? ab;
  final int? h;
  final int? hr;
  final int? rbi;
  final double? avg;
  final int? pitchingG;
  final int? w;
  final int? l;
  final double? era;
  final int? so;

  factory TeamRosterPlayer.fromJson(Map<String, dynamic> json) {
    return TeamRosterPlayer(
      playerId: stringOrEmpty(json['player_id']),
      firstName: stringOrEmpty(json['first_name']),
      lastName: stringOrEmpty(json['last_name']),
      position: nullableString(json['position']),
      battingG: nullableInt(json['batting_g']),
      ab: nullableInt(json['ab']),
      h: nullableInt(json['h']),
      hr: nullableInt(json['hr']),
      rbi: nullableInt(json['rbi']),
      avg: nullableDouble(json['avg']),
      pitchingG: nullableInt(json['pitching_g']),
      w: nullableInt(json['w']),
      l: nullableInt(json['l']),
      era: nullableDouble(json['era']),
      so: nullableInt(json['so']),
    );
  }

  String get fullName {
    final value = '$firstName $lastName'.trim();
    return value.isEmpty ? playerId : value;
  }

  String get initials {
    final first = firstName.isNotEmpty ? firstName[0] : '?';
    final last = lastName.isNotEmpty ? lastName[0] : '?';
    return '$first$last'.toUpperCase();
  }

  String get statLine {
    if (avg != null || hr != null || rbi != null) {
      final avgText = avg == null ? '—' : avg!.toStringAsFixed(3).replaceFirst('0', '');
      return 'AVG $avgText · HR ${hr ?? 0} · RBI ${rbi ?? 0}';
    }

    if (era != null || so != null || w != null || l != null) {
      final eraText = era == null ? '—' : era!.toStringAsFixed(2);
      return 'W-L ${w ?? 0}-${l ?? 0} · ERA $eraText · SO ${so ?? 0}';
    }

    return playerId;
  }
}

class TeamGameSummary {
  const TeamGameSummary({
    required this.id,
    required this.season,
    required this.date,
    required this.homeTeam,
    required this.awayTeam,
    required this.homeScore,
    required this.awayScore,
    required this.parkName,
  });

  final String id;
  final int season;
  final DateTime? date;
  final String homeTeam;
  final String awayTeam;
  final int homeScore;
  final int awayScore;
  final String? parkName;

  factory TeamGameSummary.fromJson(Map<String, dynamic> json) {
    return TeamGameSummary(
      id: stringOrEmpty(json['id']),
      season: intOrZero(json['season']),
      date: DateTime.tryParse(stringOrEmpty(json['date'])),
      homeTeam: stringOrEmpty(json['home_team']),
      awayTeam: stringOrEmpty(json['away_team']),
      homeScore: intOrZero(json['home_score']),
      awayScore: intOrZero(json['away_score']),
      parkName: nullableString(json['park_name']),
    );
  }

  bool isTeamHome(String teamId) => homeTeam == teamId;

  bool hasDecisionFor(String teamId) {
    if (teamId != homeTeam && teamId != awayTeam) {
      return false;
    }

    return homeScore != awayScore;
  }

  bool didTeamWin(String teamId) {
    if (!hasDecisionFor(teamId)) {
      return false;
    }

    if (isTeamHome(teamId)) {
      return homeScore > awayScore;
    }
    return awayScore > homeScore;
  }
}

class TeamDailyLog {
  const TeamDailyLog({
    required this.date,
    required this.gamesPlayed,
    required this.wins,
    required this.losses,
    required this.runsScored,
    required this.runsAllowed,
    required this.runDiff,
  });

  final DateTime? date;
  final int gamesPlayed;
  final int wins;
  final int losses;
  final int runsScored;
  final int runsAllowed;
  final int runDiff;

  factory TeamDailyLog.fromJson(Map<String, dynamic> json) {
    return TeamDailyLog(
      date: DateTime.tryParse(stringOrEmpty(json['date'])),
      gamesPlayed: intOrZero(json['games_played']),
      wins: intOrZero(json['wins']),
      losses: intOrZero(json['losses']),
      runsScored: intOrZero(json['runs_scored']),
      runsAllowed: intOrZero(json['runs_allowed']),
      runDiff: intOrZero(json['run_diff']),
    );
  }
}

class RunDifferentialGamePoint {
  const RunDifferentialGamePoint({required this.date, required this.differential, required this.cumulativeDiff});

  final DateTime? date;
  final int differential;
  final int cumulativeDiff;

  factory RunDifferentialGamePoint.fromJson(Map<String, dynamic> json) {
    return RunDifferentialGamePoint(
      date: DateTime.tryParse(stringOrEmpty(json['date'])),
      differential: intOrZero(json['differential']),
      cumulativeDiff: intOrZero(json['cumulative_diff']),
    );
  }
}

class TeamRunDifferentialSeries {
  const TeamRunDifferentialSeries({
    required this.entityId,
    required this.season,
    required this.gamesPlayed,
    required this.runsScored,
    required this.runsAllowed,
    required this.runDifferential,
    required this.games,
  });

  final String entityId;
  final int season;
  final int gamesPlayed;
  final int runsScored;
  final int runsAllowed;
  final int runDifferential;
  final List<RunDifferentialGamePoint> games;

  factory TeamRunDifferentialSeries.fromJson(Map<String, dynamic> json) {
    return TeamRunDifferentialSeries(
      entityId: stringOrEmpty(json['entity_id']),
      season: intOrZero(json['season']),
      gamesPlayed: intOrZero(json['games_played']),
      runsScored: intOrZero(json['runs_scored']),
      runsAllowed: intOrZero(json['runs_allowed']),
      runDifferential: intOrZero(json['run_differential']),
      games: asJsonMapList(json['games']).map(RunDifferentialGamePoint.fromJson).toList(growable: false),
    );
  }
}

class TeamDetailBundle {
  const TeamDetailBundle({
    required this.team,
    required this.franchise,
    required this.recentSeasons,
    required this.roster,
    required this.schedule,
    required this.dailyLogs,
    required this.runDifferential,
    required this.themeTeamCode,
  });

  final TeamSeasonRecord team;
  final FranchiseSummary? franchise;
  final List<TeamSeasonRecord> recentSeasons;
  final List<TeamRosterPlayer> roster;
  final List<TeamGameSummary> schedule;
  final List<TeamDailyLog> dailyLogs;
  final TeamRunDifferentialSeries? runDifferential;
  final String? themeTeamCode;
}
