import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/password_reset_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  testWidgets('completes the whole journey with a fake repository', (
    tester,
  ) async {
    final repository = _Repository();
    final usecase = PasswordResetUsecase(repository: repository);
    final router = _router(usecase);
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));

    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'user@example.com');
    await tester.pump();
    await tester.tap(find.text('Enviar código'));
    await tester.pumpAndSettle();

    expect(find.text('Confirme o código'), findsOneWidget);
    ScaffoldMessenger.of(
      tester.element(find.text('Confirme o código')),
    ).hideCurrentSnackBar();
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), '012345');
    await tester.pump();
    await tester.tap(find.text('Confirmar código'));
    await tester.pumpAndSettle();

    expect(find.text('Crie uma nova senha'), findsOneWidget);
    final passwordFields = find.byType(TextField);
    await tester.enterText(passwordFields.at(0), 'Password1');
    await tester.enterText(passwordFields.at(1), 'Password1');
    await tester.pump();
    await tester.tap(find.text('Alterar senha'));
    await tester.pumpAndSettle();

    expect(find.text('Login'), findsOneWidget);
    expect(repository.requestCalls, 1);
    expect(repository.confirmCalls, 1);
    expect(repository.completeCalls, 1);
    expect(repository.authenticationCalls, 0);
  });

  testWidgets('starts again at email after leaving an active journey', (
    tester,
  ) async {
    final repository = _Repository();
    final usecases = <PasswordResetUsecase>[];
    final router = _routerWithFactory(() {
      final usecase = PasswordResetUsecase(repository: repository);
      usecases.add(usecase);
      return usecase;
    });
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));

    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'user@example.com');
    await tester.pump();
    await tester.tap(find.text('Enviar código'));
    await tester.pumpAndSettle();
    expect(find.text('Confirme o código'), findsOneWidget);

    await tester.tap(find.byTooltip('Fechar recuperação'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, 'Sair'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();

    expect(find.text('Recupere sua senha'), findsOneWidget);
    expect(find.text('Confirme o código'), findsNothing);
    expect(usecases, hasLength(2));
    expect(usecases.last.email, isNull);
  });

  testWidgets('leaves directly from the email step', (tester) async {
    final usecase = PasswordResetUsecase(repository: _Repository());
    final router = _router(usecase);
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Fechar recuperação'));
    await tester.pumpAndSettle();

    expect(find.text('Login'), findsOneWidget);
    expect(find.text('Sair da recuperação?'), findsNothing);
  });

  testWidgets('asks before leaving after the first step', (tester) async {
    final repository = _Repository();
    final usecase = PasswordResetUsecase(repository: repository);
    await usecase.requestPasswordReset('user@example.com');
    final router = _router(usecase);
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Fechar recuperação'));
    await tester.pumpAndSettle();

    expect(find.text('Sair da recuperação?'), findsOneWidget);
    await tester.tap(find.widgetWithText(FilledButton, 'Sair'));
    await tester.pumpAndSettle();
    expect(find.text('Login'), findsOneWidget);
  });

  testWidgets('applies confirmation to the system back action', (
    tester,
  ) async {
    final repository = _Repository();
    final usecase = PasswordResetUsecase(repository: repository);
    await usecase.requestPasswordReset('user@example.com');
    final router = _router(usecase);
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.tap(find.text('Open password reset'));
    await tester.pumpAndSettle();

    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();
    expect(find.text('Sair da recuperação?'), findsOneWidget);

    await tester.tap(find.widgetWithText(TextButton, 'Continuar'));
    await tester.pumpAndSettle();
    expect(find.text('Confirme o código'), findsOneWidget);
  });
}

GoRouter _router(PasswordResetUsecase usecase) => GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, _) => Scaffold(
        body: Column(
          children: [
            const Text('Login'),
            TextButton(
              onPressed: () => context.push('/forgot-password'),
              child: const Text('Open password reset'),
            ),
          ],
        ),
      ),
    ),
    GoRoute(
      path: '/forgot-password',
      builder: (_, _) => PasswordResetPage(usecase: usecase),
    ),
  ],
);

GoRouter _routerWithFactory(PasswordResetUsecase Function() createUsecase) =>
    GoRouter(
      initialLocation: '/login',
      routes: [
        GoRoute(
          path: '/login',
          builder: (context, _) => Scaffold(
            body: Column(
              children: [
                const Text('Login'),
                TextButton(
                  onPressed: () => context.push('/forgot-password'),
                  child: const Text('Open password reset'),
                ),
              ],
            ),
          ),
        ),
        GoRoute(
          path: '/forgot-password',
          builder: (_, _) => PasswordResetPage(usecase: createUsecase()),
        ),
      ],
    );

final class _Repository implements PasswordResetRepository {
  int requestCalls = 0;
  int confirmCalls = 0;
  int completeCalls = 0;
  int authenticationCalls = 0;

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) async {
    requestCalls++;
    return Success(
      PasswordResetChallenge(
        passwordResetId: 'reset-id',
        email: email,
        codeExpiresAt: DateTime.now().add(const Duration(minutes: 10)),
        resendAvailableAt: DateTime.now(),
      ),
    );
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    confirmCalls++;
    return Success(
      PasswordResetProof(
        passwordResetToken: 'reset-proof',
        expiresAt: DateTime.now().add(const Duration(minutes: 10)),
      ),
    );
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) async {
    completeCalls++;
    return const Success(unit);
  }
}
