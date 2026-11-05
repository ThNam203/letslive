import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/settings/settings_providers.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';

import '../profile/profile_test_app.dart';
import 'message_fakes.dart';
import 'messages_test_app.dart';

/// Sending pictures in a conversation (Task 16 carry-forward): pick, upload
/// through `/upload-file`, send one picture message like web.

Finder _key(String key) => find.byKey(ValueKey(key));

FakeMessagesServer _server() => FakeMessagesServer()
  ..conversations.add(conversationJson('c-1'))
  ..threads['c-1'] = threadJson('c-1', 2);

/// The composer on conversation `c-1`; the photo library returns what
/// [picks] holds (the list is read at each tap).
Future<void> _open(
  WidgetTester tester,
  FakeMessagesServer server,
  List<PickedImage> picks, {
  double width = 800,
  double textScale = 1,
  List<String> oversized = const [],
}) async {
  final app = await pumpMessages(
    tester,
    server,
    width: width,
    textScale: textScale,
    overrides: [
      imagesPickProvider.overrideWithValue(
        (limit) async =>
            (images: picks.take(limit).toList(), oversized: oversized),
      ),
    ],
  );
  await openConversation(tester, app, 'c-1');
}

Future<void> _attach(WidgetTester tester) async {
  await tester.tap(_key('attach-pictures'));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('a picture is uploaded, then sent as one picture message', (
    tester,
  ) async {
    final server = _server();
    await _open(tester, server, [pickedImage(name: 'a.jpg')]);

    await _attach(tester);
    expect(_key('picture-0'), findsOneWidget);
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();

    expect(server.uploadCalls, 1);
    expect(server.pictureBodies.single, {
      'text': 'Sent an image',
      'type': 'image',
      'imageUrls': ['/files/uploads/pic-0.jpg'],
    });
    expect(_key('picture-0'), findsNothing);
  });

  testWidgets('typed text becomes the caption of several pictures', (
    tester,
  ) async {
    final server = _server();
    await _open(tester, server, [
      pickedImage(name: 'a.jpg'),
      pickedImage(name: 'b.jpg'),
    ]);

    await _attach(tester);
    await tester.enterText(_key('message-input'), 'look');
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();

    expect(server.pictureBodies.single['text'], 'look');
    expect(server.pictureBodies.single['imageUrls'], hasLength(2));
  });

  testWidgets('without text, several pictures use the plural caption', (
    tester,
  ) async {
    final server = _server();
    await _open(tester, server, [pickedImage(), pickedImage()]);

    await _attach(tester);
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();

    expect(server.pictureBodies.single['text'], 'Sent 2 images');
  });

  testWidgets('a picture can be removed before sending', (tester) async {
    final server = _server();
    await _open(tester, server, [pickedImage(), pickedImage()]);

    await _attach(tester);
    await tester.tap(_key('remove-picture-0'));
    await tester.pump();
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();

    expect(server.pictureBodies.single['imageUrls'], hasLength(1));
  });

  testWidgets('a failed upload keeps the pictures and Send retries', (
    tester,
  ) async {
    final server = _server()..failingUploads.add(0);
    await _open(tester, server, [pickedImage()]);

    await _attach(tester);
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();
    expect(find.text('Failed to upload files'), findsOneWidget);
    expect(server.pictureBodies, isEmpty);
    expect(_key('picture-0'), findsOneWidget);

    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();
    expect(server.uploadCalls, 2);
    expect(server.pictureBodies, hasLength(1));
  });

  testWidgets('a failed send does not upload the pictures again', (
    tester,
  ) async {
    final server = _server()..sendStatus = 503;
    await _open(tester, server, [pickedImage()]);

    await _attach(tester);
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();
    expect(find.text('Retry'), findsOneWidget);
    expect(_key('picture-0'), findsOneWidget);

    server.sendStatus = 200;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    expect(server.uploadCalls, 1, reason: 'the stored picture is reused');
    expect(server.pictureBodies, hasLength(1));
  });

  testWidgets('when only some pictures upload, those are sent', (tester) async {
    final server = _server()..failingUploads.add(1);
    await _open(tester, server, [pickedImage(), pickedImage()]);

    await _attach(tester);
    await tester.tap(_key('send-message'));
    await tester.pumpAndSettle();

    expect(server.pictureBodies.single['imageUrls'], hasLength(1));
    expect(find.text('Some files failed to upload'), findsOneWidget);
  });

  testWidgets('a picture over 10 MB is refused and not shown', (tester) async {
    final server = _server();
    await _open(tester, server, const [], oversized: ['huge.jpg']);

    await _attach(tester);

    expect(find.text('huge.jpg exceeds 10MB limit'), findsOneWidget);
    expect(_key('picture-0'), findsNothing);
  });

  testWidgets('no more than ${FieldLimits.dmImagesMax} pictures per message', (
    tester,
  ) async {
    final server = _server();
    await _open(tester, server, [
      for (var i = 0; i < FieldLimits.dmImagesMax + 3; i++) pickedImage(),
    ]);

    await _attach(tester);
    await _attach(tester);

    expect(_key('picture-${FieldLimits.dmImagesMax - 1}'), findsOneWidget);
    expect(_key('picture-${FieldLimits.dmImagesMax}'), findsNothing);
  });

  testWidgets('the composer fits 320px at text scale 2.0 with pictures', (
    tester,
  ) async {
    final server = _server();
    await _open(
      tester,
      server,
      [pickedImage(), pickedImage()],
      width: 320,
      textScale: 2,
    );

    await _attach(tester);
    await tester.enterText(_key('message-input'), 'long ${'word ' * 40}');
    await tester.pump();

    expect(tester.takeException(), isNull);
  });
}
