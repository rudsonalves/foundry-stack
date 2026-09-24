import 'package:material_ui/material_ui.dart';

import '/core/result/result.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/input_text/token_input.dart';
import '/ui/components/messages/app_snackbar.dart';
import '../password_reset_messages.dart';
import 'viewmodel/password_reset_code_viewmodel.dart';

class PasswordResetCodePage extends StatefulWidget {
  final PasswordResetUsecase usecase;
  final VoidCallback onStepChanged;

  const PasswordResetCodePage({
    required this.usecase,
    required this.onStepChanged,
    super.key,
  });

  @override
  State<PasswordResetCodePage> createState() => _PasswordResetCodePageState();
}

class _PasswordResetCodePageState extends State<PasswordResetCodePage> {
  late final PasswordResetCodeViewmodel _viewmodel;
  final _codeController = TextEditingController();
  final _codeFocusNode = FocusNode();

  @override
  void initState() {
    super.initState();
    _viewmodel = PasswordResetCodeViewmodel(widget.usecase);
  }

  @override
  void dispose() {
    _viewmodel.dispose();
    _codeController.dispose();
    _codeFocusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: Listenable.merge([
        _viewmodel.confirmPasswordResetCommand,
        _viewmodel.resendPasswordResetCommand,
        _viewmodel.returnToEmailCommand,
        _viewmodel.isEnabled,
        _viewmodel.resendRemaining,
      ]),
      builder: (context, _) {
        final confirm = _viewmodel.confirmPasswordResetCommand;
        final resend = _viewmodel.resendPasswordResetCommand;
        final changeEmail = _viewmodel.returnToEmailCommand;
        final isRunning =
            confirm.isRunning || resend.isRunning || changeEmail.isRunning;

        return Column(
          key: const ValueKey('password-reset-code-step'),
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
                        'Digite o código de 6 dígitos enviado para ${_maskedEmail()}.',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      TokenInput(
                        controller: _codeController,
                        focusNode: _codeFocusNode,
                        length: 6,
                        fieldWidth: 44,
                        enabled: !isRunning,
                        onChanged: _viewmodel.updateCode,
                      ),
                      Text(
                        _expirationLabel(context),
                        textAlign: TextAlign.center,
                      ),
                      Align(
                        alignment: Alignment.centerRight,
                        child: TextButton(
                          onPressed: _viewmodel.canResend && !isRunning
                              ? _resend
                              : null,
                          child: Text(
                            resend.isRunning
                                ? 'Reenviando...'
                                : _resendLabel(
                                    _viewmodel.resendRemaining.value,
                                  ),
                          ),
                        ),
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
                    onPressed: isRunning ? null : _changeEmail,
                    child: const Text('Alterar e-mail'),
                  ),
                  SizedBox(
                    width: 200,
                    child: BigButton(
                      label: 'Confirmar código',
                      enabled: _viewmodel.isEnabled.value && !isRunning,
                      isRunning: confirm.isRunning,
                      onPressed: _confirm,
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

  String _maskedEmail() {
    final email = _viewmodel.email ?? '';
    final separator = email.indexOf('@');
    if (separator <= 0) return 'seu e-mail';

    final local = email.substring(0, separator);
    final domain = email.substring(separator);
    final visible = local.length == 1
        ? local
        : '${local[0]}${local.length > 2 ? '***' : '*'}${local[local.length - 1]}';
    return '$visible$domain';
  }

  String _expirationLabel(BuildContext context) {
    final expiration = _viewmodel.challenge?.codeExpiresAt.toLocal();
    if (expiration == null) return '';
    final time = MaterialLocalizations.of(
      context,
    ).formatTimeOfDay(TimeOfDay.fromDateTime(expiration));
    return 'Código válido até $time.';
  }

  String _resendLabel(Duration remaining) {
    if (remaining == Duration.zero) return 'Reenviar código';
    final seconds = (remaining.inMilliseconds / 1000).ceil();
    final minutesPart = seconds ~/ 60;
    final secondsPart = seconds % 60;
    return 'Reenviar código em '
        '${minutesPart.toString().padLeft(2, '0')}:'
        '${secondsPart.toString().padLeft(2, '0')}';
  }

  Future<void> _confirm() async {
    if (!_viewmodel.isEnabled.value) return;
    FocusScope.of(context).unfocus();

    final command = _viewmodel.confirmPasswordResetCommand;
    await command.execute(_codeController.text);
    if (!mounted) return;

    if (command.isSuccess) {
      _clearCode();
      widget.onStepChanged();
      return;
    }

    final error = command.error!;
    final message = error.code == AppErrorCode.invalidData
        ? PasswordResetMessages.invalid
        : PasswordResetMessages.forError(
            error,
            fallback: PasswordResetMessages.invalid,
          );
    AppSnackbar.show(context, message: message, type: SnackbarType.error);
    _codeFocusNode.requestFocus();
  }

  Future<void> _resend() async {
    final command = _viewmodel.resendPasswordResetCommand;
    await command.execute();
    if (!mounted) return;

    if (command.isSuccess) {
      _clearCode();
      _viewmodel.restartResendCountdown();
      _codeFocusNode.requestFocus();
      return;
    }

    AppSnackbar.show(
      context,
      message: PasswordResetMessages.forError(
        command.error!,
        fallback: 'Não foi possível reenviar o código. Tente novamente.',
      ),
      type: SnackbarType.error,
    );
    _codeFocusNode.requestFocus();
  }

  Future<void> _changeEmail() async {
    final command = _viewmodel.returnToEmailCommand;
    await command.execute();
    if (!mounted) return;

    if (command.isSuccess) {
      _clearCode();
      widget.onStepChanged();
      return;
    }

    AppSnackbar.show(
      context,
      message: 'Não foi possível alterar o e-mail. Tente novamente.',
      type: SnackbarType.error,
    );
  }

  void _clearCode() {
    _codeController.clear();
    _viewmodel.updateCode('');
  }
}
