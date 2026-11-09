// Guards web/mobile parity so the mobile app cannot silently drift from web.
//
//   dart run tool/parity_check.dart            run from mobile/
//   dart run tool/parity_check.dart --release  also fail while anything is
//                                              still `planned`
//
// Checks (all must pass, exit 0; any failure exits 1 and names each item):
//  1. Every gateway prefix in configs/envoy/envoy.yaml
//     (`path_separated_prefix:` / `path:` inside a route `match:`) is covered.
//  2. Every path passed to fetchClient(...) in web/lib/api/*.ts is covered.
//  3. Every web page (page.tsx under web/app, route groups stripped, dynamic
//     segments kept) has a row in docs/PARITY.md, and every `/route` row
//     still matches a web page.
//  4. Every PARITY row has a valid status: `done` (needs mobile screen and
//     tests), `planned`, or `differs: <reason>`.
//
// "Covered" means: some ApiEndpoints constant (parsed from
// lib/core/network/api_endpoints.dart), or some entry under the PARITY.md
// sections "Planned endpoints" / "Intentionally differs", is a whole-segment
// prefix of the path. Segments `{...}`, `${...}` match any single segment.
// For gateway prefixes a mobile constant that lies *under* the prefix counts
// too (gateway `/auth` is covered by `/auth/login`); web client paths are
// stricter, so a missing sibling like `/auth/password` is still caught.
//
// Options: --envoy, --web-api, --web-app, --mobile-endpoints, --parity
// (each `--name=<path>`; defaults are relative to mobile/).
import 'dart:io';

const _defaultEnvoy = '../configs/envoy/envoy.yaml';
const _defaultWebApi = '../web/lib/api';
const _defaultWebApp = '../web/app';
const _defaultEndpoints = 'lib/core/network/api_endpoints.dart';
const _defaultParity = '../docs/PARITY.md';

const _placeholder = '{}';
const _statusDone = 'done';
const _statusPlanned = 'planned';
const _noValue = '-';

/// Raised for unreadable inputs or web client calls the scanner cannot read.
class ParityException implements Exception {
  ParityException(this.message);
  final String message;

  @override
  String toString() => message;
}

/// One data row of the PARITY.md screens table.
class ParityRow {
  const ParityRow({
    required this.route,
    required this.screen,
    required this.tests,
    required this.status,
  });

  final String route;
  final String screen;
  final String tests;
  final String status;
}

/// A bullet under "Planned endpoints" or "Intentionally differs".
class EndpointEntry {
  const EndpointEntry(this.prefix, this.reason);
  final String prefix;
  final String reason;
}

/// Parsed docs/PARITY.md.
class ParityDoc {
  const ParityDoc({
    required this.rows,
    required this.planned,
    required this.differs,
    required this.problems,
  });

  final List<ParityRow> rows;
  final List<EndpointEntry> planned;
  final List<EndpointEntry> differs;
  final List<String> problems;
}

// ---------------------------------------------------------------- paths

final _braceInterpolation = RegExp(r'\$?\{[^}]*\}');
final _dartInterpolation = RegExp(r'\$[A-Za-z_]\w*');

/// Normalizes a path: interpolations and `{name}` become `{}`, the query
/// string and any trailing slash are dropped.
String normalizePath(String raw) {
  var path = raw.split('?').first;
  path = path.replaceAll(_braceInterpolation, _placeholder);
  path = path.replaceAll(_dartInterpolation, _placeholder);
  while (path.length > 1 && path.endsWith('/')) {
    path = path.substring(0, path.length - 1);
  }
  return path;
}

List<String> _segments(String path) =>
    path.split('/').where((s) => s.isNotEmpty).toList();

/// True when [prefix] is a whole-segment prefix of [path]; a `{}` segment on
/// either side matches any single segment.
bool covers(String prefix, String path) {
  final p = _segments(prefix);
  final s = _segments(path);
  if (p.length > s.length) return false;
  for (var i = 0; i < p.length; i++) {
    if (p[i] != s[i] && p[i] != _placeholder && s[i] != _placeholder) {
      return false;
    }
  }
  return true;
}

// ---------------------------------------------------------------- sources

final _envoyKey = RegExp(
  r'^(\s*)(?:-\s+)?(path_separated_prefix|path):\s*(\S+)',
);
final _envoyMatch = RegExp(r'^(\s*)(?:-\s+)?match:\s*$');

/// Client-facing gateway paths from an envoy.yaml, line-based: only keys that
/// are direct children of a route `match:` block count (so `path:` in access
/// log config is ignored). Result is de-duplicated, in file order.
List<String> extractEnvoyPaths(String yaml) {
  final found = <String>[];
  int? matchChildIndent;
  for (final line in yaml.split('\n')) {
    if (line.trim().isEmpty) continue;
    final indent = line.length - line.trimLeft().length;
    final match = _envoyMatch.firstMatch(line);
    if (match != null) {
      final dash = line.trimLeft().startsWith('-') ? 2 : 0;
      matchChildIndent = match.group(1)!.length + dash + 1;
      continue;
    }
    final inMatch = matchChildIndent;
    if (inMatch == null) continue;
    if (indent < inMatch) {
      matchChildIndent = null;
      continue;
    }
    final key = _envoyKey.firstMatch(line);
    if (key != null && key.group(3)!.startsWith('/')) {
      final path = normalizePath(key.group(3)!);
      if (!found.contains(path)) found.add(path);
    }
  }
  return found;
}

final _fetchCall = RegExp(
  r'fetchClient\s*(?:<[\s\S]*?>)?\s*\(\s*'
  r'(?:([`"'
  "'"
  r'])(/[^`"'
  "'"
  r'\n]*)\1|([A-Za-z_]\w*))',
);

/// Paths passed to `fetchClient(...)`, keyed by file name -> content. A
/// variable argument is resolved through `let|const <name> = "<literal>"` in
/// the same file; an unresolvable one throws [ParityException].
List<String> extractWebClientPaths(Map<String, String> files) {
  final found = <String>{};
  for (final entry in files.entries) {
    for (final call in _fetchCall.allMatches(entry.value)) {
      final literal = call.group(2);
      if (literal != null) {
        found.add(normalizePath(literal));
        continue;
      }
      final name = call.group(3)!;
      final assign = RegExp(
        '(?:let|const)\\s+$name\\s*=\\s*([`"\'])(/[^`"\'\\n]*)\\1',
      ).firstMatch(entry.value);
      if (assign == null) {
        throw ParityException(
          '${entry.key}: fetchClient($name) - cannot resolve the path; '
          'assign a literal starting with "/" to it in the same file',
        );
      }
      found.add(normalizePath(assign.group(2)!));
    }
  }
  return found.toList()..sort();
}

final _endpointConst = RegExp(
  r'''static\s+const\s+String\s+\w+\s*=\s*['"]([^'"]+)['"]\s*;''',
);

/// `static const String name = '/path';` values from api_endpoints.dart.
List<String> extractMobileEndpoints(String dart) => [
  for (final m in _endpointConst.allMatches(dart))
    if (m.group(1)!.startsWith('/')) normalizePath(m.group(1)!),
];

/// Route paths for every page.tsx under [webApp] (route groups stripped,
/// dynamic segments kept). Keys are file paths relative to the app dir.
List<String> extractWebRoutes(Iterable<String> pageFiles) {
  final routes = <String>{};
  for (final file in pageFiles) {
    final segments = _segments(file.replaceAll(r'\', '/'))
      ..removeLast()
      ..removeWhere((s) => s.startsWith('(') && s.endsWith(')'));
    routes.add('/${segments.join('/')}');
  }
  return routes.toList()..sort();
}

// ---------------------------------------------------------------- PARITY.md

final _bullet = RegExp(r'^\s*[-*]\s+');
final _entry = RegExp(r'^\s*[-*]\s+`(/[^`]*)`(?:\s+-\s+(.*\S))?\s*$');
final _differsStatus = RegExp(r'^differs:\s*(\S.*)$');

String _stripTicks(String cell) => cell.replaceAll('`', '').trim();

/// Parses docs/PARITY.md. Format problems go into [ParityDoc.problems].
ParityDoc parseParity(String markdown) {
  final rows = <ParityRow>[];
  final planned = <EndpointEntry>[];
  final differs = <EndpointEntry>[];
  final problems = <String>[];
  var section = '';

  for (final line in markdown.split('\n')) {
    if (line.startsWith('#')) {
      final heading = line.toLowerCase();
      if (heading.contains('planned endpoints')) {
        section = 'planned';
      } else if (heading.contains('intentionally differs')) {
        section = 'differs';
      } else {
        section = '';
      }
      continue;
    }
    if (section.isNotEmpty && _bullet.hasMatch(line)) {
      final m = _entry.firstMatch(line);
      if (m == null) {
        problems.add(
          'PARITY.md: malformed endpoint entry "${line.trim()}"; '
          'use: - `/path` - reason',
        );
        continue;
      }
      final reason = m.group(2);
      if (reason == null) {
        problems.add(
          'PARITY.md: endpoint entry `${m.group(1)}` has no reason; '
          'use: - `/path` - reason',
        );
        continue;
      }
      final entry = EndpointEntry(normalizePath(m.group(1)!), reason);
      (section == 'planned' ? planned : differs).add(entry);
      continue;
    }
    if (section.isEmpty && line.trimLeft().startsWith('|')) {
      final cells = line.trim().split('|').map((c) => c.trim()).toList()
        ..removeAt(0)
        ..removeLast();
      if (cells.length != 4) continue;
      final first = _stripTicks(cells[0]);
      if (first == 'web route/feature' || RegExp(r'^[-: ]+$').hasMatch(first)) {
        continue;
      }
      rows.add(
        ParityRow(
          route: first,
          screen: _stripTicks(cells[1]),
          tests: _stripTicks(cells[2]),
          status: cells[3].trim(),
        ),
      );
    }
  }
  return ParityDoc(
    rows: rows,
    planned: planned,
    differs: differs,
    problems: problems,
  );
}

// ---------------------------------------------------------------- checks

bool _hasValue(String cell) => cell.isNotEmpty && cell != _noValue;

List<String> _checkRows(ParityDoc doc, List<String> webRoutes) {
  final failures = <String>[];
  final rowRoutes = <String>{};
  for (final row in doc.rows) {
    final label = '`${row.route}`';
    if (!rowRoutes.add(row.route) && row.route.startsWith('/')) {
      failures.add('PARITY.md: duplicate row for $label');
    }
    final status = row.status;
    if (status == _statusDone) {
      if (!_hasValue(row.screen) || !_hasValue(row.tests)) {
        failures.add(
          'PARITY.md: $label is done but has no mobile screen or tests',
        );
      }
    } else if (status != _statusPlanned && !_differsStatus.hasMatch(status)) {
      failures.add(
        'PARITY.md: $label has invalid status "$status" '
        '(use done, planned, or "differs: <reason>")',
      );
    }
    if (row.route.startsWith('/') && !webRoutes.contains(row.route)) {
      failures.add(
        'PARITY.md: row $label does not match any web page (stale row?)',
      );
    }
  }
  for (final route in webRoutes) {
    if (!rowRoutes.contains(route)) {
      failures.add('web route `$route` has no row in docs/PARITY.md');
    }
  }
  return failures;
}

List<String> _checkEndpoints({
  required List<String> gateway,
  required List<String> webClient,
  required List<String> mobile,
  required ParityDoc doc,
}) {
  final accounted = [
    for (final e in doc.planned) e.prefix,
    for (final e in doc.differs) e.prefix,
  ];
  bool accountedFor(String path) => accounted.any((e) => covers(e, path));

  final failures = <String>[];
  for (final prefix in gateway) {
    final inMobile = mobile.any((m) => covers(m, prefix) || covers(prefix, m));
    if (!inMobile && !accountedFor(prefix)) {
      failures.add(
        'envoy gateway prefix `$prefix` is in neither ApiEndpoints nor '
        'PARITY.md (Planned endpoints / Intentionally differs)',
      );
    }
  }
  for (final path in webClient) {
    final inMobile = mobile.any((m) => covers(m, path));
    if (!inMobile && !accountedFor(path)) {
      failures.add(
        'web client path `$path` is in neither ApiEndpoints nor '
        'PARITY.md (Planned endpoints / Intentionally differs)',
      );
    }
  }
  return failures;
}

List<String> _checkRelease(ParityDoc doc) => [
  for (final row in doc.rows)
    if (row.status == _statusPlanned)
      '--release: `${row.route}` is still planned',
  for (final entry in doc.planned)
    '--release: planned endpoint `${entry.prefix}` is still listed under '
        'Planned endpoints',
];

// ---------------------------------------------------------------- driver

String _read(String path) {
  final file = File(path);
  if (!file.existsSync()) throw ParityException('input not found: $path');
  return file.readAsStringSync();
}

Map<String, String> _readTree(String dir, bool Function(String) accept) {
  final root = Directory(dir);
  if (!root.existsSync()) throw ParityException('input not found: $dir');
  final files = <String, String>{};
  for (final entity in root.listSync(recursive: true, followLinks: false)) {
    if (entity is! File) continue;
    final relative = entity.path.substring(root.path.length + 1);
    if (accept(relative)) files[relative] = entity.readAsStringSync();
  }
  return files;
}

/// Runs the check; returns the process exit code (0 pass, 1 fail, 2 usage).
int run(List<String> args, {StringSink? out, StringSink? err}) {
  final stdoutSink = out ?? stdout;
  final stderrSink = err ?? stderr;
  final options = <String, String>{};
  var release = false;
  for (final arg in args) {
    if (arg == '--release') {
      release = true;
      continue;
    }
    final m = RegExp(r'^--([a-z-]+)=(.+)$').firstMatch(arg);
    if (m == null) {
      stderrSink.writeln('unknown argument: $arg');
      return 2;
    }
    options[m.group(1)!] = m.group(2)!;
  }

  try {
    final gateway = extractEnvoyPaths(_read(options['envoy'] ?? _defaultEnvoy));
    final webClient = extractWebClientPaths(
      _readTree(options['web-api'] ?? _defaultWebApi, (f) => f.endsWith('.ts')),
    );
    final mobile = extractMobileEndpoints(
      _read(options['mobile-endpoints'] ?? _defaultEndpoints),
    );
    final pages = _readTree(
      options['web-app'] ?? _defaultWebApp,
      (f) => f == 'page.tsx' || f.endsWith('/page.tsx'),
    );
    final webRoutes = extractWebRoutes(pages.keys);
    final doc = parseParity(_read(options['parity'] ?? _defaultParity));

    final failures = [
      ...doc.problems,
      ..._checkRows(doc, webRoutes),
      ..._checkEndpoints(
        gateway: gateway,
        webClient: webClient,
        mobile: mobile,
        doc: doc,
      ),
      if (release) ..._checkRelease(doc),
    ];
    if (failures.isNotEmpty) {
      for (final f in failures) {
        stderrSink.writeln('FAIL: $f');
      }
      stderrSink.writeln('parity check failed: ${failures.length} problem(s)');
      return 1;
    }
    stdoutSink.writeln(
      'parity OK: ${webRoutes.length} web routes, ${gateway.length} gateway '
      'prefixes, ${webClient.length} web client paths, ${mobile.length} '
      'mobile endpoints${release ? ' (release mode)' : ''}',
    );
    return 0;
  } on ParityException catch (e) {
    stderrSink.writeln('FAIL: $e');
    return 1;
  }
}

void main(List<String> args) {
  exitCode = run(args);
}
