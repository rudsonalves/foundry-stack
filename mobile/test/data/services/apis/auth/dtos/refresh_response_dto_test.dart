import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/data/services/apis/auth/dtos/refresh_response_dto.dart';

void main() {
  group('RefreshResponseDto', () {
    test('fromMap parses the response body', () {
      final dto = RefreshResponseDto.fromMap({
        'access_token': 'access-token',
        'token_type': 'Bearer',
        'expires_in': 1800,
      });

      expect(dto.accessToken, 'access-token');
      expect(dto.tokenType, 'Bearer');
      expect(dto.expiresIn, const Duration(seconds: 1800));
    });

    test('fromMap throws when access_token is missing', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'token_type': 'Bearer',
          'expires_in': 1800,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when token_type is missing', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'expires_in': 1800,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is missing', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 'Bearer',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when token_type is not Bearer', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 'Basic',
          'expires_in': 1800,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is not int', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 'Bearer',
          'expires_in': '1800',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when expires_in is negative', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 'Bearer',
          'expires_in': -1,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap accepts expires_in equal to zero', () {
      final dto = RefreshResponseDto.fromMap({
        'access_token': 'access-token',
        'token_type': 'Bearer',
        'expires_in': 0,
      });

      expect(dto.expiresIn, Duration.zero);
    });

    test('fromMap throws when access_token is not String', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 123,
          'token_type': 'Bearer',
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });

    test('fromMap throws when token_type is not String', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': 123,
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });

    test('fromMap throws when acess_token is empty', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': '',
          'token_type': 'Bearer',
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });

    test('fromMap throws when access_token is spaces', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': '  ',
          'token_type': 'Bearer',
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });

    test('fromMap throws when token_type is empty', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': '',
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });

    test('fromMap throws when token_type is spaces', () {
      expect(
        () => RefreshResponseDto.fromMap({
          'access_token': 'access-token',
          'token_type': '   ',
          'expires_in': 1800,
        }),
        throwsA(anything),
      );
    });
  });
}
