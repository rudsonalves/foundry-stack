import 'package:auto_injector/auto_injector.dart';

import '/core/services/client_http/client/rest_client.dart';
import '/core/services/secure_storage/local_secure_storage.dart';
import 'bindings/onboarding_bindings.dart';
import 'onboarding_dependencies.dart';

class OnboardingScope {
  final OnboardingDependencies _dependencies;
  final AutoInjector _injector;
  final void Function()? _onDispose;

  OnboardingScope(
    this._dependencies, {
    AutoInjector? injector,
    void Function()? onDispose,
  }) : _injector = injector ?? AutoInjector(tag: 'onboarding'),
       _onDispose = onDispose;

  bool _initialized = false;
  bool _disposed = false;

  void initialize() {
    if (_disposed) {
      throw StateError('OnboardingScope has been disposed.');
    }
    if (_initialized) {
      throw StateError('OnboardingScope has already been initialized.');
    }

    _injector
      ..addInstance<RestClient>(_dependencies.restClient)
      ..addInstance<LocalSecureStorage>(
        _dependencies.localSecureStorage,
      );
    registerOnboardingBindings(_injector);
    _injector.commit();

    _initialized = true;
  }

  T get<T>() {
    if (!_initialized || _disposed) {
      throw StateError('OnboardingScope is not available.');
    }

    return _injector.get<T>();
  }

  void dispose() {
    if (_disposed) return;

    _disposed = true;
    _initialized = false;
    _injector.dispose();
    _onDispose?.call();
  }
}
