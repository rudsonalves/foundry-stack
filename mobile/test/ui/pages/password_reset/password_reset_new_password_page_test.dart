import 'dart:async';

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
  late _Repository repository;
  late PasswordResetUsecase usecase;

  setUp(() async {
    repository = _Repository();
    usecase = PasswordResetUsecase(repository: repository);
    await usecase.requestPasswordReset('user@example.com');
    await usecase.confirmPasswordReset('123456');
  });

  Future<void> pumpPage(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(home: PasswordResetPage(usecase: usecase)),
    );
  }

  testWidgets('validates password and confirmation before completion', (
    tester,
  ) async {
    await pumpPage(tester);

    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'different');
    await tester.pump();
    expect(_completeButton(tester).onPressed, isNull);

    await tester.enterText(_confirmationField(), 'Password1');
    await tester.pump();
    expect(_completeButton(tester).onPressed, isNotNull);
  });

  testWidgets('toggles each password visibility independently', (
    tester,
  ) async {
    await pumpPage(tester);
    final visibilityButtons = find.byIcon(Icons.visibility_outlined);

    expect(tester.widget<TextField>(_passwordField()).obscureText, isTrue);
    expect(tester.widget<TextField>(_confirmationField()).obscureText, isTrue);

    await tester.tap(visibilityButtons.first);
    await tester.pump();

    expect(tester.widget<TextField>(_passwordField()).obscureText, isFalse);
    expect(tester.widget<TextField>(_confirmationField()).obscureText, isTrue);
  });

  testWidgets('sends only password, clears fields and closes on success', (
    tester,
  ) async {
    final router = GoRouter(
      initialLocation: '/host',
      routes: [
        GoRoute(
          path: '/host',
          builder: (context, _) => Scaffold(
            body: TextButton(
              onPressed: () => context.push('/forgot-password'),
              child: const Text('Open'),
            ),
          ),
        ),
        GoRoute(
          path: '/forgot-password',
          builder: (_, _) => PasswordResetPage(usecase: usecase),
        ),
      ],
    );
    addTearDown(router.dispose);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.tap(find.text('Open'));
    await tester.pumpAndSettle();
    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'Password1');
    await tester.pump();

    await tester.tap(find.text('Alterar senha'));
    await tester.pump();
    await tester.pump(const Duration(seconds: 1));

    expect(repository.completedPasswords, ['Password1']);
    expect(usecase.email, isNull);
    expect(usecase.proof, isNull);
  });

  testWidgets('keeps values and step after recoverable failure', (
    tester,
  ) async {
    repository.completeResult = Future.value(
      const Failure(
        AppError(code: AppErrorCode.networkError, message: 'internal'),
      ),
    );
    await pumpPage(tester);
    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'Password1');
    await tester.pump();

    await tester.tap(find.text('Alterar senha'));
    await tester.pump();

    expect(
      tester.widget<TextField>(_passwordField()).controller?.text,
      'Password1',
    );
    expect(
      tester.widget<TextField>(_confirmationField()).controller?.text,
      'Password1',
    );
    expect(
      find.text('Não foi possível conectar ao servidor. Tente novamente.'),
      findsOneWidget,
    );
    expect(find.text('internal'), findsNothing);
  });

  testWidgets('clears fields and returns to email for invalid reset', (
    tester,
  ) async {
    repository.completeResult = Future.value(
      const Failure(
        AppError(
          code: AppErrorCode.invalidData,
          message: 'internal proof details',
          details: {'code': BackendErrorCodes.invalidPasswordReset},
        ),
      ),
    );
    await pumpPage(tester);
    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'Password1');
    await tester.pump();

    await tester.tap(find.text('Alterar senha'));
    await tester.pump();

    expect(find.text('Recupere sua senha'), findsOneWidget);
    expect(
      find.text(
        'A recuperação não é mais válida. Solicite um novo código e tente novamente.',
      ),
      findsOneWidget,
    );
    expect(find.text('internal proof details'), findsNothing);
  });

  testWidgets('disables fields and duplicate completion while running', (
    tester,
  ) async {
    final pending = Completer<Result<Unit>>();
    repository.completeResult = pending.future;
    await pumpPage(tester);
    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'Password1');
    await tester.pump();

    await tester.tap(find.text('Alterar senha'));
    await tester.pump();

    expect(tester.widget<TextField>(_passwordField()).enabled, isFalse);
    expect(_completeButton(tester).onPressed, isNull);
    expect(repository.completeCalls, 1);

    pending.complete(
      const Failure(
        AppError(code: AppErrorCode.networkError, message: 'offline'),
      ),
    );
    await tester.pump();
  });

  testWidgets('clears passwords and returns to code', (tester) async {
    await pumpPage(tester);
    await tester.enterText(_passwordField(), 'Password1');
    await tester.enterText(_confirmationField(), 'Password1');

    await tester.tap(find.widgetWithText(TextButton, 'Voltar'));
    await tester.pumpAndSettle();

    expect(find.text('Confirme o código'), findsOneWidget);
    expect(usecase.proof, isNull);
  });
}

Finder _passwordField() => find.widgetWithText(TextField, 'Nova senha');
Finder _confirmationField() =>
    find.widgetWithText(TextField, 'Confirmar nova senha');
ElevatedButton _completeButton(WidgetTester tester) => tester.widget(
  find.widgetWithText(ElevatedButton, 'Alterar senha'),
);

final class _Repository implements PasswordResetRepository {
  Future<Result<Unit>>? completeResult;
  final completedPasswords = <String>[];
  int completeCalls = 0;

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) async {
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
    return Success(
      PasswordResetProof(
        passwordResetToken: 'proof',
        expiresAt: DateTime.now().add(const Duration(minutes: 10)),
      ),
    );
  }

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) {
    completeCalls++;
    completedPasswords.add(newPassword);
    return completeResult ?? Future.value(const Success(unit));
  }
}
