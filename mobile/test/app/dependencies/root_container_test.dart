import 'package:foundry_stack_mobile/app/dependencies/root_container.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const baseUrl = 'https://api.example.test';

  test('composes core and app in order', () {
    final root = RootContainer();
    addTearDown(root.dispose);

    expect(root.initializeApp, throwsStateError);

    root.initializeCore(baseUrl: baseUrl);
    root.initializeApp();

    expect(root.appContainer.get<AuthService>(), isA<AuthService>());
  });

  test('composes onboarding without initializing app', () {
    final root = RootContainer();
    addTearDown(root.dispose);

    root.initializeCore(baseUrl: baseUrl);
    final scope = root.createOnboardingScope();

    expect(scope.get<OnboardingUsecase>(), isA<OnboardingUsecase>());
    expect(() => root.appContainer, throwsStateError);
  });

  test('disposes active onboarding before app and core become unavailable', () {
    final root = RootContainer();
    root.initializeCore(baseUrl: baseUrl);
    root.initializeApp();
    final scope = root.createOnboardingScope();

    root.dispose();
    root.dispose();

    expect(() => scope.get<OnboardingUsecase>(), throwsStateError);
    expect(() => root.appContainer, throwsStateError);
    expect(root.createOnboardingScope, throwsStateError);
  });
}
