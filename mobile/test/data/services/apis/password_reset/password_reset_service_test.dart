import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/data/services/apis/password_reset/password_reset_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeRestClient client;
  late PasswordResetService service;

  setUp(() {
    client = _FakeRestClient();
    service = PasswordResetService(client);
  });

  group('PasswordResetService.requestPasswordReset', () {
    test('sends the email to /password-resets', () async {
      client.postResult = const Success(_challengeResponse);

      await service.requestPasswordReset('user@example.com');

      expect(client.lastPostRequest?.path, '/password-resets');
      expect(client.lastPostRequest?.body, {
        'email': 'user@example.com',
      });
    });

    test('returns the challenge for a 202 response', () async {
      client.postResult = const Success(_challengeResponse);

      final result = await service.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isSuccess, isTrue);
      expect(result.value?.passwordResetId, 'password-reset-id');
      expect(result.value?.email, 'user@example.com');
      expect(
        result.value?.codeExpiresAt,
        DateTime.parse('2026-09-23T15:15:00Z'),
      );
      expect(
        result.value?.resendAvailableAt,
        DateTime.parse('2026-09-23T15:01:00Z'),
      );
    });

    test('rejects a success status different from 202', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 200,
          data: _challengeEnvelope,
        ),
      );

      final result = await service.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.statusCode, 200);
      expect(result.error?.code, AppErrorCode.invalidData);
      expect(
        result.error?.message,
        'Unexpected response status: expected 202, received 200',
      );
    });

    test('preserves an AppError returned by the client', () async {
      const error = AppError(
        statusCode: 429,
        code: AppErrorCode.httpError,
        message: 'rate limit exceeded',
        details: {
          'code': 'RATE_LIMIT_EXCEEDED',
        },
      );
      client.postResult = const Failure(error);

      final result = await service.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });

    test('returns parsing failure for incomplete data', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 202,
          data: {
            'data': {
              'password_reset_id': 'password-reset-id',
            },
          },
        ),
      );

      final result = await service.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.code, AppErrorCode.parsingError);
    });
  });

  group('PasswordResetService.confirmPasswordReset', () {
    test('sends id and code to the confirmation path', () async {
      client.postResult = const Success(_confirmationResponse);

      await service.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(
        client.lastPostRequest?.path,
        '/password-resets/confirm',
      );
      expect(client.lastPostRequest?.body, {
        'password_reset_id': 'password-reset-id',
        'code': '012345',
      });
    });

    test('returns the proof for a 200 response', () async {
      client.postResult = const Success(_confirmationResponse);

      final result = await service.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isSuccess, isTrue);
      expect(
        result.value?.passwordResetToken,
        'opaque-password-reset-token',
      );
      expect(
        result.value?.expiresAt,
        DateTime.parse('2026-09-23T16:00:00Z'),
      );
    });

    test('rejects a success status different from 200', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 202,
          data: _confirmationEnvelope,
        ),
      );

      final result = await service.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.statusCode, 202);
      expect(result.error?.code, AppErrorCode.invalidData);
      expect(
        result.error?.message,
        'Unexpected response status: expected 200, received 202',
      );
    });

    test('preserves an AppError returned by the client', () async {
      const error = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid password reset',
        details: {
          'code': 'INVALID_PASSWORD_RESET',
        },
      );
      client.postResult = const Failure(error);

      final result = await service.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });

    test('returns parsing failure for incomplete data', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 200,
          data: {
            'data': {
              'password_reset_token': 'opaque-password-reset-token',
            },
          },
        ),
      );

      final result = await service.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.code, AppErrorCode.parsingError);
    });
  });

  group('PasswordResetService.completePasswordReset', () {
    test('sends id, proof and password to the completion path', () async {
      client.postResult = const Success(
        RestClientResponse(statusCode: 204),
      );

      await service.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-password-reset-token',
        newPassword: 'new-password123',
      );

      expect(
        client.lastPostRequest?.path,
        '/password-resets/complete',
      );
      expect(client.lastPostRequest?.body, {
        'password_reset_id': 'password-reset-id',
        'password_reset_token': 'opaque-password-reset-token',
        'new_password': 'new-password123',
      });
    });

    test('returns Unit for a 204 response without parsing body', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 204,
          data: 'body-must-be-ignored',
        ),
      );

      final result = await service.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-password-reset-token',
        newPassword: 'new-password123',
      );

      expect(result.isSuccess, isTrue);
      expect(result.value, unit);
    });

    test('rejects a success status different from 204', () async {
      client.postResult = const Success(
        RestClientResponse(statusCode: 200),
      );

      final result = await service.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-password-reset-token',
        newPassword: 'new-password123',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.statusCode, 200);
      expect(result.error?.code, AppErrorCode.invalidData);
      expect(
        result.error?.message,
        'Unexpected response status: expected 204, received 200',
      );
    });

    test('preserves an AppError returned by the client', () async {
      const error = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid password reset',
        details: {
          'code': 'INVALID_PASSWORD_RESET',
        },
      );
      client.postResult = const Failure(error);

      final result = await service.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-password-reset-token',
        newPassword: 'new-password123',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });

  test('does not expose sensitive recovery values in logs', () async {
    final messages = <String>[];
    final originalDebugPrint = debugPrint;

    debugPrint = (String? message, {int? wrapWidth}) {
      if (message != null) {
        messages.add(message);
      }
    };
    addTearDown(() => debugPrint = originalDebugPrint);

    client.postResult = const Success(_challengeResponse);
    await service.requestPasswordReset(
      'sensitive-email@example.com',
    );

    client.postResult = const Success(_confirmationResponse);
    await service.confirmPasswordReset(
      passwordResetId: 'password-reset-id',
      code: '012345',
    );

    client.postResult = const Success(
      RestClientResponse(statusCode: 204),
    );
    await service.completePasswordReset(
      passwordResetId: 'password-reset-id',
      passwordResetToken: 'opaque-password-reset-token',
      newPassword: 'sensitive-password123',
    );

    final logs = messages.join('\n');

    expect(
      logs,
      isNot(contains('sensitive-email@example.com')),
    );
    expect(logs, isNot(contains('012345')));
    expect(
      logs,
      isNot(contains('opaque-password-reset-token')),
    );
    expect(
      logs,
      isNot(contains('sensitive-password123')),
    );
  });
}

const _challengeEnvelope = {
  'data': {
    'password_reset_id': 'password-reset-id',
    'code_expires_at': '2026-09-23T15:15:00Z',
    'resend_available_at': '2026-09-23T15:01:00Z',
  },
};

const _challengeResponse = RestClientResponse(
  statusCode: 202,
  data: _challengeEnvelope,
);

const _confirmationEnvelope = {
  'data': {
    'password_reset_token': 'opaque-password-reset-token',
    'expires_at': '2026-09-23T16:00:00Z',
  },
};

const _confirmationResponse = RestClientResponse(
  statusCode: 200,
  data: _confirmationEnvelope,
);

final class _FakeRestClient implements RestClient {
  RestClientRequest? lastPostRequest;
  Result<RestClientResponse>? postResult;

  @override
  AsyncResult<RestClientResponse> post(
    RestClientRequest request,
  ) async {
    lastPostRequest = request;

    final result = postResult;
    if (result == null) {
      throw StateError(
        'postResult was not configured for test',
      );
    }

    return result;
  }

  @override
  AsyncResult<RestClientResponse> delete(
    RestClientRequest request,
  ) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> get(
    RestClientRequest request,
  ) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> patch(
    RestClientRequest request,
  ) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<RestClientResponse> put(
    RestClientRequest request,
  ) {
    throw UnimplementedError();
  }
}
