import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final now = DateTime.utc(2026, 9, 23, 15);

  late _FakePasswordResetRepository repository;
  late PasswordResetUsecase usecase;
  late PasswordResetChallenge challenge;
  late PasswordResetProof proof;

  setUp(() {
    challenge = PasswordResetChallenge(
      passwordResetId: 'reset-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 10)),
      resendAvailableAt: now,
    );

    proof = PasswordResetProof(
      passwordResetToken: 'reset-proof',
      expiresAt: now.add(const Duration(minutes: 10)),
    );

    repository = _FakePasswordResetRepository(
      requestResult: Success(challenge),
      confirmResult: Success(proof),
    );

    usecase = PasswordResetUsecase.withNow(
      repository: repository,
      now: () => now,
    );
  });

  test('starts at email without state', () {
    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, isNull);
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
  });

  test('normalizes email and advances to code after request', () async {
    final result = await usecase.requestPasswordReset(
      '  user@example.com  ',
    );

    expect(result.isSuccess, isTrue);
    expect(repository.requestedEmails, ['user@example.com']);
    expect(usecase.email, 'user@example.com');
    expect(usecase.challenge, same(challenge));
    expect(usecase.step, PasswordResetStep.code);
  });

  test('rejects invalid email without calling repository', () async {
    final result = await usecase.requestPasswordReset('invalid');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(repository.requestedEmails, isEmpty);
    expect(usecase.step, PasswordResetStep.email);
  });

  test('preserves leading zero and advances after confirmation', () async {
    await usecase.requestPasswordReset('user@example.com');

    final result = await usecase.confirmPasswordReset(' 012345 ');

    expect(result.isSuccess, isTrue);
    expect(repository.confirmations, [
      (passwordResetId: 'reset-id', code: '012345'),
    ]);
    expect(usecase.proof, same(proof));
    expect(usecase.step, PasswordResetStep.newPassword);
  });

  test('rejects malformed codes without calling repository', () async {
    await usecase.requestPasswordReset('user@example.com');

    for (final code in ['12345', '1234567', '12a456']) {
      final result = await usecase.confirmPasswordReset(code);
      expect(result.error?.code, AppErrorCode.invalidData);
    }

    expect(repository.confirmations, isEmpty);
    expect(usecase.step, PasswordResetStep.code);
  });

  test('rejects an expired code without calling repository', () async {
    final expiredChallenge = PasswordResetChallenge(
      passwordResetId: 'expired-reset-id',
      email: 'user@example.com',
      codeExpiresAt: now,
      resendAvailableAt: now,
    );
    repository.requestResult = Success(expiredChallenge);

    await usecase.requestPasswordReset('user@example.com');
    final result = await usecase.confirmPasswordReset('123456');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(repository.confirmations, isEmpty);
    expect(usecase.step, PasswordResetStep.code);
    expect(usecase.challenge, same(expiredChallenge));
  });

  test('rejects resend before resend availability', () async {
    final coolingDownChallenge = PasswordResetChallenge(
      passwordResetId: 'reset-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 10)),
      resendAvailableAt: now.add(const Duration(seconds: 1)),
    );
    repository.requestResult = Success(coolingDownChallenge);

    await usecase.requestPasswordReset('user@example.com');
    repository.requestedEmails.clear();

    final result = await usecase.resendPasswordReset();

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(repository.requestedEmails, isEmpty);
    expect(usecase.challenge, same(coolingDownChallenge));
  });

  test('replaces challenge after successful resend', () async {
    await usecase.requestPasswordReset('user@example.com');

    final replacement = PasswordResetChallenge(
      passwordResetId: 'replacement-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 20)),
      resendAvailableAt: now.add(const Duration(minutes: 1)),
    );
    repository.requestResult = Success(replacement);
    repository.requestedEmails.clear();

    final result = await usecase.resendPasswordReset();

    expect(result.isSuccess, isTrue);
    expect(repository.requestedEmails, ['user@example.com']);
    expect(usecase.challenge, same(replacement));
    expect(usecase.step, PasswordResetStep.code);
  });

  test('keeps current challenge when resend fails', () async {
    const error = AppError(
      code: AppErrorCode.networkError,
      message: 'offline',
    );

    await usecase.requestPasswordReset('user@example.com');
    repository.requestResult = const Failure(error);
    repository.requestedEmails.clear();

    final result = await usecase.resendPasswordReset();

    expect(result.error, same(error));
    expect(repository.requestedEmails, ['user@example.com']);
    expect(usecase.challenge, same(challenge));
    expect(usecase.step, PasswordResetStep.code);
  });

  test('rejects completion with expired proof', () async {
    final expiredProof = PasswordResetProof(
      passwordResetToken: 'expired-proof',
      expiresAt: now,
    );
    repository.confirmResult = Success(expiredProof);

    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    final result = await usecase.completePasswordReset('Password123');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(repository.completions, isEmpty);
    expect(usecase.step, PasswordResetStep.newPassword);
    expect(usecase.proof, same(expiredProof));
  });

  test('completes with id and proof and clears all state', () async {
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    final result = await usecase.completePasswordReset('Password123');

    expect(result.isSuccess, isTrue);
    expect(repository.completions, [
      (
        passwordResetId: 'reset-id',
        passwordResetToken: 'reset-proof',
        newPassword: 'Password123',
      ),
    ]);
    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, isNull);
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
  });

  test('keeps state after recoverable completion failure', () async {
    const error = AppError(
      code: AppErrorCode.networkError,
      message: 'offline',
    );
    repository.completeResult = const Failure(error);

    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    final result = await usecase.completePasswordReset('Password123');

    expect(result.error, same(error));
    expect(usecase.step, PasswordResetStep.newPassword);
    expect(usecase.email, 'user@example.com');
    expect(usecase.challenge, same(challenge));
    expect(usecase.proof, same(proof));
  });

  test('restarts at email after INVALID_PASSWORD_RESET', () async {
    const error = AppError(
      code: AppErrorCode.invalidData,
      details: {'code': BackendErrorCodes.invalidPasswordReset},
      message: '...',
    );
    repository.completeResult = const Failure(error);

    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    final result = await usecase.completePasswordReset('Password123');

    expect(result.error, same(error));
    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, isNull);
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
  });

  test('returns from new password to code discarding proof', () async {
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    final result = await usecase.returnToCode();

    expect(result.isSuccess, isTrue);
    expect(usecase.step, PasswordResetStep.code);
    expect(usecase.email, 'user@example.com');
    expect(usecase.challenge, same(challenge));
    expect(usecase.proof, isNull);
  });

  test('returns from code to email discarding challenge', () async {
    await usecase.requestPasswordReset('user@example.com');

    final result = await usecase.returnToEmail();

    expect(result.isSuccess, isTrue);
    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, 'user@example.com');
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
  });

  test('dispose clears every ephemeral reference', () async {
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');

    usecase.dispose();

    expect(usecase.step, PasswordResetStep.email);
    expect(usecase.email, isNull);
    expect(usecase.challenge, isNull);
    expect(usecase.proof, isNull);
  });
}

class _FakePasswordResetRepository implements PasswordResetRepository {
  _FakePasswordResetRepository({
    required this.requestResult,
    required this.confirmResult,
  });

  Result<PasswordResetChallenge> requestResult;
  Result<PasswordResetProof> confirmResult;
  Result<Unit> completeResult = const Success(unit);

  final requestedEmails = <String>[];
  final confirmations = <({String passwordResetId, String code})>[];
  final completions =
      <
        ({
          String passwordResetId,
          String passwordResetToken,
          String newPassword,
        })
      >[];

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(
    String email,
  ) async {
    requestedEmails.add(email);
    return requestResult;
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    confirmations.add((
      passwordResetId: passwordResetId,
      code: code,
    ));
    return confirmResult;
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) async {
    completions.add((
      passwordResetId: passwordResetId,
      passwordResetToken: passwordResetToken,
      newPassword: newPassword,
    ));
    return completeResult;
  }
}
