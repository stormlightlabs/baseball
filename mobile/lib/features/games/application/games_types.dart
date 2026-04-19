enum GamesStatus { initial, loading, ready, failure }

enum GamesQuickFilter { extraInnings, doubleheaders, postseason }

String gamesQuickFilterLabel(GamesQuickFilter filter) {
  return switch (filter) {
    GamesQuickFilter.extraInnings => 'Extra innings',
    GamesQuickFilter.doubleheaders => 'Doubleheaders',
    GamesQuickFilter.postseason => 'Postseason',
  };
}
