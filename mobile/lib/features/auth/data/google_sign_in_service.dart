import 'package:google_sign_in/google_sign_in.dart';

/// Native Google sign-in (google_sign_in 7). Yields the Google id token the
/// app posts to `POST /auth/google/mobile`.
///
/// [serverClientId] is the WEB OAuth client id, so the id token's `aud` is
/// the id the backend already trusts (GOOGLE_OAUTH_CLIENT_ID). [iosClientId]
/// is required on iOS only.
class GoogleSignInService {
  GoogleSignInService({
    required this._serverClientId,
    this._iosClientId,
    GoogleSignIn? signIn,
  }) : _signIn = signIn ?? GoogleSignIn.instance;

  final String _serverClientId;
  final String? _iosClientId;
  final GoogleSignIn _signIn;
  Future<void>? _initialized;

  // The plugin must be initialized exactly once before any other call.
  Future<void> _ensureInitialized() => _initialized ??= _signIn.initialize(
    clientId: _iosClientId,
    serverClientId: _serverClientId,
  );

  /// Shows the Google account sheet. Returns the id token, or null when the
  /// user cancels. Throws [GoogleSignInException] for any other failure.
  Future<String?> obtainIdToken() async {
    await _ensureInitialized();
    final GoogleSignInAccount account;
    try {
      account = await _signIn.authenticate();
    } on GoogleSignInException catch (e) {
      if (e.code == GoogleSignInExceptionCode.canceled) return null;
      rethrow;
    }
    final idToken = account.authentication.idToken;
    if (idToken == null || idToken.isEmpty) {
      throw const GoogleSignInException(
        code: GoogleSignInExceptionCode.unknownError,
        description: 'Google returned no id token',
      );
    }
    return idToken;
  }

  /// Forgets the chosen Google account, so the next sign-in asks again.
  /// Registered as a logout hook.
  Future<void> signOut() async {
    await _ensureInitialized();
    await _signIn.signOut();
  }
}
