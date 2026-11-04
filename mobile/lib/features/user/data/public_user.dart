import 'package:letslive_mobile/core/network/json_fields.dart';

/// Another user as seen by everyone, typed after `PublicUser` in
/// `web/types/user.ts`. Only the fields the app shows so far are modelled
/// (the profile screens add more); those are parsed strictly.
class PublicUser {
  const PublicUser({
    required this.id,
    required this.username,
    required this.followerCount,
    this.profilePicture,
  });

  factory PublicUser.fromJson(Map<String, Object?> json) {
    final f = JsonFields('PublicUser', json);
    return PublicUser(
      id: f.string('id'),
      username: f.string('username'),
      followerCount: f.integer('followerCount'),
      profilePicture: f.optString('profilePicture'),
    );
  }

  /// `ParseData<PublicUser>` for `ApiClient`.
  static PublicUser parse(Object? data) =>
      parseJsonObject(data, 'PublicUser', PublicUser.fromJson);

  /// `ParseData<List<PublicUser>>` for `ApiClient`.
  static List<PublicUser> parseList(Object? data) =>
      parseJsonList(data, 'PublicUser', PublicUser.fromJson);

  final String id;

  /// Empty when the user has not finished account setup.
  final String username;
  final int followerCount;

  /// Absolute URL of the avatar; null when none was uploaded.
  final String? profilePicture;
}
