import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/chat_parser/command.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/chat_command.dart';
import '../../../providers.dart';

final RegExp _namePattern = RegExp(r'^[a-z0-9_-]{1,32}$');
const int _maxResponse = 500;
const int _maxDescription = 120;

class ChatCommandsSettingsScreen extends ConsumerStatefulWidget {
  const ChatCommandsSettingsScreen({super.key});

  @override
  ConsumerState<ChatCommandsSettingsScreen> createState() =>
      _ChatCommandsSettingsScreenState();
}

class _ChatCommandsSettingsScreenState
    extends ConsumerState<ChatCommandsSettingsScreen> {
  MyChatCommands _data = const MyChatCommands(user: [], channel: []);
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  Future<void> _refresh() async {
    setState(() => _loading = true);
    final repo = ref.read(chatCommandRepositoryProvider);
    final res = await repo.listMine();
    if (!mounted) return;
    if (res.success && res.data != null) {
      setState(() {
        _data = res.data!;
        _loading = false;
      });
    } else {
      setState(() => _loading = false);
    }
  }

  Future<void> _delete(ChatCommand cmd) async {
    final l10n = AppLocalizations.of(context);
    final repo = ref.read(chatCommandRepositoryProvider);
    final res = await repo.delete(cmd.id);
    if (!mounted) return;
    if (res.success) {
      showFToast(
        context: context,
        title: Text(l10n.chatCommandsRemovedToast),
        icon: const Icon(FIcons.check),
      );
      _refresh();
    } else {
      showFToast(
        context: context,
        title: Text(l10n.chatCommandsRemoveFailedToast),
        icon: const Icon(FIcons.circleAlert),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(title: Text(l10n.settingsNavChatCommands)),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _ScopeSection(
            title: l10n.chatCommandsPersonalTitle,
            description: l10n.chatCommandsPersonalDescription,
            scope: ChatCommandScope.user,
            items: _data.user,
            onCreated: _refresh,
            onDelete: _delete,
          ),
          const SizedBox(height: 24),
          _ScopeSection(
            title: l10n.chatCommandsChannelTitle,
            description: l10n.chatCommandsChannelDescription,
            scope: ChatCommandScope.channel,
            items: _data.channel,
            onCreated: _refresh,
            onDelete: _delete,
          ),
          const SizedBox(height: 24),
          _BuiltinsSection(),
          if (_loading)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Center(child: CircularProgressIndicator()),
            ),
        ],
      ),
    );
  }
}

class _ScopeSection extends ConsumerStatefulWidget {
  final String title;
  final String description;
  final ChatCommandScope scope;
  final List<ChatCommand> items;
  final VoidCallback onCreated;
  final ValueChanged<ChatCommand> onDelete;

  const _ScopeSection({
    required this.title,
    required this.description,
    required this.scope,
    required this.items,
    required this.onCreated,
    required this.onDelete,
  });

  @override
  ConsumerState<_ScopeSection> createState() => _ScopeSectionState();
}

class _ScopeSectionState extends ConsumerState<_ScopeSection> {
  final _nameController = TextEditingController();
  final _responseController = TextEditingController();
  final _descController = TextEditingController();
  bool _submitting = false;
  ChatCommand? _editing;

  @override
  void dispose() {
    _nameController.dispose();
    _responseController.dispose();
    _descController.dispose();
    super.dispose();
  }

  void _startEdit(ChatCommand cmd) {
    setState(() {
      _editing = cmd;
      _nameController.text = cmd.name;
      _responseController.text = cmd.response;
      _descController.text = cmd.description;
    });
  }

  void _resetForm() {
    setState(() {
      _editing = null;
      _nameController.clear();
      _responseController.clear();
      _descController.clear();
    });
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context);
    final name = _nameController.text.trim().toLowerCase();
    if (!_namePattern.hasMatch(name)) {
      showFToast(
        context: context,
        title: Text(l10n.chatCommandsNameInvalidToast),
        icon: const Icon(FIcons.circleAlert),
      );
      return;
    }
    final response = _responseController.text.trim();
    if (response.isEmpty) {
      showFToast(
        context: context,
        title: Text(l10n.chatCommandsResponseRequiredToast),
        icon: const Icon(FIcons.circleAlert),
      );
      return;
    }

    setState(() => _submitting = true);
    try {
      final repo = ref.read(chatCommandRepositoryProvider);
      final editing = _editing;
      if (editing != null) {
        final res = await repo.update(
          id: editing.id,
          name: name,
          response: response,
          description: _descController.text.trim(),
        );
        if (!mounted) return;
        if (res.success) {
          showFToast(
            context: context,
            title: Text(l10n.chatCommandsUpdatedToast(name)),
            icon: const Icon(FIcons.check),
          );
          _resetForm();
          widget.onCreated();
        } else {
          showFToast(
            context: context,
            title: Text(l10n.chatCommandsUpdateFailedToast),
            icon: const Icon(FIcons.circleAlert),
          );
        }
      } else {
        final res = await repo.create(
          scope: widget.scope,
          name: name,
          response: response,
          description: _descController.text.trim(),
        );
        if (!mounted) return;
        if (res.success) {
          showFToast(
            context: context,
            title: Text(l10n.chatCommandsAddedToast(name)),
            icon: const Icon(FIcons.check),
          );
          _resetForm();
          widget.onCreated();
        } else {
          showFToast(
            context: context,
            title: Text(l10n.chatCommandsCreateFailedToast),
            icon: const Icon(FIcons.circleAlert),
          );
        }
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          widget.title,
          style: typography.lg.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 4),
        Text(
          widget.description,
          style: typography.sm.copyWith(color: colors.mutedForeground),
        ),
        const SizedBox(height: 12),
        if (widget.items.isEmpty)
          Text(
            l10n.chatCommandsEmpty,
            style: typography.sm.copyWith(color: colors.mutedForeground),
          )
        else
          for (final c in widget.items) ...[
            _CommandTile(
              command: c,
              onEdit: () => _startEdit(c),
              onDelete: () => widget.onDelete(c),
            ),
            const SizedBox(height: 8),
          ],
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            border: Border.all(color: colors.border),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_editing != null) ...[
                Text(
                  l10n.chatCommandsEditTitle(_editing!.name),
                  style: typography.sm.copyWith(color: colors.mutedForeground),
                ),
                const SizedBox(height: 8),
              ],
              FTextField(
                control: FTextFieldControl.managed(
                  controller: _nameController,
                ),
                label: Text(l10n.chatCommandsNameLabel),
                hint: l10n.chatCommandsNameHint,
                maxLength: 32,
              ),
              const SizedBox(height: 12),
              FTextField(
                control: FTextFieldControl.managed(
                  controller: _responseController,
                ),
                label: Text(l10n.chatCommandsResponseLabel),
                hint: l10n.chatCommandsResponseHint,
                maxLength: _maxResponse,
                maxLines: 2,
              ),
              const SizedBox(height: 12),
              FTextField(
                control: FTextFieldControl.managed(
                  controller: _descController,
                ),
                label: Text(l10n.chatCommandsDescriptionLabel),
                maxLength: _maxDescription,
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  if (_editing != null) ...[
                    FButton(
                      variant: FButtonVariant.ghost,
                      onPress: _submitting ? null : _resetForm,
                      child: Text(l10n.chatCommandsCancelEdit),
                    ),
                    const SizedBox(width: 8),
                  ],
                  FButton(
                    onPress: _submitting ? null : _submit,
                    child: _submitting
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(
                            _editing != null
                                ? l10n.chatCommandsSubmitEdit
                                : l10n.chatCommandsSubmit,
                          ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _CommandTile extends StatelessWidget {
  final ChatCommand command;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _CommandTile({
    required this.command,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        border: Border.all(color: colors.border),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '/${command.name}',
                  style: typography.sm.copyWith(
                    fontWeight: FontWeight.w600,
                    fontFamily: 'monospace',
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  command.response,
                  style: typography.sm.copyWith(color: colors.mutedForeground),
                ),
                if (command.description.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    command.description,
                    style: typography.xs.copyWith(color: colors.mutedForeground),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 8),
          FButton.icon(
            onPress: onEdit,
            child: Icon(
              FIcons.pencil,
              semanticLabel: l10n.chatCommandsEditAria,
            ),
          ),
          const SizedBox(width: 4),
          FButton.icon(
            onPress: onDelete,
            child: Icon(
              FIcons.x,
              semanticLabel: l10n.chatCommandsRemoveAria,
            ),
          ),
        ],
      ),
    );
  }
}

class _BuiltinsSection extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.chatCommandsBuiltinTitle,
          style: typography.lg.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 4),
        Text(
          l10n.chatCommandsBuiltinDescription,
          style: typography.sm.copyWith(color: colors.mutedForeground),
        ),
        const SizedBox(height: 12),
        for (final c in builtinChatCommands) ...[
          Row(
            children: [
              Expanded(
                child: Text(
                  c.usage,
                  style: typography.sm.copyWith(fontFamily: 'monospace'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  c.describe(l10n),
                  style: typography.xs.copyWith(color: colors.mutedForeground),
                  textAlign: TextAlign.right,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
        ],
      ],
    );
  }
}
