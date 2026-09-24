extension StringExtension on String {
  /// Removes all non-numeric characters from the string.
  String get onlyNumbers => replaceAll(RegExp(r'[^\d]'), '');

  String? trimToNull() {
    final trimmed = trim();
    return trimmed.isEmpty ? null : trimmed;
  }

  /// Checks if the string contains only numeric characters.
  bool get isNumbers => onlyNumbers.length == length;

  /// Validates if the string is a valid CPF (Brazilian individual
  /// taxpayer registry identification).
  bool get isValidCpf {
    final cleaned = onlyNumbers;
    final isValidLength = cleaned.length == 11;

    if (!isValidLength) return false;
    if (RegExp(r'^(\d)\1{10}$').hasMatch(cleaned)) return false;

    final firstNineDigits = cleaned.substring(0, 9);
    final firstCheckDigit = _calculateCPFCheckDigit(firstNineDigits);
    final secondCheckDigit = _calculateCPFCheckDigit(
      firstNineDigits + firstCheckDigit,
    );

    return cleaned.endsWith(firstCheckDigit + secondCheckDigit);
  }

  /// Validates if the string is a valid CNPJ (Brazilian corporate taxpayer
  /// registry identification).
  bool get isValidCnpj {
    final cleaned = onlyNumbers;
    final isValidLength = cleaned.length == 14;

    if (!isValidLength) return false;
    if (RegExp(r'^(\d)\1{13}$').hasMatch(cleaned)) return false;

    final firstTwelveDigits = cleaned.substring(0, 12);
    final firstCheckDigit = _calculateCNPJCheckDigit(firstTwelveDigits);
    final secondCheckDigit = _calculateCNPJCheckDigit(
      firstTwelveDigits + firstCheckDigit,
    );

    return cleaned.endsWith(firstCheckDigit + secondCheckDigit);
  }

  /// Validates if the string is a valid email address.
  bool get isValidEmail {
    final email = trim();
    return RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}',
    ).hasMatch(email);
  }

  bool get hasUppercase => contains(RegExp(r'[A-Z]'));

  bool get hasLowercase => contains(RegExp(r'[a-z]'));

  bool get hasDigit => contains(RegExp(r'\d'));

  bool get isValidPassword {
    final password = trim();
    return password.isNotEmpty &&
        password.length >= 8 &&
        password.length <= 72 &&
        (hasUppercase || hasLowercase) &&
        hasDigit;
  }

  /// Validates if the string is a valid phone number (Brazilian format).
  bool get isValidPhone {
    final cleaned = onlyNumbers;

    // DDD + number
    if (!RegExp(r'^\d{10,11}$').hasMatch(cleaned)) {
      return false;
    }

    final ddd = cleaned.substring(0, 2);
    final number = cleaned.substring(2);

    // Invalid DDD
    if (ddd.startsWith('0')) {
      return false;
    }

    // Mobile: 9 digits starting with 9
    if (number.length == 9) {
      return RegExp(r'^9\d{8}$').hasMatch(number);
    }

    // Landline: 8 digits starting with 2-5
    return RegExp(r'^[2-5]\d{7}$').hasMatch(number);
  }

  // --------------------------------------------------------------
  // Private helper methods for check digit calculation
  // --------------------------------------------------------------
  // Calculates the check digit for a given CNPJ or CPF string.
  String _calculateCNPJCheckDigit(String digits) {
    final length = digits.length;
    final weights = switch (length) {
      12 => [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2],
      13 => [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2],
      _ => throw ArgumentError('CNPJ digits length must be 12 or 13'),
    };

    final sum = List.generate(
      length,
      (index) => int.parse(digits[index]) * weights[index],
    ).reduce((a, b) => a + b);
    final remainder = sum % 11;
    return (remainder < 2) ? '0' : (11 - remainder).toString();
  }

  // Calculates the check digit for a given CPF string.
  String _calculateCPFCheckDigit(String digits) {
    final length = digits.length;
    final sum = List.generate(
      length,
      (index) => int.parse(digits[index]) * (length + 1 - index),
    ).reduce((a, b) => a + b);
    final remainder = sum % 11;
    return (remainder < 2) ? '0' : (11 - remainder).toString();
  }
}
