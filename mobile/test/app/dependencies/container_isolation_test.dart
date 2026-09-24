import 'package:foundry_stack_mobile/app/dependencies/app_container.dart';
import 'package:foundry_stack_mobile/app/dependencies/app_dependencies.dart';
import 'package:foundry_stack_mobile/app/dependencies/onboarding_dependencies.dart';
import 'package:foundry_stack_mobile/app/dependencies/onboarding_scope.dart';
import 'package:foundry_stack_mobile/core/config/core_dependencies.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/dtos/app_tokens.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_request.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client_response.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/user_service.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('isolates app and onboarding while sharing core dependencies', () async {
    final restClient = _FakeRestClient();
    final storage = _MemoryStorage();
    final tokenStore = _FakeAuthTokenStore();
    final core = CoreDependencies(
      restClient: restClient,
      localSecureStorage: storage,
      authTokenStore: tokenStore,
    );
    final app = AppContainer(
      AppDependencies(
        restClient: core.restClient,
        authTokenStore: core.authTokenStore,
        localSecureStorage: core.localSecureStorage,
      ),
    )..initialize();
    final onboarding = OnboardingScope(
      OnboardingDependencies(
        restClient: core.restClient,
        localSecureStorage: core.localSecureStorage,
      ),
    )..initialize();
    addTearDown(() {
      onboarding.dispose();
      app.dispose();
    });

    expect(() => onboarding.get<AuthService>(), throwsA(anything));
    expect(() => app.get<UserService>(), throwsA(anything));
    expect(identical(app.get<RestClient>(), restClient), isTrue);
    expect(identical(onboarding.get<RestClient>(), restClient), isTrue);
    expect(identical(app.get<AuthTokenStore>(), tokenStore), isTrue);
    expect(identical(app.get<LocalSecureStorage>(), storage), isTrue);
    expect(identical(onboarding.get<LocalSecureStorage>(), storage), isTrue);

    await storage.write('onboarding.draft', '{"email":"user@example.com"}');
    onboarding.dispose();
    onboarding.dispose();

    expect(() => onboarding.get<UserService>(), throwsStateError);
    expect(app.get<AuthService>(), isA<AuthService>());
    expect(identical(app.get<RestClient>(), restClient), isTrue);
    expect(identical(app.get<AuthTokenStore>(), tokenStore), isTrue);
    expect((await storage.read('onboarding.draft')).value, isNotNull);
  });
}

final class _MemoryStorage implements LocalSecureStorage {
  final Map<String, String> values = {};

  @override
  AsyncResult<Unit> write(String key, String value) async {
    values[key] = value;
    return const Success(unit);
  }

  @override
  AsyncResult<String> read(String key) async {
    final value = values[key];
    return value == null
        ? Failure(
            AppError(
              code: AppErrorCode.storageError,
              message: 'Value not found.',
            ),
          )
        : Success(value);
  }

  @override
  AsyncResult<Unit> delete(String key) async {
    values.remove(key);
    return const Success(unit);
  }

  @override
  AsyncResult<Unit> deleteAll() async {
    values.clear();
    return const Success(unit);
  }

  @override
  AsyncResult<List<String>> keysWithPrefix(String pattern) async =>
      Success(values.keys.where((key) => key.startsWith(pattern)).toList());
}

final class _FakeRestClient implements RestClient {
  @override
  AsyncResult<RestClientResponse> get(RestClientRequest request) =>
      throw UnimplementedError();
  @override
  AsyncResult<RestClientResponse> post(RestClientRequest request) =>
      throw UnimplementedError();
  @override
  AsyncResult<RestClientResponse> put(RestClientRequest request) =>
      throw UnimplementedError();
  @override
  AsyncResult<RestClientResponse> patch(RestClientRequest request) =>
      throw UnimplementedError();
  @override
  AsyncResult<RestClientResponse> delete(RestClientRequest request) =>
      throw UnimplementedError();
}

final class _FakeAuthTokenStore implements AuthTokenStore {
  @override
  AsyncResult<Unit> saveTokens(AppTokens tokens) => throw UnimplementedError();
  @override
  AsyncResult<String> readAccessToken() => throw UnimplementedError();
  @override
  AsyncResult<String> readRefreshToken() => throw UnimplementedError();
  @override
  AsyncResult<Unit> clearTokens() => throw UnimplementedError();
  @override
  AsyncResult<Unit> updateAccessToken(String newAccessToken) =>
      throw UnimplementedError();
}
