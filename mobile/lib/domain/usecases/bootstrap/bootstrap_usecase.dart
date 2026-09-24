import '/core/result/command.dart';
import '/data/repositories/auth/auth_repository.dart';
import '/data/repositories/onboarding/onboarding_draft_repository.dart';
import 'models/bootstrap_destination.dart';

class BootstrapUsecase {
  final AuthRepository _authRepository;
  final OnboardingDraftRepository _onboardingDraftRepository;

  BootstrapUsecase({
    required AuthRepository authRepository,
    required OnboardingDraftRepository onboardingDraftRepository,
  }) : _authRepository = authRepository,
       _onboardingDraftRepository = onboardingDraftRepository;

  AsyncResult<BootstrapDestination> initialize() async {
    final sessionResult = await _authRepository.restoreSession();

    if (sessionResult.isFailure) {
      return Failure(sessionResult.error!);
    }

    if (sessionResult.value == RestoreSessionStatus.restored) {
      return const Success(BootstrapDestination.home);
    }

    final draftResult = await _onboardingDraftRepository.load();
    if (draftResult.isFailure) {
      return Failure(draftResult.error!);
    }

    return Success(
      draftResult.value!.draft == null
          ? BootstrapDestination.login
          : BootstrapDestination.onboarding,
    );
  }
}
