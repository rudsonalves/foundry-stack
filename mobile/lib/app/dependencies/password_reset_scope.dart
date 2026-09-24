import 'package:auto_injector/auto_injector.dart';

import '/core/services/client_http/client/rest_client.dart';
import 'bindings/password_reset_bindings.dart';
import 'password_reset_dependencies.dart';

class PasswordResetScope {
  final PasswordResetDependencies _dependencies;
  final AutoInjector _injector;
  final void Function()? _onDispose;

  PasswordResetScope(
    this._dependencies, {
    AutoInjector? injector,
    void Function()? onDispose,
  }) : _injector = injector ?? AutoInjector(tag: 'password-reset'),
       _onDispose = onDispose;

  bool _initialized = false;
  bool _disposed = false;

  void initialize() {
    if (_disposed) {
      throw StateError('PasswordResetScope has been disposed.');
    }
    if (_initialized) {
      throw StateError(
        'PasswordResetScope has already been initialized.',
      );
    }

    _injector.addInstance<RestClient>(_dependencies.restClient);
    registerPasswordResetBindings(_injector);
    _injector.commit();

    _initialized = true;
  }

  T get<T>() {
    if (!_initialized || _disposed) {
      throw StateError('PasswordResetScope is not available.');
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
