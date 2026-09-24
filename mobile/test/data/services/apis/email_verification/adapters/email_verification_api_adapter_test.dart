import 'package:foundry_stack_mobile/data/services/apis/email_verification/adapters/email_verification_api_adapter.dart';
import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/confirm_email_verification_response_dto.dart';
import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/email_verification_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('EmailVerificationApiAdapter', () {
    test('maps request values to DTOs', () {
      final request = EmailVerificationApiAdapter.toRequest('ada@example.com');
      final confirmation = EmailVerificationApiAdapter.toConfirmRequest(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(request.email, 'ada@example.com');
      expect(confirmation.verificationId, 'verification-id');
      expect(confirmation.code, '012345');
    });

    test('maps challenge response to application model', () {
      final codeExpiresAt = DateTime.utc(2026, 9, 10, 15, 15);
      final resendAvailableAt = DateTime.utc(2026, 9, 10, 15, 1);
      final challenge = EmailVerificationApiAdapter.toChallenge(
        EmailVerificationResponseDto(
          verificationId: 'verification-id',
          codeExpiresAt: codeExpiresAt,
          resendAvailableAt: resendAvailableAt,
        ),
      );

      expect(challenge.verificationId, 'verification-id');
      expect(challenge.codeExpiresAt, codeExpiresAt);
      expect(challenge.resendAvailableAt, resendAvailableAt);
    });

    test('maps proof response to application model', () {
      final expiresAt = DateTime.utc(2026, 9, 11, 15);
      final proof = EmailVerificationApiAdapter.toProof(
        ConfirmEmailVerificationResponseDto(
          emailVerificationToken: 'opaque-secret',
          expiresAt: expiresAt,
        ),
      );

      expect(proof.token, 'opaque-secret');
      expect(proof.expiresAt, expiresAt);
    });
  });
}
