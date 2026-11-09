import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/theme/app_theme.dart';

void main() {
  for (final (name, theme) in [('light', lightTheme), ('dark', darkTheme)]) {
    test('$name theme gives text fields a rounded outline, like the web', () {
      final inputs = theme.inputDecorationTheme;
      final scheme = theme.colorScheme;

      final enabled = inputs.enabledBorder;
      expect(enabled, isA<OutlineInputBorder>());
      enabled as OutlineInputBorder;
      expect(enabled.borderRadius, BorderRadius.circular(6));
      expect(enabled.borderSide.color, scheme.outline);

      final focused = inputs.focusedBorder as OutlineInputBorder;
      expect(focused.borderSide.color, scheme.primary);

      final error = inputs.errorBorder as OutlineInputBorder;
      expect(error.borderSide.color, scheme.error);
    });
  }
}
