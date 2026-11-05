import 'package:letslive_mobile/core/network/json_fields.dart';
import 'package:letslive_mobile/features/user/data/social_links.dart';

/// Another user as seen by everyone, typed after `PublicUser` in
/// `web/types/user.ts`. Only the fields the app shows are modelled; those
/// are parsed strictly.
class PublicUser {
  const PublicUser({
    required this.id,
    required this.username,
    required this.followerCount,
    required this.createdAt,
    this.profilePicture,
    this.backgroundPicture,
    this.bio,
    this.isFollowing,
    this.socialMediaLinks = const {},
  });

  factory PublicUser.fromJson(Map<String, Object?> json) {
    final f = JsonFields('PublicUser', json);
    return PublicUser(
      id: f.string('id'),
      username: f.string('username'),
      followerCount: f.integer('followerCount'),
      createdAt: f.date('createdAt'),
      profilePicture: f.optString('profilePicture'),
      backgroundPicture: f.optString('backgroundPicture'),
      bio: f.optString('bio'),
      isFollowing: json['isFollowing'] as bool?,
      socialMediaLinks: parseSocialLinks(
        json['socialMediaLinks'],
        'PublicUser',
      ),
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

  /// Absolute URL of the profile banner; null when none was uploaded.
  final String? backgroundPicture;

  /// Short self-description; null when the user wrote none.
  final String? bio;

  /// When the account was created ("Joined ...").
  final DateTime createdAt;

  /// Whether the signed-in viewer follows this user; null for a guest.
  final bool? isFollowing;

  /// Platform -> URL, see [parseSocialLinks].
  final Map<String, String> socialMediaLinks;

  /// This user as followed ([following] true) or not, with the follower
  /// count moved by one to match (never below zero).
  PublicUser withFollowing({required bool following}) => PublicUser(
    id: id,
    username: username,
    followerCount: (followerCount + (following ? 1 : -1)).clamp(0, 1 << 30),
    createdAt: createdAt,
    profilePicture: profilePicture,
    backgroundPicture: backgroundPicture,
    bio: bio,
    isFollowing: following,
    socialMediaLinks: socialMediaLinks,
  );
}
