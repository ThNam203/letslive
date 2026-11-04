import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/features/search/data/search_repository.dart';

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

  test('users sends the query as `username` with page and limit', () async {
    adapter.handler = (_) => okEnvelope([publicUserJson(username: 'jane')]);

    final res = await SearchRepository(client)
        .users(q: 'trò chơi', page: 2, limit: 5);

    expect(res.data!.single.username, 'jane');
    final request = adapter.requests.single;
    expect(request.path, '/users/search');
    expect(request.queryParameters, {
      'username': 'trò chơi',
      'page': 2,
      'limit': 5,
    });
  });

  test(
    'lives and vods send the query as `q`, 20 per page by default',
    () async {
      adapter.handler = (o) => okEnvelope(
        o.path == '/livestreams/search' ? [livestreamJson()] : [vodJson()],
      );
      final repository = SearchRepository(client);

      expect((await repository.lives(q: 'a')).data!.single.id, 'ls-1');
      expect((await repository.vods(q: 'a')).data!.single.id, 'vod-1');

      expect(adapter.requests[0].path, '/livestreams/search');
      expect(adapter.requests[0].queryParameters, {
        'q': 'a',
        'page': 0,
        'limit': 20,
      });
      expect(adapter.requests[1].path, '/vods/search');
    },
  );

  test('Vietnamese and reserved characters reach the wire encoded', () async {
    adapter.handler = (_) => okEnvelope(null);

    await SearchRepository(client).vods(q: 'rock & roll Việt');

    expect(
      adapter.requests.single.uri.toString(),
      contains('q=rock+%26+roll+Vi%E1%BB%87t'),
    );
  });

  test('an empty result (null data) is a successful response', () async {
    adapter.handler = (_) => okEnvelope(null);

    final res = await SearchRepository(client).lives(q: 'zzz');

    expect(res.success, isTrue);
  });
}
