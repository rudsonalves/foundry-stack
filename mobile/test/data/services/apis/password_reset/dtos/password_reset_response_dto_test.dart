import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/password_reset_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('PasswordResetResponseDto', () {
    const validMap = {
      'password_reset_id': '7fd1ec52-6f58-4fe2-b482-e31c06adfba5',
      'code_expires_at': '2026-09-23T15:15:00Z',
      'resend_available_at': '2026-09-23T15:01:00Z',
    };

    test('fromMap parses the response body and dates', () {
      final dto = PasswordResetResponseDto.fromMap(validMap);

      expect(dto.passwordResetId, validMap['password_reset_id']);
      expect(
        dto.codeExpiresAt,
        DateTime.parse(validMap['code_expires_at']!),
      );
      expect(
        dto.resendAvailableAt,
        DateTime.parse(validMap['resend_available_at']!),
      );
    });

    for (final field in validMap.keys) {
      test('fromMap throws when $field is missing', () {
        final map = Map<String, dynamic>.from(validMap)..remove(field);

        expect(
          () => PasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field has the wrong type', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 1;

        expect(
          () => PasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field is blank', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = '   ';

        expect(
          () => PasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }

    for (final field in [
      'code_expires_at',
      'resend_available_at',
    ]) {
      test('fromMap throws when $field is not a valid date', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 'not-a-date';

        expect(
          () => PasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }
  });
}
