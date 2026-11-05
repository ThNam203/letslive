/// The creator's custom chat commands (web `/settings/chat-commands`): a
/// personal list (usable in any channel), a channel list (usable by anyone
/// chatting in the creator's channel), and the built-in commands for
/// reference.
///
/// Each list has its own add / edit form, like web. The name is lowercased and
/// must be 1-32 characters of `a-z 0-9 _ -`; the response is required.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/chat_parser/command.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/network/api_response.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';
import 'package:letslive_mobile/features/creator/creator_providers.dart';
import 'package:letslive_mobile/features/creator/data/chat_command_admin_repository.dart';
import 'package:letslive_mobile/features/home/presentation/feed_section.dart';

/// `NAME_PATTERN` in web's page.
final RegExp _namePattern = RegExp(r'^[a-z0-9_-]{1,32}$');

/// The chat commands screen.
class ChatCommandsScreen extends ConsumerWidget {
  const ChatCommandsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final commands = ref.watch(myChatCommandsProvider);
    final body = commands.hasValue
        ? null
        : feedStatusBody<MyChatCommands>(
            value: commands.whenData((c) => [c]),
            onRetry: () => ref.invalidate(myChatCommandsProvider),
          );
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settings_navigation_chat_commands)),
      body: body != null
          ? SingleChildScrollView(child: body)
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _ScopeSection(
                  scope: ChatCommandScope.user,
                  title: l10n.chat_commands_page_personal_title,
                  description: l10n.chat_commands_page_personal_description,
                  items: commands.requireValue.user,
                ),
                const SizedBox(height: 24),
                _ScopeSection(
                  scope: ChatCommandScope.channel,
                  title: l10n.chat_commands_page_channel_title,
                  description: l10n.chat_commands_page_channel_description,
                  items: commands.requireValue.channel,
                ),
                const SizedBox(height: 24),
                const _BuiltinSection(),
              ],
            ),
    );
  }
}

/// One scope: its commands and the add / edit form.
class _ScopeSection extends ConsumerStatefulWidget {
  const _ScopeSection({
    required this.scope,
    required this.title,
    required this.description,
    required this.items,
  });

  final ChatCommandScope scope;
  final String title;
  final String description;
  final List<ChatCommandEntry> items;

  @override
  ConsumerState<_ScopeSection> createState() => _ScopeSectionState();
}

class _ScopeSectionState extends ConsumerState<_ScopeSection> {
  final _name = TextEditingController();
  final _response = TextEditingController();
  final _description = TextEditingController();
  ChatCommandEntry? _editing;
  bool _busy = false;

  String get _prefix => widget.scope.name;

  @override
  void dispose() {
    _name.dispose();
    _response.dispose();
    _description.dispose();
    super.dispose();
  }

  void _say(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  void _reset() => setState(() {
    _editing = null;
    _name.clear();
    _response.clear();
    _description.clear();
  });

  void _startEdit(ChatCommandEntry command) => setState(() {
    _editing = command;
    _name.text = command.name;
    _response.text = command.response;
    _description.text = command.description;
  });

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context);
    final name = _name.text.trim().toLowerCase();
    if (!_namePattern.hasMatch(name)) {
      return _say(l10n.chat_commands_page_name_invalid_toast);
    }
    final response = _response.text.trim();
    if (response.isEmpty) {
      return _say(l10n.chat_commands_page_response_required_toast);
    }
    final editing = _editing;
    final repo = ref.read(chatCommandAdminRepositoryProvider);
    setState(() => _busy = true);
    final result = editing == null
        ? await repo.create(
            scope: widget.scope,
            name: name,
            response: response,
            description: _description.text.trim(),
          )
        : await repo.update(
            editing.id,
            name: name,
            response: response,
            description: _description.text.trim(),
          );
    if (!mounted) return;
    setState(() => _busy = false);
    if (result.success) {
      _say(
        editing == null
            ? l10n.chat_commands_page_added_toast(name)
            : l10n.chat_commands_page_updated_toast(name),
      );
      _reset();
      ref.invalidate(myChatCommandsProvider);
    } else {
      // The usual cause is a taken name (web's wording says so).
      _say(
        editing == null
            ? l10n.chat_commands_page_create_failed_toast
            : l10n.chat_commands_page_update_failed_toast,
      );
    }
  }

  Future<void> _remove(ChatCommandEntry command) async {
    final l10n = AppLocalizations.of(context);
    final ApiResponse<Object?> result = await ref
        .read(chatCommandAdminRepositoryProvider)
        .delete(command.id);
    if (!mounted) return;
    if (result.success) {
      _say(l10n.chat_commands_page_removed_toast);
      if (_editing?.id == command.id) _reset();
      ref.invalidate(myChatCommandsProvider);
    } else {
      _say(l10n.chat_commands_page_remove_failed_toast);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final editing = _editing;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(widget.title, style: theme.textTheme.titleMedium),
        Text(widget.description),
        const SizedBox(height: 8),
        if (widget.items.isEmpty)
          Text(l10n.chat_commands_page_empty)
        else
          for (final command in widget.items)
            Card(
              child: ListTile(
                key: ValueKey('command-$_prefix-${command.name}'),
                title: Text('/${command.name}'),
                subtitle: Text(
                  command.description.isEmpty
                      ? command.response
                      : '${command.response}\n${command.description}',
                ),
                isThreeLine: command.description.isNotEmpty,
                trailing: Wrap(
                  children: [
                    IconButton(
                      key: ValueKey('edit-$_prefix-${command.name}'),
                      tooltip: l10n.chat_commands_page_edit_aria,
                      icon: const Icon(Icons.edit_outlined),
                      onPressed: () => _startEdit(command),
                    ),
                    IconButton(
                      key: ValueKey('remove-$_prefix-${command.name}'),
                      tooltip: l10n.chat_commands_page_remove_aria,
                      icon: const Icon(Icons.close),
                      onPressed: () => _remove(command),
                    ),
                  ],
                ),
              ),
            ),
        const Divider(height: 32),
        if (editing != null)
          Text(l10n.chat_commands_page_edit_title(editing.name)),
        TextField(
          key: ValueKey('name-$_prefix'),
          controller: _name,
          enabled: !_busy,
          maxLength: FieldLimits.chatCommandNameMax,
          decoration: InputDecoration(
            labelText: l10n.chat_commands_page_name_label,
            hintText: l10n.chat_commands_page_name_placeholder,
          ),
        ),
        TextField(
          key: ValueKey('response-$_prefix'),
          controller: _response,
          enabled: !_busy,
          maxLength: FieldLimits.chatCommandResponseMax,
          decoration: InputDecoration(
            labelText: l10n.chat_commands_page_response_label,
            hintText: l10n.chat_commands_page_response_placeholder,
          ),
        ),
        TextField(
          key: ValueKey('description-$_prefix'),
          controller: _description,
          enabled: !_busy,
          maxLength: FieldLimits.chatCommandDescriptionMax,
          decoration: InputDecoration(
            labelText: l10n.chat_commands_page_description_label,
            hintText: l10n.chat_commands_page_description_placeholder,
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          alignment: WrapAlignment.end,
          spacing: 8,
          children: [
            if (editing != null)
              TextButton(
                onPressed: _busy ? null : _reset,
                child: Text(l10n.chat_commands_page_cancel_edit),
              ),
            SubmitButton(
              key: ValueKey('submit-$_prefix'),
              expanded: false,
              busy: _busy,
              label: editing == null
                  ? l10n.chat_commands_page_submit
                  : l10n.chat_commands_page_submit_edit,
              onPressed: _submit,
            ),
          ],
        ),
      ],
    );
  }
}

/// The commands that ship with the app, for reference.
class _BuiltinSection extends StatelessWidget {
  const _BuiltinSection();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.chat_commands_page_builtin_title,
          style: Theme.of(context).textTheme.titleMedium,
        ),
        Text(l10n.chat_commands_page_builtin_description),
        const SizedBox(height: 8),
        for (final command in builtinChatCommands)
          ListTile(
            dense: true,
            contentPadding: EdgeInsets.zero,
            title: Text(
              command.usage,
              style: const TextStyle(fontFamily: 'monospace'),
            ),
            subtitle: Text(command.description(l10n)),
          ),
      ],
    );
  }
}
