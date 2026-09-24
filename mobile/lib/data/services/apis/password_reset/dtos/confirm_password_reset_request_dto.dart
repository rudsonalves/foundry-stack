class ConfirmPasswordResetRequestDto {
  final String passwordResetId;
  final String code;

  const ConfirmPasswordResetRequestDto({
    required this.passwordResetId,
    required this.code,
  });

  Map<String, dynamic> toMap() {
    return {
      'password_reset_id': passwordResetId,
      'code': code,
    };
  }
}
