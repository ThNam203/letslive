import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../providers.dart';
import '../../data/models/shop_item_model.dart';
import '../widgets/shop_item_card.dart';
import 'shop_item_detail_sheet.dart';

class ShopScreen extends ConsumerStatefulWidget {
  const ShopScreen({super.key});

  @override
  ConsumerState<ShopScreen> createState() => _ShopScreenState();
}

class _ShopScreenState extends ConsumerState<ShopScreen> {
  List<ShopItem> _items = [];
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _fetchItems();
  }

  Future<void> _fetchItems() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final repo = ref.read(shopRepositoryProvider);
      final response = await repo.getItems();

      if (!mounted) return;

      if (response.success && response.data != null) {
        setState(() => _items = response.data!);
      } else {
        setState(() {
          _errorMessage = response.message.isNotEmpty
              ? response.message
              : 'Failed to load shop items.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => _errorMessage = 'Failed to load shop items.');
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return FScaffold(
      header: const FHeader(
        title: Text('Shop'),
      ),
      child: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    final typography = context.theme.typography;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                FIcons.circleAlert,
                size: 48,
                color: isDark
                    ? AppColors.darkForegroundMuted
                    : AppColors.lightForegroundMuted,
              ),
              const SizedBox(height: 16),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: typography.sm.copyWith(
                  color: isDark
                      ? AppColors.darkForegroundMuted
                      : AppColors.lightForegroundMuted,
                ),
              ),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _fetchItems,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    if (_items.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                FIcons.shoppingBag,
                size: 48,
                color: isDark
                    ? AppColors.darkForegroundMuted
                    : AppColors.lightForegroundMuted,
              ),
              const SizedBox(height: 16),
              Text(
                'No items available',
                style: typography.lg.copyWith(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 8),
              Text(
                'Check back later for new items',
                textAlign: TextAlign.center,
                style: typography.sm.copyWith(
                  color: isDark
                      ? AppColors.darkForegroundMuted
                      : AppColors.lightForegroundMuted,
                ),
              ),
            ],
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchItems,
      child: GridView.builder(
        padding: const EdgeInsets.all(16),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          crossAxisSpacing: 12,
          mainAxisSpacing: 12,
          childAspectRatio: 0.75,
        ),
        itemCount: _items.length,
        itemBuilder: (context, index) {
          final item = _items[index];
          return ShopItemCard(
            item: item,
            onTap: () => showShopItemDetail(context, item),
          );
        },
      ),
    );
  }
}
