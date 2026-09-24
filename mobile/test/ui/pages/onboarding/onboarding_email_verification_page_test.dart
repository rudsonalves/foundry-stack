import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/email_verification_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/email_verification/onboarding_email_verification_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  Widget subject(_FakeOnboardingUsecase usecase, VoidCallback onStepChanged) {
    return MaterialApp(
      home: Scaffold(
        body: OnboardingEmailVerificationPage(
          usecase: usecase,
          onStepChanged: onStepChanged,
        ),
      ),
    );
  }

  testWidgets('does not confirm or advance when six digits are entered', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase();
    var stepChanges = 0;

    await tester.pumpWidget(subject(usecase, () => stepChanges++));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.pump();

    expect(usecase.receivedToken, isNull);
    expect(stepChanges, 0);
  });

  testWidgets('confirms and advances only after tapping continue', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase();
    var stepChanges = 0;

    await tester.pumpWidget(subject(usecase, () => stepChanges++));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.pump();
    await tester.tap(find.text('Continuar'));
    await tester.pump();

    expect(usecase.receivedToken, '012345');
    expect(stepChanges, 1);
    expect(
      find.textContaining('ada@example.com', findRichText: true),
      findsOneWidget,
    );
  });

  testWidgets('clears token and starts a new countdown after resend', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase();

    await tester.pumpWidget(subject(usecase, () {}));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.tap(find.text('Reenviar código'));
    await tester.pump();

    final textField = tester.widget<TextField>(find.byType(TextField));
    expect(usecase.resendCalls, 1);
    expect(textField.controller?.text, isEmpty);
    expect(find.textContaining('Reenviar código em'), findsOneWidget);
  });

  testWidgets('preserves editable token when resend fails', (tester) async {
    final usecase = _FakeOnboardingUsecase()
      ..resendResult = const Failure(
        AppError(
          code: AppErrorCode.conflict,
          message: 'rate limited',
          details: {'code': BackendErrorCodes.rateLimitExceeded},
        ),
      );

    await tester.pumpWidget(subject(usecase, () {}));
    await tester.enterText(find.byType(TextField), '012345');
    await tester.tap(find.text('Reenviar código'));
    await tester.pump();

    final textField = tester.widget<TextField>(find.byType(TextField));
    expect(textField.controller?.text, '012345');
    expect(
      find.text('Muitas tentativas. Aguarde antes de tentar novamente.'),
      findsOneWidget,
    );
  });

  testWidgets('cancels the countdown timer when the page is disposed', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase(
      challenge: EmailVerificationChallenge(
        verificationId: 'active',
        codeExpiresAt: DateTime.now().add(const Duration(minutes: 10)),
        resendAvailableAt: DateTime.now().add(const Duration(minutes: 1)),
      ),
    );

    await tester.pumpWidget(subject(usecase, () {}));
    expect(find.textContaining('Reenviar código em'), findsOneWidget);

    await tester.pumpWidget(const MaterialApp(home: SizedBox()));
    await tester.pump(const Duration(seconds: 2));

    expect(tester.takeException(), isNull);
  });
}

class _FakeOnboardingUsecase implements OnboardingUsecase {
  _FakeOnboardingUsecase({EmailVerificationChallenge? challenge}) {
    currentDraft = OnboardingRegistrationDraft(
      step: OnboardingStep.emailVerification,
      name: 'Ada',
      email: 'ada@example.com',
      challenge: challenge,
      expiresAt: DateTime.now().add(const Duration(hours: 24)),
    );
  }

  @override
  OnboardingRegistrationDraft? currentDraft;

  String? receivedToken;
  int resendCalls = 0;
  Result<OnboardingRegistrationDraft>? resendResult;

  @override
  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) async {
    receivedToken = otp;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() async {
    resendCalls++;

    if (resendResult case final result?) return result;

    currentDraft = currentDraft!.copyWith(
      challenge: EmailVerificationChallenge(
        verificationId: 'replacement',
        codeExpiresAt: DateTime.now().add(const Duration(minutes: 10)),
        resendAvailableAt: DateTime.now().add(const Duration(minutes: 1)),
      ),
    );

    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> initialize() async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(
    String email,
  ) async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<Unit> cancelRegistration() async => const Success(unit);

  @override
  AsyncResult<Unit> createAccount(String password) async => const Success(unit);
}
