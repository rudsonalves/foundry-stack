class CreateUserRequestDto {
  final String name;
  final String email;
  final String emailVerificationToken;
  final String password;

  const CreateUserRequestDto({
    required this.name,
    required this.email,
    required this.emailVerificationToken,
    required this.password,
  });

  Map<String, dynamic> toMap() {
    return {
      'name': name,
      'email': email,
      'email_verification_token': emailVerificationToken,
      'password': password,
    };
  }
}
