import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';
import 'package:go_router/go_router.dart';

import '../../../core/config/app_config.dart';
import '../../../core/router/app_router.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/user.dart';
import '../../../providers.dart';
import '../../../shared/widgets/user_avatar.dart';

class SearchScreen extends ConsumerStatefulWidget {
  const SearchScreen({super.key});

  @override
  ConsumerState<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends ConsumerState<SearchScreen> {
  final _searchController = TextEditingController();
  Timer? _debounce;
  List<User> _results = [];
  bool _isLoading = false;
  bool _hasSearched = false;

  @override
  void dispose() {
    _debounce?.cancel();
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged(String query) {
    _debounce?.cancel();
    if (query.trim().isEmpty) {
      setState(() {
        _results = [];
        _isLoading = false;
        _hasSearched = false;
      });
      return;
    }

    setState(() => _isLoading = true);
    _debounce = Timer(const Duration(seconds: 1), () {
      _performSearch(query.trim());
    });
  }

  Future<void> _performSearch(String query) async {
    try {
      final userRepo = ref.read(userRepositoryProvider);
      final response = await userRepo.searchUsers(query: query);
      if (!mounted) return;

      setState(() {
        _results = response.data ?? [];
        _isLoading = false;
        _hasSearched = true;
      });
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _results = [];
          _isLoading = false;
          _hasSearched = true;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(title: Text(l10n.searchUsers)),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: FTextField(
              control: FTextFieldControl.managed(
                controller: _searchController,
                onChange: (value) => _onSearchChanged(value.text),
              ),
              hint: l10n.messagesSearchUsersPlaceholder,
              keyboardType: TextInputType.text,
              textInputAction: TextInputAction.search,
            ),
          ),
          if (_searchController.text.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 8),
              child: Align(
                alignment: Alignment.centerRight,
                child: FButton.icon(
                  variant: FButtonVariant.ghost,
                  onPress: () {
                    _searchController.clear();
                    _onSearchChanged('');
                    setState(() {});
                  },
                  child: const Icon(FIcons.x),
                ),
              ),
            ),
            Expanded(child: _buildResults(colors, typography, l10n)),
        ],
      ),
    );
  }

  Widget _buildResults(
    FColors colors,
    FTypography typography,
    AppLocalizations l10n,
  ) {
    if (_isLoading) {
      return Center(
        child: Text(
          l10n.searching,
          style: typography.sm.copyWith(color: colors.mutedForeground),
        ),
      );
    }

    if (_hasSearched && _results.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(FIcons.searchX, size: 48, color: colors.mutedForeground),
            const SizedBox(height: 16),
            Text(
              l10n.noUsersFound,
              style: typography.base.copyWith(color: colors.mutedForeground),
            ),
          ],
        ),
      );
    }

    if (_results.isEmpty) {
      return const SizedBox.shrink();
    }

    return ListView.separated(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: _results.length,
      separatorBuilder: (_, _) => const SizedBox(height: 4),
      itemBuilder: (context, index) {
        final user = _results[index];
        return _UserResultTile(
          user: user,
          onTap: () => context.push(AppRoutes.userProfile(user.id)),
        );
      },
    );
  }
}

class _UserResultTile extends StatelessWidget {
  final User user;
  final VoidCallback onTap;

  const _UserResultTile({required this.user, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FTile(
      onPress: onTap,
      prefix: UserAvatar(
        imageUrl: user.profilePicture != null
            ? '${AppConfig.apiUrl}/${user.profilePicture}'
            : null,
        size: 40,
        fallbackText: (user.displayName ?? user.username).characters.first
            .toUpperCase(),
        textStyle: typography.sm.copyWith(fontWeight: FontWeight.w600),
      ),
      title: Text(
        user.displayName ?? user.username,
        style: typography.sm.copyWith(fontWeight: FontWeight.w600),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(
        '@${user.username}',
        style: typography.xs.copyWith(
          color: colors.mutedForeground,
        ),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      suffix: Icon(FIcons.chevronRight, size: 16, color: colors.mutedForeground),
    );
  }
}
