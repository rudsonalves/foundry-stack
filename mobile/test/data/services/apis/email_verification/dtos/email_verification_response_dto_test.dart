import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/email_verification_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('EmailVerificationResponseDto', () {
    const validMap = {
      'verification_id': '7fd1ec52-6f58-4fe2-b482-e31c06adfba5',
      'code_expires_at': '2026-09-10T15:15:00Z',
      'resend_available_at': '2026-09-10T15:01:00Z',
    };

    test('fromMap parses the response body', () {
      final dto = EmailVerificationResponseDto.fromMap(validMap);

      expect(dto.verificationId, validMap['verification_id']);
      expect(dto.codeExpiresAt, DateTime.parse(validMap['code_expires_at']!));
      expect(
        dto.resendAvailableAt,
        DateTime.parse(validMap['resend_available_at']!),
      );
    });

    for (final field in validMap.keys) {
      test('fromMap throws when $field is missing', () {
        final map = Map<String, dynamic>.from(validMap)..remove(field);

        expect(
          () => EmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field has the wrong type', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 1;

        expect(
          () => EmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field is blank', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = '   ';

        expect(
          () => EmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }

    for (final field in ['code_expires_at', 'resend_available_at']) {
      test('fromMap throws when $field is not a valid date-time', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 'not-a-date';

        expect(
          () => EmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }
  });
}
