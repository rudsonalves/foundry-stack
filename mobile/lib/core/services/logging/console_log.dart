import 'package:flutter/foundation.dart';

enum LogLevel { error, warn, info }

final class ConsoleLog {
  final String context;

  const ConsoleLog(this.context);

  void error(String message, {Object? error, StackTrace? stack}) {
    if (!kDebugMode) return;
    debugPrint('[ERROR][$context] $message');
    if (error != null) debugPrint(error.toString());
    if (stack != null) debugPrint(stack.toString());
  }

  void warn(String message) {
    if (kDebugMode) debugPrint('[WARN][$context] $message');
  }

  void info(String message) {
    if (kDebugMode) debugPrint('[INFO][$context] $message');
  }
}
