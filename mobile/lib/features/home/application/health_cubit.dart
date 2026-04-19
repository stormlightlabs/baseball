import 'package:bigfly_mobile/features/home/data/repositories/health_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

enum HealthLoadStatus { initial, loading, success, failure }

class HealthState {
  const HealthState({required this.loadStatus, this.statusText, this.errorMessage, this.fromCache = false});

  const HealthState.initial() : this(loadStatus: HealthLoadStatus.initial, statusText: null, errorMessage: null);

  final HealthLoadStatus loadStatus;
  final String? statusText;
  final String? errorMessage;
  final bool fromCache;

  HealthState copyWith({HealthLoadStatus? loadStatus, String? statusText, String? errorMessage, bool? fromCache}) {
    return HealthState(
      loadStatus: loadStatus ?? this.loadStatus,
      statusText: statusText ?? this.statusText,
      errorMessage: errorMessage ?? this.errorMessage,
      fromCache: fromCache ?? this.fromCache,
    );
  }
}

class HealthCubit extends Cubit<HealthState> {
  HealthCubit(this._repository) : super(_initialState(_repository));

  final HealthRepository _repository;

  static HealthState _initialState(HealthRepository repository) {
    final cached = repository.getCachedHealth();
    if (cached == null) {
      return const HealthState.initial();
    }
    return HealthState(loadStatus: HealthLoadStatus.success, statusText: cached.status, fromCache: true);
  }

  Future<void> loadHealth() async {
    emit(state.copyWith(loadStatus: HealthLoadStatus.loading, errorMessage: null));

    try {
      final health = await _repository.fetchHealth();
      emit(HealthState(loadStatus: HealthLoadStatus.success, statusText: health.status, fromCache: false));
    } catch (error) {
      final cached = _repository.getCachedHealth();
      if (cached != null) {
        emit(
          HealthState(
            loadStatus: HealthLoadStatus.success,
            statusText: cached.status,
            fromCache: true,
            errorMessage: error.toString(),
          ),
        );
        return;
      }

      emit(HealthState(loadStatus: HealthLoadStatus.failure, errorMessage: error.toString()));
    }
  }
}
