class PasswordResetRequestDto {
  final String email;

  const PasswordResetRequestDto({
    required this.email,
  });

  Map<String, dynamic> toMap() {
    return {
      'email': email,
    };
  }
}
