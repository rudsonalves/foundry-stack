import 'package:foundry_stack_mobile/app/dependencies/app_container.dart';
import 'package:foundry_stack_mobile/app/dependencies/app_dependencies.dart';
import 'package:foundry_stack_mobile/core/config/core_container.dart';
import 'package:foundry_stack_mobile/core/services/auth_token_store/auth_token_store.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/user_service.dart';
import 'package:foundry_stack_mobile/ui/pages/auth/login/viewmodel/login_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/home/viewmodel/home_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/splash/viewmodel/splash_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const baseUrl = 'https://api.example.test';

  late CoreContainer coreContainer;
  late AppContainer appContainer;

  setUp(() {
    coreContainer = CoreContainer();
    final core = coreContainer.initialize(baseUrl: baseUrl);
    appContainer = AppContainer(
      AppDependencies(
        restClient: core.restClient,
        authTokenStore: core.authTokenStore,
        localSecureStorage: core.localSecureStorage,
      ),
    );
  });

  tearDown(() {
    appContainer.dispose();
    coreContainer.dispose();
  });

  test('does not resolve dependencies before initialization', () {
    expect(() => appContainer.get<AuthService>(), throwsStateError);
  });

  test('resolves the complete operational graph', () {
    appContainer.initialize();

    expect(appContainer.get<AuthService>(), isA<AuthService>());
    expect(appContainer.get<AuthRepository>(), isA<AuthRepository>());
    expect(appContainer.get<HomeViewmodel>(), isA<HomeViewmodel>());
    expect(appContainer.get<SplashViewmodel>(), isA<SplashViewmodel>());
    expect(appContainer.get<LoginViewmodel>(), isA<LoginViewmodel>());
    expect(
      appContainer.get<OnboardingDraftRepository>(),
      isA<OnboardingDraftRepository>(),
    );
  });

  test('uses the same dependencies owned by core', () {
    final core = coreContainer.dependencies;
    appContainer.initialize();

    expect(identical(appContainer.get<RestClient>(), core.restClient), isTrue);
    expect(
      identical(appContainer.get<AuthTokenStore>(), core.authTokenStore),
      isTrue,
    );
    expect(
      identical(
        appContainer.get<LocalSecureStorage>(),
        core.localSecureStorage,
      ),
      isTrue,
    );
  });

  test('does not contain onboarding registrations', () {
    appContainer.initialize();

    expect(() => appContainer.get<UserService>(), throwsA(anything));
  });

  test('preserves current registration lifetimes', () {
    appContainer.initialize();

    final authService = appContainer.get<AuthService>();
    final authRepository = appContainer.get<AuthRepository>();

    expect(identical(appContainer.get<AuthService>(), authService), isTrue);
    expect(
      identical(appContainer.get<AuthRepository>(), authRepository),
      isFalse,
    );
  });

  test('cannot initialize twice or resolve after disposal', () {
    appContainer.initialize();

    expect(appContainer.initialize, throwsStateError);

    appContainer.dispose();

    expect(() => appContainer.get<AuthService>(), throwsStateError);
  });
}
