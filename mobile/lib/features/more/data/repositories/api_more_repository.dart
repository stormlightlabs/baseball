import 'dart:math' as math;

import 'package:bigfly_mobile/core/data/json/json_helpers.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_batting_career.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_pitching_career.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_player_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/data_sources_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';
import 'package:bigfly_mobile/features/more/data/models/leaders_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/postseason_series_record.dart';
import 'package:bigfly_mobile/features/more/data/models/season_award_item.dart';
import 'package:bigfly_mobile/features/more/data/models/season_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_summary_record.dart';
import 'package:bigfly_mobile/features/more/data/repositories/more_repository.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';

class ApiMoreRepository implements MoreRepository {
  ApiMoreRepository(this._api);

  final BaseballApiClient _api;

  @override
  Future<List<SeasonSummaryRecord>> listSeasons() async {
    final payload = await _api.getSeasons();
    final rows =
        payload.map(asJsonMap).map(SeasonSummaryRecord.fromJson).where((row) => row.year > 0).toList(growable: false)
          ..sort((a, b) => b.year.compareTo(a.year));

    if (rows.isNotEmpty) {
      return rows;
    }

    final currentYear = DateTime.now().year;
    return <SeasonSummaryRecord>[
      SeasonSummaryRecord(year: currentYear, leagues: const <String>['AL', 'NL'], teamCount: 30),
    ];
  }

  @override
  Future<SeasonSnapshot> fetchSeasonSnapshot({required int year, required String leagueFilter}) async {
    final league = _normalizeLeagueFilter(leagueFilter);

    final teamsFuture = _safeFetchMap(() => _api.getSeasonTeams(year: year, perPage: 80));
    final teamBattingFuture = _safeFetchMap(() => _api.getTeamBattingStats(season: year, perPage: 80));
    final hrLeadersFuture = fetchLeaders(
      mode: LeadersMode.batting,
      scope: LeadersScope.season,
      season: year,
      stat: 'hr',
      league: league,
      perPage: 5,
    );
    final avgLeadersFuture = fetchLeaders(
      mode: LeadersMode.batting,
      scope: LeadersScope.season,
      season: year,
      stat: 'avg',
      league: league,
      perPage: 5,
    );
    final awardsFuture = _safeFetchMap(() => _api.getSeasonAwards(year: year, perPage: 6));
    final postseasonFuture = _safeFetchMap(() => _api.getSeasonPostseasonSeries(year: year));

    final teamsRaw = await teamsFuture;
    final teamBattingRaw = await teamBattingFuture;
    final hrLeaders = await hrLeadersFuture;
    final avgLeaders = await avgLeadersFuture;
    final awardsRaw = await awardsFuture;
    final postseasonRaw = await postseasonFuture;

    final teams =
        asJsonMapList(teamsRaw?['data'])
            .map(TeamSeasonRecord.fromJson)
            .where((row) => row.year == year)
            .where((row) => league == null || row.league.toUpperCase() == league)
            .toList(growable: false)
          ..sort((a, b) {
            final leagueCmp = a.league.compareTo(b.league);
            if (leagueCmp != 0) {
              return leagueCmp;
            }
            final divCmp = (a.division ?? '').compareTo(b.division ?? '');
            if (divCmp != 0) {
              return divCmp;
            }
            return b.wins.compareTo(a.wins);
          });

    final teamBattingRows = asJsonMapList(
      teamBattingRaw?['data'],
    ).where((row) => league == null || stringOrEmpty(row['league']).toUpperCase() == league).toList(growable: false);

    final totalHomeRuns = teamBattingRows.fold<int>(0, (sum, row) => sum + intOrZero(row['hr']));
    final averageAvg = teamBattingRows.isEmpty
        ? 0.0
        : teamBattingRows.fold<double>(0.0, (sum, row) => sum + doubleOrZero(row['avg'])) / teamBattingRows.length;

    final avgGamesPerTeam = teams.isEmpty
        ? 0
        : (teams.fold<int>(0, (sum, row) => sum + row.games) / math.max(1, teams.length)).round();

    final awards = asJsonMapList(awardsRaw?['data']).map(SeasonAwardItem.fromJson).toList(growable: false);
    final postseasonSeries = asJsonMapList(
      postseasonRaw?['series'],
    ).map(PostseasonSeriesRecord.fromJson).toList(growable: false);

    return SeasonSnapshot(
      year: year,
      teams: teams,
      totalHomeRuns: totalHomeRuns,
      leagueAverage: averageAvg,
      avgGamesPerTeam: avgGamesPerTeam,
      hrLeaders: hrLeaders.entries,
      avgLeaders: avgLeaders.entries,
      awards: awards,
      postseasonSeries: postseasonSeries,
    );
  }

  @override
  Future<LeadersSnapshot> fetchLeaders({
    required LeadersMode mode,
    required LeadersScope scope,
    required int season,
    required String stat,
    String? league,
    int page = 1,
    int perPage = 15,
  }) async {
    final normalizedStat = stat.toLowerCase();

    final payload = switch ((mode, scope)) {
      (LeadersMode.batting, LeadersScope.season) => await _api.getSeasonBattingLeaders(
        year: season,
        stat: normalizedStat,
        league: league,
        page: page,
        perPage: perPage,
      ),
      (LeadersMode.pitching, LeadersScope.season) => await _api.getSeasonPitchingLeaders(
        year: season,
        stat: normalizedStat,
        league: league,
        page: page,
        perPage: perPage,
      ),
      (LeadersMode.batting, LeadersScope.career) => await _api.getCareerBattingLeaders(
        stat: normalizedStat,
        page: page,
        perPage: perPage,
      ),
      (LeadersMode.pitching, LeadersScope.career) => await _api.getCareerPitchingLeaders(
        stat: normalizedStat,
        page: page,
        perPage: perPage,
      ),
    };

    final rows = asJsonMapList(payload['leaders']);
    final playerIds = rows
        .map((row) => stringOrEmpty(row['player_id']))
        .where((id) => id.isNotEmpty)
        .toSet()
        .toList(growable: false);
    final names = await _resolvePlayerNames(playerIds);

    final entries = rows
        .map((row) {
          final playerId = stringOrEmpty(row['player_id']);
          final rawValue = _extractStatValue(row, normalizedStat);
          return LeaderboardEntry(
            playerId: playerId,
            playerName: names[playerId] ?? playerId,
            teamId: stringOrEmpty(row['team_id']),
            league: stringOrEmpty(row['league']),
            year: nullableInt(row['year']),
            rawValue: rawValue,
            displayValue: _formatStatValue(normalizedStat, rawValue),
          );
        })
        .toList(growable: false);

    final responsePage = intOrZero(payload['page']);
    final responsePerPage = intOrZero(payload['per_page']);

    return LeadersSnapshot(
      stat: stringOrEmpty(payload['stat']).isEmpty ? normalizedStat : stringOrEmpty(payload['stat']),
      page: responsePage > 0 ? responsePage : page,
      perPage: responsePerPage > 0 ? responsePerPage : perPage,
      total: intOrZero(payload['total']),
      entries: entries,
    );
  }

  @override
  Future<List<PlayerSearchResult>> searchPlayers(String query, {int limit = 8}) async {
    final trimmed = query.trim();
    if (trimmed.isEmpty) {
      return const <PlayerSearchResult>[];
    }

    final payload = await _api.searchPlayers(query: trimmed, perPage: limit);
    return asJsonMapList(payload['data']).map(PlayerSearchResult.fromJson).toList(growable: false);
  }

  @override
  Future<ComparePlayerSnapshot> fetchComparePlayer(String playerId) async {
    final profileFuture = _api.getPlayer(playerId);
    final battingFuture = _safeFetchMap(() => _api.getPlayerBattingStats(playerId));
    final pitchingFuture = _safeFetchMap(() => _api.getPlayerPitchingStats(playerId));

    final profileRaw = await profileFuture;
    final battingRaw = await battingFuture;
    final pitchingRaw = await pitchingFuture;

    return ComparePlayerSnapshot(
      profile: PlayerProfile.fromJson(profileRaw),
      battingCareer: CompareBattingCareer.fromJson(battingRaw ?? const <String, dynamic>{}),
      pitchingCareer: ComparePitchingCareer.fromJson(pitchingRaw ?? const <String, dynamic>{}),
    );
  }

  @override
  Future<DataSourcesSnapshot> fetchDataSources() async {
    final metaFuture = _api.getMeta();
    final datasetsFuture = _safeFetchList(_api.getMetaDatasets);

    final metaRaw = await metaFuture;
    final datasetsRaw = await datasetsFuture;

    final meta = MetaSnapshot.fromJson(metaRaw);
    final datasets = (datasetsRaw ?? const <dynamic>[])
        .map(asJsonMap)
        .map(DatasetStatusSnapshot.fromJson)
        .toList(growable: false);

    return DataSourcesSnapshot(meta: meta, datasets: datasets.isEmpty ? meta.datasets : datasets);
  }

  String? _normalizeLeagueFilter(String value) {
    final normalized = value.trim().toUpperCase();
    if (normalized == 'AL' || normalized == 'NL') {
      return normalized;
    }
    return null;
  }

  Future<Map<String, String>> _resolvePlayerNames(List<String> playerIds) async {
    if (playerIds.isEmpty) {
      return const <String, String>{};
    }

    final entries = await Future.wait(
      playerIds.map((id) async {
        final payload = await _safeFetchMap(() => _api.getPlayer(id));
        if (payload == null || payload.isEmpty) {
          return MapEntry<String, String>(id, id);
        }
        final profile = PlayerProfile.fromJson(payload);
        return MapEntry<String, String>(id, profile.fullName);
      }),
    );

    return Map<String, String>.fromEntries(entries);
  }

  double _extractStatValue(Map<String, dynamic> row, String stat) {
    final aliases = <String, List<String>>{
      'k_per_9': <String>['k_per_9', 'k9'],
      'bb_per_9': <String>['bb_per_9', 'bb9'],
      'hr_per_9': <String>['hr_per_9', 'hr9'],
      'avg': <String>['avg'],
      'obp': <String>['obp'],
      'slg': <String>['slg'],
      'ops': <String>['ops'],
      'era': <String>['era'],
      'whip': <String>['whip'],
      'w': <String>['w'],
      'l': <String>['l'],
      'so': <String>['so'],
      'sv': <String>['sv'],
      'hr': <String>['hr'],
      'h': <String>['h'],
      'rbi': <String>['rbi'],
      'sb': <String>['sb'],
      'r': <String>['r'],
      'bb': <String>['bb'],
    };

    final candidates = aliases[stat] ?? <String>[stat];
    for (final candidate in candidates) {
      if (row.containsKey(candidate)) {
        return doubleOrZero(row[candidate]);
      }
    }
    return 0;
  }

  String _formatStatValue(String stat, double value) {
    const triplePrecision = <String>{'avg', 'obp', 'slg', 'ops'};
    const doublePrecision = <String>{'era', 'whip', 'k_per_9', 'bb_per_9', 'hr_per_9'};

    if (triplePrecision.contains(stat)) {
      return value.toStringAsFixed(3).replaceFirst('0.', '.');
    }
    if (doublePrecision.contains(stat)) {
      return value.toStringAsFixed(2);
    }

    if ((value - value.round()).abs() < 0.0001) {
      return value.round().toString();
    }
    return value.toStringAsFixed(1);
  }

  Future<Map<String, dynamic>?> _safeFetchMap(Future<Map<String, dynamic>> Function() operation) async {
    try {
      return await operation();
    } catch (_) {
      return null;
    }
  }

  Future<List<dynamic>?> _safeFetchList(Future<List<dynamic>> Function() operation) async {
    try {
      return await operation();
    } catch (_) {
      return null;
    }
  }
}
