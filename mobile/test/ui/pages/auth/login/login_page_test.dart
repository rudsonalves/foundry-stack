import 'dart:async';

import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/app/routing/routes.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';
import 'package:foundry_stack_mobile/ui/pages/auth/login/login_page.dart';
import 'package:foundry_stack_mobile/ui/pages/auth/login/viewmodel/login_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/password_reset_route_result.dart';

void main() {
  testWidgets('keeps submit disabled for empty fields and invalid e-mail', (
    tester,
  ) async {
    final repository = _FakeAuthRepository();
    await tester.pumpWidget(_testApp(repository));

    expect(_submitButton(tester).onPressed, isNull);

    await tester.enterText(_emailField(), 'invalid');
    await tester.enterText(_passwordField(), 'secret');
    await tester.pump();

    expect(find.text('E-mail inválido'), findsOneWidget);
    expect(_submitButton(tester).onPressed, isNull);
    expect(repository.loginCalls, 0);
  });

  testWidgets('shows loading and blocks duplicate submissions', (tester) async {
    final completer = Completer<Result<Unit>>();
    final repository = _FakeAuthRepository(() => completer.future);
    await tester.pumpWidget(_testApp(repository));
    await _fillValidCredentials(tester);

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(repository.loginCalls, 1);
    expect(_submitButton(tester).onPressed, isNull);
    expect(find.byType(CircularProgressIndicator), findsWidgets);

    await tester.tap(_submitFinder());
    expect(repository.loginCalls, 1);

    completer.complete(const Success(unit));
    await tester.pumpAndSettle();
  });

  testWidgets('shows a generic invalid-credentials message', (tester) async {
    final repository = _FakeAuthRepository(
      () async => const Failure(
        AppError(
          statusCode: 401,
          code: AppErrorCode.unauthenticated,
          message: 'invalid email or password',
        ),
      ),
    );
    await tester.pumpWidget(_testApp(repository));
    await _fillValidCredentials(tester);

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(
      find.text(
        'E-mail ou senha incorretos. Verifique os dados e tente novamente.',
      ),
      findsOneWidget,
    );
    expect(find.text('Home destination'), findsNothing);
  });

  testWidgets('keeps fields after network failure and allows retry', (
    tester,
  ) async {
    late _FakeAuthRepository repository;
    repository = _FakeAuthRepository(() async {
      if (repository.loginCalls == 1) {
        return const Failure(
          AppError(
            code: AppErrorCode.networkError,
            message: 'network failure',
          ),
        );
      }
      return const Success(unit);
    });
    await tester.pumpWidget(_testApp(repository));
    await _fillValidCredentials(tester);

    await tester.tap(_submitFinder());
    await tester.pump();

    expect(
      find.text('Não foi possível entrar agora. Tente novamente em instantes'),
      findsOneWidget,
    );
    expect(
      tester.widget<TextField>(_emailField()).controller?.text,
      'ada@example.com',
    );
    expect(
      tester.widget<TextField>(_passwordField()).controller?.text,
      'wrong-secret',
    );

    await tester.tap(_submitFinder());
    await tester.pumpAndSettle();

    expect(repository.loginCalls, 2);
    expect(find.text('Home destination'), findsOneWidget);
  });

  testWidgets('replaces Login with Home after successful login', (
    tester,
  ) async {
    final repository = _FakeAuthRepository();
    final router = _testRouter(repository);
    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await _fillValidCredentials(tester);

    await tester.tap(_submitFinder());
    await tester.pumpAndSettle();

    expect(find.text('Home destination'), findsOneWidget);
    expect(router.canPop(), isFalse);
  });

  testWidgets('navigates to user registration', (tester) async {
    final repository = _FakeAuthRepository();
    await tester.pumpWidget(_testApp(repository));

    await tester.tap(find.text('Não tem conta? Cadastre-se'));
    await tester.pumpAndSettle();

    expect(find.text('Register destination'), findsOneWidget);
  });

  testWidgets('opens password reset without logging in and shows its result', (
    tester,
  ) async {
    final repository = _FakeAuthRepository();
    await tester.pumpWidget(_testApp(repository));

    await tester.tap(find.text('Esqueci minha senha'));
    await tester.pumpAndSettle();
    expect(find.text('Password reset destination'), findsOneWidget);

    await tester.tap(find.text('Complete password reset'));
    await tester.pumpAndSettle();

    expect(repository.loginCalls, 0);
    expect(
      find.text('Senha alterada com sucesso. Entre com a nova senha.'),
      findsOneWidget,
    );
  });
}

Widget _testApp(_FakeAuthRepository repository) => MaterialApp.router(
  routerConfig: _testRouter(repository),
);

GoRouter _testRouter(_FakeAuthRepository repository) => GoRouter(
  initialLocation: AuthRoutes.login.routePath,
  routes: [
    GoRoute(
      path: AuthRoutes.login.routePath,
      name: AuthRoutes.login.routeName,
      builder: (context, state) => LoginPage(
        viewmodel: LoginViewmodel(repository),
      ),
    ),
    GoRoute(
      path: BaseRoutes.home.routePath,
      name: BaseRoutes.home.routeName,
      builder: (context, state) => const Scaffold(
        body: Text('Home destination'),
      ),
    ),
    GoRoute(
      path: BaseRoutes.register.routePath,
      name: BaseRoutes.register.routeName,
      builder: (context, state) => const Scaffold(
        body: Text('Register destination'),
      ),
    ),
    GoRoute(
      path: BaseRoutes.forgotPassword.routePath,
      name: BaseRoutes.forgotPassword.routeName,
      builder: (context, state) => Scaffold(
        body: Column(
          children: [
            const Text('Password reset destination'),
            TextButton(
              onPressed: () => context.pop(
                PasswordResetRouteResult.completed,
              ),
              child: const Text('Complete password reset'),
            ),
          ],
        ),
      ),
    ),
  ],
);

Finder _emailField() => find.widgetWithText(TextField, 'E-mail');
Finder _passwordField() => find.widgetWithText(TextField, 'Senha');
Finder _submitFinder() => find.widgetWithText(ElevatedButton, 'Entrar');

ElevatedButton _submitButton(WidgetTester tester) =>
    tester.widget<ElevatedButton>(_submitFinder());

Future<void> _fillValidCredentials(WidgetTester tester) async {
  await tester.enterText(_emailField(), 'ada@example.com');
  await tester.enterText(_passwordField(), 'wrong-secret');
  await tester.pump();
}

final class _FakeAuthRepository implements AuthRepository {
  _FakeAuthRepository([
    AsyncResult<Unit> Function()? login,
  ]) : _login = login ?? (() async => const Success(unit));

  final AsyncResult<Unit> Function() _login;
  int loginCalls = 0;

  @override
  AsyncResult<Unit> login(LoginCredentials request) {
    loginCalls++;
    return _login();
  }

  @override
  AsyncResult<Unit> clearLocalSession() => throw UnimplementedError();

  @override
  AsyncResult<RestoreSessionStatus> restoreSession() =>
      throw UnimplementedError();

  @override
  AsyncResult<Unit> logout() => throw UnimplementedError();
}
