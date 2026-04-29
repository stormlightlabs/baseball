// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_database.dart';

// ignore_for_file: type=lint
class $AppSettingsTable extends AppSettings
    with TableInfo<$AppSettingsTable, AppSetting> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $AppSettingsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _keyMeta = const VerificationMeta('key');
  @override
  late final GeneratedColumn<String> key = GeneratedColumn<String>(
    'key',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _valueMeta = const VerificationMeta('value');
  @override
  late final GeneratedColumn<String> value = GeneratedColumn<String>(
    'value',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [key, value];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'app_settings';
  @override
  VerificationContext validateIntegrity(
    Insertable<AppSetting> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('key')) {
      context.handle(
        _keyMeta,
        key.isAcceptableOrUnknown(data['key']!, _keyMeta),
      );
    } else if (isInserting) {
      context.missing(_keyMeta);
    }
    if (data.containsKey('value')) {
      context.handle(
        _valueMeta,
        value.isAcceptableOrUnknown(data['value']!, _valueMeta),
      );
    } else if (isInserting) {
      context.missing(_valueMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {key};
  @override
  AppSetting map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return AppSetting(
      key: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}key'],
      )!,
      value: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}value'],
      )!,
    );
  }

  @override
  $AppSettingsTable createAlias(String alias) {
    return $AppSettingsTable(attachedDatabase, alias);
  }
}

class AppSetting extends DataClass implements Insertable<AppSetting> {
  final String key;
  final String value;
  const AppSetting({required this.key, required this.value});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['key'] = Variable<String>(key);
    map['value'] = Variable<String>(value);
    return map;
  }

  AppSettingsCompanion toCompanion(bool nullToAbsent) {
    return AppSettingsCompanion(key: Value(key), value: Value(value));
  }

  factory AppSetting.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return AppSetting(
      key: serializer.fromJson<String>(json['key']),
      value: serializer.fromJson<String>(json['value']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'key': serializer.toJson<String>(key),
      'value': serializer.toJson<String>(value),
    };
  }

  AppSetting copyWith({String? key, String? value}) =>
      AppSetting(key: key ?? this.key, value: value ?? this.value);
  AppSetting copyWithCompanion(AppSettingsCompanion data) {
    return AppSetting(
      key: data.key.present ? data.key.value : this.key,
      value: data.value.present ? data.value.value : this.value,
    );
  }

  @override
  String toString() {
    return (StringBuffer('AppSetting(')
          ..write('key: $key, ')
          ..write('value: $value')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(key, value);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is AppSetting &&
          other.key == this.key &&
          other.value == this.value);
}

class AppSettingsCompanion extends UpdateCompanion<AppSetting> {
  final Value<String> key;
  final Value<String> value;
  final Value<int> rowid;
  const AppSettingsCompanion({
    this.key = const Value.absent(),
    this.value = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  AppSettingsCompanion.insert({
    required String key,
    required String value,
    this.rowid = const Value.absent(),
  }) : key = Value(key),
       value = Value(value);
  static Insertable<AppSetting> custom({
    Expression<String>? key,
    Expression<String>? value,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (key != null) 'key': key,
      if (value != null) 'value': value,
      if (rowid != null) 'rowid': rowid,
    });
  }

  AppSettingsCompanion copyWith({
    Value<String>? key,
    Value<String>? value,
    Value<int>? rowid,
  }) {
    return AppSettingsCompanion(
      key: key ?? this.key,
      value: value ?? this.value,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (key.present) {
      map['key'] = Variable<String>(key.value);
    }
    if (value.present) {
      map['value'] = Variable<String>(value.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('AppSettingsCompanion(')
          ..write('key: $key, ')
          ..write('value: $value, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $RecentPlayersTable extends RecentPlayers
    with TableInfo<$RecentPlayersTable, RecentPlayer> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $RecentPlayersTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _playerIdMeta = const VerificationMeta(
    'playerId',
  );
  @override
  late final GeneratedColumn<String> playerId = GeneratedColumn<String>(
    'player_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _rankIndexMeta = const VerificationMeta(
    'rankIndex',
  );
  @override
  late final GeneratedColumn<int> rankIndex = GeneratedColumn<int>(
    'rank_index',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _updatedAtMeta = const VerificationMeta(
    'updatedAt',
  );
  @override
  late final GeneratedColumn<DateTime> updatedAt = GeneratedColumn<DateTime>(
    'updated_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    clientDefault: () => DateTime.now(),
  );
  @override
  List<GeneratedColumn> get $columns => [playerId, rankIndex, updatedAt];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'recent_players';
  @override
  VerificationContext validateIntegrity(
    Insertable<RecentPlayer> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('player_id')) {
      context.handle(
        _playerIdMeta,
        playerId.isAcceptableOrUnknown(data['player_id']!, _playerIdMeta),
      );
    } else if (isInserting) {
      context.missing(_playerIdMeta);
    }
    if (data.containsKey('rank_index')) {
      context.handle(
        _rankIndexMeta,
        rankIndex.isAcceptableOrUnknown(data['rank_index']!, _rankIndexMeta),
      );
    } else if (isInserting) {
      context.missing(_rankIndexMeta);
    }
    if (data.containsKey('updated_at')) {
      context.handle(
        _updatedAtMeta,
        updatedAt.isAcceptableOrUnknown(data['updated_at']!, _updatedAtMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {playerId};
  @override
  RecentPlayer map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return RecentPlayer(
      playerId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}player_id'],
      )!,
      rankIndex: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}rank_index'],
      )!,
      updatedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}updated_at'],
      )!,
    );
  }

  @override
  $RecentPlayersTable createAlias(String alias) {
    return $RecentPlayersTable(attachedDatabase, alias);
  }
}

class RecentPlayer extends DataClass implements Insertable<RecentPlayer> {
  final String playerId;
  final int rankIndex;
  final DateTime updatedAt;
  const RecentPlayer({
    required this.playerId,
    required this.rankIndex,
    required this.updatedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['player_id'] = Variable<String>(playerId);
    map['rank_index'] = Variable<int>(rankIndex);
    map['updated_at'] = Variable<DateTime>(updatedAt);
    return map;
  }

  RecentPlayersCompanion toCompanion(bool nullToAbsent) {
    return RecentPlayersCompanion(
      playerId: Value(playerId),
      rankIndex: Value(rankIndex),
      updatedAt: Value(updatedAt),
    );
  }

  factory RecentPlayer.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return RecentPlayer(
      playerId: serializer.fromJson<String>(json['playerId']),
      rankIndex: serializer.fromJson<int>(json['rankIndex']),
      updatedAt: serializer.fromJson<DateTime>(json['updatedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'playerId': serializer.toJson<String>(playerId),
      'rankIndex': serializer.toJson<int>(rankIndex),
      'updatedAt': serializer.toJson<DateTime>(updatedAt),
    };
  }

  RecentPlayer copyWith({
    String? playerId,
    int? rankIndex,
    DateTime? updatedAt,
  }) => RecentPlayer(
    playerId: playerId ?? this.playerId,
    rankIndex: rankIndex ?? this.rankIndex,
    updatedAt: updatedAt ?? this.updatedAt,
  );
  RecentPlayer copyWithCompanion(RecentPlayersCompanion data) {
    return RecentPlayer(
      playerId: data.playerId.present ? data.playerId.value : this.playerId,
      rankIndex: data.rankIndex.present ? data.rankIndex.value : this.rankIndex,
      updatedAt: data.updatedAt.present ? data.updatedAt.value : this.updatedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('RecentPlayer(')
          ..write('playerId: $playerId, ')
          ..write('rankIndex: $rankIndex, ')
          ..write('updatedAt: $updatedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(playerId, rankIndex, updatedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is RecentPlayer &&
          other.playerId == this.playerId &&
          other.rankIndex == this.rankIndex &&
          other.updatedAt == this.updatedAt);
}

class RecentPlayersCompanion extends UpdateCompanion<RecentPlayer> {
  final Value<String> playerId;
  final Value<int> rankIndex;
  final Value<DateTime> updatedAt;
  final Value<int> rowid;
  const RecentPlayersCompanion({
    this.playerId = const Value.absent(),
    this.rankIndex = const Value.absent(),
    this.updatedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  RecentPlayersCompanion.insert({
    required String playerId,
    required int rankIndex,
    this.updatedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : playerId = Value(playerId),
       rankIndex = Value(rankIndex);
  static Insertable<RecentPlayer> custom({
    Expression<String>? playerId,
    Expression<int>? rankIndex,
    Expression<DateTime>? updatedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (playerId != null) 'player_id': playerId,
      if (rankIndex != null) 'rank_index': rankIndex,
      if (updatedAt != null) 'updated_at': updatedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  RecentPlayersCompanion copyWith({
    Value<String>? playerId,
    Value<int>? rankIndex,
    Value<DateTime>? updatedAt,
    Value<int>? rowid,
  }) {
    return RecentPlayersCompanion(
      playerId: playerId ?? this.playerId,
      rankIndex: rankIndex ?? this.rankIndex,
      updatedAt: updatedAt ?? this.updatedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (playerId.present) {
      map['player_id'] = Variable<String>(playerId.value);
    }
    if (rankIndex.present) {
      map['rank_index'] = Variable<int>(rankIndex.value);
    }
    if (updatedAt.present) {
      map['updated_at'] = Variable<DateTime>(updatedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('RecentPlayersCompanion(')
          ..write('playerId: $playerId, ')
          ..write('rankIndex: $rankIndex, ')
          ..write('updatedAt: $updatedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $ScorecardGamesTable extends ScorecardGames
    with TableInfo<$ScorecardGamesTable, ScorecardGame> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ScorecardGamesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _uuidMeta = const VerificationMeta('uuid');
  @override
  late final GeneratedColumn<String> uuid = GeneratedColumn<String>(
    'uuid',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _awayTeamNameMeta = const VerificationMeta(
    'awayTeamName',
  );
  @override
  late final GeneratedColumn<String> awayTeamName = GeneratedColumn<String>(
    'away_team_name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _awayTeamAbbreviationMeta =
      const VerificationMeta('awayTeamAbbreviation');
  @override
  late final GeneratedColumn<String> awayTeamAbbreviation =
      GeneratedColumn<String>(
        'away_team_abbreviation',
        aliasedName,
        false,
        type: DriftSqlType.string,
        requiredDuringInsert: true,
      );
  static const VerificationMeta _homeTeamNameMeta = const VerificationMeta(
    'homeTeamName',
  );
  @override
  late final GeneratedColumn<String> homeTeamName = GeneratedColumn<String>(
    'home_team_name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _homeTeamAbbreviationMeta =
      const VerificationMeta('homeTeamAbbreviation');
  @override
  late final GeneratedColumn<String> homeTeamAbbreviation =
      GeneratedColumn<String>(
        'home_team_abbreviation',
        aliasedName,
        false,
        type: DriftSqlType.string,
        requiredDuringInsert: true,
      );
  static const VerificationMeta _venueMeta = const VerificationMeta('venue');
  @override
  late final GeneratedColumn<String> venue = GeneratedColumn<String>(
    'venue',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _gameDateMeta = const VerificationMeta(
    'gameDate',
  );
  @override
  late final GeneratedColumn<DateTime> gameDate = GeneratedColumn<DateTime>(
    'game_date',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
    'status',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _awayScoreMeta = const VerificationMeta(
    'awayScore',
  );
  @override
  late final GeneratedColumn<int> awayScore = GeneratedColumn<int>(
    'away_score',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _homeScoreMeta = const VerificationMeta(
    'homeScore',
  );
  @override
  late final GeneratedColumn<int> homeScore = GeneratedColumn<int>(
    'home_score',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _awayHitsMeta = const VerificationMeta(
    'awayHits',
  );
  @override
  late final GeneratedColumn<int> awayHits = GeneratedColumn<int>(
    'away_hits',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _homeHitsMeta = const VerificationMeta(
    'homeHits',
  );
  @override
  late final GeneratedColumn<int> homeHits = GeneratedColumn<int>(
    'home_hits',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _awayErrorsMeta = const VerificationMeta(
    'awayErrors',
  );
  @override
  late final GeneratedColumn<int> awayErrors = GeneratedColumn<int>(
    'away_errors',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _homeErrorsMeta = const VerificationMeta(
    'homeErrors',
  );
  @override
  late final GeneratedColumn<int> homeErrors = GeneratedColumn<int>(
    'home_errors',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _pitchCountMeta = const VerificationMeta(
    'pitchCount',
  );
  @override
  late final GeneratedColumn<int> pitchCount = GeneratedColumn<int>(
    'pitch_count',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _currentInningMeta = const VerificationMeta(
    'currentInning',
  );
  @override
  late final GeneratedColumn<int> currentInning = GeneratedColumn<int>(
    'current_inning',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(1),
  );
  static const VerificationMeta _currentHalfMeta = const VerificationMeta(
    'currentHalf',
  );
  @override
  late final GeneratedColumn<String> currentHalf = GeneratedColumn<String>(
    'current_half',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _outsMeta = const VerificationMeta('outs');
  @override
  late final GeneratedColumn<int> outs = GeneratedColumn<int>(
    'outs',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _ballsMeta = const VerificationMeta('balls');
  @override
  late final GeneratedColumn<int> balls = GeneratedColumn<int>(
    'balls',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _strikesMeta = const VerificationMeta(
    'strikes',
  );
  @override
  late final GeneratedColumn<int> strikes = GeneratedColumn<int>(
    'strikes',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _awayBatterIndexMeta = const VerificationMeta(
    'awayBatterIndex',
  );
  @override
  late final GeneratedColumn<int> awayBatterIndex = GeneratedColumn<int>(
    'away_batter_index',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _homeBatterIndexMeta = const VerificationMeta(
    'homeBatterIndex',
  );
  @override
  late final GeneratedColumn<int> homeBatterIndex = GeneratedColumn<int>(
    'home_batter_index',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  static const VerificationMeta _lastPlaySequenceMeta = const VerificationMeta(
    'lastPlaySequence',
  );
  @override
  late final GeneratedColumn<int> lastPlaySequence = GeneratedColumn<int>(
    'last_play_sequence',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta(
    'createdAt',
  );
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    clientDefault: () => DateTime.now(),
  );
  static const VerificationMeta _lastModifiedAtMeta = const VerificationMeta(
    'lastModifiedAt',
  );
  @override
  late final GeneratedColumn<DateTime> lastModifiedAt =
      GeneratedColumn<DateTime>(
        'last_modified_at',
        aliasedName,
        false,
        type: DriftSqlType.dateTime,
        requiredDuringInsert: false,
        clientDefault: () => DateTime.now(),
      );
  @override
  List<GeneratedColumn> get $columns => [
    uuid,
    awayTeamName,
    awayTeamAbbreviation,
    homeTeamName,
    homeTeamAbbreviation,
    venue,
    gameDate,
    status,
    awayScore,
    homeScore,
    awayHits,
    homeHits,
    awayErrors,
    homeErrors,
    pitchCount,
    currentInning,
    currentHalf,
    outs,
    balls,
    strikes,
    awayBatterIndex,
    homeBatterIndex,
    lastPlaySequence,
    createdAt,
    lastModifiedAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'scorecard_games';
  @override
  VerificationContext validateIntegrity(
    Insertable<ScorecardGame> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('uuid')) {
      context.handle(
        _uuidMeta,
        uuid.isAcceptableOrUnknown(data['uuid']!, _uuidMeta),
      );
    } else if (isInserting) {
      context.missing(_uuidMeta);
    }
    if (data.containsKey('away_team_name')) {
      context.handle(
        _awayTeamNameMeta,
        awayTeamName.isAcceptableOrUnknown(
          data['away_team_name']!,
          _awayTeamNameMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_awayTeamNameMeta);
    }
    if (data.containsKey('away_team_abbreviation')) {
      context.handle(
        _awayTeamAbbreviationMeta,
        awayTeamAbbreviation.isAcceptableOrUnknown(
          data['away_team_abbreviation']!,
          _awayTeamAbbreviationMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_awayTeamAbbreviationMeta);
    }
    if (data.containsKey('home_team_name')) {
      context.handle(
        _homeTeamNameMeta,
        homeTeamName.isAcceptableOrUnknown(
          data['home_team_name']!,
          _homeTeamNameMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_homeTeamNameMeta);
    }
    if (data.containsKey('home_team_abbreviation')) {
      context.handle(
        _homeTeamAbbreviationMeta,
        homeTeamAbbreviation.isAcceptableOrUnknown(
          data['home_team_abbreviation']!,
          _homeTeamAbbreviationMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_homeTeamAbbreviationMeta);
    }
    if (data.containsKey('venue')) {
      context.handle(
        _venueMeta,
        venue.isAcceptableOrUnknown(data['venue']!, _venueMeta),
      );
    }
    if (data.containsKey('game_date')) {
      context.handle(
        _gameDateMeta,
        gameDate.isAcceptableOrUnknown(data['game_date']!, _gameDateMeta),
      );
    } else if (isInserting) {
      context.missing(_gameDateMeta);
    }
    if (data.containsKey('status')) {
      context.handle(
        _statusMeta,
        status.isAcceptableOrUnknown(data['status']!, _statusMeta),
      );
    } else if (isInserting) {
      context.missing(_statusMeta);
    }
    if (data.containsKey('away_score')) {
      context.handle(
        _awayScoreMeta,
        awayScore.isAcceptableOrUnknown(data['away_score']!, _awayScoreMeta),
      );
    }
    if (data.containsKey('home_score')) {
      context.handle(
        _homeScoreMeta,
        homeScore.isAcceptableOrUnknown(data['home_score']!, _homeScoreMeta),
      );
    }
    if (data.containsKey('away_hits')) {
      context.handle(
        _awayHitsMeta,
        awayHits.isAcceptableOrUnknown(data['away_hits']!, _awayHitsMeta),
      );
    }
    if (data.containsKey('home_hits')) {
      context.handle(
        _homeHitsMeta,
        homeHits.isAcceptableOrUnknown(data['home_hits']!, _homeHitsMeta),
      );
    }
    if (data.containsKey('away_errors')) {
      context.handle(
        _awayErrorsMeta,
        awayErrors.isAcceptableOrUnknown(data['away_errors']!, _awayErrorsMeta),
      );
    }
    if (data.containsKey('home_errors')) {
      context.handle(
        _homeErrorsMeta,
        homeErrors.isAcceptableOrUnknown(data['home_errors']!, _homeErrorsMeta),
      );
    }
    if (data.containsKey('pitch_count')) {
      context.handle(
        _pitchCountMeta,
        pitchCount.isAcceptableOrUnknown(data['pitch_count']!, _pitchCountMeta),
      );
    }
    if (data.containsKey('current_inning')) {
      context.handle(
        _currentInningMeta,
        currentInning.isAcceptableOrUnknown(
          data['current_inning']!,
          _currentInningMeta,
        ),
      );
    }
    if (data.containsKey('current_half')) {
      context.handle(
        _currentHalfMeta,
        currentHalf.isAcceptableOrUnknown(
          data['current_half']!,
          _currentHalfMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_currentHalfMeta);
    }
    if (data.containsKey('outs')) {
      context.handle(
        _outsMeta,
        outs.isAcceptableOrUnknown(data['outs']!, _outsMeta),
      );
    }
    if (data.containsKey('balls')) {
      context.handle(
        _ballsMeta,
        balls.isAcceptableOrUnknown(data['balls']!, _ballsMeta),
      );
    }
    if (data.containsKey('strikes')) {
      context.handle(
        _strikesMeta,
        strikes.isAcceptableOrUnknown(data['strikes']!, _strikesMeta),
      );
    }
    if (data.containsKey('away_batter_index')) {
      context.handle(
        _awayBatterIndexMeta,
        awayBatterIndex.isAcceptableOrUnknown(
          data['away_batter_index']!,
          _awayBatterIndexMeta,
        ),
      );
    }
    if (data.containsKey('home_batter_index')) {
      context.handle(
        _homeBatterIndexMeta,
        homeBatterIndex.isAcceptableOrUnknown(
          data['home_batter_index']!,
          _homeBatterIndexMeta,
        ),
      );
    }
    if (data.containsKey('last_play_sequence')) {
      context.handle(
        _lastPlaySequenceMeta,
        lastPlaySequence.isAcceptableOrUnknown(
          data['last_play_sequence']!,
          _lastPlaySequenceMeta,
        ),
      );
    }
    if (data.containsKey('created_at')) {
      context.handle(
        _createdAtMeta,
        createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta),
      );
    }
    if (data.containsKey('last_modified_at')) {
      context.handle(
        _lastModifiedAtMeta,
        lastModifiedAt.isAcceptableOrUnknown(
          data['last_modified_at']!,
          _lastModifiedAtMeta,
        ),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {uuid};
  @override
  ScorecardGame map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ScorecardGame(
      uuid: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}uuid'],
      )!,
      awayTeamName: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}away_team_name'],
      )!,
      awayTeamAbbreviation: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}away_team_abbreviation'],
      )!,
      homeTeamName: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}home_team_name'],
      )!,
      homeTeamAbbreviation: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}home_team_abbreviation'],
      )!,
      venue: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}venue'],
      ),
      gameDate: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}game_date'],
      )!,
      status: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}status'],
      )!,
      awayScore: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}away_score'],
      )!,
      homeScore: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}home_score'],
      )!,
      awayHits: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}away_hits'],
      )!,
      homeHits: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}home_hits'],
      )!,
      awayErrors: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}away_errors'],
      )!,
      homeErrors: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}home_errors'],
      )!,
      pitchCount: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}pitch_count'],
      )!,
      currentInning: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}current_inning'],
      )!,
      currentHalf: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}current_half'],
      )!,
      outs: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}outs'],
      )!,
      balls: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}balls'],
      )!,
      strikes: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}strikes'],
      )!,
      awayBatterIndex: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}away_batter_index'],
      )!,
      homeBatterIndex: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}home_batter_index'],
      )!,
      lastPlaySequence: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}last_play_sequence'],
      ),
      createdAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}created_at'],
      )!,
      lastModifiedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}last_modified_at'],
      )!,
    );
  }

  @override
  $ScorecardGamesTable createAlias(String alias) {
    return $ScorecardGamesTable(attachedDatabase, alias);
  }
}

class ScorecardGame extends DataClass implements Insertable<ScorecardGame> {
  final String uuid;
  final String awayTeamName;
  final String awayTeamAbbreviation;
  final String homeTeamName;
  final String homeTeamAbbreviation;
  final String? venue;
  final DateTime gameDate;
  final String status;
  final int awayScore;
  final int homeScore;
  final int awayHits;
  final int homeHits;
  final int awayErrors;
  final int homeErrors;
  final int pitchCount;
  final int currentInning;
  final String currentHalf;
  final int outs;
  final int balls;
  final int strikes;
  final int awayBatterIndex;
  final int homeBatterIndex;
  final int? lastPlaySequence;
  final DateTime createdAt;
  final DateTime lastModifiedAt;
  const ScorecardGame({
    required this.uuid,
    required this.awayTeamName,
    required this.awayTeamAbbreviation,
    required this.homeTeamName,
    required this.homeTeamAbbreviation,
    this.venue,
    required this.gameDate,
    required this.status,
    required this.awayScore,
    required this.homeScore,
    required this.awayHits,
    required this.homeHits,
    required this.awayErrors,
    required this.homeErrors,
    required this.pitchCount,
    required this.currentInning,
    required this.currentHalf,
    required this.outs,
    required this.balls,
    required this.strikes,
    required this.awayBatterIndex,
    required this.homeBatterIndex,
    this.lastPlaySequence,
    required this.createdAt,
    required this.lastModifiedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['uuid'] = Variable<String>(uuid);
    map['away_team_name'] = Variable<String>(awayTeamName);
    map['away_team_abbreviation'] = Variable<String>(awayTeamAbbreviation);
    map['home_team_name'] = Variable<String>(homeTeamName);
    map['home_team_abbreviation'] = Variable<String>(homeTeamAbbreviation);
    if (!nullToAbsent || venue != null) {
      map['venue'] = Variable<String>(venue);
    }
    map['game_date'] = Variable<DateTime>(gameDate);
    map['status'] = Variable<String>(status);
    map['away_score'] = Variable<int>(awayScore);
    map['home_score'] = Variable<int>(homeScore);
    map['away_hits'] = Variable<int>(awayHits);
    map['home_hits'] = Variable<int>(homeHits);
    map['away_errors'] = Variable<int>(awayErrors);
    map['home_errors'] = Variable<int>(homeErrors);
    map['pitch_count'] = Variable<int>(pitchCount);
    map['current_inning'] = Variable<int>(currentInning);
    map['current_half'] = Variable<String>(currentHalf);
    map['outs'] = Variable<int>(outs);
    map['balls'] = Variable<int>(balls);
    map['strikes'] = Variable<int>(strikes);
    map['away_batter_index'] = Variable<int>(awayBatterIndex);
    map['home_batter_index'] = Variable<int>(homeBatterIndex);
    if (!nullToAbsent || lastPlaySequence != null) {
      map['last_play_sequence'] = Variable<int>(lastPlaySequence);
    }
    map['created_at'] = Variable<DateTime>(createdAt);
    map['last_modified_at'] = Variable<DateTime>(lastModifiedAt);
    return map;
  }

  ScorecardGamesCompanion toCompanion(bool nullToAbsent) {
    return ScorecardGamesCompanion(
      uuid: Value(uuid),
      awayTeamName: Value(awayTeamName),
      awayTeamAbbreviation: Value(awayTeamAbbreviation),
      homeTeamName: Value(homeTeamName),
      homeTeamAbbreviation: Value(homeTeamAbbreviation),
      venue: venue == null && nullToAbsent
          ? const Value.absent()
          : Value(venue),
      gameDate: Value(gameDate),
      status: Value(status),
      awayScore: Value(awayScore),
      homeScore: Value(homeScore),
      awayHits: Value(awayHits),
      homeHits: Value(homeHits),
      awayErrors: Value(awayErrors),
      homeErrors: Value(homeErrors),
      pitchCount: Value(pitchCount),
      currentInning: Value(currentInning),
      currentHalf: Value(currentHalf),
      outs: Value(outs),
      balls: Value(balls),
      strikes: Value(strikes),
      awayBatterIndex: Value(awayBatterIndex),
      homeBatterIndex: Value(homeBatterIndex),
      lastPlaySequence: lastPlaySequence == null && nullToAbsent
          ? const Value.absent()
          : Value(lastPlaySequence),
      createdAt: Value(createdAt),
      lastModifiedAt: Value(lastModifiedAt),
    );
  }

  factory ScorecardGame.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ScorecardGame(
      uuid: serializer.fromJson<String>(json['uuid']),
      awayTeamName: serializer.fromJson<String>(json['awayTeamName']),
      awayTeamAbbreviation: serializer.fromJson<String>(
        json['awayTeamAbbreviation'],
      ),
      homeTeamName: serializer.fromJson<String>(json['homeTeamName']),
      homeTeamAbbreviation: serializer.fromJson<String>(
        json['homeTeamAbbreviation'],
      ),
      venue: serializer.fromJson<String?>(json['venue']),
      gameDate: serializer.fromJson<DateTime>(json['gameDate']),
      status: serializer.fromJson<String>(json['status']),
      awayScore: serializer.fromJson<int>(json['awayScore']),
      homeScore: serializer.fromJson<int>(json['homeScore']),
      awayHits: serializer.fromJson<int>(json['awayHits']),
      homeHits: serializer.fromJson<int>(json['homeHits']),
      awayErrors: serializer.fromJson<int>(json['awayErrors']),
      homeErrors: serializer.fromJson<int>(json['homeErrors']),
      pitchCount: serializer.fromJson<int>(json['pitchCount']),
      currentInning: serializer.fromJson<int>(json['currentInning']),
      currentHalf: serializer.fromJson<String>(json['currentHalf']),
      outs: serializer.fromJson<int>(json['outs']),
      balls: serializer.fromJson<int>(json['balls']),
      strikes: serializer.fromJson<int>(json['strikes']),
      awayBatterIndex: serializer.fromJson<int>(json['awayBatterIndex']),
      homeBatterIndex: serializer.fromJson<int>(json['homeBatterIndex']),
      lastPlaySequence: serializer.fromJson<int?>(json['lastPlaySequence']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      lastModifiedAt: serializer.fromJson<DateTime>(json['lastModifiedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'uuid': serializer.toJson<String>(uuid),
      'awayTeamName': serializer.toJson<String>(awayTeamName),
      'awayTeamAbbreviation': serializer.toJson<String>(awayTeamAbbreviation),
      'homeTeamName': serializer.toJson<String>(homeTeamName),
      'homeTeamAbbreviation': serializer.toJson<String>(homeTeamAbbreviation),
      'venue': serializer.toJson<String?>(venue),
      'gameDate': serializer.toJson<DateTime>(gameDate),
      'status': serializer.toJson<String>(status),
      'awayScore': serializer.toJson<int>(awayScore),
      'homeScore': serializer.toJson<int>(homeScore),
      'awayHits': serializer.toJson<int>(awayHits),
      'homeHits': serializer.toJson<int>(homeHits),
      'awayErrors': serializer.toJson<int>(awayErrors),
      'homeErrors': serializer.toJson<int>(homeErrors),
      'pitchCount': serializer.toJson<int>(pitchCount),
      'currentInning': serializer.toJson<int>(currentInning),
      'currentHalf': serializer.toJson<String>(currentHalf),
      'outs': serializer.toJson<int>(outs),
      'balls': serializer.toJson<int>(balls),
      'strikes': serializer.toJson<int>(strikes),
      'awayBatterIndex': serializer.toJson<int>(awayBatterIndex),
      'homeBatterIndex': serializer.toJson<int>(homeBatterIndex),
      'lastPlaySequence': serializer.toJson<int?>(lastPlaySequence),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'lastModifiedAt': serializer.toJson<DateTime>(lastModifiedAt),
    };
  }

  ScorecardGame copyWith({
    String? uuid,
    String? awayTeamName,
    String? awayTeamAbbreviation,
    String? homeTeamName,
    String? homeTeamAbbreviation,
    Value<String?> venue = const Value.absent(),
    DateTime? gameDate,
    String? status,
    int? awayScore,
    int? homeScore,
    int? awayHits,
    int? homeHits,
    int? awayErrors,
    int? homeErrors,
    int? pitchCount,
    int? currentInning,
    String? currentHalf,
    int? outs,
    int? balls,
    int? strikes,
    int? awayBatterIndex,
    int? homeBatterIndex,
    Value<int?> lastPlaySequence = const Value.absent(),
    DateTime? createdAt,
    DateTime? lastModifiedAt,
  }) => ScorecardGame(
    uuid: uuid ?? this.uuid,
    awayTeamName: awayTeamName ?? this.awayTeamName,
    awayTeamAbbreviation: awayTeamAbbreviation ?? this.awayTeamAbbreviation,
    homeTeamName: homeTeamName ?? this.homeTeamName,
    homeTeamAbbreviation: homeTeamAbbreviation ?? this.homeTeamAbbreviation,
    venue: venue.present ? venue.value : this.venue,
    gameDate: gameDate ?? this.gameDate,
    status: status ?? this.status,
    awayScore: awayScore ?? this.awayScore,
    homeScore: homeScore ?? this.homeScore,
    awayHits: awayHits ?? this.awayHits,
    homeHits: homeHits ?? this.homeHits,
    awayErrors: awayErrors ?? this.awayErrors,
    homeErrors: homeErrors ?? this.homeErrors,
    pitchCount: pitchCount ?? this.pitchCount,
    currentInning: currentInning ?? this.currentInning,
    currentHalf: currentHalf ?? this.currentHalf,
    outs: outs ?? this.outs,
    balls: balls ?? this.balls,
    strikes: strikes ?? this.strikes,
    awayBatterIndex: awayBatterIndex ?? this.awayBatterIndex,
    homeBatterIndex: homeBatterIndex ?? this.homeBatterIndex,
    lastPlaySequence: lastPlaySequence.present
        ? lastPlaySequence.value
        : this.lastPlaySequence,
    createdAt: createdAt ?? this.createdAt,
    lastModifiedAt: lastModifiedAt ?? this.lastModifiedAt,
  );
  ScorecardGame copyWithCompanion(ScorecardGamesCompanion data) {
    return ScorecardGame(
      uuid: data.uuid.present ? data.uuid.value : this.uuid,
      awayTeamName: data.awayTeamName.present
          ? data.awayTeamName.value
          : this.awayTeamName,
      awayTeamAbbreviation: data.awayTeamAbbreviation.present
          ? data.awayTeamAbbreviation.value
          : this.awayTeamAbbreviation,
      homeTeamName: data.homeTeamName.present
          ? data.homeTeamName.value
          : this.homeTeamName,
      homeTeamAbbreviation: data.homeTeamAbbreviation.present
          ? data.homeTeamAbbreviation.value
          : this.homeTeamAbbreviation,
      venue: data.venue.present ? data.venue.value : this.venue,
      gameDate: data.gameDate.present ? data.gameDate.value : this.gameDate,
      status: data.status.present ? data.status.value : this.status,
      awayScore: data.awayScore.present ? data.awayScore.value : this.awayScore,
      homeScore: data.homeScore.present ? data.homeScore.value : this.homeScore,
      awayHits: data.awayHits.present ? data.awayHits.value : this.awayHits,
      homeHits: data.homeHits.present ? data.homeHits.value : this.homeHits,
      awayErrors: data.awayErrors.present
          ? data.awayErrors.value
          : this.awayErrors,
      homeErrors: data.homeErrors.present
          ? data.homeErrors.value
          : this.homeErrors,
      pitchCount: data.pitchCount.present
          ? data.pitchCount.value
          : this.pitchCount,
      currentInning: data.currentInning.present
          ? data.currentInning.value
          : this.currentInning,
      currentHalf: data.currentHalf.present
          ? data.currentHalf.value
          : this.currentHalf,
      outs: data.outs.present ? data.outs.value : this.outs,
      balls: data.balls.present ? data.balls.value : this.balls,
      strikes: data.strikes.present ? data.strikes.value : this.strikes,
      awayBatterIndex: data.awayBatterIndex.present
          ? data.awayBatterIndex.value
          : this.awayBatterIndex,
      homeBatterIndex: data.homeBatterIndex.present
          ? data.homeBatterIndex.value
          : this.homeBatterIndex,
      lastPlaySequence: data.lastPlaySequence.present
          ? data.lastPlaySequence.value
          : this.lastPlaySequence,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      lastModifiedAt: data.lastModifiedAt.present
          ? data.lastModifiedAt.value
          : this.lastModifiedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardGame(')
          ..write('uuid: $uuid, ')
          ..write('awayTeamName: $awayTeamName, ')
          ..write('awayTeamAbbreviation: $awayTeamAbbreviation, ')
          ..write('homeTeamName: $homeTeamName, ')
          ..write('homeTeamAbbreviation: $homeTeamAbbreviation, ')
          ..write('venue: $venue, ')
          ..write('gameDate: $gameDate, ')
          ..write('status: $status, ')
          ..write('awayScore: $awayScore, ')
          ..write('homeScore: $homeScore, ')
          ..write('awayHits: $awayHits, ')
          ..write('homeHits: $homeHits, ')
          ..write('awayErrors: $awayErrors, ')
          ..write('homeErrors: $homeErrors, ')
          ..write('pitchCount: $pitchCount, ')
          ..write('currentInning: $currentInning, ')
          ..write('currentHalf: $currentHalf, ')
          ..write('outs: $outs, ')
          ..write('balls: $balls, ')
          ..write('strikes: $strikes, ')
          ..write('awayBatterIndex: $awayBatterIndex, ')
          ..write('homeBatterIndex: $homeBatterIndex, ')
          ..write('lastPlaySequence: $lastPlaySequence, ')
          ..write('createdAt: $createdAt, ')
          ..write('lastModifiedAt: $lastModifiedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hashAll([
    uuid,
    awayTeamName,
    awayTeamAbbreviation,
    homeTeamName,
    homeTeamAbbreviation,
    venue,
    gameDate,
    status,
    awayScore,
    homeScore,
    awayHits,
    homeHits,
    awayErrors,
    homeErrors,
    pitchCount,
    currentInning,
    currentHalf,
    outs,
    balls,
    strikes,
    awayBatterIndex,
    homeBatterIndex,
    lastPlaySequence,
    createdAt,
    lastModifiedAt,
  ]);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ScorecardGame &&
          other.uuid == this.uuid &&
          other.awayTeamName == this.awayTeamName &&
          other.awayTeamAbbreviation == this.awayTeamAbbreviation &&
          other.homeTeamName == this.homeTeamName &&
          other.homeTeamAbbreviation == this.homeTeamAbbreviation &&
          other.venue == this.venue &&
          other.gameDate == this.gameDate &&
          other.status == this.status &&
          other.awayScore == this.awayScore &&
          other.homeScore == this.homeScore &&
          other.awayHits == this.awayHits &&
          other.homeHits == this.homeHits &&
          other.awayErrors == this.awayErrors &&
          other.homeErrors == this.homeErrors &&
          other.pitchCount == this.pitchCount &&
          other.currentInning == this.currentInning &&
          other.currentHalf == this.currentHalf &&
          other.outs == this.outs &&
          other.balls == this.balls &&
          other.strikes == this.strikes &&
          other.awayBatterIndex == this.awayBatterIndex &&
          other.homeBatterIndex == this.homeBatterIndex &&
          other.lastPlaySequence == this.lastPlaySequence &&
          other.createdAt == this.createdAt &&
          other.lastModifiedAt == this.lastModifiedAt);
}

class ScorecardGamesCompanion extends UpdateCompanion<ScorecardGame> {
  final Value<String> uuid;
  final Value<String> awayTeamName;
  final Value<String> awayTeamAbbreviation;
  final Value<String> homeTeamName;
  final Value<String> homeTeamAbbreviation;
  final Value<String?> venue;
  final Value<DateTime> gameDate;
  final Value<String> status;
  final Value<int> awayScore;
  final Value<int> homeScore;
  final Value<int> awayHits;
  final Value<int> homeHits;
  final Value<int> awayErrors;
  final Value<int> homeErrors;
  final Value<int> pitchCount;
  final Value<int> currentInning;
  final Value<String> currentHalf;
  final Value<int> outs;
  final Value<int> balls;
  final Value<int> strikes;
  final Value<int> awayBatterIndex;
  final Value<int> homeBatterIndex;
  final Value<int?> lastPlaySequence;
  final Value<DateTime> createdAt;
  final Value<DateTime> lastModifiedAt;
  final Value<int> rowid;
  const ScorecardGamesCompanion({
    this.uuid = const Value.absent(),
    this.awayTeamName = const Value.absent(),
    this.awayTeamAbbreviation = const Value.absent(),
    this.homeTeamName = const Value.absent(),
    this.homeTeamAbbreviation = const Value.absent(),
    this.venue = const Value.absent(),
    this.gameDate = const Value.absent(),
    this.status = const Value.absent(),
    this.awayScore = const Value.absent(),
    this.homeScore = const Value.absent(),
    this.awayHits = const Value.absent(),
    this.homeHits = const Value.absent(),
    this.awayErrors = const Value.absent(),
    this.homeErrors = const Value.absent(),
    this.pitchCount = const Value.absent(),
    this.currentInning = const Value.absent(),
    this.currentHalf = const Value.absent(),
    this.outs = const Value.absent(),
    this.balls = const Value.absent(),
    this.strikes = const Value.absent(),
    this.awayBatterIndex = const Value.absent(),
    this.homeBatterIndex = const Value.absent(),
    this.lastPlaySequence = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.lastModifiedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  ScorecardGamesCompanion.insert({
    required String uuid,
    required String awayTeamName,
    required String awayTeamAbbreviation,
    required String homeTeamName,
    required String homeTeamAbbreviation,
    this.venue = const Value.absent(),
    required DateTime gameDate,
    required String status,
    this.awayScore = const Value.absent(),
    this.homeScore = const Value.absent(),
    this.awayHits = const Value.absent(),
    this.homeHits = const Value.absent(),
    this.awayErrors = const Value.absent(),
    this.homeErrors = const Value.absent(),
    this.pitchCount = const Value.absent(),
    this.currentInning = const Value.absent(),
    required String currentHalf,
    this.outs = const Value.absent(),
    this.balls = const Value.absent(),
    this.strikes = const Value.absent(),
    this.awayBatterIndex = const Value.absent(),
    this.homeBatterIndex = const Value.absent(),
    this.lastPlaySequence = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.lastModifiedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : uuid = Value(uuid),
       awayTeamName = Value(awayTeamName),
       awayTeamAbbreviation = Value(awayTeamAbbreviation),
       homeTeamName = Value(homeTeamName),
       homeTeamAbbreviation = Value(homeTeamAbbreviation),
       gameDate = Value(gameDate),
       status = Value(status),
       currentHalf = Value(currentHalf);
  static Insertable<ScorecardGame> custom({
    Expression<String>? uuid,
    Expression<String>? awayTeamName,
    Expression<String>? awayTeamAbbreviation,
    Expression<String>? homeTeamName,
    Expression<String>? homeTeamAbbreviation,
    Expression<String>? venue,
    Expression<DateTime>? gameDate,
    Expression<String>? status,
    Expression<int>? awayScore,
    Expression<int>? homeScore,
    Expression<int>? awayHits,
    Expression<int>? homeHits,
    Expression<int>? awayErrors,
    Expression<int>? homeErrors,
    Expression<int>? pitchCount,
    Expression<int>? currentInning,
    Expression<String>? currentHalf,
    Expression<int>? outs,
    Expression<int>? balls,
    Expression<int>? strikes,
    Expression<int>? awayBatterIndex,
    Expression<int>? homeBatterIndex,
    Expression<int>? lastPlaySequence,
    Expression<DateTime>? createdAt,
    Expression<DateTime>? lastModifiedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (uuid != null) 'uuid': uuid,
      if (awayTeamName != null) 'away_team_name': awayTeamName,
      if (awayTeamAbbreviation != null)
        'away_team_abbreviation': awayTeamAbbreviation,
      if (homeTeamName != null) 'home_team_name': homeTeamName,
      if (homeTeamAbbreviation != null)
        'home_team_abbreviation': homeTeamAbbreviation,
      if (venue != null) 'venue': venue,
      if (gameDate != null) 'game_date': gameDate,
      if (status != null) 'status': status,
      if (awayScore != null) 'away_score': awayScore,
      if (homeScore != null) 'home_score': homeScore,
      if (awayHits != null) 'away_hits': awayHits,
      if (homeHits != null) 'home_hits': homeHits,
      if (awayErrors != null) 'away_errors': awayErrors,
      if (homeErrors != null) 'home_errors': homeErrors,
      if (pitchCount != null) 'pitch_count': pitchCount,
      if (currentInning != null) 'current_inning': currentInning,
      if (currentHalf != null) 'current_half': currentHalf,
      if (outs != null) 'outs': outs,
      if (balls != null) 'balls': balls,
      if (strikes != null) 'strikes': strikes,
      if (awayBatterIndex != null) 'away_batter_index': awayBatterIndex,
      if (homeBatterIndex != null) 'home_batter_index': homeBatterIndex,
      if (lastPlaySequence != null) 'last_play_sequence': lastPlaySequence,
      if (createdAt != null) 'created_at': createdAt,
      if (lastModifiedAt != null) 'last_modified_at': lastModifiedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  ScorecardGamesCompanion copyWith({
    Value<String>? uuid,
    Value<String>? awayTeamName,
    Value<String>? awayTeamAbbreviation,
    Value<String>? homeTeamName,
    Value<String>? homeTeamAbbreviation,
    Value<String?>? venue,
    Value<DateTime>? gameDate,
    Value<String>? status,
    Value<int>? awayScore,
    Value<int>? homeScore,
    Value<int>? awayHits,
    Value<int>? homeHits,
    Value<int>? awayErrors,
    Value<int>? homeErrors,
    Value<int>? pitchCount,
    Value<int>? currentInning,
    Value<String>? currentHalf,
    Value<int>? outs,
    Value<int>? balls,
    Value<int>? strikes,
    Value<int>? awayBatterIndex,
    Value<int>? homeBatterIndex,
    Value<int?>? lastPlaySequence,
    Value<DateTime>? createdAt,
    Value<DateTime>? lastModifiedAt,
    Value<int>? rowid,
  }) {
    return ScorecardGamesCompanion(
      uuid: uuid ?? this.uuid,
      awayTeamName: awayTeamName ?? this.awayTeamName,
      awayTeamAbbreviation: awayTeamAbbreviation ?? this.awayTeamAbbreviation,
      homeTeamName: homeTeamName ?? this.homeTeamName,
      homeTeamAbbreviation: homeTeamAbbreviation ?? this.homeTeamAbbreviation,
      venue: venue ?? this.venue,
      gameDate: gameDate ?? this.gameDate,
      status: status ?? this.status,
      awayScore: awayScore ?? this.awayScore,
      homeScore: homeScore ?? this.homeScore,
      awayHits: awayHits ?? this.awayHits,
      homeHits: homeHits ?? this.homeHits,
      awayErrors: awayErrors ?? this.awayErrors,
      homeErrors: homeErrors ?? this.homeErrors,
      pitchCount: pitchCount ?? this.pitchCount,
      currentInning: currentInning ?? this.currentInning,
      currentHalf: currentHalf ?? this.currentHalf,
      outs: outs ?? this.outs,
      balls: balls ?? this.balls,
      strikes: strikes ?? this.strikes,
      awayBatterIndex: awayBatterIndex ?? this.awayBatterIndex,
      homeBatterIndex: homeBatterIndex ?? this.homeBatterIndex,
      lastPlaySequence: lastPlaySequence ?? this.lastPlaySequence,
      createdAt: createdAt ?? this.createdAt,
      lastModifiedAt: lastModifiedAt ?? this.lastModifiedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (uuid.present) {
      map['uuid'] = Variable<String>(uuid.value);
    }
    if (awayTeamName.present) {
      map['away_team_name'] = Variable<String>(awayTeamName.value);
    }
    if (awayTeamAbbreviation.present) {
      map['away_team_abbreviation'] = Variable<String>(
        awayTeamAbbreviation.value,
      );
    }
    if (homeTeamName.present) {
      map['home_team_name'] = Variable<String>(homeTeamName.value);
    }
    if (homeTeamAbbreviation.present) {
      map['home_team_abbreviation'] = Variable<String>(
        homeTeamAbbreviation.value,
      );
    }
    if (venue.present) {
      map['venue'] = Variable<String>(venue.value);
    }
    if (gameDate.present) {
      map['game_date'] = Variable<DateTime>(gameDate.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (awayScore.present) {
      map['away_score'] = Variable<int>(awayScore.value);
    }
    if (homeScore.present) {
      map['home_score'] = Variable<int>(homeScore.value);
    }
    if (awayHits.present) {
      map['away_hits'] = Variable<int>(awayHits.value);
    }
    if (homeHits.present) {
      map['home_hits'] = Variable<int>(homeHits.value);
    }
    if (awayErrors.present) {
      map['away_errors'] = Variable<int>(awayErrors.value);
    }
    if (homeErrors.present) {
      map['home_errors'] = Variable<int>(homeErrors.value);
    }
    if (pitchCount.present) {
      map['pitch_count'] = Variable<int>(pitchCount.value);
    }
    if (currentInning.present) {
      map['current_inning'] = Variable<int>(currentInning.value);
    }
    if (currentHalf.present) {
      map['current_half'] = Variable<String>(currentHalf.value);
    }
    if (outs.present) {
      map['outs'] = Variable<int>(outs.value);
    }
    if (balls.present) {
      map['balls'] = Variable<int>(balls.value);
    }
    if (strikes.present) {
      map['strikes'] = Variable<int>(strikes.value);
    }
    if (awayBatterIndex.present) {
      map['away_batter_index'] = Variable<int>(awayBatterIndex.value);
    }
    if (homeBatterIndex.present) {
      map['home_batter_index'] = Variable<int>(homeBatterIndex.value);
    }
    if (lastPlaySequence.present) {
      map['last_play_sequence'] = Variable<int>(lastPlaySequence.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (lastModifiedAt.present) {
      map['last_modified_at'] = Variable<DateTime>(lastModifiedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardGamesCompanion(')
          ..write('uuid: $uuid, ')
          ..write('awayTeamName: $awayTeamName, ')
          ..write('awayTeamAbbreviation: $awayTeamAbbreviation, ')
          ..write('homeTeamName: $homeTeamName, ')
          ..write('homeTeamAbbreviation: $homeTeamAbbreviation, ')
          ..write('venue: $venue, ')
          ..write('gameDate: $gameDate, ')
          ..write('status: $status, ')
          ..write('awayScore: $awayScore, ')
          ..write('homeScore: $homeScore, ')
          ..write('awayHits: $awayHits, ')
          ..write('homeHits: $homeHits, ')
          ..write('awayErrors: $awayErrors, ')
          ..write('homeErrors: $homeErrors, ')
          ..write('pitchCount: $pitchCount, ')
          ..write('currentInning: $currentInning, ')
          ..write('currentHalf: $currentHalf, ')
          ..write('outs: $outs, ')
          ..write('balls: $balls, ')
          ..write('strikes: $strikes, ')
          ..write('awayBatterIndex: $awayBatterIndex, ')
          ..write('homeBatterIndex: $homeBatterIndex, ')
          ..write('lastPlaySequence: $lastPlaySequence, ')
          ..write('createdAt: $createdAt, ')
          ..write('lastModifiedAt: $lastModifiedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $ScorecardLineupsTable extends ScorecardLineups
    with TableInfo<$ScorecardLineupsTable, ScorecardLineup> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ScorecardLineupsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
    'id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _gameUuidMeta = const VerificationMeta(
    'gameUuid',
  );
  @override
  late final GeneratedColumn<String> gameUuid = GeneratedColumn<String>(
    'game_uuid',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES scorecard_games (uuid)',
    ),
  );
  static const VerificationMeta _teamSideMeta = const VerificationMeta(
    'teamSide',
  );
  @override
  late final GeneratedColumn<String> teamSide = GeneratedColumn<String>(
    'team_side',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _battingOrderMeta = const VerificationMeta(
    'battingOrder',
  );
  @override
  late final GeneratedColumn<int> battingOrder = GeneratedColumn<int>(
    'batting_order',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _playerNameMeta = const VerificationMeta(
    'playerName',
  );
  @override
  late final GeneratedColumn<String> playerName = GeneratedColumn<String>(
    'player_name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _positionCodeMeta = const VerificationMeta(
    'positionCode',
  );
  @override
  late final GeneratedColumn<String> positionCode = GeneratedColumn<String>(
    'position_code',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  @override
  List<GeneratedColumn> get $columns => [
    id,
    gameUuid,
    teamSide,
    battingOrder,
    playerName,
    positionCode,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'scorecard_lineups';
  @override
  VerificationContext validateIntegrity(
    Insertable<ScorecardLineup> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('game_uuid')) {
      context.handle(
        _gameUuidMeta,
        gameUuid.isAcceptableOrUnknown(data['game_uuid']!, _gameUuidMeta),
      );
    } else if (isInserting) {
      context.missing(_gameUuidMeta);
    }
    if (data.containsKey('team_side')) {
      context.handle(
        _teamSideMeta,
        teamSide.isAcceptableOrUnknown(data['team_side']!, _teamSideMeta),
      );
    } else if (isInserting) {
      context.missing(_teamSideMeta);
    }
    if (data.containsKey('batting_order')) {
      context.handle(
        _battingOrderMeta,
        battingOrder.isAcceptableOrUnknown(
          data['batting_order']!,
          _battingOrderMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_battingOrderMeta);
    }
    if (data.containsKey('player_name')) {
      context.handle(
        _playerNameMeta,
        playerName.isAcceptableOrUnknown(data['player_name']!, _playerNameMeta),
      );
    } else if (isInserting) {
      context.missing(_playerNameMeta);
    }
    if (data.containsKey('position_code')) {
      context.handle(
        _positionCodeMeta,
        positionCode.isAcceptableOrUnknown(
          data['position_code']!,
          _positionCodeMeta,
        ),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  ScorecardLineup map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ScorecardLineup(
      id: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}id'],
      )!,
      gameUuid: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}game_uuid'],
      )!,
      teamSide: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}team_side'],
      )!,
      battingOrder: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}batting_order'],
      )!,
      playerName: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}player_name'],
      )!,
      positionCode: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}position_code'],
      ),
    );
  }

  @override
  $ScorecardLineupsTable createAlias(String alias) {
    return $ScorecardLineupsTable(attachedDatabase, alias);
  }
}

class ScorecardLineup extends DataClass implements Insertable<ScorecardLineup> {
  final int id;
  final String gameUuid;
  final String teamSide;
  final int battingOrder;
  final String playerName;
  final String? positionCode;
  const ScorecardLineup({
    required this.id,
    required this.gameUuid,
    required this.teamSide,
    required this.battingOrder,
    required this.playerName,
    this.positionCode,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['game_uuid'] = Variable<String>(gameUuid);
    map['team_side'] = Variable<String>(teamSide);
    map['batting_order'] = Variable<int>(battingOrder);
    map['player_name'] = Variable<String>(playerName);
    if (!nullToAbsent || positionCode != null) {
      map['position_code'] = Variable<String>(positionCode);
    }
    return map;
  }

  ScorecardLineupsCompanion toCompanion(bool nullToAbsent) {
    return ScorecardLineupsCompanion(
      id: Value(id),
      gameUuid: Value(gameUuid),
      teamSide: Value(teamSide),
      battingOrder: Value(battingOrder),
      playerName: Value(playerName),
      positionCode: positionCode == null && nullToAbsent
          ? const Value.absent()
          : Value(positionCode),
    );
  }

  factory ScorecardLineup.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ScorecardLineup(
      id: serializer.fromJson<int>(json['id']),
      gameUuid: serializer.fromJson<String>(json['gameUuid']),
      teamSide: serializer.fromJson<String>(json['teamSide']),
      battingOrder: serializer.fromJson<int>(json['battingOrder']),
      playerName: serializer.fromJson<String>(json['playerName']),
      positionCode: serializer.fromJson<String?>(json['positionCode']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'gameUuid': serializer.toJson<String>(gameUuid),
      'teamSide': serializer.toJson<String>(teamSide),
      'battingOrder': serializer.toJson<int>(battingOrder),
      'playerName': serializer.toJson<String>(playerName),
      'positionCode': serializer.toJson<String?>(positionCode),
    };
  }

  ScorecardLineup copyWith({
    int? id,
    String? gameUuid,
    String? teamSide,
    int? battingOrder,
    String? playerName,
    Value<String?> positionCode = const Value.absent(),
  }) => ScorecardLineup(
    id: id ?? this.id,
    gameUuid: gameUuid ?? this.gameUuid,
    teamSide: teamSide ?? this.teamSide,
    battingOrder: battingOrder ?? this.battingOrder,
    playerName: playerName ?? this.playerName,
    positionCode: positionCode.present ? positionCode.value : this.positionCode,
  );
  ScorecardLineup copyWithCompanion(ScorecardLineupsCompanion data) {
    return ScorecardLineup(
      id: data.id.present ? data.id.value : this.id,
      gameUuid: data.gameUuid.present ? data.gameUuid.value : this.gameUuid,
      teamSide: data.teamSide.present ? data.teamSide.value : this.teamSide,
      battingOrder: data.battingOrder.present
          ? data.battingOrder.value
          : this.battingOrder,
      playerName: data.playerName.present
          ? data.playerName.value
          : this.playerName,
      positionCode: data.positionCode.present
          ? data.positionCode.value
          : this.positionCode,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardLineup(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('teamSide: $teamSide, ')
          ..write('battingOrder: $battingOrder, ')
          ..write('playerName: $playerName, ')
          ..write('positionCode: $positionCode')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    id,
    gameUuid,
    teamSide,
    battingOrder,
    playerName,
    positionCode,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ScorecardLineup &&
          other.id == this.id &&
          other.gameUuid == this.gameUuid &&
          other.teamSide == this.teamSide &&
          other.battingOrder == this.battingOrder &&
          other.playerName == this.playerName &&
          other.positionCode == this.positionCode);
}

class ScorecardLineupsCompanion extends UpdateCompanion<ScorecardLineup> {
  final Value<int> id;
  final Value<String> gameUuid;
  final Value<String> teamSide;
  final Value<int> battingOrder;
  final Value<String> playerName;
  final Value<String?> positionCode;
  const ScorecardLineupsCompanion({
    this.id = const Value.absent(),
    this.gameUuid = const Value.absent(),
    this.teamSide = const Value.absent(),
    this.battingOrder = const Value.absent(),
    this.playerName = const Value.absent(),
    this.positionCode = const Value.absent(),
  });
  ScorecardLineupsCompanion.insert({
    this.id = const Value.absent(),
    required String gameUuid,
    required String teamSide,
    required int battingOrder,
    required String playerName,
    this.positionCode = const Value.absent(),
  }) : gameUuid = Value(gameUuid),
       teamSide = Value(teamSide),
       battingOrder = Value(battingOrder),
       playerName = Value(playerName);
  static Insertable<ScorecardLineup> custom({
    Expression<int>? id,
    Expression<String>? gameUuid,
    Expression<String>? teamSide,
    Expression<int>? battingOrder,
    Expression<String>? playerName,
    Expression<String>? positionCode,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (gameUuid != null) 'game_uuid': gameUuid,
      if (teamSide != null) 'team_side': teamSide,
      if (battingOrder != null) 'batting_order': battingOrder,
      if (playerName != null) 'player_name': playerName,
      if (positionCode != null) 'position_code': positionCode,
    });
  }

  ScorecardLineupsCompanion copyWith({
    Value<int>? id,
    Value<String>? gameUuid,
    Value<String>? teamSide,
    Value<int>? battingOrder,
    Value<String>? playerName,
    Value<String?>? positionCode,
  }) {
    return ScorecardLineupsCompanion(
      id: id ?? this.id,
      gameUuid: gameUuid ?? this.gameUuid,
      teamSide: teamSide ?? this.teamSide,
      battingOrder: battingOrder ?? this.battingOrder,
      playerName: playerName ?? this.playerName,
      positionCode: positionCode ?? this.positionCode,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (gameUuid.present) {
      map['game_uuid'] = Variable<String>(gameUuid.value);
    }
    if (teamSide.present) {
      map['team_side'] = Variable<String>(teamSide.value);
    }
    if (battingOrder.present) {
      map['batting_order'] = Variable<int>(battingOrder.value);
    }
    if (playerName.present) {
      map['player_name'] = Variable<String>(playerName.value);
    }
    if (positionCode.present) {
      map['position_code'] = Variable<String>(positionCode.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardLineupsCompanion(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('teamSide: $teamSide, ')
          ..write('battingOrder: $battingOrder, ')
          ..write('playerName: $playerName, ')
          ..write('positionCode: $positionCode')
          ..write(')'))
        .toString();
  }
}

class $ScorecardInningsTable extends ScorecardInnings
    with TableInfo<$ScorecardInningsTable, ScorecardInningRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ScorecardInningsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
    'id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _gameUuidMeta = const VerificationMeta(
    'gameUuid',
  );
  @override
  late final GeneratedColumn<String> gameUuid = GeneratedColumn<String>(
    'game_uuid',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES scorecard_games (uuid)',
    ),
  );
  static const VerificationMeta _inningNumberMeta = const VerificationMeta(
    'inningNumber',
  );
  @override
  late final GeneratedColumn<int> inningNumber = GeneratedColumn<int>(
    'inning_number',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _halfMeta = const VerificationMeta('half');
  @override
  late final GeneratedColumn<String> half = GeneratedColumn<String>(
    'half',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [id, gameUuid, inningNumber, half];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'scorecard_innings';
  @override
  VerificationContext validateIntegrity(
    Insertable<ScorecardInningRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('game_uuid')) {
      context.handle(
        _gameUuidMeta,
        gameUuid.isAcceptableOrUnknown(data['game_uuid']!, _gameUuidMeta),
      );
    } else if (isInserting) {
      context.missing(_gameUuidMeta);
    }
    if (data.containsKey('inning_number')) {
      context.handle(
        _inningNumberMeta,
        inningNumber.isAcceptableOrUnknown(
          data['inning_number']!,
          _inningNumberMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_inningNumberMeta);
    }
    if (data.containsKey('half')) {
      context.handle(
        _halfMeta,
        half.isAcceptableOrUnknown(data['half']!, _halfMeta),
      );
    } else if (isInserting) {
      context.missing(_halfMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  ScorecardInningRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ScorecardInningRow(
      id: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}id'],
      )!,
      gameUuid: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}game_uuid'],
      )!,
      inningNumber: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}inning_number'],
      )!,
      half: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}half'],
      )!,
    );
  }

  @override
  $ScorecardInningsTable createAlias(String alias) {
    return $ScorecardInningsTable(attachedDatabase, alias);
  }
}

class ScorecardInningRow extends DataClass
    implements Insertable<ScorecardInningRow> {
  final int id;
  final String gameUuid;
  final int inningNumber;
  final String half;
  const ScorecardInningRow({
    required this.id,
    required this.gameUuid,
    required this.inningNumber,
    required this.half,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['game_uuid'] = Variable<String>(gameUuid);
    map['inning_number'] = Variable<int>(inningNumber);
    map['half'] = Variable<String>(half);
    return map;
  }

  ScorecardInningsCompanion toCompanion(bool nullToAbsent) {
    return ScorecardInningsCompanion(
      id: Value(id),
      gameUuid: Value(gameUuid),
      inningNumber: Value(inningNumber),
      half: Value(half),
    );
  }

  factory ScorecardInningRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ScorecardInningRow(
      id: serializer.fromJson<int>(json['id']),
      gameUuid: serializer.fromJson<String>(json['gameUuid']),
      inningNumber: serializer.fromJson<int>(json['inningNumber']),
      half: serializer.fromJson<String>(json['half']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'gameUuid': serializer.toJson<String>(gameUuid),
      'inningNumber': serializer.toJson<int>(inningNumber),
      'half': serializer.toJson<String>(half),
    };
  }

  ScorecardInningRow copyWith({
    int? id,
    String? gameUuid,
    int? inningNumber,
    String? half,
  }) => ScorecardInningRow(
    id: id ?? this.id,
    gameUuid: gameUuid ?? this.gameUuid,
    inningNumber: inningNumber ?? this.inningNumber,
    half: half ?? this.half,
  );
  ScorecardInningRow copyWithCompanion(ScorecardInningsCompanion data) {
    return ScorecardInningRow(
      id: data.id.present ? data.id.value : this.id,
      gameUuid: data.gameUuid.present ? data.gameUuid.value : this.gameUuid,
      inningNumber: data.inningNumber.present
          ? data.inningNumber.value
          : this.inningNumber,
      half: data.half.present ? data.half.value : this.half,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardInningRow(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('inningNumber: $inningNumber, ')
          ..write('half: $half')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, gameUuid, inningNumber, half);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ScorecardInningRow &&
          other.id == this.id &&
          other.gameUuid == this.gameUuid &&
          other.inningNumber == this.inningNumber &&
          other.half == this.half);
}

class ScorecardInningsCompanion extends UpdateCompanion<ScorecardInningRow> {
  final Value<int> id;
  final Value<String> gameUuid;
  final Value<int> inningNumber;
  final Value<String> half;
  const ScorecardInningsCompanion({
    this.id = const Value.absent(),
    this.gameUuid = const Value.absent(),
    this.inningNumber = const Value.absent(),
    this.half = const Value.absent(),
  });
  ScorecardInningsCompanion.insert({
    this.id = const Value.absent(),
    required String gameUuid,
    required int inningNumber,
    required String half,
  }) : gameUuid = Value(gameUuid),
       inningNumber = Value(inningNumber),
       half = Value(half);
  static Insertable<ScorecardInningRow> custom({
    Expression<int>? id,
    Expression<String>? gameUuid,
    Expression<int>? inningNumber,
    Expression<String>? half,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (gameUuid != null) 'game_uuid': gameUuid,
      if (inningNumber != null) 'inning_number': inningNumber,
      if (half != null) 'half': half,
    });
  }

  ScorecardInningsCompanion copyWith({
    Value<int>? id,
    Value<String>? gameUuid,
    Value<int>? inningNumber,
    Value<String>? half,
  }) {
    return ScorecardInningsCompanion(
      id: id ?? this.id,
      gameUuid: gameUuid ?? this.gameUuid,
      inningNumber: inningNumber ?? this.inningNumber,
      half: half ?? this.half,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (gameUuid.present) {
      map['game_uuid'] = Variable<String>(gameUuid.value);
    }
    if (inningNumber.present) {
      map['inning_number'] = Variable<int>(inningNumber.value);
    }
    if (half.present) {
      map['half'] = Variable<String>(half.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardInningsCompanion(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('inningNumber: $inningNumber, ')
          ..write('half: $half')
          ..write(')'))
        .toString();
  }
}

class $ScorecardPlaysTable extends ScorecardPlays
    with TableInfo<$ScorecardPlaysTable, ScorecardPlayRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ScorecardPlaysTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
    'id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _gameUuidMeta = const VerificationMeta(
    'gameUuid',
  );
  @override
  late final GeneratedColumn<String> gameUuid = GeneratedColumn<String>(
    'game_uuid',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES scorecard_games (uuid)',
    ),
  );
  static const VerificationMeta _inningIdMeta = const VerificationMeta(
    'inningId',
  );
  @override
  late final GeneratedColumn<int> inningId = GeneratedColumn<int>(
    'inning_id',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES scorecard_innings (id)',
    ),
  );
  static const VerificationMeta _battingTeamMeta = const VerificationMeta(
    'battingTeam',
  );
  @override
  late final GeneratedColumn<String> battingTeam = GeneratedColumn<String>(
    'batting_team',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _sequenceMeta = const VerificationMeta(
    'sequence',
  );
  @override
  late final GeneratedColumn<int> sequence = GeneratedColumn<int>(
    'sequence',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _batterIndexMeta = const VerificationMeta(
    'batterIndex',
  );
  @override
  late final GeneratedColumn<int> batterIndex = GeneratedColumn<int>(
    'batter_index',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _batterNameMeta = const VerificationMeta(
    'batterName',
  );
  @override
  late final GeneratedColumn<String> batterName = GeneratedColumn<String>(
    'batter_name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _batterPositionMeta = const VerificationMeta(
    'batterPosition',
  );
  @override
  late final GeneratedColumn<String> batterPosition = GeneratedColumn<String>(
    'batter_position',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _outcomeCodeMeta = const VerificationMeta(
    'outcomeCode',
  );
  @override
  late final GeneratedColumn<String> outcomeCode = GeneratedColumn<String>(
    'outcome_code',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _putoutSequenceMeta = const VerificationMeta(
    'putoutSequence',
  );
  @override
  late final GeneratedColumn<String> putoutSequence = GeneratedColumn<String>(
    'putout_sequence',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _pitchLogJsonMeta = const VerificationMeta(
    'pitchLogJson',
  );
  @override
  late final GeneratedColumn<String> pitchLogJson = GeneratedColumn<String>(
    'pitch_log_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _baseStateBeforeMeta = const VerificationMeta(
    'baseStateBefore',
  );
  @override
  late final GeneratedColumn<String> baseStateBefore = GeneratedColumn<String>(
    'base_state_before',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _baseStateAfterMeta = const VerificationMeta(
    'baseStateAfter',
  );
  @override
  late final GeneratedColumn<String> baseStateAfter = GeneratedColumn<String>(
    'base_state_after',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _rbiMeta = const VerificationMeta('rbi');
  @override
  late final GeneratedColumn<int> rbi = GeneratedColumn<int>(
    'rbi',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _scoredMeta = const VerificationMeta('scored');
  @override
  late final GeneratedColumn<bool> scored = GeneratedColumn<bool>(
    'scored',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'CHECK ("scored" IN (0, 1))',
    ),
  );
  static const VerificationMeta _recordedAtMeta = const VerificationMeta(
    'recordedAt',
  );
  @override
  late final GeneratedColumn<DateTime> recordedAt = GeneratedColumn<DateTime>(
    'recorded_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
    clientDefault: () => DateTime.now(),
  );
  @override
  List<GeneratedColumn> get $columns => [
    id,
    gameUuid,
    inningId,
    battingTeam,
    sequence,
    batterIndex,
    batterName,
    batterPosition,
    outcomeCode,
    putoutSequence,
    pitchLogJson,
    baseStateBefore,
    baseStateAfter,
    rbi,
    scored,
    recordedAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'scorecard_plays';
  @override
  VerificationContext validateIntegrity(
    Insertable<ScorecardPlayRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('game_uuid')) {
      context.handle(
        _gameUuidMeta,
        gameUuid.isAcceptableOrUnknown(data['game_uuid']!, _gameUuidMeta),
      );
    } else if (isInserting) {
      context.missing(_gameUuidMeta);
    }
    if (data.containsKey('inning_id')) {
      context.handle(
        _inningIdMeta,
        inningId.isAcceptableOrUnknown(data['inning_id']!, _inningIdMeta),
      );
    } else if (isInserting) {
      context.missing(_inningIdMeta);
    }
    if (data.containsKey('batting_team')) {
      context.handle(
        _battingTeamMeta,
        battingTeam.isAcceptableOrUnknown(
          data['batting_team']!,
          _battingTeamMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_battingTeamMeta);
    }
    if (data.containsKey('sequence')) {
      context.handle(
        _sequenceMeta,
        sequence.isAcceptableOrUnknown(data['sequence']!, _sequenceMeta),
      );
    } else if (isInserting) {
      context.missing(_sequenceMeta);
    }
    if (data.containsKey('batter_index')) {
      context.handle(
        _batterIndexMeta,
        batterIndex.isAcceptableOrUnknown(
          data['batter_index']!,
          _batterIndexMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_batterIndexMeta);
    }
    if (data.containsKey('batter_name')) {
      context.handle(
        _batterNameMeta,
        batterName.isAcceptableOrUnknown(data['batter_name']!, _batterNameMeta),
      );
    } else if (isInserting) {
      context.missing(_batterNameMeta);
    }
    if (data.containsKey('batter_position')) {
      context.handle(
        _batterPositionMeta,
        batterPosition.isAcceptableOrUnknown(
          data['batter_position']!,
          _batterPositionMeta,
        ),
      );
    }
    if (data.containsKey('outcome_code')) {
      context.handle(
        _outcomeCodeMeta,
        outcomeCode.isAcceptableOrUnknown(
          data['outcome_code']!,
          _outcomeCodeMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_outcomeCodeMeta);
    }
    if (data.containsKey('putout_sequence')) {
      context.handle(
        _putoutSequenceMeta,
        putoutSequence.isAcceptableOrUnknown(
          data['putout_sequence']!,
          _putoutSequenceMeta,
        ),
      );
    }
    if (data.containsKey('pitch_log_json')) {
      context.handle(
        _pitchLogJsonMeta,
        pitchLogJson.isAcceptableOrUnknown(
          data['pitch_log_json']!,
          _pitchLogJsonMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_pitchLogJsonMeta);
    }
    if (data.containsKey('base_state_before')) {
      context.handle(
        _baseStateBeforeMeta,
        baseStateBefore.isAcceptableOrUnknown(
          data['base_state_before']!,
          _baseStateBeforeMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_baseStateBeforeMeta);
    }
    if (data.containsKey('base_state_after')) {
      context.handle(
        _baseStateAfterMeta,
        baseStateAfter.isAcceptableOrUnknown(
          data['base_state_after']!,
          _baseStateAfterMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_baseStateAfterMeta);
    }
    if (data.containsKey('rbi')) {
      context.handle(
        _rbiMeta,
        rbi.isAcceptableOrUnknown(data['rbi']!, _rbiMeta),
      );
    } else if (isInserting) {
      context.missing(_rbiMeta);
    }
    if (data.containsKey('scored')) {
      context.handle(
        _scoredMeta,
        scored.isAcceptableOrUnknown(data['scored']!, _scoredMeta),
      );
    } else if (isInserting) {
      context.missing(_scoredMeta);
    }
    if (data.containsKey('recorded_at')) {
      context.handle(
        _recordedAtMeta,
        recordedAt.isAcceptableOrUnknown(data['recorded_at']!, _recordedAtMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  ScorecardPlayRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ScorecardPlayRow(
      id: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}id'],
      )!,
      gameUuid: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}game_uuid'],
      )!,
      inningId: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}inning_id'],
      )!,
      battingTeam: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}batting_team'],
      )!,
      sequence: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}sequence'],
      )!,
      batterIndex: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}batter_index'],
      )!,
      batterName: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}batter_name'],
      )!,
      batterPosition: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}batter_position'],
      ),
      outcomeCode: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}outcome_code'],
      )!,
      putoutSequence: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}putout_sequence'],
      ),
      pitchLogJson: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}pitch_log_json'],
      )!,
      baseStateBefore: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}base_state_before'],
      )!,
      baseStateAfter: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}base_state_after'],
      )!,
      rbi: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}rbi'],
      )!,
      scored: attachedDatabase.typeMapping.read(
        DriftSqlType.bool,
        data['${effectivePrefix}scored'],
      )!,
      recordedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}recorded_at'],
      )!,
    );
  }

  @override
  $ScorecardPlaysTable createAlias(String alias) {
    return $ScorecardPlaysTable(attachedDatabase, alias);
  }
}

class ScorecardPlayRow extends DataClass
    implements Insertable<ScorecardPlayRow> {
  final int id;
  final String gameUuid;
  final int inningId;
  final String battingTeam;
  final int sequence;
  final int batterIndex;
  final String batterName;
  final String? batterPosition;
  final String outcomeCode;
  final String? putoutSequence;
  final String pitchLogJson;
  final String baseStateBefore;
  final String baseStateAfter;
  final int rbi;
  final bool scored;
  final DateTime recordedAt;
  const ScorecardPlayRow({
    required this.id,
    required this.gameUuid,
    required this.inningId,
    required this.battingTeam,
    required this.sequence,
    required this.batterIndex,
    required this.batterName,
    this.batterPosition,
    required this.outcomeCode,
    this.putoutSequence,
    required this.pitchLogJson,
    required this.baseStateBefore,
    required this.baseStateAfter,
    required this.rbi,
    required this.scored,
    required this.recordedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['game_uuid'] = Variable<String>(gameUuid);
    map['inning_id'] = Variable<int>(inningId);
    map['batting_team'] = Variable<String>(battingTeam);
    map['sequence'] = Variable<int>(sequence);
    map['batter_index'] = Variable<int>(batterIndex);
    map['batter_name'] = Variable<String>(batterName);
    if (!nullToAbsent || batterPosition != null) {
      map['batter_position'] = Variable<String>(batterPosition);
    }
    map['outcome_code'] = Variable<String>(outcomeCode);
    if (!nullToAbsent || putoutSequence != null) {
      map['putout_sequence'] = Variable<String>(putoutSequence);
    }
    map['pitch_log_json'] = Variable<String>(pitchLogJson);
    map['base_state_before'] = Variable<String>(baseStateBefore);
    map['base_state_after'] = Variable<String>(baseStateAfter);
    map['rbi'] = Variable<int>(rbi);
    map['scored'] = Variable<bool>(scored);
    map['recorded_at'] = Variable<DateTime>(recordedAt);
    return map;
  }

  ScorecardPlaysCompanion toCompanion(bool nullToAbsent) {
    return ScorecardPlaysCompanion(
      id: Value(id),
      gameUuid: Value(gameUuid),
      inningId: Value(inningId),
      battingTeam: Value(battingTeam),
      sequence: Value(sequence),
      batterIndex: Value(batterIndex),
      batterName: Value(batterName),
      batterPosition: batterPosition == null && nullToAbsent
          ? const Value.absent()
          : Value(batterPosition),
      outcomeCode: Value(outcomeCode),
      putoutSequence: putoutSequence == null && nullToAbsent
          ? const Value.absent()
          : Value(putoutSequence),
      pitchLogJson: Value(pitchLogJson),
      baseStateBefore: Value(baseStateBefore),
      baseStateAfter: Value(baseStateAfter),
      rbi: Value(rbi),
      scored: Value(scored),
      recordedAt: Value(recordedAt),
    );
  }

  factory ScorecardPlayRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ScorecardPlayRow(
      id: serializer.fromJson<int>(json['id']),
      gameUuid: serializer.fromJson<String>(json['gameUuid']),
      inningId: serializer.fromJson<int>(json['inningId']),
      battingTeam: serializer.fromJson<String>(json['battingTeam']),
      sequence: serializer.fromJson<int>(json['sequence']),
      batterIndex: serializer.fromJson<int>(json['batterIndex']),
      batterName: serializer.fromJson<String>(json['batterName']),
      batterPosition: serializer.fromJson<String?>(json['batterPosition']),
      outcomeCode: serializer.fromJson<String>(json['outcomeCode']),
      putoutSequence: serializer.fromJson<String?>(json['putoutSequence']),
      pitchLogJson: serializer.fromJson<String>(json['pitchLogJson']),
      baseStateBefore: serializer.fromJson<String>(json['baseStateBefore']),
      baseStateAfter: serializer.fromJson<String>(json['baseStateAfter']),
      rbi: serializer.fromJson<int>(json['rbi']),
      scored: serializer.fromJson<bool>(json['scored']),
      recordedAt: serializer.fromJson<DateTime>(json['recordedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'gameUuid': serializer.toJson<String>(gameUuid),
      'inningId': serializer.toJson<int>(inningId),
      'battingTeam': serializer.toJson<String>(battingTeam),
      'sequence': serializer.toJson<int>(sequence),
      'batterIndex': serializer.toJson<int>(batterIndex),
      'batterName': serializer.toJson<String>(batterName),
      'batterPosition': serializer.toJson<String?>(batterPosition),
      'outcomeCode': serializer.toJson<String>(outcomeCode),
      'putoutSequence': serializer.toJson<String?>(putoutSequence),
      'pitchLogJson': serializer.toJson<String>(pitchLogJson),
      'baseStateBefore': serializer.toJson<String>(baseStateBefore),
      'baseStateAfter': serializer.toJson<String>(baseStateAfter),
      'rbi': serializer.toJson<int>(rbi),
      'scored': serializer.toJson<bool>(scored),
      'recordedAt': serializer.toJson<DateTime>(recordedAt),
    };
  }

  ScorecardPlayRow copyWith({
    int? id,
    String? gameUuid,
    int? inningId,
    String? battingTeam,
    int? sequence,
    int? batterIndex,
    String? batterName,
    Value<String?> batterPosition = const Value.absent(),
    String? outcomeCode,
    Value<String?> putoutSequence = const Value.absent(),
    String? pitchLogJson,
    String? baseStateBefore,
    String? baseStateAfter,
    int? rbi,
    bool? scored,
    DateTime? recordedAt,
  }) => ScorecardPlayRow(
    id: id ?? this.id,
    gameUuid: gameUuid ?? this.gameUuid,
    inningId: inningId ?? this.inningId,
    battingTeam: battingTeam ?? this.battingTeam,
    sequence: sequence ?? this.sequence,
    batterIndex: batterIndex ?? this.batterIndex,
    batterName: batterName ?? this.batterName,
    batterPosition: batterPosition.present
        ? batterPosition.value
        : this.batterPosition,
    outcomeCode: outcomeCode ?? this.outcomeCode,
    putoutSequence: putoutSequence.present
        ? putoutSequence.value
        : this.putoutSequence,
    pitchLogJson: pitchLogJson ?? this.pitchLogJson,
    baseStateBefore: baseStateBefore ?? this.baseStateBefore,
    baseStateAfter: baseStateAfter ?? this.baseStateAfter,
    rbi: rbi ?? this.rbi,
    scored: scored ?? this.scored,
    recordedAt: recordedAt ?? this.recordedAt,
  );
  ScorecardPlayRow copyWithCompanion(ScorecardPlaysCompanion data) {
    return ScorecardPlayRow(
      id: data.id.present ? data.id.value : this.id,
      gameUuid: data.gameUuid.present ? data.gameUuid.value : this.gameUuid,
      inningId: data.inningId.present ? data.inningId.value : this.inningId,
      battingTeam: data.battingTeam.present
          ? data.battingTeam.value
          : this.battingTeam,
      sequence: data.sequence.present ? data.sequence.value : this.sequence,
      batterIndex: data.batterIndex.present
          ? data.batterIndex.value
          : this.batterIndex,
      batterName: data.batterName.present
          ? data.batterName.value
          : this.batterName,
      batterPosition: data.batterPosition.present
          ? data.batterPosition.value
          : this.batterPosition,
      outcomeCode: data.outcomeCode.present
          ? data.outcomeCode.value
          : this.outcomeCode,
      putoutSequence: data.putoutSequence.present
          ? data.putoutSequence.value
          : this.putoutSequence,
      pitchLogJson: data.pitchLogJson.present
          ? data.pitchLogJson.value
          : this.pitchLogJson,
      baseStateBefore: data.baseStateBefore.present
          ? data.baseStateBefore.value
          : this.baseStateBefore,
      baseStateAfter: data.baseStateAfter.present
          ? data.baseStateAfter.value
          : this.baseStateAfter,
      rbi: data.rbi.present ? data.rbi.value : this.rbi,
      scored: data.scored.present ? data.scored.value : this.scored,
      recordedAt: data.recordedAt.present
          ? data.recordedAt.value
          : this.recordedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardPlayRow(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('inningId: $inningId, ')
          ..write('battingTeam: $battingTeam, ')
          ..write('sequence: $sequence, ')
          ..write('batterIndex: $batterIndex, ')
          ..write('batterName: $batterName, ')
          ..write('batterPosition: $batterPosition, ')
          ..write('outcomeCode: $outcomeCode, ')
          ..write('putoutSequence: $putoutSequence, ')
          ..write('pitchLogJson: $pitchLogJson, ')
          ..write('baseStateBefore: $baseStateBefore, ')
          ..write('baseStateAfter: $baseStateAfter, ')
          ..write('rbi: $rbi, ')
          ..write('scored: $scored, ')
          ..write('recordedAt: $recordedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    id,
    gameUuid,
    inningId,
    battingTeam,
    sequence,
    batterIndex,
    batterName,
    batterPosition,
    outcomeCode,
    putoutSequence,
    pitchLogJson,
    baseStateBefore,
    baseStateAfter,
    rbi,
    scored,
    recordedAt,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ScorecardPlayRow &&
          other.id == this.id &&
          other.gameUuid == this.gameUuid &&
          other.inningId == this.inningId &&
          other.battingTeam == this.battingTeam &&
          other.sequence == this.sequence &&
          other.batterIndex == this.batterIndex &&
          other.batterName == this.batterName &&
          other.batterPosition == this.batterPosition &&
          other.outcomeCode == this.outcomeCode &&
          other.putoutSequence == this.putoutSequence &&
          other.pitchLogJson == this.pitchLogJson &&
          other.baseStateBefore == this.baseStateBefore &&
          other.baseStateAfter == this.baseStateAfter &&
          other.rbi == this.rbi &&
          other.scored == this.scored &&
          other.recordedAt == this.recordedAt);
}

class ScorecardPlaysCompanion extends UpdateCompanion<ScorecardPlayRow> {
  final Value<int> id;
  final Value<String> gameUuid;
  final Value<int> inningId;
  final Value<String> battingTeam;
  final Value<int> sequence;
  final Value<int> batterIndex;
  final Value<String> batterName;
  final Value<String?> batterPosition;
  final Value<String> outcomeCode;
  final Value<String?> putoutSequence;
  final Value<String> pitchLogJson;
  final Value<String> baseStateBefore;
  final Value<String> baseStateAfter;
  final Value<int> rbi;
  final Value<bool> scored;
  final Value<DateTime> recordedAt;
  const ScorecardPlaysCompanion({
    this.id = const Value.absent(),
    this.gameUuid = const Value.absent(),
    this.inningId = const Value.absent(),
    this.battingTeam = const Value.absent(),
    this.sequence = const Value.absent(),
    this.batterIndex = const Value.absent(),
    this.batterName = const Value.absent(),
    this.batterPosition = const Value.absent(),
    this.outcomeCode = const Value.absent(),
    this.putoutSequence = const Value.absent(),
    this.pitchLogJson = const Value.absent(),
    this.baseStateBefore = const Value.absent(),
    this.baseStateAfter = const Value.absent(),
    this.rbi = const Value.absent(),
    this.scored = const Value.absent(),
    this.recordedAt = const Value.absent(),
  });
  ScorecardPlaysCompanion.insert({
    this.id = const Value.absent(),
    required String gameUuid,
    required int inningId,
    required String battingTeam,
    required int sequence,
    required int batterIndex,
    required String batterName,
    this.batterPosition = const Value.absent(),
    required String outcomeCode,
    this.putoutSequence = const Value.absent(),
    required String pitchLogJson,
    required String baseStateBefore,
    required String baseStateAfter,
    required int rbi,
    required bool scored,
    this.recordedAt = const Value.absent(),
  }) : gameUuid = Value(gameUuid),
       inningId = Value(inningId),
       battingTeam = Value(battingTeam),
       sequence = Value(sequence),
       batterIndex = Value(batterIndex),
       batterName = Value(batterName),
       outcomeCode = Value(outcomeCode),
       pitchLogJson = Value(pitchLogJson),
       baseStateBefore = Value(baseStateBefore),
       baseStateAfter = Value(baseStateAfter),
       rbi = Value(rbi),
       scored = Value(scored);
  static Insertable<ScorecardPlayRow> custom({
    Expression<int>? id,
    Expression<String>? gameUuid,
    Expression<int>? inningId,
    Expression<String>? battingTeam,
    Expression<int>? sequence,
    Expression<int>? batterIndex,
    Expression<String>? batterName,
    Expression<String>? batterPosition,
    Expression<String>? outcomeCode,
    Expression<String>? putoutSequence,
    Expression<String>? pitchLogJson,
    Expression<String>? baseStateBefore,
    Expression<String>? baseStateAfter,
    Expression<int>? rbi,
    Expression<bool>? scored,
    Expression<DateTime>? recordedAt,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (gameUuid != null) 'game_uuid': gameUuid,
      if (inningId != null) 'inning_id': inningId,
      if (battingTeam != null) 'batting_team': battingTeam,
      if (sequence != null) 'sequence': sequence,
      if (batterIndex != null) 'batter_index': batterIndex,
      if (batterName != null) 'batter_name': batterName,
      if (batterPosition != null) 'batter_position': batterPosition,
      if (outcomeCode != null) 'outcome_code': outcomeCode,
      if (putoutSequence != null) 'putout_sequence': putoutSequence,
      if (pitchLogJson != null) 'pitch_log_json': pitchLogJson,
      if (baseStateBefore != null) 'base_state_before': baseStateBefore,
      if (baseStateAfter != null) 'base_state_after': baseStateAfter,
      if (rbi != null) 'rbi': rbi,
      if (scored != null) 'scored': scored,
      if (recordedAt != null) 'recorded_at': recordedAt,
    });
  }

  ScorecardPlaysCompanion copyWith({
    Value<int>? id,
    Value<String>? gameUuid,
    Value<int>? inningId,
    Value<String>? battingTeam,
    Value<int>? sequence,
    Value<int>? batterIndex,
    Value<String>? batterName,
    Value<String?>? batterPosition,
    Value<String>? outcomeCode,
    Value<String?>? putoutSequence,
    Value<String>? pitchLogJson,
    Value<String>? baseStateBefore,
    Value<String>? baseStateAfter,
    Value<int>? rbi,
    Value<bool>? scored,
    Value<DateTime>? recordedAt,
  }) {
    return ScorecardPlaysCompanion(
      id: id ?? this.id,
      gameUuid: gameUuid ?? this.gameUuid,
      inningId: inningId ?? this.inningId,
      battingTeam: battingTeam ?? this.battingTeam,
      sequence: sequence ?? this.sequence,
      batterIndex: batterIndex ?? this.batterIndex,
      batterName: batterName ?? this.batterName,
      batterPosition: batterPosition ?? this.batterPosition,
      outcomeCode: outcomeCode ?? this.outcomeCode,
      putoutSequence: putoutSequence ?? this.putoutSequence,
      pitchLogJson: pitchLogJson ?? this.pitchLogJson,
      baseStateBefore: baseStateBefore ?? this.baseStateBefore,
      baseStateAfter: baseStateAfter ?? this.baseStateAfter,
      rbi: rbi ?? this.rbi,
      scored: scored ?? this.scored,
      recordedAt: recordedAt ?? this.recordedAt,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (gameUuid.present) {
      map['game_uuid'] = Variable<String>(gameUuid.value);
    }
    if (inningId.present) {
      map['inning_id'] = Variable<int>(inningId.value);
    }
    if (battingTeam.present) {
      map['batting_team'] = Variable<String>(battingTeam.value);
    }
    if (sequence.present) {
      map['sequence'] = Variable<int>(sequence.value);
    }
    if (batterIndex.present) {
      map['batter_index'] = Variable<int>(batterIndex.value);
    }
    if (batterName.present) {
      map['batter_name'] = Variable<String>(batterName.value);
    }
    if (batterPosition.present) {
      map['batter_position'] = Variable<String>(batterPosition.value);
    }
    if (outcomeCode.present) {
      map['outcome_code'] = Variable<String>(outcomeCode.value);
    }
    if (putoutSequence.present) {
      map['putout_sequence'] = Variable<String>(putoutSequence.value);
    }
    if (pitchLogJson.present) {
      map['pitch_log_json'] = Variable<String>(pitchLogJson.value);
    }
    if (baseStateBefore.present) {
      map['base_state_before'] = Variable<String>(baseStateBefore.value);
    }
    if (baseStateAfter.present) {
      map['base_state_after'] = Variable<String>(baseStateAfter.value);
    }
    if (rbi.present) {
      map['rbi'] = Variable<int>(rbi.value);
    }
    if (scored.present) {
      map['scored'] = Variable<bool>(scored.value);
    }
    if (recordedAt.present) {
      map['recorded_at'] = Variable<DateTime>(recordedAt.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ScorecardPlaysCompanion(')
          ..write('id: $id, ')
          ..write('gameUuid: $gameUuid, ')
          ..write('inningId: $inningId, ')
          ..write('battingTeam: $battingTeam, ')
          ..write('sequence: $sequence, ')
          ..write('batterIndex: $batterIndex, ')
          ..write('batterName: $batterName, ')
          ..write('batterPosition: $batterPosition, ')
          ..write('outcomeCode: $outcomeCode, ')
          ..write('putoutSequence: $putoutSequence, ')
          ..write('pitchLogJson: $pitchLogJson, ')
          ..write('baseStateBefore: $baseStateBefore, ')
          ..write('baseStateAfter: $baseStateAfter, ')
          ..write('rbi: $rbi, ')
          ..write('scored: $scored, ')
          ..write('recordedAt: $recordedAt')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $AppSettingsTable appSettings = $AppSettingsTable(this);
  late final $RecentPlayersTable recentPlayers = $RecentPlayersTable(this);
  late final $ScorecardGamesTable scorecardGames = $ScorecardGamesTable(this);
  late final $ScorecardLineupsTable scorecardLineups = $ScorecardLineupsTable(
    this,
  );
  late final $ScorecardInningsTable scorecardInnings = $ScorecardInningsTable(
    this,
  );
  late final $ScorecardPlaysTable scorecardPlays = $ScorecardPlaysTable(this);
  late final Index scorecardGamesStatusLastModifiedIdx = Index(
    'scorecard_games_status_last_modified_idx',
    'CREATE INDEX scorecard_games_status_last_modified_idx ON scorecard_games (status, last_modified_at)',
  );
  late final Index scorecardLineupsGameSideOrderIdx = Index(
    'scorecard_lineups_game_side_order_idx',
    'CREATE INDEX scorecard_lineups_game_side_order_idx ON scorecard_lineups (game_uuid, team_side, batting_order)',
  );
  late final Index scorecardInningsGameInningHalfIdx = Index(
    'scorecard_innings_game_inning_half_idx',
    'CREATE INDEX scorecard_innings_game_inning_half_idx ON scorecard_innings (game_uuid, inning_number, half)',
  );
  late final Index scorecardPlaysGameSequenceIdx = Index(
    'scorecard_plays_game_sequence_idx',
    'CREATE INDEX scorecard_plays_game_sequence_idx ON scorecard_plays (game_uuid, sequence)',
  );
  late final Index scorecardPlaysInningSequenceIdx = Index(
    'scorecard_plays_inning_sequence_idx',
    'CREATE INDEX scorecard_plays_inning_sequence_idx ON scorecard_plays (inning_id, sequence)',
  );
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [
    appSettings,
    recentPlayers,
    scorecardGames,
    scorecardLineups,
    scorecardInnings,
    scorecardPlays,
    scorecardGamesStatusLastModifiedIdx,
    scorecardLineupsGameSideOrderIdx,
    scorecardInningsGameInningHalfIdx,
    scorecardPlaysGameSequenceIdx,
    scorecardPlaysInningSequenceIdx,
  ];
}

typedef $$AppSettingsTableCreateCompanionBuilder =
    AppSettingsCompanion Function({
      required String key,
      required String value,
      Value<int> rowid,
    });
typedef $$AppSettingsTableUpdateCompanionBuilder =
    AppSettingsCompanion Function({
      Value<String> key,
      Value<String> value,
      Value<int> rowid,
    });

class $$AppSettingsTableFilterComposer
    extends Composer<_$AppDatabase, $AppSettingsTable> {
  $$AppSettingsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get key => $composableBuilder(
    column: $table.key,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get value => $composableBuilder(
    column: $table.value,
    builder: (column) => ColumnFilters(column),
  );
}

class $$AppSettingsTableOrderingComposer
    extends Composer<_$AppDatabase, $AppSettingsTable> {
  $$AppSettingsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get key => $composableBuilder(
    column: $table.key,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get value => $composableBuilder(
    column: $table.value,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$AppSettingsTableAnnotationComposer
    extends Composer<_$AppDatabase, $AppSettingsTable> {
  $$AppSettingsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get key =>
      $composableBuilder(column: $table.key, builder: (column) => column);

  GeneratedColumn<String> get value =>
      $composableBuilder(column: $table.value, builder: (column) => column);
}

class $$AppSettingsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $AppSettingsTable,
          AppSetting,
          $$AppSettingsTableFilterComposer,
          $$AppSettingsTableOrderingComposer,
          $$AppSettingsTableAnnotationComposer,
          $$AppSettingsTableCreateCompanionBuilder,
          $$AppSettingsTableUpdateCompanionBuilder,
          (
            AppSetting,
            BaseReferences<_$AppDatabase, $AppSettingsTable, AppSetting>,
          ),
          AppSetting,
          PrefetchHooks Function()
        > {
  $$AppSettingsTableTableManager(_$AppDatabase db, $AppSettingsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$AppSettingsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$AppSettingsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$AppSettingsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> key = const Value.absent(),
                Value<String> value = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => AppSettingsCompanion(key: key, value: value, rowid: rowid),
          createCompanionCallback:
              ({
                required String key,
                required String value,
                Value<int> rowid = const Value.absent(),
              }) => AppSettingsCompanion.insert(
                key: key,
                value: value,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$AppSettingsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $AppSettingsTable,
      AppSetting,
      $$AppSettingsTableFilterComposer,
      $$AppSettingsTableOrderingComposer,
      $$AppSettingsTableAnnotationComposer,
      $$AppSettingsTableCreateCompanionBuilder,
      $$AppSettingsTableUpdateCompanionBuilder,
      (
        AppSetting,
        BaseReferences<_$AppDatabase, $AppSettingsTable, AppSetting>,
      ),
      AppSetting,
      PrefetchHooks Function()
    >;
typedef $$RecentPlayersTableCreateCompanionBuilder =
    RecentPlayersCompanion Function({
      required String playerId,
      required int rankIndex,
      Value<DateTime> updatedAt,
      Value<int> rowid,
    });
typedef $$RecentPlayersTableUpdateCompanionBuilder =
    RecentPlayersCompanion Function({
      Value<String> playerId,
      Value<int> rankIndex,
      Value<DateTime> updatedAt,
      Value<int> rowid,
    });

class $$RecentPlayersTableFilterComposer
    extends Composer<_$AppDatabase, $RecentPlayersTable> {
  $$RecentPlayersTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get playerId => $composableBuilder(
    column: $table.playerId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get rankIndex => $composableBuilder(
    column: $table.rankIndex,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get updatedAt => $composableBuilder(
    column: $table.updatedAt,
    builder: (column) => ColumnFilters(column),
  );
}

class $$RecentPlayersTableOrderingComposer
    extends Composer<_$AppDatabase, $RecentPlayersTable> {
  $$RecentPlayersTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get playerId => $composableBuilder(
    column: $table.playerId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get rankIndex => $composableBuilder(
    column: $table.rankIndex,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get updatedAt => $composableBuilder(
    column: $table.updatedAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$RecentPlayersTableAnnotationComposer
    extends Composer<_$AppDatabase, $RecentPlayersTable> {
  $$RecentPlayersTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get playerId =>
      $composableBuilder(column: $table.playerId, builder: (column) => column);

  GeneratedColumn<int> get rankIndex =>
      $composableBuilder(column: $table.rankIndex, builder: (column) => column);

  GeneratedColumn<DateTime> get updatedAt =>
      $composableBuilder(column: $table.updatedAt, builder: (column) => column);
}

class $$RecentPlayersTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $RecentPlayersTable,
          RecentPlayer,
          $$RecentPlayersTableFilterComposer,
          $$RecentPlayersTableOrderingComposer,
          $$RecentPlayersTableAnnotationComposer,
          $$RecentPlayersTableCreateCompanionBuilder,
          $$RecentPlayersTableUpdateCompanionBuilder,
          (
            RecentPlayer,
            BaseReferences<_$AppDatabase, $RecentPlayersTable, RecentPlayer>,
          ),
          RecentPlayer,
          PrefetchHooks Function()
        > {
  $$RecentPlayersTableTableManager(_$AppDatabase db, $RecentPlayersTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$RecentPlayersTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$RecentPlayersTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$RecentPlayersTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> playerId = const Value.absent(),
                Value<int> rankIndex = const Value.absent(),
                Value<DateTime> updatedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => RecentPlayersCompanion(
                playerId: playerId,
                rankIndex: rankIndex,
                updatedAt: updatedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String playerId,
                required int rankIndex,
                Value<DateTime> updatedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => RecentPlayersCompanion.insert(
                playerId: playerId,
                rankIndex: rankIndex,
                updatedAt: updatedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$RecentPlayersTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $RecentPlayersTable,
      RecentPlayer,
      $$RecentPlayersTableFilterComposer,
      $$RecentPlayersTableOrderingComposer,
      $$RecentPlayersTableAnnotationComposer,
      $$RecentPlayersTableCreateCompanionBuilder,
      $$RecentPlayersTableUpdateCompanionBuilder,
      (
        RecentPlayer,
        BaseReferences<_$AppDatabase, $RecentPlayersTable, RecentPlayer>,
      ),
      RecentPlayer,
      PrefetchHooks Function()
    >;
typedef $$ScorecardGamesTableCreateCompanionBuilder =
    ScorecardGamesCompanion Function({
      required String uuid,
      required String awayTeamName,
      required String awayTeamAbbreviation,
      required String homeTeamName,
      required String homeTeamAbbreviation,
      Value<String?> venue,
      required DateTime gameDate,
      required String status,
      Value<int> awayScore,
      Value<int> homeScore,
      Value<int> awayHits,
      Value<int> homeHits,
      Value<int> awayErrors,
      Value<int> homeErrors,
      Value<int> pitchCount,
      Value<int> currentInning,
      required String currentHalf,
      Value<int> outs,
      Value<int> balls,
      Value<int> strikes,
      Value<int> awayBatterIndex,
      Value<int> homeBatterIndex,
      Value<int?> lastPlaySequence,
      Value<DateTime> createdAt,
      Value<DateTime> lastModifiedAt,
      Value<int> rowid,
    });
typedef $$ScorecardGamesTableUpdateCompanionBuilder =
    ScorecardGamesCompanion Function({
      Value<String> uuid,
      Value<String> awayTeamName,
      Value<String> awayTeamAbbreviation,
      Value<String> homeTeamName,
      Value<String> homeTeamAbbreviation,
      Value<String?> venue,
      Value<DateTime> gameDate,
      Value<String> status,
      Value<int> awayScore,
      Value<int> homeScore,
      Value<int> awayHits,
      Value<int> homeHits,
      Value<int> awayErrors,
      Value<int> homeErrors,
      Value<int> pitchCount,
      Value<int> currentInning,
      Value<String> currentHalf,
      Value<int> outs,
      Value<int> balls,
      Value<int> strikes,
      Value<int> awayBatterIndex,
      Value<int> homeBatterIndex,
      Value<int?> lastPlaySequence,
      Value<DateTime> createdAt,
      Value<DateTime> lastModifiedAt,
      Value<int> rowid,
    });

final class $$ScorecardGamesTableReferences
    extends BaseReferences<_$AppDatabase, $ScorecardGamesTable, ScorecardGame> {
  $$ScorecardGamesTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static MultiTypedResultKey<$ScorecardLineupsTable, List<ScorecardLineup>>
  _scorecardLineupsRefsTable(_$AppDatabase db) => MultiTypedResultKey.fromTable(
    db.scorecardLineups,
    aliasName: $_aliasNameGenerator(
      db.scorecardGames.uuid,
      db.scorecardLineups.gameUuid,
    ),
  );

  $$ScorecardLineupsTableProcessedTableManager get scorecardLineupsRefs {
    final manager = $$ScorecardLineupsTableTableManager(
      $_db,
      $_db.scorecardLineups,
    ).filter((f) => f.gameUuid.uuid.sqlEquals($_itemColumn<String>('uuid')!));

    final cache = $_typedResult.readTableOrNull(
      _scorecardLineupsRefsTable($_db),
    );
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }

  static MultiTypedResultKey<$ScorecardInningsTable, List<ScorecardInningRow>>
  _scorecardInningsRefsTable(_$AppDatabase db) => MultiTypedResultKey.fromTable(
    db.scorecardInnings,
    aliasName: $_aliasNameGenerator(
      db.scorecardGames.uuid,
      db.scorecardInnings.gameUuid,
    ),
  );

  $$ScorecardInningsTableProcessedTableManager get scorecardInningsRefs {
    final manager = $$ScorecardInningsTableTableManager(
      $_db,
      $_db.scorecardInnings,
    ).filter((f) => f.gameUuid.uuid.sqlEquals($_itemColumn<String>('uuid')!));

    final cache = $_typedResult.readTableOrNull(
      _scorecardInningsRefsTable($_db),
    );
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }

  static MultiTypedResultKey<$ScorecardPlaysTable, List<ScorecardPlayRow>>
  _scorecardPlaysRefsTable(_$AppDatabase db) => MultiTypedResultKey.fromTable(
    db.scorecardPlays,
    aliasName: $_aliasNameGenerator(
      db.scorecardGames.uuid,
      db.scorecardPlays.gameUuid,
    ),
  );

  $$ScorecardPlaysTableProcessedTableManager get scorecardPlaysRefs {
    final manager = $$ScorecardPlaysTableTableManager(
      $_db,
      $_db.scorecardPlays,
    ).filter((f) => f.gameUuid.uuid.sqlEquals($_itemColumn<String>('uuid')!));

    final cache = $_typedResult.readTableOrNull(_scorecardPlaysRefsTable($_db));
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }
}

class $$ScorecardGamesTableFilterComposer
    extends Composer<_$AppDatabase, $ScorecardGamesTable> {
  $$ScorecardGamesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get uuid => $composableBuilder(
    column: $table.uuid,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get awayTeamName => $composableBuilder(
    column: $table.awayTeamName,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get awayTeamAbbreviation => $composableBuilder(
    column: $table.awayTeamAbbreviation,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get homeTeamName => $composableBuilder(
    column: $table.homeTeamName,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get homeTeamAbbreviation => $composableBuilder(
    column: $table.homeTeamAbbreviation,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get venue => $composableBuilder(
    column: $table.venue,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get gameDate => $composableBuilder(
    column: $table.gameDate,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get awayScore => $composableBuilder(
    column: $table.awayScore,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get homeScore => $composableBuilder(
    column: $table.homeScore,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get awayHits => $composableBuilder(
    column: $table.awayHits,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get homeHits => $composableBuilder(
    column: $table.homeHits,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get awayErrors => $composableBuilder(
    column: $table.awayErrors,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get homeErrors => $composableBuilder(
    column: $table.homeErrors,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get pitchCount => $composableBuilder(
    column: $table.pitchCount,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get currentInning => $composableBuilder(
    column: $table.currentInning,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get currentHalf => $composableBuilder(
    column: $table.currentHalf,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get outs => $composableBuilder(
    column: $table.outs,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get balls => $composableBuilder(
    column: $table.balls,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get strikes => $composableBuilder(
    column: $table.strikes,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get awayBatterIndex => $composableBuilder(
    column: $table.awayBatterIndex,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get homeBatterIndex => $composableBuilder(
    column: $table.homeBatterIndex,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get lastPlaySequence => $composableBuilder(
    column: $table.lastPlaySequence,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get lastModifiedAt => $composableBuilder(
    column: $table.lastModifiedAt,
    builder: (column) => ColumnFilters(column),
  );

  Expression<bool> scorecardLineupsRefs(
    Expression<bool> Function($$ScorecardLineupsTableFilterComposer f) f,
  ) {
    final $$ScorecardLineupsTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardLineups,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardLineupsTableFilterComposer(
            $db: $db,
            $table: $db.scorecardLineups,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }

  Expression<bool> scorecardInningsRefs(
    Expression<bool> Function($$ScorecardInningsTableFilterComposer f) f,
  ) {
    final $$ScorecardInningsTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardInnings,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardInningsTableFilterComposer(
            $db: $db,
            $table: $db.scorecardInnings,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }

  Expression<bool> scorecardPlaysRefs(
    Expression<bool> Function($$ScorecardPlaysTableFilterComposer f) f,
  ) {
    final $$ScorecardPlaysTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardPlays,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardPlaysTableFilterComposer(
            $db: $db,
            $table: $db.scorecardPlays,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$ScorecardGamesTableOrderingComposer
    extends Composer<_$AppDatabase, $ScorecardGamesTable> {
  $$ScorecardGamesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get uuid => $composableBuilder(
    column: $table.uuid,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get awayTeamName => $composableBuilder(
    column: $table.awayTeamName,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get awayTeamAbbreviation => $composableBuilder(
    column: $table.awayTeamAbbreviation,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get homeTeamName => $composableBuilder(
    column: $table.homeTeamName,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get homeTeamAbbreviation => $composableBuilder(
    column: $table.homeTeamAbbreviation,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get venue => $composableBuilder(
    column: $table.venue,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get gameDate => $composableBuilder(
    column: $table.gameDate,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get awayScore => $composableBuilder(
    column: $table.awayScore,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get homeScore => $composableBuilder(
    column: $table.homeScore,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get awayHits => $composableBuilder(
    column: $table.awayHits,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get homeHits => $composableBuilder(
    column: $table.homeHits,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get awayErrors => $composableBuilder(
    column: $table.awayErrors,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get homeErrors => $composableBuilder(
    column: $table.homeErrors,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get pitchCount => $composableBuilder(
    column: $table.pitchCount,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get currentInning => $composableBuilder(
    column: $table.currentInning,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get currentHalf => $composableBuilder(
    column: $table.currentHalf,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get outs => $composableBuilder(
    column: $table.outs,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get balls => $composableBuilder(
    column: $table.balls,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get strikes => $composableBuilder(
    column: $table.strikes,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get awayBatterIndex => $composableBuilder(
    column: $table.awayBatterIndex,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get homeBatterIndex => $composableBuilder(
    column: $table.homeBatterIndex,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get lastPlaySequence => $composableBuilder(
    column: $table.lastPlaySequence,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
    column: $table.createdAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get lastModifiedAt => $composableBuilder(
    column: $table.lastModifiedAt,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$ScorecardGamesTableAnnotationComposer
    extends Composer<_$AppDatabase, $ScorecardGamesTable> {
  $$ScorecardGamesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get uuid =>
      $composableBuilder(column: $table.uuid, builder: (column) => column);

  GeneratedColumn<String> get awayTeamName => $composableBuilder(
    column: $table.awayTeamName,
    builder: (column) => column,
  );

  GeneratedColumn<String> get awayTeamAbbreviation => $composableBuilder(
    column: $table.awayTeamAbbreviation,
    builder: (column) => column,
  );

  GeneratedColumn<String> get homeTeamName => $composableBuilder(
    column: $table.homeTeamName,
    builder: (column) => column,
  );

  GeneratedColumn<String> get homeTeamAbbreviation => $composableBuilder(
    column: $table.homeTeamAbbreviation,
    builder: (column) => column,
  );

  GeneratedColumn<String> get venue =>
      $composableBuilder(column: $table.venue, builder: (column) => column);

  GeneratedColumn<DateTime> get gameDate =>
      $composableBuilder(column: $table.gameDate, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<int> get awayScore =>
      $composableBuilder(column: $table.awayScore, builder: (column) => column);

  GeneratedColumn<int> get homeScore =>
      $composableBuilder(column: $table.homeScore, builder: (column) => column);

  GeneratedColumn<int> get awayHits =>
      $composableBuilder(column: $table.awayHits, builder: (column) => column);

  GeneratedColumn<int> get homeHits =>
      $composableBuilder(column: $table.homeHits, builder: (column) => column);

  GeneratedColumn<int> get awayErrors => $composableBuilder(
    column: $table.awayErrors,
    builder: (column) => column,
  );

  GeneratedColumn<int> get homeErrors => $composableBuilder(
    column: $table.homeErrors,
    builder: (column) => column,
  );

  GeneratedColumn<int> get pitchCount => $composableBuilder(
    column: $table.pitchCount,
    builder: (column) => column,
  );

  GeneratedColumn<int> get currentInning => $composableBuilder(
    column: $table.currentInning,
    builder: (column) => column,
  );

  GeneratedColumn<String> get currentHalf => $composableBuilder(
    column: $table.currentHalf,
    builder: (column) => column,
  );

  GeneratedColumn<int> get outs =>
      $composableBuilder(column: $table.outs, builder: (column) => column);

  GeneratedColumn<int> get balls =>
      $composableBuilder(column: $table.balls, builder: (column) => column);

  GeneratedColumn<int> get strikes =>
      $composableBuilder(column: $table.strikes, builder: (column) => column);

  GeneratedColumn<int> get awayBatterIndex => $composableBuilder(
    column: $table.awayBatterIndex,
    builder: (column) => column,
  );

  GeneratedColumn<int> get homeBatterIndex => $composableBuilder(
    column: $table.homeBatterIndex,
    builder: (column) => column,
  );

  GeneratedColumn<int> get lastPlaySequence => $composableBuilder(
    column: $table.lastPlaySequence,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<DateTime> get lastModifiedAt => $composableBuilder(
    column: $table.lastModifiedAt,
    builder: (column) => column,
  );

  Expression<T> scorecardLineupsRefs<T extends Object>(
    Expression<T> Function($$ScorecardLineupsTableAnnotationComposer a) f,
  ) {
    final $$ScorecardLineupsTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardLineups,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardLineupsTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardLineups,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }

  Expression<T> scorecardInningsRefs<T extends Object>(
    Expression<T> Function($$ScorecardInningsTableAnnotationComposer a) f,
  ) {
    final $$ScorecardInningsTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardInnings,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardInningsTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardInnings,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }

  Expression<T> scorecardPlaysRefs<T extends Object>(
    Expression<T> Function($$ScorecardPlaysTableAnnotationComposer a) f,
  ) {
    final $$ScorecardPlaysTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.uuid,
      referencedTable: $db.scorecardPlays,
      getReferencedColumn: (t) => t.gameUuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardPlaysTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardPlays,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$ScorecardGamesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ScorecardGamesTable,
          ScorecardGame,
          $$ScorecardGamesTableFilterComposer,
          $$ScorecardGamesTableOrderingComposer,
          $$ScorecardGamesTableAnnotationComposer,
          $$ScorecardGamesTableCreateCompanionBuilder,
          $$ScorecardGamesTableUpdateCompanionBuilder,
          (ScorecardGame, $$ScorecardGamesTableReferences),
          ScorecardGame,
          PrefetchHooks Function({
            bool scorecardLineupsRefs,
            bool scorecardInningsRefs,
            bool scorecardPlaysRefs,
          })
        > {
  $$ScorecardGamesTableTableManager(
    _$AppDatabase db,
    $ScorecardGamesTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$ScorecardGamesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$ScorecardGamesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$ScorecardGamesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> uuid = const Value.absent(),
                Value<String> awayTeamName = const Value.absent(),
                Value<String> awayTeamAbbreviation = const Value.absent(),
                Value<String> homeTeamName = const Value.absent(),
                Value<String> homeTeamAbbreviation = const Value.absent(),
                Value<String?> venue = const Value.absent(),
                Value<DateTime> gameDate = const Value.absent(),
                Value<String> status = const Value.absent(),
                Value<int> awayScore = const Value.absent(),
                Value<int> homeScore = const Value.absent(),
                Value<int> awayHits = const Value.absent(),
                Value<int> homeHits = const Value.absent(),
                Value<int> awayErrors = const Value.absent(),
                Value<int> homeErrors = const Value.absent(),
                Value<int> pitchCount = const Value.absent(),
                Value<int> currentInning = const Value.absent(),
                Value<String> currentHalf = const Value.absent(),
                Value<int> outs = const Value.absent(),
                Value<int> balls = const Value.absent(),
                Value<int> strikes = const Value.absent(),
                Value<int> awayBatterIndex = const Value.absent(),
                Value<int> homeBatterIndex = const Value.absent(),
                Value<int?> lastPlaySequence = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<DateTime> lastModifiedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ScorecardGamesCompanion(
                uuid: uuid,
                awayTeamName: awayTeamName,
                awayTeamAbbreviation: awayTeamAbbreviation,
                homeTeamName: homeTeamName,
                homeTeamAbbreviation: homeTeamAbbreviation,
                venue: venue,
                gameDate: gameDate,
                status: status,
                awayScore: awayScore,
                homeScore: homeScore,
                awayHits: awayHits,
                homeHits: homeHits,
                awayErrors: awayErrors,
                homeErrors: homeErrors,
                pitchCount: pitchCount,
                currentInning: currentInning,
                currentHalf: currentHalf,
                outs: outs,
                balls: balls,
                strikes: strikes,
                awayBatterIndex: awayBatterIndex,
                homeBatterIndex: homeBatterIndex,
                lastPlaySequence: lastPlaySequence,
                createdAt: createdAt,
                lastModifiedAt: lastModifiedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String uuid,
                required String awayTeamName,
                required String awayTeamAbbreviation,
                required String homeTeamName,
                required String homeTeamAbbreviation,
                Value<String?> venue = const Value.absent(),
                required DateTime gameDate,
                required String status,
                Value<int> awayScore = const Value.absent(),
                Value<int> homeScore = const Value.absent(),
                Value<int> awayHits = const Value.absent(),
                Value<int> homeHits = const Value.absent(),
                Value<int> awayErrors = const Value.absent(),
                Value<int> homeErrors = const Value.absent(),
                Value<int> pitchCount = const Value.absent(),
                Value<int> currentInning = const Value.absent(),
                required String currentHalf,
                Value<int> outs = const Value.absent(),
                Value<int> balls = const Value.absent(),
                Value<int> strikes = const Value.absent(),
                Value<int> awayBatterIndex = const Value.absent(),
                Value<int> homeBatterIndex = const Value.absent(),
                Value<int?> lastPlaySequence = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<DateTime> lastModifiedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ScorecardGamesCompanion.insert(
                uuid: uuid,
                awayTeamName: awayTeamName,
                awayTeamAbbreviation: awayTeamAbbreviation,
                homeTeamName: homeTeamName,
                homeTeamAbbreviation: homeTeamAbbreviation,
                venue: venue,
                gameDate: gameDate,
                status: status,
                awayScore: awayScore,
                homeScore: homeScore,
                awayHits: awayHits,
                homeHits: homeHits,
                awayErrors: awayErrors,
                homeErrors: homeErrors,
                pitchCount: pitchCount,
                currentInning: currentInning,
                currentHalf: currentHalf,
                outs: outs,
                balls: balls,
                strikes: strikes,
                awayBatterIndex: awayBatterIndex,
                homeBatterIndex: homeBatterIndex,
                lastPlaySequence: lastPlaySequence,
                createdAt: createdAt,
                lastModifiedAt: lastModifiedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$ScorecardGamesTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback:
              ({
                scorecardLineupsRefs = false,
                scorecardInningsRefs = false,
                scorecardPlaysRefs = false,
              }) {
                return PrefetchHooks(
                  db: db,
                  explicitlyWatchedTables: [
                    if (scorecardLineupsRefs) db.scorecardLineups,
                    if (scorecardInningsRefs) db.scorecardInnings,
                    if (scorecardPlaysRefs) db.scorecardPlays,
                  ],
                  addJoins: null,
                  getPrefetchedDataCallback: (items) async {
                    return [
                      if (scorecardLineupsRefs)
                        await $_getPrefetchedData<
                          ScorecardGame,
                          $ScorecardGamesTable,
                          ScorecardLineup
                        >(
                          currentTable: table,
                          referencedTable: $$ScorecardGamesTableReferences
                              ._scorecardLineupsRefsTable(db),
                          managerFromTypedResult: (p0) =>
                              $$ScorecardGamesTableReferences(
                                db,
                                table,
                                p0,
                              ).scorecardLineupsRefs,
                          referencedItemsForCurrentItem:
                              (item, referencedItems) => referencedItems.where(
                                (e) => e.gameUuid == item.uuid,
                              ),
                          typedResults: items,
                        ),
                      if (scorecardInningsRefs)
                        await $_getPrefetchedData<
                          ScorecardGame,
                          $ScorecardGamesTable,
                          ScorecardInningRow
                        >(
                          currentTable: table,
                          referencedTable: $$ScorecardGamesTableReferences
                              ._scorecardInningsRefsTable(db),
                          managerFromTypedResult: (p0) =>
                              $$ScorecardGamesTableReferences(
                                db,
                                table,
                                p0,
                              ).scorecardInningsRefs,
                          referencedItemsForCurrentItem:
                              (item, referencedItems) => referencedItems.where(
                                (e) => e.gameUuid == item.uuid,
                              ),
                          typedResults: items,
                        ),
                      if (scorecardPlaysRefs)
                        await $_getPrefetchedData<
                          ScorecardGame,
                          $ScorecardGamesTable,
                          ScorecardPlayRow
                        >(
                          currentTable: table,
                          referencedTable: $$ScorecardGamesTableReferences
                              ._scorecardPlaysRefsTable(db),
                          managerFromTypedResult: (p0) =>
                              $$ScorecardGamesTableReferences(
                                db,
                                table,
                                p0,
                              ).scorecardPlaysRefs,
                          referencedItemsForCurrentItem:
                              (item, referencedItems) => referencedItems.where(
                                (e) => e.gameUuid == item.uuid,
                              ),
                          typedResults: items,
                        ),
                    ];
                  },
                );
              },
        ),
      );
}

typedef $$ScorecardGamesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ScorecardGamesTable,
      ScorecardGame,
      $$ScorecardGamesTableFilterComposer,
      $$ScorecardGamesTableOrderingComposer,
      $$ScorecardGamesTableAnnotationComposer,
      $$ScorecardGamesTableCreateCompanionBuilder,
      $$ScorecardGamesTableUpdateCompanionBuilder,
      (ScorecardGame, $$ScorecardGamesTableReferences),
      ScorecardGame,
      PrefetchHooks Function({
        bool scorecardLineupsRefs,
        bool scorecardInningsRefs,
        bool scorecardPlaysRefs,
      })
    >;
typedef $$ScorecardLineupsTableCreateCompanionBuilder =
    ScorecardLineupsCompanion Function({
      Value<int> id,
      required String gameUuid,
      required String teamSide,
      required int battingOrder,
      required String playerName,
      Value<String?> positionCode,
    });
typedef $$ScorecardLineupsTableUpdateCompanionBuilder =
    ScorecardLineupsCompanion Function({
      Value<int> id,
      Value<String> gameUuid,
      Value<String> teamSide,
      Value<int> battingOrder,
      Value<String> playerName,
      Value<String?> positionCode,
    });

final class $$ScorecardLineupsTableReferences
    extends
        BaseReferences<_$AppDatabase, $ScorecardLineupsTable, ScorecardLineup> {
  $$ScorecardLineupsTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static $ScorecardGamesTable _gameUuidTable(_$AppDatabase db) =>
      db.scorecardGames.createAlias(
        $_aliasNameGenerator(
          db.scorecardLineups.gameUuid,
          db.scorecardGames.uuid,
        ),
      );

  $$ScorecardGamesTableProcessedTableManager get gameUuid {
    final $_column = $_itemColumn<String>('game_uuid')!;

    final manager = $$ScorecardGamesTableTableManager(
      $_db,
      $_db.scorecardGames,
    ).filter((f) => f.uuid.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_gameUuidTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }
}

class $$ScorecardLineupsTableFilterComposer
    extends Composer<_$AppDatabase, $ScorecardLineupsTable> {
  $$ScorecardLineupsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get teamSide => $composableBuilder(
    column: $table.teamSide,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get battingOrder => $composableBuilder(
    column: $table.battingOrder,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get playerName => $composableBuilder(
    column: $table.playerName,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get positionCode => $composableBuilder(
    column: $table.positionCode,
    builder: (column) => ColumnFilters(column),
  );

  $$ScorecardGamesTableFilterComposer get gameUuid {
    final $$ScorecardGamesTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableFilterComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardLineupsTableOrderingComposer
    extends Composer<_$AppDatabase, $ScorecardLineupsTable> {
  $$ScorecardLineupsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get teamSide => $composableBuilder(
    column: $table.teamSide,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get battingOrder => $composableBuilder(
    column: $table.battingOrder,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get playerName => $composableBuilder(
    column: $table.playerName,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get positionCode => $composableBuilder(
    column: $table.positionCode,
    builder: (column) => ColumnOrderings(column),
  );

  $$ScorecardGamesTableOrderingComposer get gameUuid {
    final $$ScorecardGamesTableOrderingComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableOrderingComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardLineupsTableAnnotationComposer
    extends Composer<_$AppDatabase, $ScorecardLineupsTable> {
  $$ScorecardLineupsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get teamSide =>
      $composableBuilder(column: $table.teamSide, builder: (column) => column);

  GeneratedColumn<int> get battingOrder => $composableBuilder(
    column: $table.battingOrder,
    builder: (column) => column,
  );

  GeneratedColumn<String> get playerName => $composableBuilder(
    column: $table.playerName,
    builder: (column) => column,
  );

  GeneratedColumn<String> get positionCode => $composableBuilder(
    column: $table.positionCode,
    builder: (column) => column,
  );

  $$ScorecardGamesTableAnnotationComposer get gameUuid {
    final $$ScorecardGamesTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardLineupsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ScorecardLineupsTable,
          ScorecardLineup,
          $$ScorecardLineupsTableFilterComposer,
          $$ScorecardLineupsTableOrderingComposer,
          $$ScorecardLineupsTableAnnotationComposer,
          $$ScorecardLineupsTableCreateCompanionBuilder,
          $$ScorecardLineupsTableUpdateCompanionBuilder,
          (ScorecardLineup, $$ScorecardLineupsTableReferences),
          ScorecardLineup,
          PrefetchHooks Function({bool gameUuid})
        > {
  $$ScorecardLineupsTableTableManager(
    _$AppDatabase db,
    $ScorecardLineupsTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$ScorecardLineupsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$ScorecardLineupsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$ScorecardLineupsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                Value<String> gameUuid = const Value.absent(),
                Value<String> teamSide = const Value.absent(),
                Value<int> battingOrder = const Value.absent(),
                Value<String> playerName = const Value.absent(),
                Value<String?> positionCode = const Value.absent(),
              }) => ScorecardLineupsCompanion(
                id: id,
                gameUuid: gameUuid,
                teamSide: teamSide,
                battingOrder: battingOrder,
                playerName: playerName,
                positionCode: positionCode,
              ),
          createCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                required String gameUuid,
                required String teamSide,
                required int battingOrder,
                required String playerName,
                Value<String?> positionCode = const Value.absent(),
              }) => ScorecardLineupsCompanion.insert(
                id: id,
                gameUuid: gameUuid,
                teamSide: teamSide,
                battingOrder: battingOrder,
                playerName: playerName,
                positionCode: positionCode,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$ScorecardLineupsTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: ({gameUuid = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [],
              addJoins:
                  <
                    T extends TableManagerState<
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic
                    >
                  >(state) {
                    if (gameUuid) {
                      state =
                          state.withJoin(
                                currentTable: table,
                                currentColumn: table.gameUuid,
                                referencedTable:
                                    $$ScorecardLineupsTableReferences
                                        ._gameUuidTable(db),
                                referencedColumn:
                                    $$ScorecardLineupsTableReferences
                                        ._gameUuidTable(db)
                                        .uuid,
                              )
                              as T;
                    }

                    return state;
                  },
              getPrefetchedDataCallback: (items) async {
                return [];
              },
            );
          },
        ),
      );
}

typedef $$ScorecardLineupsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ScorecardLineupsTable,
      ScorecardLineup,
      $$ScorecardLineupsTableFilterComposer,
      $$ScorecardLineupsTableOrderingComposer,
      $$ScorecardLineupsTableAnnotationComposer,
      $$ScorecardLineupsTableCreateCompanionBuilder,
      $$ScorecardLineupsTableUpdateCompanionBuilder,
      (ScorecardLineup, $$ScorecardLineupsTableReferences),
      ScorecardLineup,
      PrefetchHooks Function({bool gameUuid})
    >;
typedef $$ScorecardInningsTableCreateCompanionBuilder =
    ScorecardInningsCompanion Function({
      Value<int> id,
      required String gameUuid,
      required int inningNumber,
      required String half,
    });
typedef $$ScorecardInningsTableUpdateCompanionBuilder =
    ScorecardInningsCompanion Function({
      Value<int> id,
      Value<String> gameUuid,
      Value<int> inningNumber,
      Value<String> half,
    });

final class $$ScorecardInningsTableReferences
    extends
        BaseReferences<
          _$AppDatabase,
          $ScorecardInningsTable,
          ScorecardInningRow
        > {
  $$ScorecardInningsTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static $ScorecardGamesTable _gameUuidTable(_$AppDatabase db) =>
      db.scorecardGames.createAlias(
        $_aliasNameGenerator(
          db.scorecardInnings.gameUuid,
          db.scorecardGames.uuid,
        ),
      );

  $$ScorecardGamesTableProcessedTableManager get gameUuid {
    final $_column = $_itemColumn<String>('game_uuid')!;

    final manager = $$ScorecardGamesTableTableManager(
      $_db,
      $_db.scorecardGames,
    ).filter((f) => f.uuid.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_gameUuidTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }

  static MultiTypedResultKey<$ScorecardPlaysTable, List<ScorecardPlayRow>>
  _scorecardPlaysRefsTable(_$AppDatabase db) => MultiTypedResultKey.fromTable(
    db.scorecardPlays,
    aliasName: $_aliasNameGenerator(
      db.scorecardInnings.id,
      db.scorecardPlays.inningId,
    ),
  );

  $$ScorecardPlaysTableProcessedTableManager get scorecardPlaysRefs {
    final manager = $$ScorecardPlaysTableTableManager(
      $_db,
      $_db.scorecardPlays,
    ).filter((f) => f.inningId.id.sqlEquals($_itemColumn<int>('id')!));

    final cache = $_typedResult.readTableOrNull(_scorecardPlaysRefsTable($_db));
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }
}

class $$ScorecardInningsTableFilterComposer
    extends Composer<_$AppDatabase, $ScorecardInningsTable> {
  $$ScorecardInningsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get inningNumber => $composableBuilder(
    column: $table.inningNumber,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get half => $composableBuilder(
    column: $table.half,
    builder: (column) => ColumnFilters(column),
  );

  $$ScorecardGamesTableFilterComposer get gameUuid {
    final $$ScorecardGamesTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableFilterComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  Expression<bool> scorecardPlaysRefs(
    Expression<bool> Function($$ScorecardPlaysTableFilterComposer f) f,
  ) {
    final $$ScorecardPlaysTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.id,
      referencedTable: $db.scorecardPlays,
      getReferencedColumn: (t) => t.inningId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardPlaysTableFilterComposer(
            $db: $db,
            $table: $db.scorecardPlays,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$ScorecardInningsTableOrderingComposer
    extends Composer<_$AppDatabase, $ScorecardInningsTable> {
  $$ScorecardInningsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get inningNumber => $composableBuilder(
    column: $table.inningNumber,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get half => $composableBuilder(
    column: $table.half,
    builder: (column) => ColumnOrderings(column),
  );

  $$ScorecardGamesTableOrderingComposer get gameUuid {
    final $$ScorecardGamesTableOrderingComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableOrderingComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardInningsTableAnnotationComposer
    extends Composer<_$AppDatabase, $ScorecardInningsTable> {
  $$ScorecardInningsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<int> get inningNumber => $composableBuilder(
    column: $table.inningNumber,
    builder: (column) => column,
  );

  GeneratedColumn<String> get half =>
      $composableBuilder(column: $table.half, builder: (column) => column);

  $$ScorecardGamesTableAnnotationComposer get gameUuid {
    final $$ScorecardGamesTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  Expression<T> scorecardPlaysRefs<T extends Object>(
    Expression<T> Function($$ScorecardPlaysTableAnnotationComposer a) f,
  ) {
    final $$ScorecardPlaysTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.id,
      referencedTable: $db.scorecardPlays,
      getReferencedColumn: (t) => t.inningId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardPlaysTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardPlays,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$ScorecardInningsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ScorecardInningsTable,
          ScorecardInningRow,
          $$ScorecardInningsTableFilterComposer,
          $$ScorecardInningsTableOrderingComposer,
          $$ScorecardInningsTableAnnotationComposer,
          $$ScorecardInningsTableCreateCompanionBuilder,
          $$ScorecardInningsTableUpdateCompanionBuilder,
          (ScorecardInningRow, $$ScorecardInningsTableReferences),
          ScorecardInningRow,
          PrefetchHooks Function({bool gameUuid, bool scorecardPlaysRefs})
        > {
  $$ScorecardInningsTableTableManager(
    _$AppDatabase db,
    $ScorecardInningsTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$ScorecardInningsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$ScorecardInningsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$ScorecardInningsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                Value<String> gameUuid = const Value.absent(),
                Value<int> inningNumber = const Value.absent(),
                Value<String> half = const Value.absent(),
              }) => ScorecardInningsCompanion(
                id: id,
                gameUuid: gameUuid,
                inningNumber: inningNumber,
                half: half,
              ),
          createCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                required String gameUuid,
                required int inningNumber,
                required String half,
              }) => ScorecardInningsCompanion.insert(
                id: id,
                gameUuid: gameUuid,
                inningNumber: inningNumber,
                half: half,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$ScorecardInningsTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback:
              ({gameUuid = false, scorecardPlaysRefs = false}) {
                return PrefetchHooks(
                  db: db,
                  explicitlyWatchedTables: [
                    if (scorecardPlaysRefs) db.scorecardPlays,
                  ],
                  addJoins:
                      <
                        T extends TableManagerState<
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic
                        >
                      >(state) {
                        if (gameUuid) {
                          state =
                              state.withJoin(
                                    currentTable: table,
                                    currentColumn: table.gameUuid,
                                    referencedTable:
                                        $$ScorecardInningsTableReferences
                                            ._gameUuidTable(db),
                                    referencedColumn:
                                        $$ScorecardInningsTableReferences
                                            ._gameUuidTable(db)
                                            .uuid,
                                  )
                                  as T;
                        }

                        return state;
                      },
                  getPrefetchedDataCallback: (items) async {
                    return [
                      if (scorecardPlaysRefs)
                        await $_getPrefetchedData<
                          ScorecardInningRow,
                          $ScorecardInningsTable,
                          ScorecardPlayRow
                        >(
                          currentTable: table,
                          referencedTable: $$ScorecardInningsTableReferences
                              ._scorecardPlaysRefsTable(db),
                          managerFromTypedResult: (p0) =>
                              $$ScorecardInningsTableReferences(
                                db,
                                table,
                                p0,
                              ).scorecardPlaysRefs,
                          referencedItemsForCurrentItem:
                              (item, referencedItems) => referencedItems.where(
                                (e) => e.inningId == item.id,
                              ),
                          typedResults: items,
                        ),
                    ];
                  },
                );
              },
        ),
      );
}

typedef $$ScorecardInningsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ScorecardInningsTable,
      ScorecardInningRow,
      $$ScorecardInningsTableFilterComposer,
      $$ScorecardInningsTableOrderingComposer,
      $$ScorecardInningsTableAnnotationComposer,
      $$ScorecardInningsTableCreateCompanionBuilder,
      $$ScorecardInningsTableUpdateCompanionBuilder,
      (ScorecardInningRow, $$ScorecardInningsTableReferences),
      ScorecardInningRow,
      PrefetchHooks Function({bool gameUuid, bool scorecardPlaysRefs})
    >;
typedef $$ScorecardPlaysTableCreateCompanionBuilder =
    ScorecardPlaysCompanion Function({
      Value<int> id,
      required String gameUuid,
      required int inningId,
      required String battingTeam,
      required int sequence,
      required int batterIndex,
      required String batterName,
      Value<String?> batterPosition,
      required String outcomeCode,
      Value<String?> putoutSequence,
      required String pitchLogJson,
      required String baseStateBefore,
      required String baseStateAfter,
      required int rbi,
      required bool scored,
      Value<DateTime> recordedAt,
    });
typedef $$ScorecardPlaysTableUpdateCompanionBuilder =
    ScorecardPlaysCompanion Function({
      Value<int> id,
      Value<String> gameUuid,
      Value<int> inningId,
      Value<String> battingTeam,
      Value<int> sequence,
      Value<int> batterIndex,
      Value<String> batterName,
      Value<String?> batterPosition,
      Value<String> outcomeCode,
      Value<String?> putoutSequence,
      Value<String> pitchLogJson,
      Value<String> baseStateBefore,
      Value<String> baseStateAfter,
      Value<int> rbi,
      Value<bool> scored,
      Value<DateTime> recordedAt,
    });

final class $$ScorecardPlaysTableReferences
    extends
        BaseReferences<_$AppDatabase, $ScorecardPlaysTable, ScorecardPlayRow> {
  $$ScorecardPlaysTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static $ScorecardGamesTable _gameUuidTable(_$AppDatabase db) =>
      db.scorecardGames.createAlias(
        $_aliasNameGenerator(
          db.scorecardPlays.gameUuid,
          db.scorecardGames.uuid,
        ),
      );

  $$ScorecardGamesTableProcessedTableManager get gameUuid {
    final $_column = $_itemColumn<String>('game_uuid')!;

    final manager = $$ScorecardGamesTableTableManager(
      $_db,
      $_db.scorecardGames,
    ).filter((f) => f.uuid.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_gameUuidTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }

  static $ScorecardInningsTable _inningIdTable(_$AppDatabase db) =>
      db.scorecardInnings.createAlias(
        $_aliasNameGenerator(
          db.scorecardPlays.inningId,
          db.scorecardInnings.id,
        ),
      );

  $$ScorecardInningsTableProcessedTableManager get inningId {
    final $_column = $_itemColumn<int>('inning_id')!;

    final manager = $$ScorecardInningsTableTableManager(
      $_db,
      $_db.scorecardInnings,
    ).filter((f) => f.id.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_inningIdTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }
}

class $$ScorecardPlaysTableFilterComposer
    extends Composer<_$AppDatabase, $ScorecardPlaysTable> {
  $$ScorecardPlaysTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get battingTeam => $composableBuilder(
    column: $table.battingTeam,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get sequence => $composableBuilder(
    column: $table.sequence,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get batterIndex => $composableBuilder(
    column: $table.batterIndex,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get batterName => $composableBuilder(
    column: $table.batterName,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get batterPosition => $composableBuilder(
    column: $table.batterPosition,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get outcomeCode => $composableBuilder(
    column: $table.outcomeCode,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get putoutSequence => $composableBuilder(
    column: $table.putoutSequence,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get pitchLogJson => $composableBuilder(
    column: $table.pitchLogJson,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get baseStateBefore => $composableBuilder(
    column: $table.baseStateBefore,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get baseStateAfter => $composableBuilder(
    column: $table.baseStateAfter,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get rbi => $composableBuilder(
    column: $table.rbi,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<bool> get scored => $composableBuilder(
    column: $table.scored,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get recordedAt => $composableBuilder(
    column: $table.recordedAt,
    builder: (column) => ColumnFilters(column),
  );

  $$ScorecardGamesTableFilterComposer get gameUuid {
    final $$ScorecardGamesTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableFilterComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  $$ScorecardInningsTableFilterComposer get inningId {
    final $$ScorecardInningsTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.inningId,
      referencedTable: $db.scorecardInnings,
      getReferencedColumn: (t) => t.id,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardInningsTableFilterComposer(
            $db: $db,
            $table: $db.scorecardInnings,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardPlaysTableOrderingComposer
    extends Composer<_$AppDatabase, $ScorecardPlaysTable> {
  $$ScorecardPlaysTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get battingTeam => $composableBuilder(
    column: $table.battingTeam,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get sequence => $composableBuilder(
    column: $table.sequence,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get batterIndex => $composableBuilder(
    column: $table.batterIndex,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get batterName => $composableBuilder(
    column: $table.batterName,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get batterPosition => $composableBuilder(
    column: $table.batterPosition,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get outcomeCode => $composableBuilder(
    column: $table.outcomeCode,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get putoutSequence => $composableBuilder(
    column: $table.putoutSequence,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get pitchLogJson => $composableBuilder(
    column: $table.pitchLogJson,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get baseStateBefore => $composableBuilder(
    column: $table.baseStateBefore,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get baseStateAfter => $composableBuilder(
    column: $table.baseStateAfter,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get rbi => $composableBuilder(
    column: $table.rbi,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<bool> get scored => $composableBuilder(
    column: $table.scored,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get recordedAt => $composableBuilder(
    column: $table.recordedAt,
    builder: (column) => ColumnOrderings(column),
  );

  $$ScorecardGamesTableOrderingComposer get gameUuid {
    final $$ScorecardGamesTableOrderingComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableOrderingComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  $$ScorecardInningsTableOrderingComposer get inningId {
    final $$ScorecardInningsTableOrderingComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.inningId,
      referencedTable: $db.scorecardInnings,
      getReferencedColumn: (t) => t.id,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardInningsTableOrderingComposer(
            $db: $db,
            $table: $db.scorecardInnings,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardPlaysTableAnnotationComposer
    extends Composer<_$AppDatabase, $ScorecardPlaysTable> {
  $$ScorecardPlaysTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get battingTeam => $composableBuilder(
    column: $table.battingTeam,
    builder: (column) => column,
  );

  GeneratedColumn<int> get sequence =>
      $composableBuilder(column: $table.sequence, builder: (column) => column);

  GeneratedColumn<int> get batterIndex => $composableBuilder(
    column: $table.batterIndex,
    builder: (column) => column,
  );

  GeneratedColumn<String> get batterName => $composableBuilder(
    column: $table.batterName,
    builder: (column) => column,
  );

  GeneratedColumn<String> get batterPosition => $composableBuilder(
    column: $table.batterPosition,
    builder: (column) => column,
  );

  GeneratedColumn<String> get outcomeCode => $composableBuilder(
    column: $table.outcomeCode,
    builder: (column) => column,
  );

  GeneratedColumn<String> get putoutSequence => $composableBuilder(
    column: $table.putoutSequence,
    builder: (column) => column,
  );

  GeneratedColumn<String> get pitchLogJson => $composableBuilder(
    column: $table.pitchLogJson,
    builder: (column) => column,
  );

  GeneratedColumn<String> get baseStateBefore => $composableBuilder(
    column: $table.baseStateBefore,
    builder: (column) => column,
  );

  GeneratedColumn<String> get baseStateAfter => $composableBuilder(
    column: $table.baseStateAfter,
    builder: (column) => column,
  );

  GeneratedColumn<int> get rbi =>
      $composableBuilder(column: $table.rbi, builder: (column) => column);

  GeneratedColumn<bool> get scored =>
      $composableBuilder(column: $table.scored, builder: (column) => column);

  GeneratedColumn<DateTime> get recordedAt => $composableBuilder(
    column: $table.recordedAt,
    builder: (column) => column,
  );

  $$ScorecardGamesTableAnnotationComposer get gameUuid {
    final $$ScorecardGamesTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.gameUuid,
      referencedTable: $db.scorecardGames,
      getReferencedColumn: (t) => t.uuid,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardGamesTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardGames,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  $$ScorecardInningsTableAnnotationComposer get inningId {
    final $$ScorecardInningsTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.inningId,
      referencedTable: $db.scorecardInnings,
      getReferencedColumn: (t) => t.id,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$ScorecardInningsTableAnnotationComposer(
            $db: $db,
            $table: $db.scorecardInnings,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$ScorecardPlaysTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ScorecardPlaysTable,
          ScorecardPlayRow,
          $$ScorecardPlaysTableFilterComposer,
          $$ScorecardPlaysTableOrderingComposer,
          $$ScorecardPlaysTableAnnotationComposer,
          $$ScorecardPlaysTableCreateCompanionBuilder,
          $$ScorecardPlaysTableUpdateCompanionBuilder,
          (ScorecardPlayRow, $$ScorecardPlaysTableReferences),
          ScorecardPlayRow,
          PrefetchHooks Function({bool gameUuid, bool inningId})
        > {
  $$ScorecardPlaysTableTableManager(
    _$AppDatabase db,
    $ScorecardPlaysTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$ScorecardPlaysTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$ScorecardPlaysTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$ScorecardPlaysTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                Value<String> gameUuid = const Value.absent(),
                Value<int> inningId = const Value.absent(),
                Value<String> battingTeam = const Value.absent(),
                Value<int> sequence = const Value.absent(),
                Value<int> batterIndex = const Value.absent(),
                Value<String> batterName = const Value.absent(),
                Value<String?> batterPosition = const Value.absent(),
                Value<String> outcomeCode = const Value.absent(),
                Value<String?> putoutSequence = const Value.absent(),
                Value<String> pitchLogJson = const Value.absent(),
                Value<String> baseStateBefore = const Value.absent(),
                Value<String> baseStateAfter = const Value.absent(),
                Value<int> rbi = const Value.absent(),
                Value<bool> scored = const Value.absent(),
                Value<DateTime> recordedAt = const Value.absent(),
              }) => ScorecardPlaysCompanion(
                id: id,
                gameUuid: gameUuid,
                inningId: inningId,
                battingTeam: battingTeam,
                sequence: sequence,
                batterIndex: batterIndex,
                batterName: batterName,
                batterPosition: batterPosition,
                outcomeCode: outcomeCode,
                putoutSequence: putoutSequence,
                pitchLogJson: pitchLogJson,
                baseStateBefore: baseStateBefore,
                baseStateAfter: baseStateAfter,
                rbi: rbi,
                scored: scored,
                recordedAt: recordedAt,
              ),
          createCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                required String gameUuid,
                required int inningId,
                required String battingTeam,
                required int sequence,
                required int batterIndex,
                required String batterName,
                Value<String?> batterPosition = const Value.absent(),
                required String outcomeCode,
                Value<String?> putoutSequence = const Value.absent(),
                required String pitchLogJson,
                required String baseStateBefore,
                required String baseStateAfter,
                required int rbi,
                required bool scored,
                Value<DateTime> recordedAt = const Value.absent(),
              }) => ScorecardPlaysCompanion.insert(
                id: id,
                gameUuid: gameUuid,
                inningId: inningId,
                battingTeam: battingTeam,
                sequence: sequence,
                batterIndex: batterIndex,
                batterName: batterName,
                batterPosition: batterPosition,
                outcomeCode: outcomeCode,
                putoutSequence: putoutSequence,
                pitchLogJson: pitchLogJson,
                baseStateBefore: baseStateBefore,
                baseStateAfter: baseStateAfter,
                rbi: rbi,
                scored: scored,
                recordedAt: recordedAt,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$ScorecardPlaysTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: ({gameUuid = false, inningId = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [],
              addJoins:
                  <
                    T extends TableManagerState<
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic
                    >
                  >(state) {
                    if (gameUuid) {
                      state =
                          state.withJoin(
                                currentTable: table,
                                currentColumn: table.gameUuid,
                                referencedTable: $$ScorecardPlaysTableReferences
                                    ._gameUuidTable(db),
                                referencedColumn:
                                    $$ScorecardPlaysTableReferences
                                        ._gameUuidTable(db)
                                        .uuid,
                              )
                              as T;
                    }
                    if (inningId) {
                      state =
                          state.withJoin(
                                currentTable: table,
                                currentColumn: table.inningId,
                                referencedTable: $$ScorecardPlaysTableReferences
                                    ._inningIdTable(db),
                                referencedColumn:
                                    $$ScorecardPlaysTableReferences
                                        ._inningIdTable(db)
                                        .id,
                              )
                              as T;
                    }

                    return state;
                  },
              getPrefetchedDataCallback: (items) async {
                return [];
              },
            );
          },
        ),
      );
}

typedef $$ScorecardPlaysTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ScorecardPlaysTable,
      ScorecardPlayRow,
      $$ScorecardPlaysTableFilterComposer,
      $$ScorecardPlaysTableOrderingComposer,
      $$ScorecardPlaysTableAnnotationComposer,
      $$ScorecardPlaysTableCreateCompanionBuilder,
      $$ScorecardPlaysTableUpdateCompanionBuilder,
      (ScorecardPlayRow, $$ScorecardPlaysTableReferences),
      ScorecardPlayRow,
      PrefetchHooks Function({bool gameUuid, bool inningId})
    >;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$AppSettingsTableTableManager get appSettings =>
      $$AppSettingsTableTableManager(_db, _db.appSettings);
  $$RecentPlayersTableTableManager get recentPlayers =>
      $$RecentPlayersTableTableManager(_db, _db.recentPlayers);
  $$ScorecardGamesTableTableManager get scorecardGames =>
      $$ScorecardGamesTableTableManager(_db, _db.scorecardGames);
  $$ScorecardLineupsTableTableManager get scorecardLineups =>
      $$ScorecardLineupsTableTableManager(_db, _db.scorecardLineups);
  $$ScorecardInningsTableTableManager get scorecardInnings =>
      $$ScorecardInningsTableTableManager(_db, _db.scorecardInnings);
  $$ScorecardPlaysTableTableManager get scorecardPlays =>
      $$ScorecardPlaysTableTableManager(_db, _db.scorecardPlays);
}
