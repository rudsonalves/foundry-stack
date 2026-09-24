import '/domain/common/auth/models/auth_session.dart';
import '/domain/common/auth/models/login_credentials.dart';
import '../dtos/login_request_dto.dart';
import '../dtos/login_response_dto.dart';

abstract final class AuthApiAdapter {
  static LoginRequestDto toLoginRequest(LoginCredentials credentials) {
    return LoginRequestDto(
      email: credentials.email,
      password: credentials.password,
    );
  }

  static AuthSession toAuthSession(LoginResponseDto response) {
    return AuthSession(
      accessToken: response.accessToken,
      refreshToken: response.refreshToken,
    );
  }
}
