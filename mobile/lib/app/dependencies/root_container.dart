import '/core/config/core_container.dart';
import 'app_container.dart';
import 'app_dependencies.dart';
import 'onboarding_dependencies.dart';
import 'onboarding_scope.dart';
import 'password_reset_dependencies.dart';
import 'password_reset_scope.dart';

class RootContainer {
  final CoreContainer _coreContainer;

  RootContainer({CoreContainer? coreContainer})
    : _coreContainer = coreContainer ?? CoreContainer();

  AppContainer? _appContainer;

  OnboardingDependencies? _onboardingDependencies;
  final Set<OnboardingScope> _activeOnboardingScopes = {};

  PasswordResetDependencies? _passwordResetDependencies;
  final Set<PasswordResetScope> _activePasswordResetScopes = {};

  bool _coreInitialized = false;
  bool _disposed = false;

  void initializeCore({String? baseUrl}) {
    _ensureNotDisposed();
    if (_coreInitialized) {
      throw StateError('RootContainer core has already been initialized.');
    }

    final core = _coreContainer.initialize(baseUrl: baseUrl);

    _onboardingDependencies = OnboardingDependencies(
      restClient: core.restClient,
      localSecureStorage: core.localSecureStorage,
    );
    _passwordResetDependencies = PasswordResetDependencies(
      restClient: core.restClient,
    );

    _coreInitialized = true;
  }

  void initializeApp() {
    _ensureCoreAvailable();
    if (_appContainer != null) {
      throw StateError('RootContainer app has already been initialized.');
    }

    final core = _coreContainer.dependencies;
    final app = AppContainer(
      AppDependencies(
        restClient: core.restClient,
        authTokenStore: core.authTokenStore,
        localSecureStorage: core.localSecureStorage,
      ),
    )..initialize();
    _appContainer = app;
  }

  AppContainer get appContainer {
    _ensureNotDisposed();
    final app = _appContainer;
    if (app == null) {
      throw StateError('RootContainer app is not available.');
    }
    return app;
  }

  OnboardingScope createOnboardingScope() {
    _ensureCoreAvailable();

    late final OnboardingScope scope;
    scope = OnboardingScope(
      _onboardingDependencies!,
      onDispose: () => _activeOnboardingScopes.remove(scope),
    );
    _activeOnboardingScopes.add(scope);
    scope.initialize();
    return scope;
  }

  PasswordResetScope createPasswordResetScope() {
    _ensureCoreAvailable();

    late final PasswordResetScope scope;
    scope = PasswordResetScope(
      _passwordResetDependencies!,
      onDispose: () => _activePasswordResetScopes.remove(scope),
    );

    _activePasswordResetScopes.add(scope);
    scope.initialize();

    return scope;
  }

  void dispose() {
    if (_disposed) return;

    _disposed = true;
    for (final scope in _activeOnboardingScopes.toList()) {
      scope.dispose();
    }
    _activeOnboardingScopes.clear();

    for (final scope in _activePasswordResetScopes.toList()) {
      scope.dispose();
    }
    _activePasswordResetScopes.clear();

    _appContainer?.dispose();
    _coreContainer.dispose();

    _onboardingDependencies = null;
    _passwordResetDependencies = null;

    _appContainer = null;
    _coreInitialized = false;
  }

  void _ensureCoreAvailable() {
    _ensureNotDisposed();
    if (!_coreInitialized) {
      throw StateError('RootContainer core is not available.');
    }
  }

  void _ensureNotDisposed() {
    if (_disposed) {
      throw StateError('RootContainer has been disposed.');
    }
  }
}
