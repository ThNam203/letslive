import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/auth/domain/auth_validators.dart';

/// A password that passes every signup rule, padded to [length].
String validPassword(int length) => 'Aa!'.padRight(length, 'x');

void main() {
  group('signup validates same field limits as web', () {
    // Limits from web/constant/field-limits.ts and web/constant/password.ts.
    test('limits match the web constants', () {
      expect(emailMaxLength, 320);
      expect(passwordMinLength, 8);
      expect(passwordMaxLength, 72);
      expect(usernameMinLength, 6);
      expect(usernameMaxLength, 30);
    });

    test('email', () {
      final local = 'a' * 308; // 308 + '@example.com' (12) = 320
      expect(validateEmail('$local@example.com'), isNull);
      expect(validateEmail('a$local@example.com'), AuthFieldError.emailTooLong);
      expect(validateEmail(''), AuthFieldError.emailRequired);
      expect(validateEmail('not-an-email'), AuthFieldError.emailInvalid);
      expect(validateEmail('a..b@example.com'), AuthFieldError.emailInvalid);
      expect(validateEmail('user@example.com'), isNull);
    });

    test('username', () {
      expect(validateUsername(''), AuthFieldError.usernameTooShort);
      expect(validateUsername('a' * 5), AuthFieldError.usernameTooShort);
      expect(validateUsername('a' * 6), isNull);
      expect(validateUsername('a' * 30), isNull);
      expect(validateUsername('a' * 31), AuthFieldError.usernameTooLong);
    });

    test('password', () {
      expect(validateSignupPassword(validPassword(8)), isNull);
      expect(validateSignupPassword(validPassword(72)), isNull);
      expect(
        validateSignupPassword(validPassword(7)),
        AuthFieldError.passwordTooShort,
      );
      expect(
        validateSignupPassword(validPassword(73)),
        AuthFieldError.passwordTooLong,
      );
      expect(
        validateSignupPassword('AA!XXXXX'),
        AuthFieldError.passwordMissingLowercase,
      );
      expect(
        validateSignupPassword('aa!xxxxx'),
        AuthFieldError.passwordMissingUppercase,
      );
      expect(
        validateSignupPassword('Aa1xxxxx'),
        AuthFieldError.passwordMissingSpecial,
      );
    });

    test('like zod, the last failing rule is the one reported', () {
      // Empty fails every rule; zod reports the last one (missing special).
      expect(validateSignupPassword(''), AuthFieldError.passwordMissingSpecial);
      expect(validateLoginPassword(''), AuthFieldError.passwordTooShort);
    });

    test('confirm password', () {
      final password = validPassword(10);
      expect(validateConfirmPassword(password, password), isNull);
      expect(
        validateConfirmPassword(validPassword(11), password),
        AuthFieldError.passwordsDoNotMatch,
      );
      expect(
        validateConfirmPassword('', password),
        AuthFieldError.passwordsDoNotMatch,
      );
    });

    test('account setup username is trimmed and needs the minimum length', () {
      expect(
        validateAccountSetupUsername('  abcde  '),
        AuthFieldError.usernameTooShort,
      );
      expect(validateAccountSetupUsername(' abcdef '), isNull);
    });
  });

  group('login', () {
    test('password only checks length', () {
      expect(validateLoginPassword('alllowercase'), isNull);
      expect(validateLoginPassword('x' * 7), AuthFieldError.passwordTooShort);
      expect(validateLoginPassword('x' * 73), AuthFieldError.passwordTooLong);
    });
  });
}
