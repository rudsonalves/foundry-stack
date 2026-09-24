import 'package:foundry_stack_mobile/app/dependencies/onboarding_dependencies.dart';
import 'package:foundry_stack_mobile/app/dependencies/onboarding_scope.dart';
import 'package:foundry_stack_mobile/core/config/core_container.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/client/rest_client.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/repositories/email_verification/email_verification_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/user/user_repository.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/auth_service.dart';
import 'package:foundry_stack_mobile/data/services/apis/email_verification/email_verification_service.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/user_service.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/domain/usecases/users_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/viewmodel/onboarding_coordinator_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  const baseUrl = 'https://api.example.test';

  late CoreContainer coreContainer;
  late OnboardingDependencies dependencies;
  final scopes = <OnboardingScope>[];

  setUp(() {
    coreContainer = CoreContainer();
    final core = coreContainer.initialize(baseUrl: baseUrl);
    dependencies = OnboardingDependencies(
      restClient: core.restClient,
      localSecureStorage: core.localSecureStorage,
    );
  });

  tearDown(() {
    for (final scope in scopes) {
      scope.dispose();
    }
    scopes.clear();
    coreContainer.dispose();
  });

  OnboardingScope createScope() {
    final scope = OnboardingScope(dependencies)..initialize();
    scopes.add(scope);
    return scope;
  }

  test('blocks resolution before initialization', () {
    final core = coreContainer.dependencies;
    final scope = OnboardingScope(
      OnboardingDependencies(
        restClient: core.restClient,
        localSecureStorage: core.localSecureStorage,
      ),
    );
    scopes.add(scope);

    expect(() => scope.get<UserService>(), throwsStateError);
  });

  test('initialized scope contains the registration graph', () {
    final scope = createScope();

    expect(
      scope.get<EmailVerificationService>(),
      isA<EmailVerificationService>(),
    );
    expect(
      scope.get<EmailVerificationRepository>(),
      isA<EmailVerificationRepository>(),
    );
    expect(
      scope.get<OnboardingDraftRepository>(),
      isA<OnboardingDraftRepository>(),
    );
    expect(scope.get<UserService>(), isA<UserService>());
    expect(scope.get<UserRepository>(), isA<UserRepository>());
    expect(scope.get<OnboardingUsecase>(), isA<OnboardingUsecase>());
    expect(
      () => scope.get<OnboardingCoordinatorViewmodel>(),
      throwsA(anything),
    );
  });

  test('uses core instances without exposing operational registrations', () {
    final core = coreContainer.dependencies;
    final scope = createScope();

    expect(identical(scope.get<RestClient>(), core.restClient), isTrue);
    expect(
      identical(
        scope.get<LocalSecureStorage>(),
        core.localSecureStorage,
      ),
      isTrue,
    );
    expect(() => scope.get<AuthService>(), throwsA(anything));
    expect(() => scope.get<UsersUsecase>(), throwsA(anything));
  });

  test('dispose is idempotent and blocks subsequent resolutions', () {
    final scope = createScope();
    scope.get<OnboardingUsecase>();

    scope.dispose();
    scope.dispose();

    expect(() => scope.get<OnboardingUsecase>(), throwsStateError);
  });

  test('keeps one use case per scope and isolates different journeys', () {
    final first = createScope();
    final second = createScope();
    final firstUsecase = first.get<OnboardingUsecase>();
    final secondUsecase = second.get<OnboardingUsecase>();

    first.dispose();

    expect(identical(firstUsecase, secondUsecase), isFalse);
    expect(
      identical(second.get<OnboardingUsecase>(), secondUsecase),
      isTrue,
    );
    expect(
      identical(
        second.get<RestClient>(),
        coreContainer.dependencies.restClient,
      ),
      isTrue,
    );
  });

  test('disposing a scope does not delete the persisted draft', () async {
    final storage = _FakeLocalSecureStorage();
    final scope = OnboardingScope(
      OnboardingDependencies(
        restClient: coreContainer.dependencies.restClient,
        localSecureStorage: storage,
      ),
    )..initialize();
    scopes.add(scope);
    final repository = scope.get<OnboardingDraftRepository>();

    await repository.save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.email,
        name: 'Ada',
        expiresAt: DateTime.now().toUtc().add(const Duration(hours: 1)),
      ),
    );
    scope.dispose();

    expect(storage.deleteCalls, 0);
    expect(storage.values, isNotEmpty);
  });

  test('a new scope can load the draft saved by a previous scope', () async {
    final storage = _FakeLocalSecureStorage();
    final sharedDependencies = OnboardingDependencies(
      restClient: coreContainer.dependencies.restClient,
      localSecureStorage: storage,
    );
    final first = OnboardingScope(sharedDependencies)..initialize();
    scopes.add(first);

    await first.get<OnboardingDraftRepository>().save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.email,
        name: 'Ada',
        expiresAt: DateTime.now().toUtc().add(const Duration(hours: 1)),
      ),
    );
    first.dispose();

    final second = OnboardingScope(sharedDependencies)..initialize();
    scopes.add(second);
    final result = await second.get<OnboardingDraftRepository>().load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.draft?.step, OnboardingStep.email);
    expect(result.value?.draft?.name, 'Ada');
  });
}

class _FakeLocalSecureStorage implements LocalSecureStorage {
  final values = <String, String>{};
  int deleteCalls = 0;

  @override
  AsyncResult<String> read(String key) async {
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
    values[key] = value;
    return const Success(unit);
  }

  @override
  AsyncResult<Unit> delete(String key) async {
    deleteCalls++;
    values.remove(key);
    return const Success(unit);
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
