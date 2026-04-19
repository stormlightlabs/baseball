import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';

enum HomeMetaStatus { initial, loading, ready, failure }

enum HomeSearchStatus { idle, loading, ready, failure }

class HomeState {
  const HomeState({
    required this.activeEntity,
    required this.query,
    required this.metaStatus,
    required this.searchStatus,
    required this.searchResults,
    this.meta,
    this.metaError,
    this.searchError,
  });

  const HomeState.initial()
    : this(
        activeEntity: HomeEntityType.players,
        query: '',
        metaStatus: HomeMetaStatus.initial,
        searchStatus: HomeSearchStatus.idle,
        searchResults: const <HomeSearchItem>[],
      );

  final HomeEntityType activeEntity;
  final String query;
  final HomeMetaStatus metaStatus;
  final HomeSearchStatus searchStatus;
  final List<HomeSearchItem> searchResults;
  final MetaSnapshot? meta;
  final String? metaError;
  final String? searchError;

  HomeState copyWith({
    HomeEntityType? activeEntity,
    String? query,
    HomeMetaStatus? metaStatus,
    HomeSearchStatus? searchStatus,
    List<HomeSearchItem>? searchResults,
    MetaSnapshot? meta,
    String? metaError,
    String? searchError,
  }) {
    return HomeState(
      activeEntity: activeEntity ?? this.activeEntity,
      query: query ?? this.query,
      metaStatus: metaStatus ?? this.metaStatus,
      searchStatus: searchStatus ?? this.searchStatus,
      searchResults: searchResults ?? this.searchResults,
      meta: meta ?? this.meta,
      metaError: metaError,
      searchError: searchError,
    );
  }
}
