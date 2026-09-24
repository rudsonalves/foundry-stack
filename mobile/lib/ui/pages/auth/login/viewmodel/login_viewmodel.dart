import '/core/extensions/string.dart';
import '/core/result/command.dart';
import '/data/repositories/auth/auth_repository.dart';
import '/domain/common/auth/models/login_credentials.dart';
import '../model/login_request_model.dart';

class LoginViewmodel {
  final AuthRepository _authRepository;

  LoginViewmodel(this._authRepository) {
    loginCommand = Command1<Unit, LoginRequestModel>(
      _login,
    );
  }
  late final Command1<Unit, LoginRequestModel> loginCommand;

  AsyncResult<Unit> _login(LoginRequestModel request) async {
    final credentials = LoginCredentials(
      email: request.email,
      password: request.password,
    );

    return await _authRepository.login(credentials);
  }

  bool isValid(LoginRequestModel model) {
    final email = model.email.trim();
    final password = model.password.trim();

    return email.isValidEmail && password.isNotEmpty;
  }
}
