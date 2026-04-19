import 'package:bigfly_mobile/features/more/application/types/leaders_mode.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_scope.dart';
import 'package:bigfly_mobile/features/more/application/types/leaders_stat_option.dart';
import 'package:bigfly_mobile/features/more/application/types/more_section.dart';
import 'package:bigfly_mobile/features/more/application/types/more_status.dart';
import 'package:bigfly_mobile/features/more/data/models/compare_player_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/data_sources_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/leaders_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_snapshot.dart';
import 'package:bigfly_mobile/features/more/data/models/season_summary_record.dart';
import 'package:bigfly_mobile/features/players/data/models/player_models.dart';

class MoreState {
  const MoreState({
    required this.selectedSection,
    required this.availableSeasons,
    required this.selectedSeason,
    required this.seasonLeagueFilter,
    required this.seasonStatus,
    required this.leadersStatus,
    required this.dataSourcesStatus,
    required this.leadersMode,
    required this.leadersScope,
    required this.leadersStat,
    required this.compareSearchA,
    required this.compareSearchB,
    required this.compareSearchResultsA,
    required this.compareSearchResultsB,
    required this.compareSearchLoadingA,
    required this.compareSearchLoadingB,
    required this.compareLoadStatusA,
    required this.compareLoadStatusB,
    this.seasonSnapshot,
    this.leadersSnapshot,
    this.dataSourcesSnapshot,
    this.comparePlayerA,
    this.comparePlayerB,
    this.seasonError,
    this.leadersError,
    this.dataSourcesError,
    this.compareErrorA,
    this.compareErrorB,
    this.compareSearchErrorA,
    this.compareSearchErrorB,
  });

  const MoreState.initial()
    : this(
        selectedSection: MoreSection.seasons,
        availableSeasons: const <SeasonSummaryRecord>[],
        selectedSeason: 0,
        seasonLeagueFilter: 'Both',
        seasonStatus: MoreStatus.initial,
        leadersStatus: MoreStatus.initial,
        dataSourcesStatus: MoreStatus.initial,
        leadersMode: LeadersMode.batting,
        leadersScope: LeadersScope.season,
        leadersStat: 'hr',
        compareSearchA: '',
        compareSearchB: '',
        compareSearchResultsA: const <PlayerSearchResult>[],
        compareSearchResultsB: const <PlayerSearchResult>[],
        compareSearchLoadingA: false,
        compareSearchLoadingB: false,
        compareLoadStatusA: MoreStatus.initial,
        compareLoadStatusB: MoreStatus.initial,
      );

  final MoreSection selectedSection;
  final List<SeasonSummaryRecord> availableSeasons;
  final int selectedSeason;
  final String seasonLeagueFilter;

  final MoreStatus seasonStatus;
  final SeasonSnapshot? seasonSnapshot;
  final String? seasonError;

  final MoreStatus leadersStatus;
  final LeadersMode leadersMode;
  final LeadersScope leadersScope;
  final String leadersStat;
  final LeadersSnapshot? leadersSnapshot;
  final String? leadersError;

  final MoreStatus dataSourcesStatus;
  final DataSourcesSnapshot? dataSourcesSnapshot;
  final String? dataSourcesError;

  final String compareSearchA;
  final String compareSearchB;
  final List<PlayerSearchResult> compareSearchResultsA;
  final List<PlayerSearchResult> compareSearchResultsB;
  final bool compareSearchLoadingA;
  final bool compareSearchLoadingB;
  final String? compareSearchErrorA;
  final String? compareSearchErrorB;

  final MoreStatus compareLoadStatusA;
  final MoreStatus compareLoadStatusB;
  final ComparePlayerSnapshot? comparePlayerA;
  final ComparePlayerSnapshot? comparePlayerB;
  final String? compareErrorA;
  final String? compareErrorB;

  bool get hasCompareSelection => comparePlayerA != null && comparePlayerB != null;

  List<LeadersStatOption> get activeStatOptions =>
      leadersMode == LeadersMode.batting ? battingStatOptions : pitchingStatOptions;

  MoreState copyWith({
    MoreSection? selectedSection,
    List<SeasonSummaryRecord>? availableSeasons,
    int? selectedSeason,
    String? seasonLeagueFilter,
    MoreStatus? seasonStatus,
    SeasonSnapshot? seasonSnapshot,
    MoreStatus? leadersStatus,
    LeadersMode? leadersMode,
    LeadersScope? leadersScope,
    String? leadersStat,
    LeadersSnapshot? leadersSnapshot,
    MoreStatus? dataSourcesStatus,
    DataSourcesSnapshot? dataSourcesSnapshot,
    String? compareSearchA,
    String? compareSearchB,
    List<PlayerSearchResult>? compareSearchResultsA,
    List<PlayerSearchResult>? compareSearchResultsB,
    bool? compareSearchLoadingA,
    bool? compareSearchLoadingB,
    MoreStatus? compareLoadStatusA,
    MoreStatus? compareLoadStatusB,
    ComparePlayerSnapshot? comparePlayerA,
    ComparePlayerSnapshot? comparePlayerB,
    String? seasonError,
    String? leadersError,
    String? dataSourcesError,
    String? compareErrorA,
    String? compareErrorB,
    String? compareSearchErrorA,
    String? compareSearchErrorB,
  }) {
    return MoreState(
      selectedSection: selectedSection ?? this.selectedSection,
      availableSeasons: availableSeasons ?? this.availableSeasons,
      selectedSeason: selectedSeason ?? this.selectedSeason,
      seasonLeagueFilter: seasonLeagueFilter ?? this.seasonLeagueFilter,
      seasonStatus: seasonStatus ?? this.seasonStatus,
      seasonSnapshot: seasonSnapshot ?? this.seasonSnapshot,
      leadersStatus: leadersStatus ?? this.leadersStatus,
      leadersMode: leadersMode ?? this.leadersMode,
      leadersScope: leadersScope ?? this.leadersScope,
      leadersStat: leadersStat ?? this.leadersStat,
      leadersSnapshot: leadersSnapshot ?? this.leadersSnapshot,
      dataSourcesStatus: dataSourcesStatus ?? this.dataSourcesStatus,
      dataSourcesSnapshot: dataSourcesSnapshot ?? this.dataSourcesSnapshot,
      compareSearchA: compareSearchA ?? this.compareSearchA,
      compareSearchB: compareSearchB ?? this.compareSearchB,
      compareSearchResultsA: compareSearchResultsA ?? this.compareSearchResultsA,
      compareSearchResultsB: compareSearchResultsB ?? this.compareSearchResultsB,
      compareSearchLoadingA: compareSearchLoadingA ?? this.compareSearchLoadingA,
      compareSearchLoadingB: compareSearchLoadingB ?? this.compareSearchLoadingB,
      compareLoadStatusA: compareLoadStatusA ?? this.compareLoadStatusA,
      compareLoadStatusB: compareLoadStatusB ?? this.compareLoadStatusB,
      comparePlayerA: comparePlayerA ?? this.comparePlayerA,
      comparePlayerB: comparePlayerB ?? this.comparePlayerB,
      seasonError: seasonError,
      leadersError: leadersError,
      dataSourcesError: dataSourcesError,
      compareErrorA: compareErrorA,
      compareErrorB: compareErrorB,
      compareSearchErrorA: compareSearchErrorA,
      compareSearchErrorB: compareSearchErrorB,
    );
  }
}
