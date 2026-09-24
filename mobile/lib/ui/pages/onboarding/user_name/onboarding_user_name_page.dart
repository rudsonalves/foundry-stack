import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/app/routing/routes.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';
import '/ui/components/buttons/big_button.dart';
import '/ui/components/input_text/basic_input_text.dart';
import '/ui/components/validators/validators.dart';
import '../../../components/texts/text_header.dart';
import 'viewmodel/onboarding_user_name_viewmodel.dart';

class OnboardingUserNamePage extends StatefulWidget {
  final OnboardingUsecase usecase;
  final VoidCallback onStepChanged;

  const OnboardingUserNamePage({
    super.key,
    required this.usecase,
    required this.onStepChanged,
  });

  @override
  State<OnboardingUserNamePage> createState() => _OnboardingUserNamePageState();
}

class _OnboardingUserNamePageState extends State<OnboardingUserNamePage> {
  late final OnboardingUserNameViewmodel _viewmodel;
  late String _name;

  @override
  void initState() {
    super.initState();

    _viewmodel = OnboardingUserNameViewmodel(widget.usecase);
    _name = _viewmodel.name?.trim() ?? '';
  }

  @override
  void dispose() {
    _viewmodel.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: .start,
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
                  TextHeader('Entre com seu nome completo'),
                  BasicInputText(
                    initialValue: _name,
                    onChanged: _onChange,
                    onSubmitted: (_) => _advanceName(),
                    keyboardType: .name,
                    textCapitalization: .words,
                    textInputAction: .done,
                    autofillHints: const [AutofillHints.name],
                    validator: Validators.notEmpty,
                    // labelText: 'Nome',
                    hintText: 'Nome completo',
                    prefixIcon: const Icon(Icons.person_outline),
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
              _viewmodel.advanceNameCommand,
              _viewmodel.isEnabled,
            ]),
            builder: (context, _) {
              final command = _viewmodel.advanceNameCommand;

              return OverflowBar(
                alignment: .spaceBetween,
                children: [
                  TextButton(
                    onPressed: command.isRunning ? null : _cancelOnboarding,
                    child: const Text('Cancelar'),
                  ),
                  SizedBox(
                    width: 200,
                    child: BigButton(
                      label: 'Continuar',
                      enabled: _viewmodel.isEnabled.value,
                      onPressed: _advanceName,
                      isRunning: command.isRunning,
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

  Future<void> _advanceName() async {
    FocusScope.of(context).unfocus();

    if (Validators.notEmpty(_name) != null) return;

    await _viewmodel.advanceNameCommand.execute(_name);

    if (!mounted || !_viewmodel.advanceNameCommand.isSuccess) return;

    widget.onStepChanged();
  }

  void _onChange(String value) {
    _name = value.trim();
    _viewmodel.updateName(_name);
  }

  void _cancelOnboarding() {
    context.goNamed(AuthRoutes.login.routeName);
  }
}
