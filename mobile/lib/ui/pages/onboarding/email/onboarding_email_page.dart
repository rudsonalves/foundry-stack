import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/app/routing/routes.dart';
import '/core/result/command.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/formaters/formatters.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/texts/text_header.dart';
import '/ui/components/validators/validators.dart';
import 'viewmodel/onboarding_email_viewmodel.dart';

class OnboardingEmailPage extends StatefulWidget {
  final OnboardingUsecase usecase;
  final VoidCallback onStepChanged;
  final bool emailAlreadyRegistered;
  final VoidCallback onEmailEdited;

  const OnboardingEmailPage({
    super.key,
    required this.usecase,
    required this.onStepChanged,
    required this.emailAlreadyRegistered,
    required this.onEmailEdited,
  });

  @override
  State<OnboardingEmailPage> createState() => _OnboardingEmailPageState();
}

class _OnboardingEmailPageState extends State<OnboardingEmailPage> {
  late final OnboardingEmailViewmodel _viewmodel;
  late String _email;

  @override
  void initState() {
    super.initState();

    _viewmodel = OnboardingEmailViewmodel(widget.usecase);
    _email = _viewmodel.email?.trim() ?? '';
    if (widget.emailAlreadyRegistered) {
      _viewmodel.markEmailAsRegistered();
    }
  }

  @override
  void dispose() {
    _viewmodel.dispose();

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
                  TextHeader('Entre um email válido para cadastro'),
                  BasicInputText(
                    initialValue: _email,
                    onChanged: _onChange,
                    keyboardType: .emailAddress,
                    textCapitalization: .none,
                    autofillHints: const [AutofillHints.email],
                    inputFormatters: [Formatters.noSpaces],
                    textInputAction: .done,
                    // onSubmitted: (_) => _advanceEmail(),
                    validator: Validators.email,
                    // labelText: 'E-mail',
                    hintText: 'email@server.com',
                    prefixIcon: const Icon(Icons.mail_outline),
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
              _viewmodel.advanceEmailCommand,
              _viewmodel.returnToPreviousStepCommand,
              _viewmodel.cancelRegistrationCommand,
              _viewmodel.emailAlreadyRegistered,
              _viewmodel.isEnabled,
            ]),
            builder: (context, _) {
              final advanceCommand = _viewmodel.advanceEmailCommand;
              final returnCommand = _viewmodel.returnToPreviousStepCommand;
              final cancelCommand = _viewmodel.cancelRegistrationCommand;
              final isRunning =
                  advanceCommand.isRunning ||
                  returnCommand.isRunning ||
                  cancelCommand.isRunning;

              return OverflowBar(
                alignment: .spaceBetween,
                children: [
                  _viewmodel.emailAlreadyRegistered.value
                      ? TextButton(
                          onPressed: isRunning ? null : _cancelRegistration,
                          child: Text(
                            'Cancelar',
                            style: TextStyle(color: Colors.redAccent),
                          ),
                        )
                      : TextButton(
                          onPressed: isRunning ? null : _returnToPreviousStep,
                          child: const Text('Voltar'),
                        ),
                  SizedBox(
                    width: 200,
                    child: BigButton(
                      label: 'Continuar',
                      enabled:
                          _viewmodel.isEnabled.value &&
                          !returnCommand.isRunning &&
                          !cancelCommand.isRunning,
                      onPressed: _advanceEmail,
                      isRunning: advanceCommand.isRunning,
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

  Future<void> _advanceEmail() async {
    FocusScope.of(context).unfocus();

    if (Validators.email(_email) != null) return;

    final command = _viewmodel.advanceEmailCommand;

    await command.execute(_email);

    if (!mounted) return;

    if (command.isSuccess) {
      widget.onStepChanged();
      return;
    }

    _showAdvanceError(command.error!);
  }

  void _onChange(String value) {
    widget.onEmailEdited();
    _email = value.trim();
    _viewmodel.updateEmail(_email);
  }

  Future<void> _returnToPreviousStep() async {
    final command = _viewmodel.returnToPreviousStepCommand;

    await command.execute();

    if (!mounted) return;

    if (command.isSuccess) {
      widget.onStepChanged();
      return;
    }

    AppSnackbar.show(
      context,
      message: 'Não foi possível retornar. Tente novamente.',
      type: SnackbarType.error,
    );
  }

  Future<void> _cancelRegistration() async {
    final command = _viewmodel.cancelRegistrationCommand;

    await command.execute();

    if (!mounted) return;

    if (command.isSuccess) {
      context.goNamed(AuthRoutes.login.routeName);
      return;
    }

    AppSnackbar.show(
      context,
      message: 'Não foi possível cancelar o cadastro.',
      type: .error,
    );
  }

  void _showAdvanceError(AppError error) {
    final message = switch (backendErrorCode(error)) {
      BackendErrorCodes.emailAlreadyRegistered => 'E-mail em uso',
      BackendErrorCodes.rateLimitExceeded =>
        'Muitas tentativas. Aguarde antes de tentar novamente.',
      BackendErrorCodes.emailDeliveryUnavailable =>
        'Não foi possível enviar o código de verificação.',
      _ => switch (error.code) {
        AppErrorCode.networkError || AppErrorCode.timeout =>
          'Não foi possível conectar ao servidor. Tente novamente.',
        _ => 'Não foi possível continuar. Tente novamente.',
      },
    };

    if (backendErrorCode(error) == BackendErrorCodes.emailAlreadyRegistered) {
      _viewmodel.markEmailAsRegistered();
    }

    AppSnackbar.show(
      context,
      message: message,
      type: SnackbarType.error,
    );
  }
}
