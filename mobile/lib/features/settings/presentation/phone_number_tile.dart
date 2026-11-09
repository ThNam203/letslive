/// The phone number row of the security settings (web `phone-number.tsx`):
/// shows the number or "Add a number"; a tap opens an editor that checks the
/// number ([isValidPhoneNumber]) before `PUT /user/me {phoneNumber}`. An
/// empty number removes it, as on web.
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:letslive_mobile/core/l10n/api_message.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/core/validation/field_limits.dart';
import 'package:letslive_mobile/features/auth/auth_providers.dart';
import 'package:letslive_mobile/features/settings/domain/phone_number.dart';

/// The phone number of the signed-in user, editable.
class PhoneNumberTile extends ConsumerWidget {
  const PhoneNumberTile({required this.phoneNumber, super.key});

  /// The saved number; null when there is none.
  final String? phoneNumber;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final number = phoneNumber;
    return ListTile(
      key: const ValueKey('phone-number'),
      contentPadding: EdgeInsets.zero,
      title: Text(l10n.settings_security_contact_phone),
      subtitle: Text(number ?? l10n.settings_security_contact_add_phone),
      trailing: const Icon(Icons.edit_outlined),
      onTap: () => showDialog<void>(
        context: context,
        builder: (_) => _PhoneNumberDialog(initial: number ?? ''),
      ),
    );
  }
}

class _PhoneNumberDialog extends ConsumerStatefulWidget {
  const _PhoneNumberDialog({required this.initial});

  final String initial;

  @override
  ConsumerState<_PhoneNumberDialog> createState() => _PhoneNumberDialogState();
}

class _PhoneNumberDialogState extends ConsumerState<_PhoneNumberDialog> {
  late final _controller = TextEditingController(text: widget.initial);
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    final navigator = Navigator.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final number = _controller.text.trim();
    if (!isValidPhoneNumber(number)) {
      setState(() => _error = l10n.settings_security_contact_phone_invalid);
      return;
    }
    if (number == widget.initial) {
      navigator.pop();
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    final res = await ref
        .read(sessionProvider.notifier)
        .updateProfile(phoneNumber: number);
    if (!mounted) return;
    if (!res.success) {
      setState(() {
        _busy = false;
        _error = apiErrorMessage(l10n, res);
      });
      return;
    }
    navigator.pop();
    messenger.showSnackBar(
      SnackBar(content: Text(l10n.settings_profile_update_success)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AlertDialog(
      title: Text(l10n.settings_security_contact_phone),
      content: TextField(
        key: const ValueKey('phone-number-field'),
        controller: _controller,
        enabled: !_busy,
        autofocus: true,
        keyboardType: TextInputType.phone,
        maxLength: FieldLimits.phoneNumberMax,
        decoration: InputDecoration(
          // Web's placeholder.
          hintText: '+1 234 567 8901',
          errorText: _error,
          errorMaxLines: 3,
        ),
        onSubmitted: (_) => _save(),
      ),
      actions: [
        TextButton(
          onPressed: _busy ? null : () => Navigator.of(context).pop(),
          child: Text(l10n.common_cancel),
        ),
        FilledButton(
          key: const ValueKey('phone-number-save'),
          onPressed: _busy ? null : _save,
          child: Text(l10n.common_save),
        ),
      ],
    );
  }
}
