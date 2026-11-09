import 'package:dio/dio.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/network/json_fields.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

/// What the stream settings screen shows: the stream key and the title,
/// description and thumbnail used for the next livestream (web
/// `MeUser.streamAPIKey` and `MeUser.livestreamInformation`).
///
/// The stream key is a secret: it is never logged and never put in an error
/// message, so [toString] is deliberately not overridden to include it.
class StreamSettings {
  const StreamSettings({
    required this.streamKey,
    required this.title,
    required this.description,
    this.thumbnailUrl,
  });

  /// Reads the streaming part of a `/user/me` body.
  factory StreamSettings.fromMe(Map<String, Object?> json) {
    final info = json['livestreamInformation'];
    if (info is! Map<String, Object?>) {
      throw FormatException('livestreamInformation: expected an object');
    }
    final f = JsonFields('LivestreamInformation', info);
    return StreamSettings(
      streamKey: JsonFields('MeUser', json).string('streamAPIKey'),
      title: f.optString('title') ?? '',
      description: f.optString('description') ?? '',
      thumbnailUrl: _emptyAsNull(f.optString('thumbnailUrl')),
    );
  }

  /// `ParseData<StreamSettings>` for `GET /user/me`.
  static StreamSettings parse(Object? data) {
    if (data is! Map<String, Object?>) {
      throw FormatException('MeUser: expected a JSON object');
    }
    return StreamSettings.fromMe(data);
  }

  /// The key for the streaming software (OBS and similar). Secret.
  final String streamKey;
  final String title;
  final String description;

  /// Absolute URL of the saved thumbnail; null when there is none.
  final String? thumbnailUrl;

  /// This with a new [streamKey] (after a reset).
  StreamSettings withKey(String key) => StreamSettings(
    streamKey: key,
    title: title,
    description: description,
    thumbnailUrl: thumbnailUrl,
  );
}

String? _emptyAsNull(String? value) =>
    value == null || value.isEmpty ? null : value;

/// The stream key and stream information calls (web `web/lib/api/user.ts`:
/// `RequestToGenerateNewAPIKey`, `UpdateLivestreamInformation`). Failures are
/// values, never thrown.
class StreamSettingsRepository {
  const StreamSettingsRepository(this._client);

  final ApiClient _client;

  /// `GET /user/me`, read for the stream parts only.
  Future<ApiResponse<StreamSettings>> fetch() =>
      _client.get(ApiEndpoints.userMe, parse: StreamSettings.parse);

  /// `PATCH /user/me/api-key`: replaces the key; the old one stops working at
  /// once. Answers the new key.
  Future<ApiResponse<String>> resetKey() => _client.patch(
    ApiEndpoints.streamKey,
    parse: (data) => data is String && data.isNotEmpty
        ? data
        : throw const FormatException('stream key: expected a string'),
  );

  /// `PATCH /user/me/livestream-information`, like web: a new [thumbnail]
  /// file wins, else [keepThumbnailUrl] is kept; with neither the server
  /// clears the thumbnail. Answers the saved information.
  Future<ApiResponse<StreamInfo>> update({
    required String title,
    required String description,
    PickedImage? thumbnail,
    String? keepThumbnailUrl,
    void Function(int sent, int total)? onProgress,
  }) => _client.upload(
    ApiEndpoints.livestreamInformation,
    method: 'PATCH',
    formData: FormData.fromMap({
      'title': title,
      'description': description,
      'thumbnail': ?thumbnail?.toMultipart(),
      if (thumbnail == null && keepThumbnailUrl != null)
        'thumbnailUrl': keepThumbnailUrl,
    }),
    onProgress: onProgress,
    parse: StreamInfo.parse,
  );
}

/// The `LivestreamInformation` the update answers.
class StreamInfo {
  const StreamInfo({
    required this.title,
    required this.description,
    this.thumbnailUrl,
  });

  /// `ParseData<StreamInfo>`.
  static StreamInfo parse(Object? data) {
    if (data is! Map<String, Object?>) {
      throw FormatException('LivestreamInformation: expected an object');
    }
    final f = JsonFields('LivestreamInformation', data);
    return StreamInfo(
      title: f.optString('title') ?? '',
      description: f.optString('description') ?? '',
      thumbnailUrl: _emptyAsNull(f.optString('thumbnailUrl')),
    );
  }

  final String title;
  final String description;
  final String? thumbnailUrl;
}
