import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/data/services/apis/email_verification/email_verification_service.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeRestClient client;
  late EmailVerificationService service;

  setUp(() {
    client = _FakeRestClient();
    service = EmailVerificationService(client);
  });

  group('EmailVerificationService.requestVerification', () {
    test('sends the email to /email-verifications', () async {
      client.postResult = const Success(_challengeResponse);

      await service.requestVerification('ada@example.com');

      expect(client.lastPostRequest?.path, '/email-verifications');
      expect(client.lastPostRequest?.body, {'email': 'ada@example.com'});
    });

    test('returns the application model for a 202 response', () async {
      client.postResult = const Success(_challengeResponse);

      final result = await service.requestVerification('ada@example.com');

      expect(result.isSuccess, isTrue);
      expect(result.value?.verificationId, 'verification-id');
      expect(
        result.value?.codeExpiresAt,
        DateTime.parse('2026-09-18T15:15:00Z'),
      );
      expect(
        result.value?.resendAvailableAt,
        DateTime.parse('2026-09-18T15:01:00Z'),
      );
    });

    test('rejects a success status different from 202', () async {
      client.postResult = const Success(
        RestClientResponse(statusCode: 200, data: _challengeEnvelope),
      );

      final result = await service.requestVerification('ada@example.com');

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
          'code': BackendErrorCodes.rateLimitExceeded,
          'message': 'rate limit exceeded',
        },
      );
      client.postResult = const Failure(error);

      final result = await service.requestVerification('ada@example.com');

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(
        backendErrorCode(result.error),
        BackendErrorCodes.rateLimitExceeded,
      );
    });

    test('returns parsing failure for an incomplete response', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 202,
          data: {
            'data': {'verification_id': 'verification-id'},
          },
        ),
      );

      final result = await service.requestVerification('ada@example.com');

      expect(result.isFailure, isTrue);
      expect(result.error?.code, AppErrorCode.parsingError);
    });
  });

  group('EmailVerificationService.confirmVerification', () {
    test('sends verification id and code to the confirmation path', () async {
      client.postResult = const Success(_confirmationResponse);

      await service.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(
        client.lastPostRequest?.path,
        '/email-verifications/confirm',
      );
      expect(client.lastPostRequest?.body, {
        'verification_id': 'verification-id',
        'code': '012345',
      });
    });

    test('returns the application model for a 200 response', () async {
      client.postResult = const Success(_confirmationResponse);

      final result = await service.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(result.isSuccess, isTrue);
      expect(result.value?.token, 'opaque-secret');
      expect(
        result.value?.expiresAt,
        DateTime.parse('2026-09-19T15:00:00Z'),
      );
    });

    test('rejects a success status different from 200', () async {
      client.postResult = const Success(
        RestClientResponse(statusCode: 202, data: _confirmationEnvelope),
      );

      final result = await service.confirmVerification(
        verificationId: 'verification-id',
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
        message: 'invalid verification',
        details: {
          'code': BackendErrorCodes.invalidEmailVerification,
          'message': 'invalid verification',
        },
      );
      client.postResult = const Failure(error);

      final result = await service.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(
        backendErrorCode(result.error),
        BackendErrorCodes.invalidEmailVerification,
      );
    });

    test('returns parsing failure for an incomplete response', () async {
      client.postResult = const Success(
        RestClientResponse(
          statusCode: 200,
          data: {
            'data': {'email_verification_token': 'opaque-secret'},
          },
        ),
      );

      final result = await service.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(result.error?.code, AppErrorCode.parsingError);
    });

    test('does not expose code or proof in logs', () async {
      final messages = <String>[];
      final originalDebugPrint = debugPrint;
      debugPrint = (String? message, {int? wrapWidth}) {
        if (message != null) messages.add(message);
      };
      addTearDown(() => debugPrint = originalDebugPrint);
      client.postResult = const Success(_confirmationResponse);

      await service.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      final logs = messages.join('\n');
      expect(logs, isNot(contains('012345')));
      expect(logs, isNot(contains('opaque-secret')));
    });
  });
}

const _challengeEnvelope = {
  'data': {
    'verification_id': 'verification-id',
    'code_expires_at': '2026-09-18T15:15:00Z',
    'resend_available_at': '2026-09-18T15:01:00Z',
  },
};

const _challengeResponse = RestClientResponse(
  statusCode: 202,
  data: _challengeEnvelope,
);

const _confirmationEnvelope = {
  'data': {
    'email_verification_token': 'opaque-secret',
    'expires_at': '2026-09-19T15:00:00Z',
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
