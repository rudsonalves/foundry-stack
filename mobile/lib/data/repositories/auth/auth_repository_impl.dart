import '/core/result/result.dart';
import '/core/services/auth_token_store/auth_token_store.dart';
import '/core/services/auth_token_store/dtos/app_tokens.dart';
import '/core/services/logging/console_log.dart';
import '/data/services/apis/auth/auth_service.dart';
import '/domain/common/auth/models/login_credentials.dart';
import 'auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthService _authService;
  final AuthTokenStore _authTokenStore;

  AuthRepositoryImpl({
    required AuthService authService,
    required AuthTokenStore authTokenStore,
  }) : _authService = authService,
       _authTokenStore = authTokenStore;

  final _log = ConsoleLog('AuthRepositoryImpl');

  @override
  AsyncResult<Unit> login(LoginCredentials credentials) async {
    final result = await _authService.login(credentials);

    if (result.isFailure) {
      _log.error('Login failed: ${result.error}');
      return Failure(result.error!);
    }

    final response = result.value!;
    final tokens = AppTokens(
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
    );
    final saveResult = await _authTokenStore.saveTokens(tokens);

    if (saveResult.isFailure) {
      _log.error('Failed to save tokens after login: ${saveResult.error}');
      return Failure(saveResult.error!);
    }

    return const Success(unit);
  }

  @override
  AsyncResult<RestoreSessionStatus> restoreSession() async {
    final refreshTokenResult = await _authTokenStore.readRefreshToken();

    if (refreshTokenResult.isFailure) {
      final error = refreshTokenResult.error!;

      if (error.code == AppErrorCode.storageNotFound) {
        final clearResult = await _authTokenStore.clearTokens();

        if (clearResult.isFailure) {
          _log.error(
            'Failed to clear local session without refresh token: '
            '${clearResult.error}',
          );
          return Failure(clearResult.error!);
        }

        return const Success(RestoreSessionStatus.noSession);
      }

      _log.error('Failed to read refresh token: $error');
      return Failure(error);
    }

    final refreshResult = await _authService.refreshToken(
      refreshTokenResult.value!,
    );

    if (refreshResult.isFailure) {
      final error = refreshResult.error!;

      if (error.code == AppErrorCode.unauthenticated) {
        final clearResult = await _authTokenStore.clearTokens();

        if (clearResult.isFailure) {
          _log.error(
            'Failed to clear rejected session: ${clearResult.error}',
          );
          return Failure(clearResult.error!);
        }

        return const Success(RestoreSessionStatus.noSession);
      }

      _log.error('Failed to refresh session: $error');
      return Failure(error);
    }

    final updateResult = await _authTokenStore.updateAccessToken(
      refreshResult.value!,
    );

    if (updateResult.isFailure) {
      _log.error(
        'Failed to save refreshed access token: ${updateResult.error}',
      );
      return Failure(updateResult.error!);
    }

    return const Success(RestoreSessionStatus.restored);
  }

  @override
  AsyncResult<Unit> logout() async {
    final refreshTokenResult = await _authTokenStore.readRefreshToken();

    if (refreshTokenResult.isSuccess) {
      final logoutResult = await _authService.logout(
        refreshTokenResult.value!,
      );

      if (logoutResult.isFailure) {
        _log.error(
          'Remote logout failed: ${logoutResult.error}',
        );
      }
    } else if (refreshTokenResult.error!.code != AppErrorCode.storageNotFound) {
      _log.error(
        'Failed to read refresh token for remote logout: '
        '${refreshTokenResult.error}',
      );
    }

    final clearResult = await _authTokenStore.clearTokens();

    if (clearResult.isFailure) {
      _log.error(
        'Failed to clear local session: ${clearResult.error}',
      );
      return Failure(clearResult.error!);
    }

    return const Success(unit);
  }

  @override
  AsyncResult<Unit> clearLocalSession() async {
    final result = await _authTokenStore.clearTokens();
    if (result.isFailure) {
      final error = result.error!;
      _log.error('Failed to clear local session: $error');
      return Failure(error);
    }

    return const Success(unit);
  }
}
