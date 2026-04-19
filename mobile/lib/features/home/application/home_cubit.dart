import 'package:bigfly_mobile/features/home/application/home_state.dart';
import 'package:bigfly_mobile/features/home/application/home_types.dart';
import 'package:bigfly_mobile/features/home/data/repositories/home_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

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
