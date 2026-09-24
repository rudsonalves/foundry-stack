import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/repositories/auth/auth_repository.dart';
import 'package:foundry_stack_mobile/domain/common/auth/models/login_credentials.dart';
import 'package:foundry_stack_mobile/ui/pages/auth/login/model/login_request_model.dart';
import 'package:foundry_stack_mobile/ui/pages/auth/login/viewmodel/login_viewmodel.dart';

void main() {
  late _FakeAuthRepository repository;
  late LoginViewmodel viewmodel;

  setUp(() {
    repository = _FakeAuthRepository();
    viewmodel = LoginViewmodel(repository);
  });

  test('delegates valid credentials to the repository', () async {
    await viewmodel.loginCommand.execute(
      LoginRequestModel(email: 'ada@example.com', password: 'secret'),
    );

    expect(viewmodel.loginCommand.isSuccess, isTrue);
    expect(repository.loginCalls, 1);
    expect(repository.lastRequest?.email, 'ada@example.com');
    expect(repository.lastRequest?.password, 'secret');
  });

  test('preserves invalid credentials returned by the repository', () async {
    const error = AppError(
      statusCode: 401,
      code: AppErrorCode.unauthenticated,
      message: 'invalid credentials',
    );
    repository.loginResult = const Failure(error);

    await viewmodel.loginCommand.execute(
      LoginRequestModel(email: 'ada@example.com', password: 'wrong'),
    );

    expect(viewmodel.loginCommand.isFailure, isTrue);
    expect(identical(viewmodel.loginCommand.error, error), isTrue);
  });

  test('does not execute login twice while the command is running', () async {
    final completer = Completer<Result<Unit>>();
    repository.loginFuture = completer.future;
    final request = LoginRequestModel(
      email: 'ada@example.com',
      password: 'secret',
    );

    final firstExecution = viewmodel.loginCommand.execute(request);
    final secondExecution = viewmodel.loginCommand.execute(request);

    expect(viewmodel.loginCommand.isRunning, isTrue);
    expect(repository.loginCalls, 1);

    completer.complete(const Success(unit));
    await Future.wait([firstExecution, secondExecution]);

    expect(repository.loginCalls, 1);
    expect(viewmodel.loginCommand.isSuccess, isTrue);
  });

  test('preserves a network failure returned by the repository', () async {
    const error = AppError(
      code: AppErrorCode.networkError,
      message: 'network failure',
    );
    repository.loginResult = const Failure(error);

    await viewmodel.loginCommand.execute(
      LoginRequestModel(email: 'ada@example.com', password: 'secret'),
    );

    expect(viewmodel.loginCommand.isFailure, isTrue);
    expect(identical(viewmodel.loginCommand.error, error), isTrue);
  });

  test('validates e-mail format and requires a non-empty password', () {
    expect(viewmodel.isValid(LoginRequestModel()), isFalse);
    expect(
      viewmodel.isValid(
        LoginRequestModel(email: 'invalid', password: 'secret'),
      ),
      isFalse,
    );
    expect(
      viewmodel.isValid(
        LoginRequestModel(email: 'ada@example.com', password: 'x'),
      ),
      isTrue,
    );
  });
}

final class _FakeAuthRepository implements AuthRepository {
  Result<Unit> loginResult = const Success(unit);
  Future<Result<Unit>>? loginFuture;
  int loginCalls = 0;
  LoginCredentials? lastRequest;

  @override
  AsyncResult<Unit> login(LoginCredentials request) {
    loginCalls++;
    lastRequest = request;
    return loginFuture ?? Future.value(loginResult);
  }

  @override
  AsyncResult<Unit> clearLocalSession() => throw UnimplementedError();

  @override
  AsyncResult<RestoreSessionStatus> restoreSession() =>
      throw UnimplementedError();

  @override
  AsyncResult<Unit> logout() => throw UnimplementedError();
}
