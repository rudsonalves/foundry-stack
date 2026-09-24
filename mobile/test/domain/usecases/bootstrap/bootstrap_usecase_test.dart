import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_draft_load.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/bootstrap_usecase.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/models/bootstrap_destination.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeAuthRepository authRepository;
  late _FakeOnboardingDraftRepository draftRepository;
  late BootstrapUsecase usecase;

  setUp(() {
    authRepository = _FakeAuthRepository();
    draftRepository = _FakeOnboardingDraftRepository();
    usecase = BootstrapUsecase(
      authRepository: authRepository,
      onboardingDraftRepository: draftRepository,
    );
  });

  test('authenticated session goes home without loading the draft', () async {
    authRepository.result = const Success(RestoreSessionStatus.restored);

    final result = await usecase.initialize();

    expect(result.value, BootstrapDestination.home);
    expect(authRepository.restoreSessionCalls, 1);
    expect(draftRepository.loadCalls, 0);
  });

  test('valid draft goes to onboarding when there is no session', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.email,
      expiresAt: DateTime.utc(2026, 9, 22),
    );
    draftRepository.result = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.value, BootstrapDestination.onboarding);
    expect(authRepository.restoreSessionCalls, 1);
    expect(draftRepository.loadCalls, 1);
  });

  test('absent draft goes to login when there is no session', () async {
    final result = await usecase.initialize();

    expect(result.value, BootstrapDestination.login);
    expect(authRepository.restoreSessionCalls, 1);
    expect(draftRepository.loadCalls, 1);
  });

  for (final code in <AppErrorCode>[
    AppErrorCode.networkError,
    AppErrorCode.timeout,
    AppErrorCode.storageError,
  ]) {
    test('preserves ${code.name} from session restoration', () async {
      final error = AppError(code: code, message: '${code.name} failure');
      authRepository.result = Failure(error);

      final result = await usecase.initialize();

      expect(result.error, same(error));
      expect(draftRepository.loadCalls, 0);
    });
  }

  test('preserves a failure while loading the draft', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'draft read failed',
    );
    draftRepository.result = const Failure(error);

    final result = await usecase.initialize();

    expect(result.error, same(error));
    expect(authRepository.restoreSessionCalls, 1);
    expect(draftRepository.loadCalls, 1);
  });
}

class _FakeAuthRepository implements AuthRepository {
  Result<RestoreSessionStatus> result = const Success(
    RestoreSessionStatus.noSession,
  );
  int restoreSessionCalls = 0;

  @override
  AsyncResult<RestoreSessionStatus> restoreSession() async {
    restoreSessionCalls++;
    return result;
  }

  @override
  AsyncResult<Unit> clearLocalSession() => throw UnimplementedError();

  @override
  AsyncResult<Unit> login(LoginCredentials credentials) =>
      throw UnimplementedError();

  @override
  AsyncResult<Unit> logout() => throw UnimplementedError();
}

class _FakeOnboardingDraftRepository implements OnboardingDraftRepository {
  Result<OnboardingDraftLoad> result = const Success(
    OnboardingDraftLoad.absent(),
  );
  int loadCalls = 0;

  @override
  AsyncResult<OnboardingDraftLoad> load() async {
    loadCalls++;
    return result;
  }

  @override
  AsyncResult<Unit> delete() => throw UnimplementedError();

  @override
  AsyncResult<Unit> save(OnboardingRegistrationDraft draft) =>
      throw UnimplementedError();
}
