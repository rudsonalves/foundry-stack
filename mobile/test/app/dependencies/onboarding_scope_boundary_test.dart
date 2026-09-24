import 'package:foundry_stack_mobile/app/dependencies/onboarding_scope.dart';
import 'package:foundry_stack_mobile/app/dependencies/onboarding_scope_boundary.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  testWidgets('creates one scope and reuses the journey across rebuilds', (
    tester,
  ) async {
    final factory = _FakeOnboardingScopeCreator();
    var builds = 0;

    Widget app() => MaterialApp(
      home: OnboardingScopeBoundary(
        createScope: factory.create,
        builder: (scope) {
          builds++;
          return const Text('Onboarding');
        },
      ),
    );

    await tester.pumpWidget(app());
    await tester.pumpWidget(app());

    expect(factory.createCalls, 1);
    expect(builds, 1);
    expect(factory.scopes.single.disposeCalls, 0);
  });

  testWidgets('disposes the scope when removed by external navigation', (
    tester,
  ) async {
    final factory = _FakeOnboardingScopeCreator();
    final router = GoRouter(
      initialLocation: '/onboarding',
      routes: [
        GoRoute(
          path: '/onboarding',
          builder: (context, state) => OnboardingScopeBoundary(
            createScope: factory.create,
            builder: (scope) => const Scaffold(body: Text('Onboarding')),
          ),
        ),
        GoRoute(
          path: '/outside',
          builder: (context, state) => const Scaffold(body: Text('Outside')),
        ),
      ],
    );
    addTearDown(router.dispose);

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    final scope = factory.scopes.single;

    router.go('/outside');
    await tester.pumpAndSettle();

    expect(find.text('Outside'), findsOneWidget);
    expect(scope.disposeCalls, 1);
  });

  testWidgets('keeps the same scope across internal navigation', (
    tester,
  ) async {
    final factory = _FakeOnboardingScopeCreator();

    await tester.pumpWidget(
      MaterialApp(
        home: OnboardingScopeBoundary(
          createScope: factory.create,
          builder: (scope) => Navigator(
            onGenerateRoute: (settings) => MaterialPageRoute<void>(
              builder: (context) => Scaffold(
                body: TextButton(
                  onPressed: () => Navigator.of(context).push<void>(
                    MaterialPageRoute<void>(
                      builder: (_) => const Text('Second step'),
                    ),
                  ),
                  child: const Text('Next step'),
                ),
              ),
            ),
          ),
        ),
      ),
    );
    final scope = factory.scopes.single;

    await tester.tap(find.text('Next step'));
    await tester.pumpAndSettle();

    expect(find.text('Second step'), findsOneWidget);
    expect(factory.createCalls, 1);
    expect(scope.disposeCalls, 0);
  });

  testWidgets('does not expose the scope to the rendered page', (tester) async {
    final factory = _FakeOnboardingScopeCreator();

    await tester.pumpWidget(
      MaterialApp(
        home: OnboardingScopeBoundary(
          createScope: factory.create,
          builder: (scope) => const _PageWithConstructorDependency(
            value: 'resolved at boundary',
          ),
        ),
      ),
    );

    expect(find.text('resolved at boundary'), findsOneWidget);
  });
}

final class _PageWithConstructorDependency extends StatelessWidget {
  final String value;

  const _PageWithConstructorDependency({required this.value});

  @override
  Widget build(BuildContext context) => Text(value);
}

class _FakeOnboardingScopeCreator {
  final scopes = <_FakeOnboardingScope>[];
  int createCalls = 0;

  OnboardingScope create() {
    createCalls++;
    final scope = _FakeOnboardingScope();
    scopes.add(scope);
    return scope;
  }
}

final class _FakeOnboardingScope implements OnboardingScope {
  int disposeCalls = 0;

  @override
  void initialize() {}

  @override
  T get<T>() => throw UnimplementedError();

  @override
  void dispose() => disposeCalls++;
}
