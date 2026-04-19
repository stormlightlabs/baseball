Map<String, dynamic> asJsonMap(dynamic value) {
  if (value is Map<String, dynamic>) {
    return value;
  }
  if (value is Map) {
    return value.map((key, val) => MapEntry(key.toString(), val));
  }
  return <String, dynamic>{};
}

List<Map<String, dynamic>> asJsonMapList(dynamic value) {
  if (value is! List) {
    return <Map<String, dynamic>>[];
  }
  return value.map((item) => asJsonMap(item)).toList(growable: false);
}

String stringOrEmpty(dynamic value) {
  if (value == null) {
    return '';
  }
  return value.toString();
}

String? nullableString(dynamic value) {
  final normalized = stringOrEmpty(value).trim();
  return normalized.isEmpty ? null : normalized;
}

int intOrZero(dynamic value) {
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  if (value is String) {
    return int.tryParse(value) ?? 0;
  }
  return 0;
}

int? nullableInt(dynamic value) {
  if (value == null) {
    return null;
  }
  final parsed = intOrZero(value);
  return parsed == 0 && value.toString() != '0' ? null : parsed;
}

double doubleOrZero(dynamic value) {
  if (value is double) {
    return value;
  }
  if (value is num) {
    return value.toDouble();
  }
  if (value is String) {
    return double.tryParse(value) ?? 0;
  }
  return 0;
}

double? nullableDouble(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is String && value.trim().isEmpty) {
    return null;
  }
  return doubleOrZero(value);
}

bool boolOrFalse(dynamic value) {
  if (value is bool) {
    return value;
  }
  if (value is String) {
    return value.toLowerCase() == 'true';
  }
  return false;
}
