import 'dart:async';

import 'package:flutter/foundation.dart';

import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/domain/common/password_reset/models/password_reset_challenge.dart';
import '/domain/common/password_reset/models/password_reset_step.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';

class PasswordResetCodeViewmodel {
  final PasswordResetUsecase _usecase;
  final DateTime Function() _now;

  PasswordResetCodeViewmodel(
    this._usecase, {
    DateTime Function()? now,
  }) : _now = now ?? DateTime.now {
    confirmPasswordResetCommand = Command1(_usecase.confirmPasswordReset);
    resendPasswordResetCommand = Command0(_usecase.resendPasswordReset);
    returnToEmailCommand = Command0(_usecase.returnToEmail);
    resendRemaining = ValueNotifier(_calculateResendRemaining());

    _startResendTimer();
  }

  late final Command1<Unit, String> confirmPasswordResetCommand;
  late final Command0<Unit> resendPasswordResetCommand;
  late final Command0<Unit> returnToEmailCommand;

  final isEnabled = ValueNotifier<bool>(false);
  late final ValueNotifier<Duration> resendRemaining;
  Timer? _resendTimer;

  PasswordResetStep get step => _usecase.step;
  String? get email => _usecase.email;
  PasswordResetChallenge? get challenge => _usecase.challenge;
  bool get canResend => resendRemaining.value == Duration.zero;

  void updateCode(String code) {
    isEnabled.value = code.isNumbers && code.length == 6;
  }

  void restartResendCountdown() {
    _startResendTimer();
  }

  Duration _calculateResendRemaining() {
    return challenge?.resendRemainingAt(_now().toUtc()) ?? Duration.zero;
  }

  void _startResendTimer() {
    _resendTimer?.cancel();
    resendRemaining.value = _calculateResendRemaining();

    if (resendRemaining.value == Duration.zero) return;

    _resendTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      resendRemaining.value = _calculateResendRemaining();
      if (resendRemaining.value == Duration.zero) {
        _resendTimer?.cancel();
        _resendTimer = null;
      }
    });
  }

  void dispose() {
    _resendTimer?.cancel();
    _resendTimer = null;
    confirmPasswordResetCommand.dispose();
    resendPasswordResetCommand.dispose();
    returnToEmailCommand.dispose();
    isEnabled.dispose();
    resendRemaining.dispose();
  }
}
