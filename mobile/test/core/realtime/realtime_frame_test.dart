import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/realtime/realtime_frame.dart';

void main() {
  test('parses an event frame', () {
    final frame = parseRealtimeFrame(
      '{"op":"event","topic":"room:1","type":"chat.message","data":{"a":1}}',
    );
    expect(frame, isA<RealtimeEventFrame>());
    final event = frame! as RealtimeEventFrame;
    expect(event.topic, 'room:1');
    expect(event.type, 'chat.message');
    expect(event.data, {'a': 1});
  });

  test('parses error and pong frames', () {
    expect(
      (parseRealtimeFrame('{"op":"error","code":"bad_frame"}')!
              as RealtimeErrorFrame)
          .code,
      'bad_frame',
    );
    expect(parseRealtimeFrame('{"op":"pong"}'), isA<RealtimePongFrame>());
  });

  test('returns null for malformed or unknown frames', () {
    expect(parseRealtimeFrame('not json'), isNull);
    expect(parseRealtimeFrame('[1]'), isNull);
    expect(parseRealtimeFrame('{"op":"wat"}'), isNull);
    expect(parseRealtimeFrame('{"op":"event","topic":"x"}'), isNull);
    expect(parseRealtimeFrame('{"op":"error"}'), isNull);
    expect(parseRealtimeFrame(42), isNull);
    expect(parseRealtimeFrame(null), isNull);
  });

  test('encodes client frames', () {
    expect(
      encodeRealtimeFrame('subscribe', topic: 'room:1'),
      '{"op":"subscribe","topic":"room:1"}',
    );
    expect(encodeRealtimeFrame('ping'), '{"op":"ping"}');
  });
}
