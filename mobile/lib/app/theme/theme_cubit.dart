import 'package:bigfly_mobile/colors.dart';
import 'package:bigfly_mobile/data/local/cache_store.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class AppThemeState {
  const AppThemeState({this.selectedTeamCode});

  final String? selectedTeamCode;

  MlbTeam? get selectedTeam => MlbTeam.fromCode(selectedTeamCode);
  Color? get selectedSeedColor => selectedTeam?.primaryColor;
}

class ThemeCubit extends Cubit<AppThemeState> {
  ThemeCubit(this._cacheStore) : super(AppThemeState(selectedTeamCode: _cacheStore.selectedTeamCode));

  final CacheStore _cacheStore;

  Future<void> selectTeam(String? teamCode) async {
    final normalized = MlbTeam.fromCode(teamCode)?.code;
    emit(AppThemeState(selectedTeamCode: normalized));
    await _cacheStore.setSelectedTeamCode(normalized);
  }
}
