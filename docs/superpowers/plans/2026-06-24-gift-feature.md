# Gift Feature + Inventory Tab — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a complete gift feature to the Flutter mobile app: send gifts from inventory, receive gifts, view gift history, and show animation on send.

**Architecture:** New `features/gift/` directory mirrors the shop feature's data → presentation split. Gift data layer talks to user-service `/v1/` endpoints. Animation overlay is a standalone stateful widget shown imperatively via `showDialog` (zero-dependency approach). The wallet screen gains a third tab showing owned inventory items. Profile screen gains a gift-button for non-own profiles.

**Tech Stack:** Flutter 3.x, Riverpod (ConsumerStatefulWidget), GoRouter, forui component library, `Image.network` + `FadeTransition` for animation overlay (no new packages).

## Global Constraints

- No new pub.dev dependencies — use only what is already in `pubspec.yaml`.
- No `dynamic` in variable declarations — `Map<String, dynamic>` in `fromJson` factory methods only.
- No `// TODO` comments left in produced code.
- No mock data — all data from real API responses.
- `flutter analyze` must pass with zero NEW errors (one pre-existing error in `emote_picker.dart` is allowed).
- All screens: `ConsumerStatefulWidget` + `ConsumerState<T>` pattern.
- Import paths: always relative, matching existing files' conventions.
- API base URL prefix for user-service routes: paths go through `ApiClient` which already has `AppConfig.apiUrl` as `baseUrl`. Use path strings matching the backend (e.g. `/user/me/inventory`).
- Error handling: on API failure show error string; never silently swallow errors in UI.

---

### Task 1: Data Models — Gift and InventoryItem

**Files:**
- Create: `mobile/lib/features/gift/data/models/gift_model.dart`
- Create: `mobile/lib/features/shop/data/models/inventory_item_model.dart`

**Interfaces:**
- Consumes: nothing (pure models)
- Produces:
  - `Gift` — fields: `id String`, `senderUserId String`, `recipientUserId String`, `shopItemId String`, `quantity int`, `message String?`, `sentAt String`, `shopItemName String?`, `shopItemImageUrl String?`, `senderUsername String?`
  - `InventoryItem` — fields: `id String`, `userId String`, `shopItemId String`, `quantity int`, `updatedAt String`, `shopItemName String?`, `shopItemImageUrl String?`
  - Both have `fromJson(Map<String, dynamic> json)` factory constructors

- [ ] **Step 1: Create `gift_model.dart`**

```dart
// mobile/lib/features/gift/data/models/gift_model.dart

class Gift {
  final String id;
  final String senderUserId;
  final String recipientUserId;
  final String shopItemId;
  final int quantity;
  final String? message;
  final String sentAt;
  final String? shopItemName;
  final String? shopItemImageUrl;
  final String? senderUsername;

  const Gift({
    required this.id,
    required this.senderUserId,
    required this.recipientUserId,
    required this.shopItemId,
    required this.quantity,
    this.message,
    required this.sentAt,
    this.shopItemName,
    this.shopItemImageUrl,
    this.senderUsername,
  });

  factory Gift.fromJson(Map<String, dynamic> json) {
    return Gift(
      id: json['id'] as String,
      senderUserId: json['senderUserId'] as String,
      recipientUserId: json['recipientUserId'] as String,
      shopItemId: json['shopItemId'] as String,
      quantity: (json['quantity'] as num).toInt(),
      message: json['message'] as String?,
      sentAt: json['sentAt'] as String,
      shopItemName: json['shopItemName'] as String?,
      shopItemImageUrl: json['shopItemImageUrl'] as String?,
      senderUsername: json['senderUsername'] as String?,
    );
  }
}
```

- [ ] **Step 2: Create `inventory_item_model.dart`**

```dart
// mobile/lib/features/shop/data/models/inventory_item_model.dart

class InventoryItem {
  final String id;
  final String userId;
  final String shopItemId;
  final int quantity;
  final String updatedAt;
  final String? shopItemName;
  final String? shopItemImageUrl;
  final String? shopItemAnimationUrl;

  const InventoryItem({
    required this.id,
    required this.userId,
    required this.shopItemId,
    required this.quantity,
    required this.updatedAt,
    this.shopItemName,
    this.shopItemImageUrl,
    this.shopItemAnimationUrl,
  });

  factory InventoryItem.fromJson(Map<String, dynamic> json) {
    return InventoryItem(
      id: json['id'] as String,
      userId: json['userId'] as String,
      shopItemId: json['shopItemId'] as String,
      quantity: (json['quantity'] as num).toInt(),
      updatedAt: json['updatedAt'] as String,
      shopItemName: json['shopItemName'] as String?,
      shopItemImageUrl: json['shopItemImageUrl'] as String?,
      shopItemAnimationUrl: json['shopItemAnimationUrl'] as String?,
    );
  }
}
```

- [ ] **Step 3: Verify no analyze errors**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/data/models/ lib/features/shop/data/models/
```

Expected: no errors (only the pre-existing emote_picker.dart error is allowed).

- [ ] **Step 4: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/gift/data/models/gift_model.dart \
        lib/features/shop/data/models/inventory_item_model.dart
git commit -m "feat(mobile): add Gift and InventoryItem data models

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 2: API Endpoints + Repositories

**Files:**
- Modify: `mobile/lib/core/network/api_endpoints.dart`
- Create: `mobile/lib/features/gift/data/repositories/gift_repository.dart`
- Create: `mobile/lib/features/shop/data/repositories/inventory_repository.dart`
- Modify: `mobile/lib/providers.dart`

**Interfaces:**
- Consumes: `Gift` from Task 1, `InventoryItem` from Task 1, `ApiClient`, `ApiEndpoints`, `ApiResponse`
- Produces:
  - `GiftRepository` with methods:
    - `sendFromInventory({required String shopItemId, required String recipientUserId, String? message}) → Future<ApiResponse<Gift>>`
    - `getReceivedGifts(String userId, {int page = 0, int limit = 20}) → Future<ApiResponse<List<Gift>>>`
    - `getSentGifts({int page = 0, int limit = 20}) → Future<ApiResponse<List<Gift>>>`
  - `InventoryRepository` with methods:
    - `getMyInventory({int page = 0, int limit = 20}) → Future<ApiResponse<List<InventoryItem>>>`
  - `giftRepositoryProvider` Riverpod provider
  - `inventoryRepositoryProvider` Riverpod provider
  - New `ApiEndpoints` constants: `userGifts`, `userGiftsReceived(String userId)`, `userGiftsSent`, `userInventory`

- [ ] **Step 1: Add endpoints to `api_endpoints.dart`**

Open `/Users/ictsaigon.vn/mywork/letslive/mobile/lib/core/network/api_endpoints.dart` and add after the existing user endpoints section (after line 27 `static const userApiKey = '/user/me/api-key';`):

```dart
  // Gifts
  static const userGifts = '/user/v1/gifts';
  static String userGiftsReceived(String userId) =>
      '/user/v1/user/$userId/gifts/received';
  static const userGiftsSent = '/user/v1/user/me/gifts/sent';
  static const userInventory = '/user/v1/user/me/inventory';
```

- [ ] **Step 2: Create `gift_repository.dart`**

```dart
// mobile/lib/features/gift/data/repositories/gift_repository.dart

import '../../../../core/network/api_client.dart';
import '../../../../core/network/api_endpoints.dart';
import '../../../../core/network/api_response.dart';
import '../models/gift_model.dart';

class GiftRepository {
  final ApiClient _client;

  GiftRepository(this._client);

  Future<ApiResponse<Gift>> sendFromInventory({
    required String shopItemId,
    required String recipientUserId,
    String? message,
  }) {
    final data = <String, dynamic>{
      'shop_item_id': shopItemId,
      'recipient_user_id': recipientUserId,
    };
    if (message != null && message.isNotEmpty) {
      data['message'] = message;
    }
    return _client.post(
      ApiEndpoints.userGifts,
      data: data,
      fromJsonT: (json) => Gift.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<ApiResponse<List<Gift>>> getReceivedGifts(
    String userId, {
    int page = 0,
    int limit = 20,
  }) {
    return _client.get(
      ApiEndpoints.userGiftsReceived(userId),
      queryParameters: {'page': page, 'limit': limit},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Gift.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<ApiResponse<List<Gift>>> getSentGifts({
    int page = 0,
    int limit = 20,
  }) {
    return _client.get(
      ApiEndpoints.userGiftsSent,
      queryParameters: {'page': page, 'limit': limit},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => Gift.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
```

- [ ] **Step 3: Create `inventory_repository.dart`**

```dart
// mobile/lib/features/shop/data/repositories/inventory_repository.dart

import '../../../../core/network/api_client.dart';
import '../../../../core/network/api_endpoints.dart';
import '../../../../core/network/api_response.dart';
import '../models/inventory_item_model.dart';

class InventoryRepository {
  final ApiClient _client;

  InventoryRepository(this._client);

  Future<ApiResponse<List<InventoryItem>>> getMyInventory({
    int page = 0,
    int limit = 20,
  }) {
    return _client.get(
      ApiEndpoints.userInventory,
      queryParameters: {'page': page, 'limit': limit},
      fromJsonT: (json) => (json as List<dynamic>)
          .map((e) => InventoryItem.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
```

- [ ] **Step 4: Register providers in `providers.dart`**

Add these imports at the top of `/Users/ictsaigon.vn/mywork/letslive/mobile/lib/providers.dart`, after the existing imports:

```dart
import 'features/gift/data/repositories/gift_repository.dart';
import 'features/shop/data/repositories/inventory_repository.dart';
```

Add these providers at the end of the file, before the closing:

```dart
/// Gift repository.
final giftRepositoryProvider = Provider<GiftRepository>((ref) {
  return GiftRepository(ref.watch(apiClientProvider));
});

/// Inventory repository.
final inventoryRepositoryProvider = Provider<InventoryRepository>((ref) {
  return InventoryRepository(ref.watch(apiClientProvider));
});
```

- [ ] **Step 5: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/ lib/features/shop/data/repositories/ lib/providers.dart lib/core/network/api_endpoints.dart
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/core/network/api_endpoints.dart \
        lib/features/gift/data/repositories/gift_repository.dart \
        lib/features/shop/data/repositories/inventory_repository.dart \
        lib/providers.dart
git commit -m "feat(mobile): add GiftRepository and InventoryRepository with providers

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 3: Gift Animation Overlay

**Files:**
- Create: `mobile/lib/features/gift/presentation/widgets/gift_animation_overlay.dart`

**Interfaces:**
- Consumes: nothing external
- Produces: `showGiftAnimationOverlay(BuildContext context, String animationUrl) → Future<void>` — a top-level function that shows the overlay and awaits its dismissal

The overlay is shown via `showDialog` with `barrierDismissible: true` and `barrierColor: Colors.black54`. It uses `Image.network` inside a `FadeTransition` driven by an `AnimationController` (2-second forward, then 1-second reverse, then pop). Tapping outside or tapping the overlay also dismisses it.

- [ ] **Step 1: Create `gift_animation_overlay.dart`**

```dart
// mobile/lib/features/gift/presentation/widgets/gift_animation_overlay.dart

import 'package:flutter/material.dart';

/// Shows a full-screen gift animation overlay and resolves when dismissed.
///
/// [animationUrl] is the URL of a GIF or image that represents the animation.
/// The overlay fades in, displays the image for 2 seconds, fades out,
/// then auto-dismisses. Tapping the screen dismisses it immediately.
Future<void> showGiftAnimationOverlay(
  BuildContext context,
  String animationUrl,
) {
  return showDialog<void>(
    context: context,
    barrierDismissible: true,
    barrierColor: Colors.black54,
    builder: (dialogContext) =>
        _GiftAnimationDialog(animationUrl: animationUrl),
  );
}

class _GiftAnimationDialog extends StatefulWidget {
  final String animationUrl;

  const _GiftAnimationDialog({required this.animationUrl});

  @override
  State<_GiftAnimationDialog> createState() => _GiftAnimationDialogState();
}

class _GiftAnimationDialogState extends State<_GiftAnimationDialog>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _opacity;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
      reverseDuration: const Duration(milliseconds: 400),
    );
    _opacity = CurvedAnimation(parent: _controller, curve: Curves.easeIn);

    _runAnimation();
  }

  Future<void> _runAnimation() async {
    await _controller.forward();
    // Hold for 2 seconds, then fade out and close.
    await Future<void>.delayed(const Duration(seconds: 2));
    if (!mounted) return;
    await _controller.reverse();
    if (mounted) Navigator.of(context).pop();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => Navigator.of(context).pop(),
      behavior: HitTestBehavior.opaque,
      child: Center(
        child: FadeTransition(
          opacity: _opacity,
          child: ConstrainedBox(
            constraints: const BoxConstraints(
              maxWidth: 320,
              maxHeight: 320,
            ),
            child: Image.network(
              widget.animationUrl,
              fit: BoxFit.contain,
              loadingBuilder: (context, child, progress) {
                if (progress == null) return child;
                return const Center(child: CircularProgressIndicator());
              },
              errorBuilder: (context, error, stackTrace) {
                return const Center(
                  child: Icon(Icons.card_giftcard, size: 80, color: Colors.white),
                );
              },
            ),
          ),
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/presentation/widgets/gift_animation_overlay.dart
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/gift/presentation/widgets/gift_animation_overlay.dart
git commit -m "feat(mobile): add gift animation overlay (FadeTransition + Image.network)

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 4: Wire Animation Into shop_item_detail_sheet.dart

**Files:**
- Modify: `mobile/lib/features/shop/presentation/screens/shop_item_detail_sheet.dart`

**Interfaces:**
- Consumes: `showGiftAnimationOverlay` from Task 3 (`../../../gift/presentation/widgets/gift_animation_overlay.dart` relative path from the shop file)
- Produces: after a successful `_purchase()` call, if `result.data!.animationUrl.isNotEmpty`, calls `showGiftAnimationOverlay(context, animationUrl)` before popping.

Currently `_purchase()` does: success → set message → delay 1s → pop. Change it to: success → show animation overlay (await) → pop. Remove the `_successMessage` state and the delay since the overlay replaces that feedback.

- [ ] **Step 1: Add import to `shop_item_detail_sheet.dart`**

Open `/Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/shop/presentation/screens/shop_item_detail_sheet.dart`.

Add after the existing imports (after `import '../../data/models/shop_item_model.dart';`):

```dart
import '../../../gift/presentation/widgets/gift_animation_overlay.dart';
```

- [ ] **Step 2: Replace `_purchase()` method**

Find the existing `_purchase()` method in `_ShopItemDetailSheetState` and replace it with:

```dart
  Future<void> _purchase({String? recipientUserId}) async {
    setState(() {
      _isBuying = true;
      _errorMessage = null;
      _successMessage = null;
    });

    try {
      final repo = ref.read(shopRepositoryProvider);
      final result = await repo.purchase(
        shopItemId: widget.item.id,
        quantity: _quantity,
        recipientUserId: recipientUserId,
      );

      if (!mounted) return;

      if (result.success && result.data != null) {
        final animationUrl = result.data!.animationUrl;
        if (animationUrl.isNotEmpty && mounted) {
          await showGiftAnimationOverlay(context, animationUrl);
        }
        if (mounted) Navigator.of(context).pop();
      } else {
        setState(() {
          _errorMessage = result.message.isNotEmpty
              ? result.message
              : 'Purchase failed, please try again.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _errorMessage = 'Purchase failed, please try again.';
        });
      }
    } finally {
      if (mounted) setState(() => _isBuying = false);
    }
  }
```

- [ ] **Step 3: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/shop/presentation/screens/shop_item_detail_sheet.dart
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/shop/presentation/screens/shop_item_detail_sheet.dart
git commit -m "feat(mobile): show gift animation overlay after purchase

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 5: Gifts Received Screen

**Files:**
- Create: `mobile/lib/features/gift/presentation/screens/gifts_received_screen.dart`
- Modify: `mobile/lib/core/router/app_router.dart`

**Interfaces:**
- Consumes: `GiftRepository.getReceivedGifts(userId)` from Task 2, `giftRepositoryProvider` from Task 2, `Gift` from Task 1
- Produces: `GiftsReceivedScreen(userId: String)` widget, route `/users/:userId/gifts` added to `AppRoutes` and `appRouter`

The screen pattern mirrors `NotificationsScreen`: `ConsumerStatefulWidget`, `initState` calls fetch, paginated list via `_currentPage`/`_hasMore`, `RefreshIndicator`, `FScaffold` + `FHeader`.

Each list item shows: gift icon, item name (`shopItemName ?? shopItemId`), sender (`senderUsername ?? senderUserId`), quantity, date (formatted like `_formatDate` in profile_screen).

- [ ] **Step 1: Create `gifts_received_screen.dart`**

```dart
// mobile/lib/features/gift/presentation/screens/gifts_received_screen.dart

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../providers.dart';
import '../../../../shared/widgets/empty_state_view.dart';
import '../../../../shared/widgets/error_display.dart';
import '../../../../shared/widgets/loading_indicator.dart';
import '../../../gift/data/models/gift_model.dart';

class GiftsReceivedScreen extends ConsumerStatefulWidget {
  final String userId;

  const GiftsReceivedScreen({super.key, required this.userId});

  @override
  ConsumerState<GiftsReceivedScreen> createState() =>
      _GiftsReceivedScreenState();
}

class _GiftsReceivedScreenState extends ConsumerState<GiftsReceivedScreen> {
  List<Gift> _gifts = [];
  bool _isLoading = true;
  String? _error;
  int _currentPage = 0;
  bool _hasMore = true;
  bool _isLoadingMore = false;

  @override
  void initState() {
    super.initState();
    _fetchGifts();
  }

  Future<void> _fetchGifts() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(giftRepositoryProvider);
      final response = await repo.getReceivedGifts(
        widget.userId,
        page: 0,
      );
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _gifts = response.data ?? [];
          _isLoading = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() {
          _error = response.message.isNotEmpty
              ? response.message
              : 'Failed to load gifts.';
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Network error. Please try again.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(giftRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response = await repo.getReceivedGifts(
        widget.userId,
        page: nextPage,
      );
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _currentPage = nextPage;
          _gifts.addAll(response.data ?? []);
          _isLoadingMore = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  String _formatDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      const months = [
        'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
      ];
      return '${months[date.month - 1]} ${date.day}, ${date.year}';
    } catch (_) {
      return dateStr;
    }
  }

  @override
  Widget build(BuildContext context) {
    return FScaffold(
      header: const FHeader(title: Text('Gifts Received')),
      child: _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    if (_isLoading) {
      return LoadingIndicator(message: 'Loading gifts...');
    }

    if (_error != null) {
      return ErrorDisplay(
        title: 'Error',
        message: _error,
        onRetry: _fetchGifts,
      );
    }

    if (_gifts.isEmpty) {
      return const EmptyStateView(
        icon: FIcons.gift,
        title: 'No gifts yet',
        description: 'Gifts received will appear here.',
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchGifts,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _gifts.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _gifts.length) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.outline,
                onPress: _isLoadingMore ? null : _loadMore,
                child: _isLoadingMore
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Load more'),
              ),
            );
          }

          final gift = _gifts[index];
          return _GiftTile(gift: gift, formattedDate: _formatDate(gift.sentAt));
        },
      ),
    );
  }
}

class _GiftTile extends StatelessWidget {
  final Gift gift;
  final String formattedDate;

  const _GiftTile({required this.gift, required this.formattedDate});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: colors.border, width: 0.5)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: colors.primary.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(FIcons.gift, size: 18, color: colors.primary),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  gift.shopItemName ?? gift.shopItemId,
                  style: typography.sm
                      .copyWith(fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  'From: ${gift.senderUsername ?? gift.senderUserId}'
                  ' · Qty: ${gift.quantity}',
                  style: typography.xs
                      .copyWith(color: colors.mutedForeground),
                ),
                const SizedBox(height: 2),
                Text(
                  formattedDate,
                  style: typography.xs
                      .copyWith(color: colors.mutedForeground),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Add route to `app_router.dart`**

In `/Users/ictsaigon.vn/mywork/letslive/mobile/lib/core/router/app_router.dart`:

Add import after existing imports (e.g., after `import '../../features/wallet/presentation/wallet_deposit_screen.dart';`):

```dart
import '../../features/gift/presentation/screens/gifts_received_screen.dart';
```

Add to `AppRoutes`:
```dart
  static String userGiftsReceived(String userId) => '/users/$userId/gifts';
```

Add route to `appRouter` (after the `/users/:userId` route block):

```dart
    // Gifts received (public profile page)
    GoRoute(
      path: '/users/:userId/gifts',
      builder: (context, state) {
        final userId = state.pathParameters['userId']!;
        return GiftsReceivedScreen(userId: userId);
      },
    ),
```

- [ ] **Step 3: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/presentation/screens/gifts_received_screen.dart lib/core/router/app_router.dart
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/gift/presentation/screens/gifts_received_screen.dart \
        lib/core/router/app_router.dart
git commit -m "feat(mobile): add gifts received screen and route /users/:id/gifts

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 6: Sent Gifts History Screen

**Files:**
- Create: `mobile/lib/features/gift/presentation/screens/sent_gifts_screen.dart`
- Modify: `mobile/lib/core/router/app_router.dart`

**Interfaces:**
- Consumes: `GiftRepository.getSentGifts()` from Task 2, `giftRepositoryProvider` from Task 2, `Gift` from Task 1
- Produces: `SentGiftsScreen` widget, route `/gifts/sent` (auth-required) added to `AppRoutes` and `appRouter`

Mirrors `GiftsReceivedScreen` but calls `getSentGifts()`. Shows recipient instead of sender in each tile.

- [ ] **Step 1: Create `sent_gifts_screen.dart`**

```dart
// mobile/lib/features/gift/presentation/screens/sent_gifts_screen.dart

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../providers.dart';
import '../../../../shared/widgets/empty_state_view.dart';
import '../../../../shared/widgets/error_display.dart';
import '../../../../shared/widgets/loading_indicator.dart';
import '../../data/models/gift_model.dart';

class SentGiftsScreen extends ConsumerStatefulWidget {
  const SentGiftsScreen({super.key});

  @override
  ConsumerState<SentGiftsScreen> createState() => _SentGiftsScreenState();
}

class _SentGiftsScreenState extends ConsumerState<SentGiftsScreen> {
  List<Gift> _gifts = [];
  bool _isLoading = true;
  String? _error;
  int _currentPage = 0;
  bool _hasMore = true;
  bool _isLoadingMore = false;

  @override
  void initState() {
    super.initState();
    _fetchGifts();
  }

  Future<void> _fetchGifts() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(giftRepositoryProvider);
      final response = await repo.getSentGifts(page: 0);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _gifts = response.data ?? [];
          _isLoading = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() {
          _error = response.message.isNotEmpty
              ? response.message
              : 'Failed to load sent gifts.';
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Network error. Please try again.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(giftRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response = await repo.getSentGifts(page: nextPage);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _currentPage = nextPage;
          _gifts.addAll(response.data ?? []);
          _isLoadingMore = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  String _formatDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      const months = [
        'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
      ];
      return '${months[date.month - 1]} ${date.day}, ${date.year}';
    } catch (_) {
      return dateStr;
    }
  }

  @override
  Widget build(BuildContext context) {
    return FScaffold(
      header: const FHeader(title: Text('Sent Gifts')),
      child: _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    if (_isLoading) {
      return LoadingIndicator(message: 'Loading sent gifts...');
    }

    if (_error != null) {
      return ErrorDisplay(
        title: 'Error',
        message: _error,
        onRetry: _fetchGifts,
      );
    }

    if (_gifts.isEmpty) {
      return const EmptyStateView(
        icon: FIcons.gift,
        title: 'No gifts sent',
        description: 'Gifts you send will appear here.',
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchGifts,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _gifts.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _gifts.length) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.outline,
                onPress: _isLoadingMore ? null : _loadMore,
                child: _isLoadingMore
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Load more'),
              ),
            );
          }

          final gift = _gifts[index];
          return _SentGiftTile(
            gift: gift,
            formattedDate: _formatDate(gift.sentAt),
          );
        },
      ),
    );
  }
}

class _SentGiftTile extends StatelessWidget {
  final Gift gift;
  final String formattedDate;

  const _SentGiftTile({required this.gift, required this.formattedDate});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: colors.border, width: 0.5)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: colors.primary.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(FIcons.gift, size: 18, color: colors.primary),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  gift.shopItemName ?? gift.shopItemId,
                  style: typography.sm
                      .copyWith(fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 2),
                Text(
                  'To: ${gift.recipientUserId}'
                  ' · Qty: ${gift.quantity}',
                  style: typography.xs
                      .copyWith(color: colors.mutedForeground),
                ),
                const SizedBox(height: 2),
                Text(
                  formattedDate,
                  style: typography.xs
                      .copyWith(color: colors.mutedForeground),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Add route to `app_router.dart`**

Add import:
```dart
import '../../features/gift/presentation/screens/sent_gifts_screen.dart';
```

Add to `AppRoutes`:
```dart
  static const sentGifts = '/gifts/sent';
```

Add route (after the `/users/:userId/gifts` route):
```dart
    // Sent gifts history (auth required)
    GoRoute(
      path: AppRoutes.sentGifts,
      redirect: _requireAuth,
      builder: (context, state) => const SentGiftsScreen(),
    ),
```

- [ ] **Step 3: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/presentation/screens/sent_gifts_screen.dart lib/core/router/app_router.dart
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/gift/presentation/screens/sent_gifts_screen.dart \
        lib/core/router/app_router.dart
git commit -m "feat(mobile): add sent gifts history screen and /gifts/sent route

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 7: Inventory Tab in Wallet Screen

**Files:**
- Create: `mobile/lib/features/wallet/presentation/inventory_tab.dart`
- Modify: `mobile/lib/features/wallet/presentation/wallet_screen.dart`

**Interfaces:**
- Consumes: `InventoryRepository.getMyInventory()` from Task 2, `inventoryRepositoryProvider` from Task 2, `InventoryItem` from Task 1, `showShopItemDetail` from `shop_item_detail_sheet.dart` (to allow viewing item)
- Produces: `InventoryTab` widget (stateful, fetches own inventory on mount), `WalletScreen` modified to show three tabs: Overview (existing content), Inventory (new), Transactions (existing link)

**Implementation note:** The existing `WalletScreen` is a `ListView`-based single-page layout. Wrap its body in a `DefaultTabController(length: 2)` approach, showing two tabs: "Overview" (existing wallet content) and "Inventory" (new tab). Use `TabBar` inside the `FHeader`'s `bottom` slot if supported, or below the header — check forui docs. If `FHeader` doesn't have a `bottom` slot, place the `TabBar` as the first item in the `ListView`.

Since the forui `FHeader` widget may not support a `bottom` slot, use this pattern: replace `FScaffold` with a `Scaffold` + `SafeArea` wrapping a `Column` with `TabBar` + `Expanded(child: TabBarView(...))`, or keep `FScaffold` and put a `Column` as the child.

The simplest approach that matches existing code: keep `FScaffold`, make child a `Column` with a `TabBar` + `Expanded(child: TabBarView(...))`, wrapped in a `DefaultTabController`.

- [ ] **Step 1: Create `inventory_tab.dart`**

```dart
// mobile/lib/features/wallet/presentation/inventory_tab.dart

import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../providers.dart';
import '../../../shared/widgets/empty_state_view.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../shop/data/models/inventory_item_model.dart';
import '../../shop/data/models/shop_item_model.dart';
import '../../shop/presentation/screens/shop_item_detail_sheet.dart';

class InventoryTab extends ConsumerStatefulWidget {
  const InventoryTab({super.key});

  @override
  ConsumerState<InventoryTab> createState() => _InventoryTabState();
}

class _InventoryTabState extends ConsumerState<InventoryTab> {
  List<InventoryItem> _items = [];
  bool _isLoading = true;
  String? _error;
  int _currentPage = 0;
  bool _hasMore = true;
  bool _isLoadingMore = false;

  @override
  void initState() {
    super.initState();
    _fetchInventory();
  }

  Future<void> _fetchInventory() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _currentPage = 0;
    });

    try {
      final repo = ref.read(inventoryRepositoryProvider);
      final response = await repo.getMyInventory(page: 0);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _items = response.data ?? [];
          _isLoading = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() {
          _error = response.message.isNotEmpty
              ? response.message
              : 'Failed to load inventory.';
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Network error. Please try again.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loadMore() async {
    if (_isLoadingMore || !_hasMore) return;
    setState(() => _isLoadingMore = true);

    try {
      final repo = ref.read(inventoryRepositoryProvider);
      final nextPage = _currentPage + 1;
      final response = await repo.getMyInventory(page: nextPage);
      if (!mounted) return;

      if (response.success) {
        final pageSize = response.meta?.pageSize ?? 20;
        final total = response.meta?.total ?? 0;
        final totalPages = total == 0 ? 1 : (total + pageSize - 1) ~/ pageSize;
        setState(() {
          _currentPage = nextPage;
          _items.addAll(response.data ?? []);
          _isLoadingMore = false;
          _hasMore = (_currentPage + 1) < totalPages;
        });
      } else {
        setState(() => _isLoadingMore = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isLoadingMore = false);
    }
  }

  void _openItemDetail(InventoryItem item) {
    if (item.shopItemImageUrl == null) return;
    final shopItem = ShopItem(
      id: item.shopItemId,
      name: item.shopItemName ?? item.shopItemId,
      imageUrl: item.shopItemImageUrl!,
      animationUrl: item.shopItemAnimationUrl ?? '',
      price: 0,
      createdAt: item.updatedAt,
    );
    showShopItemDetail(context, shopItem);
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return LoadingIndicator(message: 'Loading inventory...');
    }

    if (_error != null) {
      return ErrorDisplay(
        title: 'Error',
        message: _error,
        onRetry: _fetchInventory,
      );
    }

    if (_items.isEmpty) {
      return const EmptyStateView(
        icon: FIcons.package,
        title: 'Inventory is empty',
        description: 'Items you purchase from the shop will appear here.',
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchInventory,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _items.length + (_hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == _items.length) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: FButton(
                variant: FButtonVariant.outline,
                onPress: _isLoadingMore ? null : _loadMore,
                child: _isLoadingMore
                    ? const SizedBox(
                        height: 16,
                        width: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Load more'),
              ),
            );
          }

          final item = _items[index];
          return _InventoryItemTile(
            item: item,
            onTap: () => _openItemDetail(item),
          );
        },
      ),
    );
  }
}

class _InventoryItemTile extends StatelessWidget {
  final InventoryItem item;
  final VoidCallback onTap;

  const _InventoryItemTile({required this.item, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          border: Border(bottom: BorderSide(color: colors.border, width: 0.5)),
        ),
        child: Row(
          children: [
            // Item image or fallback icon
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: SizedBox(
                width: 48,
                height: 48,
                child: item.shopItemImageUrl != null
                    ? CachedNetworkImage(
                        imageUrl: item.shopItemImageUrl!,
                        fit: BoxFit.cover,
                        placeholder: (_, _) => ColoredBox(color: colors.muted),
                        errorWidget: (_, _, _) => ColoredBox(
                          color: colors.muted,
                          child: Icon(
                            FIcons.package,
                            size: 24,
                            color: colors.mutedForeground,
                          ),
                        ),
                      )
                    : ColoredBox(
                        color: colors.muted,
                        child: Icon(
                          FIcons.package,
                          size: 24,
                          color: colors.mutedForeground,
                        ),
                      ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                item.shopItemName ?? item.shopItemId,
                style: typography.sm.copyWith(fontWeight: FontWeight.w600),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 8),
            // Quantity badge
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: colors.primary.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                'x${item.quantity}',
                style: typography.sm.copyWith(
                  color: colors.primary,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Modify `wallet_screen.dart` to add Inventory tab**

The current `WalletScreen._WalletScreenState` uses `FScaffold`. Wrap the child in a `DefaultTabController`, add a `TabBar` + `TabBarView`. The "Overview" tab shows the existing `ListView` content (balance cards + recent transactions). The "Inventory" tab shows `InventoryTab()`.

Replace the class `_WalletScreenState` `build` method entirely. The `FScaffold.child` becomes:

```dart
child: DefaultTabController(
  length: 2,
  child: Column(
    children: [
      TabBar(
        tabs: const [
          Tab(text: 'Overview'),
          Tab(text: 'Inventory'),
        ],
      ),
      Expanded(
        child: TabBarView(
          children: [
            // Tab 0: existing wallet content
            _isLoading
                ? const Center(child: CircularProgressIndicator())
                : RefreshIndicator(
                    onRefresh: _fetchData,
                    child: ListView(
                      padding: const EdgeInsets.all(16),
                      children: [
                        // balance cards
                        Row(
                          children: [
                            Expanded(
                              child: _BalanceCard(
                                icon: Icons.bolt,
                                name: l10n.walletCurrencySpark,
                                balance: _wallet?.balanceFor(CurrencyCode.spark) ?? '0',
                                gradientColors: [
                                  Colors.amber.shade500,
                                  Colors.orange.shade600,
                                ],
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: _BalanceCard(
                                icon: Icons.diamond,
                                name: l10n.walletCurrencyFlare,
                                balance: _wallet?.balanceFor(CurrencyCode.flare) ?? '0',
                                gradientColors: [
                                  Colors.purple.shade500,
                                  Colors.indigo.shade600,
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 24),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              l10n.walletRecentTransactions,
                              style: typography.lg.copyWith(fontWeight: FontWeight.w600),
                            ),
                            FButton.raw(
                              onPress: () => context.push(AppRoutes.walletTransactions),
                              child: Text(
                                l10n.walletViewAll,
                                style: typography.sm.copyWith(color: colors.primary),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        if (_recentTxns.isEmpty)
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 32),
                            child: Center(
                              child: Text(
                                l10n.walletNoTransactions,
                                style: typography.sm.copyWith(
                                  color: isDark
                                      ? AppColors.darkForegroundMuted
                                      : AppColors.lightForegroundMuted,
                                ),
                              ),
                            ),
                          )
                        else
                          WalletTransactionList(transactions: _recentTxns),
                      ],
                    ),
                  ),
            // Tab 1: inventory
            const InventoryTab(),
          ],
        ),
      ),
    ],
  ),
),
```

Add import at top of `wallet_screen.dart`:
```dart
import 'inventory_tab.dart';
```

The `build` method needs `final typography = context.theme.typography;` and `final isDark = ...` kept since they're used by the overview tab content. Keep `l10n`, `colors`, `typography`, `isDark` declarations.

- [ ] **Step 3: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/wallet/presentation/
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/wallet/presentation/inventory_tab.dart \
        lib/features/wallet/presentation/wallet_screen.dart
git commit -m "feat(mobile): add inventory tab to wallet screen

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 8: Gift Picker Sheet

**Files:**
- Create: `mobile/lib/features/gift/presentation/widgets/gift_picker_sheet.dart`

**Interfaces:**
- Consumes:
  - `InventoryRepository.getMyInventory()` from Task 2, `inventoryRepositoryProvider` from Task 2
  - `GiftRepository.sendFromInventory(...)` from Task 2, `giftRepositoryProvider` from Task 2
  - `InventoryItem` from Task 1
  - `showGiftAnimationOverlay` from Task 3
- Produces: `showGiftPickerSheet(BuildContext context, String recipientUserId) → Future<void>` top-level function

The picker sheet is a `DraggableScrollableSheet` (pattern matches `_UserPickerSheet` in shop_item_detail_sheet). It fetches inventory on mount, shows items as a scrollable list. Tapping an item opens a confirm dialog (`showDialog`) showing: item name, "Send 1 to this user?" with Cancel/Send buttons. On confirm → calls `sendFromInventory` → on success shows `showGiftAnimationOverlay`. If inventory empty shows "Your inventory is empty."

- [ ] **Step 1: Create `gift_picker_sheet.dart`**

```dart
// mobile/lib/features/gift/presentation/widgets/gift_picker_sheet.dart

import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../providers.dart';
import '../../../shop/data/models/inventory_item_model.dart';
import 'gift_animation_overlay.dart';

/// Opens a bottom sheet for picking an inventory item to send as a gift.
Future<void> showGiftPickerSheet(
  BuildContext context,
  String recipientUserId,
) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    clipBehavior: Clip.antiAlias,
    builder: (sheetContext) => ProviderScope.containerOf(context).read(
          // We need ProviderScope access inside the sheet; use a Consumer.
          // Wrapping with Consumer gives access to Riverpod via sheetContext.
          // Instead, build directly:
          _GiftPickerSheet(recipientUserId: recipientUserId),
        ),
  );
}

// ---------------------------------------------------------------------------
// Internal: use a regular showModalBottomSheet but pass ProviderScope down
// ---------------------------------------------------------------------------

/// Opens a bottom sheet for picking an inventory item to send as a gift.
///
/// Shows the caller's Riverpod scope inside the sheet via UncontrolledProviderScope.
Future<void> showGiftPickerSheetScoped(
  BuildContext context,
  String recipientUserId,
) {
  final container = ProviderScope.containerOf(context);
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    clipBehavior: Clip.antiAlias,
    builder: (sheetContext) => UncontrolledProviderScope(
      container: container,
      child: _GiftPickerSheet(recipientUserId: recipientUserId),
    ),
  );
}

class _GiftPickerSheet extends ConsumerStatefulWidget {
  final String recipientUserId;

  const _GiftPickerSheet({required this.recipientUserId});

  @override
  ConsumerState<_GiftPickerSheet> createState() => _GiftPickerSheetState();
}

class _GiftPickerSheetState extends ConsumerState<_GiftPickerSheet> {
  List<InventoryItem> _items = [];
  bool _isLoading = true;
  String? _error;
  bool _isSending = false;
  String? _sendError;

  @override
  void initState() {
    super.initState();
    _fetchInventory();
  }

  Future<void> _fetchInventory() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = ref.read(inventoryRepositoryProvider);
      final response = await repo.getMyInventory();
      if (!mounted) return;

      if (response.success) {
        setState(() {
          _items = response.data ?? [];
          _isLoading = false;
        });
      } else {
        setState(() {
          _error = response.message.isNotEmpty
              ? response.message
              : 'Failed to load inventory.';
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Network error. Please try again.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _sendGift(InventoryItem item) async {
    // Confirm dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Send Gift'),
        content: Text(
          'Send 1 "${item.shopItemName ?? item.shopItemId}" to this user?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Send'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;
    if (!mounted) return;

    setState(() {
      _isSending = true;
      _sendError = null;
    });

    try {
      final giftRepo = ref.read(giftRepositoryProvider);
      final response = await giftRepo.sendFromInventory(
        shopItemId: item.shopItemId,
        recipientUserId: widget.recipientUserId,
      );

      if (!mounted) return;

      if (response.success) {
        // Close the sheet first, then show animation from the parent context.
        Navigator.of(context).pop();
        // NOTE: animationUrl on Gift — use shopItemAnimationUrl from inventory.
        if (item.shopItemAnimationUrl != null &&
            item.shopItemAnimationUrl!.isNotEmpty) {
          // The context here might be disposed after pop. Use rootNavigator key.
          // We delay to allow the sheet to close before overlay.
          await Future<void>.delayed(const Duration(milliseconds: 200));
          if (mounted) {
            await showGiftAnimationOverlay(context, item.shopItemAnimationUrl!);
          }
        }
      } else {
        setState(() {
          _sendError = response.message.isNotEmpty
              ? response.message
              : 'Failed to send gift. Please try again.';
          _isSending = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _sendError = 'Failed to send gift. Please try again.';
          _isSending = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final typography = context.theme.typography;
    final colors = context.theme.colors;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return DraggableScrollableSheet(
      initialChildSize: 0.6,
      minChildSize: 0.4,
      maxChildSize: 0.9,
      expand: false,
      builder: (ctx, scrollController) {
        return Column(
          children: [
            // Drag handle
            Center(
              child: Container(
                margin: const EdgeInsets.only(top: 12, bottom: 8),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: isDark ? Colors.white24 : Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
              child: Text(
                'Send a Gift',
                style: typography.lg.copyWith(fontWeight: FontWeight.bold),
              ),
            ),
            if (_sendError != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: colors.destructive.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _sendError!,
                    style: typography.xs.copyWith(color: colors.destructive),
                  ),
                ),
              ),
            if (_isSending)
              const Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              )
            else
              Expanded(
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : _error != null
                        ? Center(
                            child: Padding(
                              padding: const EdgeInsets.all(24),
                              child: Text(
                                _error!,
                                style: typography.sm.copyWith(
                                  color: colors.mutedForeground,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                          )
                        : _items.isEmpty
                            ? Center(
                                child: Padding(
                                  padding: const EdgeInsets.all(24),
                                  child: Text(
                                    'Your inventory is empty.\nPurchase items from the shop first.',
                                    style: typography.sm.copyWith(
                                      color: colors.mutedForeground,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                              )
                            : ListView.builder(
                                controller: scrollController,
                                itemCount: _items.length,
                                itemBuilder: (ctx, index) {
                                  final item = _items[index];
                                  return ListTile(
                                    leading: ClipRRect(
                                      borderRadius: BorderRadius.circular(6),
                                      child: SizedBox(
                                        width: 40,
                                        height: 40,
                                        child: item.shopItemImageUrl != null
                                            ? CachedNetworkImage(
                                                imageUrl:
                                                    item.shopItemImageUrl!,
                                                fit: BoxFit.cover,
                                                placeholder: (_, _) =>
                                                    ColoredBox(
                                                      color: colors.muted,
                                                    ),
                                                errorWidget: (_, _, _) =>
                                                    ColoredBox(
                                                      color: colors.muted,
                                                      child: Icon(
                                                        FIcons.package,
                                                        size: 20,
                                                        color:
                                                            colors.mutedForeground,
                                                      ),
                                                    ),
                                              )
                                            : ColoredBox(
                                                color: colors.muted,
                                                child: Icon(
                                                  FIcons.package,
                                                  size: 20,
                                                  color: colors.mutedForeground,
                                                ),
                                              ),
                                      ),
                                    ),
                                    title: Text(
                                      item.shopItemName ?? item.shopItemId,
                                    ),
                                    trailing: Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 8,
                                        vertical: 3,
                                      ),
                                      decoration: BoxDecoration(
                                        color: colors.primary
                                            .withValues(alpha: 0.12),
                                        borderRadius: BorderRadius.circular(10),
                                      ),
                                      child: Text(
                                        'x${item.quantity}',
                                        style: typography.xs.copyWith(
                                          color: colors.primary,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                    onTap: () => _sendGift(item),
                                  );
                                },
                              ),
              ),
          ],
        );
      },
    );
  }
}
```

**IMPORTANT NOTE on `showGiftPickerSheet`:** The two-function approach above is messy. Replace the entire file's exported function with just `showGiftPickerSheetScoped`, and rename it `showGiftPickerSheet` — that's the single exported function. Remove the broken first function. So the file should export only:

```dart
Future<void> showGiftPickerSheet(BuildContext context, String recipientUserId) {
  final container = ProviderScope.containerOf(context);
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    clipBehavior: Clip.antiAlias,
    builder: (sheetContext) => UncontrolledProviderScope(
      container: container,
      child: _GiftPickerSheet(recipientUserId: recipientUserId),
    ),
  );
}
```

- [ ] **Step 2: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/gift/presentation/widgets/gift_picker_sheet.dart
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/gift/presentation/widgets/gift_picker_sheet.dart
git commit -m "feat(mobile): add gift picker sheet with inventory selection

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 9: Gift Button on Profile Screen + Navigation Links

**Files:**
- Modify: `mobile/lib/features/profile/presentation/profile_screen.dart`

**Interfaces:**
- Consumes:
  - `showGiftPickerSheet` from Task 8
  - `AppRoutes.userGiftsReceived(userId)` from Task 5
  - `AppRoutes.sentGifts` from Task 6
- Produces: a gift icon button next to the Follow button on non-own profiles, a "View gifts" link on all profiles

**Change 1:** In `_buildUserInfo`, in the `Row` containing the follow button, add a gift icon button before the follow button (visible only when `!_isOwnProfile`):

```dart
if (!_isOwnProfile) ...[
  FButton.icon(
    onPress: () => showGiftPickerSheet(context, widget.userId),
    child: const Icon(FIcons.gift),
  ),
  const SizedBox(width: 8),
],
```

**Change 2:** Add a "View Gifts" link row after the follower count row, visible to everyone. For the own profile, push to `AppRoutes.sentGifts`. For other profiles, push to `AppRoutes.userGiftsReceived(widget.userId)`:

```dart
const SizedBox(height: 8),
FButton.raw(
  onPress: () => _isOwnProfile
      ? context.push(AppRoutes.sentGifts)
      : context.push(AppRoutes.userGiftsReceived(widget.userId)),
  child: Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Icon(FIcons.gift, size: 16, color: colors.primary),
      const SizedBox(width: 6),
      Text(
        _isOwnProfile ? 'My Sent Gifts' : 'View Gift Collection',
        style: typography.sm.copyWith(color: colors.primary),
      ),
    ],
  ),
),
```

- [ ] **Step 1: Add imports to `profile_screen.dart`**

```dart
import '../../gift/presentation/widgets/gift_picker_sheet.dart';
```

- [ ] **Step 2: Modify `_buildUserInfo` — add gift button to action row**

Find the `Row` children block in `_buildUserInfo` that has the `Expanded` column and the follow button. The current structure is:
```dart
Row(
  children: [
    Expanded(child: Column(...)),
    if (!_isOwnProfile) FButton(...)  // follow button
  ],
)
```

Replace with:
```dart
Row(
  children: [
    Expanded(child: Column(...)),
    if (!_isOwnProfile) ...[
      FButton.icon(
        onPress: () => showGiftPickerSheet(context, widget.userId),
        child: const Icon(FIcons.gift),
      ),
      const SizedBox(width: 8),
      FButton(
        onPress: _isFollowLoading ? null : _toggleFollow,
        variant: user.isFollowing == true
            ? FButtonVariant.destructive
            : null,
        child: _isFollowLoading
            ? const SizedBox(
                height: 16,
                width: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : Text(
                user.isFollowing == true
                    ? l10n.unfollow
                    : l10n.follow,
              ),
      ),
    ],
  ],
)
```

- [ ] **Step 3: Add "View Gifts" link after follower count Row in `_buildUserInfo`**

After the follower count `Row` and `const SizedBox(height: 16)` (before "About" text), add:

```dart
          const SizedBox(height: 8),
          FButton.raw(
            onPress: () => _isOwnProfile
                ? context.push(AppRoutes.sentGifts)
                : context.push(AppRoutes.userGiftsReceived(widget.userId)),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(FIcons.gift, size: 16, color: colors.primary),
                const SizedBox(width: 6),
                Text(
                  _isOwnProfile ? 'My Sent Gifts' : 'View Gift Collection',
                  style: typography.sm.copyWith(color: colors.primary),
                ),
              ],
            ),
          ),
```

- [ ] **Step 4: Verify analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze lib/features/profile/presentation/profile_screen.dart
```

Expected: no errors.

- [ ] **Step 5: Final full analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze 2>&1 | grep -v "emote_picker.dart"
```

Expected: output ends with `No issues found!` (or shows only the pre-existing emote_picker.dart issue after filtering).

- [ ] **Step 6: Commit**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile
git add lib/features/profile/presentation/profile_screen.dart
git commit -m "feat(mobile): add gift button and gift collection link to profile screen

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```

---

### Task 10: Final Analyze + Report

**Files:**
- Create: `/Users/ictsaigon.vn/mywork/letslive/.superpowers/sdd/task-12-report.md`

- [ ] **Step 1: Run full flutter analyze**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/mobile && flutter analyze 2>&1
```

Record the output. Expected: `No issues found!` (only emote_picker.dart error pre-existing is acceptable).

- [ ] **Step 2: Verify all new files exist**

```bash
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/data/models/gift_model.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/shop/data/models/inventory_item_model.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/data/repositories/gift_repository.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/shop/data/repositories/inventory_repository.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/presentation/widgets/gift_animation_overlay.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/presentation/widgets/gift_picker_sheet.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/presentation/screens/gifts_received_screen.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/gift/presentation/screens/sent_gifts_screen.dart
ls /Users/ictsaigon.vn/mywork/letslive/mobile/lib/features/wallet/presentation/inventory_tab.dart
```

- [ ] **Step 3: Write report**

Write `/Users/ictsaigon.vn/mywork/letslive/.superpowers/sdd/task-12-report.md` with:
- Files created/modified
- Animation approach used (FadeTransition + Image.network in showDialog, auto-dismiss after 2s)
- Routes added (/users/:userId/gifts, /gifts/sent)
- Analyze result (paste output)
- Commit hash (`git log --oneline -10`)
- Concerns (API field names assumed; if backend returns different field names the fromJson will fail silently returning null)

- [ ] **Step 4: Final commit for report**

```bash
cd /Users/ictsaigon.vn/mywork/letslive
git add .superpowers/sdd/task-12-report.md
git commit -m "feat(mobile): add gift feature and inventory tab

Refs: docs/superpowers/plans/2026-06-24-gift-feature.md"
```
