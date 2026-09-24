import 'package:material_ui/material_ui.dart';

import '../../../domain/usecases/onboarding/onboarding_usecase.dart';
import '../../components/base/safe_scaffold.dart';
import 'email/onboarding_email_page.dart';
import 'email_verification/onboarding_email_verification_page.dart';
import 'password/onboarding_password_page.dart';
import 'user_name/onboarding_user_name_page.dart';
import 'viewmodel/onboarding_coordinator_viewmodel.dart';

class OnboardingPage extends StatefulWidget {
  final OnboardingCoordinatorViewmodel viewmodel;
  final OnboardingUsecase usecase;

  const OnboardingPage({
    super.key,
    required this.viewmodel,
    required this.usecase,
  });

  @override
  State<OnboardingPage> createState() => _OnboardingPageState();
}

class _OnboardingPageState extends State<OnboardingPage> {
  OnboardingCoordinatorViewmodel get _viewmodel => widget.viewmodel;
  OnboardingUsecase get _usecase => widget.usecase;

  bool _emailAlreadyRegistered = false;

  @override
  void initState() {
    super.initState();

    _viewmodel.initializeCommand.execute();
  }

  @override
  void dispose() {
    _viewmodel.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: _viewmodel.initializeCommand,
      builder: (context, _) => SafeScaffold(
        appBar: AppBar(
          title: Text(_title),
          bottom: _buildProgressIndicator(),
          automaticallyImplyLeading: false,
        ),
        body: _buildBody(),
      ),
    );
  }

  String get _title => switch (_viewmodel.step) {
    .userName => 'Seu nome',
    .email => 'Seu e-mail',
    .emailVerification => 'Verifique seu e-mail',
    .password => 'Crie sua senha',
    null => 'Cadastro',
  };

  double? get _progress => switch (_viewmodel.step) {
    .userName => .25,
    .email => .5,
    .emailVerification => .75,
    .password => 1.0,
    null => null,
  };

  PreferredSizeWidget? _buildProgressIndicator() {
    final progress = _progress;
    if (progress == null) return null;

    return PreferredSize(
      preferredSize: const Size.fromHeight(4),
      child: LinearProgressIndicator(value: progress),
    );
  }

  Widget _buildBody() {
    final command = _viewmodel.initializeCommand;

    if (command.isIdle || command.isRunning) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }

    if (command.isFailure) {
      return Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Column(
            mainAxisSize: .min,
            spacing: 16,
            children: [
              const Text(
                'Não foi possível iniciar o cadastro.',
                textAlign: .center,
              ),
              FilledButton.icon(
                onPressed: command.isRunning ? null : command.execute,
                icon: const Icon(Icons.refresh),
                label: Text('Tentar novamente'),
              ),
            ],
          ),
        ),
      );
    }

    return _buildCurrentStep();
  }

  Widget _buildCurrentStep() {
    return switch (_viewmodel.step) {
      .userName => OnboardingUserNamePage(
        usecase: _usecase,
        onStepChanged: _refreshStep,
      ),
      .email => OnboardingEmailPage(
        usecase: _usecase,
        onStepChanged: _refreshStep,
        emailAlreadyRegistered: _emailAlreadyRegistered,
        onEmailEdited: () => _emailAlreadyRegistered = false,
      ),
      .emailVerification => OnboardingEmailVerificationPage(
        usecase: _usecase,
        onStepChanged: _refreshStep,
      ),
      .password => OnboardingPasswordPage(
        usecase: _usecase,
        onStepChanged: _refreshStep,
        onEmailAlreadyRegistered: _handleEmailAlreadyRegistered,
      ),
      null => const SizedBox.shrink(),
    };
  }

  void _handleEmailAlreadyRegistered() {
    if (!mounted) return;

    setState(() {
      _emailAlreadyRegistered = true;
    });
  }

  void _refreshStep() {
    if (!mounted) return;

    setState(() {});
  }
}
