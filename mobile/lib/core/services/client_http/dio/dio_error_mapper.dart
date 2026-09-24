import 'package:dio/dio.dart';

import '../../../result/result.dart';

AppError mapHttpError(Object error) {
  if (error is AppError) {
    return error;
  }

  if (error is! DioException) {
    return AppError(
      code: AppErrorCode.unexpected,
      message: error.toString(),
      details: error,
    );
  }

  final response = error.response;

  switch (error.type) {
    case DioExceptionType.connectionTimeout:
    case DioExceptionType.receiveTimeout:
    case DioExceptionType.sendTimeout:
      return AppError(
        statusCode: response?.statusCode,
        code: AppErrorCode.timeout,
        message: 'Tempo limite da conexão excedido.',
      );
    case DioExceptionType.connectionError:
      return const AppError(
        code: AppErrorCode.networkError,
        message: 'Não foi possível conectar à internet.',
      );
    case DioExceptionType.cancel:
      return const AppError(
        code: AppErrorCode.unexpected,
        message: 'Requisição cancelada.',
      );
    default:
      break;
  }

  final data = response?.data;
  final apiError = _apiError(data);
  final statusCode = response?.statusCode;

  return AppError(
    statusCode: statusCode,
    code: _codeForStatus(statusCode),
    message:
        apiError?['message']?.toString() ??
        error.message ??
        'Erro ao realizar a requisição.',
    details: apiError ?? data,
  );
}

Map<String, dynamic>? _apiError(Object? data) {
  if (data is! Map) return null;
  final error = data['error'];
  if (error is! Map) return null;
  return error.map((key, value) => MapEntry(key.toString(), value));
}

AppErrorCode _codeForStatus(int? statusCode) => switch (statusCode) {
  401 => AppErrorCode.unauthenticated,
  403 => AppErrorCode.forbidden,
  404 => AppErrorCode.notFound,
  409 => AppErrorCode.conflict,
  400 || 422 => AppErrorCode.invalidData,
  _ => AppErrorCode.httpError,
};
