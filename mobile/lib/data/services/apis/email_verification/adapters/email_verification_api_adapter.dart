import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '../dtos/confirm_email_verification_request_dto.dart';
import '../dtos/confirm_email_verification_response_dto.dart';
import '../dtos/email_verification_request_dto.dart';
import '../dtos/email_verification_response_dto.dart';

abstract final class EmailVerificationApiAdapter {
  static EmailVerificationRequestDto toRequest(String email) {
    return EmailVerificationRequestDto(email: email);
  }

  static ConfirmEmailVerificationRequestDto toConfirmRequest({
    required String verificationId,
    required String code,
  }) {
    return ConfirmEmailVerificationRequestDto(
      verificationId: verificationId,
      code: code,
    );
  }

  static EmailVerificationChallenge toChallenge(
    EmailVerificationResponseDto response,
  ) {
    return EmailVerificationChallenge(
      verificationId: response.verificationId,
      codeExpiresAt: response.codeExpiresAt,
      resendAvailableAt: response.resendAvailableAt,
    );
  }

  static EmailVerificationProof toProof(
    ConfirmEmailVerificationResponseDto response,
  ) {
    return EmailVerificationProof(
      token: response.emailVerificationToken,
      expiresAt: response.expiresAt,
    );
  }
}
