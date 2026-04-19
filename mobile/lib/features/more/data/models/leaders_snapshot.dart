import 'package:bigfly_mobile/features/more/data/models/leaderboard_entry.dart';

class LeadersSnapshot {
  const LeadersSnapshot({
    required this.stat,
    required this.page,
    required this.perPage,
    required this.total,
    required this.entries,
  });

  final String stat;
  final int page;
  final int perPage;
  final int total;
  final List<LeaderboardEntry> entries;
}
