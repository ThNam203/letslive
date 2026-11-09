/// When watching a video counts as a view.
///
/// Web registers a view once the player passed 15 seconds or 10% of the
/// video, whichever is less (at least 1 second for a video with a known
/// length), and the backend refuses earlier ones (`res_err_vod_view_threshold`).
/// Mobile applies the same rule so it never sends a request that is refused.
library;

/// Seconds a viewer must have played, for a video of [duration] seconds,
/// before the view counts (web `getViewThreshold`). 0 for an unknown length.
int viewThreshold(int duration) {
  const maxThreshold = 15;
  final tenPercent = (duration * 0.1).floor();
  var threshold = tenPercent < maxThreshold ? tenPercent : maxThreshold;
  if (duration > 0 && threshold < 1) threshold = 1;
  return threshold;
}

/// Whether [playedSeconds] into a video of [duration] seconds is enough to
/// register the view. At least one whole second is always needed: the backend
/// rejects `watchedSeconds` 0.
bool shouldRegisterView({
  required double playedSeconds,
  required int duration,
}) {
  final watched = playedSeconds.floor();
  return watched >= 1 && watched >= viewThreshold(duration);
}
