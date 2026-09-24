import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_proof.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('EmailVerificationChallenge', () {
    final challenge = EmailVerificationChallenge(
      verificationId: 'verification-id',
      codeExpiresAt: DateTime.utc(2026, 9, 10, 15, 15),
      resendAvailableAt: DateTime.utc(2026, 9, 10, 15, 1),
    );

    test('detects code expiration including the boundary', () {
      expect(
        challenge.isCodeExpiredAt(DateTime.utc(2026, 9, 10, 15, 14)),
        isFalse,
      );
      expect(
        challenge.isCodeExpiredAt(DateTime.utc(2026, 9, 10, 15, 15)),
        isTrue,
      );
    });

    test('allows resend beginning at the available instant', () {
      expect(challenge.canResendAt(DateTime.utc(2026, 9, 10, 15)), isFalse);
      expect(challenge.canResendAt(DateTime.utc(2026, 9, 10, 15, 1)), isTrue);
    });

    test('calculates resend remaining time including the boundary', () {
      expect(
        challenge.resendRemainingAt(DateTime.utc(2026, 9, 10, 15)),
        const Duration(minutes: 1),
      );
      expect(
        challenge.resendRemainingAt(DateTime.utc(2026, 9, 10, 15, 1)),
        Duration.zero,
      );
      expect(
        challenge.resendRemainingAt(DateTime.utc(2026, 9, 10, 15, 2)),
        Duration.zero,
      );
    });
  });

  test('EmailVerificationProof detects expiration including the boundary', () {
    final proof = EmailVerificationProof(
      token: 'opaque-secret',
      expiresAt: DateTime.utc(2026, 9, 11, 15),
    );

    expect(proof.isExpiredAt(DateTime.utc(2026, 9, 11, 14, 59)), isFalse);
    expect(proof.isExpiredAt(DateTime.utc(2026, 9, 11, 15)), isTrue);
  });
}
