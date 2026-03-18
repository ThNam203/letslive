import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../providers.dart';
import 'vod_repository.dart';

const _maxConcurrent = 3;

enum UploadItemStatus { queued, uploading, processing, completed, failed, cancelled }

class UploadItem {
  final String id;
  final File file;
  final String title;
  final String description;
  final String visibility;
  final UploadItemStatus status;
  final double progress; // 0.0 - 1.0
  final int loaded;
  final int total;
  final String? error;
  final CancelToken? cancelToken;

  const UploadItem({
    required this.id,
    required this.file,
    required this.title,
    required this.description,
    required this.visibility,
    this.status = UploadItemStatus.queued,
    this.progress = 0.0,
    this.loaded = 0,
    this.total = 0,
    this.error,
    this.cancelToken,
  });

  UploadItem copyWith({
    UploadItemStatus? status,
    double? progress,
    int? loaded,
    int? total,
    String? error,
    CancelToken? cancelToken,
  }) {
    return UploadItem(
      id: id,
      file: file,
      title: title,
      description: description,
      visibility: visibility,
      status: status ?? this.status,
      progress: progress ?? this.progress,
      loaded: loaded ?? this.loaded,
      total: total ?? this.total,
      error: error ?? this.error,
      cancelToken: cancelToken ?? this.cancelToken,
    );
  }
}

class UploadQueueState {
  final List<UploadItem> items;
  final bool isCollapsed;

  const UploadQueueState({
    this.items = const [],
    this.isCollapsed = false,
  });

  UploadQueueState copyWith({
    List<UploadItem>? items,
    bool? isCollapsed,
  }) {
    return UploadQueueState(
      items: items ?? this.items,
      isCollapsed: isCollapsed ?? this.isCollapsed,
    );
  }

  int get activeCount =>
      items.where((i) => i.status == UploadItemStatus.uploading || i.status == UploadItemStatus.queued).length;

  int get completedCount =>
      items.where((i) => i.status == UploadItemStatus.completed).length;

  int get failedCount =>
      items.where((i) => i.status == UploadItemStatus.failed).length;
}

class UploadQueueNotifier extends Notifier<UploadQueueState> {
  @override
  UploadQueueState build() => const UploadQueueState();

  VodRepository get _repo => ref.read(vodRepositoryProvider);

  void enqueue({
    required File file,
    required String title,
    String description = '',
    String visibility = 'public',
  }) {
    final item = UploadItem(
      id: DateTime.now().microsecondsSinceEpoch.toString(),
      file: file,
      title: title,
      description: description,
      visibility: visibility,
    );
    state = state.copyWith(items: [...state.items, item]);
    _processQueue();
  }

  void cancel(String id) {
    final item = state.items.where((i) => i.id == id).firstOrNull;
    item?.cancelToken?.cancel();
    state = state.copyWith(
      items: state.items
          .map((i) => i.id == id ? i.copyWith(status: UploadItemStatus.cancelled) : i)
          .toList(),
    );
    _processQueue();
  }

  void retry(String id) {
    state = state.copyWith(
      items: state.items
          .map((i) => i.id == id
              ? UploadItem(
                  id: i.id,
                  file: i.file,
                  title: i.title,
                  description: i.description,
                  visibility: i.visibility,
                )
              : i)
          .toList(),
    );
    _processQueue();
  }

  void dismiss(String id) {
    state = state.copyWith(
      items: state.items.where((i) => i.id != id).toList(),
    );
  }

  void dismissCompleted() {
    state = state.copyWith(
      items: state.items
          .where((i) =>
              i.status != UploadItemStatus.completed &&
              i.status != UploadItemStatus.failed &&
              i.status != UploadItemStatus.cancelled)
          .toList(),
    );
  }

  void toggleCollapsed() {
    state = state.copyWith(isCollapsed: !state.isCollapsed);
  }

  void _processQueue() {
    final activeCount = state.items.where((i) => i.status == UploadItemStatus.uploading).length;
    final availableSlots = _maxConcurrent - activeCount;
    if (availableSlots <= 0) return;

    final queued = state.items.where((i) => i.status == UploadItemStatus.queued).toList();
    final toStart = queued.take(availableSlots);

    for (final item in toStart) {
      _startUpload(item);
    }
  }

  Future<void> _startUpload(UploadItem item) async {
    final cancelToken = CancelToken();

    _updateItem(item.id, (i) => i.copyWith(
      status: UploadItemStatus.uploading,
      cancelToken: cancelToken,
    ));

    try {
      final response = await _repo.uploadVod(
        videoFile: item.file,
        title: item.title,
        description: item.description,
        visibility: item.visibility,
        cancelToken: cancelToken,
        onSendProgress: (sent, total) {
          _updateItem(item.id, (i) => i.copyWith(
            loaded: sent,
            total: total,
            progress: total > 0 ? sent / total : 0,
          ));
        },
      );

      // Check if cancelled during upload
      final current = state.items.where((i) => i.id == item.id).firstOrNull;
      if (current?.status == UploadItemStatus.cancelled) return;

      if (response.success) {
        _updateItem(item.id, (i) => i.copyWith(
          status: UploadItemStatus.completed,
          progress: 1.0,
        ));
      } else {
        _updateItem(item.id, (i) => i.copyWith(
          status: UploadItemStatus.failed,
          error: response.message,
        ));
      }
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      _updateItem(item.id, (i) => i.copyWith(
        status: UploadItemStatus.failed,
        error: e.message ?? 'Upload failed',
      ));
    } catch (e) {
      _updateItem(item.id, (i) => i.copyWith(
        status: UploadItemStatus.failed,
        error: e.toString(),
      ));
    } finally {
      _processQueue();
    }
  }

  void _updateItem(String id, UploadItem Function(UploadItem) updater) {
    state = state.copyWith(
      items: state.items
          .map((i) => i.id == id ? updater(i) : i)
          .toList(),
    );
  }
}

final uploadQueueProvider =
    NotifierProvider<UploadQueueNotifier, UploadQueueState>(
        UploadQueueNotifier.new);
