import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_client.dart';
import 'package:letslive_mobile/core/network/api_endpoints.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/auth/data/auth_repository.dart';
import 'package:letslive_mobile/features/auth/data/bot_proof.dart';
import 'package:letslive_mobile/features/auth/domain/session_lifecycle.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/data/my_vods_repository.dart';
import 'package:letslive_mobile/features/creator/domain/picked_video.dart';
import 'package:letslive_mobile/features/creator/upload_manager.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

import '../../core/network/fake_adapter.dart';
import '../home/home_fixtures.dart';
import 'creator_fakes.dart';
import '../watch/fake_player.dart';

const _base = 'http://localhost:8000';

const _video = PickedVideo(path: '/tmp/a.mp4', name: 'a.mp4', size: 1000);

void main() {
  late CookieJar jar;
  late ApiClient client;
  late FakeScreenWake wake;
  late FakeMyVods myVods;
  late FakeVodStatus status;

  ProviderContainer makeContainer() {
    final container = ProviderContainer(
      overrides: [
        myVodsRepositoryProvider.overrideWithValue(myVods),
        vodRepositoryProvider.overrideWithValue(status),
        screenWakeProvider.overrideWithValue(wake.call),
        uploadPollIntervalProvider.overrideWithValue(
          const Duration(milliseconds: 1),
        ),
        sessionLifecycleProvider.overrideWith(
          (ref) => SessionLifecycle(
            auth: AuthRepository(
              client: client,
              botProof: const BotProof.turnstile('t'),
            ),
            cookieJar: jar,
            restartRealtime: () {},
          ),
        ),
      ],
    );
    addTearDown(container.dispose);
    return container;
  }

  Future<void> until(bool Function() done) async {
    for (var i = 0; i < 200 && !done(); i++) {
      await Future<void>.delayed(const Duration(milliseconds: 2));
    }
    expect(done(), isTrue, reason: 'condition not reached in time');
  }

  setUp(() {
    jar = CookieJar();
    client = ApiClient(
      baseUrl: _base,
      cookieJar: jar,
      adapter: FakeAdapter((_) => ResponseBody.fromString('', 204)),
    );
    wake = FakeScreenWake();
    status = FakeVodStatus(['processing', 'processing', 'ready']);
    myVods = FakeMyVods(
      (_, _) async => okResponse(vodWithStatus('processing')),
    );
  });

  tearDown(() => client.dispose());

  group('checkVideo', () {
    test('rejects unsupported extension before upload', () {
      const gif = PickedVideo(path: '/x', name: 'clip.gif', size: 10);
      const noExtension = PickedVideo(path: '/x', name: 'clip', size: 10);
      expect(checkVideo(gif), VideoProblem.unsupportedFormat);
      expect(checkVideo(noExtension), VideoProblem.unsupportedFormat);
    });

    test('accepts every supported extension in any case', () {
      for (final name in ['a.mp4', 'a.MOV', 'a.avi', 'a.mkv', 'a.WebM']) {
        expect(checkVideo(PickedVideo(path: '/x', name: name, size: 10)), null);
      }
    });

    test('rejects a file over 100 MB and accepts exactly 100 MB', () {
      const limit = 100 * 1024 * 1024;
      expect(
        checkVideo(const PickedVideo(path: '/x', name: 'a.mp4', size: limit)),
        isNull,
      );
      expect(
        checkVideo(
          const PickedVideo(path: '/x', name: 'a.mp4', size: limit + 1),
        ),
        VideoProblem.tooLarge,
      );
    });

    test('suggests the file name without extension as the title', () {
      expect(suggestedTitle('holiday.final.mp4'), 'holiday.final');
      expect(suggestedTitle('${'x' * 300}.mp4'), hasLength(255));
    });
  });

  test('progress events reach 100 then status polling until ready', () async {
    myVods.onUpload = (onProgress, _) async {
      onProgress?.call(500, 1000);
      onProgress?.call(1000, 1000);
      return okResponse(vodWithStatus('processing'));
    };
    final container = makeContainer();
    final manager = container.read(uploadManagerProvider.notifier);
    final seen = <UploadStatus>[];
    container.listen(uploadManagerProvider, (_, items) {
      if (items.isNotEmpty) seen.add(items.single.status);
    });

    final id = manager.enqueue(
      _video,
      title: 'T',
      description: 'D',
      isPublic: true,
    );
    await until(
      () =>
          container.read(uploadManagerProvider).single.status ==
          UploadStatus.completed,
    );

    expect(seen.toSet(), {
      UploadStatus.queued,
      UploadStatus.uploading,
      UploadStatus.processing,
      UploadStatus.completed,
    });
    final item = container.read(uploadManagerProvider).single;
    expect(item.id, id);
    expect(item.percent, 100);
    expect(status.polls, 3, reason: 'two processing answers, then ready');
  });

  test('keeps the screen awake only while a file is being sent', () async {
    final sending = Completer<ApiResponse<Vod>>();
    myVods.onUpload = (_, _) => sending.future;
    final container = makeContainer();
    container
        .read(uploadManagerProvider.notifier)
        .enqueue(_video, title: 'T', description: '', isPublic: true);
    expect(wake.calls, [true]);

    sending.complete(okResponse(vodWithStatus('processing')));
    await until(() => wake.calls.length == 2);
    expect(wake.calls, [true, false]);
  });

  test('an upload keeps the screen awake after a player lets go', () async {
    final sending = Completer<ApiResponse<Vod>>();
    myVods.onUpload = (_, _) => sending.future;
    final container = makeContainer();
    final holder = container.read(screenWakeHolderProvider);
    final player = Object();

    holder.acquire(player);
    container
        .read(uploadManagerProvider.notifier)
        .enqueue(_video, title: 'T', description: '', isPublic: true);
    holder.release(player);
    expect(wake.calls, [true], reason: 'the upload still holds the screen');

    sending.complete(okResponse(vodWithStatus('processing')));
    await until(() => wake.calls.length == 2);
    expect(wake.calls, [true, false]);
  });

  test('cancel aborts the request and removes the entry', () async {
    final aborted = Completer<void>();
    myVods.onUpload = (_, cancel) async {
      await cancel!.whenCancel;
      aborted.complete();
      return ApiResponse<Vod>.transportFailure(
        ApiFailureKind.cancelled,
        'Request cancelled',
      );
    };
    final container = makeContainer();
    final manager = container.read(uploadManagerProvider.notifier);
    final id = manager.enqueue(
      _video,
      title: 'T',
      description: '',
      isPublic: true,
    );

    manager.cancel(id);
    await aborted.future;
    await Future<void>.delayed(Duration.zero);

    expect(container.read(uploadManagerProvider), isEmpty);
    expect(wake.isAwake, isFalse);
  });

  test(
    'a failed upload shows the reason and retry sends the file again',
    () async {
      var calls = 0;
      myVods.onUpload = (_, _) async {
        calls++;
        return calls == 1
            ? failResponse(503)
            : okResponse(vodWithStatus('processing'));
      };
      final container = makeContainer();
      final manager = container.read(uploadManagerProvider.notifier);
      final id = manager.enqueue(
        _video,
        title: 'T',
        description: '',
        isPublic: true,
      );
      await until(
        () =>
            container.read(uploadManagerProvider).single.status ==
            UploadStatus.failed,
      );
      final failed = container.read(uploadManagerProvider).single;
      expect(failed.failure?.kind, UploadFailureKind.request);
      expect(failed.failure?.response?.statusCode, 503);

      manager.retry(id);
      await until(
        () =>
            container.read(uploadManagerProvider).single.status ==
            UploadStatus.completed,
      );
      expect(myVods.uploadedPaths, [_video.path, _video.path]);
    },
  );

  test('a video the server cannot transcode ends as failed', () async {
    status.statuses = ['processing', 'failed'];
    final container = makeContainer();
    container
        .read(uploadManagerProvider.notifier)
        .enqueue(_video, title: 'T', description: '', isPublic: true);
    await until(
      () =>
          container.read(uploadManagerProvider).single.status ==
          UploadStatus.failed,
    );
    expect(
      container.read(uploadManagerProvider).single.failure?.kind,
      UploadFailureKind.processing,
    );
  });

  test('polling stops when the entry is dismissed', () async {
    status.statuses = ['processing'];
    final container = makeContainer();
    final manager = container.read(uploadManagerProvider.notifier);
    final id = manager.enqueue(
      _video,
      title: 'T',
      description: '',
      isPublic: true,
    );
    await until(() => status.polls >= 2);

    manager.dismiss(id);
    await Future<void>.delayed(const Duration(milliseconds: 20));
    final pollsAtDismiss = status.polls;
    await Future<void>.delayed(const Duration(milliseconds: 30));
    expect(status.polls, pollsAtDismiss);
  });

  test('only three uploads run at once; the others wait', () async {
    final gates = <Completer<ApiResponse<Vod>>>[];
    myVods.onUpload = (_, _) {
      final gate = Completer<ApiResponse<Vod>>();
      gates.add(gate);
      return gate.future;
    };
    final container = makeContainer();
    final manager = container.read(uploadManagerProvider.notifier);
    for (var i = 0; i < 4; i++) {
      manager.enqueue(_video, title: 'T$i', description: '', isPublic: true);
    }
    List<UploadStatus> statuses() => [
      for (final i in container.read(uploadManagerProvider)) i.status,
    ];
    expect(statuses(), [
      UploadStatus.uploading,
      UploadStatus.uploading,
      UploadStatus.uploading,
      UploadStatus.queued,
    ]);

    gates.first.complete(okResponse(vodWithStatus('processing')));
    await until(() => gates.length == 4);
    expect(statuses().last, UploadStatus.uploading);
  });

  test(
    'logout cancels in-flight uploads and the next account starts clean',
    () async {
      final aborted = Completer<void>();
      myVods.onUpload = (_, cancel) async {
        await cancel!.whenCancel;
        aborted.complete();
        return ApiResponse<Vod>.transportFailure(
          ApiFailureKind.cancelled,
          'Request cancelled',
        );
      };
      final container = makeContainer();
      container
          .read(uploadManagerProvider.notifier)
          .enqueue(
            _video,
            title: 'Account A video',
            description: '',
            isPublic: true,
          );
      expect(container.read(uploadManagerProvider), hasLength(1));

      await container
          .read(sessionLifecycleProvider)
          .endSession(notifyServer: false);
      await aborted.future;

      expect(container.read(uploadManagerProvider), isEmpty);
      expect(wake.isAwake, isFalse);

      // Account B queues its own upload on the same manager.
      myVods.onUpload = (_, _) async => okResponse(vodWithStatus('processing'));
      container
          .read(uploadManagerProvider.notifier)
          .enqueue(
            _video,
            title: 'Account B video',
            description: '',
            isPublic: true,
          );
      expect(container.read(uploadManagerProvider).map((i) => i.title), [
        'Account B video',
      ]);
    },
  );

  test('a picked file that vanished fails with a clear reason', () async {
    myVods = FakeMyVods(
      (_, _) => throw const FileSystemException('gone', '/tmp/a.mp4'),
    );
    final container = makeContainer();
    container
        .read(uploadManagerProvider.notifier)
        .enqueue(_video, title: 'T', description: '', isPublic: true);
    await until(
      () =>
          container.read(uploadManagerProvider).single.status ==
          UploadStatus.failed,
    );
    expect(
      container.read(uploadManagerProvider).single.failure?.kind,
      UploadFailureKind.fileMissing,
    );
  });

  group('token expiry mid-upload (Review Focus 1)', () {
    late Directory dir;
    late File file;
    late List<int> bodyLengths;
    late int refreshes;
    late int uploadCalls;
    late ApiClient real;

    setUp(() async {
      dir = await Directory.systemTemp.createTemp('upload_test');
      file = File('${dir.path}/clip.mp4')
        ..writeAsBytesSync(Uint8List.fromList(List.filled(50000, 7)));
      bodyLengths = [];
      refreshes = 0;
      uploadCalls = 0;
      final signedIn = CookieJar();
      await signedIn.saveFromResponse(Uri.parse(_base), [
        Cookie('ACCESS_TOKEN', 'old'),
        Cookie('REFRESH_TOKEN', 'refresh'),
      ]);
      real = ApiClient(
        baseUrl: _base,
        cookieJar: signedIn,
        adapter: _StreamingAdapter((options, bytes) async {
          if (options.path == ApiEndpoints.refreshToken) {
            refreshes++;
            return okEnvelope(null);
          }
          uploadCalls++;
          bodyLengths.add(bytes);
          return uploadCalls == 1
              ? jsonBody({
                  'requestId': 'r',
                  'success': false,
                  'code': 1,
                  'key': 'res_err_unauthorized',
                  'message': 'expired',
                }, status: 401)
              : okEnvelope(vodJson(id: 'vod-9'));
        }),
      );
      addTearDown(() async {
        await real.dispose();
        await dir.delete(recursive: true);
      });
    });

    test('a 401 is refreshed once and the whole file is sent again', () async {
      final response = await MyVodsRepository(real).upload(
        path: file.path,
        fileName: 'clip.mp4',
        title: 'T',
        description: '',
        isPublic: true,
      );

      expect(response.success, isTrue);
      expect(refreshes, 1);
      expect(uploadCalls, 2);
      expect(bodyLengths[1], bodyLengths[0]);
      expect(bodyLengths[1], greaterThan(50000));
    });

    test('a failed refresh leaves the upload failed, without a loop', () async {
      final expired = CookieJar();
      await expired.saveFromResponse(Uri.parse(_base), [
        Cookie('ACCESS_TOKEN', 'old'),
        Cookie('REFRESH_TOKEN', 'also-expired'),
      ]);
      var rejectedRefreshes = 0;
      final noRefresh = ApiClient(
        baseUrl: _base,
        cookieJar: expired,
        adapter: _StreamingAdapter((options, bytes) async {
          if (options.path == ApiEndpoints.refreshToken) {
            rejectedRefreshes++;
          } else {
            uploadCalls++;
          }
          return jsonBody({
            'requestId': 'r',
            'success': false,
            'code': 1,
            'key': 'res_err_unauthorized',
            'message': 'expired',
          }, status: 401);
        }),
      );
      addTearDown(noRefresh.dispose);

      final response = await MyVodsRepository(noRefresh).upload(
        path: file.path,
        fileName: 'clip.mp4',
        title: 'T',
        description: '',
        isPublic: true,
      );

      expect(response.success, isFalse);
      expect(response.statusCode, 401);
      expect(rejectedRefreshes, 1, reason: 'refreshed once, not in a loop');
      expect(uploadCalls, 1, reason: 'the upload is not replayed');
    });
  });
}

/// Reads the whole request body (like a real socket would) before answering,
/// and reports how many bytes it held.
class _StreamingAdapter implements HttpClientAdapter {
  _StreamingAdapter(this.handler);

  final Future<ResponseBody> Function(RequestOptions options, int bodyBytes)
  handler;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    var total = 0;
    if (requestStream != null) {
      await for (final chunk in requestStream) {
        total += chunk.length;
      }
    }
    return handler(options, total);
  }

  @override
  void close({bool force = false}) {}
}
