import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/data/vod_reaction.dart';

/// Page size of the "other videos" list under a video (web
/// `GetPublicVODsOfUser` default).
const int userVodsPageSize = 10;

/// The video endpoints behind the VOD screen, mirroring
/// `web/lib/api/vod.ts`. Failures are values ([ApiResponse.success] false),
/// never thrown.
class VodRepository {
  const VodRepository(this._client);

  final ApiClient _client;

  /// `GET /vods/<id>`: one video. 404 (`res_err_vod_not_found`) for an
  /// unknown id and for a private video of someone else; 400 for an id that
  /// is not a UUID.
  Future<ApiResponse<Vod>> byId(String vodId) =>
      _client.get('${ApiEndpoints.vods}/$vodId', parse: Vod.parse);

  /// `GET /vods?userId=`: the public videos of [userId].
  Future<ApiResponse<List<Vod>>> ofUser(String userId, {int page = 0}) =>
      _client.get(
        ApiEndpoints.vods,
        query: {'userId': userId, 'page': page, 'limit': userVodsPageSize},
        parse: Vod.parseList,
      );

  /// `POST /vods/<id>/view`: counts one view after [watchedSeconds] of
  /// playback. The backend refuses less than the view threshold.
  Future<ApiResponse<Object?>> registerView(String vodId, int watchedSeconds) =>
      _client.post(
        '${ApiEndpoints.vods}/$vodId/view',
        body: {'watchedSeconds': watchedSeconds},
        parse: (data) => data,
      );

  /// `GET /vods/<id>/reaction`: the signed-in viewer's reaction and the like
  /// count.
  Future<ApiResponse<VodReactionState>> myReaction(String vodId) => _client.get(
    '${ApiEndpoints.vods}/$vodId/reaction',
    parse: VodReactionState.parse,
  );

  /// `PUT /vods/<id>/reaction`: sets the reaction; answers the new state.
  Future<ApiResponse<VodReactionState>> react(
    String vodId,
    VodReaction reaction,
  ) => _client.put(
    '${ApiEndpoints.vods}/$vodId/reaction',
    body: {'reaction': reaction.name},
    parse: VodReactionState.parse,
  );

  /// `DELETE /vods/<id>/reaction`: removes the reaction; answers the new
  /// state.
  Future<ApiResponse<VodReactionState>> removeReaction(String vodId) =>
      _client.delete(
        '${ApiEndpoints.vods}/$vodId/reaction',
        parse: VodReactionState.parse,
      );
}
