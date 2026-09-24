abstract final class AppHttpHeaders {
  static const accept = 'Accept';
  static const authorization = 'Authorization';
  static const contentType = 'Content-Type';
  static const requestId = 'X-Request-ID';

  static const sensitiveLowercase = {'authorization'};

  static String bearer(String token) => 'Bearer $token';
}
