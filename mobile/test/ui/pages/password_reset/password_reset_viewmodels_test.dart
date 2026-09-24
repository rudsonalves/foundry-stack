import 'dart:async';

import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/code/viewmodel/password_reset_code_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/email/viewmodel/password_reset_email_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/new_password/viewmodel/password_reset_new_password_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final now = DateTime.utc(2026, 9, 23, 15);
  late _FakePasswordResetRepository repository;
  late PasswordResetUsecase usecase;

  setUp(() {
    repository = _FakePasswordResetRepository(
      challenge: PasswordResetChallenge(
        passwordResetId: 'reset-id',
        email: 'user@example.com',
        codeExpiresAt: now.add(const Duration(minutes: 10)),
        resendAvailableAt: now.add(const Duration(seconds: 30)),
      ),
      proof: PasswordResetProof(
        passwordResetToken: 'proof',
        expiresAt: now.add(const Duration(minutes: 10)),
      ),
    );
    usecase = PasswordResetUsecase.withNow(
      repository: repository,
      now: () => now,
    );
  });

  test('email validates locally and delegates normalized request', () async {
    final viewmodel = PasswordResetEmailViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    viewmodel.updateEmail('invalid');
    expect(viewmodel.isEnabled.value, isFalse);

    viewmodel.updateEmail(' user@example.com ');
    expect(viewmodel.isEnabled.value, isTrue);

    await viewmodel.requestPasswordResetCommand.execute(
      ' user@example.com ',
    );

    expect(repository.requestedEmails, ['user@example.com']);
    expect(viewmodel.step, PasswordResetStep.code);
  });

  test('Command blocks repeated email requests while running', () async {
    final pending = Completer<Result<PasswordResetChallenge>>();
    repository.requestOverride = (_) => pending.future;
    final viewmodel = PasswordResetEmailViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    final first = viewmodel.requestPasswordResetCommand.execute(
      'user@example.com',
    );
    final second = viewmodel.requestPasswordResetCommand.execute(
      'user@example.com',
    );
    await Future<void>.delayed(Duration.zero);

    expect(repository.requestedEmails, ['user@example.com']);

    pending.complete(Success(repository.challenge));
    await Future.wait([first, second]);
  });

  test('code validates input and calculates resend countdown', () async {
    await usecase.requestPasswordReset('user@example.com');
    final viewmodel = PasswordResetCodeViewmodel(usecase, now: () => now);
    addTearDown(viewmodel.dispose);

    viewmodel.updateCode('12345');
    expect(viewmodel.isEnabled.value, isFalse);
    viewmodel.updateCode('012345');
    expect(viewmodel.isEnabled.value, isTrue);
    expect(viewmodel.resendRemaining.value, const Duration(seconds: 30));

    await viewmodel.confirmPasswordResetCommand.execute('012345');

    expect(repository.confirmedCodes, ['012345']);
    expect(viewmodel.step, PasswordResetStep.newPassword);
  });

  test('code delegates resend and return to email', () async {
    repository.challenge = PasswordResetChallenge(
      passwordResetId: 'reset-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 10)),
      resendAvailableAt: now,
    );
    await usecase.requestPasswordReset('user@example.com');
    final viewmodel = PasswordResetCodeViewmodel(usecase, now: () => now);
    addTearDown(viewmodel.dispose);

    await viewmodel.resendPasswordResetCommand.execute();
    expect(repository.requestedEmails, hasLength(2));

    await viewmodel.returnToEmailCommand.execute();
    expect(viewmodel.step, PasswordResetStep.email);
  });

  testWidgets('code updates the countdown and cancels it at zero', (
    tester,
  ) async {
    var currentTime = now;
    await usecase.requestPasswordReset('user@example.com');
    final viewmodel = PasswordResetCodeViewmodel(
      usecase,
      now: () => currentTime,
    );
    addTearDown(viewmodel.dispose);

    expect(viewmodel.resendRemaining.value, const Duration(seconds: 30));

    currentTime = now.add(const Duration(seconds: 29));
    await tester.pump(const Duration(seconds: 1));
    expect(viewmodel.resendRemaining.value, const Duration(seconds: 1));

    currentTime = now.add(const Duration(seconds: 30));
    await tester.pump(const Duration(seconds: 1));
    expect(viewmodel.resendRemaining.value, Duration.zero);
    expect(viewmodel.canResend, isTrue);
  });

  test('new password validates equality and delegates only password', () async {
    repository.challenge = PasswordResetChallenge(
      passwordResetId: 'reset-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 10)),
      resendAvailableAt: now,
    );
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');
    final viewmodel = PasswordResetNewPasswordViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    viewmodel.updatePasswords('Password1', 'different');
    expect(viewmodel.isEnabled.value, isFalse);
    viewmodel.updatePasswords('Password1', 'Password1');
    expect(viewmodel.isEnabled.value, isTrue);

    await viewmodel.completePasswordResetCommand.execute('Password1');

    expect(repository.completedPasswords, ['Password1']);
    expect(viewmodel.step, PasswordResetStep.email);
  });

  test('new password delegates return to code and discards proof', () async {
    repository.challenge = PasswordResetChallenge(
      passwordResetId: 'reset-id',
      email: 'user@example.com',
      codeExpiresAt: now.add(const Duration(minutes: 10)),
      resendAvailableAt: now,
    );
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');
    final viewmodel = PasswordResetNewPasswordViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.returnToCodeCommand.execute();

    expect(viewmodel.step, PasswordResetStep.code);
    expect(usecase.proof, isNull);
  });
}

final class _FakePasswordResetRepository implements PasswordResetRepository {
  _FakePasswordResetRepository({required this.challenge, required this.proof});

  PasswordResetChallenge challenge;
  final PasswordResetProof proof;
  Future<Result<PasswordResetChallenge>> Function(String email)?
  requestOverride;
  final requestedEmails = <String>[];
  final confirmedCodes = <String>[];
  final completedPasswords = <String>[];

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) {
    requestedEmails.add(email);
    return requestOverride?.call(email) ?? Future.value(Success(challenge));
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    confirmedCodes.add(code);
    return Success(proof);
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) async {
    completedPasswords.add(newPassword);
    return const Success(unit);
  }
}
