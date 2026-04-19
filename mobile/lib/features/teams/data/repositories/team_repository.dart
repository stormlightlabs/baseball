import 'package:bigfly_mobile/colors.dart';
import 'package:bigfly_mobile/core/data/json/json_helpers.dart';
import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';

abstract class TeamRepository {
  Future<List<FranchiseSummary>> listFranchises({bool active = true});
  Future<List<TeamSeasonRecord>> searchTeams(String query, {int limit = 12});
  Future<TeamSeasonRecord?> seedTeamForCode(String? teamCode);
  Future<TeamDetailBundle> fetchTeamDetail(TeamSeasonRecord team);
}

class ApiTeamRepository implements TeamRepository {
  ApiTeamRepository(this._api);

  final BaseballApiClient _api;

  @override
  Future<List<FranchiseSummary>> listFranchises({bool active = true}) async {
    final payload = await _api.getFranchises(active: active);
    return asJsonMapList(payload['franchises']).map(FranchiseSummary.fromJson).toList(growable: false)
      ..sort((a, b) => a.name.compareTo(b.name));
  }

  @override
  Future<List<TeamSeasonRecord>> searchTeams(String query, {int limit = 12}) async {
    final payload = await _api.searchTeams(query: query, perPage: limit);
    final rows = asJsonMapList(payload['data']).map(TeamSeasonRecord.fromJson).toList(growable: false);
    return _sortedByYear(rows);
  }

  @override
  Future<TeamSeasonRecord?> seedTeamForCode(String? teamCode) async {
    final normalized = normalizeMlbTeamCode(teamCode);
    final query = normalized ?? teamCode ?? 'NYY';

    final candidates = await searchTeams(query, limit: 40);
    if (candidates.isEmpty) {
      return null;
    }

    final exact = candidates
        .where((row) {
          if (normalized == null) {
            return false;
          }
          return normalizeMlbTeamCode(row.teamId) == normalized || normalizeMlbTeamCode(row.franchiseId) == normalized;
        })
        .toList(growable: false);

    if (exact.isNotEmpty) {
      return _sortedByYear(exact).first;
    }

    return candidates.first;
  }

  @override
  Future<TeamDetailBundle> fetchTeamDetail(TeamSeasonRecord team) async {
    final seasonFuture = _api.getTeamSeason(team.teamId, year: team.year);
    final franchiseFuture = _safeFetch<Map<String, dynamic>>(() => _api.getFranchise(team.franchiseId));
    final recentSeasonsFuture = _safeFetch<Map<String, dynamic>>(
      () => _api.searchTeams(query: team.teamId, perPage: 120),
    );
    final rosterFuture = _safeFetch<List<dynamic>>(() => _api.getTeamRoster(year: team.year, teamId: team.teamId));
    final scheduleFuture = _safeFetch<Map<String, dynamic>>(
      () => _api.getTeamSchedule(year: team.year, teamId: team.teamId, perPage: 60),
    );
    final dailyLogsFuture = _safeFetch<Map<String, dynamic>>(
      () => _api.getTeamDailyLogs(year: team.year, teamId: team.teamId, perPage: 60),
    );
    final runDiffFuture = _safeFetch<Map<String, dynamic>>(
      () => _api.getTeamRunDifferential(teamId: team.teamId, season: team.year, windows: const <int>[30]),
    );

    final seasonRaw = await seasonFuture;
    final franchiseRaw = await franchiseFuture;
    final recentSeasonsRaw = await recentSeasonsFuture;
    final rosterRaw = await rosterFuture;
    final scheduleRaw = await scheduleFuture;
    final dailyLogsRaw = await dailyLogsFuture;
    final runDiffRaw = await runDiffFuture;

    final resolvedTeam = TeamSeasonRecord.fromJson(seasonRaw);
    final franchise = franchiseRaw == null ? null : FranchiseSummary.fromJson(franchiseRaw);

    final recentSeasons = _sortedByYear(
      asJsonMapList(
        recentSeasonsRaw?['data'],
      ).map(TeamSeasonRecord.fromJson).where((row) => row.teamId == resolvedTeam.teamId).toList(growable: false),
    );

    final roster = asJsonMapList(rosterRaw)
      ..sort((a, b) => stringOrEmpty(a['last_name']).compareTo(stringOrEmpty(b['last_name'])));
    final parsedRoster = roster.map(TeamRosterPlayer.fromJson).toList(growable: false);

    final schedule = asJsonMapList(scheduleRaw?['data']).map(TeamGameSummary.fromJson).toList(growable: false)
      ..sort((a, b) => _compareDatesDesc(a.date, b.date));

    final dailyLogs = asJsonMapList(dailyLogsRaw?['data']).map(TeamDailyLog.fromJson).toList(growable: false)
      ..sort((a, b) => _compareDatesDesc(a.date, b.date));

    final runDifferential = runDiffRaw == null ? null : TeamRunDifferentialSeries.fromJson(runDiffRaw);

    return TeamDetailBundle(
      team: resolvedTeam,
      franchise: franchise,
      recentSeasons: recentSeasons.take(5).toList(growable: false),
      roster: parsedRoster,
      schedule: schedule,
      dailyLogs: dailyLogs,
      runDifferential: runDifferential,
      themeTeamCode: normalizeMlbTeamCode(resolvedTeam.franchiseId) ?? normalizeMlbTeamCode(resolvedTeam.teamId),
    );
  }

  List<TeamSeasonRecord> _sortedByYear(List<TeamSeasonRecord> rows) {
    final sorted = [...rows]..sort((a, b) => b.year.compareTo(a.year));
    return sorted;
  }

  Future<T?> _safeFetch<T>(Future<T> Function() operation) async {
    try {
      return await operation();
    } catch (_) {
      return null;
    }
  }

  int _compareDatesDesc(DateTime? a, DateTime? b) {
    if (a == null && b == null) {
      return 0;
    }
    if (a == null) {
      return 1;
    }
    if (b == null) {
      return -1;
    }
    return b.compareTo(a);
  }
}
