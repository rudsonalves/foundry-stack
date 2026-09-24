import 'package:dio/dio.dart';

import '../../../resources/app_env.dart';
import '../../../resources/app_http_headers.dart';

abstract final class DioFactory {
  static Dio create({
    String? baseUrl,
    Map<String, String>? defaultHeaders,
    List<Interceptor> interceptors = const [],
  }) {
    final dio = Dio(
      BaseOptions(
        baseUrl: baseUrl ?? AppEnv.baseUrl,
        connectTimeout: Duration(milliseconds: AppEnv.connectTimeout),
        receiveTimeout: Duration(milliseconds: AppEnv.receiveTimeout),
        headers: {
          AppHttpHeaders.accept: 'application/json',
          AppHttpHeaders.contentType: 'application/json',
          ...?defaultHeaders,
        },
      ),
    );

    for (final interceptor in interceptors) {
      if (dio.interceptors.every(
        (item) => item.runtimeType != interceptor.runtimeType,
      )) {
        dio.interceptors.add(interceptor);
      }
    }

    return dio;
  }
}
