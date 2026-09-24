import '/core/result/result.dart';
import '/data/services/apis/email_verification/email_verification_service.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import 'email_verification_repository.dart';

class EmailVerificationRepositoryImpl implements EmailVerificationRepository {
  final EmailVerificationService _service;

  EmailVerificationRepositoryImpl({
    required EmailVerificationService service,
  }) : _service = service;

  @override
  AsyncResult<EmailVerificationChallenge> requestVerification(String email) {
    return _service.requestVerification(email);
  }

  @override
  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  }) {
    return _service.confirmVerification(
      verificationId: verificationId,
      code: code,
    );
  }
}
