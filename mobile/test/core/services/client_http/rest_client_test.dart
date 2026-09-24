import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/resources/app_http_headers.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client_http.dart';

void main() {
  test('DioRestClient forwards a request and maps its response', () async {
    late RequestOptions capturedOptions;
    final dio = Dio();
    dio.httpClientAdapter = _FakeHttpClientAdapter((options) async {
      capturedOptions = options;
      return ResponseBody.fromString(
        jsonEncode({'data': []}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    });
    final client = DioRestClient(dio);

    final result = await client.get(
      const RestClientRequest(
        '/listings',
        headers: {AppHttpHeaders.authorization: 'Bearer secret'},
        queryParameters: {'page': 1},
      ),
    );

    expect(result, isA<Success<RestClientResponse>>());
    expect(capturedOptions.method, 'GET');
    expect(capturedOptions.path, '/listings');
    expect(capturedOptions.queryParameters, {'page': 1});
    expect(result.value?.statusCode, 200);
    expect(result.value?.data, {'data': []});
  });

  test('RestClientResponse identifies successful status codes', () {
    expect(const RestClientResponse(statusCode: 200).isSuccess, isTrue);
    expect(const RestClientResponse(statusCode: 299).isSuccess, isTrue);
    expect(const RestClientResponse(statusCode: 400).isSuccess, isFalse);
  });
}

final class _FakeHttpClientAdapter implements HttpClientAdapter {
  final Future<ResponseBody> Function(RequestOptions options) handler;

  _FakeHttpClientAdapter(this.handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) {
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}
