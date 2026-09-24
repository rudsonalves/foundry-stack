import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/core/result/errors/app_error_code.dart';
import '/app/routing/routes.dart';
import '/ui/components/base/safe_scaffold.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/formaters/formatters.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/messages/app_snackbar.dart';
import '/ui/components/validators/validators.dart';
import '/ui/pages/password_reset/password_reset_route_result.dart';
import 'model/login_request_model.dart';
import 'viewmodel/login_viewmodel.dart';

class LoginPage extends StatefulWidget {
  final LoginViewmodel viewmodel;

  const LoginPage({
    super.key,
    required this.viewmodel,
  });

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _obscurePassword = ValueNotifier<bool>(true);
  final _isReady = ValueNotifier<bool>(false);

  LoginViewmodel get _viewmodel => widget.viewmodel;

  LoginRequestModel request = LoginRequestModel();

  @override
  void dispose() {
    _obscurePassword.dispose();
    _isReady.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return SafeScaffold(
      appBar: AppBar(
        title: const Text('Entrar'),
      ),
      body: GestureDetector(
        onTap: () => FocusScope.of(context).unfocus(),
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 460),
              child: Column(
                spacing: 16,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    'Acesse sua conta para continuar.',
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 12),
                  BasicInputText(
                    onChanged: _emailOnChanged,
                    keyboardType: TextInputType.emailAddress,
                    autofillHints: const [AutofillHints.email],
                    textInputAction: TextInputAction.next,
                    labelText: 'E-mail',
                    inputFormatters: [Formatters.noSpaces],
                    validator: Validators.email,
                    hintText: 'voce@exemplo.com',
                    prefixIcon: const Icon(Icons.email_outlined),
                  ),
                  ValueListenableBuilder<bool>(
                    valueListenable: _obscurePassword,
                    builder: (context, value, _) => BasicInputText(
                      onChanged: _passwordOnChanged,
                      obscureText: value,
                      autofillHints: const [AutofillHints.password],
                      textInputAction: TextInputAction.done,
                      labelText: 'Senha',
                      inputFormatters: [Formatters.noSpaces],
                      validator: Validators.notEmpty,
                      hintText: '********',
                      prefixIcon: const Icon(Icons.lock_outline),
                      suffixIcon: IconButton(
                        onPressed: _togglePasswordVisibility,
                        icon: Icon(
                          value
                              ? Icons.visibility_outlined
                              : Icons.visibility_off_outlined,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 6),
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton(
                      onPressed: _recoverPassword,
                      child: const Text('Esqueci minha senha'),
                    ),
                  ),
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton(
                      onPressed: _navToRegisterUser,
                      child: const Text('Não tem conta? Cadastre-se'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
      bottomNavigationBar: ListenableBuilder(
        listenable: Listenable.merge([
          _viewmodel.loginCommand,
          _isReady,
        ]),
        builder: (context, _) => BigButton(
          onPressed: _login,
          label: 'Entrar',
          rightIcon: Icon(Icons.login_rounded, size: 24),
          isRunning: _viewmodel.loginCommand.isRunning,
          enabled: _isReady.value,
        ),
      ),
    );
  }

  void _emailOnChanged(String value) {
    request = request.copyWith(
      email: value,
    );

    _isReady.value = _viewmodel.isValid(request);
  }

  void _passwordOnChanged(String value) {
    request = request.copyWith(
      password: value,
    );

    _isReady.value = _viewmodel.isValid(request);
  }

  Future<void> _login() async {
    if (_viewmodel.loginCommand.isRunning || !_isReady.value) return;

    await _viewmodel.loginCommand.execute(request);

    if (!mounted) return;
    if (_viewmodel.loginCommand.isFailure) {
      final error = _viewmodel.loginCommand.error!;

      switch (error.code) {
        case AppErrorCode.unauthenticated:
          AppSnackbar.show(
            context,
            message:
                'E-mail ou senha incorretos. Verifique os dados e tente novamente.',
            type: SnackbarType.info,
          );
          break;
        default:
          AppSnackbar.show(
            context,
            message:
                'Não foi possível entrar agora. Tente novamente em instantes',
            type: SnackbarType.error,
          );
      }

      return;
    }

    context.goNamed(BaseRoutes.home.routeName);
  }

  void _navToRegisterUser() {
    context.goNamed(BaseRoutes.register.routeName);
  }

  Future<void> _recoverPassword() async {
    final result = await context.pushNamed<PasswordResetRouteResult>(
      BaseRoutes.forgotPassword.routeName,
    );

    if (!mounted || result != PasswordResetRouteResult.completed) return;

    AppSnackbar.show(
      context,
      message: 'Senha alterada com sucesso. Entre com a nova senha.',
      type: SnackbarType.success,
    );
  }

  void _togglePasswordVisibility() {
    _obscurePassword.value = !_obscurePassword.value;
  }
}
