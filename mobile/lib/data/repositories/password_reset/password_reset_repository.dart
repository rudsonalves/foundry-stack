import '/core/result/result.dart';
import '/domain/common/password_reset/models/password_reset_challenge.dart';
import '/domain/common/password_reset/models/password_reset_proof.dart';

abstract class PasswordResetRepository {
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email);

  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  });

  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  });
}
