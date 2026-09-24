import '../../../core/result/result.dart';
import '/domain/common/auth/models/login_credentials.dart';

enum RestoreSessionStatus {
  restored,
  noSession,
}

abstract class AuthRepository {
  AsyncResult<Unit> login(LoginCredentials credentials);
  AsyncResult<RestoreSessionStatus> restoreSession();
  AsyncResult<Unit> logout();
  AsyncResult<Unit> clearLocalSession();
}
