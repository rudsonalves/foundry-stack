import 'package:material_ui/material_ui.dart';

import '/core/result/result.dart';
import '/domain/common/password_reset/models/password_reset_step.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/formaters/formatters.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/validators/validators.dart';
import '../password_reset_messages.dart';
import 'viewmodel/password_reset_new_password_viewmodel.dart';

class PasswordResetNewPasswordPage extends StatefulWidget {
  final PasswordResetUsecase usecase;
  final VoidCallback onStepChanged;
  final VoidCallback onCompleted;

  const PasswordResetNewPasswordPage({
    required this.usecase,
    required this.onStepChanged,
    required this.onCompleted,
    super.key,
  });

  @override
  State<PasswordResetNewPasswordPage> createState() =>
      _PasswordResetNewPasswordPageState();
}

class _PasswordResetNewPasswordPageState
    extends State<PasswordResetNewPasswordPage> {
  late final PasswordResetNewPasswordViewmodel _viewmodel;
  final _passwordController = TextEditingController();
  final _confirmationController = TextEditingController();
  final _passwordFocusNode = FocusNode();
  final _confirmationFocusNode = FocusNode();
  final _obscurePassword = ValueNotifier(true);
  final _obscureConfirmation = ValueNotifier(true);

  @override
  void initState() {
    super.initState();
    _viewmodel = PasswordResetNewPasswordViewmodel(widget.usecase);
  }

  @override
  void dispose() {
    _viewmodel.dispose();
    _passwordController.dispose();
    _confirmationController.dispose();
    _passwordFocusNode.dispose();
    _confirmationFocusNode.dispose();
    _obscurePassword.dispose();
    _obscureConfirmation.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: Listenable.merge([
        _viewmodel.completePasswordResetCommand,
        _viewmodel.returnToCodeCommand,
        _viewmodel.isEnabled,
        _obscurePassword,
        _obscureConfirmation,
      ]),
      builder: (context, _) {
        final complete = _viewmodel.completePasswordResetCommand;
        final goBack = _viewmodel.returnToCodeCommand;
        final isRunning = complete.isRunning || goBack.isRunning;

        return Column(
          key: const ValueKey('password-reset-new-password-step'),
          children: [
            Expanded(
              child: GestureDetector(
                onTap: () => FocusScope.of(context).unfocus(),
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    spacing: 16,
                    children: [
                      Text(
                        'Use pelo menos 8 caracteres, com letras e números.',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      BasicInputText(
                        controller: _passwordController,
                        focusNode: _passwordFocusNode,
                        enabled: !isRunning,
                        onChanged: (_) => _passwordsChanged(),
                        validator: Validators.password,
                        keyboardType: TextInputType.visiblePassword,
                        autofillHints: const [AutofillHints.newPassword],
                        textInputAction: TextInputAction.next,
                        inputFormatters: [Formatters.noSpaces],
                        labelText: 'Nova senha',
                        obscureText: _obscurePassword.value,
                        prefixIcon: const Icon(Icons.lock_outline),
                        suffixIcon: IconButton(
                          onPressed: isRunning
                              ? null
                              : () => _obscurePassword.value =
                                    !_obscurePassword.value,
                          icon: Icon(
                            _obscurePassword.value
                                ? Icons.visibility_outlined
                                : Icons.visibility_off_outlined,
                          ),
                        ),
                      ),
                      BasicInputText(
                        controller: _confirmationController,
                        focusNode: _confirmationFocusNode,
                        enabled: !isRunning,
                        onChanged: (_) => _passwordsChanged(),
                        validator: (value) => Validators.confirmPassword(
                          value,
                          _passwordController.text,
                        ),
                        keyboardType: TextInputType.visiblePassword,
                        autofillHints: const [AutofillHints.newPassword],
                        textInputAction: TextInputAction.done,
                        inputFormatters: [Formatters.noSpaces],
                        labelText: 'Confirmar nova senha',
                        obscureText: _obscureConfirmation.value,
                        prefixIcon: const Icon(Icons.lock_outline),
                        suffixIcon: IconButton(
                          onPressed: isRunning
                              ? null
                              : () => _obscureConfirmation.value =
                                    !_obscureConfirmation.value,
                          icon: Icon(
                            _obscureConfirmation.value
                                ? Icons.visibility_outlined
                                : Icons.visibility_off_outlined,
                          ),
                        ),
                        onSubmitted: (_) => _complete(),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: OverflowBar(
                alignment: MainAxisAlignment.spaceBetween,
                children: [
                  TextButton(
                    onPressed: isRunning ? null : _returnToCode,
                    child: const Text('Voltar'),
                  ),
                  SizedBox(
                    width: 220,
                    child: BigButton(
                      label: 'Alterar senha',
                      enabled: _viewmodel.isEnabled.value && !isRunning,
                      isRunning: complete.isRunning,
                      onPressed: _complete,
                    ),
                  ),
                ],
              ),
            ),
          ],
        );
      },
    );
  }

  void _passwordsChanged() {
    _viewmodel.updatePasswords(
      _passwordController.text,
      _confirmationController.text,
    );
  }

  Future<void> _complete() async {
    if (!_viewmodel.isEnabled.value) return;
    FocusScope.of(context).unfocus();

    final command = _viewmodel.completePasswordResetCommand;
    await command.execute(_passwordController.text);
    if (!mounted) return;

    if (command.isSuccess) {
      _clearPasswords();
      widget.onCompleted();
      return;
    }

    final error = command.error!;
    if (backendErrorCode(error) == BackendErrorCodes.invalidPasswordReset ||
        _viewmodel.step != PasswordResetStep.newPassword) {
      _clearPasswords();
      AppSnackbar.show(
        context,
        message: PasswordResetMessages.invalid,
        type: SnackbarType.error,
      );
      widget.onStepChanged();
      return;
    }

    AppSnackbar.show(
      context,
      message: PasswordResetMessages.forError(
        error,
        fallback: 'Não foi possível alterar a senha. Tente novamente.',
      ),
      type: SnackbarType.error,
    );
  }

  Future<void> _returnToCode() async {
    final command = _viewmodel.returnToCodeCommand;
    await command.execute();
    if (!mounted || !command.isSuccess) return;

    _clearPasswords();
    widget.onStepChanged();
  }

  void _clearPasswords() {
    _passwordController.clear();
    _confirmationController.clear();
    _viewmodel.updatePasswords('', '');
  }
}
