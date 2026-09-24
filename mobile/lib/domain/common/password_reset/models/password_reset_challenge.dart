class PasswordResetChallenge {
  final String passwordResetId;
  final String email;
  final DateTime codeExpiresAt;
  final DateTime resendAvailableAt;

  const PasswordResetChallenge({
    required this.passwordResetId,
    required this.email,
    required this.codeExpiresAt,
    required this.resendAvailableAt,
  });

  bool isCodeExpiredAt(DateTime instant) => !instant.isBefore(codeExpiresAt);

  bool canResendAt(DateTime instant) => !instant.isBefore(resendAvailableAt);

  Duration resendRemainingAt(DateTime instant) {
    final remaining = resendAvailableAt.difference(instant);

    return remaining.isNegative ? Duration.zero : remaining;
  }
}
