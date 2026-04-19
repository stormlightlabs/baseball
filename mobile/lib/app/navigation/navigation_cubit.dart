import 'package:bigfly_mobile/app/navigation/navigation_state.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class NavigationCubit extends Cubit<NavigationState> {
  NavigationCubit() : super(const NavigationState.initial());

  void setTab(AppTab tab) => emit(state.copyWith(tab: tab));

  void setIndex(int index) => setTab(appTabFromIndex(index));
}
