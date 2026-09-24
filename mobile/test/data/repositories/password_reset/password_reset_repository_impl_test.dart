import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository_impl.dart';
import 'package:foundry_stack_mobile/data/services/apis/password_reset/password_reset_service.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakePasswordResetService service;
  late PasswordResetRepositoryImpl repository;

  setUp(() {
    service = _FakePasswordResetService();
    repository = PasswordResetRepositoryImpl(
      service: service,
    );
  });

  group('PasswordResetRepositoryImpl.requestPasswordReset', () {
    test('returns challenge and forwards email', () async {
      final challenge = PasswordResetChallenge(
        passwordResetId: 'password-reset-id',
        email: 'user@example.com',
        codeExpiresAt: DateTime.utc(
          2026,
          9,
          23,
          15,
          15,
        ),
        resendAvailableAt: DateTime.utc(
          2026,
          9,
          23,
          15,
          1,
        ),
      );
      service.requestResult = Success(challenge);

      final result = await repository.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isSuccess, isTrue);
      expect(identical(result.value, challenge), isTrue);
      expect(service.requestedEmail, 'user@example.com');
    });

    test('preserves service failure', () async {
      const error = AppError(
        statusCode: 429,
        code: AppErrorCode.httpError,
        message: 'rate limit exceeded',
        details: {
          'code': 'RATE_LIMIT_EXCEEDED',
        },
      );
      service.requestResult = const Failure(error);

      final result = await repository.requestPasswordReset(
        'user@example.com',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(result.error?.details, same(error.details));
    });
  });

  group('PasswordResetRepositoryImpl.confirmPasswordReset', () {
    test('returns proof and forwards id and code', () async {
      final proof = PasswordResetProof(
        passwordResetToken: 'opaque-proof',
        expiresAt: DateTime.utc(2026, 9, 23, 16),
      );
      service.confirmResult = Success(proof);

      final result = await repository.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isSuccess, isTrue);
      expect(identical(result.value, proof), isTrue);
      expect(
        service.confirmedPasswordResetId,
        'password-reset-id',
      );
      expect(service.confirmedCode, '012345');
    });

    test('preserves service failure', () async {
      const error = AppError(
        statusCode: 400,
        code: AppErrorCode.invalidData,
        message: 'invalid password reset',
        details: {
          'code': 'INVALID_PASSWORD_RESET',
        },
      );
      service.confirmResult = const Failure(error);

      final result = await repository.confirmPasswordReset(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(result.error?.details, same(error.details));
    });
  });

  group('PasswordResetRepositoryImpl.completePasswordReset', () {
    test('returns Unit and forwards id, proof and password', () async {
      service.completeResult = const Success(unit);

      final result = await repository.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-proof',
        newPassword: 'new-password123',
      );

      expect(result.isSuccess, isTrue);
      expect(result.value, unit);
      expect(
        service.completedPasswordResetId,
        'password-reset-id',
      );
      expect(
        service.completedPasswordResetToken,
        'opaque-proof',
      );
      expect(
        service.completedNewPassword,
        'new-password123',
      );
    });

    test('preserves service failure', () async {
      const error = AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      service.completeResult = const Failure(error);

      final result = await repository.completePasswordReset(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-proof',
        newPassword: 'new-password123',
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });
}

class _FakePasswordResetService implements PasswordResetService {
  Result<PasswordResetChallenge>? requestResult;
  Result<PasswordResetProof>? confirmResult;
  Result<Unit>? completeResult;

  String? requestedEmail;

  String? confirmedPasswordResetId;
  String? confirmedCode;

  String? completedPasswordResetId;
  String? completedPasswordResetToken;
  String? completedNewPassword;

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(
    String email,
  ) async {
    requestedEmail = email;
    return requestResult!;
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    confirmedPasswordResetId = passwordResetId;
    confirmedCode = code;
    return confirmResult!;
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) async {
    completedPasswordResetId = passwordResetId;
    completedPasswordResetToken = passwordResetToken;
    completedNewPassword = newPassword;
    return completeResult!;
  }
}
