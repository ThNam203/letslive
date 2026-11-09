/// Social media links of a user: platform key -> URL, as `socialMediaLinks`
/// in `web/types/user.ts` and `backend/user/dto/social_media_links.go`.
library;

/// The platforms the backend accepts, in the order web lists them.
const List<String> socialPlatforms = [
  'facebook',
  'twitter',
  'instagram',
  'linkedin',
  'github',
  'youtube',
  'website',
  'tiktok',
];

/// Reads the optional `socialMediaLinks` object: only known platforms with a
/// non-empty string URL are kept, in [socialPlatforms] order. A missing value
/// is no links; anything that is not an object throws [FormatException].
Map<String, String> parseSocialLinks(Object? value, String model) {
  if (value == null) return const {};
  if (value is! Map<String, Object?>) {
    throw FormatException('$model.socialMediaLinks: expected object: $value');
  }
  return {
    for (final platform in socialPlatforms)
      if (value[platform] case final String url when url.isNotEmpty)
        platform: url,
  };
}
