import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';

void main() {
  group('Result', () {
    test('exposes a successful value', () {
      const result = Success<String>('ok');

      expect(result.isSuccess, isTrue);
      expect(result.isFailure, isFalse);
      expect(result.value, 'ok');
      expect(result.error, isNull);
      expect(
        result.fold(onSuccess: (value) => value, onFailure: (_) => 'error'),
        'ok',
      );
    });

    test('exposes an application error', () {
      const error = AppError(
        code: AppErrorCode.invalidData,
        message: 'Dados inválidos.',
      );
      const result = Failure<String>(error);

      expect(result.isFailure, isTrue);
      expect(result.value, isNull);
      expect(result.error, same(error));
      expect(
        result.fold(
          onSuccess: (value) => value,
          onFailure: (error) => error.toString(),
        ),
        'AppError(null, invalidData, Dados inválidos.)',
      );
    });
  });
}
