import 'package:foundry_stack_mobile/data/services/apis/email_verification/dtos/confirm_email_verification_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ConfirmEmailVerificationResponseDto', () {
    const validMap = {
      'email_verification_token': 'opaque-secret',
      'expires_at': '2026-09-11T15:00:00Z',
    };

    test('fromMap parses the response body', () {
      final dto = ConfirmEmailVerificationResponseDto.fromMap(validMap);

      expect(dto.emailVerificationToken, 'opaque-secret');
      expect(dto.expiresAt, DateTime.parse('2026-09-11T15:00:00Z'));
    });

    for (final field in validMap.keys) {
      test('fromMap throws when $field is missing', () {
        final map = Map<String, dynamic>.from(validMap)..remove(field);

        expect(
          () => ConfirmEmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field has the wrong type', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 1;

        expect(
          () => ConfirmEmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field is blank', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = '   ';

        expect(
          () => ConfirmEmailVerificationResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }

    test('fromMap throws when expires_at is not a valid date-time', () {
      final map = Map<String, dynamic>.from(validMap)
        ..['expires_at'] = 'not-a-date';

      expect(
        () => ConfirmEmailVerificationResponseDto.fromMap(map),
        throwsA(isA<Exception>()),
      );
    });
  });
}
