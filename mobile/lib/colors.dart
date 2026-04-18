import 'package:flutter/material.dart';

enum MlbTeam {
  diamondbacks(167, 25, 48, 227, 212, 173),
  braves(206, 17, 65, 19, 39, 79),
  orioles(223, 70, 1, 0, 0, 0),
  redSox(189, 48, 57, 12, 35, 64),
  cubs(14, 51, 134, 204, 52, 51),
  whiteSox(39, 37, 31, 196, 206, 212),
  reds(198, 1, 31, 0, 0, 0),
  guardians(227, 25, 55, 12, 35, 64),
  rockies(51, 0, 111, 0, 0, 0),
  tigers(12, 35, 64, 250, 70, 22),
  astros(0, 45, 98, 235, 110, 31),
  royals(0, 70, 135, 189, 155, 96),
  angels(186, 0, 33, 0, 50, 99),
  dodgers(0, 90, 156, 255, 255, 255),
  marlins(0, 0, 0, 0, 119, 200),
  brewers(18, 40, 75, 255, 197, 47),
  twins(0, 43, 92, 211, 17, 69),
  mets(0, 45, 114, 255, 89, 16),
  yankees(19, 36, 72, 196, 206, 212),
  athletics(0, 56, 49, 239, 178, 30),
  phillies(232, 24, 40, 0, 45, 114),
  pirates(39, 37, 31, 253, 184, 39),
  padres(47, 36, 29, 162, 170, 173),
  giants(253, 90, 30, 39, 37, 31),
  mariners(12, 44, 86, 0, 92, 92),
  cardinals(196, 30, 58, 12, 35, 64),
  rays(9, 44, 92, 143, 188, 230),
  rangers(0, 50, 120, 192, 17, 31),
  blueJays(19, 74, 142, 29, 45, 92),
  nationals(171, 0, 3, 20, 34, 90);

  const MlbTeam(
    this.primaryR,
    this.primaryG,
    this.primaryB,
    this.secondaryR,
    this.secondaryG,
    this.secondaryB,
  );

  final int primaryR;
  final int primaryG;
  final int primaryB;
  final int secondaryR;
  final int secondaryG;
  final int secondaryB;

  Color get primaryColor => Color.fromARGB(255, primaryR, primaryG, primaryB);
  Color get secondaryColor =>
      Color.fromARGB(255, secondaryR, secondaryG, secondaryB);

  String get displayName => switch (this) {
    MlbTeam.diamondbacks => 'Arizona Diamondbacks',
    MlbTeam.braves => 'Atlanta Braves',
    MlbTeam.orioles => 'Baltimore Orioles',
    MlbTeam.redSox => 'Boston Red Sox',
    MlbTeam.cubs => 'Chicago Cubs',
    MlbTeam.whiteSox => 'Chicago White Sox',
    MlbTeam.reds => 'Cincinnati Reds',
    MlbTeam.guardians => 'Cleveland Guardians',
    MlbTeam.rockies => 'Colorado Rockies',
    MlbTeam.tigers => 'Detroit Tigers',
    MlbTeam.astros => 'Houston Astros',
    MlbTeam.royals => 'Kansas City Royals',
    MlbTeam.angels => 'Los Angeles Angels',
    MlbTeam.dodgers => 'Los Angeles Dodgers',
    MlbTeam.marlins => 'Miami Marlins',
    MlbTeam.brewers => 'Milwaukee Brewers',
    MlbTeam.twins => 'Minnesota Twins',
    MlbTeam.mets => 'New York Mets',
    MlbTeam.yankees => 'New York Yankees',
    MlbTeam.athletics => 'Oakland Athletics',
    MlbTeam.phillies => 'Philadelphia Phillies',
    MlbTeam.pirates => 'Pittsburgh Pirates',
    MlbTeam.padres => 'San Diego Padres',
    MlbTeam.giants => 'San Francisco Giants',
    MlbTeam.mariners => 'Seattle Mariners',
    MlbTeam.cardinals => 'St. Louis Cardinals',
    MlbTeam.rays => 'Tampa Bay Rays',
    MlbTeam.rangers => 'Texas Rangers',
    MlbTeam.blueJays => 'Toronto Blue Jays',
    MlbTeam.nationals => 'Washington Nationals',
  };

  String get primaryHex =>
      '#${primaryR.toRadixString(16).padLeft(2, '0').toUpperCase()}'
      '${primaryG.toRadixString(16).padLeft(2, '0').toUpperCase()}'
      '${primaryB.toRadixString(16).padLeft(2, '0').toUpperCase()}';

  String get secondaryHex =>
      '#${secondaryR.toRadixString(16).padLeft(2, '0').toUpperCase()}'
      '${secondaryG.toRadixString(16).padLeft(2, '0').toUpperCase()}'
      '${secondaryB.toRadixString(16).padLeft(2, '0').toUpperCase()}';

  ColorScheme colorScheme({Brightness brightness = Brightness.light}) =>
      ColorScheme.fromSeed(
        seedColor: primaryColor,
        brightness: brightness,
        primary: primaryColor,
        secondary: secondaryColor,
      );
}
