import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/ui/pages/password_reset/password_reset_messages.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('maps semantic and transport errors to public messages', () {
    const invalid = AppError(
      code: AppErrorCode.invalidData,
      message: 'internal invalid reason',
      details: {'code': BackendErrorCodes.invalidPasswordReset},
    );
    const rateLimit = AppError(
      code: AppErrorCode.httpError,
      message: 'internal rate limit',
      details: {'code': BackendErrorCodes.rateLimitExceeded},
    );
    const network = AppError(
      code: AppErrorCode.networkError,
      message: 'socket details',
    );

    expect(
      PasswordResetMessages.forError(invalid, fallback: 'fallback'),
      PasswordResetMessages.invalid,
    );
    expect(
      PasswordResetMessages.forError(rateLimit, fallback: 'fallback'),
      PasswordResetMessages.rateLimit,
    );
    expect(
      PasswordResetMessages.forError(network, fallback: 'fallback'),
      PasswordResetMessages.connection,
    );
  });

  test('uses caller fallback without exposing the API message', () {
    const error = AppError(
      code: AppErrorCode.httpError,
      message: 'sensitive internal response',
    );

    expect(
      PasswordResetMessages.forError(error, fallback: 'Tente novamente.'),
      'Tente novamente.',
    );
  });
}
