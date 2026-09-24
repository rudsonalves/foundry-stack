import 'package:flutter/services.dart';

abstract final class Formatters {
  static final noSpaces = FilteringTextInputFormatter.deny(RegExp(r'\s'));
}
