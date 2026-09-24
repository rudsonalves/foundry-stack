class EmailVerificationChallenge {
  final String verificationId;
  final DateTime codeExpiresAt;
  final DateTime resendAvailableAt;

  const EmailVerificationChallenge({
    required this.verificationId,
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
