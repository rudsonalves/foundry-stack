class EmailVerificationResponseDto {
  final String verificationId;
  final DateTime codeExpiresAt;
  final DateTime resendAvailableAt;

  const EmailVerificationResponseDto({
    required this.verificationId,
    required this.codeExpiresAt,
    required this.resendAvailableAt,
  });

  factory EmailVerificationResponseDto.fromMap(Map<String, dynamic> map) {
    final verificationId = map['verification_id'];
    final codeExpiresAt = _parseDateTime(map['code_expires_at']);
    final resendAvailableAt = _parseDateTime(map['resend_available_at']);

    if (verificationId is! String ||
        verificationId.trim().isEmpty ||
        codeExpiresAt == null ||
        resendAvailableAt == null) {
      throw Exception('Invalid map structure for EmailVerificationResponseDto');
    }

    return EmailVerificationResponseDto(
      verificationId: verificationId,
      codeExpiresAt: codeExpiresAt,
      resendAvailableAt: resendAvailableAt,
    );
  }
}

DateTime? _parseDateTime(Object? value) {
  if (value is! String || value.trim().isEmpty) return null;
  return DateTime.tryParse(value);
}
