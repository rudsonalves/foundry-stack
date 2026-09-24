import 'dart:async';

import 'package:flutter/foundation.dart';

import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/domain/common/onboarding/models/email_verification_challenge.dart';
import '/domain/common/onboarding/models/email_verification_proof.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import '/domain/common/onboarding/models/onboarding_step.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';

class OnboardingEmailVerificationViewmodel {
  final OnboardingUsecase _usecase;

  OnboardingEmailVerificationViewmodel(this._usecase) {
    confirmVerificationCommand = Command1(_usecase.confirmEmailVerification);
    resendVerificationCommand = Command0(_usecase.resendEmailVerification);
    returnToPreviousStepCommand = Command0(_usecase.returnToPreviousStep);

    resendRemaining = ValueNotifier(_calculateResendRemaining());
    _startResendTimer();
  }

  late final Command1<OnboardingRegistrationDraft, String>
  confirmVerificationCommand;
  late final Command0<OnboardingRegistrationDraft> resendVerificationCommand;
  late final Command0<OnboardingRegistrationDraft> returnToPreviousStepCommand;

  late final ValueNotifier<Duration> resendRemaining;
  Timer? _resendTimer;

  OnboardingStep? get step => _usecase.currentDraft?.step;
  String? get name => _usecase.currentDraft?.name;
  String? get email => _usecase.currentDraft?.email;
  EmailVerificationChallenge? get challenge => _usecase.currentDraft?.challenge;
  EmailVerificationProof? get proof => _usecase.currentDraft?.proof;

  final isEnabled = ValueNotifier<bool>(false);

  void dispose() {
    _resendTimer?.cancel();

    confirmVerificationCommand.dispose();
    resendVerificationCommand.dispose();
    returnToPreviousStepCommand.dispose();

    isEnabled.dispose();
    resendRemaining.dispose();
  }

  void updateEmailVerification(String token) {
    isEnabled.value = token.isNumbers && token.length == 6;
  }

  Duration _calculateResendRemaining() {
    return challenge?.resendRemainingAt(DateTime.now()) ?? Duration.zero;
  }

  void _startResendTimer() {
    _resendTimer?.cancel();

    resendRemaining.value = _calculateResendRemaining();

    if (resendRemaining.value == Duration.zero) return;

    _resendTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      final remaining = _calculateResendRemaining();

      resendRemaining.value = remaining;

      if (remaining == Duration.zero) {
        _resendTimer?.cancel();
        _resendTimer = null;
      }
    });
  }

  void restartResendCountdown() {
    _startResendTimer();
  }
}
