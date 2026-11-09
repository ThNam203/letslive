import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

import '../../tool/parity_check.dart';

const _envoy = '''
static_resources:
  access_log:
  - typed_config:
      path: /dev/stdout
  listeners:
  - filter_chains:
    - filters:
      - typed_config:
          route_config:
            virtual_hosts:
            - routes:
              - match:
                  path_separated_prefix: /auth
              - match:
                  path: /user/me
                  headers:
                  - name: :method
              - match:
                  path_separated_prefix: /vods
              - match:
                  path_separated_prefix: /vods
''';

const _envoyExtra = '''
              - match:
                  path_separated_prefix: ''';

const _webApi = r'''
export async function A() {
  return fetchClient<ApiResponse<void>>("/auth/login", { method: "POST" });
}
export async function B(id: string) {
  return fetchClient<
    ApiResponse<VOD>
  >(
    `/vods/${id}/comments?page=${page}`,
  );
}
export async function C(id: string) {
  let url = `/user/me/following?limit=${limit}`;
  return fetchClient<ApiResponse<void>>(url);
}
''';

const _endpoints = '''
abstract final class ApiEndpoints {
  static const String login = '/auth/login';
  static const String me = '/user/me';
  static const String vods = '/vods';
}
''';

const _parity = '''
# Parity

| web route/feature | mobile screen | tests | status |
|---|---|---|---|
| `/` | HomeScreen | home_test.dart | done |
| `/login` | LoginScreen | login_test.dart | done |
| `/users/[userId]` | - | - | planned |
| `/wallet/deposit` | - | - | differs: out of scope |

## Planned endpoints

## Intentionally differs

''';

void main() {
  late Directory tmp;
  late String envoyPath;
  late String webApiDir;
  late String webAppDir;
  late String endpointsPath;
  late String parityPath;

  void write(String path, String content) {
    File(path)
      ..createSync(recursive: true)
      ..writeAsStringSync(content);
  }

  void page(String relative) =>
      write('$webAppDir/$relative/page.tsx', 'export default function P() {}');

  void setParity(String content) => write(parityPath, content);

  void addGatewayPrefix(String prefix) =>
      write(envoyPath, '$_envoy$_envoyExtra$prefix\n');

  /// Runs the checker; returns exit code and combined output.
  ({int code, String output}) check([List<String> extra = const []]) {
    final sink = StringBuffer();
    final code = run(
      [
        '--envoy=$envoyPath',
        '--web-api=$webApiDir',
        '--web-app=$webAppDir',
        '--mobile-endpoints=$endpointsPath',
        '--parity=$parityPath',
        ...extra,
      ],
      out: sink,
      err: sink,
    );
    return (code: code, output: sink.toString());
  }

  setUp(() {
    tmp = Directory.systemTemp.createTempSync('parity_check_test');
    envoyPath = '${tmp.path}/envoy.yaml';
    webApiDir = '${tmp.path}/web/lib/api';
    webAppDir = '${tmp.path}/web/app';
    endpointsPath = '${tmp.path}/api_endpoints.dart';
    parityPath = '${tmp.path}/PARITY.md';
    write(envoyPath, _envoy);
    write('$webApiDir/x.ts', _webApi);
    write(endpointsPath, _endpoints);
    page('(main)');
    page('(auth)/login');
    page('(main)/users/[userId]');
    page('(main)/wallet/deposit');
    setParity(_parity);
  });

  tearDown(() => tmp.deleteSync(recursive: true));

  test('passes when everything is accounted for', () {
    final result = check();
    expect(result.output, contains('OK'));
    expect(result.code, 0);
  });

  test('flags an envoy prefix missing from mobile and PARITY', () {
    addGatewayPrefix('/wallet');
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/wallet'));
    expect(result.output, contains('envoy'));
  });

  test('accepts a prefix listed as intentionally differs', () {
    addGatewayPrefix('/payments');
    setParity(
      _parity.replaceFirst(
        '## Intentionally differs\n',
        '## Intentionally differs\n\n- `/payments` - Stripe is out of scope\n',
      ),
    );
    expect(check().code, 0);
  });

  test('accepts a prefix listed under Planned endpoints', () {
    addGatewayPrefix('/wallet');
    setParity(
      _parity.replaceFirst(
        '## Planned endpoints\n',
        '## Planned endpoints\n\n- `/wallet` - Task 15\n',
      ),
    );
    expect(check().code, 0);
  });

  test('rejects an endpoint entry without a reason', () {
    setParity(
      _parity.replaceFirst(
        '## Intentionally differs\n',
        '## Intentionally differs\n\n- `/payments`\n',
      ),
    );
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/payments'));
    expect(result.output, contains('reason'));
  });

  test('flags a web page with no PARITY row', () {
    page('(main)/shop');
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/shop'));
  });

  test('covers pages under (auth) and keeps dynamic segments', () {
    page('(auth)/signup');
    page('(main)/users/[userId]/vods/[vodId]');
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/signup'));
    expect(result.output, contains('/users/[userId]/vods/[vodId]'));
  });

  test('flags a PARITY route row that no longer exists on web', () {
    setParity(
      _parity.replaceFirst(
        '| `/login` |',
        '| `/gone` | - | - | planned |\n| `/login` |',
      ),
    );
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/gone'));
  });

  test('flags a web client path missing from mobile and PARITY', () {
    write(
      '$webApiDir/y.ts',
      'fetchClient<ApiResponse<void>>(`/auth/password`, { method: "PUT" });',
    );
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/auth/password'));
    expect(result.output, contains('web client'));
  });

  test('normalizes template interpolations and drops query strings', () {
    final paths = extractWebClientPaths({'x.ts': _webApi});
    expect(
      paths,
      containsAll(['/auth/login', '/vods/{}/comments', '/user/me/following']),
    );
  });

  test('matches on whole path segments only', () {
    write(endpointsPath, _endpoints.replaceFirst("'/vods'", "'/vod'"));
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('/vods'));
  });

  test('a mobile endpoint with a parameter placeholder matches', () {
    write(
      endpointsPath,
      _endpoints.replaceFirst("'/vods'", "'/vods/{id}/comments'"),
    );
    final result = check();
    expect(result.output, isNot(contains('/vods/{}/comments')));
  });

  test('ignores path: keys outside a route match block', () {
    expect(extractEnvoyPaths(_envoy), isNot(contains('/dev/stdout')));
    expect(extractEnvoyPaths(_envoy), equals(['/auth', '/user/me', '/vods']));
  });

  test('flags an unknown status and a differs row without a reason', () {
    setParity(
      _parity
          .replaceFirst('| done |\n| `/login`', '| almost |\n| `/login`')
          .replaceFirst('differs: out of scope', 'differs'),
    );
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('almost'));
    expect(result.output, contains('/wallet/deposit'));
  });

  test('flags a done row without a mobile screen or tests', () {
    setParity(_parity.replaceFirst('HomeScreen | home_test.dart', '- | -'));
    final result = check();
    expect(result.code, 1);
    expect(result.output, contains('`/`'));
  });

  test('--release fails while anything is still planned', () {
    final result = check(['--release']);
    expect(result.code, 1);
    expect(result.output, contains('/users/[userId]'));
    expect(result.output, contains('planned'));
  });

  test('--release fails on leftover Planned endpoints entries', () {
    setParity(
      _parity
          .replaceFirst(
            '| - | - | planned |',
            '| UserScreen | u_test.dart | done |',
          )
          .replaceFirst(
            '## Planned endpoints\n',
            '## Planned endpoints\n\n- `/wallet` - Task 15\n',
          ),
    );
    final result = check(['--release']);
    expect(result.code, 1);
    expect(result.output, contains('/wallet'));
  });

  test('--release passes when nothing is planned', () {
    setParity(
      _parity.replaceFirst(
        '| - | - | planned |',
        '| UserScreen | u_test.dart | done |',
      ),
    );
    expect(check(['--release']).code, 0);
  });
}
