import '../../result/result.dart';
import '../logging/console_log.dart';
import '../secure_storage/local_secure_storage.dart';
import 'auth_token_store.dart';
import 'dtos/app_tokens.dart';

enum _StorageKeys {
  accessToken('access_token'),
  refreshToken('refresh_token');

  final String key;
  const _StorageKeys(this.key);
}

class SecureAuthTokenStore implements AuthTokenStore {
  final LocalSecureStorage _store;

  SecureAuthTokenStore(this._store);

  final _log = ConsoleLog('SecureAuthTokenStore');

  @override
  AsyncResult<Unit> saveTokens(AppTokens tokens) async {
    final accessToken = tokens.accessToken.trim();
    final refreshToken = tokens.refreshToken.trim();

    if (accessToken.isEmpty || refreshToken.isEmpty) {
      return Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Access token or refresh token is empty',
        ),
      );
    }

    final result = await _store.write(
      _StorageKeys.accessToken.key,
      accessToken,
    );
    if (result.isFailure) {
      return Failure(result.error!);
    }

    final refreshResult = await _store.write(
      _StorageKeys.refreshToken.key,
      refreshToken,
    );
    if (refreshResult.isFailure) {
      final deleteResult = await _store.delete(_StorageKeys.accessToken.key);
      if (deleteResult.isFailure) {
        _log.error(
          'Failed to delete access token after refresh token write failure: ${deleteResult.error}',
        );
      }
      return Failure(refreshResult.error!);
    }

    return const Success(unit);
  }

  @override
  AsyncResult<String> readAccessToken() async {
    final result = await _store.read(_StorageKeys.accessToken.key);
    if (result.isFailure) {
      return Failure(result.error!);
    }

    if (result.value == null || result.value!.trim().isEmpty) {
      return Failure(
        AppError(
          code: AppErrorCode.storageError,
          message: 'Access token is empty',
        ),
      );
    }

    return Success(result.value!);
  }

  @override
  AsyncResult<String> readRefreshToken() async {
    final result = await _store.read(_StorageKeys.refreshToken.key);
    if (result.isFailure) {
      return Failure(result.error!);
    }
    if (result.value == null || result.value!.trim().isEmpty) {
      return Failure(
        AppError(
          code: AppErrorCode.storageError,
          message: 'Refresh token is empty',
        ),
      );
    }

    return Success(result.value!);
  }

  @override
  AsyncResult<Unit> clearTokens() async {
    final accessResult = await _store.delete(_StorageKeys.accessToken.key);
    final refreshResult = await _store.delete(_StorageKeys.refreshToken.key);

    if (accessResult.isFailure) {
      return Failure(accessResult.error!);
    }

    if (refreshResult.isFailure) {
      return Failure(refreshResult.error!);
    }

    return const Success(unit);
  }

  @override
  AsyncResult<Unit> updateAccessToken(String newAccessToken) async {
    final token = newAccessToken.trim();
    if (token.isEmpty) {
      return Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'New access token is empty',
        ),
      );
    }

    final result = await _store.write(
      _StorageKeys.accessToken.key,
      token,
    );
    if (result.isFailure) {
      return Failure(result.error!);
    }

    return const Success(unit);
  }
}
