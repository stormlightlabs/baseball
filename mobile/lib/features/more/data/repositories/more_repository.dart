import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_player_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/data_sources_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/leaders_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_summary_record.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';

abstract class MoreRepository {
  Future<List<SeasonSummaryRecord>> listSeasons();
  Future<SeasonSnapshot> fetchSeasonSnapshot({required int year, required String leagueFilter});
  Future<LeadersSnapshot> fetchLeaders({
    required LeadersMode mode,
    required LeadersScope scope,
    required int season,
    required String stat,
    String? league,
    int page,
    int perPage,
  });
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 8});
  Future<ComparePlayerSnapshot> fetchComparePlayer(String playerId);
  Future<DataSourcesSnapshot> fetchDataSources();
}
