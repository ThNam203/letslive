import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/features/home/data/home_repository.dart';
import 'package:letslive_mobile/features/user/data/user_repository.dart';

import '../../core/network/fake_adapter.dart';
import 'home_fixtures.dart';

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

  test('popularLive requests the first page and parses livestreams', () async {
    adapter.handler = (_) => okEnvelope([livestreamJson(title: 'Live 1')]);

    final res = await HomeRepository(client).popularLive();

    expect(res.success, isTrue);
    expect(res.data!.single.title, 'Live 1');
    expect(res.data!.single.endedAt, isNull);
    final request = adapter.requests.single;
    expect(request.path, '/popular-livestreams');
    expect(request.queryParameters, {'page': 0, 'limit': 10});
  });

  test('popularVods requests 20 per page and parses videos', () async {
    adapter.handler = (_) => okEnvelope([vodJson(title: 'Vod 1')]);

    final res = await HomeRepository(client).popularVods();

    expect(res.data!.single.duration, 3725);
    expect(adapter.requests.single.path, '/popular-vods');
    expect(adapter.requests.single.queryParameters, {'page': 0, 'limit': 20});
  });

  test('a mistyped field is a malformed response, not wrong data', () async {
    adapter.handler = (_) => okEnvelope([
      {...livestreamJson(), 'viewCount': 'many'},
    ]);

    final res = await HomeRepository(client).popularLive();

    expect(res.success, isFalse);
    expect(res.failure, isNotNull);
  });

  test('user repository reads channels and one user', () async {
    adapter.handler = (o) => switch (o.path) {
      '/user/me/following' || '/users/recommendations' => okEnvelope([
        publicUserJson(id: 'u-9', username: 'nine'),
      ]),
      _ => okEnvelope(publicUserJson(id: 'u-1', username: 'one')),
    };
    final users = UserRepository(client);

    expect((await users.following()).data!.single.username, 'nine');
    expect((await users.recommended()).data!.single.id, 'u-9');
    expect((await users.byId('u-1')).data!.username, 'one');
    expect(adapter.requests.last.path, '/user/u-1');
    expect(adapter.requests[1].queryParameters, {'page': 0});
  });
}
