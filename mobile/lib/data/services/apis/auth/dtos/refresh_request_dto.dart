class RefreshRequestDto {
  final String refreshToken;

  RefreshRequestDto({
    required this.refreshToken,
  });

  Map<String, dynamic> toMap() {
    return {
      'refresh_token': refreshToken,
    };
  }
}
