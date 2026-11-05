import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/wallet/domain/decimal_amount.dart';

DecimalAmount _d(String text) => DecimalAmount.parse(text);

void main() {
  group('parse', () {
    test('reads the server formats exactly', () {
      expect(_d('15.00').units, BigInt.from(1500));
      expect(_d('15.00').scale, 2);
      expect(_d('-1.00').units, BigInt.from(-100));
      expect(_d('7').scale, 0);
    });

    test('rejects anything that is not plain decimal notation', () {
      for (final bad in ['', ' 1', '1 ', '+1', '1e5', '1,5', '.5', '1.', 'x']) {
        expect(
          () => DecimalAmount.parse(bad),
          throwsFormatException,
          reason: '"$bad"',
        );
      }
    });
  });

  group('arithmetic', () {
    test('adds exactly where doubles drift', () {
      // 0.1 + 0.2 is 0.30000000000000004 as a double.
      expect((_d('0.10') + _d('0.20')).format('en'), '0.3');
      expect((_d('0.1') + _d('0.2')).format('en'), '0.3');
    });

    test('adds amounts of different scales', () {
      expect((_d('1') + _d('0.25')).format('en'), '1.25');
    });

    test('is exact beyond double precision', () {
      final big = _d('9007199254740993.01') + _d('0.01');
      expect(big.format('en'), '9,007,199,254,740,993.02');
    });
  });

  group('format', () {
    test('groups, trims trailing zeros and keeps up to two digits', () {
      expect(_d('1234567.50').format('en'), '1,234,567.5');
      expect(_d('15.00').format('en'), '15');
      expect(_d('0.05').format('en'), '0.05');
      expect(_d('0').format('en'), '0');
    });

    test('rounds half away from zero past two digits', () {
      expect(_d('1.005').format('en'), '1.01');
      expect(_d('-1.005').format('en'), '-1.01');
      expect(_d('1.004').format('en'), '1');
    });

    test('shows the sign: minus always, plus on request', () {
      expect(_d('-3.50').format('en'), '-3.5');
      expect(_d('3.50').format('en', withPlus: true), '+3.5');
      expect(_d('0').format('en', withPlus: true), '+0');
    });

    test('follows the locale separators', () {
      expect(_d('1234.5').format('vi'), '1.234,5');
    });
  });
}
