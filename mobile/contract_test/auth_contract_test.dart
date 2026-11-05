import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/app_config.dart';
import 'package:letslive_mobile/core/config/env.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/user/data/me_repository.dart';

import 'harness.dart';

const _password = 'Contract-Pass1!';

/// The app's auth repository and `/user/me` repository on a fresh cookie
/// jar, as the dev flavor wires them (Turnstile dev token as bot proof).
class _App {
  _App(AppConfig config, {String? turnstileToken})
    : client = ApiClient(baseUrl: config.apiBaseUrl, cookieJar: CookieJar()) {
    auth = AuthRepository(
      client: client,
      botProof: BotProof.turnstile(
        turnstileToken ?? config.turnstileDevToken ?? '',
      ),
    );
    me = MeRepository(client);
  }

  final ApiClient client;
  late final AuthRepository auth;
  late final MeRepository me;
}

void main() {
  late AppConfig config;
  final runId = DateTime.now().millisecondsSinceEpoch.toRadixString(36);
  final email = 'auth$runId@contract.test';
  final username = 'auth_$runId';

  // Holds the session signup created; only the first test uses it.
  late _App signedUp;

  // The user is created here, not in a test, so every test can run alone or
  // in any order.
  setUpAll(() async {
    config = AppConfig.fromEnv(compileTimeEnv);
    if (config.turnstileDevToken == null) {
      throw const ContractSetupException('TURNSTILE_DEV_TOKEN is not set.');
    }
    signedUp = _App(config);
    // Delivery fails on the dev stack (no SMTP), but the code is stored.
    final verify = await signedUp.auth.requestEmailVerification(email);
    final otp = await readSignupOtp(email, verifyKey: verify.key);
    final signup = await signedUp.auth.signup(
      email: email,
      username: username,
      password: _password,
      otpCode: otp,
    );
    if (!signup.success) {
      throw ContractSetupException(
        'signup failed: ${signup.key} ${signup.message}',
      );
    }
  });

  tearDownAll(() => signedUp.client.dispose());

  test(
    'signup session -> /user/me -> logout -> login -> /user/me -> logout',
    () async {
      final app = signedUp;

      // The user service creates the profile right after signup.
      var me = await app.me.fetch();
      for (var i = 0; i < 30 && !me.success; i++) {
        await Future<void>.delayed(const Duration(milliseconds: 300));
        me = await app.me.fetch();
      }
      expect(me.data?.email, email, reason: '${me.key} ${me.message}');
      expect(me.data?.username, username);

      final logout = await app.auth.logout();
      expect(logout.success, isTrue, reason: '${logout.key} ${logout.message}');
      final afterLogout = await app.me.fetch();
      expect(afterLogout.statusCode, 401);

      final login = await app.auth.login(email, _password);
      expect(login.success, isTrue, reason: '${login.key} ${login.message}');
      final again = await app.me.fetch();
      expect(again.data?.username, username);

      await app.auth.logout();
      expect((await app.me.fetch()).statusCode, 401);
    },
  );

  test('PUT /user/me username update parses as MeUser', () async {
    final ctx = await startContractContext();
    addTearDown(ctx.dispose);
    final me = MeRepository(ctx.userA.client);
    // Harness usernames are ~18 characters; one more stays within 6..30.
    final renamed = 'r${ctx.userA.username}';

    final res = await me.update(username: renamed);

    expect(res.success, isTrue, reason: '${res.key} ${res.message}');
    expect(res.data?.username, renamed);
  });

  test('wrong password is rejected with the web error key', () async {
    final app = _App(config);
    addTearDown(app.client.dispose);

    final res = await app.auth.login(email, 'Wrong-Pass1!');

    expect(res.success, isFalse);
    expect(res.key, 'res_err_email_or_password_incorrect');
  });

  test('login without bot proof is rejected', () async {
    final app = _App(config, turnstileToken: '');
    addTearDown(app.client.dispose);

    final res = await app.auth.login(email, _password);

    expect(res.success, isFalse);
    expect(res.key, 'res_err_captcha_failed');
  });

  test(
    'POST /auth/google/mobile is routed and rejects a bogus token',
    () async {
      final app = _App(config);
      addTearDown(app.client.dispose);

      final res = await app.auth.loginWithGoogle('not-a-google-id-token');

      expect(res.success, isFalse);
      expect(res.statusCode, isNot(404));
      expect(res.key, isNot('res_err_route_not_found'));
    },
  );
}
