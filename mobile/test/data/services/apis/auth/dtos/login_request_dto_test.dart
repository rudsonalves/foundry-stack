import 'package:foundry_stack_mobile/core/resources/app_env.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/dtos/login_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('LoginRequestDto serializes the API request body', () {
    final dto = LoginRequestDto(
      email: 'ada@example.com',
      password: 'secret123',
    );

    expect(dto.toMap(), {
      'client_id': AppEnv.authClientId,
      'email': 'ada@example.com',
      'password': 'secret123',
    });
  });
}
