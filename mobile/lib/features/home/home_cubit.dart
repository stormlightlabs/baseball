import 'package:bigfly_mobile/data/models/meta_models.dart';
import 'package:bigfly_mobile/data/repositories/home_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

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

class HomeCubit extends Cubit<HomeState> {
  HomeCubit(this._repository) : super(const HomeState.initial());

  final HomeRepository _repository;

  Future<void> initialize() async {
    await loadMeta();
  }

  Future<void> loadMeta() async {
    emit(state.copyWith(metaStatus: HomeMetaStatus.loading, metaError: null));
    try {
      final meta = await _repository.fetchMeta();
      emit(state.copyWith(metaStatus: HomeMetaStatus.ready, meta: meta, metaError: null));
    } catch (error) {
      emit(state.copyWith(metaStatus: HomeMetaStatus.failure, metaError: error.toString()));
    }
  }

  void setEntity(HomeEntityType entity) {
    emit(state.copyWith(activeEntity: entity));
  }

  void setQuery(String query) {
    emit(state.copyWith(query: query));
  }

  Future<void> runSearch() async {
    final query = state.query.trim();
    if (query.isEmpty) {
      emit(
        state.copyWith(searchStatus: HomeSearchStatus.idle, searchResults: const <HomeSearchItem>[], searchError: null),
      );
      return;
    }

    emit(
      state.copyWith(
        searchStatus: HomeSearchStatus.loading,
        searchResults: const <HomeSearchItem>[],
        searchError: null,
      ),
    );

    try {
      final results = await _repository.search(state.activeEntity, query);
      emit(state.copyWith(searchStatus: HomeSearchStatus.ready, searchResults: results, searchError: null));
    } catch (error) {
      emit(
        state.copyWith(
          searchStatus: HomeSearchStatus.failure,
          searchResults: const <HomeSearchItem>[],
          searchError: error.toString(),
        ),
      );
    }
  }
}
