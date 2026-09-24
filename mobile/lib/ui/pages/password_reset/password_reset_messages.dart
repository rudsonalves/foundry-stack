import '/core/result/result.dart';

abstract final class PasswordResetMessages {
  static const invalid =
      'A recuperação não é mais válida. Solicite um novo código e tente novamente.';
  static const rateLimit =
      'Muitas tentativas. Aguarde antes de tentar novamente.';
  static const connection =
      'Não foi possível conectar ao servidor. Tente novamente.';

  static String forError(
    AppError error, {
    required String fallback,
  }) {
    return switch (backendErrorCode(error)) {
      BackendErrorCodes.invalidPasswordReset => invalid,
      BackendErrorCodes.rateLimitExceeded => rateLimit,
      _ => switch (error.code) {
        AppErrorCode.networkError || AppErrorCode.timeout => connection,
        _ => fallback,
      },
    };
  }
}
