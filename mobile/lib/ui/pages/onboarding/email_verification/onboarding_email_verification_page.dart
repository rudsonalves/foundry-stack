import 'package:material_ui/material_ui.dart';

import '/core/extensions/string.dart';
import '/core/result/errors/app_error.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/input_text/token_input.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/texts/text_header.dart';
import 'viewmodel/onboarding_email_verification_viewmodel.dart';

class OnboardingEmailVerificationPage extends StatefulWidget {
  final OnboardingUsecase usecase;
  final VoidCallback onStepChanged;

  const OnboardingEmailVerificationPage({
    super.key,
    required this.usecase,
    required this.onStepChanged,
  });

  @override
  State<OnboardingEmailVerificationPage> createState() =>
      _OnboardingEmailVerificationPageState();
}

class _OnboardingEmailVerificationPageState
    extends State<OnboardingEmailVerificationPage> {
  late final OnboardingEmailVerificationViewmodel _viewmodel;
  late final FocusNode _tokenFocusNode;
  final _tokenController = TextEditingController();

  @override
  void initState() {
    super.initState();

    _tokenFocusNode = FocusNode();
    _viewmodel = OnboardingEmailVerificationViewmodel(widget.usecase);
  }

  @override
  void dispose() {
    _viewmodel.dispose();
    _tokenFocusNode.dispose();
    _tokenController.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Expanded(
          child: GestureDetector(
            onTap: () => FocusScope.of(context).unfocus(),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: .start,
                mainAxisSize: .min,
                children: [
                  _viewmodel.email == null
                      ? TextHeader('Informe o código enviado para o seu email.')
                      : TextHeader(
                          'Enviamos um código de 6 dígitos para seu email ${_viewmodel.email}.',
                        ),
                  TokenInput(
                    focusNode: _tokenFocusNode,
                    controller: _tokenController,
                    length: 6,
                    onChanged: _onChange,
                  ),
                  ListenableBuilder(
                    listenable: Listenable.merge([
                      _viewmodel.resendRemaining,
                      _viewmodel.confirmVerificationCommand,
                      _viewmodel.resendVerificationCommand,
                      _viewmodel.returnToPreviousStepCommand,
                    ]),
                    builder: (context, _) {
                      final remaining = _viewmodel.resendRemaining.value;
                      final resendCommand =
                          _viewmodel.resendVerificationCommand;

                      final isRunning =
                          _viewmodel.confirmVerificationCommand.isRunning ||
                          resendCommand.isRunning ||
                          _viewmodel.returnToPreviousStepCommand.isRunning;

                      final canResend =
                          remaining == Duration.zero && !isRunning;

                      return Align(
                        alignment: .centerRight,
                        child: TextButton(
                          onPressed: canResend
                              ? _resendEmailVerification
                              : null,
                          child: Text(
                            resendCommand.isRunning
                                ? 'Reenviando...'
                                : _resendLabel(remaining),
                          ),
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(16),
          child: ListenableBuilder(
            listenable: Listenable.merge([
              _viewmodel.confirmVerificationCommand,
              _viewmodel.resendVerificationCommand,
              _viewmodel.returnToPreviousStepCommand,
              _viewmodel.isEnabled,
            ]),
            builder: (context, _) {
              final confirmVerificationCommand =
                  _viewmodel.confirmVerificationCommand;
              final resendVerificationCommand =
                  _viewmodel.resendVerificationCommand;
              final returnToPreviousStepCommand =
                  _viewmodel.returnToPreviousStepCommand;
              final isRunning =
                  confirmVerificationCommand.isRunning ||
                  resendVerificationCommand.isRunning ||
                  returnToPreviousStepCommand.isRunning;

              return OverflowBar(
                alignment: .spaceBetween,
                children: [
                  TextButton(
                    onPressed: isRunning ? null : _returnToPreviousStep,
                    child: const Text('Voltar'),
                  ),
                  SizedBox(
                    width: 200,
                    child: BigButton(
                      label: 'Continuar',
                      enabled:
                          _viewmodel.isEnabled.value &&
                          !resendVerificationCommand.isRunning &&
                          !returnToPreviousStepCommand.isRunning,
                      onPressed: _advanceEmailVerification,
                      isRunning: confirmVerificationCommand.isRunning,
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ],
    );
  }

  String _resendLabel(Duration remaining) {
    if (remaining == Duration.zero) return 'Reenviar código';

    final totalSeconds = (remaining.inMilliseconds / 1000).ceil();
    final minutes = totalSeconds ~/ 60;
    final seconds = totalSeconds % 60;

    return 'Reenviar código em '
        '${minutes.toString().padLeft(2, '0')}:'
        '${seconds.toString().padLeft(2, '0')}';
  }

  Future<void> _resendEmailVerification() async {
    FocusScope.of(context).unfocus();

    final command = _viewmodel.resendVerificationCommand;

    await command.execute();

    if (!mounted) return;

    if (command.isSuccess) {
      _clearToken();
      _viewmodel.restartResendCountdown();
      _tokenFocusNode.requestFocus();
      return;
    }

    _showResendError(command.error!);
    _tokenFocusNode.requestFocus();
  }

  void _showResendError(AppError error) {
    final message = switch (backendErrorCode(error)) {
      BackendErrorCodes.rateLimitExceeded =>
        'Muitas tentativas. Aguarde antes de tentar novamente.',
      BackendErrorCodes.emailDeliveryUnavailable =>
        'Não foi possível enviar o código de verificação.',
      _ => switch (error.code) {
        AppErrorCode.networkError || AppErrorCode.timeout =>
          'Não foi possível conectar ao servidor. Tente novamente.',
        _ => 'Não foi possível reenviar o código. Tente novamente.',
      },
    };

    AppSnackbar.show(
      context,
      message: message,
      type: .error,
    );
  }

  Future<void> _advanceEmailVerification() async {
    FocusScope.of(context).unfocus();

    final token = _tokenController.text;

    if (!_validator(token)) {
      _tokenFocusNode.requestFocus();
      return;
    }

    final command = _viewmodel.confirmVerificationCommand;

    await command.execute(token);

    if (!mounted) return;

    if (command.isSuccess) {
      _clearToken();
      widget.onStepChanged();
      return;
    }

    _tokenFocusNode.requestFocus();
  }

  void _onChange(String? token) {
    _viewmodel.updateEmailVerification(token ?? '');
  }

  bool _validator(String otp) {
    return otp.isNumbers && otp.length == 6;
  }

  Future<void> _returnToPreviousStep() async {
    final command = _viewmodel.returnToPreviousStepCommand;

    await command.execute();

    if (!mounted || !command.isSuccess) return;

    widget.onStepChanged();
  }

  void _clearToken() {
    _tokenController.clear();
  }
}
