import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/app/routing/routes.dart';
import '/core/result/result.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/formaters/formatters.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/texts/text_header.dart';
import '/ui/components/validators/validators.dart';
import 'viewmodel/onboarding_password_viewmodel.dart';

class OnboardingPasswordPage extends StatefulWidget {
  final OnboardingUsecase usecase;
  final VoidCallback onStepChanged;
  final VoidCallback onEmailAlreadyRegistered;

  const OnboardingPasswordPage({
    super.key,
    required this.usecase,
    required this.onStepChanged,
    required this.onEmailAlreadyRegistered,
  });
  @override
  State<OnboardingPasswordPage> createState() => _OnboardingPasswordPageState();
}

class _OnboardingPasswordPageState extends State<OnboardingPasswordPage> {
  late final OnboardingPasswordViewmodel _viewmodel;

  String _password = '';
  String _passwordConfirmation = '';

  @override
  void initState() {
    super.initState();

    _viewmodel = OnboardingPasswordViewmodel(widget.usecase);
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
            child: SingleChildScrollView(
              padding: EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: .start,
                mainAxisSize: .min,
                children: [
                  TextHeader(
                    'Crie uma senha com pelo menos 8 caracteres, letras e números.',
                  ),

                  // Password
                  BasicInputText(
                    onChanged: _onPasswordChanged,
                    validator: Validators.password,
                    keyboardType: .visiblePassword,
                    autofillHints: const [AutofillHints.newPassword],
                    textInputAction: .next,
                    labelText: 'Senha',
                    inputFormatters: [Formatters.noSpaces],
                    hintText: '******',
                    prefixIcon: const Icon(Icons.lock_outlined),
                    obscureText: true,
                  ),
                  const SizedBox(height: 8),
                  ListenableBuilder(
                    listenable: _viewmodel.requirements,
                    builder: (context, _) {
                      final requirements = _viewmodel.requirements.value;

                      return Column(
                        spacing: 8,
                        children: [
                          _buildRequirement(
                            'Pelo menos 8 caracteres',
                            requirements.hasMinimumLength,
                          ),
                          _buildRequirement(
                            'Pelo menos uma letra',
                            requirements.hasLetter,
                          ),
                          _buildRequirement(
                            'Pelo menos um número',
                            requirements.hasNumber,
                          ),
                        ],
                      );
                    },
                  ),

                  const SizedBox(height: 12),

                  // Password Confirmation
                  BasicInputText(
                    onChanged: _onPasswordConfirmationChanged,
                    validator: (value) =>
                        Validators.confirmPassword(value, _password),
                    keyboardType: .visiblePassword,
                    autofillHints: const [AutofillHints.newPassword],
                    textInputAction: .next,
                    labelText: 'Confirmar Senha',
                    inputFormatters: [Formatters.noSpaces],
                    hintText: '******',
                    prefixIcon: const Icon(Icons.lock_outlined),
                    obscureText: true,
                  ),

                  const SizedBox(height: 8),
                  ListenableBuilder(
                    listenable: _viewmodel.requirements,
                    builder: (context, _) => _buildRequirement(
                      'Senhas iguais',
                      _viewmodel.requirements.value.passwordsMatch,
                    ),
                  ),

                  const SizedBox(height: 16),
                ],
              ),
            ),
          ),
        ),

        Padding(
          padding: const EdgeInsets.all(16),
          child: ListenableBuilder(
            listenable: Listenable.merge([
              _viewmodel.createAccountCommand,
              _viewmodel.returnToPreviousStepCommand,
              _viewmodel.isEnabled,
            ]),
            builder: (context, _) {
              final createCommand = _viewmodel.createAccountCommand;
              final returnCommand = _viewmodel.returnToPreviousStepCommand;
              final isRunning =
                  createCommand.isRunning || returnCommand.isRunning;

              return OverflowBar(
                alignment: MainAxisAlignment.spaceBetween,
                children: [
                  TextButton(
                    onPressed: isRunning ? null : _returnToPreviousStep,
                    child: const Text('Voltar'),
                  ),
                  SizedBox(
                    width: 260,
                    child: BigButton(
                      label: 'Criar conta',
                      enabled:
                          _viewmodel.isEnabled.value &&
                          !returnCommand.isRunning,
                      isRunning: createCommand.isRunning,
                      onPressed: _createAccount,
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

  void _onPasswordChanged(String value) {
    _password = value;
    _viewmodel.updatePassword(_password, _passwordConfirmation);
  }

  void _onPasswordConfirmationChanged(String value) {
    _passwordConfirmation = value;
    _viewmodel.updatePassword(_password, _passwordConfirmation);
  }

  bool _validatePasswords() {
    return Validators.password(_password) == null &&
        Validators.confirmPassword(_passwordConfirmation, _password) == null;
  }

  Widget _buildRequirement(String label, bool fulfilled) {
    final color = fulfilled
        ? Colors.green
        : Theme.of(context).colorScheme.error;

    return Row(
      children: [
        Icon(
          fulfilled ? Icons.check : Icons.circle_outlined,
          color: color,
          size: 18,
        ),
        const SizedBox(width: 8),
        Text(
          label,
          style: TextStyle(color: color),
        ),
      ],
    );
  }

  Future<void> _createAccount() async {
    FocusScope.of(context).unfocus();

    if (!_validatePasswords()) return;

    final command = _viewmodel.createAccountCommand;

    await command.execute(_password);

    if (!mounted) return;

    if (command.isSuccess) {
      _clearPasswords();

      AppSnackbar.show(
        context,
        message: 'Cadastro realizado com sucesso! Faça login para continuar.',
        type: .success,
      );

      context.goNamed(AuthRoutes.login.routeName);
      return;
    }

    final backendCode = backendErrorCode(command.error);

    if (backendCode == BackendErrorCodes.emailAlreadyRegistered) {
      _clearPasswords();

      AppSnackbar.show(
        context,
        message: 'E-mail em uso',
        type: .error,
      );

      widget.onEmailAlreadyRegistered();
      return;
    }

    if (backendCode == BackendErrorCodes.invalidEmailVerification ||
        _viewmodel.step != .password) {
      _clearPasswords();

      final message = _viewmodel.step == .emailVerification
          ? 'A verificação expirou. Enviamos um novo código.'
          : 'A verificação expirou. Solicite um novo código.';

      AppSnackbar.show(
        context,
        message: message,
        type: .error,
      );

      widget.onStepChanged();
      return;
    }

    final error = command.error!;

    final message = switch (backendCode) {
      BackendErrorCodes.rateLimitExceeded =>
        'Muitas tentativas. Aguarde antes de tentar novamente.',
      _ => switch (error.code) {
        AppErrorCode.networkError || AppErrorCode.timeout =>
          'Não foi possível conectar ao servidor. Tente novamente.',
        _ => 'Não foi possível criar a conta. Tente novamente.',
      },
    };

    AppSnackbar.show(
      context,
      message: message,
      type: .error,
    );
  }

  Future<void> _returnToPreviousStep() async {
    final command = _viewmodel.returnToPreviousStepCommand;

    await command.execute();

    if (!mounted || !command.isSuccess) return;

    widget.onStepChanged();
  }

  void _clearPasswords() {
    _password = '';
    _passwordConfirmation = '';
    _viewmodel.updatePassword('', '');
  }
}
