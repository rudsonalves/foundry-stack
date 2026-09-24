import 'package:material_ui/material_ui.dart';

import '/core/result/command.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/formaters/formatters.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/validators/validators.dart';
import 'viewmodel/password_reset_email_viewmodel.dart';
import '../password_reset_messages.dart';

class PasswordResetEmailPage extends StatefulWidget {
  final PasswordResetUsecase usecase;
  final VoidCallback onStepChanged;

  const PasswordResetEmailPage({
    required this.usecase,
    required this.onStepChanged,
    super.key,
  });

  @override
  State<PasswordResetEmailPage> createState() => _PasswordResetEmailPageState();
}

class _PasswordResetEmailPageState extends State<PasswordResetEmailPage> {
  late final PasswordResetEmailViewmodel _viewmodel;
  late String _email;

  @override
  void initState() {
    super.initState();
    _viewmodel = PasswordResetEmailViewmodel(widget.usecase);
    _email = _viewmodel.email ?? '';
  }

  @override
  void dispose() {
    _viewmodel.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: Listenable.merge([
        _viewmodel.requestPasswordResetCommand,
        _viewmodel.isEnabled,
      ]),
      builder: (context, _) {
        final command = _viewmodel.requestPasswordResetCommand;

        return Column(
          key: const ValueKey('password-reset-email-step'),
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
                        'Informe o e-mail usado na sua conta.',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const Text(
                        'Se houver uma conta cadastrada, enviaremos um código para continuar.',
                      ),
                      BasicInputText(
                        initialValue: _email,
                        enabled: !command.isRunning,
                        onChanged: _emailChanged,
                        onSubmitted: (_) => _requestPasswordReset(),
                        keyboardType: TextInputType.emailAddress,
                        autofillHints: const [AutofillHints.email],
                        inputFormatters: [Formatters.noSpaces],
                        textInputAction: TextInputAction.done,
                        validator: Validators.email,
                        labelText: 'E-mail',
                        hintText: 'voce@exemplo.com',
                        prefixIcon: const Icon(Icons.email_outlined),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: BigButton(
                label: 'Enviar código',
                enabled: _viewmodel.isEnabled.value,
                isRunning: command.isRunning,
                onPressed: _requestPasswordReset,
              ),
            ),
          ],
        );
      },
    );
  }

  void _emailChanged(String value) {
    _email = value.trim();
    _viewmodel.updateEmail(value);
  }

  Future<void> _requestPasswordReset() async {
    if (!_viewmodel.isEnabled.value ||
        _viewmodel.requestPasswordResetCommand.isRunning) {
      return;
    }

    FocusScope.of(context).unfocus();
    final command = _viewmodel.requestPasswordResetCommand;
    await command.execute(_email);

    if (!mounted) return;

    if (command.isSuccess) {
      AppSnackbar.show(
        context,
        message:
            'Se este e-mail estiver cadastrado, enviaremos um código de recuperação.',
        type: SnackbarType.info,
      );
      widget.onStepChanged();
      return;
    }

    _showError(command.error!);
  }

  void _showError(AppError error) {
    final message = PasswordResetMessages.forError(
      error,
      fallback: 'Não foi possível enviar o código. Tente novamente.',
    );

    AppSnackbar.show(
      context,
      message: message,
      type: SnackbarType.error,
    );
  }
}
