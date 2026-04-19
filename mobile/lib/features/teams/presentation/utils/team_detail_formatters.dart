import 'package:bigfly_mobile/features/teams/application/teams_types.dart';

const List<String> _monthAbbr = <String>[
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
];

String teamDetailSegmentLabel(TeamDetailSegment segment) {
  return switch (segment) {
    TeamDetailSegment.overview => 'Overview',
    TeamDetailSegment.roster => 'Roster',
    TeamDetailSegment.schedule => 'Schedule',
    TeamDetailSegment.daily => 'Daily',
  };
}

String formatCompactDate(DateTime? date) {
  if (date == null) {
    return '—';
  }
  return '${_monthAbbr[date.month - 1]} ${date.day}';
}

String formatIsoDate(DateTime? date) {
  if (date == null) {
    return '—';
  }
  final y = date.year.toString().padLeft(4, '0');
  final m = date.month.toString().padLeft(2, '0');
  final d = date.day.toString().padLeft(2, '0');
  return '$y-$m-$d';
}

String formatWinPct(double value) => value.toStringAsFixed(3).replaceFirst('0', '.');
