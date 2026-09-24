import 'package:auto_injector/auto_injector.dart';

import '../services/auth_token_store/auth_token_store.dart';
import '../services/client_http/client/rest_client.dart';
import 'core_bindings.dart';
import '../services/secure_storage/local_secure_storage.dart';
import 'core_dependencies.dart';

class CoreContainer {
  final AutoInjector _injector;

  CoreContainer({AutoInjector? injector})
    : _injector = injector ?? AutoInjector(tag: 'core');

  CoreDependencies? _dependencies;
  bool _disposed = false;

  CoreDependencies initialize({String? baseUrl}) {
    if (_disposed) {
      throw StateError('CoreContainer has been disposed.');
    }
    if (_dependencies != null) {
      throw StateError('CoreContainer has already been initialized.');
    }

    registerCoreBindings(_injector, baseUrl: baseUrl);
    _injector.commit();

    return _dependencies = CoreDependencies(
      restClient: _injector.get<RestClient>(),
      localSecureStorage: _injector.get<LocalSecureStorage>(),
      authTokenStore: _injector.get<AuthTokenStore>(),
    );
  }

  CoreDependencies get dependencies {
    final dependencies = _dependencies;
    if (dependencies == null || _disposed) {
      throw StateError('CoreContainer is not available.');
    }
    return dependencies;
  }

  void dispose() {
    if (_disposed) return;

    _disposed = true;
    _dependencies = null;
    _injector.dispose();
  }
}
