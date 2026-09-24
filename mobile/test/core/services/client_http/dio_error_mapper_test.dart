import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/client_http/dio/dio_error_mapper.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  for (final errorCase in [
    (
      statusCode: 400,
      appErrorCode: AppErrorCode.invalidData,
      backendCode: BackendErrorCodes.invalidEmailVerification,
      message: 'Verificação inválida.',
    ),
    (
      statusCode: 409,
      appErrorCode: AppErrorCode.conflict,
      backendCode: BackendErrorCodes.emailAlreadyRegistered,
      message: 'E-mail já cadastrado.',
    ),
    (
      statusCode: 429,
      appErrorCode: AppErrorCode.httpError,
      backendCode: BackendErrorCodes.rateLimitExceeded,
      message: 'Limite excedido.',
    ),
    (
      statusCode: 503,
      appErrorCode: AppErrorCode.httpError,
      backendCode: BackendErrorCodes.emailDeliveryUnavailable,
      message: 'Envio indisponível.',
    ),
  ]) {
    test(
      'preserves ${errorCase.backendCode} from '
      'HTTP ${errorCase.statusCode}',
      () {
        final options = RequestOptions(path: '/email-verifications');

        final error = mapHttpError(
          DioException(
            requestOptions: options,
            type: DioExceptionType.badResponse,
            response: Response(
              requestOptions: options,
              statusCode: errorCase.statusCode,
              data: {
                'error': {
                  'code': errorCase.backendCode,
                  'message': errorCase.message,
                },
              },
            ),
          ),
        );

        expect(error.statusCode, errorCase.statusCode);
        expect(error.code, errorCase.appErrorCode);
        expect(error.message, errorCase.message);
        expect(
          backendErrorCode(error),
          errorCase.backendCode,
        );
      },
    );
  }

  test('preserves an AppError', () {
    const original = AppError(
      code: AppErrorCode.invalidData,
      message: 'invalid input',
    );

    final error = mapHttpError(original);

    expect(identical(error, original), isTrue);
  });

  test('maps connection failures', () {
    final error = mapHttpError(
      DioException(
        requestOptions: RequestOptions(path: '/listings'),
        type: DioExceptionType.connectionError,
      ),
    );

    expect(error.code, AppErrorCode.networkError);
  });
}
