import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/features/vod/domain/view_threshold.dart';

/// Web's `getViewThreshold` (`vod-view.tsx`) and the backend's
/// `RegisterView` agree: 15 s or 10% of the video, whichever is less.
void main() {
  test('long videos count after 15 seconds', () {
    expect(viewThreshold(3600), 15);
    expect(viewThreshold(150), 15);
  });

  test('shorter videos count after 10 percent, at least 1 second', () {
    expect(viewThreshold(100), 10);
    expect(viewThreshold(5), 1);
  });

  test('an unknown length (0) needs no watch time', () {
    expect(viewThreshold(0), 0);
  });

  test('shouldRegisterView needs the threshold and at least one second', () {
    expect(shouldRegisterView(playedSeconds: 14.9, duration: 3600), isFalse);
    expect(shouldRegisterView(playedSeconds: 15, duration: 3600), isTrue);
    // watchedSeconds 0 is refused by the backend (`required`).
    expect(shouldRegisterView(playedSeconds: 0.5, duration: 0), isFalse);
    expect(shouldRegisterView(playedSeconds: 1, duration: 0), isTrue);
  });
}
