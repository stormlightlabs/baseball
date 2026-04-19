class NavigationState {
  const NavigationState({required this.tab});

  const NavigationState.initial() : tab = AppTab.home;

  final AppTab tab;

  int get index => tab.index;

  NavigationState copyWith({AppTab? tab}) {
    return NavigationState(tab: tab ?? this.tab);
  }
}

enum AppTab { home, players, teams, games, more }

AppTab appTabFromIndex(int index) {
  if (index < 0 || index >= AppTab.values.length) {
    return AppTab.home;
  }
  return AppTab.values[index];
}
