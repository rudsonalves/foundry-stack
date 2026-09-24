import 'package:foundry_stack_mobile/data/services/apis/password_reset/dtos/confirm_password_reset_response_dto.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('ConfirmPasswordResetResponseDto', () {
    const validMap = {
      'password_reset_token': 'opaque-proof',
      'expires_at': '2026-09-23T16:00:00Z',
    };

    test('fromMap parses the response body and expiration', () {
      final dto = ConfirmPasswordResetResponseDto.fromMap(
        validMap,
      );

      expect(dto.passwordResetToken, 'opaque-proof');
      expect(
        dto.expiresAt,
        DateTime.parse('2026-09-23T16:00:00Z'),
      );
    });

    for (final field in validMap.keys) {
      test('fromMap throws when $field is missing', () {
        final map = Map<String, dynamic>.from(validMap)..remove(field);

        expect(
          () => ConfirmPasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field has the wrong type', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = 1;

        expect(
          () => ConfirmPasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });

      test('fromMap throws when $field is blank', () {
        final map = Map<String, dynamic>.from(validMap)..[field] = '   ';

        expect(
          () => ConfirmPasswordResetResponseDto.fromMap(map),
          throwsA(isA<Exception>()),
        );
      });
    }

    test('fromMap throws when expires_at is invalid', () {
      final map = Map<String, dynamic>.from(validMap)
        ..['expires_at'] = 'not-a-date';

      expect(
        () => ConfirmPasswordResetResponseDto.fromMap(map),
        throwsA(isA<Exception>()),
      );
    });
  });
}
