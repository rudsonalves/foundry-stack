import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/complete_password_reset_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('serializes only the fields accepted by the API', () {
    const dto = CompletePasswordResetRequestDto(
      passwordResetId: 'password-reset-id',
      passwordResetToken: 'opaque-proof',
      newPassword: 'new-password123',
    );

    expect(dto.toMap(), {
      'password_reset_id': 'password-reset-id',
      'password_reset_token': 'opaque-proof',
      'new_password': 'new-password123',
    });
  });
}
