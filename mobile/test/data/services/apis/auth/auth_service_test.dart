import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/resources/app_env.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';

void main() {
  group('AuthService.login', () {
    late _FakeRestClient client;
    late AuthService service;

    setUp(() {
      client = _FakeRestClient();
      service = AuthService(client);
    });

    test('sends request to /auth/login', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'access_token': 'access.jwt.token',
              'refresh_token': 'opaque.refresh.token',
              'token_type': 'Bearer',
              'expires_in': 900,
            },
          },
        ),
      );

      await service.login(
        const LoginCredentials(
          email: 'ada@example.com',
          password: 'secret123',
        ),
      );

      expect(client.lastPostRequest, isNotNull);
      expect(client.lastPostRequest!.path, '/auth/login');
    });

    test('translates login credentials to the API body', () async {
      final request = _loginUserRequest();
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            "data": {
              "access_token": "access.jwt.token",
              "refresh_token": "opaque.refresh.token",
              "token_type": "Bearer",
              "expires_in": 900,
            },
          },
        ),
      );

      await service.login(request);

      expect(client.lastPostRequest, isNotNull);
      expect(client.lastPostRequest!.body, {
        'client_id': AppEnv.authClientId,
        'email': request.email,
        'password': request.password,
      });
    });

    test('returns success for 200 response', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            "data": {
              "access_token": "access.jwt.token",
              "refresh_token": "opaque.refresh.token",
              "token_type": "Bearer",
              "expires_in": 900,
            },
          },
        ),
      );

      final result = await service.login(_loginUserRequest());

      expect(result.isSuccess, isTrue);
      expect(result.value?.accessToken, 'access.jwt.token');
      expect(result.value?.refreshToken, 'opaque.refresh.token');
    });

    test('returns failure when status is different from 200', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          statusMessage: 'Created',
          data: {
            'data': {
              'access_token': 'access.jwt.token',
              'refresh_token': 'opaque.refresh.token',
              'token_type': 'Bearer',
              'expires_in': 900,
            },
          },
        ),
      );

      final result = await service.login(_loginUserRequest());

      expect(result.isFailure, isTrue);
      expect(result.error, isA<AppError>());
      final error = result.error!;
      expect(error.statusCode, 201);
      expect(error.code, AppErrorCode.invalidData);
      expect(
        error.message,
        'Unexpected response status: expected 200, received 201',
      );
    });

    test('preserves AppError for 400 and 401', () async {
      const badRequest = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid login body',
      );

      client.postResult = const Failure(badRequest);
      final badRequestResult = await service.login(_loginUserRequest());

      expect(badRequestResult.isFailure, isTrue);
      expect(identical(badRequestResult.error, badRequest), isTrue);

      const unauthorized = AppError(
        statusCode: 401,
        code: AppErrorCode.unauthenticated,
        message: 'invalid credentials',
      );

      client.postResult = const Failure(unauthorized);
      final unauthorizedResult = await service.login(_loginUserRequest());

      expect(unauthorizedResult.isFailure, isTrue);
      expect(identical(unauthorizedResult.error, unauthorized), isTrue);
    });

    test('returns failure on network error', () async {
      final networkError = const AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      client.postResult = Failure(networkError);

      final result = await service.login(_loginUserRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, networkError), isTrue);
    });

    test('returns failure on malformed envelope', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': 'invalid-type',
          },
        ),
      );

      final result = await service.login(_loginUserRequest());

      expect(result.isFailure, isTrue);
      expect(result.error, isA<AppError>());
      final error = result.error!;
      expect(error.code, AppErrorCode.parsingError);
    });

    test('does not expose password or tokens in logs', () async {
      final messages = <String>[];
      final originalDebugPrint = debugPrint;
      debugPrint = (String? message, {int? wrapWidth}) {
        if (message != null) {
          messages.add(message);
        }
      };
      addTearDown(() {
        debugPrint = originalDebugPrint;
      });

      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'access_token': 'SensitiveAccessToken',
              'refresh_token': 'SensitiveRefreshToken',
              'token_type': 'invalid',
              'expires_in': 900,
            },
          },
        ),
      );

      const request = LoginCredentials(
        email: 'ada@example.com',
        password: 'SensitivePassword',
      );

      await service.login(request);

      final allLogs = messages.join('\n');
      expect(messages, isNotEmpty);
      expect(allLogs.contains('SensitivePassword'), isFalse);
      expect(allLogs.contains('SensitiveAccessToken'), isFalse);
      expect(allLogs.contains('SensitiveRefreshToken'), isFalse);
    });
  });

  group('AuthService.refreshToken', () {
    late _FakeRestClient client;
    late AuthService service;

    setUp(() {
      client = _FakeRestClient();
      service = AuthService(client);
    });

    test('sends only the refresh token to /auth/refresh', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'access_token': 'new.access.token',
              'token_type': 'Bearer',
              'expires_in': 900,
            },
          },
        ),
      );

      await service.refreshToken(_refreshRequest());

      expect(client.lastPostRequest?.path, '/auth/refresh');
      expect(client.lastPostRequest?.body, {
        'refresh_token': 'refresh-token',
      });
    });

    test('returns the renewed access token for a valid response', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'access_token': 'new.access.token',
              'token_type': 'Bearer',
              'expires_in': 900,
            },
          },
        ),
      );

      final result = await service.refreshToken(_refreshRequest());

      expect(result.isSuccess, isTrue);
      expect(result.value, 'new.access.token');
    });

    test('preserves a 401 returned by the RestClient', () async {
      const unauthorized = AppError(
        statusCode: 401,
        code: AppErrorCode.unauthenticated,
        message: 'invalid refresh token',
      );
      client.postResult = const Failure(unauthorized);

      final result = await service.refreshToken(_refreshRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, unauthorized), isTrue);
    });

    test('preserves a transient network failure', () async {
      const networkError = AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      client.postResult = const Failure(networkError);

      final result = await service.refreshToken(_refreshRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, networkError), isTrue);
    });

    test('returns parsing failure for an invalid response', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'access_token': '',
              'token_type': 'Bearer',
              'expires_in': 900,
            },
          },
        ),
      );

      final result = await service.refreshToken(_refreshRequest());

      expect(result.isFailure, isTrue);
      expect(result.error!.code, AppErrorCode.parsingError);
    });

    test('rejects a success status different from 200', () async {
      client.postResult = Success(
        const RestClientResponse(statusCode: 204),
      );

      final result = await service.refreshToken(_refreshRequest());

      expect(result.isFailure, isTrue);
      final error = result.error!;
      expect(error.statusCode, 204);
      expect(error.code, AppErrorCode.invalidData);
    });

    test('does not expose the refresh token in logs', () async {
      final messages = <String>[];
      final originalDebugPrint = debugPrint;
      debugPrint = (String? message, {int? wrapWidth}) {
        if (message != null) messages.add(message);
      };
      addTearDown(() => debugPrint = originalDebugPrint);

      client.postResult = Success(
        const RestClientResponse(statusCode: 204, statusMessage: 'No Content'),
      );

      await service.refreshToken(_refreshRequest('SensitiveRefreshToken'));

      final allLogs = messages.join('\n');
      expect(messages, isNotEmpty);
      expect(allLogs.contains('SensitiveRefreshToken'), isFalse);
    });
  });

  group('AuthService.logout', () {
    late _FakeRestClient client;
    late AuthService service;

    setUp(() {
      client = _FakeRestClient();
      service = AuthService(client);
    });

    test('sends the refresh token to /auth/logout', () async {
      client.postResult = Success(
        const RestClientResponse(statusCode: 204),
      );

      await service.logout(_refreshRequest());

      expect(client.lastPostRequest?.path, '/auth/logout');
      expect(client.lastPostRequest?.body, {
        'refresh_token': 'refresh-token',
      });
    });

    test('returns Success<Unit> for 204 with an empty body', () async {
      client.postResult = Success(
        const RestClientResponse(statusCode: 204),
      );

      final result = await service.logout(_refreshRequest());

      expect(result.isSuccess, isTrue);
      expect(result.value, unit);
    });

    test('preserves a remote failure', () async {
      const remoteError = AppError(
        statusCode: 500,
        code: AppErrorCode.httpError,
        message: 'remote logout failed',
      );
      client.postResult = const Failure(remoteError);

      final result = await service.logout(_refreshRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, remoteError), isTrue);
    });

    test('rejects a success status different from 204', () async {
      client.postResult = Success(
        const RestClientResponse(statusCode: 200, data: <String, dynamic>{}),
      );

      final result = await service.logout(_refreshRequest());

      expect(result.isFailure, isTrue);
      final error = result.error!;
      expect(error.statusCode, 200);
      expect(error.code, AppErrorCode.invalidData);
    });

    test('does not expose the refresh token in logs', () async {
      final messages = <String>[];
      final originalDebugPrint = debugPrint;
      debugPrint = (String? message, {int? wrapWidth}) {
        if (message != null) messages.add(message);
      };
      addTearDown(() => debugPrint = originalDebugPrint);

      client.postResult = Success(
        const RestClientResponse(statusCode: 200, statusMessage: 'OK'),
      );

      await service.logout(_refreshRequest('SensitiveRefreshToken'));

      final allLogs = messages.join('\n');
      expect(messages, isNotEmpty);
      expect(allLogs.contains('SensitiveRefreshToken'), isFalse);
    });
  });
}

String _refreshRequest([String token = 'refresh-token']) => token;

LoginCredentials _loginUserRequest() {
  return const LoginCredentials(
    email: 'ada@example.com',
    password: 'secret123',
  );
}

final class _FakeRestClient implements RestClient {
  RestClientRequest? lastPostRequest;
  Result<RestClientResponse>? postResult;

  @override
  AsyncResult<RestClientResponse> post(RestClientRequest request) async {
    lastPostRequest = request;
    final result = postResult;
    if (result == null) {
      throw StateError('postResult was not configured for test');
    }

    return result;
  }

  @override
  AsyncResult<RestClientResponse> delete(RestClientRequest request) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> get(RestClientRequest request) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> patch(RestClientRequest request) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> put(RestClientRequest request) {
    throw UnimplementedError();
  }
}
