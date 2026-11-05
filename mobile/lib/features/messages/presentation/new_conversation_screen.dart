/// Start a conversation (web `NewConversationDialog`): pick one person for a
/// direct message, or several for a group with an optional name, then create.
/// Creating a direct message that already exists opens the existing one.
library;

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/router/app_routes.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/core/widgets/action_feedback.dart';
import 'package:letslive_mobile/core/widgets/feed_section.dart';
import 'package:letslive_mobile/features/messages/data/dm_models.dart';
import 'package:letslive_mobile/features/messages/messages_providers.dart';
import 'package:letslive_mobile/features/search/domain/search_type.dart';
import 'package:letslive_mobile/features/user/data/public_user.dart';
import 'package:letslive_mobile/features/user/presentation/user_avatar.dart';

/// Shortest query web searches for.
const int _minQueryLength = 2;

/// The new-conversation screen.
class NewConversationScreen extends ConsumerStatefulWidget {
  const NewConversationScreen({super.key});

  @override
  ConsumerState<NewConversationScreen> createState() =>
      _NewConversationScreenState();
}

class _NewConversationScreenState extends ConsumerState<NewConversationScreen> {
  final _search = TextEditingController();
  final _groupName = TextEditingController();
  Timer? _debounce;

  /// The trimmed text being searched; empty or short means no search.
  String _query = '';
  bool _isGroup = false;
  List<PublicUser> _selected = const [];
  bool _isCreating = false;

  @override
  void dispose() {
    _debounce?.cancel();
    _search.dispose();
    _groupName.dispose();
    super.dispose();
  }

  void _onQueryChanged(String text) {
    _debounce?.cancel();
    _debounce = Timer(searchDebounce, () {
      if (mounted) setState(() => _query = text.trim());
    });
  }

  void _setGroup({required bool isGroup}) => setState(() {
    _isGroup = isGroup;
    // A direct message has exactly one other person.
    if (!isGroup && _selected.length > 1) _selected = [_selected.first];
  });

  void _select(PublicUser user) {
    final l10n = AppLocalizations.of(context);
    if (_isGroup && _selected.length + 1 >= FieldLimits.groupMaxMembers) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            l10n.messages_group_member_limit(FieldLimits.groupMaxMembers),
          ),
        ),
      );
      return;
    }
    setState(() {
      _selected = _isGroup ? [..._selected, user] : [user];
      _query = '';
    });
    _search.clear();
  }

  void _remove(PublicUser user) => setState(
    () => _selected = [
      for (final u in _selected)
        if (u.id != user.id) u,
    ],
  );

  Future<void> _create() async {
    if (_selected.isEmpty || _isCreating) return;
    setState(() => _isCreating = true);
    final name = _groupName.text.trim();
    final response = await ref
        .read(messageRepositoryProvider)
        .create(
          type: _isGroup ? ConversationType.group : ConversationType.dm,
          participantIds: [for (final u in _selected) u.id],
          name: _isGroup && name.isNotEmpty ? name : null,
        );
    if (!mounted) return;
    setState(() => _isCreating = false);
    final created = response.data;
    if (!response.success || created == null) {
      showActionError(context, response);
      return;
    }
    if (ref.exists(conversationsProvider)) {
      ref.read(conversationsProvider.notifier).prepend(created);
    }
    context.pushReplacement(AppRoutes.conversation(created.id));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final myId = ref.watch(currentUserIdProvider);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.messages_new_conversation)),
      body: Column(
        children: [
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Wrap(
                  spacing: 8,
                  children: [
                    ChoiceChip(
                      key: const ValueKey('mode-dm'),
                      label: Text(l10n.messages_direct_message),
                      selected: !_isGroup,
                      onSelected: (_) => _setGroup(isGroup: false),
                    ),
                    ChoiceChip(
                      key: const ValueKey('mode-group'),
                      label: Text(l10n.messages_group),
                      selected: _isGroup,
                      onSelected: (_) => _setGroup(isGroup: true),
                    ),
                  ],
                ),
                if (_isGroup) ...[
                  const SizedBox(height: 12),
                  TextField(
                    key: const ValueKey('group-name'),
                    controller: _groupName,
                    inputFormatters: [
                      LengthLimitingTextInputFormatter(
                        FieldLimits.groupNameMax,
                      ),
                    ],
                    decoration: InputDecoration(
                      hintText: l10n.messages_group_name_placeholder,
                      border: const OutlineInputBorder(),
                    ),
                  ),
                ],
                if (_selected.isNotEmpty) ...[
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    children: [
                      for (final user in _selected)
                        InputChip(
                          key: ValueKey('selected-${user.id}'),
                          label: Text(
                            user.username,
                            overflow: TextOverflow.ellipsis,
                          ),
                          onDeleted: () => _remove(user),
                        ),
                    ],
                  ),
                ],
                const SizedBox(height: 12),
                TextField(
                  key: const ValueKey('user-search'),
                  controller: _search,
                  onChanged: _onQueryChanged,
                  decoration: InputDecoration(
                    hintText: l10n.messages_search_users_placeholder,
                    prefixIcon: const Icon(Icons.search),
                    border: const OutlineInputBorder(),
                  ),
                ),
                if (_query.length >= _minQueryLength)
                  _Results(
                    query: _query,
                    hidden: {myId, for (final u in _selected) u.id},
                    onPick: _select,
                  ),
              ],
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: SizedBox(
                width: double.infinity,
                child: FilledButton(
                  key: const ValueKey('create-conversation'),
                  onPressed: _selected.isEmpty || _isCreating ? null : _create,
                  child: Text(
                    _isCreating ? l10n.messages_creating : l10n.messages_create,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// The users matching [query], without the ones in [hidden] (the user
/// themself and those already picked).
class _Results extends ConsumerWidget {
  const _Results({
    required this.query,
    required this.hidden,
    required this.onPick,
  });

  final String query;
  final Set<String?> hidden;
  final ValueChanged<PublicUser> onPick;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final results = ref.watch(dmUserSearchProvider(query));
    if (results.hasError && !results.isLoading) {
      return FeedError(
        error: results.error!,
        onRetry: () => ref.invalidate(dmUserSearchProvider(query)),
      );
    }
    final users = results.value;
    if (users == null) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Center(child: Text(l10n.messages_searching)),
      );
    }
    return Column(
      children: [
        for (final user in users)
          if (!hidden.contains(user.id))
            ListTile(
              key: ValueKey('result-${user.id}'),
              leading: UserAvatar(
                pictureUrl: user.profilePicture,
                name: user.username,
              ),
              title: Text(
                user.username,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () => onPick(user),
            ),
      ],
    );
  }
}
