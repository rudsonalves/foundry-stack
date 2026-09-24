import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/dtos/app_tokens.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/secure_auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';

void main() {
  group('SecureAuthTokenStore reads', () {
    late _FakeLocalSecureStorage storage;
    late SecureAuthTokenStore tokenStore;

    setUp(() {
      storage = _FakeLocalSecureStorage();
      tokenStore = SecureAuthTokenStore(storage);
    });

    test(
      'reads access and refresh tokens from their respective keys',
      () async {
        storage.values['access_token'] = 'access-token';
        storage.values['refresh_token'] = 'refresh-token';

        final accessResult = await tokenStore.readAccessToken();
        final refreshResult = await tokenStore.readRefreshToken();

        expect(accessResult.value, 'access-token');
        expect(refreshResult.value, 'refresh-token');
        expect(storage.readKeys, ['access_token', 'refresh_token']);
      },
    );

    test('preserves storageNotFound when a token is absent', () async {
      const error = AppError(
        code: AppErrorCode.storageNotFound,
        message: 'key not found',
      );
      storage.readResults['access_token'] = const Failure(error);

      final result = await tokenStore.readAccessToken();

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });

    test('preserves a storage read failure', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'read failed',
      );
      storage.readResults['refresh_token'] = const Failure(error);

      final result = await tokenStore.readRefreshToken();

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });

    test('rejects an empty token returned by storage', () async {
      storage.values['access_token'] = '   ';

      final result = await tokenStore.readAccessToken();

      expect(result.isFailure, isTrue);
      expect(result.error!.code, AppErrorCode.storageError);
    });
  });

  group('SecureAuthTokenStore.saveTokens', () {
    late _FakeLocalSecureStorage storage;
    late SecureAuthTokenStore tokenStore;

    setUp(() {
      storage = _FakeLocalSecureStorage();
      tokenStore = SecureAuthTokenStore(storage);
    });

    test('saves both trimmed tokens', () async {
      final result = await tokenStore.saveTokens(
        AppTokens(
          accessToken: ' access-token ',
          refreshToken: ' refresh-token ',
        ),
      );

      expect(result.isSuccess, isTrue);
      expect(storage.writeCalls, [
        const _WriteCall('access_token', 'access-token'),
        const _WriteCall('refresh_token', 'refresh-token'),
      ]);
    });

    test('rejects empty tokens without writing', () async {
      final result = await tokenStore.saveTokens(
        AppTokens(accessToken: ' ', refreshToken: 'refresh-token'),
      );

      expect(result.isFailure, isTrue);
      expect(result.error!.code, AppErrorCode.invalidData);
      expect(storage.writeCalls, isEmpty);
    });

    test(
      'preserves first write failure and does not write refresh token',
      () async {
        const error = AppError(
          code: AppErrorCode.storageError,
          message: 'access write failed',
        );
        storage.writeResults['access_token'] = const Failure(error);

        final result = await tokenStore.saveTokens(_tokens());

        expect(identical(result.error, error), isTrue);
        expect(storage.writeCalls, [
          const _WriteCall('access_token', 'access-token'),
        ]);
        expect(storage.deleteKeys, isEmpty);
      },
    );

    test('removes access token when refresh token write fails', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'refresh write failed',
      );
      storage.writeResults['refresh_token'] = const Failure(error);

      final result = await tokenStore.saveTokens(_tokens());

      expect(identical(result.error, error), isTrue);
      expect(storage.deleteKeys, ['access_token']);
    });

    test('still preserves write failure when rollback also fails', () async {
      const writeError = AppError(
        code: AppErrorCode.storageError,
        message: 'refresh write failed',
      );
      const deleteError = AppError(
        code: AppErrorCode.storageError,
        message: 'rollback failed',
      );
      storage.writeResults['refresh_token'] = const Failure(writeError);
      storage.deleteResults['access_token'] = const Failure(deleteError);

      final result = await tokenStore.saveTokens(_tokens());

      expect(identical(result.error, writeError), isTrue);
      expect(storage.deleteKeys, ['access_token']);
    });
  });

  group('SecureAuthTokenStore.updateAccessToken', () {
    late _FakeLocalSecureStorage storage;
    late SecureAuthTokenStore tokenStore;

    setUp(() {
      storage = _FakeLocalSecureStorage();
      tokenStore = SecureAuthTokenStore(storage);
    });

    test('updates only the access token', () async {
      final result = await tokenStore.updateAccessToken(' new-access-token ');

      expect(result.isSuccess, isTrue);
      expect(storage.writeCalls, [
        const _WriteCall('access_token', 'new-access-token'),
      ]);
    });

    test('rejects an empty access token without writing', () async {
      final result = await tokenStore.updateAccessToken('   ');

      expect(result.isFailure, isTrue);
      expect(result.error!.code, AppErrorCode.invalidData);
      expect(storage.writeCalls, isEmpty);
    });

    test('preserves access token write failure', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'update failed',
      );
      storage.writeResults['access_token'] = const Failure(error);

      final result = await tokenStore.updateAccessToken('access-token');

      expect(identical(result.error, error), isTrue);
    });
  });

  group('SecureAuthTokenStore.clearTokens', () {
    late _FakeLocalSecureStorage storage;
    late SecureAuthTokenStore tokenStore;

    setUp(() {
      storage = _FakeLocalSecureStorage();
      tokenStore = SecureAuthTokenStore(storage);
    });

    test('deletes both tokens', () async {
      final result = await tokenStore.clearTokens();

      expect(result.isSuccess, isTrue);
      expect(storage.deleteKeys, ['access_token', 'refresh_token']);
    });

    test('attempts both deletes when access token deletion fails', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'access delete failed',
      );
      storage.deleteResults['access_token'] = const Failure(error);

      final result = await tokenStore.clearTokens();

      expect(identical(result.error, error), isTrue);
      expect(storage.deleteKeys, ['access_token', 'refresh_token']);
    });

    test('returns refresh token deletion failure', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'refresh delete failed',
      );
      storage.deleteResults['refresh_token'] = const Failure(error);

      final result = await tokenStore.clearTokens();

      expect(identical(result.error, error), isTrue);
      expect(storage.deleteKeys, ['access_token', 'refresh_token']);
    });
  });
}

AppTokens _tokens() => AppTokens(
  accessToken: 'access-token',
  refreshToken: 'refresh-token',
);

class _FakeLocalSecureStorage implements LocalSecureStorage {
  final values = <String, String>{};
  final readResults = <String, Result<String>>{};
  final writeResults = <String, Result<Unit>>{};
  final deleteResults = <String, Result<Unit>>{};
  final readKeys = <String>[];
  final writeCalls = <_WriteCall>[];
  final deleteKeys = <String>[];

  @override
  AsyncResult<String> read(String key) async {
    readKeys.add(key);
    final configured = readResults[key];
    if (configured != null) {
      return configured;
    }
    final value = values[key];
    if (value == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.storageNotFound,
          message: 'key not found',
        ),
      );
    }
    return Success(value);
  }

  @override
  AsyncResult<Unit> write(String key, String value) async {
    writeCalls.add(_WriteCall(key, value));
    final result = writeResults[key] ?? const Success(unit);
    if (result.isSuccess) {
      values[key] = value;
    }
    return result;
  }

  @override
  AsyncResult<Unit> delete(String key) async {
    deleteKeys.add(key);
    final result = deleteResults[key] ?? const Success(unit);
    if (result.isSuccess) {
      values.remove(key);
    }
    return result;
  }

  @override
  AsyncResult<Unit> deleteAll() async {
    values.clear();
    return const Success(unit);
  }

  @override
  AsyncResult<List<String>> keysWithPrefix(String pattern) async {
    return Success(
      values.keys.where((key) => key.startsWith(pattern)).toList(),
    );
  }
}

class _WriteCall {
  final String key;
  final String value;

  const _WriteCall(this.key, this.value);

  @override
  bool operator ==(Object other) {
    return other is _WriteCall && other.key == key && other.value == value;
  }

  @override
  int get hashCode => Object.hash(key, value);
}
