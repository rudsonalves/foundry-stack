import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/data/repositories/password_reset/password_reset_repository.dart';
import '../../common/password_reset/models/password_reset_challenge.dart';
import '../../common/password_reset/models/password_reset_proof.dart';
import '../../common/password_reset/models/password_reset_step.dart';

class PasswordResetUsecase {
  final PasswordResetRepository _repository;
  final DateTime Function() _now;

  PasswordResetUsecase({
    required PasswordResetRepository repository,
  }) : _repository = repository,
       _now = DateTime.now;

  PasswordResetUsecase.withNow({
    required PasswordResetRepository repository,
    required DateTime Function() now,
  }) : _repository = repository,
       _now = now;

  PasswordResetStep _step = .email;
  String? _email;
  PasswordResetChallenge? _challenge;
  PasswordResetProof? _proof;

  PasswordResetStep get step => _step;
  String? get email => _email;
  PasswordResetChallenge? get challenge => _challenge;
  PasswordResetProof? get proof => _proof;

  AsyncResult<Unit> requestPasswordReset(String email) async {
    if (_step != .email) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Invalid password reset step for email.',
        ),
      );
    }

    final normalizedEmail = email.trim();

    if (!normalizedEmail.isValidEmail) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Invalid email.',
        ),
      );
    }

    final result = await _repository.requestPasswordReset(normalizedEmail);

    if (result.isFailure) {
      return Failure(result.error!);
    }

    _email = normalizedEmail;
    _challenge = result.value!;
    _step = .code;

    return Success(unit);
  }

  AsyncResult<Unit> confirmPasswordReset(String code) async {
    final challenge = _challenge;

    if (_step != .code || challenge == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Invalid password reset step for code confirmation.',
        ),
      );
    }

    final normalizedCode = code.trim();

    if (!RegExp(r'^\d{6}$').hasMatch(normalizedCode)) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Code must contain exactly six digits.',
        ),
      );
    }

    if (challenge.isCodeExpiredAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset code has expired.',
        ),
      );
    }

    final result = await _repository.confirmPasswordReset(
      passwordResetId: challenge.passwordResetId,
      code: normalizedCode,
    );

    if (result.isFailure) {
      return Failure(result.error!);
    }

    _proof = result.value!;
    _step = .newPassword;

    return Success(unit);
  }

  AsyncResult<Unit> completePasswordReset(String newPassword) async {
    final challenge = _challenge;
    final proof = _proof;

    if (_step != .newPassword || challenge == null || proof == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset is not ready for completion.',
        ),
      );
    }

    if (proof.isExpiredAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset proof has expired.',
        ),
      );
    }

    if (!newPassword.isValidPassword) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password does not meet the current policy.',
        ),
      );
    }

    final result = await _repository.completePasswordReset(
      passwordResetId: challenge.passwordResetId,
      passwordResetToken: proof.passwordResetToken,
      newPassword: newPassword,
    );

    if (result.isFailure) {
      final error = result.error!;

      if (backendErrorCode(error) == BackendErrorCodes.invalidPasswordReset) {
        _step = .email;
        _email = null;
        _challenge = null;
        _proof = null;
      }

      return Failure(error);
    }

    _step = .email;
    _email = null;
    _challenge = null;
    _proof = null;

    return Success(unit);
  }

  AsyncResult<Unit> resendPasswordReset() async {
    final email = _email;
    final challenge = _challenge;

    if (_step != .code || email == null || challenge == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset is not ready for resend.',
        ),
      );
    }

    if (!challenge.canResendAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset resend is not available yet.',
        ),
      );
    }

    final result = await _repository.requestPasswordReset(email);

    if (result.isFailure) {
      return Failure(result.error!);
    }

    _challenge = result.value!;

    return Success(unit);
  }

  AsyncResult<Unit> returnToCode() async {
    if (_step != .newPassword || _challenge == null || _proof == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset is not in the new password step.',
        ),
      );
    }

    _proof = null;
    _step = .code;

    return Success(unit);
  }

  AsyncResult<Unit> returnToEmail() async {
    if (_step != .code || _email == null || _challenge == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Password reset is not in the code step.',
        ),
      );
    }

    _challenge = null;
    _proof = null;
    _step = .email;

    return Success(unit);
  }

  void dispose() {
    _step = .email;
    _email = null;
    _challenge = null;
    _proof = null;
  }
}
