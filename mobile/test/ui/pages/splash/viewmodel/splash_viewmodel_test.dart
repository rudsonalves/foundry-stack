import 'dart:async';

import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/bootstrap_usecase.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/models/bootstrap_destination.dart';
import 'package:foundry_stack_mobile/ui/pages/splash/viewmodel/splash_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  late _FakeBootstrapUsecase usecase;
  late SplashViewmodel viewmodel;

  setUp(() {
    usecase = _FakeBootstrapUsecase();
    viewmodel = SplashViewmodel(usecase);
  });

  test('delegates initialization and exposes its destination', () async {
    usecase.result = const Success(BootstrapDestination.home);

    await viewmodel.initialize.execute();

    expect(viewmodel.initialize.value, BootstrapDestination.home);
    expect(usecase.initializeCalls, 1);
  });

  test('preserves a bootstrap failure', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'bootstrap failed',
    );
    usecase.result = const Failure(error);

    await viewmodel.initialize.execute();

    expect(viewmodel.initialize.isFailure, isTrue);
    expect(viewmodel.initialize.error, same(error));
    expect(viewmodel.initialize.value, isNull);
    expect(usecase.initializeCalls, 1);
  });

  test('does not initialize twice while the command is running', () async {
    final completer = Completer<Result<BootstrapDestination>>();
    usecase.future = completer.future;

    final first = viewmodel.initialize.execute();
    final second = viewmodel.initialize.execute();

    expect(viewmodel.initialize.isRunning, isTrue);
    expect(usecase.initializeCalls, 1);

    completer.complete(const Success(BootstrapDestination.login));
    await Future.wait([first, second]);

    expect(usecase.initializeCalls, 1);
    expect(viewmodel.initialize.value, BootstrapDestination.login);
  });
}

class _FakeBootstrapUsecase implements BootstrapUsecase {
  Result<BootstrapDestination> result = const Success(
    BootstrapDestination.login,
  );
  Future<Result<BootstrapDestination>>? future;
  int initializeCalls = 0;

  @override
  AsyncResult<BootstrapDestination> initialize() {
    initializeCalls++;
    return future ?? Future.value(result);
  }
}
