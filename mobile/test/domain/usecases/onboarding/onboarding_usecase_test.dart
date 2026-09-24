import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/email_verification/email_verification_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository.dart';
import 'package:foundry_stack_mobile/data/repositories/user/user_repository.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_proof.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_draft_load.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/common/users/models/user_registration.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final now = DateTime.utc(2026, 9, 18, 12);
  late DateTime clock;
  late List<String> events;
  late _FakeOnboardingDraftRepository draftRepository;
  late _FakeEmailVerificationRepository emailVerificationRepository;
  late _FakeUserRepository userRepository;
  late OnboardingUsecase usecase;

  setUp(() {
    clock = now;
    events = [];
    draftRepository = _FakeOnboardingDraftRepository(events);
    emailVerificationRepository = _FakeEmailVerificationRepository(now, events);
    userRepository = _FakeUserRepository(events);
    usecase = OnboardingUsecase.withNow(
      emailVerificationRepository: emailVerificationRepository,
      draftRepository: draftRepository,
      userRepository: userRepository,
      now: () => clock,
    );
  });

  Future<void> initializeWith(OnboardingRegistrationDraft draft) async {
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));
    final result = await usecase.initialize();
    expect(result.isSuccess, isTrue);
    draftRepository.savedDrafts.clear();
    events.clear();
  }

  test('creates and persists an empty draft when none exists', () async {
    final result = await usecase.initialize();

    expect(result.isSuccess, isTrue);
    expect(result.value?.step, OnboardingStep.userName);
    expect(result.value?.expiresAt, now.add(const Duration(hours: 24)));
    expect(usecase.currentDraft, same(result.value));
    expect(draftRepository.savedDrafts, [same(result.value)]);
  });

  test('restores a valid draft and renews its expiration', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.email,
      name: 'Ada',
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.isSuccess, isTrue);
    expect(result.value?.step, OnboardingStep.email);
    expect(result.value?.name, 'Ada');
    expect(result.value?.expiresAt, now.add(const Duration(hours: 24)));
    expect(usecase.currentDraft, same(result.value));
    expect(draftRepository.savedDrafts, [same(result.value)]);
  });

  test('preserves a load failure without saving', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'read failed',
    );
    draftRepository.loadResult = const Failure(error);

    final result = await usecase.initialize();

    expect(identical(result.error, error), isTrue);
    expect(usecase.currentDraft, isNull);
    expect(draftRepository.savedDrafts, isEmpty);
  });

  test('preserves a save failure during initialization', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'write failed',
    );
    draftRepository.saveResult = const Failure(error);

    final result = await usecase.initialize();

    expect(identical(result.error, error), isTrue);
    expect(usecase.currentDraft, isNull);
    expect(draftRepository.savedDrafts, hasLength(1));
  });

  test('removes an expired proof and returns to verification', () async {
    final challenge = EmailVerificationChallenge(
      verificationId: 'verification-id',
      codeExpiresAt: now.subtract(const Duration(minutes: 1)),
      resendAvailableAt: now.subtract(const Duration(minutes: 2)),
    );
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.password,
      name: 'Ada',
      email: 'ada@example.com',
      challenge: challenge,
      proof: EmailVerificationProof(token: 'proof', expiresAt: now),
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.isSuccess, isTrue);
    expect(result.value?.step, OnboardingStep.emailVerification);
    expect(result.value?.proof, isNull);
    expect(identical(result.value?.challenge, challenge), isTrue);
  });

  test('returns to email when an expired proof has no challenge', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.password,
      name: 'Ada',
      email: 'ada@example.com',
      proof: EmailVerificationProof(token: 'proof', expiresAt: now),
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.isSuccess, isTrue);
    expect(result.value?.step, OnboardingStep.email);
    expect(result.value?.proof, isNull);
  });

  test('preserves a valid proof at the password step', () async {
    final proof = EmailVerificationProof(
      token: 'proof',
      expiresAt: now.add(const Duration(minutes: 1)),
    );
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.password,
      name: 'Ada',
      email: 'ada@example.com',
      proof: proof,
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.value?.step, OnboardingStep.password);
    expect(identical(result.value?.proof, proof), isTrue);
  });

  test('preserves an expired challenge for explicit resend', () async {
    final challenge = EmailVerificationChallenge(
      verificationId: 'verification-id',
      codeExpiresAt: now,
      resendAvailableAt: now,
    );
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.emailVerification,
      name: 'Ada',
      email: 'ada@example.com',
      challenge: challenge,
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));

    final result = await usecase.initialize();

    expect(result.value?.step, OnboardingStep.emailVerification);
    expect(identical(result.value?.challenge, challenge), isTrue);
  });

  test('trims the name, advances to email and persists the draft', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.userName,
      expiresAt: now.add(const Duration(hours: 1)),
    );

    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));
    await usecase.initialize();
    draftRepository.savedDrafts.clear();

    final result = await usecase.advanceWithName('  Ada Lovelace  ');

    expect(result.isSuccess, isTrue);
    expect(result.value?.step, OnboardingStep.email);
    expect(result.value?.name, 'Ada Lovelace');
    expect(result.value?.expiresAt, now.add(const Duration(hours: 24)));
    expect(usecase.currentDraft, same(result.value));
    expect(draftRepository.savedDrafts, [same(result.value)]);
  });

  test('rejects a blank name without saving', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.userName,
      expiresAt: now.add(const Duration(hours: 1)),
    );

    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));
    await usecase.initialize();
    draftRepository.savedDrafts.clear();

    final result = await usecase.advanceWithName('   ');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(draftRepository.savedDrafts, isEmpty);
  });

  test('rejects a name outside the user name step without saving', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.email,
      name: 'Ada',
      expiresAt: now.add(const Duration(hours: 1)),
    );

    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));
    await usecase.initialize();
    draftRepository.savedDrafts.clear();

    final result = await usecase.advanceWithName('Grace');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(draftRepository.savedDrafts, isEmpty);
  });

  test('preserves a save failure when advancing with the name', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'write failed',
    );
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.userName,
      expiresAt: now.add(const Duration(hours: 1)),
    );
    draftRepository.loadResult = Success(OnboardingDraftLoad.present(draft));
    await usecase.initialize();
    final persistedDraft = usecase.currentDraft;
    draftRepository.saveResult = const Failure(error);

    final result = await usecase.advanceWithName('Ada');

    expect(identical(result.error, error), isTrue);
    expect(usecase.currentDraft, same(persistedDraft));
  });

  test('rejects advancing before initialization', () async {
    final result = await usecase.advanceWithName('Ada');

    expect(result.error?.code, AppErrorCode.invalidData);
    expect(usecase.currentDraft, isNull);
    expect(draftRepository.savedDrafts, isEmpty);
  });

  group('return to previous step', () {
    final challenge = EmailVerificationChallenge(
      verificationId: 'verification-id',
      codeExpiresAt: now.add(const Duration(hours: 3)),
      resendAvailableAt: now,
    );
    final proof = EmailVerificationProof(
      token: 'proof-token',
      expiresAt: now.add(const Duration(hours: 3)),
    );

    for (final transition in <({OnboardingStep from, OnboardingStep to})>[
      (from: OnboardingStep.email, to: OnboardingStep.userName),
      (from: OnboardingStep.emailVerification, to: OnboardingStep.email),
      (from: OnboardingStep.password, to: OnboardingStep.email),
    ]) {
      test(
        'persists ${transition.from.name} to ${transition.to.name} '
        'and preserves the draft',
        () async {
          await initializeWith(
            OnboardingRegistrationDraft(
              step: transition.from,
              name: 'Ada',
              email: 'ada@example.com',
              challenge: challenge,
              proof: proof,
              expiresAt: now.add(const Duration(hours: 1)),
            ),
          );
          clock = now.add(const Duration(hours: 2));

          final result = await usecase.returnToPreviousStep();

          expect(result.isSuccess, isTrue);
          expect(result.value?.step, transition.to);
          expect(result.value?.name, 'Ada');
          expect(result.value?.email, 'ada@example.com');
          expect(result.value?.challenge, same(challenge));
          expect(result.value?.proof, same(proof));
          expect(
            result.value?.expiresAt,
            clock.add(OnboardingRegistrationDraft.validity),
          );
          expect(draftRepository.savedDrafts, [same(result.value)]);
          expect(usecase.currentDraft, same(result.value));
        },
      );
    }

    test('rejects returning from the first step without saving', () async {
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.userName,
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );

      final result = await usecase.returnToPreviousStep();

      expect(result.error?.code, AppErrorCode.invalidData);
      expect(usecase.currentDraft?.step, OnboardingStep.userName);
      expect(draftRepository.savedDrafts, isEmpty);
    });

    test('preserves the current step when persistence fails', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'write failed',
      );
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.emailVerification,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: challenge,
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );
      final persistedDraft = usecase.currentDraft;
      draftRepository.saveResult = const Failure(error);

      final result = await usecase.returnToPreviousStep();

      expect(result.error, same(error));
      expect(usecase.currentDraft, same(persistedDraft));
    });
  });

  group('email verification', () {
    test('requests and persists a challenge for a valid email', () async {
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );
      clock = now.add(const Duration(hours: 2));

      final result = await usecase.advanceWithEmail('  ada@example.com  ');

      expect(result.isSuccess, isTrue);
      expect(emailVerificationRepository.requestedEmails, ['ada@example.com']);
      expect(result.value?.step, OnboardingStep.emailVerification);
      expect(result.value?.email, 'ada@example.com');
      expect(
        result.value?.challenge,
        same(emailVerificationRepository.challenge),
      );
      expect(result.value?.expiresAt, clock.add(const Duration(hours: 24)));
      expect(draftRepository.savedDrafts, hasLength(2));
      expect(draftRepository.savedDrafts.first.step, OnboardingStep.email);
      expect(draftRepository.savedDrafts.first.challenge, isNull);
      expect(draftRepository.savedDrafts.last, same(result.value));
    });

    test(
      'reuses a valid proof for the same email without requesting',
      () async {
        final proof = EmailVerificationProof(
          token: 'proof',
          expiresAt: now.add(const Duration(hours: 1)),
        );
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.email,
            name: 'Ada',
            email: 'ada@example.com',
            proof: proof,
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final result = await usecase.advanceWithEmail('ada@example.com');

        expect(result.value?.step, OnboardingStep.password);
        expect(result.value?.proof, same(proof));
        expect(emailVerificationRepository.requestedEmails, isEmpty);
      },
    );

    test(
      'reuses a valid challenge for the same email without requesting',
      () async {
        final challenge = EmailVerificationChallenge(
          verificationId: 'existing',
          codeExpiresAt: now.add(const Duration(minutes: 5)),
          resendAvailableAt: now.add(const Duration(minutes: 1)),
        );
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.email,
            name: 'Ada',
            email: 'ada@example.com',
            challenge: challenge,
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final result = await usecase.advanceWithEmail('ada@example.com');

        expect(result.value?.step, OnboardingStep.emailVerification);
        expect(result.value?.challenge, same(challenge));
        expect(emailVerificationRepository.requestedEmails, isEmpty);
      },
    );

    test('renews validity after requesting for the unchanged email', () async {
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          email: 'ada@example.com',
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );
      clock = now.add(const Duration(hours: 2));

      final result = await usecase.advanceWithEmail('ada@example.com');

      expect(result.value?.expiresAt, clock.add(const Duration(hours: 24)));
    });

    test(
      'keeps the changed email and semantic error when request fails',
      () async {
        const error = AppError(
          statusCode: 409,
          code: AppErrorCode.conflict,
          message: 'E-mail em uso',
          details: {'code': BackendErrorCodes.emailAlreadyRegistered},
        );
        emailVerificationRepository.requestResult = const Failure(error);
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.email,
            name: 'Ada',
            email: 'old@example.com',
            challenge: EmailVerificationChallenge(
              verificationId: 'old',
              codeExpiresAt: now.add(const Duration(minutes: 5)),
              resendAvailableAt: now,
            ),
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final result = await usecase.advanceWithEmail('new@example.com');

        expect(result.error, same(error));
        expect(
          backendErrorCode(result.error),
          BackendErrorCodes.emailAlreadyRegistered,
        );
        expect(usecase.currentDraft?.step, OnboardingStep.email);
        expect(usecase.currentDraft?.email, 'new@example.com');
        expect(usecase.currentDraft?.challenge, isNull);
        expect(usecase.currentDraft?.proof, isNull);
      },
    );

    test('preserves network and delivery errors from a request', () async {
      for (final error in <AppError>[
        const AppError(code: AppErrorCode.networkError, message: 'offline'),
        const AppError(
          code: AppErrorCode.unexpected,
          message: 'delivery unavailable',
          details: {'code': BackendErrorCodes.emailDeliveryUnavailable},
        ),
      ]) {
        emailVerificationRepository.requestResult = Failure(error);
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.email,
            name: 'Ada',
            email: 'ada@example.com',
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final result = await usecase.advanceWithEmail('ada@example.com');

        expect(result.error, same(error));
      }
    });

    test('confirms six digits and persists the proof', () async {
      final challenge = EmailVerificationChallenge(
        verificationId: 'verification-id',
        codeExpiresAt: now.add(const Duration(minutes: 5)),
        resendAvailableAt: now,
      );
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.emailVerification,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: challenge,
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );

      final result = await usecase.confirmEmailVerification(' 012345 ');

      expect(emailVerificationRepository.confirmations, [
        (verificationId: 'verification-id', code: '012345'),
      ]);
      expect(result.value?.step, OnboardingStep.password);
      expect(result.value?.proof, same(emailVerificationRepository.proof));
    });

    test('rejects malformed OTP and expired challenge locally', () async {
      final challenge = EmailVerificationChallenge(
        verificationId: 'verification-id',
        codeExpiresAt: now,
        resendAvailableAt: now,
      );
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.emailVerification,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: challenge,
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );

      final malformed = await usecase.confirmEmailVerification('12345');
      final expired = await usecase.confirmEmailVerification('123456');

      expect(malformed.error?.code, AppErrorCode.invalidData);
      expect(expired.error?.code, AppErrorCode.invalidData);
      expect(emailVerificationRepository.confirmations, isEmpty);
      expect(draftRepository.savedDrafts, isEmpty);
    });

    test('preserves INVALID_EMAIL_VERIFICATION from confirmation', () async {
      const error = AppError(
        code: AppErrorCode.invalidData,
        message: 'invalid verification',
        details: {'code': BackendErrorCodes.invalidEmailVerification},
      );
      emailVerificationRepository.confirmResult = const Failure(error);
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.emailVerification,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: EmailVerificationChallenge(
            verificationId: 'verification-id',
            codeExpiresAt: now.add(const Duration(minutes: 5)),
            resendAvailableAt: now,
          ),
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );

      final result = await usecase.confirmEmailVerification('123456');

      expect(result.error, same(error));
      expect(
        backendErrorCode(result.error),
        BackendErrorCodes.invalidEmailVerification,
      );
      expect(draftRepository.savedDrafts, isEmpty);
    });

    test(
      'rejects resend during cooldown without request or persistence',
      () async {
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.emailVerification,
            name: 'Ada',
            email: 'ada@example.com',
            challenge: EmailVerificationChallenge(
              verificationId: 'verification-id',
              codeExpiresAt: now.add(const Duration(minutes: 5)),
              resendAvailableAt: now.add(const Duration(seconds: 1)),
            ),
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final result = await usecase.resendEmailVerification();

        expect(result.error?.code, AppErrorCode.invalidData);
        expect(emailVerificationRepository.requestedEmails, isEmpty);
        expect(draftRepository.savedDrafts, isEmpty);
      },
    );

    test('resends, replaces the challenge and renews validity', () async {
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.emailVerification,
          name: 'Ada',
          email: 'ada@example.com',
          challenge: EmailVerificationChallenge(
            verificationId: 'old',
            codeExpiresAt: now,
            resendAvailableAt: now,
          ),
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );
      clock = now.add(const Duration(hours: 2));

      final result = await usecase.resendEmailVerification();

      expect(emailVerificationRepository.requestedEmails, ['ada@example.com']);
      expect(draftRepository.savedDrafts, hasLength(1));
      expect(draftRepository.savedDrafts.single, same(result.value));
      expect(result.value?.step, OnboardingStep.emailVerification);
      expect(
        result.value?.challenge,
        same(emailVerificationRepository.challenge),
      );
      expect(result.value?.expiresAt, clock.add(const Duration(hours: 24)));
    });

    test(
      'preserves draft and RATE_LIMIT_EXCEEDED when resend fails',
      () async {
        const error = AppError(
          code: AppErrorCode.conflict,
          message: 'rate limited',
          details: {'code': BackendErrorCodes.rateLimitExceeded},
        );
        emailVerificationRepository.requestResult = const Failure(error);
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.emailVerification,
            name: 'Ada',
            email: 'ada@example.com',
            challenge: EmailVerificationChallenge(
              verificationId: 'old',
              codeExpiresAt: now,
              resendAvailableAt: now,
            ),
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );
        final draftBeforeResend = usecase.currentDraft;

        final result = await usecase.resendEmailVerification();

        expect(result.error, same(error));
        expect(
          backendErrorCode(result.error),
          BackendErrorCodes.rateLimitExceeded,
        );
        expect(usecase.currentDraft, same(draftBeforeResend));
        expect(usecase.currentDraft?.step, OnboardingStep.emailVerification);
        expect(usecase.currentDraft?.challenge?.verificationId, 'old');
        expect(draftRepository.savedDrafts, isEmpty);
      },
    );
  });

  group('registration cancellation', () {
    test('deletes the draft and clears the current state', () async {
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          email: 'used@example.com',
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );

      final result = await usecase.cancelRegistration();

      expect(result.isSuccess, isTrue);
      expect(draftRepository.deleteCalls, 1);
      expect(usecase.currentDraft, isNull);
    });

    test('preserves the current state when deletion fails', () async {
      const error = AppError(
        code: AppErrorCode.storageError,
        message: 'delete failed',
      );
      await initializeWith(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          email: 'used@example.com',
          expiresAt: now.add(const Duration(hours: 1)),
        ),
      );
      final draft = usecase.currentDraft;
      draftRepository.deleteResult = const Failure(error);

      final result = await usecase.cancelRegistration();

      expect(result.error, same(error));
      expect(draftRepository.deleteCalls, 1);
      expect(usecase.currentDraft, same(draft));
    });
  });

  group('account creation', () {
    OnboardingRegistrationDraft readyDraft() {
      return OnboardingRegistrationDraft(
        step: OnboardingStep.password,
        name: '  Ada Lovelace  ',
        email: '  ada@example.com  ',
        challenge: EmailVerificationChallenge(
          verificationId: 'verification-id',
          codeExpiresAt: now.add(const Duration(minutes: 5)),
          resendAvailableAt: now,
        ),
        proof: EmailVerificationProof(
          token: 'proof-token',
          expiresAt: now.add(const Duration(hours: 1)),
        ),
        expiresAt: now.add(const Duration(hours: 1)),
      );
    }

    test(
      'creates with normalized data and deletes the draft in order',
      () async {
        await initializeWith(readyDraft());

        final result = await usecase.createAccount('Secret123');

        expect(result.isSuccess, isTrue);
        expect(userRepository.registrations, hasLength(1));
        final registration = userRepository.registrations.single;
        expect(registration.name, 'Ada Lovelace');
        expect(registration.email, 'ada@example.com');
        expect(registration.password, 'Secret123');
        expect(registration.emailVerificationToken, 'proof-token');
        expect(events, ['user.create', 'draft.delete']);
        expect(draftRepository.deleteCalls, 1);
        expect(usecase.currentDraft, isNull);
      },
    );

    test(
      'keeps account creation successful when draft deletion fails',
      () async {
        const cleanupError = AppError(
          code: AppErrorCode.storageError,
          message: 'delete failed',
        );
        draftRepository.deleteResult = const Failure(cleanupError);
        await initializeWith(readyDraft());

        final result = await usecase.createAccount('Secret123');

        expect(result.isSuccess, isTrue);
        expect(draftRepository.deleteCalls, 1);
        expect(usecase.currentDraft, isNull);
        expect(events, ['user.create', 'draft.delete']);
      },
    );

    test(
      'rejects incomplete or expired registration without creating',
      () async {
        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.password,
            name: 'Ada',
            email: 'ada@example.com',
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final missingProof = await usecase.createAccount('Secret123');

        expect(missingProof.error?.code, AppErrorCode.invalidData);
        expect(userRepository.registrations, isEmpty);
        expect(events, isEmpty);

        await initializeWith(
          OnboardingRegistrationDraft(
            step: OnboardingStep.password,
            name: 'Ada',
            email: 'ada@example.com',
            proof: EmailVerificationProof(token: 'proof', expiresAt: now),
            expiresAt: now.add(const Duration(hours: 1)),
          ),
        );

        final expiredProof = await usecase.createAccount('Secret123');

        expect(expiredProof.error?.code, AppErrorCode.invalidData);
        expect(userRepository.registrations, isEmpty);
        expect(events, isEmpty);
      },
    );

    test(
      'returns to email, clears verification and preserves EMAIL_ALREADY_REGISTERED',
      () async {
        const error = AppError(
          statusCode: 409,
          code: AppErrorCode.conflict,
          message: 'E-mail em uso',
          details: {'code': BackendErrorCodes.emailAlreadyRegistered},
        );
        userRepository.createResult = const Failure(error);
        await initializeWith(readyDraft());

        final result = await usecase.createAccount('Secret123');

        expect(result.error, same(error));
        expect(usecase.currentDraft?.step, OnboardingStep.email);
        expect(usecase.currentDraft?.challenge, isNull);
        expect(usecase.currentDraft?.proof, isNull);
        expect(events, ['user.create', 'draft.save:email']);
        expect(draftRepository.deleteCalls, 0);
      },
    );

    test('renews verification after INVALID_EMAIL_VERIFICATION', () async {
      const error = AppError(
        code: AppErrorCode.invalidData,
        message: 'invalid verification',
        details: {'code': BackendErrorCodes.invalidEmailVerification},
      );
      userRepository.createResult = const Failure(error);
      await initializeWith(readyDraft());

      final result = await usecase.createAccount('Secret123');

      expect(result.error, same(error));
      expect(emailVerificationRepository.requestedEmails, ['ada@example.com']);
      expect(usecase.currentDraft?.step, OnboardingStep.emailVerification);
      expect(
        usecase.currentDraft?.challenge,
        same(emailVerificationRepository.challenge),
      );
      expect(usecase.currentDraft?.proof, isNull);
      expect(events, [
        'user.create',
        'draft.save:email',
        'email.request',
        'draft.save:emailVerification',
      ]);
    });

    test('stays at email when verification renewal request fails', () async {
      const creationError = AppError(
        code: AppErrorCode.invalidData,
        message: 'invalid verification',
        details: {'code': BackendErrorCodes.invalidEmailVerification},
      );
      const renewalError = AppError(
        code: AppErrorCode.networkError,
        message: 'offline',
      );
      userRepository.createResult = const Failure(creationError);
      emailVerificationRepository.requestResult = const Failure(renewalError);
      await initializeWith(readyDraft());

      final result = await usecase.createAccount('Secret123');

      expect(result.error, same(renewalError));
      expect(usecase.currentDraft?.step, OnboardingStep.email);
      expect(usecase.currentDraft?.challenge, isNull);
      expect(usecase.currentDraft?.proof, isNull);
      expect(events, ['user.create', 'draft.save:email', 'email.request']);
    });

    test('stays at email when persisting renewed challenge fails', () async {
      const creationError = AppError(
        code: AppErrorCode.invalidData,
        message: 'invalid verification',
        details: {'code': BackendErrorCodes.invalidEmailVerification},
      );
      const storageError = AppError(
        code: AppErrorCode.storageError,
        message: 'write failed',
      );
      userRepository.createResult = const Failure(creationError);
      await initializeWith(readyDraft());
      draftRepository.saveResults.addAll([
        const Success(unit),
        const Failure(storageError),
      ]);

      final result = await usecase.createAccount('Secret123');

      expect(result.error, same(storageError));
      expect(usecase.currentDraft?.step, OnboardingStep.email);
      expect(usecase.currentDraft?.challenge, isNull);
      expect(events, [
        'user.create',
        'draft.save:email',
        'email.request',
        'draft.save:emailVerification',
      ]);
    });

    test('preserves draft and error for other creation failures', () async {
      const error = AppError(
        code: AppErrorCode.networkError,
        message: 'offline',
      );
      userRepository.createResult = const Failure(error);
      await initializeWith(readyDraft());
      final draftBeforeCreation = usecase.currentDraft;

      final result = await usecase.createAccount('Secret123');

      expect(result.error, same(error));
      expect(usecase.currentDraft, same(draftBeforeCreation));
      expect(events, ['user.create']);
      expect(draftRepository.savedDrafts, isEmpty);
      expect(draftRepository.deleteCalls, 0);
    });
  });
}

class _FakeOnboardingDraftRepository implements OnboardingDraftRepository {
  _FakeOnboardingDraftRepository(this.events);

  final List<String> events;
  Result<OnboardingDraftLoad> loadResult = const Success(
    OnboardingDraftLoad.absent(),
  );
  Result<Unit> saveResult = const Success(unit);
  Result<Unit> deleteResult = const Success(unit);
  final saveResults = <Result<Unit>>[];
  final savedDrafts = <OnboardingRegistrationDraft>[];
  int deleteCalls = 0;

  @override
  AsyncResult<OnboardingDraftLoad> load() async => loadResult;

  @override
  AsyncResult<Unit> save(OnboardingRegistrationDraft draft) async {
    savedDrafts.add(draft);
    events.add('draft.save:${draft.step.name}');
    return saveResults.isEmpty ? saveResult : saveResults.removeAt(0);
  }

  @override
  AsyncResult<Unit> delete() async {
    deleteCalls++;
    events.add('draft.delete');
    return deleteResult;
  }
}

class _FakeEmailVerificationRepository implements EmailVerificationRepository {
  _FakeEmailVerificationRepository(DateTime now, this.events)
    : challenge = EmailVerificationChallenge(
        verificationId: 'new-verification-id',
        codeExpiresAt: now.add(const Duration(minutes: 10)),
        resendAvailableAt: now.add(const Duration(minutes: 1)),
      ),
      proof = EmailVerificationProof(
        token: 'proof-token',
        expiresAt: now.add(const Duration(hours: 1)),
      );

  final List<String> events;
  final EmailVerificationChallenge challenge;
  final EmailVerificationProof proof;
  late Result<EmailVerificationChallenge> requestResult = Success(challenge);
  late Result<EmailVerificationProof> confirmResult = Success(proof);
  final requestedEmails = <String>[];
  final confirmations = <({String verificationId, String code})>[];

  @override
  AsyncResult<EmailVerificationChallenge> requestVerification(
    String email,
  ) async {
    requestedEmails.add(email);
    events.add('email.request');
    return requestResult;
  }

  @override
  AsyncResult<EmailVerificationProof> confirmVerification({
    required String verificationId,
    required String code,
  }) async {
    confirmations.add((verificationId: verificationId, code: code));
    events.add('email.confirm');
    return confirmResult;
  }
}

class _FakeUserRepository implements UserRepository {
  _FakeUserRepository(this.events);

  final List<String> events;
  Result<Unit> createResult = const Success(unit);
  final registrations = <UserRegistration>[];

  @override
  AsyncResult<Unit> createUser(UserRegistration registration) async {
    registrations.add(registration);
    events.add('user.create');
    return createResult;
  }
}
