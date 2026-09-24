import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_draft_load.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('draft uses the current schema version by default', () {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.userName,
      expiresAt: DateTime.utc(2026, 9, 19),
    );

    expect(
      draft.schemaVersion,
      OnboardingRegistrationDraft.currentSchemaVersion,
    );
  });

  test('draft load distinguishes absent and present values', () {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.email,
      expiresAt: DateTime.utc(2026, 9, 19),
    );
    const absent = OnboardingDraftLoad.absent();
    final present = OnboardingDraftLoad.present(draft);

    expect(absent.hasDraft, isFalse);
    expect(absent.draft, isNull);
    expect(present.hasDraft, isTrue);
    expect(identical(present.draft, draft), isTrue);
  });

  test('draft expires at the configured instant', () {
    final expiresAt = DateTime.utc(2026, 9, 19);
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.userName,
      expiresAt: expiresAt,
    );

    expect(OnboardingRegistrationDraft.validity, const Duration(hours: 24));
    expect(
      draft.isExpiredAt(expiresAt.subtract(const Duration(microseconds: 1))),
      isFalse,
    );
    expect(draft.isExpiredAt(expiresAt), isTrue);
    expect(
      draft.isExpiredAt(expiresAt.add(const Duration(microseconds: 1))),
      isTrue,
    );
  });
}
