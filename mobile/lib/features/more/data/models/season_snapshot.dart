import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';
import 'package:bigfly_mobile/features/more/data/models/postseason_series_record.dart';
import 'package:bigfly_mobile/features/more/data/models/season_award_item.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';

class SeasonSnapshot {
  const SeasonSnapshot({
    required this.year,
    required this.teams,
    required this.totalHomeRuns,
    required this.leagueAverage,
    required this.avgGamesPerTeam,
    required this.hrLeaders,
    required this.avgLeaders,
    required this.awards,
    required this.postseasonSeries,
  });

  final int year;
  final List<TeamSeasonRecord> teams;
  final int totalHomeRuns;
  final double leagueAverage;
  final int avgGamesPerTeam;
  final List<LeaderboardEntry> hrLeaders;
  final List<LeaderboardEntry> avgLeaders;
  final List<SeasonAwardItem> awards;
  final List<PostseasonSeriesRecord> postseasonSeries;
}
