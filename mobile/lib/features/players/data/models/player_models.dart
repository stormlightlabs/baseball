import 'package:bigfly_mobile/core/data/json/json_helpers.dart';

class PlayerSearchResult {
  const PlayerSearchResult({required this.id, required this.name, required this.subtitle});

  final String id;
  final String name;
  final String subtitle;

  factory PlayerSearchResult.fromJson(Map<String, dynamic> json) {
    final player = PlayerProfile.fromJson(json);
    return PlayerSearchResult(id: player.id, name: player.fullName, subtitle: player.subtitle);
  }
}

class PlayerProfile {
  const PlayerProfile({
    required this.id,
    required this.firstName,
    required this.lastName,
    required this.birthYear,
    required this.birthMonth,
    required this.birthDay,
    required this.birthCity,
    required this.birthState,
    required this.bats,
    required this.throwsHand,
    required this.debut,
    required this.finalGame,
    required this.latestSeason,
    required this.latestTeam,
    required this.positions,
  });

  final String id;
  final String firstName;
  final String lastName;
  final int? birthYear;
  final int? birthMonth;
  final int? birthDay;
  final String? birthCity;
  final String? birthState;
  final String? bats;
  final String? throwsHand;
  final DateTime? debut;
  final DateTime? finalGame;
  final int? latestSeason;
  final String? latestTeam;
  final String? positions;

  factory PlayerProfile.fromJson(Map<String, dynamic> json) {
    return PlayerProfile(
      id: stringOrEmpty(json['id']),
      firstName: stringOrEmpty(json['first_name']),
      lastName: stringOrEmpty(json['last_name']),
      birthYear: nullableInt(json['birth_year']),
      birthMonth: nullableInt(json['birth_month']),
      birthDay: nullableInt(json['birth_day']),
      birthCity: nullableString(json['birth_city']),
      birthState: nullableString(json['birth_state']),
      bats: nullableString(json['bats']),
      throwsHand: nullableString(json['throws']),
      debut: DateTime.tryParse(stringOrEmpty(json['debut'])),
      finalGame: DateTime.tryParse(stringOrEmpty(json['final_game'])),
      latestSeason: nullableInt(json['latest_season']),
      latestTeam: nullableString(json['latest_team']),
      positions: nullableString(json['positions']),
    );
  }

  String get fullName {
    final combined = '$firstName $lastName'.trim();
    return combined.isEmpty ? id : combined;
  }

  String get initials {
    final a = firstName.isNotEmpty ? firstName[0] : '?';
    final b = lastName.isNotEmpty ? lastName[0] : '?';
    return '$a$b'.toUpperCase();
  }

  String get subtitle {
    final pieces = <String>[
      id,
      if (positions != null && positions!.isNotEmpty) positions!,
      if (debut != null || finalGame != null) '${debut?.year ?? '—'}–${finalGame?.year ?? latestSeason ?? '—'}',
    ];
    return pieces.join(' · ');
  }

  String get birthLine {
    final dateParts = <String>[
      if (birthMonth != null) birthMonth.toString().padLeft(2, '0'),
      if (birthDay != null) birthDay.toString().padLeft(2, '0'),
      if (birthYear != null) birthYear.toString(),
    ];
    final placeParts = <String?>[birthCity, birthState].whereType<String>().toList(growable: false);
    final date = dateParts.isEmpty ? null : dateParts.join('/');
    final place = placeParts.isEmpty ? null : placeParts.join(', ');
    if (date == null && place == null) {
      return 'Birth unknown';
    }
    if (date != null && place != null) {
      return '$date · $place';
    }
    return date ?? place!;
  }
}

class PlayerBattingSeason {
  const PlayerBattingSeason({
    required this.year,
    required this.teamId,
    required this.g,
    required this.ab,
    required this.avg,
    required this.hr,
    required this.rbi,
    required this.obp,
    required this.slg,
    required this.ops,
    required this.hits,
  });

  final int year;
  final String teamId;
  final int g;
  final int ab;
  final double avg;
  final int hr;
  final int rbi;
  final double obp;
  final double slg;
  final double ops;
  final int hits;

  factory PlayerBattingSeason.fromJson(Map<String, dynamic> json) {
    return PlayerBattingSeason(
      year: intOrZero(json['year']),
      teamId: stringOrEmpty(json['team_id']),
      g: intOrZero(json['g']),
      ab: intOrZero(json['ab']),
      avg: doubleOrZero(json['avg']),
      hr: intOrZero(json['hr']),
      rbi: intOrZero(json['rbi']),
      obp: doubleOrZero(json['obp']),
      slg: doubleOrZero(json['slg']),
      ops: doubleOrZero(json['ops']),
      hits: intOrZero(json['h']),
    );
  }
}

class PlayerPitchingSeason {
  const PlayerPitchingSeason({
    required this.year,
    required this.teamId,
    required this.wins,
    required this.losses,
    required this.games,
    required this.era,
    required this.so,
    required this.whip,
    required this.kPer9,
  });

  final int year;
  final String teamId;
  final int wins;
  final int losses;
  final int games;
  final double era;
  final int so;
  final double whip;
  final double kPer9;

  factory PlayerPitchingSeason.fromJson(Map<String, dynamic> json) {
    return PlayerPitchingSeason(
      year: intOrZero(json['year']),
      teamId: stringOrEmpty(json['team_id']),
      wins: intOrZero(json['w']),
      losses: intOrZero(json['l']),
      games: intOrZero(json['g']),
      era: doubleOrZero(json['era']),
      so: intOrZero(json['so']),
      whip: doubleOrZero(json['whip']),
      kPer9: doubleOrZero(json['k_per_9']),
    );
  }
}

class PlayerAward {
  const PlayerAward({required this.awardId, required this.year, required this.league});

  final String awardId;
  final int year;
  final String? league;

  factory PlayerAward.fromJson(Map<String, dynamic> json) {
    return PlayerAward(
      awardId: stringOrEmpty(json['award_id']),
      year: intOrZero(json['year']),
      league: nullableString(json['league']),
    );
  }
}

class PlayerHallOfFameRecord {
  const PlayerHallOfFameRecord({
    required this.year,
    required this.votedBy,
    required this.votes,
    required this.ballots,
    required this.inducted,
  });

  final int year;
  final String votedBy;
  final int? votes;
  final int? ballots;
  final bool inducted;

  factory PlayerHallOfFameRecord.fromJson(Map<String, dynamic> json) {
    return PlayerHallOfFameRecord(
      year: intOrZero(json['year']),
      votedBy: stringOrEmpty(json['voted_by']),
      votes: nullableInt(json['votes']),
      ballots: nullableInt(json['ballots']),
      inducted: boolOrFalse(json['inducted']),
    );
  }

  double? get votePercent {
    if (votes == null || ballots == null || ballots == 0) {
      return null;
    }
    return (votes! / ballots!) * 100;
  }
}

class TeamSeasonSnapshot {
  const TeamSeasonSnapshot({required this.teamId, required this.franchiseId});

  final String teamId;
  final String franchiseId;

  factory TeamSeasonSnapshot.fromJson(Map<String, dynamic> json) {
    return TeamSeasonSnapshot(teamId: stringOrEmpty(json['team_id']), franchiseId: stringOrEmpty(json['franchise_id']));
  }
}

class PlayerDetailBundle {
  const PlayerDetailBundle({
    required this.player,
    required this.battingSeasons,
    required this.pitchingSeasons,
    required this.awards,
    required this.hallOfFameRecords,
    required this.themeTeamCode,
  });

  final PlayerProfile player;
  final List<PlayerBattingSeason> battingSeasons;
  final List<PlayerPitchingSeason> pitchingSeasons;
  final List<PlayerAward> awards;
  final List<PlayerHallOfFameRecord> hallOfFameRecords;
  final String? themeTeamCode;
}
