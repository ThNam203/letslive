/// The authenticated user returned by `GET /user/me`.
///
/// Typed after `web/types/user.ts` (`MeUser`) and the Go `domains.User` of
/// `backend/user`. Only fields needed so far are modelled; every modelled field
/// is parsed strictly, so a type mismatch throws [FormatException] (which
/// `ApiClient` turns into a `malformedResponse` failure).
class MeUser {
  const MeUser({
    required this.id,
    required this.username,
    required this.email,
    required this.status,
    required this.authProvider,
    required this.createdAt,
    required this.followerCount,
    required this.streamApiKey,
    this.bio,
    this.profilePicture,
    this.backgroundPicture,
    this.phoneNumber,
    this.locale,
  });

  factory MeUser.fromJson(Map<String, Object?> json) => MeUser(
    id: _string(json, 'id'),
    // Null and empty both mean "account setup pending"; one value is simpler
    // for every caller. Any other non-string still throws.
    username: _optString(json, 'username') ?? '',
    email: _string(json, 'email'),
    status: _string(json, 'status'),
    authProvider: _string(json, 'authProvider'),
    createdAt: DateTime.parse(_string(json, 'createdAt')),
    followerCount: _int(json, 'followerCount'),
    streamApiKey: _string(json, 'streamAPIKey'),
    bio: _optString(json, 'bio'),
    profilePicture: _optString(json, 'profilePicture'),
    backgroundPicture: _optString(json, 'backgroundPicture'),
    phoneNumber: _optString(json, 'phoneNumber'),
    locale: _optString(json, 'locale'),
  );

  /// `ParseData<MeUser>` for `ApiClient`.
  static MeUser parse(Object? data) {
    if (data is! Map<String, Object?>) {
      throw FormatException('MeUser: expected a JSON object, got $data');
    }
    return MeUser.fromJson(data);
  }

  final String id;

  /// Empty until the user completes account setup (Google signup); a JSON
  /// null is read as empty too.
  final String username;
  final String email;

  /// `normal` | `disabled`.
  final String status;

  /// `local` | `google`.
  final String authProvider;
  final DateTime createdAt;
  final int followerCount;
  final String streamApiKey;
  final String? bio;
  final String? profilePicture;
  final String? backgroundPicture;
  final String? phoneNumber;

  /// BCP-47 tag; null when the user never chose one.
  final String? locale;
}

String _string(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value is String) return value;
  throw FormatException('MeUser.$key: expected string, got $value');
}

String? _optString(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value == null || value is String) return value as String?;
  throw FormatException('MeUser.$key: expected string or null, got $value');
}

int _int(Map<String, Object?> json, String key) {
  final value = json[key];
  if (value is int) return value;
  throw FormatException('MeUser.$key: expected int, got $value');
}
