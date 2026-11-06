/// The phone number check of the security settings (web `phone-number.tsx`).
library;

final RegExp _separators = RegExp(r'[\s\-.()]');

/// `+`, a country code that does not start with 0, then the rest: 7 to 15
/// digits in all (E.164).
final RegExp _international = RegExp(r'^\+[1-9]\d{6,14}$');

/// Whether [value] may be saved as the phone number. Empty is allowed (it
/// removes the number), as on web. Otherwise the number must be written in
/// international form, digits separated only by spaces, dashes, dots or
/// parentheses.
///
/// Web asks libphonenumber-js, which also knows each country's number
/// lengths; without a phone library (no new dependency) this checks the
/// E.164 shape only, so a few numbers web refuses pass here. The server
/// stores the text as typed, up to 20 characters.
bool isValidPhoneNumber(String value) {
  final trimmed = value.trim();
  if (trimmed.isEmpty) return true;
  return _international.hasMatch(trimmed.replaceAll(_separators, ''));
}
