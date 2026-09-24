import '/core/result/result.dart';
import '/domain/common/onboarding/models/onboarding_draft_load.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';

abstract class OnboardingDraftRepository {
  AsyncResult<OnboardingDraftLoad> load();

  AsyncResult<Unit> save(OnboardingRegistrationDraft draft);

  AsyncResult<Unit> delete();
}
