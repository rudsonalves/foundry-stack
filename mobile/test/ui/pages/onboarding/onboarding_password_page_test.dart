import 'dart:async';

import 'package:foundry_stack_mobile/app/routing/routes.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/password/onboarding_password_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  Widget subject(
    _FakeOnboardingUsecase usecase, {
    VoidCallback? onStepChanged,
    VoidCallback? onEmailAlreadyRegistered,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: OnboardingPasswordPage(
          usecase: usecase,
          onStepChanged: onStepChanged ?? () {},
          onEmailAlreadyRegistered: onEmailAlreadyRegistered ?? () {},
        ),
      ),
    );
  }

  Future<void> enterValidPasswords(WidgetTester tester) async {
    final fields = find.byType(TextField);
    await tester.enterText(fields.at(0), 'Secret123');
    await tester.enterText(fields.at(1), 'Secret123');
    await tester.pump();
  }

  testWidgets('keeps creation disabled until both passwords are valid', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase();

    await tester.pumpWidget(subject(usecase));
    var button = tester.widget<ElevatedButton>(
      find.widgetWithText(ElevatedButton, 'Criar conta'),
    );
    expect(button.onPressed, isNull);

    await tester.enterText(find.byType(TextField).at(0), 'Secret123');
    await tester.enterText(find.byType(TextField).at(1), 'Different123');
    await tester.pump();

    button = tester.widget<ElevatedButton>(
      find.widgetWithText(ElevatedButton, 'Criar conta'),
    );
    expect(button.onPressed, isNull);
    expect(usecase.createCalls, 0);
  });

  testWidgets('updates the password requirement indicators', (tester) async {
    final usecase = _FakeOnboardingUsecase();

    await tester.pumpWidget(subject(usecase));

    expect(find.byIcon(Icons.circle_outlined), findsNWidgets(4));
    expect(find.byIcon(Icons.check), findsNothing);

    await tester.enterText(find.byType(TextField).at(0), 'Secret123');
    await tester.pump();

    expect(find.byIcon(Icons.check), findsNWidgets(3));
    expect(find.byIcon(Icons.circle_outlined), findsOneWidget);

    await tester.enterText(find.byType(TextField).at(1), 'Secret123');
    await tester.pump();

    final checks = tester.widgetList<Icon>(find.byIcon(Icons.check));
    expect(checks, hasLength(4));
    expect(checks.every((icon) => icon.color == Colors.green), isTrue);
    expect(find.byIcon(Icons.circle_outlined), findsNothing);
  });

  testWidgets('creates the account and replaces onboarding with login', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase();
    final router = GoRouter(
      initialLocation: '/password',
      routes: [
        GoRoute(
          path: '/password',
          builder: (context, state) => Scaffold(
            body: OnboardingPasswordPage(
              usecase: usecase,
              onStepChanged: () {},
              onEmailAlreadyRegistered: () {},
            ),
          ),
        ),
        GoRoute(
          path: AuthRoutes.login.routePath,
          name: AuthRoutes.login.routeName,
          builder: (context, state) => const Scaffold(body: Text('Login')),
        ),
      ],
    );
    addTearDown(router.dispose);

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await enterValidPasswords(tester);

    expect(find.text('Login'), findsNothing);
    expect(usecase.createCalls, 0);

    await tester.tap(find.text('Criar conta'));
    await tester.pumpAndSettle();

    expect(usecase.receivedPassword, 'Secret123');
    expect(usecase.createCalls, 1);
    expect(find.text('Login'), findsOneWidget);
  });

  testWidgets('prevents a second creation while the command is running', (
    tester,
  ) async {
    final completer = Completer<Result<Unit>>();
    final usecase = _FakeOnboardingUsecase()..createFuture = completer.future;

    await tester.pumpWidget(subject(usecase));
    await enterValidPasswords(tester);
    await tester.tap(find.text('Criar conta'));
    await tester.pump();

    final button = tester.widget<ElevatedButton>(
      find.widgetWithText(ElevatedButton, 'Criar conta'),
    );
    expect(button.onPressed, isNull);
    expect(usecase.createCalls, 1);

    completer.complete(
      const Failure(
        AppError(code: AppErrorCode.networkError, message: 'offline'),
      ),
    );
    await tester.pump();
  });

  testWidgets('reports email conflict and requests return to email', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase()
      ..createResult = const Failure(
        AppError(
          code: AppErrorCode.conflict,
          message: 'email in use',
          details: {'code': BackendErrorCodes.emailAlreadyRegistered},
        ),
      )
      ..stepAfterCreate = OnboardingStep.email;
    var conflicts = 0;

    await tester.pumpWidget(
      subject(usecase, onEmailAlreadyRegistered: () => conflicts++),
    );
    await enterValidPasswords(tester);
    await tester.tap(find.text('Criar conta'));
    await tester.pump();

    expect(conflicts, 1);
    expect(find.text('E-mail em uso'), findsOneWidget);
  });

  testWidgets('renders the recovered step after invalid verification', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase()
      ..createResult = const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'invalid verification',
          details: {'code': BackendErrorCodes.invalidEmailVerification},
        ),
      )
      ..stepAfterCreate = OnboardingStep.emailVerification;
    var stepChanges = 0;

    await tester.pumpWidget(
      subject(usecase, onStepChanged: () => stepChanges++),
    );
    await enterValidPasswords(tester);
    await tester.tap(find.text('Criar conta'));
    await tester.pump();

    expect(stepChanges, 1);
    expect(
      find.text('A verificação expirou. Enviamos um novo código.'),
      findsOneWidget,
    );
  });

  testWidgets('keeps password step and values after a recoverable failure', (
    tester,
  ) async {
    final usecase = _FakeOnboardingUsecase()
      ..createResult = const Failure(
        AppError(code: AppErrorCode.networkError, message: 'offline'),
      );
    var stepChanges = 0;

    await tester.pumpWidget(
      subject(usecase, onStepChanged: () => stepChanges++),
    );
    await enterValidPasswords(tester);
    await tester.tap(find.text('Criar conta'));
    await tester.pump();

    final fields = find.byType(TextField);
    expect(stepChanges, 0);
    expect(
      tester.widget<TextField>(fields.at(0)).controller?.text,
      'Secret123',
    );
    expect(
      tester.widget<TextField>(fields.at(1)).controller?.text,
      'Secret123',
    );
    expect(
      find.text('Não foi possível conectar ao servidor. Tente novamente.'),
      findsOneWidget,
    );
  });

  testWidgets('returns to the previous onboarding step', (tester) async {
    final usecase = _FakeOnboardingUsecase();
    var stepChanges = 0;

    await tester.pumpWidget(
      subject(usecase, onStepChanged: () => stepChanges++),
    );
    await tester.tap(find.text('Voltar'));
    await tester.pump();

    expect(usecase.returnCalls, 1);
    expect(stepChanges, 1);
  });
}

class _FakeOnboardingUsecase implements OnboardingUsecase {
  @override
  OnboardingRegistrationDraft? currentDraft = OnboardingRegistrationDraft(
    step: OnboardingStep.password,
    name: 'Ada',
    email: 'ada@example.com',
    expiresAt: DateTime.now().add(const Duration(hours: 24)),
  );

  Result<Unit> createResult = const Success(unit);
  Future<Result<Unit>>? createFuture;
  OnboardingStep? stepAfterCreate;
  String? receivedPassword;
  int createCalls = 0;
  int returnCalls = 0;

  @override
  AsyncResult<Unit> createAccount(String password) async {
    createCalls++;
    receivedPassword = password;

    final result = createFuture == null ? createResult : await createFuture!;
    final nextStep = stepAfterCreate;
    if (nextStep != null) {
      currentDraft = currentDraft!.copyWith(step: nextStep);
    }
    return result;
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() async {
    returnCalls++;
    currentDraft = currentDraft!.copyWith(step: OnboardingStep.email);
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> initialize() async =>
      Success(currentDraft!);

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) async =>
      Success(currentDraft!);

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(
    String email,
  ) async => Success(currentDraft!);

  @override
  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) async => Success(currentDraft!);

  @override
  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() async =>
      Success(currentDraft!);

  @override
  AsyncResult<Unit> cancelRegistration() async => const Success(unit);
}
