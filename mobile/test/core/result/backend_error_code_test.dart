import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';

void main() {
  test('extracts a backend code from API error details', () {
    const error = AppError(
      code: AppErrorCode.conflict,
      message: 'Conflito.',
      details: {
        'error': {'code': 'EMAIL_ALREADY_EXISTS'},
      },
    );

    expect(backendErrorCode(error), 'EMAIL_ALREADY_EXISTS');
  });
}
