import 'package:foundry_stack_mobile/app/dependencies/password_reset_scope.dart';
import 'package:foundry_stack_mobile/app/dependencies/password_reset_scope_boundary.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('creates one scope and keeps it across rebuilds', (tester) async {
    final factory = _FakePasswordResetScopeCreator();
    var journeyBuilds = 0;

    Widget app() => MaterialApp(
      home: PasswordResetScopeBoundary(
        createScope: factory.create,
        builder: (scope) {
          journeyBuilds++;
          return const Text('Password reset');
        },
      ),
    );

    await tester.pumpWidget(app());
    await tester.pumpWidget(app());

    expect(factory.createCalls, 1);
    expect(journeyBuilds, 1);
    expect(factory.scopes.single.disposeCalls, 0);
  });

  testWidgets('disposes the scope when the boundary is removed', (
    tester,
  ) async {
    final factory = _FakePasswordResetScopeCreator();

    await tester.pumpWidget(
      MaterialApp(
        home: PasswordResetScopeBoundary(
          createScope: factory.create,
          builder: (scope) => const Text('Password reset'),
        ),
      ),
    );
    final scope = factory.scopes.single;

    await tester.pumpWidget(const MaterialApp(home: Text('Login')));

    expect(scope.disposeCalls, 1);
  });
}

final class _FakePasswordResetScopeCreator {
  final scopes = <_FakePasswordResetScope>[];
  int createCalls = 0;

  PasswordResetScope create() {
    createCalls++;
    final scope = _FakePasswordResetScope();
    scopes.add(scope);
    return scope;
  }
}

final class _FakePasswordResetScope implements PasswordResetScope {
  int disposeCalls = 0;

  @override
  void initialize() {}

  @override
  T get<T>() => throw UnimplementedError();

  @override
  void dispose() => disposeCalls++;
}
