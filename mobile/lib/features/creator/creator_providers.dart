/// Riverpod providers of the creator tools (web `/settings/stream`,
/// `/settings/chat-commands`, `/settings/vods`). The upload manager is in
/// `upload_manager.dart`.
///
/// Every data provider is `autoDispose`, so leaving the signed-in shell on
/// logout drops it and the next account never sees the previous account's
/// stream key, commands or videos.
library;

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/features/creator/data/chat_command_admin_repository.dart';
import 'package:letslive_mobile/features/creator/data/my_vods_repository.dart';
import 'package:letslive_mobile/features/creator/data/stream_settings_repository.dart';
import 'package:letslive_mobile/features/creator/domain/picked_video.dart';
import 'package:letslive_mobile/features/home/data/vod.dart';

/// Stream key and stream information calls.
final streamSettingsRepositoryProvider = Provider<StreamSettingsRepository>(
  (ref) => StreamSettingsRepository(ref.watch(apiClientProvider)),
);

/// Chat-command admin calls.
final chatCommandAdminRepositoryProvider = Provider<ChatCommandAdminRepository>(
  (ref) => ChatCommandAdminRepository(ref.watch(apiClientProvider)),
);

/// The creator's own video calls.
final myVodsRepositoryProvider = Provider<MyVodsRepository>(
  (ref) => MyVodsRepository(ref.watch(apiClientProvider)),
);

/// The stream key and stream information. Held in this provider so a key
/// reset or a save updates what the screen shows without another request.
final streamSettingsProvider =
    AsyncNotifierProvider.autoDispose<StreamSettingsNotifier, StreamSettings>(
      StreamSettingsNotifier.new,
      retry: noRetry,
    );

/// Loads [StreamSettings] and applies a key reset.
class StreamSettingsNotifier extends AsyncNotifier<StreamSettings> {
  @override
  Future<StreamSettings> build() async =>
      unwrapData(await ref.watch(streamSettingsRepositoryProvider).fetch());

  /// Generates a new key. The old key stops working at once, so the caller
  /// asks the user to confirm first. Returns the failed response, or null on
  /// success.
  Future<ApiResponse<Object?>?> resetKey() async {
    final response = await ref
        .read(streamSettingsRepositoryProvider)
        .resetKey();
    final key = response.data;
    final current = state.value;
    if (!ref.mounted || !response.success || key == null || current == null) {
      return response.success ? null : response;
    }
    state = AsyncData(current.withKey(key));
    return null;
  }

  /// Shows the information the server saved.
  void applyInfo(StreamInfo info) {
    final current = state.value;
    if (current == null) return;
    state = AsyncData(
      StreamSettings(
        streamKey: current.streamKey,
        title: info.title,
        description: info.description,
        thumbnailUrl: info.thumbnailUrl,
      ),
    );
  }
}

/// The creator's chat commands, by scope.
final myChatCommandsProvider = FutureProvider.autoDispose<MyChatCommands>(
  (ref) async =>
      unwrapData(await ref.watch(chatCommandAdminRepositoryProvider).mine()),
  retry: noRetry,
);

/// The creator's own videos, newest first, paged. Invalidate it after an
/// edit, a delete or a finished upload.
final myVodsProvider =
    AsyncNotifierProvider.autoDispose<MyVodsNotifier, PagedFeed<Vod>>(
      MyVodsNotifier.new,
      retry: noRetry,
    );

/// Loads the pages of `GET /vods/author`.
class MyVodsNotifier extends PagedFeedNotifier<Vod> {
  @override
  int get pageSize => myVodsPageSize;

  @override
  String idOf(Vod item) => item.id;

  @override
  Future<ApiResponse<List<Vod>>> fetchPage(int page) =>
      ref.read(myVodsRepositoryProvider).list(page);
}

/// Lets the user choose a video from the device; null when they cancel.
/// Throws [PickerException] when the picker fails. Widget tests override it
/// with a fake.
typedef VideoPick = Future<PickedVideo?> Function();

/// The system video picker (the photo library). On Android it needs no
/// permission; iOS asks through `NSPhotoLibraryUsageDescription`.
final videoPickProvider = Provider<VideoPick>(
  (ref) => () async {
    try {
      final file = await ImagePicker().pickVideo(source: ImageSource.gallery);
      if (file == null) return null;
      return PickedVideo(
        path: file.path,
        name: file.name,
        size: await file.length(),
      );
    } on PlatformException catch (e) {
      // The picker or the file could not be opened.
      throw PickerException(e);
    }
  },
);
