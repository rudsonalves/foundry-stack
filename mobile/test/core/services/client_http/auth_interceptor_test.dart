import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/dtos/app_tokens.dart';
import 'package:foundry_stack_mobile/core/services/client_http/interceptors/auth/auth_interceptor.dart';

void main() {
  group('AuthInterceptor', () {
    test(
      'sends request without Authorization when access token is absent',
      () async {
        final tokenStore = _FakeAuthTokenStore(accessToken: null);
        final mainAdapter = _FakeHttpClientAdapter((options) async {
          return _jsonResponse(200, const {'data': {}});
        });
        final harness = _Harness(
          tokenStore: tokenStore,
          mainAdapter: mainAdapter,
        );

        await harness.dio.get('/protected');

        expect(mainAdapter.requests, hasLength(1));
        expect(
          mainAdapter.requests.single.headers.containsKey('Authorization'),
          isFalse,
        );
        expect(harness.refreshAdapter.requests, isEmpty);
      },
    );

    test('shares one refresh between concurrent requests', () async {
      final refreshStarted = Completer<void>();
      final releaseRefresh = Completer<void>();
      final tokenStore = _FakeAuthTokenStore(
        accessToken: _oldAccessToken,
        refreshToken: _refreshToken,
      );
      final mainAdapter = _protectedAdapter();
      final refreshAdapter = _FakeHttpClientAdapter((options) async {
        if (!refreshStarted.isCompleted) {
          refreshStarted.complete();
        }
        await releaseRefresh.future;
        return _successfulRefresh();
      });
      final harness = _Harness(
        tokenStore: tokenStore,
        mainAdapter: mainAdapter,
        refreshAdapter: refreshAdapter,
      );

      final first = harness.dio.get('/first');
      final second = harness.dio.get('/second');
      await refreshStarted.future;
      await Future<void>.delayed(Duration.zero);
      releaseRefresh.complete();

      final responses = await Future.wait([first, second]);

      expect(responses.map((response) => response.statusCode), [200, 200]);
      expect(refreshAdapter.requests, hasLength(1));
      expect(tokenStore.updateAccessCalls, 1);
    });

    test('retries the original request with the renewed Bearer', () async {
      final tokenStore = _FakeAuthTokenStore(
        accessToken: _oldAccessToken,
        refreshToken: _refreshToken,
      );
      final mainAdapter = _protectedAdapter();
      final harness = _Harness(
        tokenStore: tokenStore,
        mainAdapter: mainAdapter,
        refreshAdapter: _FakeHttpClientAdapter(
          (_) async => _successfulRefresh(),
        ),
      );

      final response = await harness.dio.post(
        '/protected',
        data: const {'value': 1},
        queryParameters: const {'page': 2},
      );

      expect(response.statusCode, 200);
      expect(mainAdapter.requests, hasLength(2));
      final retry = mainAdapter.requests.last;
      expect(retry.method, 'POST');
      expect(retry.path, '/protected');
      expect(retry.data, const {'value': 1});
      expect(retry.queryParameters, const {'page': 2});
      expect(retry.headers['Authorization'], 'Bearer $_newAccessToken');
    });

    test('clears tokens when refresh is rejected with 401', () async {
      final tokenStore = _FakeAuthTokenStore(
        accessToken: _oldAccessToken,
        refreshToken: _refreshToken,
      );
      final harness = _Harness(
        tokenStore: tokenStore,
        mainAdapter: _protectedAdapter(),
        refreshAdapter: _FakeHttpClientAdapter((_) async {
          return _jsonResponse(
            401,
            const {
              'error': {'code': 'UNAUTHORIZED', 'message': 'rejected'},
            },
          );
        }),
      );

      await expectLater(
        harness.dio.get('/protected'),
        throwsA(isA<DioException>()),
      );

      expect(tokenStore.clearCalls, 1);
      expect(tokenStore.accessToken, isNull);
      expect(tokenStore.refreshToken, isNull);
    });

    test('preserves tokens when refresh has a connection failure', () async {
      final tokenStore = _FakeAuthTokenStore(
        accessToken: _oldAccessToken,
        refreshToken: _refreshToken,
      );
      final harness = _Harness(
        tokenStore: tokenStore,
        mainAdapter: _protectedAdapter(),
        refreshAdapter: _FakeHttpClientAdapter((options) async {
          throw DioException(
            requestOptions: options,
            type: DioExceptionType.connectionError,
            message: 'offline',
          );
        }),
      );

      await expectLater(
        harness.dio.get('/protected'),
        throwsA(isA<DioException>()),
      );

      expect(tokenStore.clearCalls, 0);
      expect(tokenStore.hasTokens, isTrue);
    });

    test(
      'does not refresh again when the retried request returns 401',
      () async {
        final tokenStore = _FakeAuthTokenStore(
          accessToken: _oldAccessToken,
          refreshToken: _refreshToken,
        );
        final mainAdapter = _FakeHttpClientAdapter((_) async {
          return _refreshableUnauthorized();
        });
        final refreshAdapter = _FakeHttpClientAdapter(
          (_) async => _successfulRefresh(),
        );
        final harness = _Harness(
          tokenStore: tokenStore,
          mainAdapter: mainAdapter,
          refreshAdapter: refreshAdapter,
        );

        await expectLater(
          harness.dio.get('/protected'),
          throwsA(isA<DioException>()),
        );

        expect(mainAdapter.requests, hasLength(2));
        expect(refreshAdapter.requests, hasLength(1));
        expect(tokenStore.updateAccessCalls, 1);
      },
    );
  });
}

const _baseUrl = 'https://api.example.test';
const _oldAccessToken = 'old-access-token';
const _newAccessToken = 'new-access-token';
const _refreshToken = 'refresh-token';

class _Harness {
  late final Dio dio;
  late final _FakeHttpClientAdapter refreshAdapter;

  _Harness({
    required _FakeAuthTokenStore tokenStore,
    required _FakeHttpClientAdapter mainAdapter,
    _FakeHttpClientAdapter? refreshAdapter,
  }) {
    dio = Dio(BaseOptions(baseUrl: _baseUrl));
    dio.httpClientAdapter = mainAdapter;

    this.refreshAdapter =
        refreshAdapter ??
        _FakeHttpClientAdapter((_) async => _successfulRefresh());
    final refreshDio = Dio(BaseOptions(baseUrl: _baseUrl));
    refreshDio.httpClientAdapter = this.refreshAdapter;

    dio.interceptors.add(
      AuthInterceptor(
        authDio: dio,
        authTokenStore: tokenStore,
        baseUrl: _baseUrl,
        refreshDio: refreshDio,
      ),
    );
  }
}

_FakeHttpClientAdapter _protectedAdapter() {
  return _FakeHttpClientAdapter((options) async {
    if (options.headers['Authorization'] == 'Bearer $_newAccessToken') {
      return _jsonResponse(200, const {'data': {}});
    }
    return _refreshableUnauthorized();
  });
}

ResponseBody _successfulRefresh() {
  return _jsonResponse(
    200,
    const {
      'data': {'access_token': _newAccessToken},
    },
  );
}

ResponseBody _refreshableUnauthorized() {
  return _jsonResponse(
    401,
    const {
      'error': {'code': 'INVALID_TOKEN', 'message': 'expired'},
    },
  );
}

ResponseBody _jsonResponse(int statusCode, Object body) {
  return ResponseBody.fromString(
    jsonEncode(body),
    statusCode,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );
}

class _FakeHttpClientAdapter implements HttpClientAdapter {
  final Future<ResponseBody> Function(RequestOptions options) _handler;
  final requests = <RequestOptions>[];

  _FakeHttpClientAdapter(this._handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return _handler(options);
  }

  @override
  void close({bool force = false}) {}
}

class _FakeAuthTokenStore implements AuthTokenStore {
  String? accessToken;
  String? refreshToken;
  var updateAccessCalls = 0;
  var clearCalls = 0;

  _FakeAuthTokenStore({
    this.accessToken,
    this.refreshToken,
  });

  bool get hasTokens => accessToken != null && refreshToken != null;

  @override
  AsyncResult<String> readAccessToken() async {
    final token = accessToken;
    if (token == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.storageNotFound,
          message: 'access token not found',
        ),
      );
    }
    return Success(token);
  }

  @override
  AsyncResult<String> readRefreshToken() async {
    final token = refreshToken;
    if (token == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.storageNotFound,
          message: 'refresh token not found',
        ),
      );
    }
    return Success(token);
  }

  @override
  AsyncResult<Unit> updateAccessToken(String newAccessToken) async {
    updateAccessCalls++;
    accessToken = newAccessToken;
    return const Success(unit);
  }

  @override
  AsyncResult<Unit> clearTokens() async {
    clearCalls++;
    accessToken = null;
    refreshToken = null;
    return const Success(unit);
  }

  @override
  AsyncResult<Unit> saveTokens(AppTokens tokens) async {
    accessToken = tokens.accessToken;
    refreshToken = tokens.refreshToken;
    return const Success(unit);
  }
}
