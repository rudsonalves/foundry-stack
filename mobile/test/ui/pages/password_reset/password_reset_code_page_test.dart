import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/password_reset/password_reset_repository.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_challenge.dart';
import 'package:foundry_stack_mobile/domain/common/password_reset/models/password_reset_proof.dart';
import 'package:foundry_stack_mobile/domain/usecases/password_reset/password_reset_usecase.dart';
import 'package:foundry_stack_mobile/ui/components/input_text/token_input.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/password_reset_page.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _Repository repository;
  late PasswordResetUsecase usecase;

  setUp(() async {
    repository = _Repository();
    usecase = PasswordResetUsecase(repository: repository);
    await usecase.requestPasswordReset('user@example.com');
  });

  Future<void> pumpPage(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(home: PasswordResetPage(usecase: usecase)),
    );
  }

  Finder tokenField() => find.descendant(
    of: find.byType(TokenInput),
    matching: find.byType(TextField),
  );

  testWidgets('masks email and does not confirm automatically', (tester) async {
    await pumpPage(tester);

    expect(
      find.text('Digite o código de 6 dígitos enviado para u***r@example.com.'),
      findsOneWidget,
    );

    await tester.enterText(tokenField(), '012345');
    await tester.pump();

    expect(repository.confirmedCodes, isEmpty);
    expect(_button(tester, 'Confirmar código').onPressed, isNotNull);
  });

  testWidgets('preserves leading zero and advances only on explicit confirm', (
    tester,
  ) async {
    await pumpPage(tester);
    await tester.enterText(tokenField(), '012345');
    await tester.pump();

    await tester.tap(find.text('Confirmar código'));
    await tester.pumpAndSettle();

    expect(repository.confirmedCodes, ['012345']);
    expect(find.text('Crie uma nova senha'), findsWidgets);
  });

  testWidgets('clears code and replaces countdown after successful resend', (
    tester,
  ) async {
    await pumpPage(tester);
    await tester.enterText(tokenField(), '123456');
    repository.challenge = repository.newChallenge();

    await tester.tap(find.text('Reenviar código'));
    await tester.pump();

    expect(tester.widget<TextField>(tokenField()).controller?.text, isEmpty);
    expect(repository.requestedEmails, hasLength(2));
    expect(find.textContaining('Reenviar código em'), findsOneWidget);
  });

  testWidgets('keeps code and challenge when resend fails', (tester) async {
    repository.requestFailure = const AppError(
      code: AppErrorCode.networkError,
      message: 'internal network details',
    );
    await pumpPage(tester);
    await tester.enterText(tokenField(), '123456');

    await tester.tap(find.text('Reenviar código'));
    await tester.pump();

    expect(tester.widget<TextField>(tokenField()).controller?.text, '123456');
    expect(
      find.text('Não foi possível conectar ao servidor. Tente novamente.'),
      findsOneWidget,
    );
    expect(find.text('internal network details'), findsNothing);
  });

  testWidgets('uses one public message for invalid confirmation', (
    tester,
  ) async {
    repository.confirmFailure = const AppError(
      code: AppErrorCode.invalidData,
      message: 'expired versus invalid internal detail',
      details: {'code': BackendErrorCodes.invalidPasswordReset},
    );
    await pumpPage(tester);
    await tester.enterText(tokenField(), '123456');
    await tester.pump();

    await tester.tap(find.text('Confirmar código'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(
      find.text(
        'A recuperação não é mais válida. Solicite um novo código e tente novamente.',
      ),
      findsOneWidget,
    );
    expect(find.text('expired versus invalid internal detail'), findsNothing);
    expect(
      tester.widget<EditableText>(find.byType(EditableText)).focusNode.hasFocus,
      isTrue,
    );
  });

  testWidgets('returns to email and discards the challenge', (tester) async {
    await pumpPage(tester);

    await tester.tap(find.text('Alterar e-mail'));
    await tester.pumpAndSettle();

    expect(find.text('Recupere sua senha'), findsOneWidget);
    expect(usecase.challenge, isNull);
  });
}

ElevatedButton _button(WidgetTester tester, String label) =>
    tester.widget<ElevatedButton>(find.widgetWithText(ElevatedButton, label));

final class _Repository implements PasswordResetRepository {
  _Repository() : challenge = _challenge(resendAvailableAt: DateTime.now());

  PasswordResetChallenge challenge;
  AppError? requestFailure;
  AppError? confirmFailure;
  final requestedEmails = <String>[];
  final confirmedCodes = <String>[];

  PasswordResetChallenge newChallenge() => _challenge(
    id: 'new-reset-id',
    resendAvailableAt: DateTime.now().add(const Duration(minutes: 1)),
  );

  @override
  AsyncResult<PasswordResetChallenge> requestPasswordReset(String email) async {
    requestedEmails.add(email);
    final error = requestFailure;
    if (error != null && requestedEmails.length > 1) return Failure(error);
    return Success(challenge);
  }

  @override
  AsyncResult<PasswordResetProof> confirmPasswordReset({
    required String passwordResetId,
    required String code,
  }) async {
    confirmedCodes.add(code);
    final error = confirmFailure;
    if (error != null) return Failure(error);
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
  }) => throw UnimplementedError();
}

PasswordResetChallenge _challenge({
  String id = 'reset-id',
  required DateTime resendAvailableAt,
}) => PasswordResetChallenge(
  passwordResetId: id,
  email: 'user@example.com',
  codeExpiresAt: DateTime.now().add(const Duration(minutes: 10)),
  resendAvailableAt: resendAvailableAt,
);
