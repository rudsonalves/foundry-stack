import 'dart:async';

import 'package:dio/dio.dart';

import '/core/services/auth_token_store/auth_token_store.dart';
import '/core/services/logging/console_log.dart';

class AuthInterceptor extends Interceptor {
  static const _refreshPath = '/auth/refresh';
  static const _refreshRetryKey = 'auth_refresh_retry';
  static const _refreshableErrorCodes = {
    'UNAUTHORIZED',
    'INVALID_TOKEN',
  };

  final Dio _authDio;
  final Dio _refreshDio;
  final AuthTokenStore _authTokenStore;
  Future<String>? _refreshInFlight;

  AuthInterceptor({
    required Dio authDio,
    required AuthTokenStore authTokenStore,
    required String baseUrl,
    Duration timeout = const Duration(seconds: 10),
    Dio? refreshDio,
  }) : _authDio = authDio,
       _authTokenStore = authTokenStore,
       _refreshDio =
           refreshDio ??
           Dio(
             BaseOptions(
               baseUrl: baseUrl,
               connectTimeout: timeout,
               receiveTimeout: timeout,
               headers: const {
                 'Accept': 'application/json',
               },
             ),
           );

  final _log = ConsoleLog('AuthInterceptor');

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // If the request already has an Authorization header, skip adding the token
    if (options.headers.containsKey('Authorization')) {
      return handler.next(options);
    }

    final tokenResult = await _authTokenStore.readAccessToken();

    tokenResult.fold(
      onSuccess: (token) {
        if (token.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $token';
        }
      },
      onFailure: (_) {},
    );

    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final statusCode = err.response?.statusCode;
    final path = err.requestOptions.path;

    if (statusCode == null) {
      _log.error(
        '[AuthInterceptor] ${err.requestOptions.method} $path failed: '
        '${err.message}',
        error: err,
        stack: err.stackTrace,
      );
    }

    final backendErrorCode = _backendErrorCode(err.response?.data);
    final canRefresh =
        statusCode == 401 &&
        !path.endsWith(_refreshPath) &&
        err.requestOptions.extra[_refreshRetryKey] != true &&
        _refreshableErrorCodes.contains(backendErrorCode);

    if (!canRefresh) {
      if (statusCode != null && statusCode >= 500) {
        _log.error(
          '[AuthInterceptor] HTTP $statusCode → '
          '${err.requestOptions.method} $path',
          error: err,
          stack: err.stackTrace,
        );
      } else if (statusCode != null) {
        _log.warn(
          '[AuthInterceptor] HTTP $statusCode → '
          '${err.requestOptions.method} $path'
          '${backendErrorCode == null ? '' : ' ($backendErrorCode)'}',
        );
      }

      return handler.next(err);
    }

    _log.warn(
      '[AuthInterceptor] HTTP 401 → ${err.requestOptions.method} $path; '
      'attempting token refresh.',
    );

    final alreadyRefreshedToken = await _accessTokenUpdatedAfter(err);
    if (alreadyRefreshedToken != null) {
      try {
        final retryResponse = await _retryWithToken(
          err.requestOptions,
          alreadyRefreshedToken,
        );
        return handler.resolve(retryResponse);
      } catch (e, s) {
        _log.error('[AuthInterceptor] retry failed: $e', error: e, stack: s);
        return handler.next(err);
      }
    }

    final refreshResult = await _authTokenStore.readRefreshToken();

    if (refreshResult.isFailure) {
      await _clearSession();
      return handler.next(err);
    }

    final refreshToken = refreshResult.value!;

    if (refreshToken.isEmpty) {
      await _clearSession();
      return handler.next(err);
    }

    late final String newAccessToken;
    try {
      newAccessToken = await _refreshAccessToken(refreshToken);
    } catch (e, s) {
      _log.error('[AuthInterceptor] refresh failed: $e', error: e, stack: s);

      if (_isDefinitiveRefreshRejection(e)) {
        await _clearSession();
        return handler.next(err);
      }

      return handler.next(err);
    }

    try {
      final retryResponse = await _retryWithToken(
        err.requestOptions,
        newAccessToken,
      );

      return handler.resolve(retryResponse);
    } catch (e, s) {
      _log.error('[AuthInterceptor] retry failed: $e', error: e, stack: s);
      return handler.next(err);
    }
  }

  Future<void> _clearSession() async {
    await _authTokenStore.clearTokens();
  }

  Future<String> _refreshAccessToken(String refreshToken) {
    final inFlight = _refreshInFlight;
    if (inFlight != null) {
      return inFlight;
    }

    final refresh = _performRefresh(refreshToken);
    _refreshInFlight = refresh;

    return refresh.whenComplete(() {
      if (identical(_refreshInFlight, refresh)) {
        _refreshInFlight = null;
      }
    });
  }

  Future<String> _performRefresh(String refreshToken) async {
    final response = await _refreshDio.post(
      _refreshPath,
      data: {'refresh_token': refreshToken},
    );

    final data = response.data;
    final payload = data is Map<String, dynamic> ? data : {};
    final inner = payload['data'] ?? payload;

    final newAccessToken = inner['access_token'] as String?;
    if (newAccessToken == null || newAccessToken.isEmpty) {
      throw Exception('Invalid refresh response');
    }

    final updateResult = await _authTokenStore.updateAccessToken(
      newAccessToken,
    );
    if (updateResult.isFailure) {
      throw StateError('Failed to update access token');
    }

    return newAccessToken;
  }

  Future<Response<dynamic>> _retryWithToken(
    RequestOptions request,
    String accessToken,
  ) {
    final newRequest =
        Options(
          method: request.method,
          headers: {
            ...request.headers,
            'Authorization': 'Bearer $accessToken',
          },
          responseType: request.responseType,
          contentType: request.contentType,
          extra: {
            ...request.extra,
            _refreshRetryKey: true,
          },
          followRedirects: request.followRedirects,
          validateStatus: request.validateStatus,
        ).compose(
          _authDio.options,
          request.path,
          data: request.data,
          queryParameters: request.queryParameters,
        );

    return _authDio.fetch(newRequest);
  }

  Future<String?> _accessTokenUpdatedAfter(DioException err) async {
    final failedToken = _bearerToken(
      err.requestOptions.headers['Authorization'],
    );
    if (failedToken == null || failedToken.isEmpty) {
      return null;
    }

    final currentTokenResult = await _authTokenStore.readAccessToken();
    if (currentTokenResult.isFailure) {
      return null;
    }

    final currentToken = currentTokenResult.value;
    if (currentToken == null ||
        currentToken.isEmpty ||
        currentToken == failedToken) {
      return null;
    }

    return currentToken;
  }

  String? _bearerToken(Object? authorization) {
    if (authorization is! String) {
      return null;
    }

    final parts = authorization.trim().split(RegExp(r'\s+'));
    if (parts.length != 2 || parts.first != 'Bearer') {
      return null;
    }

    return parts.last;
  }

  String? _backendErrorCode(Object? data) {
    if (data is! Map) {
      return null;
    }

    final error = data['error'];
    if (error is! Map) {
      return null;
    }

    final code = error['code'];
    return code is String ? code : null;
  }

  bool _isDefinitiveRefreshRejection(Object error) {
    return error is DioException && error.response?.statusCode == 401;
  }
}
