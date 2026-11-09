// Converts the web app's i18next locale JSON into Flutter ARB files so the
// mobile app speaks exactly the same strings as the web app.
//
//   dart run tool/sync_locales.dart            write l10n/app_{en,vi}.arb
//   dart run tool/sync_locales.dart --check    exit 1 if the ARB files differ
//
// Options: --web-dir=<path> (default ../web/lib/i18n/locales),
//          --out-dir=<path> (default l10n).
//
// Key scheme: `<namespace>_<nested_key_path>`, every character outside
// [A-Za-z0-9_] becomes `_`. i18next plural families (`x_one`, `x_other`, ...)
// collapse into one ICU plural message under `x`, driven by `count`.
import 'dart:convert';
import 'dart:io';

const defaultWebDir = '../web/lib/i18n/locales';
const defaultOutDir = 'l10n';

/// Web locale folder -> ARB locale code.
const localeFolders = {'en-US': 'en', 'vi-VN': 'vi'};
const templateLocale = 'en';

const _pluralForms = ['zero', 'one', 'two', 'few', 'many', 'other'];
final _pluralSuffix = RegExp('^(.+)_(${_pluralForms.join('|')})\$');
final _placeholderRe = RegExp(r'\{\{\s*([^{}]*?)\s*\}\}');
final _identifierRe = RegExp(r'^[A-Za-z_][A-Za-z0-9_]*$');
final _invalidKeyChars = RegExp(r'[^A-Za-z0-9_]');

class SyncException implements Exception {
  SyncException(this.message);
  final String message;
  @override
  String toString() => message;
}

/// One converted message: its ICU text and the placeholder names it uses.
class ArbMessage {
  const ArbMessage(this.text, this.placeholders, {this.pluralVar});
  final String text;
  final Set<String> placeholders;
  final String? pluralVar;
}

/// Escapes ICU syntax characters and converts `{{name}}` to `{name}`.
/// Returns the converted text and the placeholders it used.
(String, Set<String>) convertString(String source, String where) {
  final placeholders = <String>{};
  final buffer = StringBuffer();
  var cursor = 0;
  for (final match in _placeholderRe.allMatches(source)) {
    buffer.write(_escapeLiteral(source.substring(cursor, match.start)));
    final name = match.group(1) ?? '';
    if (!_identifierRe.hasMatch(name)) {
      throw SyncException(
        'Unsupported interpolation "${match.group(0)}" in $where '
        '(only plain {{name}} is supported)',
      );
    }
    placeholders.add(name);
    buffer.write('{$name}');
    cursor = match.end;
  }
  buffer.write(_escapeLiteral(source.substring(cursor)));
  final rest = source.replaceAll(_placeholderRe, '');
  if (rest.contains(r'$t(') || rest.contains('{{') || rest.contains('}}')) {
    throw SyncException('Unsupported i18next syntax in $where: "$source"');
  }
  return (buffer.toString(), placeholders);
}

String _escapeLiteral(String text) =>
    text.replaceAll("'", "''").replaceAll('{', "'{'").replaceAll('}', "'}'");

String sanitizeKey(String raw) => raw.replaceAll(_invalidKeyChars, '_');

void _flatten(
  Object? node,
  List<String> path,
  String where,
  Map<String, String> out,
) {
  if (node is Map<String, Object?>) {
    for (final entry in node.entries) {
      _flatten(entry.value, [...path, entry.key], where, out);
    }
  } else if (node is String) {
    out[path.join('_')] = node;
  } else {
    throw SyncException('$where/${path.join('.')} is not a string or object');
  }
}

/// Converts `namespace -> parsed JSON` of one locale into ARB messages.
/// Throws [SyncException] on key collisions or unsupported syntax.
Map<String, ArbMessage> convertLocale(Map<String, Object?> namespaces) {
  final result = <String, ArbMessage>{};
  final origins = <String, String>{};
  final collisions = <String>[];
  final shadowed = <String>[];

  void add(String key, ArbMessage message, String origin) {
    final previous = origins[key];
    if (previous != null) {
      collisions.add('$key  <-  "$previous" and "$origin"');
      return;
    }
    origins[key] = origin;
    result[key] = message;
  }

  for (final ns in (namespaces.keys.toList()..sort())) {
    final flat = <String, String>{};
    _flatten(namespaces[ns], [], ns, flat);
    final nsPrefix = sanitizeKey(ns);
    final families = <String, Map<String, String>>{};
    for (final entry in flat.entries) {
      final m = _pluralSuffix.firstMatch(entry.key);
      if (m != null) {
        families.putIfAbsent(m.group(1)!, () => {})[m.group(2)!] = entry.value;
      }
    }
    // A suffix only means "plural form" when the family has `_other`; keys
    // like typing_one / typing_two are ordinary keys.
    families.removeWhere((_, forms) => !forms.containsKey('other'));
    final pluralKeys = {
      for (final fam in families.entries)
        for (final form in fam.value.keys) '${fam.key}_$form',
    };

    for (final entry in flat.entries) {
      if (pluralKeys.contains(entry.key)) continue;
      // i18next resolves `t('x', {count})` to the plural form, so a plain
      // `x` next to an `x_other` family is redundant, but only safe to drop
      // when it says the same thing as `x_other`.
      if (families.containsKey(entry.key)) {
        if (entry.value != flat['${entry.key}_other']) {
          shadowed.add(
            '$ns:${entry.key} differs from ${entry.key}_other '
            '("${entry.value}" vs "${flat['${entry.key}_other']}")',
          );
        }
        continue;
      }
      final where = '$ns:${entry.key}';
      final (text, placeholders) = convertString(entry.value, where);
      add(
        sanitizeKey('${nsPrefix}_${entry.key}'),
        ArbMessage(text, placeholders),
        where,
      );
    }
    for (final fam in families.entries) {
      final where = '$ns:${fam.key}';
      final placeholders = <String>{};
      final forms = StringBuffer();
      for (final form in _pluralForms.where(fam.value.containsKey)) {
        final (text, used) = convertString(fam.value[form]!, '$where _$form');
        placeholders.addAll(used);
        forms.write(' $form{$text}');
      }
      placeholders.add('count');
      add(
        sanitizeKey('${nsPrefix}_${fam.key}'),
        ArbMessage('{count, plural,$forms}', placeholders, pluralVar: 'count'),
        where,
      );
    }
  }
  if (shadowed.isNotEmpty) {
    throw SyncException(
      'Plain key would be silently dropped by a plural family:\n'
      '  ${shadowed.join('\n  ')}',
    );
  }
  if (collisions.isNotEmpty) {
    throw SyncException(
      'ARB key collisions after sanitizing:\n  ${collisions.join('\n  ')}',
    );
  }
  return result;
}

/// Renders messages as ARB JSON. The template locale also carries
/// `@key` metadata declaring placeholders, as gen-l10n requires.
String renderArb(
  String locale,
  Map<String, ArbMessage> messages, {
  required bool withMetadata,
  Map<String, Set<String>> extraPlaceholders = const {},
}) {
  final keys = messages.keys.toList()..sort();
  final doc = <String, Object?>{'@@locale': locale};
  for (final key in keys) {
    final message = messages[key]!;
    doc[key] = message.text;
    if (!withMetadata) continue;
    final names = {...message.placeholders, ...?extraPlaceholders[key]}
      ..remove(message.pluralVar);
    final placeholders = <String, Object?>{
      for (final n in (names.toList()..sort())) n: <String, Object?>{},
      if (message.pluralVar != null)
        message.pluralVar!: <String, Object?>{'type': 'num'},
    };
    if (placeholders.isNotEmpty) {
      doc['@$key'] = <String, Object?>{'placeholders': placeholders};
    }
  }
  return '${const JsonEncoder.withIndent('  ').convert(doc)}\n';
}

Map<String, Object?> _readNamespaces(Directory dir) {
  if (!dir.existsSync()) {
    throw SyncException('Locale folder not found: ${dir.path}');
  }
  final files =
      dir
          .listSync()
          .whereType<File>()
          .where((f) => f.path.endsWith('.json'))
          .toList()
        ..sort((a, b) => a.path.compareTo(b.path));
  final out = <String, Object?>{};
  for (final file in files) {
    final name = file.uri.pathSegments.last.replaceAll(RegExp(r'\.json$'), '');
    try {
      out[name] = jsonDecode(file.readAsStringSync());
    } on FormatException catch (e) {
      throw SyncException('Invalid JSON in ${file.path}: ${e.message}');
    }
  }
  return out;
}

/// Builds the expected ARB file contents: ARB file name -> text.
Map<String, String> buildArbFiles(String webDir) {
  final converted = <String, Map<String, ArbMessage>>{};
  for (final entry in localeFolders.entries) {
    converted[entry.value] = convertLocale(
      _readNamespaces(Directory('$webDir/${entry.key}')),
    );
  }
  final template = converted[templateLocale]!;
  final extra = <String, Set<String>>{};
  for (final entry in converted.entries) {
    for (final m in entry.value.entries) {
      extra.putIfAbsent(m.key, () => {}).addAll(m.value.placeholders);
    }
  }
  final orphans = <String>[
    for (final entry in converted.entries)
      if (entry.key != templateLocale)
        for (final key in entry.value.keys)
          if (!template.containsKey(key)) '${entry.key}:$key',
  ];
  if (orphans.isNotEmpty) {
    throw SyncException(
      'Keys present in a translation but not in $templateLocale:\n'
      '  ${orphans.join('\n  ')}',
    );
  }
  return {
    for (final entry in converted.entries)
      'app_${entry.key}.arb': renderArb(
        entry.key,
        entry.value,
        withMetadata: entry.key == templateLocale,
        extraPlaceholders: extra,
      ),
  };
}

/// Lists differences between expected and on-disk ARB files.
List<String> diffAgainstDisk(Map<String, String> expected, String outDir) {
  final problems = <String>[];
  for (final entry in expected.entries) {
    final file = File('$outDir/${entry.key}');
    if (!file.existsSync()) {
      problems.add('${entry.key}: file is missing');
      continue;
    }
    final actual = file.readAsStringSync();
    if (actual == entry.value) continue;
    final want = jsonDecode(entry.value) as Map<String, Object?>;
    final Map<String, Object?> have;
    try {
      have = jsonDecode(actual) as Map<String, Object?>;
    } on FormatException {
      problems.add('${entry.key}: not valid JSON');
      continue;
    }
    final before = problems.length;
    for (final key in want.keys.where((k) => !have.containsKey(k))) {
      problems.add('${entry.key}: missing key $key');
    }
    for (final key in have.keys.where((k) => !want.containsKey(k))) {
      problems.add('${entry.key}: stale key $key');
    }
    for (final key in want.keys.where(have.containsKey)) {
      if (jsonEncode(want[key]) != jsonEncode(have[key])) {
        problems.add('${entry.key}: value differs for $key');
      }
    }
    if (problems.length == before) {
      problems.add('${entry.key}: formatting differs');
    }
  }
  return problems;
}

String? _option(List<String> args, String name) {
  for (final arg in args) {
    if (arg.startsWith('--$name=')) return arg.substring(name.length + 3);
  }
  return null;
}

/// Entry point usable from tests. Returns the process exit code.
int run(List<String> args, {StringSink? out, StringSink? err}) {
  final stdoutSink = out ?? stdout;
  final stderrSink = err ?? stderr;
  final check = args.contains('--check');
  final webDir = _option(args, 'web-dir') ?? defaultWebDir;
  final outDir = _option(args, 'out-dir') ?? defaultOutDir;
  try {
    final expected = buildArbFiles(webDir);
    if (check) {
      final problems = diffAgainstDisk(expected, outDir);
      if (problems.isNotEmpty) {
        stderrSink.writeln(
          'ARB files are out of sync with the web locales:\n'
          '  ${problems.join('\n  ')}\n'
          'Run: dart run tool/sync_locales.dart',
        );
        return 1;
      }
      stdoutSink.writeln('ARB files are in sync with the web locales.');
      return 0;
    }
    Directory(outDir).createSync(recursive: true);
    for (final entry in expected.entries) {
      File('$outDir/${entry.key}').writeAsStringSync(entry.value);
    }
    stdoutSink.writeln('Wrote ${expected.keys.join(', ')} to $outDir');
    return 0;
  } on SyncException catch (e) {
    stderrSink.writeln('sync_locales: ${e.message}');
    return 1;
  }
}

void main(List<String> args) {
  exitCode = run(args);
}
