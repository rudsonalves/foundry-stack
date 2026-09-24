import 'package:go_router/go_router.dart';

import '/app/dependencies/app_container.dart';
import '/app/dependencies/onboarding_scope_boundary.dart';
import '/app/dependencies/password_reset_scope_boundary.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';
import '/ui/pages/home/home_page.dart';
import '/ui/pages/home/viewmodel/home_viewmodel.dart';
import '/ui/pages/splash/splash_page.dart';
import '/ui/pages/splash/viewmodel/splash_viewmodel.dart';
import '../../../domain/usecases/onboarding/onboarding_usecase.dart';
import '../../../ui/pages/onboarding/onboarding_page.dart';
import '../../../ui/pages/onboarding/viewmodel/onboarding_coordinator_viewmodel.dart';
import '../../../ui/pages/password_reset/password_reset_page.dart';
import '../routes.dart';

List<RouteBase> baseRoutes({
  required AppContainer appContainer,
  required OnboardingScopeCreator createOnboardingScope,
  required PasswordResetScopeCreator createPasswordResetScope,
}) => [
  GoRoute(
    path: BaseRoutes.home.routePath,
    name: BaseRoutes.home.routeName,
    builder: (context, state) =>
        HomePage(viewmodel: appContainer.get<HomeViewmodel>()),
  ),

  GoRoute(
    path: BaseRoutes.splash.routePath,
    name: BaseRoutes.splash.routeName,
    builder: (context, state) =>
        SplashPage(viewmodel: appContainer.get<SplashViewmodel>()),
  ),

  GoRoute(
    path: BaseRoutes.register.routePath,
    name: BaseRoutes.register.routeName,
    builder: (context, state) => OnboardingScopeBoundary(
      createScope: createOnboardingScope,
      builder: (scope) {
        final usecase = scope.get<OnboardingUsecase>();

        return OnboardingPage(
          viewmodel: OnboardingCoordinatorViewmodel(usecase),
          usecase: usecase,
        );
      },
    ),
  ),
  GoRoute(
    path: BaseRoutes.forgotPassword.routePath,
    name: BaseRoutes.forgotPassword.routeName,
    builder: (context, state) => PasswordResetScopeBoundary(
      createScope: createPasswordResetScope,
      builder: (scope) => PasswordResetPage(
        usecase: scope.get<PasswordResetUsecase>(),
      ),
    ),
  ),
];
