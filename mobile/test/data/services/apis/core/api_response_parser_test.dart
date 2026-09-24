import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/data/services/apis/core/api_response_parser.dart';

void main() {
  group('ApiResponseParser.parse', () {
    test('decodes object data from a valid response body', () {
      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{
          'data': <String, dynamic>{'value': 'decoded'},
        },
        (data) => data['value'] as String,
      );

      expect(result.isSuccess, isTrue);
      expect(result.value, 'decoded');
    });

    test('rejects a body that is not a Map<String, dynamic>', () {
      final result = ApiResponseParser.parse<String>('invalid', (_) => '');

      expectParsingFailure(
        result,
        'Response body must be an Map<String, dynamic>',
      );
    });

    test('rejects a body without data', () {
      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{},
        (_) => '',
      );

      expectParsingFailure(result, 'Response body does not contain data');
    });

    test('rejects null data', () {
      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{'data': null},
        (_) => '',
      );

      expectParsingFailure(result, 'Response data cannot be null');
    });

    test('rejects data that is not a Map<String, dynamic>', () {
      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{'data': 'invalid'},
        (_) => '',
      );

      expectParsingFailure(
        result,
        'Response data must be an Map<String, dynamic>',
      );
    });

    test('converts decoder failures into a parsing failure', () {
      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{
          'data': <String, dynamic>{'value': 'invalid'},
        },
        (_) => throw const FormatException('invalid value'),
      );

      expectParsingFailure(result, 'Invalid API response');
    });

    test('preserves an AppError returned by the decoder', () {
      const error = AppError(
        code: AppErrorCode.invalidData,
        message: 'invalid decoded data',
      );

      final result = ApiResponseParser.parse<String>(
        const <String, dynamic>{'data': <String, dynamic>{}},
        (_) => throw error,
      );

      expect(result.isFailure, isTrue);
      expect(identical(result.error, error), isTrue);
    });
  });
}

void expectParsingFailure<T extends Object>(
  Result<T> result,
  String message,
) {
  expect(result.isFailure, isTrue);
  expect(result.error, isA<AppError>());

  final error = result.error!;
  expect(error.code, AppErrorCode.parsingError);
  expect(error.message, message);
}
