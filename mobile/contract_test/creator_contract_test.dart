import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/features/creator/data/chat_command_admin_repository.dart';
import 'package:letslive_mobile/features/creator/data/my_vods_repository.dart';
import 'package:letslive_mobile/features/creator/data/stream_settings_repository.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/data/message_repository.dart';
import 'package:letslive_mobile/features/user/data/me_user.dart';
import 'package:letslive_mobile/features/user/data/picked_image.dart';
import 'package:letslive_mobile/features/user/data/upload_file_repository.dart';
import 'package:letslive_mobile/features/vod/data/vod_repository.dart';

import 'harness.dart';
import 'png.dart';

/// How long the video is polled for `ready`. The mobile contract stack has no
/// transcode worker, so there it stays `processing`; with the worker running
/// raise it (seconds) with `--dart-define=CONTRACT_VOD_READY_WAIT=120`.
const int _readyWaitSeconds = int.fromEnvironment(
  'CONTRACT_VOD_READY_WAIT',
  defaultValue: 10,
);

/// The creator tools against the real dev stack: stream key and stream
/// information, chat commands, the general file upload with a DM picture, and
/// the video upload / list / edit / delete. User A is the creator, user B the
/// DM partner.
void main() {
  ContractContext? ctx;
  late StreamSettingsRepository stream;
  late ChatCommandAdminRepository commands;
  late MyVodsRepository myVods;
  late UploadFileRepository files;
  late Directory scratch;

  setUpAll(() async {
    final context = ctx = await startContractContext();
    stream = StreamSettingsRepository(context.userA.client);
    commands = ChatCommandAdminRepository(context.userA.client);
    myVods = MyVodsRepository(context.userA.client);
    files = UploadFileRepository(context.userA.client);
    scratch = await Directory.systemTemp.createTemp('creator_contract');
  });

  tearDownAll(() async {
    try {
      await scratch.delete(recursive: true);
    } finally {
      await ctx?.dispose();
    }
  });

  test('the stream key is read and reset', () async {
    final before = await stream.fetch();
    expect(before.success, isTrue, reason: '${before.key}: ${before.message}');
    expect(before.data!.streamKey, isNotEmpty);

    final reset = await stream.resetKey();
    expect(reset.success, isTrue, reason: '${reset.key}: ${reset.message}');
    expect(reset.data, isNot(before.data!.streamKey));

    final after = await stream.fetch();
    expect(after.data!.streamKey, reset.data);
    // The key is the same one /user/me reports.
    final me = await ctx!.userA.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    expect(me.data!.streamApiKey, reset.data);
  });

  test('stream information is saved, with and without a thumbnail', () async {
    final text = await stream.update(
      title: 'Contract stream Việt 🎉',
      description: 'About the stream',
    );
    expect(text.success, isTrue, reason: '${text.key}: ${text.message}');
    expect(text.data!.title, 'Contract stream Việt 🎉');
    expect(text.data!.thumbnailUrl, isNull);

    final withThumbnail = await stream.update(
      title: 'With thumbnail',
      description: '',
      thumbnail: PickedImage(
        name: 'thumb.png',
        bytes: solidPng(64, 36),
        mimeType: 'image/png',
      ),
    );
    expect(
      withThumbnail.success,
      isTrue,
      reason: '${withThumbnail.key}: ${withThumbnail.message}',
    );
    final url = withThumbnail.data!.thumbnailUrl;
    expect(url, isNotNull);

    // Keeping the thumbnail sends its URL back, as web does.
    final kept = await stream.update(
      title: 'Kept',
      description: '',
      keepThumbnailUrl: url,
    );
    expect(kept.data!.thumbnailUrl, url);

    final fetched = await stream.fetch();
    expect(fetched.data!.title, 'Kept');
    expect(fetched.data!.thumbnailUrl, url);
  });

  test('chat commands: create, list, edit, delete, name conflict', () async {
    final created = await commands.create(
      scope: ChatCommandScope.user,
      name: 'discord',
      response: 'Join us at discord.gg/example',
      description: 'Our server',
    );
    expect(created.success, isTrue, reason: created.key);

    final channel = await commands.create(
      scope: ChatCommandScope.channel,
      name: 'rules',
      response: 'Be kind',
      description: '',
    );
    expect(channel.success, isTrue, reason: channel.key);

    final listed = await commands.mine();
    expect(listed.success, isTrue, reason: listed.key);
    expect(listed.data!.user.map((c) => c.name), ['discord']);
    expect(listed.data!.channel.map((c) => c.name), ['rules']);
    expect(listed.data!.user.single.description, 'Our server');

    final duplicate = await commands.create(
      scope: ChatCommandScope.user,
      name: 'discord',
      response: 'again',
      description: '',
    );
    expect(duplicate.success, isFalse);

    final edited = await commands.update(
      created.data!.id,
      name: 'discord',
      response: 'New invite',
      description: '',
    );
    expect(edited.success, isTrue, reason: edited.key);
    expect(edited.data!.response, 'New invite');

    for (final id in [created.data!.id, channel.data!.id]) {
      final removed = await commands.delete(id);
      expect(removed.success, isTrue, reason: removed.key);
    }
    final empty = await commands.mine();
    expect(empty.data!.user, isEmpty);
    expect(empty.data!.channel, isEmpty);
  });

  test('a DM picture goes through /upload-file and arrives as image', () async {
    final uploaded = await files.upload(
      PickedImage(
        name: 'pic.png',
        bytes: solidPng(64, 64),
        mimeType: 'image/png',
      ),
    );
    expect(uploaded.success, isTrue, reason: uploaded.key);
    expect(uploaded.data, isNotEmpty);

    final a = MessageRepository(ctx!.userA.client);
    final b = MessageRepository(ctx!.userB.client);
    final bMe = await ctx!.userB.client.get<MeUser>(
      ApiEndpoints.userMe,
      parse: MeUser.parse,
    );
    final conversation = await a.create(
      type: ConversationType.dm,
      participantIds: [bMe.data!.id],
    );
    expect(conversation.success, isTrue, reason: conversation.key);

    final sent = await a.send(
      conversation.data!.id,
      'Sent an image',
      imageUrls: [uploaded.data!],
    );
    expect(sent.success, isTrue, reason: '${sent.key}: ${sent.message}');
    expect(sent.data!.type, DmMessageType.image);
    expect(sent.data!.imageUrls, [uploaded.data]);

    final history = await b.messages(conversation.data!.id);
    expect(history.data!.single.imageUrls, [uploaded.data]);
  });

  test('a video uploads, is listed, edited and deleted', () async {
    final mp4 = File('${scratch.path}/clip.mp4');
    final made = await Process.run('ffmpeg', [
      '-loglevel',
      'error',
      '-f',
      'lavfi',
      '-i',
      'testsrc=duration=2:size=160x90:rate=10',
      '-f',
      'lavfi',
      '-i',
      'sine=frequency=440:duration=2',
      '-c:v',
      'libx264',
      '-pix_fmt',
      'yuv420p',
      '-c:a',
      'aac',
      '-shortest',
      mp4.path,
    ]);
    expect(made.exitCode, 0, reason: "${made.stderr}");

    var lastProgress = 0;
    final uploaded = await myVods.upload(
      path: mp4.path,
      fileName: 'clip.mp4',
      title: 'Contract upload Việt 🎉',
      description: 'From the contract test',
      isPublic: false,
      onProgress: (sent, total) => lastProgress = sent,
    );
    expect(uploaded.success, isTrue, reason: uploaded.key);
    final vod = uploaded.data!;
    expect(vod.status, 'processing');
    expect(vod.visibility, 'private');
    expect(lastProgress, greaterThan(0));

    final listed = await myVods.list(0);
    expect(listed.success, isTrue, reason: listed.key);
    expect(listed.data!.map((v) => v.id), contains(vod.id));

    // The owner polls the status like the upload manager does.
    final viewer = VodRepository(ctx!.userA.client);
    final deadline = DateTime.now().add(
      const Duration(seconds: _readyWaitSeconds),
    );
    var status = vod.status;
    while (status == 'processing' && DateTime.now().isBefore(deadline)) {
      await Future<void>.delayed(const Duration(seconds: 2));
      final polled = await viewer.byId(vod.id);
      expect(polled.success, isTrue, reason: polled.key);
      status = polled.data!.status;
    }
    // ignore: avoid_print
    print('uploaded video status after ${_readyWaitSeconds}s: $status');
    expect(['processing', 'ready'], contains(status));

    final edited = await myVods.update(
      vod.id,
      title: 'Renamed upload',
      description: '',
      isPublic: true,
    );
    expect(edited.success, isTrue, reason: edited.key);
    final after = await viewer.byId(vod.id);
    expect(after.data!.title, 'Renamed upload');
    expect(after.data!.visibility, 'public');

    final removed = await myVods.delete(vod.id);
    expect(removed.success, isTrue, reason: removed.key);
    final gone = await viewer.byId(vod.id);
    expect(gone.success, isFalse);
    expect(gone.statusCode, 404);
  });

  test(
    'a file with an unsupported extension is refused by the server',
    () async {
      final txt = File('${scratch.path}/note.txt')..writeAsStringSync('hello');
      final refused = await myVods.upload(
        path: txt.path,
        fileName: 'note.txt',
        title: 'Not a video',
        description: '',
        isPublic: true,
      );
      expect(refused.success, isFalse);
    },
  );
}
