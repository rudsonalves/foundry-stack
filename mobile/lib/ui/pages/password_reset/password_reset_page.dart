import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/domain/common/password_reset/models/password_reset_step.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';
import '/ui/components/base/safe_scaffold.dart';
import 'code/password_reset_code_page.dart';
import 'email/password_reset_email_page.dart';
import 'new_password/password_reset_new_password_page.dart';
import 'password_reset_route_result.dart';

class PasswordResetPage extends StatefulWidget {
  final PasswordResetUsecase usecase;

  const PasswordResetPage({required this.usecase, super.key});

  @override
  State<PasswordResetPage> createState() => _PasswordResetPageState();
}

class _PasswordResetPageState extends State<PasswordResetPage> {
  bool _completed = false;

  PasswordResetUsecase get _usecase => widget.usecase;

  @override
  Widget build(BuildContext context) {
    if (_completed) return const SizedBox.shrink();

    final canLeaveDirectly = _usecase.step == PasswordResetStep.email;

    return PopScope<PasswordResetRouteResult>(
      canPop: canLeaveDirectly,
      onPopInvokedWithResult: (didPop, _) async {
        if (!didPop) await _requestExit();
      },
      child: SafeScaffold(
        appBar: AppBar(
          title: Text(_title),
          automaticallyImplyLeading: false,
          leading: IconButton(
            tooltip: 'Fechar recuperação',
            onPressed: _requestExit,
            icon: const Icon(Icons.close),
          ),
          bottom: PreferredSize(
            preferredSize: const Size.fromHeight(4),
            child: LinearProgressIndicator(value: _progress),
          ),
        ),
        body: AnimatedSwitcher(
          duration: const Duration(milliseconds: 200),
          child: _currentStep,
        ),
      ),
    );
  }

  String get _title => switch (_usecase.step) {
    PasswordResetStep.email => 'Recupere sua senha',
    PasswordResetStep.code => 'Confirme o código',
    PasswordResetStep.newPassword => 'Crie uma nova senha',
  };

  double get _progress => switch (_usecase.step) {
    PasswordResetStep.email => 1 / 3,
    PasswordResetStep.code => 2 / 3,
    PasswordResetStep.newPassword => 1,
  };

  Widget get _currentStep => switch (_usecase.step) {
    PasswordResetStep.email => PasswordResetEmailPage(
      key: const ValueKey(PasswordResetStep.email),
      usecase: _usecase,
      onStepChanged: _refreshStep,
    ),
    PasswordResetStep.code => PasswordResetCodePage(
      key: const ValueKey(PasswordResetStep.code),
      usecase: _usecase,
      onStepChanged: _refreshStep,
    ),
    PasswordResetStep.newPassword => PasswordResetNewPasswordPage(
      key: const ValueKey(PasswordResetStep.newPassword),
      usecase: _usecase,
      onStepChanged: _refreshStep,
      onCompleted: _completeJourney,
    ),
  };

  void _refreshStep() {
    if (mounted) setState(() {});
  }

  void _completeJourney() {
    if (!mounted) return;
    setState(() => _completed = true);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) context.pop(PasswordResetRouteResult.completed);
    });
  }

  Future<void> _requestExit() async {
    if (_usecase.step == PasswordResetStep.email) {
      context.pop();
      return;
    }

    final shouldExit = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Sair da recuperação?'),
        content: const Text(
          'O progresso atual será perdido e será necessário começar novamente.',
        ),
        actions: [
          TextButton(
            onPressed: () => dialogContext.pop(false),
            child: const Text('Continuar'),
          ),
          FilledButton(
            onPressed: () => dialogContext.pop(true),
            child: const Text('Sair'),
          ),
        ],
      ),
    );

    if (shouldExit == true && mounted) context.pop();
  }
}
