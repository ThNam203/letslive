import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:letslive_mobile/core/l10n/generated/app_localizations.dart';
import 'package:letslive_mobile/features/auth/presentation/auth_widgets.dart';

// Both dialog actions share one contract: they resolve to the localized
// error message to show, or null on success.

/// Submits the entered [code] (creates the account).
typedef OtpVerifyAction = Future<String?> Function(String code);

/// Mails a new code.
typedef OtpResendAction = Future<String?> Function();

/// Asks for the 6-digit signup code (web `SignupForm.tsx` dialog). [onVerify]
/// creates the account; [onResend] mails a new code, allowed every
/// [resendCooldown] like web's `ResendOtpButton`.
class OtpDialog extends StatefulWidget {
  const OtpDialog({
    required this.email,
    required this.onVerify,
    required this.onResend,
    super.key,
  });

  /// Web `initialCountdown={60}`.
  static const Duration resendCooldown = Duration(seconds: 60);

  /// Length of the code (`otpCode` is `min=6,max=6` in the backend DTO).
  static const int codeLength = 6;

  final String email;
  final OtpVerifyAction onVerify;
  final OtpResendAction onResend;

  @override
  State<OtpDialog> createState() => _OtpDialogState();
}

class _OtpDialogState extends State<OtpDialog> {
  final _code = TextEditingController();
  Timer? _ticker;
  int _secondsLeft = OtpDialog.resendCooldown.inSeconds;
  bool _busy = false;
  bool _resending = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _startCooldown();
  }

  @override
  void dispose() {
    _ticker?.cancel();
    _code.dispose();
    super.dispose();
  }

  void _startCooldown() {
    _ticker?.cancel();
    _secondsLeft = OtpDialog.resendCooldown.inSeconds;
    _ticker = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_secondsLeft <= 1) timer.cancel();
      setState(() => _secondsLeft--);
    });
  }

  Future<void> _verify() async {
    final l10n = AppLocalizations.of(context);
    final code = _code.text;
    if (code.length != OtpDialog.codeLength) {
      setState(() => _error = l10n.error_otp_required);
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    final error = await widget.onVerify(code);
    if (!mounted) return;
    if (error == null) {
      Navigator.of(context).pop();
      return;
    }
    // The code is spent whatever the reason, as on web.
    _code.clear();
    setState(() {
      _busy = false;
      _error = error;
    });
  }

  Future<void> _resend() async {
    setState(() => _resending = true);
    final error = await widget.onResend();
    if (!mounted) return;
    setState(() {
      _resending = false;
      _error = error;
      if (error == null) _startCooldown();
    });
  }

  String _resendLabel(AppLocalizations l10n) {
    if (_resending) return l10n.auth_sending_otp;
    if (_secondsLeft > 0) return l10n.auth_otp_resend_count_down(_secondsLeft);
    return l10n.auth_resend_otp;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AlertDialog(
      title: Text(l10n.auth_enter_verification_code),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            '${l10n.auth_otp_dialog_description_part_1} ${widget.email}'
            '${l10n.auth_otp_dialog_description_part_2}',
          ),
          const SizedBox(height: 16),
          TextField(
            key: const ValueKey('otp'),
            controller: _code,
            enabled: !_busy,
            autofocus: true,
            keyboardType: TextInputType.number,
            autofillHints: const [AutofillHints.oneTimeCode],
            maxLength: OtpDialog.codeLength,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 24, letterSpacing: 8),
            decoration: InputDecoration(errorText: _error, counterText: ''),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: _secondsLeft > 0 || _resending ? null : _resend,
          child: Text(_resendLabel(l10n)),
        ),
        SubmitButton(
          key: const ValueKey('otp-submit'),
          label: l10n.auth_verify_otp,
          busy: _busy,
          onPressed: _verify,
          expanded: false,
        ),
      ],
    );
  }
}
