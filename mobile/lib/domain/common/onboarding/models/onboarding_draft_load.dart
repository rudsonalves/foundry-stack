import 'onboarding_registration_draft.dart';

class OnboardingDraftLoad {
  final OnboardingRegistrationDraft? draft;

  const OnboardingDraftLoad.absent() : draft = null;

  const OnboardingDraftLoad.present(OnboardingRegistrationDraft value)
    : draft = value;

  bool get hasDraft => draft != null;
}
