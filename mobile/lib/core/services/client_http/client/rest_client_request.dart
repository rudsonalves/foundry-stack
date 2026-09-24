final class RestClientRequest {
  final String path;
  final Map<String, dynamic>? headers;
  final Map<String, dynamic>? queryParameters;
  final Object? body;

  const RestClientRequest(
    this.path, {
    this.headers,
    this.queryParameters,
    this.body,
  });

  RestClientRequest copyWith({
    String? path,
    Map<String, dynamic>? headers,
    Map<String, dynamic>? queryParameters,
    Object? body,
  }) {
    return RestClientRequest(
      path ?? this.path,
      headers: headers ?? this.headers,
      queryParameters: queryParameters ?? this.queryParameters,
      body: body ?? this.body,
    );
  }
}
