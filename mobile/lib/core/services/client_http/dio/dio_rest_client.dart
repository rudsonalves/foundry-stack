import 'package:dio/dio.dart';

import '../../../resources/app_http_headers.dart';
import '../../../result/result.dart';
import '../../logging/console_log.dart';
import '../client/rest_client.dart';
import '../client/rest_client_request.dart';
import '../client/rest_client_response.dart';
import 'dio_error_mapper.dart';

final class DioRestClient implements RestClient {
  final Dio _dio;
  final _log = const ConsoleLog('DioRestClient');

  DioRestClient(this._dio);

  @override
  AsyncResult<RestClientResponse> get(RestClientRequest request) {
    return _request(
      'GET',
      request,
      () => _dio.get(
        request.path,
        queryParameters: request.queryParameters,
        options: Options(headers: request.headers),
      ),
    );
  }

  @override
  AsyncResult<RestClientResponse> post(RestClientRequest request) {
    return _request(
      'POST',
      request,
      () => _dio.post(
        request.path,
        data: request.body,
        queryParameters: request.queryParameters,
        options: Options(headers: request.headers),
      ),
    );
  }

  @override
  AsyncResult<RestClientResponse> put(RestClientRequest request) {
    return _request(
      'PUT',
      request,
      () => _dio.put(
        request.path,
        data: request.body,
        queryParameters: request.queryParameters,
        options: Options(headers: request.headers),
      ),
    );
  }

  @override
  AsyncResult<RestClientResponse> patch(RestClientRequest request) {
    return _request(
      'PATCH',
      request,
      () => _dio.patch(
        request.path,
        data: request.body,
        queryParameters: request.queryParameters,
        options: Options(headers: request.headers),
      ),
    );
  }

  @override
  AsyncResult<RestClientResponse> delete(RestClientRequest request) {
    return _request(
      'DELETE',
      request,
      () => _dio.delete(
        request.path,
        data: request.body,
        queryParameters: request.queryParameters,
        options: Options(headers: request.headers),
      ),
    );
  }

  AsyncResult<RestClientResponse> _request(
    String method,
    RestClientRequest request,
    Future<Response<Object?>> Function() call,
  ) async {
    _logRequest(method, request);

    try {
      final response = await call();
      return Success(
        RestClientResponse(
          data: response.data,
          statusCode: response.statusCode,
          statusMessage: response.statusMessage,
        ),
      );
    } catch (error, stackTrace) {
      final appError = mapHttpError(error);
      _log.error(appError.message, error: error, stack: stackTrace);
      return Failure(appError);
    }
  }

  void _logRequest(String method, RestClientRequest request) {
    final headers = request.headers?.map(
      (name, value) => MapEntry(
        name,
        AppHttpHeaders.sensitiveLowercase.contains(name.toLowerCase())
            ? '<redacted>'
            : value,
      ),
    );
    _log.info('$method ${request.path}${headers == null ? '' : ' $headers'}');
  }
}
