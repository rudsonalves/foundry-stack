import '/core/extensions/string.dart';
import '/core/result/result.dart';
import '/data/repositories/email_verification/email_verification_repository.dart';
import '/data/repositories/onboarding/onboarding_draft_repository.dart';
import '/data/repositories/user/user_repository.dart';
import '../../common/onboarding/models/onboarding_registration_draft.dart';
import '../../common/onboarding/models/onboarding_step.dart';
import '../../common/users/models/user_registration.dart';

class OnboardingUsecase {
  final EmailVerificationRepository _emailVerificationRepository;
  final OnboardingDraftRepository _draftRepository;
  final UserRepository _userRepository;
  final DateTime Function() _now;

  OnboardingUsecase({
    required EmailVerificationRepository emailVerificationRepository,
    required OnboardingDraftRepository draftRepository,
    required UserRepository userRepository,
  }) : _emailVerificationRepository = emailVerificationRepository,
       _draftRepository = draftRepository,
       _userRepository = userRepository,
       _now = DateTime.now;

  OnboardingUsecase.withNow({
    required EmailVerificationRepository emailVerificationRepository,
    required OnboardingDraftRepository draftRepository,
    required UserRepository userRepository,
    required DateTime Function() now,
  }) : _emailVerificationRepository = emailVerificationRepository,
       _draftRepository = draftRepository,
       _userRepository = userRepository,
       _now = now;

  OnboardingRegistrationDraft? _currentDraft;

  OnboardingRegistrationDraft? get currentDraft => _currentDraft;

  static const AppError onboardingNotInitialized = AppError(
    code: .invalidData,
    message: 'Onboarding has not been initialized.',
  );

  AsyncResult<OnboardingRegistrationDraft> initialize() async {
    final result = await _draftRepository.load();

    if (result.isFailure) {
      return Failure(result.error!);
    }

    final draft =
        result.value!.draft ??
        OnboardingRegistrationDraft(
          step: OnboardingStep.userName,
          expiresAt: _now().toUtc(),
        );

    final normalized = _normalize(draft, _now().toUtc());
    return await _updateDraft(normalized);
  }

  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    final normalizedName = name.trim();

    if (draft.step != .userName) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Invalid onboarding step for name.',
        ),
      );
    }

    if (normalizedName.isEmpty) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Name is required.',
        ),
      );
    }

    return _updateDraft(
      draft.copyWith(
        step: .email,
        name: normalizedName,
      ),
    );
  }

  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    final OnboardingStep previousStep;

    switch (draft.step) {
      case OnboardingStep.email:
        previousStep = OnboardingStep.userName;
      case OnboardingStep.emailVerification:
      case OnboardingStep.password:
        previousStep = OnboardingStep.email;
      case OnboardingStep.userName:
        return const Failure(
          AppError(
            code: .invalidData,
            message: 'There is no previous onboarding step.',
          ),
        );
    }

    return _updateDraft(draft.copyWith(step: previousStep));
  }

  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(
    String email,
  ) async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    if (draft.step != .email) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Invalid onboarding step for email.',
        ),
      );
    }

    final normalizedEmail = email.trim();
    if (!normalizedEmail.isValidEmail) {
      return Failure(
        AppError(
          code: .invalidData,
          message: 'Invalid email.',
        ),
      );
    }

    final isSameEmail = normalizedEmail == draft.email;
    final now = _now().toUtc();

    if (isSameEmail) {
      final proof = draft.proof;

      if (proof != null && !proof.isExpiredAt(now)) {
        return _updateDraft(draft.copyWith(step: .password));
      }

      final challenge = draft.challenge;

      if (challenge != null && !challenge.isCodeExpiredAt(now)) {
        return _updateDraft(draft.copyWith(step: .emailVerification));
      }
    }

    var requestDraft = draft;

    if (!isSameEmail || draft.challenge != null || draft.proof != null) {
      final updateResult = await _updateDraft(
        draft.copyWith(
          step: .email,
          email: normalizedEmail,
          clearChallenge: true,
          clearProof: true,
        ),
      );

      if (updateResult.isFailure) {
        return Failure(updateResult.error!);
      }

      requestDraft = updateResult.value!;
    }

    final result = await _emailVerificationRepository.requestVerification(
      normalizedEmail,
    );

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return _updateDraft(
      requestDraft.copyWith(
        step: .emailVerification,
        challenge: result.value!,
        clearProof: true,
      ),
    );
  }

  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    if (draft.step != .emailVerification || draft.challenge == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'Invalid onboarding step for email verification.',
        ),
      );
    }

    final normalizedOtp = otp.trim();

    if (!RegExp(r'^\d{6}$').hasMatch(normalizedOtp)) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'OTP must contain exactly six digits.',
        ),
      );
    }

    if (draft.challenge!.isCodeExpiredAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Email verification challenge has expired.',
        ),
      );
    }

    final result = await _emailVerificationRepository.confirmVerification(
      verificationId: draft.challenge!.verificationId,
      code: normalizedOtp,
    );

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return _updateDraft(
      draft.copyWith(
        step: .password,
        proof: result.value!,
      ),
    );
  }

  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    final email = draft.email;
    final challenge = draft.challenge;

    if (draft.step != .emailVerification ||
        email == null ||
        challenge == null) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Email verification challenge is not available.',
        ),
      );
    }

    if (!challenge.canResendAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Email verification resend is not available yet.',
        ),
      );
    }

    final result = await _emailVerificationRepository.requestVerification(
      email,
    );

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return _updateDraft(
      draft.copyWith(
        step: .emailVerification,
        challenge: result.value!,
        clearProof: true,
      ),
    );
  }

  AsyncResult<Unit> cancelRegistration() async {
    final result = await _draftRepository.delete();

    if (result.isFailure) {
      return Failure(result.error!);
    }

    _currentDraft = null;

    return const Success(unit);
  }

  AsyncResult<Unit> createAccount(String password) async {
    final draft = _currentDraft;

    if (draft == null) {
      return Failure(onboardingNotInitialized);
    }

    final name = draft.name;
    final email = draft.email;
    final proof = draft.proof;

    if (draft.step != .password ||
        name == null ||
        name.trim().isEmpty ||
        email == null ||
        email.trim().isEmpty ||
        proof == null ||
        proof.isExpiredAt(_now().toUtc())) {
      return const Failure(
        AppError(
          code: .invalidData,
          message: 'Onboarding is not ready for account creation.',
        ),
      );
    }

    final registration = UserRegistration(
      name: name.trim(),
      email: email.trim(),
      password: password,
      emailVerificationToken: proof.token,
    );

    final createResult = await _userRepository.createUser(registration);

    if (createResult.isFailure) {
      final error = createResult.error!;
      final backendCode = backendErrorCode(error);

      if (backendCode == BackendErrorCodes.emailAlreadyRegistered) {
        await _updateDraft(
          draft.copyWith(
            step: .email,
            clearChallenge: true,
            clearProof: true,
          ),
        );
      }

      if (backendCode == BackendErrorCodes.invalidEmailVerification) {
        final clearVerificationResult = await _updateDraft(
          draft.copyWith(
            step: .email,
            clearChallenge: true,
            clearProof: true,
          ),
        );

        if (clearVerificationResult.isFailure) {
          return Failure(clearVerificationResult.error!);
        }

        final clearedDraft = clearVerificationResult.value!;

        final renewalResult = await _emailVerificationRepository
            .requestVerification(registration.email);

        if (renewalResult.isFailure) {
          return Failure(renewalResult.error!);
        }

        final challengeResult = await _updateDraft(
          clearedDraft.copyWith(
            step: .emailVerification,
            challenge: renewalResult.value!,
          ),
        );

        if (challengeResult.isFailure) {
          return Failure(challengeResult.error!);
        }
      }

      return Failure(error);
    }

    _currentDraft = null;

    await _draftRepository.delete();

    return const Success(unit);
  }

  // --------------------------------------------
  // Normalizes the onboarding draft by checking if the proof has expired and
  // updating the step accordingly.
  OnboardingRegistrationDraft _normalize(
    OnboardingRegistrationDraft draft,
    DateTime now,
  ) {
    final proof = draft.proof;

    if (proof == null || !proof.isExpiredAt(now)) {
      return draft;
    }

    final step = draft.step == .password
        ? draft.challenge == null
              ? OnboardingStep.email
              : OnboardingStep.emailVerification
        : draft.step;

    return draft.copyWith(
      step: step,
      clearProof: true,
    );
  }

  // Updates the onboarding draft in the repository and optionally renews
  // its validity.
  AsyncResult<OnboardingRegistrationDraft> _updateDraft(
    OnboardingRegistrationDraft draft, {
    bool renewValidity = true,
  }) async {
    final updated = renewValidity
        ? draft.copyWith(
            expiresAt: _now().toUtc().add(
              OnboardingRegistrationDraft.validity,
            ),
          )
        : draft;

    final result = await _draftRepository.save(updated);

    if (result.isFailure) {
      return Failure(result.error!);
    }

    _currentDraft = updated;
    return Success(updated);
  }
}
