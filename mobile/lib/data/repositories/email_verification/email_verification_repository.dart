import '/core/result/result.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';

abstract class EmailVerificationRepository {
  AsyncResult<EmailVerificationChallenge> requestVerification(String email);

  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  });
}
