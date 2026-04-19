import 'package:bigfly_mobile/data/models/json_helpers.dart';

class CoverageRange {
  const CoverageRange({required this.from, required this.to});

  final int? from;
  final int? to;

  factory CoverageRange.fromJson(Map<String, dynamic> json) {
    return CoverageRange(from: nullableInt(json['from']), to: nullableInt(json['to']));
  }
}

class DatasetStatusSnapshot {
  const DatasetStatusSnapshot({
    required this.id,
    required this.name,
    required this.source,
    required this.required,
    required this.healthy,
    required this.rowCount,
    required this.coverageFrom,
    required this.coverageTo,
  });

  final String id;
  final String name;
  final String source;
  final bool required;
  final bool healthy;
  final int rowCount;
  final int? coverageFrom;
  final int? coverageTo;

  factory DatasetStatusSnapshot.fromJson(Map<String, dynamic> json) {
    return DatasetStatusSnapshot(
      id: stringOrEmpty(json['id']),
      name: stringOrEmpty(json['name']),
      source: stringOrEmpty(json['source']),
      required: boolOrFalse(json['required']),
      healthy: boolOrFalse(json['healthy']),
      rowCount: intOrZero(json['row_count']),
      coverageFrom: nullableInt(json['coverage_from']),
      coverageTo: nullableInt(json['coverage_to']),
    );
  }
}

class MetaSnapshot {
  const MetaSnapshot({
    required this.version,
    required this.generatedAt,
    required this.coverage,
    required this.datasets,
  });

  final String version;
  final DateTime? generatedAt;
  final Map<String, CoverageRange> coverage;
  final List<DatasetStatusSnapshot> datasets;

  factory MetaSnapshot.fromJson(Map<String, dynamic> json) {
    final coverageRaw = asJsonMap(json['coverage']);
    final coverage = <String, CoverageRange>{};
    for (final entry in coverageRaw.entries) {
      coverage[entry.key] = CoverageRange.fromJson(asJsonMap(entry.value));
    }

    return MetaSnapshot(
      version: stringOrEmpty(json['version']),
      generatedAt: DateTime.tryParse(stringOrEmpty(json['generated_at'])),
      coverage: coverage,
      datasets: asJsonMapList(json['datasets']).map(DatasetStatusSnapshot.fromJson).toList(growable: false),
    );
  }

  int? get minCoverageYear {
    final years = coverage.values.map((item) => item.from).whereType<int>().toList(growable: false);
    if (years.isEmpty) {
      return null;
    }
    years.sort();
    return years.first;
  }

  int? get maxCoverageYear {
    final years = coverage.values.map((item) => item.to).whereType<int>().toList(growable: false);
    if (years.isEmpty) {
      return null;
    }
    years.sort();
    return years.last;
  }

  int get sourceCount => coverage.keys.length;

  bool get allRequiredHealthy {
    final requiredDatasets = datasets.where((dataset) => dataset.required);
    if (requiredDatasets.isEmpty) {
      return true;
    }
    return requiredDatasets.every((dataset) => dataset.healthy);
  }
}
