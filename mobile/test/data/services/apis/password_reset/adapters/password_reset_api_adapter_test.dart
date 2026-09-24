import 'package:foundry_stack_mobile/data/services/apis/password_reset/adapters/password_reset_api_adapter.dart';
import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/confirm_password_reset_response_dto.dart';
import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/password_reset_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('PasswordResetApiAdapter', () {
    test('maps request values to DTOs', () {
      final request = PasswordResetApiAdapter.toRequest(
        'user@example.com',
      );
      final confirmation = PasswordResetApiAdapter.toConfirmRequest(
        passwordResetId: 'password-reset-id',
        code: '012345',
      );
      final completion = PasswordResetApiAdapter.toCompleteRequest(
        passwordResetId: 'password-reset-id',
        passwordResetToken: 'opaque-proof',
        newPassword: 'new-password123',
      );

      expect(request.email, 'user@example.com');
      expect(confirmation.passwordResetId, 'password-reset-id');
      expect(confirmation.code, '012345');
      expect(completion.passwordResetId, 'password-reset-id');
      expect(completion.passwordResetToken, 'opaque-proof');
      expect(completion.newPassword, 'new-password123');
    });

    test('maps response and local e-mail to challenge', () {
      final codeExpiresAt = DateTime.utc(
        2026,
        9,
        23,
        15,
        15,
      );
      final resendAvailableAt = DateTime.utc(
        2026,
        9,
        23,
        15,
        1,
      );

      final challenge = PasswordResetApiAdapter.toChallenge(
        PasswordResetResponseDto(
          passwordResetId: 'password-reset-id',
          codeExpiresAt: codeExpiresAt,
          resendAvailableAt: resendAvailableAt,
        ),
        email: 'user@example.com',
      );

      expect(challenge.passwordResetId, 'password-reset-id');
      expect(challenge.email, 'user@example.com');
      expect(challenge.codeExpiresAt, codeExpiresAt);
      expect(
        challenge.resendAvailableAt,
        resendAvailableAt,
      );
    });

    test('maps confirmation response to proof', () {
      final expiresAt = DateTime.utc(2026, 9, 23, 16);

      final proof = PasswordResetApiAdapter.toProof(
        ConfirmPasswordResetResponseDto(
          passwordResetToken: 'opaque-proof',
          expiresAt: expiresAt,
        ),
      );

      expect(proof.passwordResetToken, 'opaque-proof');
      expect(proof.expiresAt, expiresAt);
    });
  });
}
