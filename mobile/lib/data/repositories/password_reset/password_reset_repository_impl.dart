import '/core/result/result.dart';
import '/domain/common/password_reset/models/password_reset_challenge.dart';
import '/domain/common/password_reset/models/password_reset_proof.dart';
import '../../services/apis/password_reset/password_reset_service.dart';
import 'password_reset_repository.dart';

class PasswordResetRepositoryImpl implements PasswordResetRepository {
  final PasswordResetService _service;

  PasswordResetRepositoryImpl({
    required PasswordResetService service,
  }) : _service = service;

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) {
    return _service.requestPasswordReset(email);
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) {
    return _service.confirmPasswordReset(
      passwordResetId: passwordResetId,
      code: code,
    );
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) {
    return _service.completePasswordReset(
      passwordResetId: passwordResetId,
      passwordResetToken: passwordResetToken,
      newPassword: newPassword,
    );
  }
}
