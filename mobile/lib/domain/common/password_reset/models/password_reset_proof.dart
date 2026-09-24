class PasswordResetProof {
  final String passwordResetToken;
  final DateTime expiresAt;

  const PasswordResetProof({
    required this.passwordResetToken,
    required this.expiresAt,
  });

  bool isExpiredAt(DateTime instant) => !instant.isBefore(expiresAt);
}
