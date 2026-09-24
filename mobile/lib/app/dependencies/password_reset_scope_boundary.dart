import 'package:flutter/widgets.dart';

import 'password_reset_scope.dart';

typedef PasswordResetScopeCreator = PasswordResetScope Function();

typedef PasswordResetScopeWidgetBuilder =
    Widget Function(PasswordResetScope scope);

class PasswordResetScopeBoundary extends StatefulWidget {
  final PasswordResetScopeCreator createScope;
  final PasswordResetScopeWidgetBuilder builder;

  const PasswordResetScopeBoundary({
    required this.createScope,
    required this.builder,
    super.key,
  });

  @override
  State<PasswordResetScopeBoundary> createState() =>
      _PasswordResetScopeBoundaryState();
}

class _PasswordResetScopeBoundaryState
    extends State<PasswordResetScopeBoundary> {
  late final PasswordResetScope _scope;
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
