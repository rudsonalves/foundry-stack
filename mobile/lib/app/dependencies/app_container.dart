import 'package:auto_injector/auto_injector.dart';

import '/core/services/auth_token_store/auth_token_store.dart';
import '/core/services/client_http/client/rest_client.dart';
import '../../core/services/secure_storage/local_secure_storage.dart';
import 'app_dependencies.dart';
import 'bindings/app_bindings.dart';

class AppContainer {
  final AppDependencies _dependencies;
  final AutoInjector _injector;

  AppContainer(
    this._dependencies, {
    AutoInjector? injector,
  }) : _injector = injector ?? AutoInjector(tag: 'app');

  bool _initialized = false;
  bool _disposed = false;

  void initialize() {
    if (_disposed) {
      throw StateError('AppContainer has been disposed.');
    }
    if (_initialized) {
      throw StateError('AppContainer has already been initialized.');
    }

    _injector
      ..addInstance<RestClient>(_dependencies.restClient)
      ..addInstance<AuthTokenStore>(_dependencies.authTokenStore)
      ..addInstance<LocalSecureStorage>(_dependencies.localSecureStorage);
    registerAppBindings(_injector);
    _injector.commit();

    _initialized = true;
  }

  T get<T>() {
    if (!_initialized || _disposed) {
      throw StateError('AppContainer is not available.');
    }
    return _injector.get<T>();
  }

  void dispose() {
    if (_disposed) return;

    _disposed = true;
    _initialized = false;
    _injector.dispose();
  }
}
