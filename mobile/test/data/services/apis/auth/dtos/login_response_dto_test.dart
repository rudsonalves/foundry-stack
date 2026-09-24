import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/dtos/login_response_dto.dart';

void main() {
  group('LoginResponseDto', () {
    test('fromMap parses the response body', () {
      final dto = LoginResponseDto.fromMap({
        'access_token': 'access-token',
        'refresh_token': 'refresh-token',
        'token_type': 'Bearer',
        'expires_in': 3600,
      });

      expect(dto.accessToken, 'access-token');
      expect(dto.refreshToken, 'refresh-token');
      expect(dto.tokenType, 'Bearer');
      expect(dto.expiresIn, const Duration(seconds: 3600));
    });

    test('fromMap throws when refresh_token is missing', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when access_token is missing', () {
      expect(
        () => LoginResponseDto.fromMap({
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when token_type is missing', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': 'refresh-token',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is missing', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when token_type is invalid', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': 'refresh-token',
          'token_type': 'InvalidTokenType',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is not int', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
          'expires_in': 'InvalidExpiresIn',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is negative', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
          'expires_in': -200,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap accepts expires_in equal to zero', () {
      final dto = LoginResponseDto.fromMap({
        'access_token': 'access-token',
        'refresh_token': 'refresh-token',
        'token_type': 'Bearer',
        'expires_in': 0,
      });

      expect(dto.expiresIn, Duration.zero);
    });

    test('fromMap throws when access_token is empty', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': '',
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when access_token is spaces', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': '   ',
          'refresh_token': 'refresh-token',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when refresh_token is empty', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': '',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when refresh_token is spaces', () {
      expect(
        () => LoginResponseDto.fromMap({
          'access_token': 'access-token',
          'refresh_token': '   ',
          'token_type': 'Bearer',
          'expires_in': 3600,
        }),
        throwsA(isA<Exception>()),
      );
    });
  });
}
