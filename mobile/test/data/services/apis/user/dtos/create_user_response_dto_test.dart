import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/data/services/apis/user/dtos/create_user_response_dto.dart';

void main() {
  group('CreateUserResponseDto', () {
    test('fromMap parses the response body', () {
      final dto = CreateUserResponseDto.fromMap({
        'id': 'user-1',
        'name': 'Ada Lovelace',
        'email': 'ada@example.com',
      });

      expect(dto.id, 'user-1');
      expect(dto.name, 'Ada Lovelace');
      expect(dto.email, 'ada@example.com');
    });

    test('fromMap throws when id is spaces', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': '   ',
          'name': 'Ada Lovelace',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when name is spaces', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': '   ',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when email is spaces', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': 'Ada Lovelace',
          'email': '   ',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when id is missing', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'name': 'Ada Lovelace',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when name is missing', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when email is missing', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': 'Ada Lovelace',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when id has the wrong type', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 1,
          'name': 'Ada Lovelace',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when name has the wrong type', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': 1,
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when email has the wrong type', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': 'Ada Lovelace',
          'email': 1,
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when id is empty', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': '',
          'name': 'Ada Lovelace',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when name is empty', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': '',
          'email': 'ada@example.com',
        }),
        throwsA(isA<Exception>()),
      );
    });

    test('fromMap throws when email is empty', () {
      expect(
        () => CreateUserResponseDto.fromMap({
          'id': 'user-1',
          'name': 'Ada Lovelace',
          'email': '',
        }),
        throwsA(isA<Exception>()),
      );
    });
  });
}
