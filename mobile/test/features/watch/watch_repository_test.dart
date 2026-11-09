import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/features/watch/data/watch_repository.dart';
import 'package:letslive_mobile/features/watch/domain/live_stream_url.dart';

import '../../core/network/fake_adapter.dart';
import '../home/home_fixtures.dart';

void main() {
  late FakeAdapter adapter;
  late ApiClient client;

  setUp(() {
    adapter = FakeAdapter((_) => okEnvelope(null));
    client = ApiClient(
      baseUrl: 'http://localhost:8000',
      cookieJar: CookieJar(),
      adapter: adapter,
    );
  });
  tearDown(() => client.dispose());

  test('liveOf asks for the user and parses the live stream', () async {
    adapter.handler = (_) => okEnvelope(livestreamJson(title: 'On air'));

    final res = await WatchRepository(client).liveOf('u-1');

    expect(res.success, isTrue);
    expect(res.data!.title, 'On air');
    final request = adapter.requests.single;
    expect(request.path, '/livestreams');
    expect(request.queryParameters, {'userId': 'u-1'});
  });

  test('liveOf is a success without data when the user is offline', () async {
    final res = await WatchRepository(client).liveOf('u-1');

    expect(res.success, isTrue);
    expect(res.data, isNull);
  });

  test('liveOf reports a mistyped stream as malformed', () async {
    adapter.handler = (_) =>
        okEnvelope({...livestreamJson(), 'startedAt': 'yesterday'});

    final res = await WatchRepository(client).liveOf('u-1');

    expect(res.success, isFalse);
    expect(res.failure, isNotNull);
  });

  test('liveStreamUrl is the transcoder playlist, like web', () {
    expect(
      liveStreamUrl('https://api.example.com:443', 'ls-1'),
      'https://api.example.com:443/transcode/ls-1/index.m3u8',
    );
  });
}
