import 'package:letslive_mobile/features/user/data/me_user.dart';

/// A signed-in user, as loaded from `GET /user/me`.
class Session {
  const Session(this.user);

  final MeUser user;

  /// Google signups have no username yet (the backend sends an empty
  /// string); they must finish account setup before using the app.
  bool get needsAccountSetup => user.username.isEmpty;
}
