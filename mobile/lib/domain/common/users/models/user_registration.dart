class UserRegistration {
  final String name;
  final String email;
  final String password;
  final String emailVerificationToken;

  const UserRegistration({
    required this.name,
    required this.email,
    required this.password,
    required this.emailVerificationToken,
  });
}
