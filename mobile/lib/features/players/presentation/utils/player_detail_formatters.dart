import 'package:bigfly_mobile/features/players/application/players_types.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:fl_chart/fl_chart.dart';

String playerDetailTabLabel(PlayerDetailTab tab) => switch (tab) {
  PlayerDetailTab.batting => 'Batting',
  PlayerDetailTab.pitching => 'Pitching',
  PlayerDetailTab.awards => 'Awards',
  PlayerDetailTab.hallOfFame => 'HOF',
};

String battingMetricLabel(BattingChartMetric metric) => switch (metric) {
  BattingChartMetric.hr => 'HR',
  BattingChartMetric.avg => 'AVG',
  BattingChartMetric.rbi => 'RBI',
  BattingChartMetric.obp => 'OBP',
  BattingChartMetric.slg => 'SLG',
  BattingChartMetric.ops => 'OPS',
};

String pitchingMetricLabel(PitchingChartMetric metric) => switch (metric) {
  PitchingChartMetric.era => 'ERA',
  PitchingChartMetric.strikeouts => 'SO',
  PitchingChartMetric.whip => 'WHIP',
  PitchingChartMetric.wins => 'W',
  PitchingChartMetric.kPer9 => 'K/9',
};

int compareBattingSeasons(PlayerBattingSeason a, PlayerBattingSeason b, BattingSortColumn column) => switch (column) {
  BattingSortColumn.year => a.year.compareTo(b.year),
  BattingSortColumn.team => a.teamId.compareTo(b.teamId),
  BattingSortColumn.g => a.g.compareTo(b.g),
  BattingSortColumn.ab => a.ab.compareTo(b.ab),
  BattingSortColumn.avg => a.avg.compareTo(b.avg),
  BattingSortColumn.hr => a.hr.compareTo(b.hr),
  BattingSortColumn.rbi => a.rbi.compareTo(b.rbi),
  BattingSortColumn.ops => a.ops.compareTo(b.ops),
};

int comparePitchingSeasons(PlayerPitchingSeason a, PlayerPitchingSeason b, PitchingSortColumn column) =>
    switch (column) {
      PitchingSortColumn.year => a.year.compareTo(b.year),
      PitchingSortColumn.team => a.teamId.compareTo(b.teamId),
      PitchingSortColumn.g => a.games.compareTo(b.games),
      PitchingSortColumn.wins => a.wins.compareTo(b.wins),
      PitchingSortColumn.losses => a.losses.compareTo(b.losses),
      PitchingSortColumn.era => a.era.compareTo(b.era),
      PitchingSortColumn.so => a.so.compareTo(b.so),
      PitchingSortColumn.whip => a.whip.compareTo(b.whip),
    };

int battingSortColumnIndex(BattingSortColumn column) => switch (column) {
  BattingSortColumn.year => 0,
  BattingSortColumn.team => 1,
  BattingSortColumn.g => 2,
  BattingSortColumn.ab => 3,
  BattingSortColumn.avg => 4,
  BattingSortColumn.hr => 5,
  BattingSortColumn.rbi => 6,
  BattingSortColumn.ops => 7,
};

int pitchingSortColumnIndex(PitchingSortColumn column) => switch (column) {
  PitchingSortColumn.year => 0,
  PitchingSortColumn.team => 1,
  PitchingSortColumn.g => 2,
  PitchingSortColumn.wins => 3,
  PitchingSortColumn.losses => 4,
  PitchingSortColumn.era => 5,
  PitchingSortColumn.so => 6,
  PitchingSortColumn.whip => 7,
};

List<FlSpot> battingMetricPoints(List<PlayerBattingSeason> seasons, BattingChartMetric metric) {
  if (seasons.isEmpty) {
    return <FlSpot>[];
  }

  return seasons
      .asMap()
      .entries
      .map((entry) => FlSpot(entry.key.toDouble(), _battingMetricValue(entry.value, metric)))
      .toList(growable: false);
}

List<FlSpot> pitchingMetricPoints(List<PlayerPitchingSeason> seasons, PitchingChartMetric metric) {
  if (seasons.isEmpty) {
    return <FlSpot>[];
  }

  return seasons
      .asMap()
      .entries
      .map((entry) => FlSpot(entry.key.toDouble(), _pitchingMetricValue(entry.value, metric)))
      .toList(growable: false);
}

double _battingMetricValue(PlayerBattingSeason season, BattingChartMetric metric) => switch (metric) {
  BattingChartMetric.hr => season.hr.toDouble(),
  BattingChartMetric.avg => season.avg,
  BattingChartMetric.rbi => season.rbi.toDouble(),
  BattingChartMetric.obp => season.obp,
  BattingChartMetric.slg => season.slg,
  BattingChartMetric.ops => season.ops,
};

double _pitchingMetricValue(PlayerPitchingSeason season, PitchingChartMetric metric) => switch (metric) {
  PitchingChartMetric.era => season.era,
  PitchingChartMetric.strikeouts => season.so.toDouble(),
  PitchingChartMetric.whip => season.whip,
  PitchingChartMetric.wins => season.wins.toDouble(),
  PitchingChartMetric.kPer9 => season.kPer9,
};
