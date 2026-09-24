import 'app_error.dart';

abstract final class BackendErrorCodes {
  static const emailAlreadyRegistered = 'EMAIL_ALREADY_REGISTERED';
  static const emailDeliveryUnavailable = 'EMAIL_DELIVERY_UNAVAILABLE';
  static const rateLimitExceeded = 'RATE_LIMIT_EXCEEDED';
  static const invalidEmailVerification = 'INVALID_EMAIL_VERIFICATION';
  static const invalidPasswordReset = 'INVALID_PASSWORD_RESET';
}

String? backendErrorCode(AppError? error) {
  if (error == null) return null;
  return _readCode(error.details);
}

String? _readCode(Object? value) {
  if (value is! Map) return null;

  final code = value['code'];
  if (code is String && code.trim().isNotEmpty) {
    return code;
  }

  return _readCode(value['error']) ?? _readCode(value['details']);
}
