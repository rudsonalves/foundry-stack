import 'package:flutter/foundation.dart';

import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/domain/common/password_reset/models/password_reset_step.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';

class PasswordResetNewPasswordViewmodel {
  final PasswordResetUsecase _usecase;

  PasswordResetNewPasswordViewmodel(this._usecase) {
    completePasswordResetCommand = Command1(_usecase.completePasswordReset);
    returnToCodeCommand = Command0(_usecase.returnToCode);
  }

  late final Command1<Unit, String> completePasswordResetCommand;
  late final Command0<Unit> returnToCodeCommand;
  final isEnabled = ValueNotifier<bool>(false);

  PasswordResetStep get step => _usecase.step;

  void updatePasswords(String password, String confirmation) {
    isEnabled.value = password.isValidPassword && password == confirmation;
  }

  void dispose() {
    completePasswordResetCommand.dispose();
    returnToCodeCommand.dispose();
    isEnabled.dispose();
  }
}
