/// Exact decimal amounts for the wallet.
///
/// The server sends amounts as decimal strings built from integer minor units
/// (`"15.00"`, `"-1.00"`, `"7"`; `backend/finance/dto/amount.go`). Parsing
/// them to `double` (as web's `parseFloat` does) can drift, so amounts are
/// kept as an integer count of 10^-[scale] units and only turned into text
/// at the very end.
library;

import 'package:flutter/foundation.dart';
import 'package:intl/intl.dart';

/// Digits after the point web shows at most (`maximumFractionDigits: 2` in
/// `web/app/(main)/wallet/_components/balance-card.tsx`).
const int _maxFractionDigits = 2;

final RegExp _decimalPattern = RegExp(r'^(-?)(\d+)(?:\.(\d+))?$');
final RegExp _trailingZeros = RegExp(r'0+$');

/// A signed decimal number: [units] x 10^-[scale]. Never uses floating point.
@immutable
class DecimalAmount implements Comparable<DecimalAmount> {
  const DecimalAmount._(this.units, this.scale);

  /// Reads the server's plain decimal notation: optional `-`, digits, an
  /// optional point with digits. Anything else (exponent, `+`, spaces, empty)
  /// throws [FormatException], so a changed backend format fails loudly.
  factory DecimalAmount.parse(String text) {
    final match = _decimalPattern.firstMatch(text);
    if (match == null) {
      throw FormatException('Not a decimal amount: "$text"');
    }
    final fraction = match.group(3) ?? '';
    final digits = BigInt.parse('${match.group(2)}$fraction');
    return DecimalAmount._(
      match.group(1) == '-' ? -digits : digits,
      fraction.length,
    );
  }

  /// Zero.
  static final DecimalAmount zero = DecimalAmount._(BigInt.zero, 0);

  /// The amount as an integer count of 10^-[scale].
  final BigInt units;

  /// Number of digits after the decimal point in [units].
  final int scale;

  /// True below zero.
  bool get isNegative => units.isNegative;

  /// The exact sum; the result keeps the larger scale.
  DecimalAmount operator +(DecimalAmount other) {
    final common = scale > other.scale ? scale : other.scale;
    return DecimalAmount._(_rescale(common) + other._rescale(common), common);
  }

  BigInt _rescale(int to) => units * BigInt.from(10).pow(to - scale);

  @override
  int compareTo(DecimalAmount other) {
    final common = scale > other.scale ? scale : other.scale;
    return _rescale(common).compareTo(other._rescale(common));
  }

  @override
  bool operator ==(Object other) =>
      other is DecimalAmount && compareTo(other) == 0;

  @override
  int get hashCode => _rescale(_maxFractionDigits + scale).hashCode;

  /// The amount as web shows it: grouped for [locale], at most two fraction
  /// digits (rounded half away from zero), no trailing zeros. A leading `-`
  /// for negative amounts; [withPlus] adds `+` to the others (transaction
  /// rows).
  String format(String locale, {bool withPlus = false}) {
    final symbols = NumberFormat.decimalPattern(locale).symbols;
    final shown = _roundedTo(_maxFractionDigits);
    final absolute = shown.units.abs();
    final divisor = BigInt.from(10).pow(shown.scale);
    final whole = _group(absolute ~/ divisor, symbols.GROUP_SEP);
    final fraction = shown.scale == 0
        ? ''
        : (absolute % divisor)
              .toString()
              .padLeft(shown.scale, '0')
              .replaceFirst(_trailingZeros, '');
    final sign = shown.isNegative ? '-' : (withPlus ? '+' : '');
    final point = fraction.isEmpty ? '' : '${symbols.DECIMAL_SEP}$fraction';
    return '$sign$whole$point';
  }

  /// This value with at most [digits] fraction digits, rounded half away
  /// from zero (what `Intl.NumberFormat` does).
  DecimalAmount _roundedTo(int digits) {
    if (scale <= digits) return this;
    final divisor = BigInt.from(10).pow(scale - digits);
    final magnitude = (units.abs() + divisor ~/ BigInt.two) ~/ divisor;
    return DecimalAmount._(isNegative ? -magnitude : magnitude, digits);
  }

  @override
  String toString() => format('en');
}

/// [digits] (a non-negative whole number) with [separator] between groups of
/// three.
String _group(BigInt digits, String separator) {
  final text = digits.toString();
  final out = StringBuffer();
  for (var i = 0; i < text.length; i++) {
    if (i > 0 && (text.length - i) % 3 == 0) out.write(separator);
    out.write(text[i]);
  }
  return out.toString();
}
