import 'package:bigfly_mobile/features/teams/application/teams_types.dart';
import 'package:bigfly_mobile/features/teams/data/models/team_models.dart';

class TeamsState {
  const TeamsState({
    required this.initialized,
    required this.searchQuery,
    required this.searchLoading,
    required this.searchResults,
    required this.detailStatus,
    required this.activeSegment,
    required this.franchises,
    this.detail,
    this.searchError,
    this.detailError,
    this.selectedTeam,
  });

  const TeamsState.initial()
    : this(
        initialized: false,
        searchQuery: '',
        searchLoading: false,
        searchResults: const <TeamSeasonRecord>[],
        detailStatus: TeamDetailStatus.initial,
        activeSegment: TeamDetailSegment.overview,
        franchises: const <FranchiseSummary>[],
      );

  final bool initialized;
  final String searchQuery;
  final bool searchLoading;
  final List<TeamSeasonRecord> searchResults;
  final TeamDetailStatus detailStatus;
  final TeamDetailBundle? detail;
  final String? searchError;
  final String? detailError;
  final TeamDetailSegment activeSegment;
  final List<FranchiseSummary> franchises;
  final TeamSeasonRecord? selectedTeam;

  TeamsState copyWith({
    bool? initialized,
    String? searchQuery,
    bool? searchLoading,
    List<TeamSeasonRecord>? searchResults,
    TeamDetailStatus? detailStatus,
    TeamDetailBundle? detail,
    String? searchError,
    String? detailError,
    TeamDetailSegment? activeSegment,
    List<FranchiseSummary>? franchises,
    TeamSeasonRecord? selectedTeam,
  }) {
    return TeamsState(
      initialized: initialized ?? this.initialized,
      searchQuery: searchQuery ?? this.searchQuery,
      searchLoading: searchLoading ?? this.searchLoading,
      searchResults: searchResults ?? this.searchResults,
      detailStatus: detailStatus ?? this.detailStatus,
      detail: detail ?? this.detail,
      searchError: searchError,
      detailError: detailError,
      activeSegment: activeSegment ?? this.activeSegment,
      franchises: franchises ?? this.franchises,
      selectedTeam: selectedTeam ?? this.selectedTeam,
    );
  }
}
