import 'package:foundry_stack_mobile/data/services/apis/user/dtos/create_user_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('CreateUserRequestDto serializes the API request body', () {
    const dto = CreateUserRequestDto(
      name: 'Ada Lovelace',
      email: 'ada@example.com',
      emailVerificationToken: 'verification-proof',
      password: 'secret123',
    );

    expect(dto.toMap(), {
      'name': 'Ada Lovelace',
      'email': 'ada@example.com',
      'email_verification_token': 'verification-proof',
      'password': 'secret123',
    });
  });
}
