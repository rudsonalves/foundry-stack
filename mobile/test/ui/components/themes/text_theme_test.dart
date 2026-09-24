import 'package:foundry_stack_mobile/ui/components/themes/text_theme.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('createTextTheme', () {
    final base = ThemeData.light().textTheme;
    final result = createTextTheme(base);

    test('uses Google Sans for display, headline, and title styles', () {
      expect(result.displayLarge?.fontFamily, 'Google Sans');
      expect(result.displayMedium?.fontFamily, 'Google Sans');
      expect(result.displaySmall?.fontFamily, 'Google Sans');
      expect(result.headlineLarge?.fontFamily, 'Google Sans');
      expect(result.headlineMedium?.fontFamily, 'Google Sans');
      expect(result.headlineSmall?.fontFamily, 'Google Sans');
      expect(result.titleLarge?.fontFamily, 'Google Sans');
      expect(result.titleMedium?.fontFamily, 'Google Sans');
      expect(result.titleSmall?.fontFamily, 'Google Sans');
    });

    test('uses Quicksand for body and label styles', () {
      expect(result.bodyLarge?.fontFamily, 'Quicksand');
      expect(result.bodyMedium?.fontFamily, 'Quicksand');
      expect(result.bodySmall?.fontFamily, 'Quicksand');
      expect(result.labelLarge?.fontFamily, 'Quicksand');
      expect(result.labelMedium?.fontFamily, 'Quicksand');
      expect(result.labelSmall?.fontFamily, 'Quicksand');
    });

    test('preserves the remaining properties from the base theme', () {
      expect(result.displayLarge?.fontSize, base.displayLarge?.fontSize);
      expect(
        result.headlineMedium?.fontWeight,
        base.headlineMedium?.fontWeight,
      );
      expect(result.titleSmall?.height, base.titleSmall?.height);
      expect(result.bodyLarge?.fontSize, base.bodyLarge?.fontSize);
      expect(result.bodyMedium?.fontWeight, base.bodyMedium?.fontWeight);
      expect(result.labelSmall?.letterSpacing, base.labelSmall?.letterSpacing);
    });
  });
}
