import 'package:auto_injector/auto_injector.dart';

import '/data/repositories/auth/auth_repository.dart';
import '/data/repositories/auth/auth_repository_impl.dart';
import '/data/repositories/onboarding/onboarding_draft_repository.dart';
import '/data/repositories/onboarding/onboarding_draft_repository_impl.dart';
import '/data/services/apis/auth/auth_service.dart';
import '/domain/usecases/bootstrap/bootstrap_usecase.dart';
import '/ui/pages/auth/login/viewmodel/login_viewmodel.dart';
import '/ui/pages/home/viewmodel/home_viewmodel.dart';
import '/ui/pages/splash/viewmodel/splash_viewmodel.dart';

void registerAppBindings(AutoInjector injector) {
  injector
    ..addSingleton<AuthService>(AuthService.new)
    ..add<AuthRepository>(AuthRepositoryImpl.new)
    ..add<OnboardingDraftRepository>(OnboardingDraftRepositoryImpl.new)
    ..add<BootstrapUsecase>(BootstrapUsecase.new)
    ..add<HomeViewmodel>(HomeViewmodel.new)
    ..add<SplashViewmodel>(SplashViewmodel.new)
    ..add<LoginViewmodel>(LoginViewmodel.new);
}
