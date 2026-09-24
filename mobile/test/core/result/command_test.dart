import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/command.dart';

void main() {
  test('Command0 moves from idle to success', () async {
    final command = Command0<String>(
      () async => const Success<String>('concluído'),
    );

    expect(command.isIdle, isTrue);

    await command.execute();

    expect(command.isSuccess, isTrue);
    expect(command.value, 'concluído');
  });

  test('Command1 forwards its input', () async {
    final command = Command1<int, int>(
      (value) async => Success<int>(value * 2),
    );

    await command.execute(4);

    expect(command.value, 8);
  });

  test('does not execute the same command concurrently', () async {
    final completer = Completer<Result<String>>();
    var executions = 0;
    final command = Command0<String>(() {
      executions++;
      return completer.future;
    });

    final first = command.execute();
    final second = command.execute();

    expect(command.isRunning, isTrue);
    expect(executions, 1);

    completer.complete(const Success<String>('ok'));
    await Future.wait([first, second]);

    expect(executions, 1);
    expect(command.value, 'ok');
  });

  test('preserves an AppError thrown by the action', () async {
    const error = AppError(
      code: AppErrorCode.invalidData,
      message: 'invalid input',
    );
    final command = Command0<String>(() async => throw error);

    await command.execute();

    expect(command.isFailure, isTrue);
    expect(identical(command.error, error), isTrue);
  });

  test('maps an unexpected exception to AppError', () async {
    final command = Command0<String>(() async => throw StateError('falha'));

    await command.execute();

    expect(command.isFailure, isTrue);
    expect(command.error?.code, AppErrorCode.unknown);
    expect(
      command.error?.message,
      'An unexpected error occurred: Bad state: falha',
    );
  });
}
