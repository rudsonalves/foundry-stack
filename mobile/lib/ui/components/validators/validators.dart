import '/core/extensions/string.dart';

abstract final class Validators {
  static String? email(String? value) {
    final email = value?.trim() ?? '';

    if (email.isValidEmail) return null;

    return 'E-mail inválido';
  }

  static String? password(String? value) {
    final password = value?.trim() ?? '';

    if (password.length < 8) {
      return 'Senha deve ter pelo menos 8 caracteres';
    }

    if (!password.hasUppercase && !password.hasLowercase) {
      return 'Senha deve conter pelo menos uma letra';
    }

    if (!password.hasDigit) {
      return 'Senha deve conter pelo menos um número';
    }

    if (password.length > 72) {
      return 'Senha deve ter no máximo 72 caracteres';
    }

    return null;
  }

  static String? notEmpty(String? value) {
    final text = value?.trim() ?? '';

    if (text.isEmpty) {
      return 'Campo obrigatório';
    }

    return null;
  }

  static String? confirmPassword(String? value, String password) {
    final confirmPassword = value?.trim() ?? '';

    if (confirmPassword != password) {
      return 'As senhas não coincidem';
    }

    return null;
  }
}
