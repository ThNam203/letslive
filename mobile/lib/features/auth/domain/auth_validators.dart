/// Auth form rules, copied from `web/lib/validations/{login,signUp}.ts` and
/// the account-setup page so mobile accepts and rejects exactly what web
/// does.
///
/// Each validator lists its rules in the web schema's order and reports the
/// LAST failing one: zod runs every check and the web form keeps the last
/// issue per field (e.g. an empty signup password reports "missing special
/// character", as on web).
library;

import 'package:letslive_mobile/core/validation/field_limits.dart';

/// Why a field is invalid; the UI maps each to its `error_*` message.
enum AuthFieldError {
  emailRequired,
  emailInvalid,
  emailTooLong,
  passwordRequired,
  passwordTooShort,
  passwordTooLong,
  passwordMissingLowercase,
  passwordMissingUppercase,
  passwordMissingSpecial,
  confirmPasswordRequired,
  passwordsDoNotMatch,
  usernameRequired,
  usernameTooShort,
  usernameTooLong,
  newPasswordMustDiffer,
}

/// zod v4's `z.email()` pattern.
final RegExp _email = RegExp(
  r"^(?!\.)(?!.*\.\.)([A-Za-z0-9_'+\-\.]*)[A-Za-z0-9_+-]@([A-Za-z0-9][A-Za-z0-9\-]*\.)+[A-Za-z]{2,}$",
);

/// The special characters web's signup schema requires one of.
final RegExp _special = RegExp(r'''[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]''');

/// The last error whose rule fails, or null when all pass.
AuthFieldError? _lastFailing(List<(bool, AuthFieldError)> rules) {
  AuthFieldError? last;
  for (final (failed, error) in rules) {
    if (failed) last = error;
  }
  return last;
}

/// Login and signup email.
AuthFieldError? validateEmail(String email) => _lastFailing([
  (!_email.hasMatch(email), AuthFieldError.emailInvalid),
  (email.isEmpty, AuthFieldError.emailRequired),
  (email.length > FieldLimits.emailMax, AuthFieldError.emailTooLong),
]);

List<(bool, AuthFieldError)> _passwordLength(String password) => [
  (password.isEmpty, AuthFieldError.passwordRequired),
  (password.length < FieldLimits.passwordMin, AuthFieldError.passwordTooShort),
  (password.length > FieldLimits.passwordMax, AuthFieldError.passwordTooLong),
];

/// Login password: length only.
AuthFieldError? validateLoginPassword(String password) =>
    _lastFailing(_passwordLength(password));

/// Signup password: length plus lowercase, uppercase and special character.
AuthFieldError? validateSignupPassword(String password) => _lastFailing([
  ..._passwordLength(password),
  (
    !password.contains(RegExp('[a-z]')),
    AuthFieldError.passwordMissingLowercase,
  ),
  (
    !password.contains(RegExp('[A-Z]')),
    AuthFieldError.passwordMissingUppercase,
  ),
  (!_special.hasMatch(password), AuthFieldError.passwordMissingSpecial),
]);

/// Signup confirm-password field.
AuthFieldError? validateConfirmPassword(
  String confirm,
  String password,
) => _lastFailing([
  (confirm.isEmpty, AuthFieldError.confirmPasswordRequired),
  (confirm.length < FieldLimits.passwordMin, AuthFieldError.passwordTooShort),
  (confirm.length > FieldLimits.passwordMax, AuthFieldError.passwordTooLong),
  (confirm != password, AuthFieldError.passwordsDoNotMatch),
]);

/// Signup username.
AuthFieldError? validateUsername(String username) => _lastFailing([
  (username.isEmpty, AuthFieldError.usernameRequired),
  (username.length < FieldLimits.usernameMin, AuthFieldError.usernameTooShort),
  (username.length > FieldLimits.usernameMax, AuthFieldError.usernameTooLong),
]);

/// Account setup: the web page trims and only checks the minimum (the input
/// itself caps the maximum).
AuthFieldError? validateAccountSetupUsername(String username) =>
    username.trim().length < FieldLimits.usernameMin
    ? AuthFieldError.usernameTooShort
    : null;

/// The three fields of the change-password form, each with its error (null =
/// valid).
typedef ChangePasswordErrors = ({
  AuthFieldError? current,
  AuthFieldError? next,
  AuthFieldError? confirm,
});

/// `web/lib/validations/changePassword.ts`: every field is 8..72 long, the
/// new password must differ from the current one (reported on the new
/// field) and match the confirmation (reported on the confirm field). As on
/// web the last failing rule of a field wins.
ChangePasswordErrors validateChangePassword({
  required String current,
  required String next,
  required String confirm,
}) {
  List<(bool, AuthFieldError)> length(String value) => [
    (value.length < FieldLimits.passwordMin, AuthFieldError.passwordTooShort),
    (value.length > FieldLimits.passwordMax, AuthFieldError.passwordTooLong),
  ];
  return (
    current: _lastFailing(length(current)),
    next: _lastFailing([
      ...length(next),
      (next == current, AuthFieldError.newPasswordMustDiffer),
    ]),
    confirm: _lastFailing([
      ...length(confirm),
      (next != confirm, AuthFieldError.passwordsDoNotMatch),
    ]),
  );
}
