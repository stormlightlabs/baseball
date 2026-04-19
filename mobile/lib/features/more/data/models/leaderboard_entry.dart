class LeaderboardEntry {
  const LeaderboardEntry({
    required this.playerId,
    required this.playerName,
    required this.teamId,
    required this.league,
    required this.year,
    required this.rawValue,
    required this.displayValue,
  });

  final String playerId;
  final String playerName;
  final String teamId;
  final String league;
  final int? year;
  final double rawValue;
  final String displayValue;

  String get initials {
    final parts = playerName
        .split(' ')
        .where((part) => part.isNotEmpty)
        .take(2)
        .map((part) => part[0].toUpperCase())
        .toList(growable: false);
    if (parts.isEmpty) {
      return playerId.length >= 2 ? playerId.substring(0, 2).toUpperCase() : playerId.toUpperCase();
    }
    return parts.join();
  }

  String get subtitle {
    final items = <String>[
      playerId,
      if (teamId.isNotEmpty) teamId,
      if (league.isNotEmpty) league,
      if (year != null && year! > 0) year.toString(),
    ];
    return items.join(' · ');
  }
}
