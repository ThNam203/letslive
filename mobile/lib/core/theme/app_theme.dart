import 'package:flutter/material.dart';

/// Builds a color from the web app's `H S% L%` CSS variable triplets
/// (`web/app/globals.css`).
Color _hsl(double h, double s, double l) =>
    HSLColor.fromAHSL(1, h, s / 100, l / 100).toColor();

// Light palette, from `:root` in web/app/globals.css.
final ColorScheme lightColorScheme = ColorScheme(
  brightness: Brightness.light,
  primary: _hsl(276, 91, 38),
  onPrimary: _hsl(0, 0, 92),
  secondary: _hsl(28, 80, 52),
  onSecondary: _hsl(0, 0, 100),
  tertiary: _hsl(140, 70, 45),
  onTertiary: _hsl(0, 0, 100),
  error: _hsl(0, 90.3, 60),
  onError: _hsl(0, 0, 100),
  surface: _hsl(0, 0, 100),
  onSurface: _hsl(222, 47, 11),
  onSurfaceVariant: _hsl(222, 10.3, 25.4),
  surfaceContainerHighest: _hsl(210, 40, 96),
  outline: _hsl(210, 10, 23),
);

// Dark palette, from the `.dark` block in web/app/globals.css.
final ColorScheme darkColorScheme = ColorScheme(
  brightness: Brightness.dark,
  primary: _hsl(276, 91, 38),
  onPrimary: _hsl(0, 0, 92),
  secondary: _hsl(28, 90, 60),
  onSecondary: _hsl(0, 0, 0),
  tertiary: _hsl(140, 70, 60),
  onTertiary: _hsl(0, 0, 0),
  error: _hsl(0, 90.3, 60),
  onError: _hsl(0, 0, 90),
  surface: _hsl(222, 47, 11),
  onSurface: _hsl(0, 0, 100),
  onSurfaceVariant: _hsl(0, 0, 80),
  surfaceContainerHighest: _hsl(220, 20, 25),
  outline: _hsl(220, 15, 30),
);

/// Success color shared by both themes (`--success` on the web).
final Color successColor = _hsl(142, 70.6, 45.3);

ThemeData _build(ColorScheme scheme) => ThemeData(
  useMaterial3: true,
  colorScheme: scheme,
  scaffoldBackgroundColor: scheme.surface,
  dividerColor: scheme.outline.withValues(alpha: 0.4),
);

final ThemeData lightTheme = _build(lightColorScheme);
final ThemeData darkTheme = _build(darkColorScheme);
