import '/core/result/result.dart';
import '/core/services/client_http/client/rest_client.dart';
import '/core/services/client_http/client/rest_client_request.dart';
import '/core/services/logging/console_log.dart';
import '/domain/common/auth/models/auth_session.dart';
import '/domain/common/auth/models/login_credentials.dart';
import '../core/api_response_parser.dart';
import 'adapters/auth_api_adapter.dart';
import 'dtos/login_response_dto.dart';
import 'dtos/refresh_request_dto.dart';
import 'dtos/refresh_response_dto.dart';

class AuthService {
  final RestClient _client;

  AuthService(this._client);

  final _log = ConsoleLog('AuthService');

  AsyncResult<AuthSession> login(LoginCredentials credentials) async {
    final response = await _client.post(
      RestClientRequest(
        '/auth/login',
        body: AuthApiAdapter.toLoginRequest(credentials).toMap(),
      ),
    );

    if (response.isFailure) {
      _log.error('login failed: ${response.error}');
      return Failure(response.error!);
    }

    final resp = response.value!;

    if (resp.statusCode != 200) {
      _log.error(
        'login failed: ${resp.statusCode} ${resp.statusMessage}',
      );

      return Failure(
        AppError(
          statusCode: resp.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 200, received ${resp.statusCode}',
        ),
      );
    }

    final result = ApiResponseParser.parse<LoginResponseDto>(
      resp.data,
      LoginResponseDto.fromMap,
    );

    if (result.isFailure) {
      _log.error('login response parsing failed: ${result.error}');
    }

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return Success(AuthApiAdapter.toAuthSession(result.value!));
  }

  AsyncResult<String> refreshToken(String refreshToken) async {
    final response = await _client.post(
      RestClientRequest(
        '/auth/refresh',
        body: RefreshRequestDto(refreshToken: refreshToken).toMap(),
      ),
    );

    if (response.isFailure) {
      _log.error('refreshToken failed: ${response.error}');
      return Failure(response.error!);
    }

    final resp = response.value!;

    if (resp.statusCode != 200) {
      _log.error(
        'refreshToken failed: ${resp.statusCode} ${resp.statusMessage}',
      );

      return Failure(
        AppError(
          statusCode: resp.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 200, received ${resp.statusCode}',
        ),
      );
    }

    final result = ApiResponseParser.parse<RefreshResponseDto>(
      resp.data,
      RefreshResponseDto.fromMap,
    );

    if (result.isFailure) {
      _log.error('refreshToken response parsing failed: ${result.error}');
    }

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return Success(result.value!.accessToken);
  }

  AsyncResult<Unit> logout(String refreshToken) async {
    final response = await _client.post(
      RestClientRequest(
        '/auth/logout',
        body: RefreshRequestDto(refreshToken: refreshToken).toMap(),
      ),
    );

    if (response.isFailure) {
      _log.error('logout failed: ${response.error}');
      return Failure(response.error!);
    }

    final resp = response.value!;

    if (resp.statusCode != 204) {
      _log.error(
        'logout failed: ${resp.statusCode} ${resp.statusMessage}',
      );

      return Failure(
        AppError(
          statusCode: resp.statusCode,
          code: AppErrorCode.invalidData,
          message:
              'Unexpected response status: expected 204, received ${resp.statusCode}',
        ),
      );
    }

    return const Success(unit);
  }
}
