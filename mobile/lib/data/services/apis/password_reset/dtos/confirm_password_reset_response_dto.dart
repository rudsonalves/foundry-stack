import '/core/extensions/datetime_extension.dart';

class ConfirmPasswordResetResponseDto {
  final String passwordResetToken;
  final DateTime expiresAt;

  const ConfirmPasswordResetResponseDto({
    required this.passwordResetToken,
    required this.expiresAt,
  });

  factory ConfirmPasswordResetResponseDto.fromMap(Map<String, dynamic> map) {
    final passwordResetToken = map['password_reset_token'];
    final expiresAt = DateParser.parseOrNull(map['expires_at']);

    if (passwordResetToken is! String ||
        passwordResetToken.trim().isEmpty ||
        expiresAt == null) {
      throw Exception(
        'Invalid map structure for '
        'ConfirmPasswordResetResponseDto',
      );
    }

    return ConfirmPasswordResetResponseDto(
      passwordResetToken: passwordResetToken.trim(),
      expiresAt: expiresAt,
    );
  }
}
