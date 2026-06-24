import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../models/user.dart';
import '../../../../providers.dart';
import '../../data/models/shop_item_model.dart';


/// Opens a modal bottom sheet with item detail and buy/send CTAs.
Future<void> showShopItemDetail(BuildContext context, ShopItem item) {
  return showModalBottomSheet(
    context: context,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    clipBehavior: Clip.antiAlias,
    builder: (sheetContext) => _ShopItemDetailSheet(item: item),
  );
}

class _ShopItemDetailSheet extends ConsumerStatefulWidget {
  final ShopItem item;

  const _ShopItemDetailSheet({required this.item});

  @override
  ConsumerState<_ShopItemDetailSheet> createState() =>
      _ShopItemDetailSheetState();
}

class _ShopItemDetailSheetState extends ConsumerState<_ShopItemDetailSheet> {
  int _quantity = 1;
  bool _isBuying = false;
  String? _errorMessage;
  String? _successMessage;

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
        setState(() {
          _successMessage = recipientUserId != null
              ? 'Gift sent successfully!'
              : 'Added to your inventory!';
        });
        await Future<void>.delayed(const Duration(seconds: 1));
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

  Future<void> _showRecipientPicker() async {
    final user = await showModalBottomSheet<User?>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      clipBehavior: Clip.antiAlias,
      builder: (ctx) => _UserPickerSheet(shopItem: widget.item),
    );

    if (user != null) {
      await _purchase(recipientUserId: user.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final typography = context.theme.typography;
    final colors = context.theme.colors;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final item = widget.item;

    return DraggableScrollableSheet(
      initialChildSize: 0.7,
      minChildSize: 0.4,
      maxChildSize: 0.92,
      expand: false,
      builder: (ctx, scrollController) {
        return SingleChildScrollView(
          controller: scrollController,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
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

              // Item image
              AspectRatio(
                aspectRatio: 16 / 9,
                child: CachedNetworkImage(
                  imageUrl: item.imageUrl,
                  fit: BoxFit.cover,
                  placeholder: (context, url) => Container(
                    color: isDark
                        ? const Color(0xFF334155)
                        : Colors.grey.shade100,
                    child: const Center(child: CircularProgressIndicator()),
                  ),
                  errorWidget: (context, url, error) => Container(
                    color: isDark
                        ? const Color(0xFF334155)
                        : Colors.grey.shade100,
                    child: Icon(
                      FIcons.imageOff,
                      color: isDark ? Colors.white38 : Colors.grey.shade400,
                      size: 48,
                    ),
                  ),
                ),
              ),

              Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Name + price
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          child: Text(
                            item.name,
                            style: typography.xl.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            gradient: LinearGradient(
                              colors: [
                                Colors.amber.shade500,
                                Colors.orange.shade600,
                              ],
                            ),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.bolt,
                                  color: Colors.white, size: 16),
                              const SizedBox(width: 4),
                              Text(
                                '${item.price}',
                                style: typography.sm.copyWith(
                                  color: Colors.white,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),

                    if (item.description != null &&
                        item.description!.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      Text(
                        item.description!,
                        style: typography.sm.copyWith(
                          color: isDark
                              ? const Color(0xFFCCCCCC)
                              : Colors.grey.shade600,
                        ),
                      ),
                    ],

                    const SizedBox(height: 24),

                    // Quantity selector
                    Row(
                      children: [
                        Text(
                          'Quantity',
                          style: typography.sm.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const Spacer(),
                        IconButton(
                          onPressed: _quantity > 1
                              ? () => setState(() => _quantity--)
                              : null,
                          icon: const Icon(FIcons.minus, size: 18),
                          padding: EdgeInsets.zero,
                          constraints:
                              const BoxConstraints(minWidth: 32, minHeight: 32),
                          style: IconButton.styleFrom(
                            backgroundColor: isDark
                                ? const Color(0xFF334155)
                                : Colors.grey.shade100,
                            shape: const CircleBorder(),
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 16),
                          child: Text(
                            '$_quantity',
                            style: typography.lg
                                .copyWith(fontWeight: FontWeight.bold),
                          ),
                        ),
                        IconButton(
                          onPressed: () => setState(() => _quantity++),
                          icon: const Icon(FIcons.plus, size: 18),
                          padding: EdgeInsets.zero,
                          constraints:
                              const BoxConstraints(minWidth: 32, minHeight: 32),
                          style: IconButton.styleFrom(
                            backgroundColor: colors.primary,
                            foregroundColor: Colors.white,
                            shape: const CircleBorder(),
                          ),
                        ),
                      ],
                    ),

                    const SizedBox(height: 8),

                    // Total cost
                    Text(
                      'Total: ${item.price * _quantity} Spark',
                      style: typography.xs.copyWith(
                        color: isDark
                            ? const Color(0xFFCCCCCC)
                            : Colors.grey.shade600,
                      ),
                    ),

                    const SizedBox(height: 24),

                    // Feedback messages
                    if (_errorMessage != null) ...[
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.red.shade50,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: Colors.red.shade200),
                        ),
                        child: Text(
                          _errorMessage!,
                          style: typography.sm
                              .copyWith(color: Colors.red.shade700),
                        ),
                      ),
                      const SizedBox(height: 12),
                    ],
                    if (_successMessage != null) ...[
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.green.shade50,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(color: Colors.green.shade200),
                        ),
                        child: Text(
                          _successMessage!,
                          style: typography.sm
                              .copyWith(color: Colors.green.shade700),
                        ),
                      ),
                      const SizedBox(height: 12),
                    ],

                    // CTA buttons
                    if (_isBuying)
                      const Center(child: CircularProgressIndicator())
                    else ...[
                      SizedBox(
                        width: double.infinity,
                        child: FilledButton.icon(
                          onPressed: () => _purchase(),
                          icon: const Icon(FIcons.shoppingBag, size: 18),
                          label: const Text('Add to Inventory'),
                          style: FilledButton.styleFrom(
                            backgroundColor: colors.primary,
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 10),
                      SizedBox(
                        width: double.infinity,
                        child: OutlinedButton.icon(
                          onPressed: _showRecipientPicker,
                          icon: const Icon(FIcons.gift, size: 18),
                          label: const Text('Buy & Send as Gift'),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: colors.primary,
                            side: BorderSide(color: colors.primary),
                            padding: const EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10),
                            ),
                          ),
                        ),
                      ),
                    ],

                    const SizedBox(height: 20),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

// ---------------------------------------------------------------------------
// User picker sheet for "Buy & Send" flow
// ---------------------------------------------------------------------------

class _UserPickerSheet extends ConsumerStatefulWidget {
  final ShopItem shopItem;

  const _UserPickerSheet({required this.shopItem});

  @override
  ConsumerState<_UserPickerSheet> createState() => _UserPickerSheetState();
}

class _UserPickerSheetState extends ConsumerState<_UserPickerSheet> {
  final TextEditingController _searchController = TextEditingController();
  List<User> _results = [];
  bool _isSearching = false;
  String _lastQuery = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _search(String query) async {
    final trimmed = query.trim();
    if (trimmed == _lastQuery) return;
    _lastQuery = trimmed;

    if (trimmed.isEmpty) {
      setState(() => _results = []);
      return;
    }

    setState(() => _isSearching = true);
    try {
      final repo = ref.read(userRepositoryProvider);
      final response = await repo.searchUsers(query: trimmed);
      if (!mounted) return;
      if (response.success && response.data != null) {
        setState(() => _results = response.data!);
      }
    } catch (_) {
      // leave results as-is
    } finally {
      if (mounted) setState(() => _isSearching = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final typography = context.theme.typography;
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
                'Send to a user',
                style: typography.lg.copyWith(fontWeight: FontWeight.bold),
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: TextField(
                controller: _searchController,
                autofocus: true,
                decoration: InputDecoration(
                  hintText: 'Search username...',
                  prefixIcon: const Icon(FIcons.search, size: 18),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(10),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 10,
                  ),
                ),
                onChanged: _search,
              ),
            ),
            const SizedBox(height: 8),
            Expanded(
              child: _isSearching
                  ? const Center(child: CircularProgressIndicator())
                  : _results.isEmpty
                      ? Center(
                          child: Text(
                            _lastQuery.isEmpty
                                ? 'Type a username to search'
                                : 'No users found',
                            style: typography.sm.copyWith(
                              color: isDark
                                  ? const Color(0xFFCCCCCC)
                                  : Colors.grey.shade600,
                            ),
                          ),
                        )
                      : ListView.builder(
                          controller: scrollController,
                          itemCount: _results.length,
                          itemBuilder: (ctx, index) {
                            final user = _results[index];
                            return ListTile(
                              leading: CircleAvatar(
                                radius: 20,
                                backgroundImage:
                                    user.profilePicture != null
                                        ? CachedNetworkImageProvider(
                                            user.profilePicture!)
                                        : null,
                                child: user.profilePicture == null
                                    ? Text(
                                        (user.username ?? '?')
                                            .substring(0, 1)
                                            .toUpperCase(),
                                        style: const TextStyle(
                                          fontWeight: FontWeight.bold,
                                        ),
                                      )
                                    : null,
                              ),
                              title: Text(user.username ?? 'Unknown'),
                              subtitle: user.bio != null && user.bio!.isNotEmpty
                                  ? Text(
                                      user.bio!,
                                      maxLines: 1,
                                      overflow: TextOverflow.ellipsis,
                                    )
                                  : null,
                              onTap: () => Navigator.of(context).pop(user),
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
