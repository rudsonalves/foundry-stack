import 'package:flutter/foundation.dart';

import '/core/result/command.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/validators/validators.dart';

class OnboardingEmailViewmodel {
  final OnboardingUsecase _usecase;

  OnboardingEmailViewmodel(this._usecase) {
    advanceEmailCommand = Command1(_usecase.advanceWithEmail);
    returnToPreviousStepCommand = Command0(_usecase.returnToPreviousStep);
    cancelRegistrationCommand = Command0(_usecase.cancelRegistration);

    updateEmail(email ?? '');
  }

  late final Command1<OnboardingRegistrationDraft, String> advanceEmailCommand;
  late final Command0<OnboardingRegistrationDraft> returnToPreviousStepCommand;
  late final Command0<Unit> cancelRegistrationCommand;

  OnboardingStep? get step => _usecase.currentDraft?.step;
  String? get name => _usecase.currentDraft?.name;
  String? get email => _usecase.currentDraft?.email;
  EmailVerificationChallenge? get challenge => _usecase.currentDraft?.challenge;
  EmailVerificationProof? get proof => _usecase.currentDraft?.proof;

  final isEnabled = ValueNotifier<bool>(false);
  final emailAlreadyRegistered = ValueNotifier<bool>(false);

  void dispose() {
    advanceEmailCommand.dispose();
    returnToPreviousStepCommand.dispose();
    cancelRegistrationCommand.dispose();

    isEnabled.dispose();
    emailAlreadyRegistered.dispose();
  }

  void updateEmail(String email) {
    isEnabled.value = Validators.email(email.trim()) == null;
    emailAlreadyRegistered.value = false;
  }

  void markEmailAsRegistered() {
    emailAlreadyRegistered.value = true;
  }
}
