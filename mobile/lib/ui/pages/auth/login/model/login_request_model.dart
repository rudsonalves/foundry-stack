class LoginRequestModel {
  final String email;
  final String password;

  LoginRequestModel({
    this.email = '',
    this.password = '',
  });

  LoginRequestModel copyWith({
    String? email,
    String? password,
  }) {
    return LoginRequestModel(
      email: email?.trim() ?? this.email,
      password: password?.trim() ?? this.password,
    );
  }
}
