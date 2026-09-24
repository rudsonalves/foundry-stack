import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/email_verification_request_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('EmailVerificationRequestDto serializes the API request body', () {
    const dto = EmailVerificationRequestDto(email: 'ada@example.com');

    expect(dto.toMap(), {'email': 'ada@example.com'});
  });
}
