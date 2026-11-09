import 'package:letslive_mobile/features/home/data/livestream.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// What the channel screen (`/users/:userId`) shows once loaded. Loading and
/// load errors are the surrounding `AsyncValue`.
sealed class ChannelState {
  const ChannelState();
}

/// [user] is broadcasting [stream]: the player.
final class ChannelLive extends ChannelState {
  const ChannelLive(this.user, this.stream);

  final PublicUser user;
  final Livestream stream;
}

/// [user] is not broadcasting.
final class ChannelOffline extends ChannelState {
  const ChannelOffline(this.user);

  final PublicUser user;
}

/// [user]'s stream stopped while this screen showed it (Review Focus 5).
final class ChannelEnded extends ChannelState {
  const ChannelEnded(this.user);

  final PublicUser user;
}

/// No such user (unknown or malformed id).
final class ChannelNotFound extends ChannelState {
  const ChannelNotFound();
}
