/// Auth form rules, copied from `web/lib/validations/{login,signUp}.ts` and
/// the account-setup page so mobile accepts and rejects exactly what web
/// does.
///
/// Each validator lists its rules in the web schema's order and reports the
/// LAST failing one: zod runs every check and the web form keeps the last
/// issue per field (e.g. an empty signup password reports "missing special
/// character", as on web).
library;

/// `EMAIL_MAX_LENGTH` in `web/constant/field-limits.ts`.
const int emailMaxLength = 320;

/// `PASSWORD_MIN_LENGTH` / `PASSWORD_MAX_LENGTH` in `web/constant/password.ts`.
const int passwordMinLength = 8;
const int passwordMaxLength = 72;

/// `USERNAME_MIN_LENGTH` / `USERNAME_MAX_LENGTH` in
/// `web/constant/field-limits.ts`.
const int usernameMinLength = 6;
const int usernameMaxLength = 30;

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
  (email.length > emailMaxLength, AuthFieldError.emailTooLong),
]);

List<(bool, AuthFieldError)> _passwordLength(String password) => [
  (password.isEmpty, AuthFieldError.passwordRequired),
  (password.length < passwordMinLength, AuthFieldError.passwordTooShort),
  (password.length > passwordMaxLength, AuthFieldError.passwordTooLong),
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
AuthFieldError? validateConfirmPassword(String confirm, String password) =>
    _lastFailing([
      (confirm.isEmpty, AuthFieldError.confirmPasswordRequired),
      (confirm.length < passwordMinLength, AuthFieldError.passwordTooShort),
      (confirm.length > passwordMaxLength, AuthFieldError.passwordTooLong),
      (confirm != password, AuthFieldError.passwordsDoNotMatch),
    ]);

/// Signup username.
AuthFieldError? validateUsername(String username) => _lastFailing([
  (username.isEmpty, AuthFieldError.usernameRequired),
  (username.length < usernameMinLength, AuthFieldError.usernameTooShort),
  (username.length > usernameMaxLength, AuthFieldError.usernameTooLong),
]);

/// Account setup: the web page trims and only checks the minimum (the input
/// itself caps the maximum).
AuthFieldError? validateAccountSetupUsername(String username) =>
    username.trim().length < usernameMinLength
    ? AuthFieldError.usernameTooShort
    : null;
