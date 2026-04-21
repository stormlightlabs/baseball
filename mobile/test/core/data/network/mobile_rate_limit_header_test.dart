import 'dart:convert';
import 'dart:typed_data';

import 'package:bigfly_mobile/core/data/network/baseball_api_client.dart';
import 'package:bigfly_mobile/main.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

class _CaptureAdapter implements HttpClientAdapter {
  RequestOptions? lastOptions;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    lastOptions = options;
    return ResponseBody.fromString(
      jsonEncode(<String, dynamic>{'ok': true}),
      200,
      headers: <String, List<String>>{
        Headers.contentTypeHeader: <String>[Headers.jsonContentType],
      },
    );
  }
}

void main() {
  test('mobile API requests include first-party client header', () async {
    final dio = buildApiDio();
    final adapter = _CaptureAdapter();
    dio.httpClientAdapter = adapter;

    final client = BaseballApiClient(dio);
    await client.getMeta();

    expect(adapter.lastOptions, isNotNull);
    expect(adapter.lastOptions!.path, '/v1/meta');
    expect(adapter.lastOptions!.headers['X-BigFly-Client'], 'mobile');
  });
}
