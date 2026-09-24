import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:material_ui/material_ui.dart';

void main() {
  testWidgets(
    'builder route creates a MaterialPage with decoupled MaterialApp',
    (tester) async {
      final router = GoRouter(
        routes: [
          GoRoute(
            path: '/',
            builder: (context, state) => const Scaffold(
              body: Text('Home'),
            ),
          ),
          GoRoute(
            path: '/details',
            builder: (context, state) => const Scaffold(
              body: Text('Details'),
            ),
          ),
        ],
      );
      addTearDown(router.dispose);

      await tester.pumpWidget(MaterialApp.router(routerConfig: router));

      router.go('/details');
      await tester.pumpAndSettle();

      final context = tester.element(find.text('Details'));
      final page = ModalRoute.of(context)?.settings;

      expect(page, isA<MaterialPage<void>>());
      expect(page, isNot(isA<NoTransitionPage<void>>()));
    },
  );
}
