class CompletePasswordResetRequestDto {
  final String passwordResetId;
  final String passwordResetToken;
  final String newPassword;

  const CompletePasswordResetRequestDto({
    required this.passwordResetId,
    required this.passwordResetToken,
    required this.newPassword,
  });

  Map<String, dynamic> toMap() {
    return {
      'password_reset_id': passwordResetId,
      'password_reset_token': passwordResetToken,
      'new_password': newPassword,
    };
  }
}
