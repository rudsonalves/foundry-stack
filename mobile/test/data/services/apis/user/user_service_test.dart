import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/user_service.dart';
import 'package:foundry_stack_mobile/domain/common/users/models/user_registration.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('UserService.createUser', () {
    late _FakeRestClient client;
    late UserService service;

    setUp(() {
      client = _FakeRestClient();
      service = UserService(client);
    });

    test('sends request to /users', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          data: {
            'data': {
              'id': 'user-1',
              'name': 'Ada',
              'email': 'ada@example.com',
            },
          },
        ),
      );

      await service.createUser(_createUserRequest());

      expect(client.lastPostRequest, isNotNull);
      expect(client.lastPostRequest!.path, '/users');
    });

    test('translates user registration to the API body', () async {
      final request = _createUserRequest();
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          data: {
            'data': {
              'id': 'user-1',
              'name': 'Ada',
              'email': 'ada@example.com',
            },
          },
        ),
      );

      await service.createUser(request);

      expect(client.lastPostRequest, isNotNull);
      expect(client.lastPostRequest!.body, {
        'name': request.name,
        'email': request.email,
        'email_verification_token': request.emailVerificationToken,
        'password': request.password,
      });
    });

    test('returns success for 201 response', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          data: {
            'data': {
              'id': 'user-1',
              'name': 'Ada',
              'email': 'ada@example.com',
            },
          },
        ),
      );

      final result = await service.createUser(_createUserRequest());

      expect(result.isSuccess, isTrue);
      expect(result.value?.id, 'user-1');
    });

    test('extracts payload from data envelope', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          data: {
            'data': {
              'id': 'user-99',
              'name': 'Ada Lovelace',
              'email': 'ada@example.com',
            },
            'meta': {'traceId': 'trace-1'},
          },
        ),
      );

      final result = await service.createUser(_createUserRequest());

      expect(result.isSuccess, isTrue);
      expect(result.value?.id, 'user-99');
      expect(result.value?.name, 'Ada Lovelace');
      expect(result.value?.email, 'ada@example.com');
    });

    test('returns failure when status is different from 201', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 200,
          statusMessage: 'OK',
          data: {
            'data': {
              'id': 'user-1',
              'name': 'Ada',
              'email': 'ada@example.com',
            },
          },
        ),
      );

      final result = await service.createUser(_createUserRequest());

      expect(result.isFailure, isTrue);
      expect(result.error, isA<AppError>());
      final error = result.error!;
      expect(error.statusCode, 200);
      expect(error.code, AppErrorCode.invalidData);
      expect(
        error.message,
        'Unexpected response status: expected 201, received 200',
      );
    });

    test('preserves AppError for 400 and 409', () async {
      final badRequest = const AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid body',
      );

      client.postResult = Failure(badRequest);
      final badRequestResult = await service.createUser(_createUserRequest());

      expect(badRequestResult.isFailure, isTrue);
      expect(identical(badRequestResult.error, badRequest), isTrue);

      final conflict = const AppError(
        statusCode: 409,
        code: AppErrorCode.conflict,
        message: 'email already exists',
      );

      client.postResult = Failure(conflict);
      final conflictResult = await service.createUser(_createUserRequest());

      expect(conflictResult.isFailure, isTrue);
      expect(identical(conflictResult.error, conflict), isTrue);
    });

    test('returns failure on network error', () async {
      final networkError = const AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      client.postResult = Failure(networkError);

      final result = await service.createUser(_createUserRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, networkError), isTrue);
    });

    test('returns failure on malformed envelope', () async {
      client.postResult = Success(
        const RestClientResponse(
          statusCode: 201,
          data: {
            'data': 'invalid-type',
          },
        ),
      );

      final result = await service.createUser(_createUserRequest());

      expect(result.isFailure, isTrue);
      expect(result.error, isA<AppError>());
      final error = result.error!;
      expect(error.code, AppErrorCode.parsingError);
    });

    test(
      'returns parsing failure for error envelope with status 201',
      () async {
        client.postResult = Success(
          const RestClientResponse(
            statusCode: 201,
            data: {
              'error': {
                'code': 'INTERNAL_ERROR',
                'message': 'unexpected error',
              },
            },
          ),
        );

        final result = await service.createUser(_createUserRequest());

        expect(result.isFailure, isTrue);
        expect(result.error, isA<AppError>());
        final error = result.error!;
        expect(error.code, AppErrorCode.parsingError);
        expect(error.message, 'Response body does not contain data');
      },
    );

    test('does not expose password or verification proof in logs', () async {
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

      const failure = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid body',
      );
      client.postResult = const Failure(failure);

      const request = UserRegistration(
        name: 'Ada',
        email: 'ada@example.com',
        password: 'S3nh@SuperSecreta',
        emailVerificationToken: 'verification-proof',
      );

      await service.createUser(request);

      final allLogs = messages.join('\n');
      expect(allLogs.contains('S3nh@SuperSecreta'), isFalse);
      expect(allLogs.contains('verification-proof'), isFalse);
    });
  });
}

UserRegistration _createUserRequest() {
  return const UserRegistration(
    name: 'Ada',
    email: 'ada@example.com',
    password: 'secret123',
    emailVerificationToken: 'verification-proof',
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
