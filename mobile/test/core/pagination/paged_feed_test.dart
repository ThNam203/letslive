import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:letslive_mobile/core/network/api_load.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/pagination/paged_feed.dart';

ApiResponse<List<String>> _ok(List<String> items) => ApiResponse(
  success: true,
  statusCode: 200,
  code: 0,
  key: '',
  message: '',
  requestId: '',
  data: items,
);

/// A feed of strings whose pages the test answers through [pending].
class _Feed extends PagedFeedNotifier<String> {
  final List<int> requested = [];
  final Map<int, Completer<ApiResponse<List<String>>>> pending = {};

  @override
  int get pageSize => 2;

  @override
  String idOf(String item) => item;

  @override
  Future<ApiResponse<List<String>>> fetchPage(int page) {
    requested.add(page);
    return (pending[page] = Completer()).future;
  }

  /// A realtime push / user action: puts [item] on top.
  void prepend(String item) {
    final feed = state.value!;
    state = AsyncData(feed.copyWith(items: [item, ...feed.items]));
  }
}

final _feedProvider = AsyncNotifierProvider<_Feed, PagedFeed<String>>(
  _Feed.new,
  retry: noRetry,
);

void main() {
  late ProviderContainer container;
  late _Feed notifier;

  setUp(() async {
    container = ProviderContainer();
    addTearDown(container.dispose);
    container.listen(_feedProvider, (_, _) {});
    notifier = container.read(_feedProvider.notifier);
    notifier.pending[0]!.complete(_ok(['a', 'b']));
    await container.read(_feedProvider.future);
  });

  PagedFeed<String> feed() => container.read(_feedProvider).requireValue;

  test(
    'a change while the next page loads keeps both, and paging goes on',
    () async {
      final loading = notifier.loadMore();
      notifier.prepend('new');
      notifier.pending[1]!.complete(_ok(['c', 'd']));
      await loading;

      expect(feed().items, ['new', 'a', 'b', 'c', 'd']);
      expect(feed().isLoadingMore, isFalse);
      expect(feed().nextPage, 2);

      final next = notifier.loadMore();
      expect(notifier.requested, [0, 1, 2]);
      notifier.pending[2]!.complete(_ok(['e']));
      await next;
      expect(feed().items.last, 'e');
      expect(feed().hasMore, isFalse);
    },
  );

  test(
    'a failed page during a change keeps the change and offers retry',
    () async {
      final loading = notifier.loadMore();
      notifier.prepend('new');
      notifier.pending[1]!.complete(
        ApiResponse.transportFailure(ApiFailureKind.network, 'offline'),
      );
      await loading;

      expect(feed().items, ['new', 'a', 'b']);
      expect(feed().isLoadingMore, isFalse);
      expect(feed().loadMoreError, isNotNull);
    },
  );

  test('a page requested before a rebuild is dropped', () async {
    final loading = notifier.loadMore();
    container.invalidate(_feedProvider);
    final rebuilt = container.read(_feedProvider.notifier);
    // Riverpod may reuse or recreate the notifier; either way page 0 is due.
    rebuilt.pending[0]!.complete(_ok(['x', 'y']));
    await container.read(_feedProvider.future);
    notifier.pending[1]!.complete(_ok(['c', 'd']));
    await loading;

    expect(feed().items, ['x', 'y']);
    expect(feed().nextPage, 1);
  });
}
