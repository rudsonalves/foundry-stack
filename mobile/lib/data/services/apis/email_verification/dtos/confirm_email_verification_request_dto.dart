class ConfirmEmailVerificationRequestDto {
  final String verificationId;
  final String code;

  const ConfirmEmailVerificationRequestDto({
    required this.verificationId,
    required this.code,
  });

  Map<String, dynamic> toMap() {
    return {'verification_id': verificationId, 'code': code};
  }
}
