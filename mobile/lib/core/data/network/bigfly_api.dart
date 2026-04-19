import 'package:bigfly_mobile/features/home/data/models/health_response.dart';
import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

part 'bigfly_api.g.dart';

@RestApi()
abstract class BigFlyApi {
  factory BigFlyApi(Dio dio, {String baseUrl}) = _BigFlyApi;

  @GET('/v1/health')
  Future<HealthResponse> getHealth();
}
