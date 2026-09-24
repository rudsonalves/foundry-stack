import 'package:auto_injector/auto_injector.dart';
import 'package:foundry_stack_mobile/core/config/core_container.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/dio/dio_rest_client.dart';
import 'package:foundry_stack_mobile/core/services/client_http/interceptors/auth/auth_interceptor.dart';
import 'package:foundry_stack_mobile/core/config/core_bindings.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const baseUrl = 'https://api.example.test';

  group('CoreContainer', () {
    late CoreContainer container;

    setUp(() => container = CoreContainer());
    tearDown(() => container.dispose());

    test('does not expose dependencies before commit', () {
      expect(() => container.dependencies, throwsStateError);
    });

    test('exposes a stable and limited dependency view after initialize', () {
      final initialized = container.initialize(baseUrl: baseUrl);

      expect(identical(container.dependencies, initialized), isTrue);
      expect(initialized.restClient, isA<RestClient>());
      expect(initialized.localSecureStorage, isA<LocalSecureStorage>());
      expect(initialized.authTokenStore, isA<AuthTokenStore>());
    });

    test('cannot initialize the same container twice', () {
      container.initialize(baseUrl: baseUrl);

      expect(
        () => container.initialize(baseUrl: baseUrl),
        throwsStateError,
      );
    });

    test('does not expose dependencies after disposal', () {
      container.initialize(baseUrl: baseUrl);
      container.dispose();

      expect(() => container.dependencies, throwsStateError);
      expect(
        () => container.initialize(baseUrl: baseUrl),
        throwsStateError,
      );
    });
  });

  test('core bindings compose shared singletons before RestClient', () {
    final injector = AutoInjector(tag: 'core-test');
    addTearDown(injector.dispose);

    registerCoreBindings(injector, baseUrl: baseUrl);
    injector.commit();

    final firstStorage = injector.get<LocalSecureStorage>();
    final tokenStore = injector.get<AuthTokenStore>();
    final dio = injector.get<Dio>();
    final interceptor = injector.get<AuthInterceptor>();
    final restClient = injector.get<RestClient>();

    expect(identical(injector.get<LocalSecureStorage>(), firstStorage), isTrue);
    expect(identical(injector.get<AuthTokenStore>(), tokenStore), isTrue);
    expect(identical(injector.get<RestClient>(), restClient), isTrue);
    expect(restClient, isA<DioRestClient>());
    expect(
      dio.interceptors.where((item) => identical(item, interceptor)),
      hasLength(1),
    );
  });
}
