/// The app-wide video upload queue (web `use-upload-store.ts` and its
/// `UploadManager` panel).
///
/// Uploads run in the foreground only: there is no background transfer. The
/// manager lives for the whole signed-in session (not `autoDispose`), so
/// leaving the upload screen does not cancel anything, like web keeps
/// uploads across page navigation. A logout hook cancels everything and
/// empties the list, so the next account starts clean.
///
/// If the app is backgrounded or killed mid-upload and the request fails, the
/// item becomes `failed` and Retry sends the whole file again (the server has
/// no resumable upload). A token that expires before or during the request is
/// refreshed and the request replayed by the API client's auth interceptor
/// (the file is re-read from disk), so the manager never sees it.
library;

import 'dart:async';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/domain/picked_video.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';
import 'package:letslive_mobile/features/vod/vod_providers.dart';
import 'package:letslive_mobile/features/watch/watch_providers.dart';

/// How often the server is asked whether a video finished processing. The
/// transcode worker itself polls its job queue every 5 seconds
/// (`backend/transcode/worker`), so asking faster finds nothing new. Web has
/// no status polling; its upload just ends as completed.
const Duration uploadPollInterval = Duration(seconds: 5);

/// Overridable in tests to make polling instant.
final uploadPollIntervalProvider = Provider<Duration>(
  (ref) => uploadPollInterval,
);

/// Where an upload is.
enum UploadStatus {
  /// Waiting for a free slot (`FieldLimits.vodUploadConcurrency`).
  queued,

  /// The file is being sent.
  uploading,

  /// Sent; the server is transcoding it.
  processing,

  /// The video is `ready`.
  completed,

  /// See [UploadItem.failure]. Retry sends the file again.
  failed,
}

/// Why an upload failed.
enum UploadFailureKind {
  /// The request failed or the server refused it; see
  /// [UploadFailure.response].
  request,

  /// The picked file is gone (the system cleaned its temp copy).
  fileMissing,

  /// The upload worked but the server could not transcode the video.
  processing,
}

/// A failed upload's reason.
class UploadFailure {
  const UploadFailure(this.kind, [this.response]);

  final UploadFailureKind kind;

  /// The failed response when [kind] is [UploadFailureKind.request].
  final ApiResponse<Object?>? response;
}

/// One video in the queue.
class UploadItem {
  const UploadItem({
    required this.id,
    required this.video,
    required this.title,
    required this.description,
    required this.isPublic,
    this.status = UploadStatus.queued,
    this.sent = 0,
    this.failure,
  });

  final String id;
  final PickedVideo video;
  final String title;
  final String description;
  final bool isPublic;
  final UploadStatus status;

  /// Bytes sent so far.
  final int sent;
  final UploadFailure? failure;

  /// Bytes to send (what `onSendProgress` totals can differ slightly by the
  /// multipart envelope, so the file size is used).
  int get total => video.size;

  /// Sent share, 0 to 100.
  int get percent =>
      total <= 0 ? 0 : (sent * 100 / total).clamp(0, 100).round();

  UploadItem copyWith({
    UploadStatus? status,
    int? sent,
    UploadFailure? failure,
    bool clearFailure = false,
  }) => UploadItem(
    id: id,
    video: video,
    title: title,
    description: description,
    isPublic: isPublic,
    status: status ?? this.status,
    sent: sent ?? this.sent,
    failure: clearFailure ? null : failure ?? this.failure,
  );
}

/// The queue, shared by the upload screen and anything that wants to show
/// progress.
final uploadManagerProvider = NotifierProvider<UploadManager, List<UploadItem>>(
  UploadManager.new,
);

/// Runs the uploads: at most `FieldLimits.vodUploadConcurrency` at once, then
/// polls the server until each video is `ready`.
class UploadManager extends Notifier<List<UploadItem>> {
  final Map<String, CancelToken> _cancelTokens = {};
  int _nextId = 0;
  bool _awake = false;

  @override
  List<UploadItem> build() {
    // Logout and session expiry end every upload of the old account.
    ref.read(sessionLifecycleProvider).addLogoutHook(cancelAll);
    return const [];
  }

  /// Adds [video] to the queue and starts it when a slot is free. Returns the
  /// id of the new item. The caller has checked the video with
  /// [checkVideo].
  String enqueue(
    PickedVideo video, {
    required String title,
    required String description,
    required bool isPublic,
  }) {
    final id = 'upload-${_nextId++}';
    _set([
      ...state,
      UploadItem(
        id: id,
        video: video,
        title: title,
        description: description,
        isPublic: isPublic,
      ),
    ]);
    _startQueued();
    return id;
  }

  /// Stops a queued or uploading item and removes it from the list; the
  /// server drops the half-received file when the connection closes. An item
  /// the server is already processing cannot be stopped; [dismiss] it.
  void cancel(String id) {
    final item = _find(id);
    if (item == null ||
        (item.status != UploadStatus.queued &&
            item.status != UploadStatus.uploading)) {
      return;
    }
    dismiss(id);
  }

  /// Sends a failed item again from the start.
  void retry(String id) {
    final item = _find(id);
    if (item == null || item.status != UploadStatus.failed) return;
    _update(
      id,
      (i) =>
          i.copyWith(status: UploadStatus.queued, sent: 0, clearFailure: true),
    );
    _startQueued();
  }

  /// Removes an item from the list. An item still uploading is aborted; one
  /// still processing stops being polled (the server finishes it anyway).
  void dismiss(String id) {
    _cancelTokens.remove(id)?.cancel();
    _set([
      for (final item in state)
        if (item.id != id) item,
    ]);
    _startQueued();
  }

  /// Removes every finished item (web "Clear").
  void dismissFinished() {
    const finished = {UploadStatus.completed, UploadStatus.failed};
    _set([
      for (final item in state)
        if (!finished.contains(item.status)) item,
    ]);
  }

  /// Cancels everything and empties the list (logout hook).
  Future<void> cancelAll() async {
    for (final token in _cancelTokens.values) {
      token.cancel();
    }
    _cancelTokens.clear();
    _set(const []);
  }

  UploadItem? _find(String id) {
    for (final item in state) {
      if (item.id == id) return item;
    }
    return null;
  }

  void _update(String id, UploadItem Function(UploadItem) change) =>
      _set([for (final item in state) item.id == id ? change(item) : item]);

  /// Publishes [items] and keeps the screen awake while a file is being sent.
  void _set(List<UploadItem> items) {
    if (!ref.mounted) return;
    state = List.unmodifiable(items);
    final uploading = items.any((i) => i.status == UploadStatus.uploading);
    if (uploading != _awake) {
      _awake = uploading;
      unawaited(ref.read(screenWakeProvider)(uploading));
    }
  }

  void _startQueued() {
    var free =
        FieldLimits.vodUploadConcurrency -
        state.where((i) => i.status == UploadStatus.uploading).length;
    for (final item in state.where((i) => i.status == UploadStatus.queued)) {
      if (free-- <= 0) return;
      unawaited(_run(item.id));
    }
  }

  Future<void> _run(String id) async {
    final item = _find(id);
    if (item == null) return;
    final cancel = CancelToken();
    _cancelTokens[id] = cancel;
    _update(id, (i) => i.copyWith(status: UploadStatus.uploading));

    // Null when the picked file is gone from disk.
    ApiResponse<Vod>? response;
    try {
      response = await ref
          .read(myVodsRepositoryProvider)
          .upload(
            path: item.video.path,
            fileName: item.video.name,
            title: item.title,
            description: item.description,
            isPublic: item.isPublic,
            cancel: cancel,
            onProgress: (sent, _) {
              if (_find(id)?.status == UploadStatus.uploading) {
                _update(id, (i) => i.copyWith(sent: sent));
              }
            },
          );
    } on FileSystemException {
      response = null;
    }
    _cancelTokens.remove(id);

    // Cancelled, dismissed or logged out while sending: nothing to record.
    if (_find(id)?.status != UploadStatus.uploading) {
      _startQueued();
      return;
    }
    final vod = response?.data;
    if (response == null) {
      _fail(id, const UploadFailure(UploadFailureKind.fileMissing));
    } else if (!response.success || vod == null) {
      _fail(id, UploadFailure(UploadFailureKind.request, response));
    } else {
      _update(
        id,
        (i) => i.copyWith(status: UploadStatus.processing, sent: i.total),
      );
      ref.invalidate(myVodsProvider);
      unawaited(_waitUntilReady(id, vod.id));
    }
    _startQueued();
  }

  void _fail(String id, UploadFailure failure) => _update(
    id,
    (i) => i.copyWith(status: UploadStatus.failed, failure: failure),
  );

  /// Asks the server about [vodId] until it is `ready` or `failed`. A failed
  /// request is skipped (a network blip must not fail a finished upload)
  /// unless the video is gone (404). Stops when the item is dismissed or the
  /// manager is cleared.
  Future<void> _waitUntilReady(String id, String vodId) async {
    final interval = ref.read(uploadPollIntervalProvider);
    while (true) {
      await Future<void>.delayed(interval);
      if (!ref.mounted || _find(id)?.status != UploadStatus.processing) return;
      final response = await ref.read(vodRepositoryProvider).byId(vodId);
      if (!ref.mounted || _find(id)?.status != UploadStatus.processing) return;
      final status = response.data?.status;
      if (status == 'ready') {
        _update(id, (i) => i.copyWith(status: UploadStatus.completed));
        ref.invalidate(myVodsProvider);
        return;
      }
      if (status == 'failed') {
        _fail(id, const UploadFailure(UploadFailureKind.processing));
        return;
      }
      if (response.statusCode == 404) {
        _fail(id, UploadFailure(UploadFailureKind.request, response));
        return;
      }
    }
  }
}
