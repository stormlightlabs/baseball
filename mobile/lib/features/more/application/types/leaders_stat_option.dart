class LeadersStatOption {
  const LeadersStatOption({required this.key, required this.label});

  final String key;
  final String label;
}

const List<LeadersStatOption> battingStatOptions = <LeadersStatOption>[
  LeadersStatOption(key: 'hr', label: 'HR'),
  LeadersStatOption(key: 'avg', label: 'AVG'),
  LeadersStatOption(key: 'rbi', label: 'RBI'),
  LeadersStatOption(key: 'ops', label: 'OPS'),
  LeadersStatOption(key: 'obp', label: 'OBP'),
  LeadersStatOption(key: 'slg', label: 'SLG'),
  LeadersStatOption(key: 'sb', label: 'SB'),
  LeadersStatOption(key: 'h', label: 'H'),
  LeadersStatOption(key: 'r', label: 'R'),
];

const List<LeadersStatOption> pitchingStatOptions = <LeadersStatOption>[
  LeadersStatOption(key: 'era', label: 'ERA'),
  LeadersStatOption(key: 'so', label: 'SO'),
  LeadersStatOption(key: 'w', label: 'W'),
  LeadersStatOption(key: 'sv', label: 'SV'),
  LeadersStatOption(key: 'whip', label: 'WHIP'),
  LeadersStatOption(key: 'k_per_9', label: 'K/9'),
  LeadersStatOption(key: 'hr', label: 'HR'),
  LeadersStatOption(key: 'bb', label: 'BB'),
];
