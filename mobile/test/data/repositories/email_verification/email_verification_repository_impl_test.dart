import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/email_verification/email_verification_repository_impl.dart';
import 'package:foundry_stack_mobile/data/services/apis/email_verification/email_verification_service.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_proof.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeEmailVerificationService service;
  late EmailVerificationRepositoryImpl repository;

  setUp(() {
    service = _FakeEmailVerificationService();
    repository = EmailVerificationRepositoryImpl(service: service);
  });

  group('EmailVerificationRepositoryImpl.requestVerification', () {
    for (final errorCase in [
      (
        code: BackendErrorCodes.emailAlreadyRegistered,
        statusCode: 409,
        appErrorCode: AppErrorCode.conflict,
      ),
      (
        code: BackendErrorCodes.emailDeliveryUnavailable,
        statusCode: 503,
        appErrorCode: AppErrorCode.httpError,
      ),
      (
        code: BackendErrorCodes.rateLimitExceeded,
        statusCode: 429,
        appErrorCode: AppErrorCode.httpError,
      ),
    ]) {
      test('preserves ${errorCase.code}', () async {
        final error = AppError(
          statusCode: errorCase.statusCode,
          code: errorCase.appErrorCode,
          message: 'request rejected',
          details: {'code': errorCase.code},
        );
        service.requestResult = Failure(error);

        final result = await repository.requestVerification(
          'ada@example.com',
        );

        expect(identical(result.error, error), isTrue);
        expect(backendErrorCode(result.error), errorCase.code);
      });
    }

    test('returns the challenge and forwards the email', () async {
      final challenge = EmailVerificationChallenge(
        verificationId: 'verification-id',
        codeExpiresAt: DateTime.utc(2026, 9, 18, 15, 15),
        resendAvailableAt: DateTime.utc(2026, 9, 18, 15, 1),
      );
      service.requestResult = Success(challenge);

      final result = await repository.requestVerification('ada@example.com');

      expect(result.isSuccess, isTrue);
      expect(identical(result.value, challenge), isTrue);
      expect(service.requestedEmail, 'ada@example.com');
    });

    test('preserves a service failure', () async {
      const error = AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      service.requestResult = const Failure(error);

      final result = await repository.requestVerification('ada@example.com');

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });

  group('EmailVerificationRepositoryImpl.confirmVerification', () {
    test('returns the proof and forwards verification id and code', () async {
      final proof = EmailVerificationProof(
        token: 'opaque-secret',
        expiresAt: DateTime.utc(2026, 9, 19, 15),
      );
      service.confirmationResult = Success(proof);

      final result = await repository.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(result.isSuccess, isTrue);
      expect(identical(result.value, proof), isTrue);
      expect(service.confirmedVerificationId, 'verification-id');
      expect(service.confirmedCode, '012345');
    });

    test('preserves a service failure', () async {
      const error = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid verification',
      );
      service.confirmationResult = const Failure(error);

      final result = await repository.confirmVerification(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });
}

class _FakeEmailVerificationService implements EmailVerificationService {
  Result<EmailVerificationChallenge>? requestResult;
  Result<EmailVerificationProof>? confirmationResult;

  String? requestedEmail;
  String? confirmedVerificationId;
  String? confirmedCode;

  @override
  AsyncResult<EmailVerificationChallenge> requestVerification(
    String email,
  ) async {
    requestedEmail = email;
    return requestResult!;
  }

  @override
  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  }) async {
    confirmedVerificationId = verificationId;
    confirmedCode = code;
    return confirmationResult!;
  }
}
