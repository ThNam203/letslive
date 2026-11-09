import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/config/env.dart';

void main() {
  // String.fromEnvironment only sees keys named in code, so a key documented
  // in .env.example but missing from compileTimeEnv is silently dropped from
  // every build (that once broke staging/prod boot: WEB_BASE_URL).
  test('compileTimeEnv reads every key documented in .env.example', () {
    final keyLine = RegExp(r'^#\s+([A-Z][A-Z0-9_]+)(\s|$)');
    final documented = File('.env.example')
        .readAsLinesSync()
        .map(keyLine.firstMatch)
        .whereType<RegExpMatch>()
        .map((match) => match.group(1)!)
        .toSet();

    expect(documented, contains('WEB_BASE_URL'));
    expect(compileTimeEnv.keys.toSet(), containsAll(documented));
  });
}
