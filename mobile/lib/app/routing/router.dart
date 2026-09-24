import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';

import '/app/dependencies/app_container.dart';
import '/app/dependencies/onboarding_scope_boundary.dart';
import '/app/dependencies/password_reset_scope_boundary.dart';
import 'extra_codec.dart';
import 'route_observer.dart';
import 'routes.dart';
import 'routes/auth_routes.dart';
import 'routes/base_routes.dart';

GoRouter router({
  required AppContainer appContainer,
  required OnboardingScopeCreator createOnboardingScope,
  required PasswordResetScopeCreator createPasswordResetScope,
}) => GoRouter(
  initialLocation: BaseRoutes.splash.routePath,
  debugLogDiagnostics: kDebugMode,
  observers: [routeObserver],
  extraCodec: const ExtraCodec(),
  routes: [
    ...baseRoutes(
      appContainer: appContainer,
      createOnboardingScope: createOnboardingScope,
      createPasswordResetScope: createPasswordResetScope,
    ),
    ...authRoutes(appContainer: appContainer),
  ],
);
