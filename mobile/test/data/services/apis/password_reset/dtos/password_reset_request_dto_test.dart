import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/password_reset_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('PasswordResetRequestDto serializes the API request body', () {
    const dto = PasswordResetRequestDto(
      email: 'user@example.com',
    );

    expect(dto.toMap(), {
      'email': 'user@example.com',
    });
  });
}
