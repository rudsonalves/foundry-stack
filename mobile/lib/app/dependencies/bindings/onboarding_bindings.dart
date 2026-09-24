import 'package:auto_injector/auto_injector.dart';

import '/data/repositories/email_verification/email_verification_repository.dart';
import '/data/repositories/email_verification/email_verification_repository_impl.dart';
import '/data/repositories/onboarding/onboarding_draft_repository.dart';
import '/data/repositories/onboarding/onboarding_draft_repository_impl.dart';
import '/data/repositories/user/user_repository.dart';
import '/data/repositories/user/user_repository_impl.dart';
import '/data/services/apis/email_verification/email_verification_service.dart';
import '/data/services/apis/user/user_service.dart';
import '/domain/usecases/onboarding/onboarding_usecase.dart';

void registerOnboardingBindings(AutoInjector injector) {
  injector
    ..add<EmailVerificationService>(EmailVerificationService.new)
    ..add<EmailVerificationRepository>(EmailVerificationRepositoryImpl.new)
    ..add<OnboardingDraftRepository>(OnboardingDraftRepositoryImpl.new)
    ..add<UserService>(UserService.new)
    ..add<UserRepository>(UserRepositoryImpl.new)
    ..addSingleton<OnboardingUsecase>(OnboardingUsecase.new);
}
