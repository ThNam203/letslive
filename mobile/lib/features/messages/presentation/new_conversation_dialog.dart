import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../../../models/user.dart';
import '../../../providers.dart';

class NewConversationDialog extends ConsumerStatefulWidget {
  final void Function(String conversationId) onCreated;

  const NewConversationDialog({super.key, required this.onCreated});

  @override
  ConsumerState<NewConversationDialog> createState() =>
      _NewConversationDialogState();
}

class _NewConversationDialogState extends ConsumerState<NewConversationDialog> {
  final _searchController = TextEditingController();
  final _groupNameController = TextEditingController();
  Timer? _debounce;

  List<User> _searchResults = [];
  final List<User> _selectedUsers = [];
  bool _isSearching = false;
  bool _isCreating = false;

  @override
  void dispose() {
    _searchController.dispose();
    _groupNameController.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _onSearchChanged(String query) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      if (query.trim().length >= 2) {
        _searchUsers(query.trim());
      } else {
        setState(() => _searchResults = []);
      }
    });
  }

  Future<void> _searchUsers(String query) async {
    setState(() => _isSearching = true);
    try {
      final repo = ref.read(userRepositoryProvider);
      final response = await repo.searchUsers(query: query);
      if (mounted && response.success) {
        final currentUser = ref.read(currentUserProvider);
        setState(() {
          _searchResults = (response.data ?? [])
              .where((u) => u.id != currentUser?.id)
              .toList();
          _isSearching = false;
        });
      } else {
        if (mounted) setState(() => _isSearching = false);
      }
    } on DioException catch (_) {
      if (mounted) setState(() => _isSearching = false);
    }
  }

  void _toggleUser(User user) {
    setState(() {
      final idx = _selectedUsers.indexWhere((u) => u.id == user.id);
      if (idx != -1) {
        _selectedUsers.removeAt(idx);
      } else {
        _selectedUsers.add(user);
      }
    });
  }

  Future<void> _createConversation() async {
    if (_selectedUsers.isEmpty) return;
    setState(() => _isCreating = true);

    final currentUser = ref.read(currentUserProvider);
    if (currentUser == null) {
      setState(() => _isCreating = false);
      return;
    }

    final isGroup = _selectedUsers.length > 1;

    final participantUsernames = <String, String>{};
    final participantDisplayNames = <String, String>{};
    final participantProfilePictures = <String, String>{};

    for (final u in _selectedUsers) {
      participantUsernames[u.id] = u.username;
      if (u.displayName != null) participantDisplayNames[u.id] = u.displayName!;
      if (u.profilePicture != null) {
        participantProfilePictures[u.id] = u.profilePicture!;
      }
    }

    try {
      final repo = ref.read(messageRepositoryProvider);
      final response = await repo.createConversation(
        type: isGroup ? 'group' : 'dm',
        participantIds: _selectedUsers.map((u) => u.id).toList(),
        participantUsernames: participantUsernames,
        participantDisplayNames:
            participantDisplayNames.isNotEmpty ? participantDisplayNames : null,
        participantProfilePictures: participantProfilePictures.isNotEmpty
            ? participantProfilePictures
            : null,
        creatorUsername: currentUser.username,
        creatorDisplayName: currentUser.displayName,
        creatorProfilePicture: currentUser.profilePicture,
        name: isGroup
            ? _groupNameController.text.trim().isNotEmpty
                  ? _groupNameController.text.trim()
                  : null
            : null,
      );
      if (!mounted) return;

      if (response.success && response.data != null) {
        Navigator.pop(context);
        widget.onCreated(response.data!.id);
      } else {
        setState(() => _isCreating = false);
        showFToast(
          context: context,
          title: Text(response.message),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() => _isCreating = false);
        showFToast(
          context: context,
          title: Text(AppLocalizations.of(context).fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Dialog(
      insetPadding: const EdgeInsets.all(16),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 400, maxHeight: 520),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                l10n.messagesNewConversation,
                style: typography.lg.copyWith(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),

              // Selected users chips
              if (_selectedUsers.isNotEmpty) ...[
                Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: _selectedUsers
                      .map(
                        (u) => Chip(
                          label: Text(
                            u.displayName ?? u.username,
                            style: typography.xs,
                          ),
                          deleteIcon: const Icon(FIcons.x, size: 14),
                          onDeleted: () => _toggleUser(u),
                          visualDensity: VisualDensity.compact,
                        ),
                      )
                      .toList(),
                ),
                const SizedBox(height: 8),
              ],

              // Group name (for multiple users)
              if (_selectedUsers.length > 1) ...[
                TextField(
                  controller: _groupNameController,
                  decoration: InputDecoration(
                    hintText: l10n.messagesGroupNamePlaceholder,
                    hintStyle: typography.sm.copyWith(
                      color: colors.mutedForeground,
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: colors.border),
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 10,
                    ),
                    isDense: true,
                  ),
                  style: typography.sm,
                ),
                const SizedBox(height: 8),
              ],

              // Search input
              TextField(
                controller: _searchController,
                onChanged: _onSearchChanged,
                decoration: InputDecoration(
                  hintText: l10n.messagesSearchUsersPlaceholder,
                  hintStyle: typography.sm.copyWith(
                    color: colors.mutedForeground,
                  ),
                  prefixIcon: const Icon(FIcons.search, size: 16),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: colors.border),
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 10,
                  ),
                  isDense: true,
                ),
                style: typography.sm,
              ),
              const SizedBox(height: 8),

              // Search results
              Expanded(
                child: _isSearching
                    ? Center(
                        child: Text(
                          l10n.messagesSearching,
                          style: typography.sm.copyWith(
                            color: colors.mutedForeground,
                          ),
                        ),
                      )
                    : ListView.builder(
                        itemCount: _searchResults.length,
                        itemBuilder: (context, index) {
                          final user = _searchResults[index];
                          final isSelected = _selectedUsers.any(
                            (u) => u.id == user.id,
                          );
                          return ListTile(
                            dense: true,
                            leading: CircleAvatar(
                              radius: 16,
                              child: Text(
                                (user.username.isNotEmpty
                                        ? user.username[0]
                                        : '?')
                                    .toUpperCase(),
                                style: typography.xs,
                              ),
                            ),
                            title: Text(
                              user.displayName ?? user.username,
                              style: typography.sm,
                            ),
                            subtitle: Text(
                              '@${user.username}',
                              style: typography.xs.copyWith(
                                color: colors.mutedForeground,
                              ),
                            ),
                            trailing: isSelected
                                ? Icon(
                                    FIcons.check,
                                    size: 16,
                                    color: colors.primary,
                                  )
                                : null,
                            onTap: () => _toggleUser(user),
                          );
                        },
                      ),
              ),

              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  FButton(
                    variant: FButtonVariant.outline,
                    onPress: () => Navigator.pop(context),
                    child: Text(l10n.cancel),
                  ),
                  const SizedBox(width: 8),
                  FButton(
                    onPress: _selectedUsers.isEmpty || _isCreating
                        ? null
                        : _createConversation,
                    child: _isCreating
                        ? const SizedBox(
                            height: 16,
                            width: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.messagesCreate),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
