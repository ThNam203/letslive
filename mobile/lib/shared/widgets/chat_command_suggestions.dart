import 'package:flutter/material.dart';
import 'package:forui/forui.dart';

import '../../core/chat_parser/command.dart';
import '../../l10n/app_localizations.dart';

class ChatCommandSuggestionsList extends StatelessWidget {
  final List<ChatCommandSuggestion> suggestions;
  final ValueChanged<ChatCommandSuggestion> onPick;

  const ChatCommandSuggestionsList({
    super.key,
    required this.suggestions,
    required this.onPick,
  });

  @override
  Widget build(BuildContext context) {
    if (suggestions.isEmpty) return const SizedBox.shrink();
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Container(
      constraints: const BoxConstraints(maxHeight: 240),
      decoration: BoxDecoration(
        color: colors.background,
        border: Border.all(color: colors.border),
        borderRadius: BorderRadius.circular(8),
      ),
      child: ListView.builder(
        shrinkWrap: true,
        padding: EdgeInsets.zero,
        itemCount: suggestions.length,
        itemBuilder: (context, index) {
          final s = suggestions[index];
          return InkWell(
            onTap: () => onPick(s),
            child: Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 8,
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          s.usage,
                          style: typography.sm.copyWith(
                            fontWeight: FontWeight.w600,
                            fontFamily: 'monospace',
                          ),
                        ),
                        Text(
                          s.description,
                          style: typography.xs.copyWith(
                            color: colors.mutedForeground,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _sourceLabel(l10n, s.source),
                    style: typography.xs.copyWith(
                      color: colors.mutedForeground,
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  String _sourceLabel(AppLocalizations l10n, ChatCommandSource source) {
    return switch (source) {
      ChatCommandSource.builtin => l10n.chatCommandsSourceBuiltin,
      ChatCommandSource.user => l10n.chatCommandsSourceUser,
      ChatCommandSource.channel => l10n.chatCommandsSourceChannel,
    };
  }
}
