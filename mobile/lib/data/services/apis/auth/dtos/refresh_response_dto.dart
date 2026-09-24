class RefreshResponseDto {
  final String accessToken;
  final String tokenType;
  final Duration expiresIn;

  RefreshResponseDto({
    required this.accessToken,
    required this.tokenType,
    required this.expiresIn,
  });

  factory RefreshResponseDto.fromMap(Map<String, dynamic> map) {
    if (!map.containsKey('access_token') ||
        !map.containsKey('token_type') ||
        !map.containsKey('expires_in') ||
        map['token_type'] != 'Bearer' ||
        map['access_token'] is! String ||
        map['access_token'].trim().isEmpty ||
        map['token_type'] is! String ||
        map['token_type'].trim().isEmpty ||
        map['expires_in'] is! int ||
        map['expires_in'] < 0) {
      throw Exception('Invalid map structure for RefreshResponseDto');
    }
    return RefreshResponseDto(
      accessToken: map['access_token'] as String,
      tokenType: map['token_type'] as String,
      expiresIn: Duration(seconds: map['expires_in'] as int),
    );
  }
}
