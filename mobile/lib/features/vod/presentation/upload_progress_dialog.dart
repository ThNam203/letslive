import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../l10n/app_localizations.dart';
import '../../../providers.dart';

enum UploadStatus { uploading, processing, complete, failed }

class UploadProgressDialog extends ConsumerStatefulWidget {
  final Animation<double> animation;
  final File videoFile;
  final String title;
  final String description;
  final String visibility;

  const UploadProgressDialog({
    super.key,
    required this.animation,
    required this.videoFile,
    required this.title,
    required this.description,
    required this.visibility,
  });

  @override
  ConsumerState<UploadProgressDialog> createState() =>
      _UploadProgressDialogState();
}

class _UploadProgressDialogState extends ConsumerState<UploadProgressDialog> {
  double _progress = 0.0;
  int _bytesSent = 0;
  int _totalBytes = 0;
  UploadStatus _status = UploadStatus.uploading;
  CancelToken? _cancelToken;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _startUpload();
  }

  Future<void> _startUpload() async {
    _cancelToken = CancelToken();
    try {
      final vodRepo = ref.read(vodRepositoryProvider);
      final response = await vodRepo.uploadVod(
        videoFile: widget.videoFile,
        title: widget.title,
        description: widget.description,
        visibility: widget.visibility,
        onSendProgress: (sent, total) {
          if (!mounted) return;
          setState(() {
            _bytesSent = sent;
            _totalBytes = total;
            _progress = total > 0 ? sent / total : 0;
          });
        },
        cancelToken: _cancelToken,
      );
      if (!mounted) return;
      setState(() {
        _status = response.success
            ? UploadStatus.processing
            : UploadStatus.failed;
        if (!response.success) {
          _errorMessage = response.message;
        }
      });

      if (response.success) {
        await Future.delayed(const Duration(seconds: 2));
        if (mounted) {
          setState(() => _status = UploadStatus.complete);
          await Future.delayed(const Duration(seconds: 1));
          if (mounted) Navigator.of(context).pop(true);
        }
      }
    } on DioException catch (e) {
      if (e.type == DioExceptionType.cancel) return;
      if (!mounted) return;
      final l10n = AppLocalizations.of(context);
      setState(() {
        _status = UploadStatus.failed;
        _errorMessage = e.message ?? l10n.uploadFailedGeneric;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _status = UploadStatus.failed;
        _errorMessage = e.toString();
      });
    }
  }

  void _cancel() {
    _cancelToken?.cancel();
    Navigator.of(context).pop(false);
  }

  void _retry() {
    setState(() {
      _progress = 0.0;
      _bytesSent = 0;
      _totalBytes = 0;
      _status = UploadStatus.uploading;
      _errorMessage = null;
    });
    _startUpload();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.theme.colors;
    final typography = context.theme.typography;

    return FDialog(
      animation: widget.animation,
      title: Text(_getTitle(l10n)),
      body: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_status == UploadStatus.uploading) ...[
            LinearProgressIndicator(value: _progress),
            const SizedBox(height: 12),
            Text(
              '${_formatBytes(_bytesSent)} / ${_formatBytes(_totalBytes)}',
              style: typography.xs.copyWith(color: colors.mutedForeground),
            ),
            const SizedBox(height: 4),
            Text(
              '${(_progress * 100).toStringAsFixed(1)}%',
              style: typography.sm.copyWith(fontWeight: FontWeight.bold),
            ),
          ],
          if (_status == UploadStatus.processing) ...[
            const LinearProgressIndicator(),
            const SizedBox(height: 12),
            Text(l10n.uploadProcessingOnServer, style: typography.sm),
          ],
          if (_status == UploadStatus.complete) ...[
            Icon(FIcons.circleCheck, color: colors.primary, size: 48),
            const SizedBox(height: 8),
            Text(l10n.uploadComplete, style: typography.sm),
          ],
          if (_status == UploadStatus.failed) ...[
            Icon(FIcons.circleAlert, color: colors.destructive, size: 48),
            const SizedBox(height: 8),
            Text(
              _errorMessage ?? l10n.uploadFailedGeneric,
              style: typography.sm,
              textAlign: TextAlign.center,
            ),
          ],
        ],
      ),
      actions: [
        if (_status == UploadStatus.uploading)
          FButton(
            variant: FButtonVariant.outline,
            onPress: _cancel,
            child: Text(l10n.cancel),
          ),
        if (_status == UploadStatus.failed) ...[
          FButton(
            variant: FButtonVariant.outline,
            onPress: () => Navigator.of(context).pop(false),
            child: Text(l10n.uploadClose),
          ),
          FButton(onPress: _retry, child: Text(l10n.retry)),
        ],
        if (_status == UploadStatus.processing)
          FButton(
            variant: FButtonVariant.outline,
            onPress: () => Navigator.of(context).pop(true),
            child: Text(l10n.uploadDismiss),
          ),
      ],
    );
  }

  String _getTitle(AppLocalizations l10n) {
    switch (_status) {
      case UploadStatus.uploading:
        return l10n.uploadUploading;
      case UploadStatus.processing:
        return l10n.uploadProcessing;
      case UploadStatus.complete:
        return l10n.uploadDone;
      case UploadStatus.failed:
        return l10n.uploadFailed;
    }
  }

  String _formatBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1048576) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1073741824) {
      return '${(bytes / 1048576).toStringAsFixed(1)} MB';
    }
    return '${(bytes / 1073741824).toStringAsFixed(1)} GB';
  }
}
