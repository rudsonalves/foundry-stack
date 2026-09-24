import 'package:foundry_stack_mobile/core/resources/storage_keys.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/email_verification/email_verification_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository_impl.dart';
import 'package:foundry_stack_mobile/data/repositories/user/user_repository.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_proof.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/common/users/models/user_registration.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/bootstrap_usecase.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/models/bootstrap_destination.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late DateTime clock;
  late _MemorySecureStorage storage;
  late _FakeEmailVerificationRepository emailRepository;
  late _FakeUserRepository userRepository;

  OnboardingDraftRepositoryImpl draftRepository() {
    return OnboardingDraftRepositoryImpl.withNow(
      storage: storage,
      now: () => clock,
    );
  }

  OnboardingUsecase journey() {
    return OnboardingUsecase.withNow(
      emailVerificationRepository: emailRepository,
      draftRepository: draftRepository(),
      userRepository: userRepository,
      now: () => clock,
    );
  }

  setUp(() {
    clock = DateTime.utc(2026, 9, 22, 12);
    storage = _MemorySecureStorage();
    emailRepository = _FakeEmailVerificationRepository(() => clock);
    userRepository = _FakeUserRepository();
  });

  test('completes, resumes and resends the registration journey', () async {
    final firstJourney = journey();

    expect(
      (await firstJourney.initialize()).value?.step,
      OnboardingStep.userName,
    );

    clock = clock.add(const Duration(minutes: 10));
    expect(
      (await firstJourney.advanceWithName('  Ada  ')).value?.step,
      OnboardingStep.email,
    );

    clock = clock.add(const Duration(minutes: 10));
    expect(
      (await firstJourney.advanceWithEmail('ada@example.com')).value?.step,
      OnboardingStep.emailVerification,
    );
    final firstChallenge = firstJourney.currentDraft!.challenge;

    expect(
      (await firstJourney.returnToPreviousStep()).value?.step,
      OnboardingStep.email,
    );
    expect(
      (await firstJourney.advanceWithEmail('ada@example.com')).value?.step,
      OnboardingStep.emailVerification,
    );
    expect(firstJourney.currentDraft!.challenge, same(firstChallenge));

    clock = clock.add(const Duration(minutes: 1));
    expect((await firstJourney.resendEmailVerification()).isSuccess, isTrue);
    expect(emailRepository.requestedEmails, hasLength(2));
    expect(
      firstJourney.currentDraft!.challenge!.verificationId,
      isNot(firstChallenge!.verificationId),
    );

    expect(
      (await firstJourney.confirmEmailVerification('012345')).value?.step,
      OnboardingStep.password,
    );

    final persistedBeforeClose =
        storage.values[StorageKeys.onboardingRegistrationDraft]!;
    expect(persistedBeforeClose, isNot(contains('012345')));
    expect(persistedBeforeClose, isNot(contains('Secret123')));
    expect(persistedBeforeClose, isNot(contains('confirmation')));
    expect(persistedBeforeClose, isNot(contains('loading')));
    expect(persistedBeforeClose, isNot(contains('message')));

    final expirationBeforeResume = firstJourney.currentDraft!.expiresAt;
    clock = clock.add(const Duration(minutes: 30));

    final resumedJourney = journey();
    final resumed = await resumedJourney.initialize();

    expect(resumed.value?.step, OnboardingStep.password);
    expect(resumed.value?.name, 'Ada');
    expect(resumed.value?.expiresAt, clock.add(const Duration(hours: 24)));
    expect(resumed.value!.expiresAt.isAfter(expirationBeforeResume), isTrue);

    final expirationBeforeFailure = resumedJourney.currentDraft!.expiresAt;
    userRepository.result = const Failure(
      AppError(code: AppErrorCode.networkError, message: 'offline'),
    );
    clock = clock.add(const Duration(minutes: 10));

    expect((await resumedJourney.createAccount('Secret123')).isFailure, isTrue);
    expect(resumedJourney.currentDraft?.expiresAt, expirationBeforeFailure);
    expect(
      storage.values[StorageKeys.onboardingRegistrationDraft],
      isNot(contains('Secret123')),
    );

    userRepository.result = const Success(unit);
    expect((await resumedJourney.createAccount('Secret123')).isSuccess, isTrue);
    expect(userRepository.registrations, hasLength(2));
    expect(userRepository.registrations.last.password, 'Secret123');
    expect(
      storage.values.containsKey(StorageKeys.onboardingRegistrationDraft),
      isFalse,
    );
  });

  test(
    'bootstrap reads drafts without renewing them and ignores expired ones',
    () async {
      final repository = draftRepository();
      final validDraft = OnboardingRegistrationDraft(
        step: OnboardingStep.email,
        name: 'Ada',
        expiresAt: clock.add(const Duration(hours: 2)),
      );
      await repository.save(validDraft);
      final persistedBeforeBootstrap =
          storage.values[StorageKeys.onboardingRegistrationDraft];

      final bootstrap = BootstrapUsecase(
        authRepository: _FakeAuthRepository(),
        onboardingDraftRepository: repository,
      );

      expect(
        (await bootstrap.initialize()).value,
        BootstrapDestination.onboarding,
      );
      expect(
        storage.values[StorageKeys.onboardingRegistrationDraft],
        persistedBeforeBootstrap,
      );

      await repository.save(
        validDraft.copyWith(
          expiresAt: clock.subtract(const Duration(seconds: 1)),
        ),
      );

      expect((await bootstrap.initialize()).value, BootstrapDestination.login);
      expect(
        storage.values.containsKey(StorageKeys.onboardingRegistrationDraft),
        isFalse,
      );

      await repository.save(validDraft);
      final authenticatedBootstrap = BootstrapUsecase(
        authRepository: _FakeAuthRepository(restored: true),
        onboardingDraftRepository: repository,
      );

      expect(
        (await authenticatedBootstrap.initialize()).value,
        BootstrapDestination.home,
      );
    },
  );

  test(
    'expired proof blocks creation and returns to email verification',
    () async {
      final repository = draftRepository();
      await repository.save(
        OnboardingRegistrationDraft(
          step: OnboardingStep.password,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: EmailVerificationChallenge(
            verificationId: 'verification-id',
            codeExpiresAt: clock.add(const Duration(minutes: 5)),
            resendAvailableAt: clock,
          ),
          proof: EmailVerificationProof(
            token: 'expired-proof',
            expiresAt: clock,
          ),
          expiresAt: clock.add(const Duration(hours: 1)),
        ),
      );

      final resumedJourney = journey();
      final resumed = await resumedJourney.initialize();

      expect(resumed.value?.step, OnboardingStep.emailVerification);
      expect(resumed.value?.proof, isNull);
      expect(
        (await resumedJourney.createAccount('Secret123')).isFailure,
        isTrue,
      );
      expect(userRepository.registrations, isEmpty);
    },
  );

  test('cancellation removes the draft without renewing it', () async {
    final currentJourney = journey();
    await currentJourney.initialize();
    await currentJourney.advanceWithName('Ada');
    final expirationBeforeCancellation = currentJourney.currentDraft!.expiresAt;
    clock = clock.add(const Duration(hours: 1));

    expect((await currentJourney.cancelRegistration()).isSuccess, isTrue);
    expect(currentJourney.currentDraft, isNull);
    expect(
      expirationBeforeCancellation,
      isNot(clock.add(const Duration(hours: 24))),
    );
    expect(
      storage.values.containsKey(StorageKeys.onboardingRegistrationDraft),
      isFalse,
    );
  });
}

class _MemorySecureStorage implements LocalSecureStorage {
  final values = <String, String>{};

  @override
  AsyncResult<Unit> write(String key, String value) async {
    values[key] = value;
    return const Success(unit);
  }

  @override
  AsyncResult<String> read(String key) async {
    final value = values[key];
    return value == null
        ? const Failure(
            AppError(
              code: AppErrorCode.storageNotFound,
              message: 'not found',
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
  AsyncResult<List<String>> keysWithPrefix(String pattern) async {
    return Success(
      values.keys.where((key) => key.startsWith(pattern)).toList(),
    );
  }
}

class _FakeEmailVerificationRepository implements EmailVerificationRepository {
  final DateTime Function() _now;
  final requestedEmails = <String>[];
  final confirmedCodes = <String>[];

  _FakeEmailVerificationRepository(this._now);

  @override
  AsyncResult<EmailVerificationChallenge> requestVerification(
    String email,
  ) async {
    requestedEmails.add(email);
    final sequence = requestedEmails.length;
    return Success(
      EmailVerificationChallenge(
        verificationId: 'verification-$sequence',
        codeExpiresAt: _now().add(const Duration(minutes: 10)),
        resendAvailableAt: _now().add(const Duration(minutes: 1)),
      ),
    );
  }

  @override
  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  }) async {
    confirmedCodes.add(code);
    return Success(
      EmailVerificationProof(
        token: 'opaque-proof',
        expiresAt: _now().add(const Duration(hours: 1)),
      ),
    );
  }
}

class _FakeUserRepository implements UserRepository {
  Result<Unit> result = const Success(unit);
  final registrations = <UserRegistration>[];

  @override
  AsyncResult<Unit> createUser(UserRegistration registration) async {
    registrations.add(registration);
    return result;
  }
}

class _FakeAuthRepository implements AuthRepository {
  final bool restored;

  _FakeAuthRepository({this.restored = false});

  @override
  AsyncResult<RestoreSessionStatus> restoreSession() async => Success(
    restored ? RestoreSessionStatus.restored : RestoreSessionStatus.noSession,
  );

  @override
  AsyncResult<Unit> clearLocalSession() => throw UnimplementedError();

  @override
  AsyncResult<Unit> login(LoginCredentials credentials) =>
      throw UnimplementedError();

  @override
  AsyncResult<Unit> logout() => throw UnimplementedError();
}
