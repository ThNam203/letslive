/// Strict readers for the fields of one JSON object, shared by the response
/// models. A missing or mistyped field throws [FormatException] naming the
/// model and key, which `ApiClient` turns into a `malformedResponse` failure,
/// so backend drift fails loudly instead of showing wrong data.
class JsonFields {
  const JsonFields(this.model, this.json);

  /// Model name, used in error messages.
  final String model;
  final Map<String, Object?> json;

  /// A required string.
  String string(String key) {
    final value = json[key];
    if (value is String) return value;
    throw _error(key, 'string', value);
  }

  /// A string the backend really sends as null for "not set".
  String? optString(String key) {
    final value = json[key];
    if (value == null || value is String) return value as String?;
    throw _error(key, 'string or null', value);
  }

  /// A boolean the backend omits or sends as null when it does not apply.
  bool? optBool(String key) {
    final value = json[key];
    if (value == null || value is bool) return value as bool?;
    throw _error(key, 'bool or null', value);
  }

  /// A required whole number.
  int integer(String key) {
    final value = json[key];
    if (value is int) return value;
    throw _error(key, 'int', value);
  }

  /// A required boolean.
  bool boolean(String key) {
    final value = json[key];
    if (value is bool) return value;
    throw _error(key, 'bool', value);
  }

  /// A required ISO 8601 timestamp.
  DateTime date(String key) {
    final value = DateTime.tryParse(string(key));
    if (value != null) return value;
    throw _error(key, 'ISO 8601 date', json[key]);
  }

  /// An ISO 8601 timestamp the backend sends as null while unset.
  DateTime? optDate(String key) => json[key] == null ? null : date(key);

  FormatException _error(String key, String expected, Object? got) =>
      FormatException('$model.$key: expected $expected, got $got');
}

/// Parses a JSON array of objects with [fromJson]; [model] names the item type
/// in errors. Use inside the `parse` of a list endpoint.
List<T> parseJsonList<T>(
  Object? data,
  String model,
  T Function(Map<String, Object?> json) fromJson,
) {
  if (data is! List<Object?>) {
    throw FormatException('$model list: expected a JSON array, got $data');
  }
  return [
    for (final item in data)
      if (item is Map<String, Object?>)
        fromJson(item)
      else
        throw FormatException('$model list: expected objects, got $item'),
  ];
}

/// Parses a single JSON object with [fromJson].
T parseJsonObject<T>(
  Object? data,
  String model,
  T Function(Map<String, Object?> json) fromJson,
) {
  if (data is! Map<String, Object?>) {
    throw FormatException('$model: expected a JSON object, got $data');
  }
  return fromJson(data);
}
