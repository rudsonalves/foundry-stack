import '/core/result/command.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';

class OnboardingCoordinatorViewmodel {
  final OnboardingUsecase _usecase;

  OnboardingCoordinatorViewmodel(this._usecase) {
    initializeCommand = Command0(_usecase.initialize);
  }

  late final Command0<OnboardingRegistrationDraft> initializeCommand;

  OnboardingStep? get step => _usecase.currentDraft?.step;

  void dispose() {
    initializeCommand.dispose();
  }
}
