import 'package:letslive_mobile/core/network/api_endpoints.dart';

/// The HLS master playlist of live stream [livestreamId], built exactly like
/// web's channel page (`${API_URL}/transcode/<id>/index.m3u8`). The
/// transcoder serves it publicly, so the player needs no cookies.
String liveStreamUrl(String apiBaseUrl, String livestreamId) =>
    '$apiBaseUrl${ApiEndpoints.transcode}/$livestreamId/index.m3u8';
