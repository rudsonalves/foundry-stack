class AuthSession {
  final String accessToken;
  final String refreshToken;

  const AuthSession({
    required this.accessToken,
    required this.refreshToken,
  });
}
