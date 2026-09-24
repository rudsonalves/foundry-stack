import 'package:flutter/widgets.dart';

import 'onboarding_scope.dart';

typedef OnboardingScopeCreator = OnboardingScope Function();
typedef OnboardingScopeWidgetBuilder = Widget Function(OnboardingScope scope);

class OnboardingScopeBoundary extends StatefulWidget {
  final OnboardingScopeCreator createScope;
  final OnboardingScopeWidgetBuilder builder;

  const OnboardingScopeBoundary({
    required this.createScope,
    required this.builder,
    super.key,
  });

  @override
  State<OnboardingScopeBoundary> createState() =>
      _OnboardingScopeBoundaryState();
}

class _OnboardingScopeBoundaryState extends State<OnboardingScopeBoundary> {
  late final OnboardingScope _scope;
  late final Widget _journey;

  @override
  void initState() {
    super.initState();
    _scope = widget.createScope();
    _journey = widget.builder(_scope);
  }

  @override
  void dispose() {
    _scope.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => _journey;
}
