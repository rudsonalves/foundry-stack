import 'package:flutter/foundation.dart';

import '/core/result/command.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/validators/validators.dart';

class OnboardingUserNameViewmodel {
  final OnboardingUsecase _usecase;

  OnboardingUserNameViewmodel(this._usecase) {
    advanceNameCommand = Command1(_usecase.advanceWithName);

    updateName(name ?? '');
  }

  late final Command1<OnboardingRegistrationDraft, String> advanceNameCommand;

  OnboardingStep? get step => _usecase.currentDraft?.step;
  String? get name => _usecase.currentDraft?.name;
  String? get email => _usecase.currentDraft?.email;
  EmailVerificationChallenge? get challenge => _usecase.currentDraft?.challenge;
  EmailVerificationProof? get proof => _usecase.currentDraft?.proof;

  final isEnabled = ValueNotifier<bool>(false);

  void dispose() {
    advanceNameCommand.dispose();

    isEnabled.dispose();
  }

  void updateName(String name) {
    isEnabled.value = Validators.notEmpty(name) == null;
  }
}
