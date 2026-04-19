import 'package:flutter/material.dart';

enum MlbTeam {
  ari('ARI', 'Arizona Diamondbacks', '#A71930'),
  atl('ATL', 'Atlanta Braves', '#CE1141'),
  bal('BAL', 'Baltimore Orioles', '#DF4601'),
  bos('BOS', 'Boston Red Sox', '#BD3039'),
  chc('CHC', 'Chicago Cubs', '#0E3386'),
  cws('CWS', 'Chicago White Sox', '#27251F'),
  cin('CIN', 'Cincinnati Reds', '#C6011F'),
  cle('CLE', 'Cleveland Guardians', '#E31937'),
  col('COL', 'Colorado Rockies', '#33006F'),
  det('DET', 'Detroit Tigers', '#0C2340'),
  hou('HOU', 'Houston Astros', '#002D62'),
  kc('KC', 'Kansas City Royals', '#004687'),
  laa('LAA', 'Los Angeles Angels', '#BA0021'),
  lad('LAD', 'Los Angeles Dodgers', '#005A9C'),
  mia('MIA', 'Miami Marlins', '#000000'),
  mil('MIL', 'Milwaukee Brewers', '#12284B'),
  min('MIN', 'Minnesota Twins', '#002B5C'),
  nym('NYM', 'New York Mets', '#002D72'),
  nyy('NYY', 'New York Yankees', '#132448'),
  ath('ATH', 'Athletics', '#003831'),
  phi('PHI', 'Philadelphia Phillies', '#E81828'),
  pit('PIT', 'Pittsburgh Pirates', '#27251F'),
  sd('SD', 'San Diego Padres', '#2F241D'),
  sf('SF', 'San Francisco Giants', '#FD5A1E'),
  sea('SEA', 'Seattle Mariners', '#0C2C56'),
  stl('STL', 'St. Louis Cardinals', '#C41E3A'),
  tb('TB', 'Tampa Bay Rays', '#092C5C'),
  tex('TEX', 'Texas Rangers', '#003278'),
  tor('TOR', 'Toronto Blue Jays', '#134A8E'),
  wsh('WSH', 'Washington Nationals', '#AB0003');

  const MlbTeam(this.code, this.displayName, this.primaryHex);

  final String code;
  final String displayName;
  final String primaryHex;

  Color get primaryColor => colorFromHex(primaryHex);

  static MlbTeam? fromCode(String? code) {
    final normalized = normalizeMlbTeamCode(code);
    if (normalized == null) {
      return null;
    }
    for (final team in MlbTeam.values) {
      if (team.code == normalized) {
        return team;
      }
    }
    return null;
  }
}

final Map<String, String> mlbTeamPrimaryHex = Map<String, String>.unmodifiable(<String, String>{
  for (final team in MlbTeam.values) team.code: team.primaryHex,
});

Color? teamPrimaryColor(String? teamCode) => MlbTeam.fromCode(teamCode)?.primaryColor;

const Map<String, String> _teamCodeAliases = <String, String>{
  'ANA': 'LAA',
  'CAL': 'LAA',
  'CHA': 'CWS',
  'CHN': 'CHC',
  'CLN': 'CLE',
  'FLO': 'MIA',
  'KCA': 'KC',
  'KCN': 'KC',
  'LAN': 'LAD',
  'BRO': 'LAD',
  'MON': 'WSH',
  'NYA': 'NYY',
  'NYN': 'NYM',
  'PHA': 'ATH',
  'OAK': 'ATH',
  'SDN': 'SD',
  'SFN': 'SF',
  'SLN': 'STL',
  'TBA': 'TB',
  'TBD': 'TB',
  'WAS': 'WSH',
  'WSN': 'WSH',
};

String? normalizeMlbTeamCode(String? rawTeamCode) {
  if (rawTeamCode == null || rawTeamCode.isEmpty) {
    return null;
  }

  final code = rawTeamCode.trim().toUpperCase();
  if (mlbTeamPrimaryHex.containsKey(code)) {
    return code;
  }
  return _teamCodeAliases[code];
}

Color colorFromHex(String hexColor) {
  final normalized = hexColor.replaceFirst('#', '');
  final value = int.parse('FF$normalized', radix: 16);
  return Color(value);
}
