import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/dtos/app_tokens.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository_impl.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/auth_session.dart';

void main() {
  group('AuthRepositoryImpl.login', () {
    late _FakeAuthService authService;
    late _FakeAuthTokenStore tokenStore;
    late AuthRepositoryImpl repository;

    setUp(() {
      authService = _FakeAuthService();
      tokenStore = _FakeAuthTokenStore();
      repository = AuthRepositoryImpl(
        authService: authService,
        authTokenStore: tokenStore,
      );
    });

    test('persists the complete session before returning success', () async {
      authService.loginResult = Success(_loginResponse());
      tokenStore.expectedAccessToken = _accessToken;
      tokenStore.expectedRefreshToken = _refreshToken;
      final request = _loginRequest();

      final result = await repository.login(request);

      expect(result.isSuccess, isTrue);
      expect(result.value, unit);
      expect(identical(authService.lastLoginRequest, request), isTrue);
      expect(tokenStore.saveCalls, 1);
      expect(tokenStore.receivedExpectedSession, isTrue);
    });

    test('does not persist tokens after an HTTP failure', () async {
      const error = AppError(
        statusCode: 401,
        code: AppErrorCode.unauthenticated,
        message: 'invalid credentials',
      );
      authService.loginResult = const Failure(error);

      final result = await repository.login(_loginRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(tokenStore.saveCalls, 0);
    });

    test('returns failure when the session cannot be persisted', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'session persistence failed',
      );
      authService.loginResult = Success(_loginResponse());
      tokenStore.saveResult = const Failure(error);

      final result = await repository.login(_loginRequest());

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
      expect(tokenStore.saveCalls, 1);
      expect(tokenStore.hasUsableSession, isFalse);
    });
  });

  group('AuthRepositoryImpl.restoreSession', () {
    late _FakeAuthService authService;
    late _FakeAuthTokenStore tokenStore;
    late AuthRepositoryImpl repository;

    setUp(() {
      authService = _FakeAuthService();
      tokenStore = _FakeAuthTokenStore();
      repository = AuthRepositoryImpl(
        authService: authService,
        authTokenStore: tokenStore,
      );
    });

    test(
      'returns noSession without refreshing when credentials are absent',
      () async {
        tokenStore.readRefreshResult = const Failure(
          AppError(
            code: AppErrorCode.storageNotFound,
            message: 'credential not found',
          ),
        );

        final result = await repository.restoreSession();

        expect(result.value, RestoreSessionStatus.noSession);
        expect(authService.refreshCalls, 0);
        expect(tokenStore.clearCalls, 1);
      },
    );

    test('refreshes and persists the new access token', () async {
      tokenStore.readRefreshResult = const Success(_refreshToken);
      authService.expectedRefreshToken = _refreshToken;
      tokenStore.expectedAccessToken = _newAccessToken;
      authService.refreshResult = Success(_refreshResponse());

      final result = await repository.restoreSession();

      expect(result.value, RestoreSessionStatus.restored);
      expect(authService.refreshCalls, 1);
      expect(authService.receivedExpectedRefreshToken, isTrue);
      expect(tokenStore.updateCalls, 1);
      expect(tokenStore.receivedExpectedAccessToken, isTrue);
      expect(tokenStore.clearCalls, 0);
    });

    test(
      'clears a session rejected by the API and returns noSession',
      () async {
        const error = AppError(
          statusCode: 401,
          code: AppErrorCode.unauthenticated,
          message: 'session rejected',
        );
        tokenStore.readRefreshResult = const Success(_refreshToken);
        authService.refreshResult = const Failure(error);

        final result = await repository.restoreSession();

        expect(result.value, RestoreSessionStatus.noSession);
        expect(tokenStore.clearCalls, 1);
        expect(tokenStore.updateCalls, 0);
      },
    );

    test('preserves credentials and error after a transient failure', () async {
      const error = AppError(
        code: AppErrorCode.networkError,
        message: 'network unavailable',
      );
      tokenStore.readRefreshResult = const Success(_refreshToken);
      authService.refreshResult = const Failure(error);

      final result = await repository.restoreSession();

      expect(identical(result.error, error), isTrue);
      expect(tokenStore.clearCalls, 0);
      expect(tokenStore.updateCalls, 0);
    });

    test(
      'does not report a restored session when access update fails',
      () async {
        const error = AppError(
          code: AppErrorCode.storageError,
          message: 'access update failed',
        );
        tokenStore.readRefreshResult = const Success(_refreshToken);
        tokenStore.updateResult = const Failure(error);
        authService.refreshResult = Success(_refreshResponse());

        final result = await repository.restoreSession();

        expect(identical(result.error, error), isTrue);
        expect(tokenStore.updateCalls, 1);
        expect(tokenStore.clearCalls, 0);
      },
    );
  });

  group('AuthRepositoryImpl.logout', () {
    late _FakeAuthService authService;
    late _FakeAuthTokenStore tokenStore;
    late AuthRepositoryImpl repository;

    setUp(() {
      authService = _FakeAuthService();
      tokenStore = _FakeAuthTokenStore();
      repository = AuthRepositoryImpl(
        authService: authService,
        authTokenStore: tokenStore,
      );
    });

    test('clears local credentials when there is no session', () async {
      tokenStore.readRefreshResult = const Failure(
        AppError(
          code: AppErrorCode.storageNotFound,
          message: 'credential not found',
        ),
      );

      final result = await repository.logout();

      expect(result.isSuccess, isTrue);
      expect(authService.logoutCalls, 0);
      expect(tokenStore.clearCalls, 1);
    });

    test(
      'revokes the remote session before clearing local credentials',
      () async {
        tokenStore.readRefreshResult = const Success(_refreshToken);
        authService.expectedRefreshToken = _refreshToken;
        authService.logoutResult = const Success(unit);

        final result = await repository.logout();

        expect(result.isSuccess, isTrue);
        expect(authService.logoutCalls, 1);
        expect(authService.receivedExpectedLogoutToken, isTrue);
        expect(tokenStore.clearCalls, 1);
      },
    );

    test(
      'completes logout when remote revocation fails but cleanup succeeds',
      () async {
        tokenStore.readRefreshResult = const Success(_refreshToken);
        authService.logoutResult = const Failure(
          AppError(
            code: AppErrorCode.networkError,
            message: 'remote logout failed',
          ),
        );

        final result = await repository.logout();

        expect(result.isSuccess, isTrue);
        expect(authService.logoutCalls, 1);
        expect(tokenStore.clearCalls, 1);
      },
    );

    test(
      'returns cleanup failure even after successful remote revocation',
      () async {
        const error = AppError(
          code: AppErrorCode.storageError,
          message: 'cleanup failed',
        );
        tokenStore.readRefreshResult = const Success(_refreshToken);
        tokenStore.clearResult = const Failure(error);
        authService.logoutResult = const Success(unit);

        final result = await repository.logout();

        expect(identical(result.error, error), isTrue);
        expect(authService.logoutCalls, 1);
        expect(tokenStore.clearCalls, 1);
      },
    );
  });

  group('AuthRepositoryImpl.clearLocalSession', () {
    late _FakeAuthTokenStore tokenStore;
    late AuthRepositoryImpl repository;

    setUp(() {
      tokenStore = _FakeAuthTokenStore();
      repository = AuthRepositoryImpl(
        authService: _FakeAuthService(),
        authTokenStore: tokenStore,
      );
    });

    test('delegates local cleanup to the token store', () async {
      final result = await repository.clearLocalSession();

      expect(result.isSuccess, isTrue);
      expect(tokenStore.clearCalls, 1);
    });

    test('preserves a local cleanup failure', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'cleanup failed',
      );
      tokenStore.clearResult = const Failure(error);

      final result = await repository.clearLocalSession();

      expect(identical(result.error, error), isTrue);
      expect(tokenStore.clearCalls, 1);
    });
  });
}

const _accessToken = 'access-token-fixture';
const _newAccessToken = 'new-access-token-fixture';
const _refreshToken = 'refresh-token-fixture';

LoginCredentials _loginRequest() {
  return const LoginCredentials(
    email: 'user@example.com',
    password: 'password-fixture',
  );
}

AuthSession _loginResponse() {
  return const AuthSession(
    accessToken: _accessToken,
    refreshToken: _refreshToken,
  );
}

String _refreshResponse() => _newAccessToken;

class _FakeAuthService implements AuthService {
  Result<AuthSession>? loginResult;
  Result<String>? refreshResult;
  Result<Unit>? logoutResult;
  LoginCredentials? lastLoginRequest;
  String? expectedRefreshToken;
  var refreshCalls = 0;
  var logoutCalls = 0;
  var receivedExpectedRefreshToken = false;
  var receivedExpectedLogoutToken = false;

  @override
  AsyncResult<AuthSession> login(LoginCredentials request) async {
    lastLoginRequest = request;
    return loginResult!;
  }

  @override
  AsyncResult<String> refreshToken(String refreshToken) async {
    refreshCalls++;
    receivedExpectedRefreshToken = refreshToken == expectedRefreshToken;
    return refreshResult!;
  }

  @override
  AsyncResult<Unit> logout(String refreshToken) async {
    logoutCalls++;
    receivedExpectedLogoutToken = refreshToken == expectedRefreshToken;
    return logoutResult!;
  }
}

class _FakeAuthTokenStore implements AuthTokenStore {
  Result<Unit> saveResult = const Success(unit);
  Result<String>? readRefreshResult;
  Result<Unit> updateResult = const Success(unit);
  Result<Unit> clearResult = const Success(unit);
  String? expectedAccessToken;
  String? expectedRefreshToken;
  var saveCalls = 0;
  var updateCalls = 0;
  var clearCalls = 0;
  var receivedExpectedSession = false;
  var receivedExpectedAccessToken = false;
  var hasUsableSession = false;

  @override
  AsyncResult<Unit> saveTokens(AppTokens tokens) async {
    saveCalls++;
    receivedExpectedSession =
        tokens.accessToken == expectedAccessToken &&
        tokens.refreshToken == expectedRefreshToken;
    hasUsableSession = saveResult.isSuccess;
    return saveResult;
  }

  @override
  AsyncResult<Unit> clearTokens() async {
    clearCalls++;
    if (clearResult.isSuccess) {
      hasUsableSession = false;
    }
    return clearResult;
  }

  @override
  AsyncResult<String> readAccessToken() {
    throw UnimplementedError();
  }

  @override
  AsyncResult<String> readRefreshToken() async {
    return readRefreshResult!;
  }

  @override
  AsyncResult<Unit> updateAccessToken(String newAccessToken) async {
    updateCalls++;
    receivedExpectedAccessToken = newAccessToken == expectedAccessToken;
    return updateResult;
  }
}
