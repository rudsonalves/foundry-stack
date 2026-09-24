import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/email/onboarding_email_page.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/user_name/onboarding_user_name_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final expiresAt = DateTime.utc(2026, 9, 22);

  Widget page(Widget child) => MaterialApp(home: Scaffold(body: child));

  group('OnboardingUserNamePage', () {
    testWidgets('restores the name and advances with the normalized value', (
      tester,
    ) async {
      final usecase = _FakeOnboardingUsecase(
        OnboardingRegistrationDraft(
          step: OnboardingStep.userName,
          name: '  Ada Lovelace  ',
          expiresAt: expiresAt,
        ),
      );
      var stepChanges = 0;

      await tester.pumpWidget(
        page(
          OnboardingUserNamePage(
            usecase: usecase,
            onStepChanged: () => stepChanges++,
          ),
        ),
      );

      expect(find.widgetWithText(TextField, 'Ada Lovelace'), findsOneWidget);

      await tester.tap(find.text('Continuar'));
      await tester.pump();

      expect(usecase.receivedName, 'Ada Lovelace');
      expect(stepChanges, 1);
    });

    testWidgets('keeps continue disabled for an empty name', (tester) async {
      final usecase = _FakeOnboardingUsecase(
        OnboardingRegistrationDraft(
          step: OnboardingStep.userName,
          expiresAt: expiresAt,
        ),
      );

      await tester.pumpWidget(
        page(
          OnboardingUserNamePage(
            usecase: usecase,
            onStepChanged: () {},
          ),
        ),
      );

      final button = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Continuar'),
      );
      expect(button.onPressed, isNull);
      expect(usecase.receivedName, isNull);
    });
  });

  group('OnboardingEmailPage', () {
    testWidgets('restores the email and advances successfully', (tester) async {
      final usecase = _FakeOnboardingUsecase(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          email: '  ada@example.com  ',
          expiresAt: expiresAt,
        ),
      );
      var stepChanges = 0;

      await tester.pumpWidget(
        page(
          OnboardingEmailPage(
            usecase: usecase,
            onStepChanged: () => stepChanges++,
            emailAlreadyRegistered: false,
            onEmailEdited: () {},
          ),
        ),
      );

      expect(find.widgetWithText(TextField, 'ada@example.com'), findsOneWidget);

      await tester.tap(find.text('Continuar'));
      await tester.pump();

      expect(usecase.receivedEmail, 'ada@example.com');
      expect(stepChanges, 1);
    });

    testWidgets('rejects an invalid email without calling the use case', (
      tester,
    ) async {
      final usecase = _FakeOnboardingUsecase(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          expiresAt: expiresAt,
        ),
      );

      await tester.pumpWidget(
        page(
          OnboardingEmailPage(
            usecase: usecase,
            onStepChanged: () {},
            emailAlreadyRegistered: false,
            onEmailEdited: () {},
          ),
        ),
      );
      await tester.enterText(find.byType(TextField), 'invalid');
      await tester.pump();

      final button = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Continuar'),
      );
      expect(button.onPressed, isNull);
      expect(usecase.receivedEmail, isNull);
    });

    testWidgets('shows email in use and preserves the entered address', (
      tester,
    ) async {
      final usecase =
          _FakeOnboardingUsecase(
              OnboardingRegistrationDraft(
                step: OnboardingStep.email,
                name: 'Ada',
                expiresAt: expiresAt,
              ),
            )
            ..advanceEmailResult = const Failure(
              AppError(
                code: AppErrorCode.conflict,
                message: 'conflict',
                details: {'code': BackendErrorCodes.emailAlreadyRegistered},
              ),
            );

      await tester.pumpWidget(
        page(
          OnboardingEmailPage(
            usecase: usecase,
            onStepChanged: () {},
            emailAlreadyRegistered: false,
            onEmailEdited: () {},
          ),
        ),
      );
      await tester.enterText(find.byType(TextField), 'used@example.com');
      await tester.pump();
      await tester.tap(find.text('Continuar'));
      await tester.pumpAndSettle();

      expect(find.text('E-mail em uso'), findsOneWidget);
      expect(find.text('Cancelar'), findsOneWidget);
      expect(find.text('Voltar'), findsNothing);
      expect(find.text('used@example.com'), findsOneWidget);
    });

    testWidgets('shows a recoverable network failure without losing email', (
      tester,
    ) async {
      final usecase =
          _FakeOnboardingUsecase(
              OnboardingRegistrationDraft(
                step: OnboardingStep.email,
                name: 'Ada',
                expiresAt: expiresAt,
              ),
            )
            ..advanceEmailResult = const Failure(
              AppError(code: AppErrorCode.networkError, message: 'offline'),
            );

      await tester.pumpWidget(
        page(
          OnboardingEmailPage(
            usecase: usecase,
            onStepChanged: () {},
            emailAlreadyRegistered: false,
            onEmailEdited: () {},
          ),
        ),
      );
      await tester.enterText(find.byType(TextField), 'ada@example.com');
      await tester.pump();
      await tester.tap(find.text('Continuar'));
      await tester.pumpAndSettle();

      expect(
        find.text('Não foi possível conectar ao servidor. Tente novamente.'),
        findsOneWidget,
      );
      expect(find.text('ada@example.com'), findsOneWidget);
    });

    testWidgets('returns to the previous step', (tester) async {
      final usecase = _FakeOnboardingUsecase(
        OnboardingRegistrationDraft(
          step: OnboardingStep.email,
          name: 'Ada',
          expiresAt: expiresAt,
        ),
      );
      var stepChanges = 0;

      await tester.pumpWidget(
        page(
          OnboardingEmailPage(
            usecase: usecase,
            onStepChanged: () => stepChanges++,
            emailAlreadyRegistered: false,
            onEmailEdited: () {},
          ),
        ),
      );
      await tester.tap(find.text('Voltar'));
      await tester.pump();

      expect(usecase.returnCalls, 1);
      expect(stepChanges, 1);
    });
  });
}

class _FakeOnboardingUsecase implements OnboardingUsecase {
  _FakeOnboardingUsecase(this.currentDraft);

  @override
  OnboardingRegistrationDraft? currentDraft;

  Result<OnboardingRegistrationDraft>? advanceEmailResult;
  String? receivedName;
  String? receivedEmail;
  int returnCalls = 0;
  int cancelCalls = 0;

  @override
  AsyncResult<OnboardingRegistrationDraft> initialize() async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) async {
    receivedName = name;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(
    String email,
  ) async {
    receivedEmail = email;
    return advanceEmailResult ?? Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() async {
    returnCalls++;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<Unit> cancelRegistration() async {
    cancelCalls++;
    return const Success(unit);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<Unit> createAccount(String password) async {
    return const Success(unit);
  }
}
