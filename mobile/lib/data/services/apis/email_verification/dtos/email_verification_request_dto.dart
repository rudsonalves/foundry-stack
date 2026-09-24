class EmailVerificationRequestDto {
  final String email;

  const EmailVerificationRequestDto({required this.email});

  Map<String, dynamic> toMap() {
    return {'email': email};
  }
}
