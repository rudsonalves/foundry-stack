import '/core/extensions/datetime_extension.dart';

class PasswordResetResponseDto {
  final String passwordResetId;
  final DateTime codeExpiresAt;
  final DateTime resendAvailableAt;

  const PasswordResetResponseDto({
    required this.passwordResetId,
    required this.codeExpiresAt,
    required this.resendAvailableAt,
  });

  factory PasswordResetResponseDto.fromMap(Map<String, dynamic> map) {
    final passwordResetId = map['password_reset_id'];
    final codeExpiresAt = DateParser.parseOrNull(map['code_expires_at']);
    final resendAvailableAt = DateParser.parseOrNull(
      map['resend_available_at'],
    );

    if (passwordResetId is! String ||
        passwordResetId.trim().isEmpty ||
        codeExpiresAt == null ||
        resendAvailableAt == null) {
      throw Exception(
        'Invalid map structure for PasswordResetResponseDto',
      );
    }

    return PasswordResetResponseDto(
      passwordResetId: passwordResetId.trim(),
      codeExpiresAt: codeExpiresAt,
      resendAvailableAt: resendAvailableAt,
    );
  }
}
