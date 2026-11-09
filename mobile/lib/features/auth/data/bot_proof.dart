/// Header and body fields one request carries to pass the bot check.
typedef BotProofFields = ({
  Map<String, String> headers,
  Map<String, Object?> body,
});

/// Proves to the backend that a request comes from the real app
/// (`verifyNotBot` in `backend/auth/handlers/auth.go`, used by
/// `POST /auth/login` and `POST /auth/verify-email`).
///
/// Either a Firebase App Check header or, in dev without Firebase, the
/// Cloudflare Turnstile test token in the body. Only those two endpoints get
/// it, never the global header provider.
class BotProof {
  /// App Check: [limitedUseToken] must return a NEW token on every call; the
  /// backend accepts each token once, so retries need a fresh one too.
  const BotProof.appCheck(Future<String> Function() limitedUseToken)
    : _limitedUseToken = limitedUseToken,
      _turnstileToken = null;

  /// Dev only: the dev stack's Turnstile test secret accepts any token.
  const BotProof.turnstile(String token)
    : _turnstileToken = token,
      _limitedUseToken = null;

  /// Header the backend reads the App Check token from.
  static const String appCheckHeader = 'X-Firebase-AppCheck';

  final Future<String> Function()? _limitedUseToken;
  final String? _turnstileToken;

  /// Fields for one request. May throw if App Check cannot mint a token.
  Future<BotProofFields> next() async {
    final limitedUseToken = _limitedUseToken;
    if (limitedUseToken != null) {
      return (
        headers: {appCheckHeader: await limitedUseToken()},
        body: const <String, Object?>{},
      );
    }
    return (
      headers: const <String, String>{},
      body: <String, Object?>{'turnstileToken': _turnstileToken},
    );
  }
}
