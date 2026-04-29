class ScorekeeperRoutes {
  static const String hub = '/more/scorekeeper';
  static const String newGame = '/more/scorekeeper/new';

  static String activeGame(String uuid) => '/more/scorekeeper/$uuid';
  static String grid(String uuid) => '/more/scorekeeper/$uuid/grid';
  static String export(String uuid) => '/more/scorekeeper/$uuid/export';
}
