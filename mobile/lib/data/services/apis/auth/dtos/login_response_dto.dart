class LoginResponseDto {
  final String accessToken;
  final String refreshToken;
  final String tokenType;
  final Duration expiresIn;

  LoginResponseDto({
    required this.accessToken,
    required this.refreshToken,
    required this.tokenType,
    required this.expiresIn,
  });

  factory LoginResponseDto.fromMap(Map<String, dynamic> map) {
    if (!map.containsKey('access_token') ||
        !map.containsKey('refresh_token') ||
        !map.containsKey('token_type') ||
        !map.containsKey('expires_in') ||
        map['access_token'] is! String ||
        map['access_token'].trim().isEmpty ||
        map['refresh_token'] is! String ||
        map['refresh_token'].trim().isEmpty ||
        map['token_type'] != 'Bearer' ||
        map['expires_in'] is! int ||
        (map['expires_in'] as int) < 0) {
      throw Exception('Invalid map structure for LoginResponseDto');
    }
    return LoginResponseDto(
      accessToken: map['access_token'] as String,
      refreshToken: map['refresh_token'] as String,
      tokenType: map['token_type'] as String,
      expiresIn: Duration(seconds: map['expires_in'] as int),
    );
  }
}
