class EmailVerificationProof {
  final String token;
  final DateTime expiresAt;

  const EmailVerificationProof({
    required this.token,
    required this.expiresAt,
  });

  bool isExpiredAt(DateTime instant) => !instant.isBefore(expiresAt);
}
