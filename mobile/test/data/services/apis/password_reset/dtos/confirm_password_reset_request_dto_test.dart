import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/confirm_password_reset_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('serializes the confirmation preserving leading zeros', () {
    const dto = ConfirmPasswordResetRequestDto(
      passwordResetId: 'password-reset-id',
      code: '012345',
    );

    expect(dto.toMap(), {
      'password_reset_id': 'password-reset-id',
      'code': '012345',
    });
  });
}
