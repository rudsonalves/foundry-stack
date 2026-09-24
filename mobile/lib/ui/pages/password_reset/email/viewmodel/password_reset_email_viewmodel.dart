import 'package:flutter/foundation.dart';

import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/domain/common/password_reset/models/password_reset_step.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';

class PasswordResetEmailViewmodel {
  final PasswordResetUsecase _usecase;

  PasswordResetEmailViewmodel(this._usecase) {
    requestPasswordResetCommand = Command1(_usecase.requestPasswordReset);

    updateEmail(email ?? '');
  }

  late final Command1<Unit, String> requestPasswordResetCommand;
  final isEnabled = ValueNotifier<bool>(false);

  PasswordResetStep get step => _usecase.step;
  String? get email => _usecase.email;

  void updateEmail(String value) {
    isEnabled.value = value.trim().isValidEmail;
  }

  void dispose() {
    requestPasswordResetCommand.dispose();
    isEnabled.dispose();
  }
}
