import 'package:letslive_mobile/core/network/json_fields.dart';

/// A live stream, typed after `web/types/livestream.ts` and the Go
/// `domains.Livestream` of `backend/livestream`. Parsed strictly; only the
/// fields the backend really sends as null are optional.
class Livestream {
  const Livestream({
    required this.id,
    required this.userId,
    required this.title,
    required this.viewCount,
    required this.visibility,
    required this.startedAt,
    required this.createdAt,
    required this.updatedAt,
    this.description,
    this.thumbnailUrl,
    this.endedAt,
    this.vodId,
  });

  factory Livestream.fromJson(Map<String, Object?> json) {
    final f = JsonFields('Livestream', json);
    return Livestream(
      id: f.string('id'),
      userId: f.string('userId'),
      title: f.string('title'),
      description: f.optString('description'),
      thumbnailUrl: f.optString('thumbnailUrl'),
      viewCount: f.integer('viewCount'),
      visibility: f.string('visibility'),
      startedAt: f.date('startedAt'),
      endedAt: f.optDate('endedAt'),
      createdAt: f.date('createdAt'),
      updatedAt: f.date('updatedAt'),
      vodId: f.optString('vodId'),
    );
  }

  /// `ParseData<Livestream>` for `ApiClient`.
  static Livestream parse(Object? data) =>
      parseJsonObject(data, 'Livestream', Livestream.fromJson);

  /// `ParseData<List<Livestream>>` for `ApiClient`.
  static List<Livestream> parseList(Object? data) =>
      parseJsonList(data, 'Livestream', Livestream.fromJson);

  final String id;
  final String userId;
  final String title;
  final String? description;

  /// Null until a thumbnail exists; see `mediaThumbnailUrl`.
  final String? thumbnailUrl;
  final int viewCount;

  /// `public` | `private`.
  final String visibility;
  final DateTime startedAt;

  /// Null while the stream is live.
  final DateTime? endedAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  /// The recording, once the stream ended.
  final String? vodId;
}
