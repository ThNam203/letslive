/// Riverpod providers of direct messages (web `/messages`).
///
/// Everything is `autoDispose`, so leaving the signed-in shell on logout
/// drops it and the next account never sees the previous one's data. The
/// shell watches [unreadDmCountProvider] (the tab badge) for the whole
/// session, which keeps the unread counts and the presence state live; the
/// conversation list and an open conversation live while their screens do.
///
/// Realtime: [dmEventsProvider] turns the socket's DM pushes into typed
/// [DmEvent]s once; each provider below listens to them and applies what
/// concerns it, like web's `useDmRealtime` does to its query cache. After a
/// reconnect (pushes sent meanwhile were lost) the unread counts, the list
/// and the open conversation are fetched again and merged by message id, so
/// nothing shows twice.
library;

import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart' show StringCharacters;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';
import 'package:letslive_mobile/core/providers.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/data/message_repository.dart';
import 'package:letslive_mobile/features/messages/domain/dm_event.dart';
import 'package:letslive_mobile/features/messages/domain/dm_rules.dart';
import 'package:letslive_mobile/features/search/search_providers.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';

/// Longest preview kept of a new message in the list
/// (`LastMessagePreviewLen` in the chat service, `substring(0, 100)` in web).
const int _previewLength = 100;

/// Conversation REST calls.
final messageRepositoryProvider = Provider<MessageRepository>(
  (ref) => MessageRepository(ref.watch(apiClientProvider)),
);

/// The signed-in user's id; null when logged out.
final currentUserIdProvider = Provider<String?>(
  (ref) => ref.watch(sessionProvider).value?.user.id,
);

/// The DM pushes of the socket as typed events (plus [DmReconnected]).
/// Listeners that are not attached miss events, which is intended: a
/// provider that is gone has nothing to update.
final dmEventsProvider = Provider.autoDispose<Stream<DmEvent>>((ref) {
  final realtime = ref.read(realtimeManagerProvider);
  final controller = StreamController<DmEvent>.broadcast();
  final cleanups = [
    for (final type in DmEventType.all)
      realtime.onEvent(type, (data) {
        final event = parseDmEvent(type, data);
        if (event != null) controller.add(event);
      }),
    realtime.onReconnect(() => controller.add(const DmReconnected())),
  ];
  ref.onDispose(() {
    for (final cleanup in cleanups) {
      cleanup();
    }
    unawaited(controller.close());
  });
  return controller.stream;
});

/// Listens to [dmEventsProvider] until the owner of [ref] is disposed.
void _listenToEvents(Ref ref, void Function(DmEvent event) onEvent) {
  final subscription = ref.watch(dmEventsProvider).listen(onEvent);
  ref.onDispose(subscription.cancel);
}

// --- conversation list -----------------------------------------------------

/// The user's conversations, newest activity first, paged.
final conversationsProvider =
    AsyncNotifierProvider.autoDispose<
      ConversationsNotifier,
      PagedFeed<Conversation>
    >(ConversationsNotifier.new, retry: noRetry);

/// Loads the conversation list and keeps it current from pushes: a new
/// message moves its conversation to the top (web updates it in place; a
/// list of recent chats is expected to bubble up), a message in an unknown
/// conversation (someone just wrote to you) refetches the list.
class ConversationsNotifier extends PagedFeedNotifier<Conversation> {
  @override
  int get pageSize => dmConversationsPageSize;

  @override
  String idOf(Conversation item) => item.id;

  @override
  Future<ApiResponse<List<Conversation>>> fetchPage(int page) =>
      ref.read(messageRepositoryProvider).list(page: page);

  @override
  Future<PagedFeed<Conversation>> build() {
    _listenToEvents(ref, _onEvent);
    return super.build();
  }

  /// Adds a conversation the user just created (nothing if it is already
  /// there: creating an existing DM answers the existing one).
  void prepend(Conversation conversation) {
    final feed = state.value;
    if (feed == null || feed.items.any((c) => c.id == conversation.id)) return;
    final total = feed.total;
    state = AsyncData(
      feed.copyWith(
        items: [conversation, ...feed.items],
        total: total == null ? null : total + 1,
      ),
    );
  }

  /// Fetches page 0 again and merges it into what is loaded: the fresh
  /// items come first, in server order, and the conversations of the
  /// later pages already loaded stay. The loaded pages and the scroll
  /// position survive, unlike a rebuild. A failed fetch changes nothing.
  Future<void> _refreshFirstPage() async {
    final response = await fetchPage(0);
    final fresh = response.data;
    final feed = state.value;
    if (!ref.mounted || !response.success || fresh == null || feed == null) {
      return;
    }
    final freshIds = {for (final c in fresh) c.id};
    state = AsyncData(
      feed.copyWith(
        items: [
          ...fresh,
          for (final c in feed.items)
            if (!freshIds.contains(c.id)) c,
        ],
        total: response.meta?.total,
      ),
    );
  }

  void _onEvent(DmEvent event) {
    final feed = state.value;
    switch (event) {
      case DmReconnected():
        unawaited(_refreshFirstPage());
      case DmNewMessage(:final conversationId, :final message):
        if (feed == null) return;
        final index = feed.items.indexWhere((c) => c.id == conversationId);
        if (index < 0) return unawaited(_refreshFirstPage());
        final updated = feed.items[index].withLastMessage(
          LastMessage(
            id: message.id,
            senderId: message.senderId,
            senderUsername: message.senderUsername,
            text: message.text.characters.take(_previewLength).toString(),
            createdAt: message.createdAt,
          ),
        );
        state = AsyncData(
          feed.copyWith(
            items: [
              updated,
              for (var i = 0; i < feed.items.length; i++)
                if (i != index) feed.items[i],
            ],
          ),
        );
      case DmConversationUpdated(:final conversationId, :final update):
        if (feed == null) return;
        try {
          state = AsyncData(
            feed.copyWith(
              items: [
                for (final c in feed.items)
                  c.id == conversationId ? c.withUpdate(update) : c,
              ],
            ),
          );
        } on FormatException {
          // A payload this app does not understand; web ignores it too.
        }
      case DmMessageEdited() ||
          DmMessageDeleted() ||
          DmTyping() ||
          DmPresence():
        break;
    }
  }
}

/// One conversation's details, or null when it does not exist or the user is
/// not a member (404 / 403): the screen's not-found state.
final conversationProvider = FutureProvider.autoDispose
    .family<Conversation?, String>((ref, id) async {
      final response = await ref.watch(messageRepositoryProvider).byId(id);
      if (response.statusCode == 404 || response.statusCode == 403) {
        return null;
      }
      return unwrapData(response);
    }, retry: noRetry);

// --- unread ------------------------------------------------------------------

/// Unread messages per conversation id.
final dmUnreadProvider =
    AsyncNotifierProvider.autoDispose<DmUnreadNotifier, Map<String, int>>(
      DmUnreadNotifier.new,
      retry: noRetry,
    );

/// Loads the unread counts and counts new messages from pushes. A message is
/// not unread when the user sent it or its conversation is open (its
/// [dmThreadProvider] exists), like web's "visible conversation" check.
class DmUnreadNotifier extends AsyncNotifier<Map<String, int>> {
  @override
  Future<Map<String, int>> build() async {
    _listenToEvents(ref, _onEvent);
    final response = await ref.read(messageRepositoryProvider).unreadCounts();
    if (!response.success) throw ApiLoadException(response);
    // Go sends null for an empty map. A conversation that is open is being
    // read: a count fetched now (cold start, reconnect) may predate its
    // mark-read, so it must not bring the count back.
    return {
      for (final e in (response.data ?? const <String, int>{}).entries)
        if (!ref.exists(dmThreadProvider(e.key))) e.key: e.value,
    };
  }

  /// The user read [conversationId]: its count goes away.
  void clear(String conversationId) {
    final counts = state.value;
    if (counts == null || !counts.containsKey(conversationId)) return;
    state = AsyncData({
      for (final e in counts.entries)
        if (e.key != conversationId) e.key: e.value,
    });
  }

  void _onEvent(DmEvent event) {
    switch (event) {
      case DmReconnected():
        ref.invalidateSelf();
      case DmNewMessage(:final conversationId, :final message):
        final counts = state.value;
        final isOwn = message.senderId == ref.read(currentUserIdProvider);
        final isOpen = ref.exists(dmThreadProvider(conversationId));
        if (counts == null || isOwn || isOpen) return;
        state = AsyncData({
          ...counts,
          conversationId: (counts[conversationId] ?? 0) + 1,
        });
      case DmMessageEdited() ||
          DmMessageDeleted() ||
          DmTyping() ||
          DmPresence() ||
          DmConversationUpdated():
        break;
    }
  }
}

/// Total unread messages: the Messages tab badge. The shell watches it for
/// the whole session, which also keeps the typing and online state alive
/// (it is listened to below), as web mounts all of it once in its header.
final unreadDmCountProvider = Provider.autoDispose<int>((ref) {
  ref.listen(dmPresenceProvider, (_, _) {});
  final counts = ref.watch(dmUnreadProvider).value;
  return counts == null ? 0 : counts.values.fold(0, (sum, n) => sum + n);
});

// --- typing and online ---------------------------------------------------------

/// Who is typing where, and who is online.
@immutable
class DmPresenceState {
  const DmPresenceState({this.typing = const {}, this.online = const {}});

  /// Usernames typing, per conversation id.
  final Map<String, List<String>> typing;

  /// Ids of users known to be online. Like web, only users seen coming
  /// online since the app opened are known.
  final Set<String> online;
}

/// Typing and online state from pushes.
final dmPresenceProvider =
    NotifierProvider.autoDispose<DmPresenceNotifier, DmPresenceState>(
      DmPresenceNotifier.new,
    );

/// Applies typing and presence pushes. A "typing" hint expires after
/// [dmTypingTimeout] unless refreshed; a "stopped typing" push clears it.
class DmPresenceNotifier extends Notifier<DmPresenceState> {
  final Map<String, Timer> _timers = {};

  @override
  DmPresenceState build() {
    _listenToEvents(ref, _onEvent);
    ref.onDispose(() {
      for (final timer in _timers.values) {
        timer.cancel();
      }
      _timers.clear();
    });
    return const DmPresenceState();
  }

  void _onEvent(DmEvent event) {
    switch (event) {
      case DmTyping(:final conversationId, :final username, :final isTyping):
        _setTyping(conversationId, username, isTyping: isTyping);
      case DmPresence(:final userId, :final isOnline):
        state = DmPresenceState(
          typing: state.typing,
          online: isOnline
              ? {...state.online, userId}
              : {
                  for (final id in state.online)
                    if (id != userId) id,
                },
        );
      case DmNewMessage() ||
          DmMessageEdited() ||
          DmMessageDeleted() ||
          DmConversationUpdated() ||
          DmReconnected():
        break;
    }
  }

  void _setTyping(
    String conversationId,
    String username, {
    required bool isTyping,
  }) {
    final key = '$conversationId:$username';
    _timers.remove(key)?.cancel();
    final names = state.typing[conversationId] ?? const [];
    final next = isTyping
        ? (names.contains(username) ? names : [...names, username])
        : [
            for (final n in names)
              if (n != username) n,
          ];
    if (isTyping) {
      _timers[key] = Timer(
        dmTypingTimeout,
        () => _setTyping(conversationId, username, isTyping: false),
      );
    }
    state = DmPresenceState(
      typing: {
        for (final e in state.typing.entries)
          if (e.key != conversationId) e.key: e.value,
        if (next.isNotEmpty) conversationId: next,
      },
      online: state.online,
    );
  }
}

// --- one conversation's messages --------------------------------------------

/// The messages of one open conversation.
@immutable
class DmThread {
  const DmThread({
    required this.messages,
    required this.hasMore,
    this.isLoadingOlder = false,
    this.olderError,
  });

  /// Loaded messages, oldest first, without duplicates.
  final List<DmMessage> messages;

  /// False once the beginning of the conversation is loaded.
  final bool hasMore;
  final bool isLoadingOlder;

  /// The last request for older messages failed; the screen offers a retry.
  final ApiLoadException? olderError;

  DmThread copyWith({
    List<DmMessage>? messages,
    bool? hasMore,
    bool? isLoadingOlder,
    ApiLoadException? olderError,
    bool clearError = false,
  }) => DmThread(
    messages: messages ?? this.messages,
    hasMore: hasMore ?? this.hasMore,
    isLoadingOlder: isLoadingOlder ?? this.isLoadingOlder,
    olderError: clearError ? null : olderError ?? this.olderError,
  );
}

/// The open conversation's messages. Its existence means the conversation is
/// on screen: opening it marks it read, and new messages in it are read too.
final dmThreadProvider = AsyncNotifierProvider.autoDispose
    .family<DmThreadNotifier, DmThread, String>(
      DmThreadNotifier.new,
      retry: noRetry,
    );

/// Loads the newest page, pages back with a cursor ([loadOlder]), applies
/// pushes and sends, edits and deletes the user's messages.
class DmThreadNotifier extends AsyncNotifier<DmThread> {
  DmThreadNotifier(this.conversationId);

  /// The conversation shown.
  final String conversationId;

  // Pushes that arrive before the first page has loaded.
  final List<DmMessage> _early = [];

  // Bumped when the loaded messages are replaced, so a page that was
  // requested for the old ones is dropped.
  int _epoch = 0;

  MessageRepository get _repo => ref.read(messageRepositoryProvider);

  @override
  Future<DmThread> build() async {
    _early.clear();
    // Subscribe before fetching so nothing sent in between is lost.
    _listenToEvents(ref, _onEvent);
    // Not now: build must not change another provider.
    unawaited(Future.microtask(_markRead));
    final response = await _repo.messages(conversationId);
    final page = unwrapList(response);
    final messages = mergeDmMessages(page, _early);
    _early.clear();
    return DmThread(
      messages: messages,
      hasMore: page.length >= dmMessagesPageSize,
    );
  }

  /// Loads the page before the oldest loaded message. Does nothing while one
  /// is loading or at the beginning of the conversation; also the retry after
  /// an error.
  Future<void> loadOlder() async {
    if (!ref.mounted) return;
    final thread = state.value;
    if (thread == null ||
        !thread.hasMore ||
        thread.isLoadingOlder ||
        thread.messages.isEmpty) {
      return;
    }
    final epoch = _epoch;
    state = AsyncData(thread.copyWith(isLoadingOlder: true, clearError: true));
    final response = await _repo.messages(
      conversationId,
      before: thread.messages.first.id,
    );
    final current = state.value;
    if (!ref.mounted || epoch != _epoch || current == null) return;
    final page = response.data;
    if (!response.success || page == null) {
      state = AsyncData(
        current.copyWith(
          isLoadingOlder: false,
          olderError: ApiLoadException(response),
        ),
      );
      return;
    }
    state = AsyncData(
      current.copyWith(
        messages: mergeDmMessages(current.messages, page),
        hasMore: page.length >= dmMessagesPageSize,
        isLoadingOlder: false,
      ),
    );
  }

  /// Sends [text]. On success the message is added at once (merged by id
  /// with its own push); a failure is returned for the caller to show.
  Future<ApiResponse<DmMessage>> send(String text) async {
    final response = await _repo.send(conversationId, text);
    final sent = response.data;
    if (response.success && sent != null && ref.mounted) _merge([sent]);
    return response;
  }

  /// Edits the user's own message [messageId]. Returns null, sending
  /// nothing, when the message is not loaded or may not be edited (see
  /// [canEditMessage]); else the server's response.
  Future<ApiResponse<Object?>?> edit(String messageId, String text) async {
    final message = _find(messageId);
    if (message == null || !canEditMessage(message, _myId)) return null;
    final response = await _repo.edit(conversationId, messageId, text);
    final edited = response.data;
    if (response.success && edited != null && ref.mounted) _merge([edited]);
    return response;
  }

  /// Deletes the user's own message [messageId]; same contract as [edit].
  Future<ApiResponse<Object?>?> delete(String messageId) async {
    final message = _find(messageId);
    if (message == null || !canDeleteMessage(message, _myId)) return null;
    final response = await _repo.delete(conversationId, messageId);
    if (response.success && ref.mounted) _merge([message.asDeleted()]);
    return response;
  }

  /// Tells the others the user started or stopped typing. Best effort: a
  /// lost hint is not worth showing an error.
  void setTyping({required bool isTyping}) =>
      unawaited(_repo.typing(conversationId, isTyping: isTyping));

  String? get _myId => ref.read(currentUserIdProvider);

  DmMessage? _find(String id) {
    for (final m in state.value?.messages ?? const <DmMessage>[]) {
      if (m.id == id) return m;
    }
    return null;
  }

  void _onEvent(DmEvent event) {
    switch (event) {
      case DmNewMessage(:final conversationId, :final message)
          when conversationId == this.conversationId:
        _merge([message]);
        if (message.senderId != _myId) unawaited(_markRead());
      case DmMessageEdited(
            :final conversationId,
            :final messageId,
            :final newText,
            :final updatedAt,
          )
          when conversationId == this.conversationId:
        final known = _find(messageId);
        if (known != null) _merge([known.withText(newText, updatedAt)]);
      case DmMessageDeleted(:final conversationId, :final messageId)
          when conversationId == this.conversationId:
        final known = _find(messageId);
        if (known != null) _merge([known.asDeleted()]);
      case DmReconnected():
        unawaited(_catchUp());
      case DmNewMessage() ||
          DmMessageEdited() ||
          DmMessageDeleted() ||
          DmTyping() ||
          DmPresence() ||
          DmConversationUpdated():
        break;
    }
  }

  void _merge(List<DmMessage> messages) {
    final thread = state.value;
    if (thread == null) {
      _early.addAll(messages);
    } else {
      state = AsyncData(
        thread.copyWith(messages: mergeDmMessages(thread.messages, messages)),
      );
    }
  }

  /// After a reconnect: fetch the newest page and merge it. When it shares
  /// no message with what is loaded, more than a page was missed, so it
  /// replaces the list (merging would leave a hole the cursor never fills).
  /// A failed fetch changes nothing; the next reconnect tries again.
  Future<void> _catchUp() async {
    unawaited(_markRead());
    final response = await _repo.messages(conversationId);
    final page = response.data;
    final current = state.value;
    if (!ref.mounted || !response.success || page == null || current == null) {
      return;
    }
    final known = {for (final m in current.messages) m.id};
    final overlaps = page.isEmpty || page.any((m) => known.contains(m.id));
    if (overlaps) {
      _merge(page);
      return;
    }
    _epoch++;
    state = AsyncData(
      DmThread(messages: page, hasMore: page.length >= dmMessagesPageSize),
    );
  }

  /// Clears the unread count at once and tells the server (web does the same
  /// when a conversation opens and whenever its messages change).
  Future<void> _markRead() async {
    if (!ref.mounted) return;
    if (ref.exists(dmUnreadProvider)) {
      ref.read(dmUnreadProvider.notifier).clear(conversationId);
    }
    await _repo.markRead(conversationId);
  }
}

// --- finding people to write to -----------------------------------------------

/// Users whose name matches a query, for the new-conversation screen.
final dmUserSearchProvider = FutureProvider.autoDispose
    .family<List<PublicUser>, String>(
      (ref, query) async =>
          unwrapList(await ref.watch(searchRepositoryProvider).users(q: query)),
      retry: noRetry,
    );
