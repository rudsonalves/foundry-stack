import '/core/result/result.dart';

abstract final class ApiResponseParser {
  static Result<T> parse<T extends Object>(
    Object? body,
    T Function(Map<String, dynamic> data) decode,
  ) {
    if (body is! Map<String, dynamic>) {
      return _failure('Response body must be an Map<String, dynamic>');
    }

    if (!body.containsKey('data')) {
      return _failure('Response body does not contain data');
    }

    final data = body['data'];
    if (data == null) {
      return _failure('Response data cannot be null');
    }

    if (data is! Map<String, dynamic>) {
      return _failure('Response data must be an Map<String, dynamic>');
    }

    try {
      return Success(decode(data));
    } on AppError catch (error) {
      return Failure(error);
    } catch (error) {
      return _failure('Invalid API response');
    }
  }

  static Failure<T> _failure<T extends Object>(String message) {
    return Failure(
      AppError(
        code: AppErrorCode.parsingError,
        message: message,
      ),
    );
  }
}
