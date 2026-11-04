import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// cookie_jar [Storage] backed by Keychain/Keystore, so the ACCESS_TOKEN and
/// REFRESH_TOKEN cookies never touch plain files.
class SecureCookieStorage implements Storage {
  SecureCookieStorage({FlutterSecureStorage? storage})
    : _storage = storage ?? const FlutterSecureStorage();

  static const _prefix = 'letslive.cookies.';

  final FlutterSecureStorage _storage;

  @override
  Future<void> init(bool persistSession, bool ignoreExpires) async {}

  @override
  Future<String?> read(String key) => _storage.read(key: '$_prefix$key');

  @override
  Future<void> write(String key, String value) =>
      _storage.write(key: '$_prefix$key', value: value);

  @override
  Future<void> delete(String key) => _storage.delete(key: '$_prefix$key');

  @override
  Future<void> deleteAll(List<String> keys) async {
    for (final key in keys) {
      await delete(key);
    }
  }
}

/// The app's persistent cookie jar (secure-storage backed).
PersistCookieJar createSecureCookieJar({Storage? storage}) =>
    PersistCookieJar(storage: storage ?? SecureCookieStorage());
