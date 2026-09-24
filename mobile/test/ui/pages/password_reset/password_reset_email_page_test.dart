import 'dart:async';

import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/password_reset_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _Repository repository;
  late PasswordResetUsecase usecase;

  setUp(() {
    repository = _Repository();
    usecase = PasswordResetUsecase(repository: repository);
  });

  Future<void> pumpPage(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(home: PasswordResetPage(usecase: usecase)),
    );
  }

  testWidgets('shows the email step with coherent title and progress', (
    tester,
  ) async {
    await pumpPage(tester);

    expect(find.text('Recupere sua senha'), findsOneWidget);
    expect(find.byKey(const ValueKey('password-reset-email-step')), findsOne);
    expect(
      tester
          .widget<LinearProgressIndicator>(
            find.byType(LinearProgressIndicator),
          )
          .value,
      closeTo(1 / 3, 0.001),
    );
  });

  testWidgets('rejects invalid email without requesting recovery', (
    tester,
  ) async {
    await pumpPage(tester);

    await tester.enterText(_emailField(), 'invalid');
    await tester.pump();

    expect(find.text('E-mail inválido'), findsOneWidget);
    expect(_submitButton(tester).onPressed, isNull);
    expect(repository.requestedEmails, isEmpty);
  });

  testWidgets('disables input and duplicate action while requesting', (
    tester,
  ) async {
    final pending = Completer<Result<PasswordResetChallenge>>();
    repository.requestResult = pending.future;
    await pumpPage(tester);
    await tester.enterText(_emailField(), 'user@example.com');
    await tester.pump();

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(repository.requestedEmails, ['user@example.com']);
    expect(tester.widget<TextField>(_emailField()).enabled, isFalse);
    expect(find.byType(CircularProgressIndicator), findsWidgets);

    await tester.tap(_submitFinder());
    expect(repository.requestedEmails, hasLength(1));

    pending.complete(Success(repository.challenge));
    await tester.pumpAndSettle();
  });

  testWidgets('shows neutral success message and advances to code', (
    tester,
  ) async {
    await pumpPage(tester);
    await tester.enterText(_emailField(), ' user@example.com ');
    await tester.pump();

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(repository.requestedEmails, ['user@example.com']);
    expect(
      find.text(
        'Se este e-mail estiver cadastrado, enviaremos um código de recuperação.',
      ),
      findsOneWidget,
    );
    expect(find.text('Confirme o código'), findsOneWidget);
    expect(find.byKey(const ValueKey('password-reset-code-step')), findsOne);
  });

  testWidgets('shows rate limit guidance and keeps email step', (
    tester,
  ) async {
    repository.requestResult = Future.value(
      const Failure(
        AppError(
          code: AppErrorCode.httpError,
          message: 'rate limited',
          details: {'code': BackendErrorCodes.rateLimitExceeded},
        ),
      ),
    );
    await pumpPage(tester);
    await tester.enterText(_emailField(), 'user@example.com');
    await tester.pump();

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(
      find.text('Muitas tentativas. Aguarde antes de tentar novamente.'),
      findsOneWidget,
    );
    expect(find.byKey(const ValueKey('password-reset-email-step')), findsOne);
    expect(
      tester.widget<TextField>(_emailField()).controller?.text,
      'user@example.com',
    );
  });

  testWidgets('shows generic network message and allows retry', (
    tester,
  ) async {
    repository.requestResult = Future.value(
      const Failure(
        AppError(code: AppErrorCode.networkError, message: 'internal'),
      ),
    );
    await pumpPage(tester);
    await tester.enterText(_emailField(), 'user@example.com');
    await tester.pump();

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(
      find.text('Não foi possível conectar ao servidor. Tente novamente.'),
      findsOneWidget,
    );
    expect(_submitButton(tester).onPressed, isNotNull);
    expect(find.text('internal'), findsNothing);
  });
}

Finder _emailField() => find.widgetWithText(TextField, 'E-mail');
Finder _submitFinder() => find.widgetWithText(ElevatedButton, 'Enviar código');
ElevatedButton _submitButton(WidgetTester tester) =>
    tester.widget<ElevatedButton>(_submitFinder());

final class _Repository implements PasswordResetRepository {
  final challenge = PasswordResetChallenge(
    passwordResetId: 'reset-id',
    email: 'user@example.com',
    codeExpiresAt: DateTime.utc(2026, 9, 23, 15, 15),
    resendAvailableAt: DateTime.utc(2026, 9, 23, 15, 1),
  );
  Future<Result<PasswordResetChallenge>>? requestResult;
  final requestedEmails = <String>[];

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) {
    requestedEmails.add(email);
    return requestResult ?? Future.value(Success(challenge));
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) => throw UnimplementedError();

  @override
  AsyncResult<Unit> completePasswordReset({
    required String passwordResetId,
    required String passwordResetToken,
    required String newPassword,
  }) => throw UnimplementedError();
}
