import 'package:letslive_mobile/core/network/json_fields.dart';

/// A video on demand, typed after `web/types/vod.ts` and the Go `domains.VOD`
/// of `backend/vod`. Parsed strictly; only the fields the backend really sends
/// as null are optional.
class Vod {
  const Vod({
    required this.id,
    required this.userId,
    required this.title,
    required this.visibility,
    required this.viewCount,
    required this.likeCount,
    required this.duration,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
    this.livestreamId,
    this.description,
    this.thumbnailUrl,
    this.playbackUrl,
    this.originalFileUrl,
  });

  factory Vod.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Vod', json);
    return Vod(
      id: f.string('id'),
      livestreamId: f.optString('livestreamId'),
      userId: f.string('userId'),
      title: f.string('title'),
      description: f.optString('description'),
      thumbnailUrl: f.optString('thumbnailUrl'),
      visibility: f.string('visibility'),
      viewCount: f.integer('viewCount'),
      likeCount: f.integer('likeCount'),
      duration: f.integer('duration'),
      playbackUrl: f.optString('playbackUrl'),
      status: f.string('status'),
      originalFileUrl: f.optString('originalFileUrl'),
      createdAt: f.date('createdAt'),
      updatedAt: f.date('updatedAt'),
    );
  }

  /// `ParseData<Vod>` for `ApiClient`.
  static Vod parse(Object? data) => parseJsonObject(data, 'Vod', Vod.fromJson);

  /// `ParseData<List<Vod>>` for `ApiClient`.
  static List<Vod> parseList(Object? data) =>
      parseJsonList(data, 'Vod', Vod.fromJson);

  final String id;

  /// The stream this video was recorded from; null for uploads.
  final String? livestreamId;
  final String userId;
  final String title;
  final String? description;

  /// Null until a thumbnail exists; see `mediaThumbnailUrl`.
  final String? thumbnailUrl;

  /// `public` | `private`.
  final String visibility;
  final int viewCount;
  final int likeCount;

  /// Length in seconds.
  final int duration;
  final String? playbackUrl;

  /// `uploading` | `processing` | `ready` | `failed`.
  final String status;
  final String? originalFileUrl;
  final DateTime createdAt;
  final DateTime updatedAt;
}
