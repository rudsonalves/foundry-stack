import 'package:auto_injector/auto_injector.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'sibling scopes resolve each other through a shared upward parent',
    () {
      final core = AutoInjector(tag: 'core');
      final app = AutoInjector(tag: 'app')
        ..addSingleton<_AppOnly>(_AppOnly.new);
      final onboarding = AutoInjector(tag: 'onboarding');

      core
        ..addInjector(app, resolveUpward: true)
        ..addInjector(onboarding, resolveUpward: true)
        ..commit();

      addTearDown(core.disposeRecursive);

      expect(onboarding.get<_AppOnly>(), isA<_AppOnly>());
    },
  );
}

final class _AppOnly {}
