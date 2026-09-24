import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/confirm_email_verification_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'ConfirmEmailVerificationRequestDto serializes the API request body',
    () {
      const dto = ConfirmEmailVerificationRequestDto(
        verificationId: 'verification-id',
        code: '012345',
      );

      expect(dto.toMap(), {
        'verification_id': 'verification-id',
        'code': '012345',
      });
    },
  );
}
