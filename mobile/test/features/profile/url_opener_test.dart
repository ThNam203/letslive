import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/profile/profile_providers.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  // Links come from other users (social links, DM pictures): only web links
  // may leave the app, never tel:, sms:, intent: or another app's scheme.
  test(
    'urlOpener refuses every scheme but http(s), without launching',
    () async {
      final container = ProviderContainer();
      addTearDown(container.dispose);
      final open = container.read(urlOpenerProvider);

      for (final url in [
        'tel:+123456',
        'sms:+123456',
        'intent://scan/#Intent;scheme=zxing;end',
        'otherapp://pay?amount=100',
        'file:///etc/passwd',
        '/files/uploads/relative.jpg',
      ]) {
        expect(await open(Uri.parse(url)), isFalse, reason: url);
      }
    },
  );
}
