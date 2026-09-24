import '/core/resources/app_env.dart';

class LoginRequestDto {
  final String clientId;
  final String email;
  final String password;

  LoginRequestDto({
    required this.email,
    required this.password,
  }) : clientId = AppEnv.authClientId;

  Map<String, dynamic> toMap() {
    return {
      'client_id': clientId,
      'email': email,
      'password': password,
    };
  }
}
