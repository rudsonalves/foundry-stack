import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/extensions/string.dart';

void main() {
  group('StringExtension.isValidCpf', () {
    test('returns true for a valid CPF without formatting', () {
      expect('52998224725'.isValidCpf, isTrue);
    });

    test('returns true for a valid CPF with formatting', () {
      expect('529.982.247-25'.isValidCpf, isTrue);
    });

    test('returns false for invalid CPF check digits', () {
      expect('52998224724'.isValidCpf, isFalse);
    });

    test('returns false for repeated CPF digits', () {
      expect('11111111111'.isValidCpf, isFalse);
    });

    test('returns false for CPF with invalid length', () {
      expect('1234567890'.isValidCpf, isFalse);
      expect('123456789012'.isValidCpf, isFalse);
    });

    test(
      'returns false for non-numeric content that becomes invalid length',
      () {
        expect('abc'.isValidCpf, isFalse);
      },
    );
  });

  group('StringExtension.isValidCnpj', () {
    test('returns true for a valid CNPJ without formatting', () {
      expect('04252011000110'.isValidCnpj, isTrue);
    });

    test('returns true for a valid CNPJ with formatting', () {
      expect('04.252.011/0001-10'.isValidCnpj, isTrue);
    });

    test('returns false for invalid CNPJ check digits', () {
      expect('04252011000111'.isValidCnpj, isFalse);
    });

    test('returns false for repeated CNPJ digits', () {
      expect('11111111111111'.isValidCnpj, isFalse);
    });

    test('returns false for CNPJ with invalid length', () {
      expect('1234567890123'.isValidCnpj, isFalse);
      expect('123456789012345'.isValidCnpj, isFalse);
    });

    test(
      'returns false for non-numeric content that becomes invalid length',
      () {
        expect('cnpj'.isValidCnpj, isFalse);
      },
    );
  });

  group('StringExtension.isValidEmail', () {
    test('returns true for valid email', () {
      expect('user@example.com'.isValidEmail, isTrue);
    });

    test('returns true for valid email with plus and subdomain', () {
      expect('user.name+tag@mail.example.com'.isValidEmail, isTrue);
    });

    test('returns true for valid email with surrounding spaces', () {
      expect('  user@example.com  '.isValidEmail, isTrue);
    });

    test('returns false for email without at symbol', () {
      expect('user.example.com'.isValidEmail, isFalse);
    });

    test('returns false for email without domain', () {
      expect('user@'.isValidEmail, isFalse);
    });

    test('returns false for email without top-level domain', () {
      expect('user@example'.isValidEmail, isFalse);
    });
  });

  group('StringExtension.isValidPassword', () {
    test('returns true for password with lowercase and digit', () {
      expect('abcde123'.isValidPassword, isTrue);
    });

    test('returns true for password with uppercase and digit', () {
      expect('ABCDE123'.isValidPassword, isTrue);
    });

    test('returns true for password with mixed case and digit', () {
      expect('Abcde123'.isValidPassword, isTrue);
    });

    test('returns true for password with surrounding spaces after trim', () {
      expect('  Abcde123  '.isValidPassword, isTrue);
    });

    test('returns false for empty password', () {
      expect(''.isValidPassword, isFalse);
      expect('   '.isValidPassword, isFalse);
    });

    test('returns false for password shorter than minimum length', () {
      expect('A1b2'.isValidPassword, isFalse);
    });

    test('returns false for password without digits', () {
      expect('abcdefgh'.isValidPassword, isFalse);
      expect('ABCDEFGH'.isValidPassword, isFalse);
    });

    test('returns false for password without letters', () {
      expect('12345678'.isValidPassword, isFalse);
    });
  });

  group('StringExtension.isValidPhone', () {
    test('returns true for valid mobile phone without formatting', () {
      expect('11987654321'.isValidPhone, isTrue);
    });

    test('returns true for valid mobile phone with formatting', () {
      expect('(11) 98765-4321'.isValidPhone, isTrue);
    });

    test('returns true for valid landline phone', () {
      expect('1132345678'.isValidPhone, isTrue);
    });

    test('returns false for DDD starting with zero', () {
      expect('0132345678'.isValidPhone, isFalse);
      expect('01987654321'.isValidPhone, isFalse);
    });

    test('returns false for mobile phone not starting with 9', () {
      expect('11876543210'.isValidPhone, isFalse);
    });

    test('returns false for landline starting outside 2-5 range', () {
      expect('1162345678'.isValidPhone, isFalse);
      expect('1172345678'.isValidPhone, isFalse);
      expect('1182345678'.isValidPhone, isFalse);
      expect('1192345678'.isValidPhone, isFalse);
    });

    test('returns false for invalid phone length', () {
      expect('1198765432'.isValidPhone, isFalse);
      expect('119876543210'.isValidPhone, isFalse);
    });

    test(
      'returns false for non-numeric content that becomes invalid length',
      () {
        expect('telefone'.isValidPhone, isFalse);
      },
    );
  });
}
