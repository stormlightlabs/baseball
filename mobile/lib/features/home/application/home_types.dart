enum HomeEntityType { players, teams, games, franchises, seasons }

extension HomeEntityTypeLabel on HomeEntityType {
  String get label => switch (this) {
    HomeEntityType.players => 'Players',
    HomeEntityType.teams => 'Teams',
    HomeEntityType.games => 'Games',
    HomeEntityType.franchises => 'Franchises',
    HomeEntityType.seasons => 'Seasons',
  };
}

class HomeSearchItem {
  const HomeSearchItem({required this.entity, required this.title, required this.subtitle, this.id});

  final HomeEntityType entity;
  final String title;
  final String subtitle;
  final String? id;
}
