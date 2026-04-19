import 'package:bigfly_mobile/features/home/data/models/meta_models.dart';

class DataSourcesSnapshot {
  const DataSourcesSnapshot({required this.meta, required this.datasets});

  final MetaSnapshot meta;
  final List<DatasetStatusSnapshot> datasets;

  int get totalRows => datasets.fold<int>(0, (sum, item) => sum + item.rowCount);

  int? get minCoverage {
    final values = datasets.map((item) => item.coverageFrom).whereType<int>().toList(growable: false);
    if (values.isEmpty) {
      return null;
    }
    values.sort();
    return values.first;
  }

  int? get maxCoverage {
    final values = datasets.map((item) => item.coverageTo).whereType<int>().toList(growable: false);
    if (values.isEmpty) {
      return null;
    }
    values.sort();
    return values.last;
  }
}
