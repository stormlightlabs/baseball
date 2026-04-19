import 'package:json_annotation/json_annotation.dart';

part 'health_response.g.dart';

@JsonSerializable()
class HealthResponse {
  const HealthResponse({required this.status});

  factory HealthResponse.fromJson(Map<String, dynamic> json) => _$HealthResponseFromJson(json);

  final String status;

  Map<String, dynamic> toJson() => _$HealthResponseToJson(this);
}
