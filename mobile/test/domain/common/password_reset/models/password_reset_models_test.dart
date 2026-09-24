import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_step.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('PasswordResetChallenge', () {
    final codeExpiresAt = DateTime.utc(2026, 9, 23, 15, 15);
    final resendAvailableAt = DateTime.utc(2026, 9, 23, 15, 1);

    final challenge = PasswordResetChallenge(
      passwordResetId: 'password-reset-id',
      email: 'user@example.com',
      codeExpiresAt: codeExpiresAt,
      resendAvailableAt: resendAvailableAt,
    );

    test('preserves the challenge data', () {
      expect(challenge.passwordResetId, 'password-reset-id');
      expect(challenge.email, 'user@example.com');
      expect(challenge.codeExpiresAt, same(codeExpiresAt));
      expect(challenge.resendAvailableAt, same(resendAvailableAt));
    });

    test('detects code expiration including the boundary', () {
      expect(
        challenge.isCodeExpiredAt(
          DateTime.utc(2026, 9, 23, 15, 14, 59),
        ),
        isFalse,
      );
      expect(
        challenge.isCodeExpiredAt(
          DateTime.utc(2026, 9, 23, 15, 15),
        ),
        isTrue,
      );
      expect(
        challenge.isCodeExpiredAt(
          DateTime.utc(2026, 9, 23, 15, 16),
        ),
        isTrue,
      );
    });

    test('allows resend beginning at the available instant', () {
      expect(
        challenge.canResendAt(
          DateTime.utc(2026, 9, 23, 15, 0, 59),
        ),
        isFalse,
      );
      expect(
        challenge.canResendAt(
          DateTime.utc(2026, 9, 23, 15, 1),
        ),
        isTrue,
      );
    });

    test('calculates resend remaining time without negative values', () {
      expect(
        challenge.resendRemainingAt(
          DateTime.utc(2026, 9, 23, 15),
        ),
        const Duration(minutes: 1),
      );
      expect(
        challenge.resendRemainingAt(
          DateTime.utc(2026, 9, 23, 15, 1),
        ),
        Duration.zero,
      );
      expect(
        challenge.resendRemainingAt(
          DateTime.utc(2026, 9, 23, 15, 2),
        ),
        Duration.zero,
      );
    });
  });

  group('PasswordResetProof', () {
    final expiresAt = DateTime.utc(2026, 9, 23, 16);

    final proof = PasswordResetProof(
      passwordResetToken: 'opaque-password-reset-token',
      expiresAt: expiresAt,
    );

    test('preserves the proof data', () {
      expect(
        proof.passwordResetToken,
        'opaque-password-reset-token',
      );
      expect(proof.expiresAt, same(expiresAt));
    });

    test('detects expiration including the boundary', () {
      expect(
        proof.isExpiredAt(
          DateTime.utc(2026, 9, 23, 15, 59, 59),
        ),
        isFalse,
      );
      expect(
        proof.isExpiredAt(
          DateTime.utc(2026, 9, 23, 16),
        ),
        isTrue,
      );
      expect(
        proof.isExpiredAt(
          DateTime.utc(2026, 9, 23, 16, 1),
        ),
        isTrue,
      );
    });
  });

  test('PasswordResetStep preserves the journey order', () {
    expect(
      PasswordResetStep.values,
      [
        PasswordResetStep.email,
        PasswordResetStep.code,
        PasswordResetStep.newPassword,
      ],
    );
  });
}
