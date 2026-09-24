import '/domain/common/password_reset/models/password_reset_challenge.dart';
import '/domain/common/password_reset/models/password_reset_proof.dart';
import '../dtos/complete_password_reset_request_dto.dart';
import '../dtos/confirm_password_reset_request_dto.dart';
import '../dtos/confirm_password_reset_response_dto.dart';
import '../dtos/password_reset_request_dto.dart';
import '../dtos/password_reset_response_dto.dart';

abstract final class PasswordResetApiAdapter {
  static PasswordResetRequestDto toRequest(String email) {
    return PasswordResetRequestDto(email: email);
  }

  static ConfirmPasswordResetRequestDto toConfirmRequest({
    required String passwordResetId,
    required String code,
  }) {
    return ConfirmPasswordResetRequestDto(
      passwordResetId: passwordResetId,
      code: code,
    );
  }

  static CompletePasswordResetRequestDto toCompleteRequest({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) {
    return CompletePasswordResetRequestDto(
      passwordResetId: passwordResetId,
      passwordResetToken: passwordResetToken,
      newPassword: newPassword,
    );
  }

  static PasswordResetChallenge toChallenge(
    PasswordResetResponseDto response, {
    required String email,
  }) {
    return PasswordResetChallenge(
      passwordResetId: response.passwordResetId,
      email: email,
      codeExpiresAt: response.codeExpiresAt,
      resendAvailableAt: response.resendAvailableAt,
    );
  }

  static PasswordResetProof toProof(ConfirmPasswordResetResponseDto response) {
    return PasswordResetProof(
      passwordResetToken: response.passwordResetToken,
      expiresAt: response.expiresAt,
    );
  }
}
