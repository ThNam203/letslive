import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import 'fake_player.dart';

void main() {
  late FakeScreenWake wake;
  late ScreenWakeHolder holder;

  setUp(() {
    wake = FakeScreenWake();
    holder = ScreenWakeHolder(wake.call);
  });

  test('the first owner turns the screen on, the last one turns it off', () {
    final player = Object();
    final upload = Object();

    holder.acquire(player);
    holder.acquire(upload);
    expect(wake.calls, [true]);

    holder.release(player);
    expect(wake.calls, [true], reason: 'the upload still needs it');

    holder.release(upload);
    expect(wake.calls, [true, false]);
  });

  test('acquiring twice or releasing without holding changes nothing', () {
    final owner = Object();

    holder.release(owner);
    holder.acquire(owner);
    holder.acquire(owner);
    expect(wake.calls, [true]);

    holder.release(owner);
    holder.release(owner);
    expect(wake.calls, [true, false]);
  });
}
