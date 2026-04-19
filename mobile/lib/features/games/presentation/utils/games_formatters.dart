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

String formatGamesCompactDate(DateTime? date) {
  if (date == null) {
    return '—';
  }
  return '${_monthAbbr[date.month - 1]} ${date.day}';
}

String formatGamesLongDate(DateTime? date) {
  if (date == null) {
    return '—';
  }
  return '${_monthAbbr[date.month - 1]} ${date.day}, ${date.year}';
}

String formatAttendance(int? attendance) {
  if (attendance == null || attendance <= 0) {
    return '—';
  }
  return _formatWithCommas(attendance);
}

String formatDuration(int? durationMinutes) {
  if (durationMinutes == null || durationMinutes <= 0) {
    return '—';
  }

  final hours = durationMinutes ~/ 60;
  final minutes = durationMinutes % 60;

  if (hours <= 0) {
    return '${minutes}m';
  }

  return '${hours}h ${minutes.toString().padLeft(2, '0')}m';
}

String formatProbabilityPercent(double probability) {
  final bounded = probability.clamp(0.0, 1.0);
  return '${(bounded * 100).round()}%';
}

String _formatWithCommas(int value) {
  final raw = value.toString();
  final buffer = StringBuffer();

  for (var i = 0; i < raw.length; i++) {
    final indexFromEnd = raw.length - i;
    buffer.write(raw[i]);
    if (indexFromEnd > 1 && indexFromEnd % 3 == 1) {
      buffer.write(',');
    }
  }

  return buffer.toString();
}
