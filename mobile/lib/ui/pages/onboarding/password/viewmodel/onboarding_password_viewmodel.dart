import 'package:material_ui/material_ui.dart';

import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/validators/validators.dart';

class OnboardingPasswordViewmodel {
  final OnboardingUsecase _usecase;

  OnboardingPasswordViewmodel(this._usecase) {
    createAccountCommand = Command1(_usecase.createAccount);
    returnToPreviousStepCommand = Command0(_usecase.returnToPreviousStep);
  }

  late final Command1<Unit, String> createAccountCommand;
  late final Command0<OnboardingRegistrationDraft> returnToPreviousStepCommand;

  OnboardingStep? get step => _usecase.currentDraft?.step;
  String? get name => _usecase.currentDraft?.name;
  String? get email => _usecase.currentDraft?.email;
  EmailVerificationChallenge? get challenge => _usecase.currentDraft?.challenge;
  EmailVerificationProof? get proof => _usecase.currentDraft?.proof;

  final isEnabled = ValueNotifier<bool>(false);
  final requirements = ValueNotifier<PasswordRequirements>(
    const PasswordRequirements(),
  );

  void dispose() {
    createAccountCommand.dispose();
    returnToPreviousStepCommand.dispose();

    isEnabled.dispose();
    requirements.dispose();
  }

  void updatePassword(String password, String confirmation) {
    requirements.value = PasswordRequirements(
      hasMinimumLength: password.length >= 8,
      hasLetter: password.hasUppercase || password.hasLowercase,
      hasNumber: password.hasDigit,
      passwordsMatch: confirmation.isNotEmpty && confirmation == password,
    );

    isEnabled.value =
        Validators.password(password) == null &&
        Validators.confirmPassword(confirmation, password) == null;
  }
}

class PasswordRequirements {
  final bool hasMinimumLength;
  final bool hasLetter;
  final bool hasNumber;
  final bool passwordsMatch;

  const PasswordRequirements({
    this.hasMinimumLength = false,
    this.hasLetter = false,
    this.hasNumber = false,
    this.passwordsMatch = false,
  });
}
