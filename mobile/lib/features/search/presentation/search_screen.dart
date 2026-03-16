import 'dart:async';

import 'package:cached_network_image/cached_network_image.dart';
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
      header: FHeader.nested(
        title: Text(l10n.searchUsers),
      ),
      child: Column(
        children: [
          // Search input
          Padding(
            padding: const EdgeInsets.all(16),
            child: TextField(
              controller: _searchController,
              onChanged: _onSearchChanged,
              maxLength: 100,
              textInputAction: TextInputAction.search,
              style: typography.sm,
              decoration: InputDecoration(
                hintText: l10n.messagesSearchUsersPlaceholder,
                hintStyle:
                    typography.sm.copyWith(color: colors.mutedForeground),
                prefixIcon:
                    Icon(FIcons.search, color: colors.mutedForeground),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: Icon(FIcons.x,
                            color: colors.mutedForeground, size: 18),
                        onPressed: () {
                          _searchController.clear();
                          _onSearchChanged('');
                        },
                      )
                    : null,
                counterText: '',
                contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16, vertical: 12),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: colors.border),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: colors.primary),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: colors.border),
                ),
              ),
            ),
          ),

          // Results
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
            Icon(FIcons.searchX, size: 48,
                color: colors.mutedForeground),
            const SizedBox(height: 16),
            Text(
              l10n.noUsersFound,
              style: typography.base
                  .copyWith(color: colors.mutedForeground),
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
      separatorBuilder: (_, _) => Divider(
        height: 1,
        color: colors.border,
      ),
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

    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 12),
        child: Row(
          children: [
            // Avatar
            CircleAvatar(
              radius: 20,
              backgroundColor: colors.muted,
              backgroundImage: user.profilePicture != null
                  ? CachedNetworkImageProvider(
                      '${AppConfig.apiUrl}/${user.profilePicture}')
                  : null,
              child: user.profilePicture == null
                  ? Text(
                      (user.displayName ?? user.username)
                          .characters
                          .first
                          .toUpperCase(),
                      style: typography.sm
                          .copyWith(fontWeight: FontWeight.w600),
                    )
                  : null,
            ),
            const SizedBox(width: 12),
            // User info
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    user.displayName ?? user.username,
                    style: typography.sm
                        .copyWith(fontWeight: FontWeight.w600),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    '@${user.username}',
                    style: typography.xs
                        .copyWith(color: colors.mutedForeground),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            Icon(FIcons.chevronRight, size: 16,
                color: colors.mutedForeground),
          ],
        ),
      ),
    );
  }
}
