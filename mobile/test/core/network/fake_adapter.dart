import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';

typedef FakeHandler = FutureOr<ResponseBody> Function(RequestOptions options);

/// In-memory [HttpClientAdapter]: no real network. Records every request.
class FakeAdapter implements HttpClientAdapter {
  FakeAdapter(this.handler);

  FakeHandler handler;
  final List<RequestOptions> requests = [];

  int count(String path) => requests.where((r) => r.path == path).length;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}

ResponseBody jsonBody(
  Object? json, {
  int status = 200,
  Map<String, List<String>> headers = const {},
}) => ResponseBody.fromString(
  jsonEncode(json),
  status,
  headers: {
    Headers.contentTypeHeader: ['application/json; charset=utf-8'],
    ...headers,
  },
);

ResponseBody okEnvelope(Object? data, {Map<String, Object?>? meta}) =>
    jsonBody({
      'requestId': 'req-1',
      'success': true,
      'code': 0,
      'key': 'ok',
      'message': 'ok',
      'data': data,
      'meta': ?meta,
    });
