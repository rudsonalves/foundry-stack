import '../../result/result.dart';
import 'dtos/app_tokens.dart';

abstract class AuthTokenStore {
  AsyncResult<Unit> saveTokens(AppTokens tokens);
  AsyncResult<String> readAccessToken();
  AsyncResult<String> readRefreshToken();
  AsyncResult<Unit> clearTokens();
  AsyncResult<Unit> updateAccessToken(String newAccessToken);
}
