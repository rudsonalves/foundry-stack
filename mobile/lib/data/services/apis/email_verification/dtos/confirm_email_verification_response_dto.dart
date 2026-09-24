class ConfirmEmailVerificationResponseDto {
  final String emailVerificationToken;
  final DateTime expiresAt;

  const ConfirmEmailVerificationResponseDto({
    required this.emailVerificationToken,
    required this.expiresAt,
  });

  factory ConfirmEmailVerificationResponseDto.fromMap(
    Map<String, dynamic> map,
  ) {
    final token = map['email_verification_token'];
    final expiresAt = _parseDateTime(map['expires_at']);

    if (token is! String || token.trim().isEmpty || expiresAt == null) {
      throw Exception(
        'Invalid map structure for ConfirmEmailVerificationResponseDto',
      );
    }

    return ConfirmEmailVerificationResponseDto(
      emailVerificationToken: token,
      expiresAt: expiresAt,
    );
  }
}

DateTime? _parseDateTime(Object? value) {
  if (value is! String || value.trim().isEmpty) return null;
  return DateTime.tryParse(value);
}
