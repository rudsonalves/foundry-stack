final class RestClientResponse {
  final Object? data;
  final int? statusCode;
  final String? statusMessage;

  const RestClientResponse({this.data, this.statusCode, this.statusMessage});

  bool get isSuccess =>
      statusCode != null && statusCode! >= 200 && statusCode! < 300;

  T parse<T>(T Function(Object? json) fromJson) => fromJson(data);
}
