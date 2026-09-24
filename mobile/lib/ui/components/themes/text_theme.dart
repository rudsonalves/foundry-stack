import 'package:material_ui/material_ui.dart';

const _bodyFontFamily = 'Quicksand';
const _displayFontFamily = 'Google Sans';

TextTheme createTextTheme(TextTheme baseTextTheme) {
  return baseTextTheme.copyWith(
    displayLarge: baseTextTheme.displayLarge?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    displayMedium: baseTextTheme.displayMedium?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    displaySmall: baseTextTheme.displaySmall?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    headlineLarge: baseTextTheme.headlineLarge?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    headlineMedium: baseTextTheme.headlineMedium?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    headlineSmall: baseTextTheme.headlineSmall?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    titleLarge: baseTextTheme.titleLarge?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    titleMedium: baseTextTheme.titleMedium?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    titleSmall: baseTextTheme.titleSmall?.copyWith(
      fontFamily: _displayFontFamily,
    ),
    bodyLarge: baseTextTheme.bodyLarge?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
    bodyMedium: baseTextTheme.bodyMedium?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
    bodySmall: baseTextTheme.bodySmall?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
    labelLarge: baseTextTheme.labelLarge?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
    labelMedium: baseTextTheme.labelMedium?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
    labelSmall: baseTextTheme.labelSmall?.copyWith(
      fontFamily: _bodyFontFamily,
    ),
  );
}
