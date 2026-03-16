import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:forui/forui.dart';

import '../../../core/config/app_config.dart';
import '../../../core/constants/field_limits.dart';
import '../../../core/utils/api_error_localizer.dart';
import '../../../l10n/app_localizations.dart';
import '../../../models/vod.dart';
import '../../../providers.dart';
import '../../../shared/widgets/error_display.dart';
import '../../../shared/widgets/loading_indicator.dart';
import '../../vod/presentation/upload_vod_screen.dart';

class VodsSettingsScreen extends ConsumerStatefulWidget {
  const VodsSettingsScreen({super.key});

  @override
  ConsumerState<VodsSettingsScreen> createState() => _VodsSettingsScreenState();
}

class _VodsSettingsScreenState extends ConsumerState<VodsSettingsScreen> {
  List<Vod> _vods = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _fetchVods();
  }

  Future<void> _fetchVods() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = ref.read(vodRepositoryProvider);
      final response = await repo.getAuthorVods();
      if (!mounted) return;

      if (response.success) {
        setState(() {
          _vods = response.data ?? [];
          _isLoading = false;
        });
      } else {
        setState(() {
          _error = response.message;
          _isLoading = false;
        });
      }
    } on DioException catch (_) {
      if (mounted) {
        setState(() {
          _error = AppLocalizations.of(context).fetchError;
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _deleteVod(String vodId) async {
    try {
      final repo = ref.read(vodRepositoryProvider);
      final response = await repo.deleteVod(vodId);

      if (!mounted) return;

      if (response.success) {
        setState(() {
          _vods.removeWhere((v) => v.id == vodId);
        });
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.settingsVodsDeleteSuccess),
          icon: const Icon(FIcons.check),
        );
      } else {
        final errorMsg = getLocalizedApiMessage(context, response.key);
        showFToast(
          context: context,
          title: Text(errorMsg),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    } on DioException catch (_) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        showFToast(
          context: context,
          title: Text(l10n.fetchError),
          icon: const Icon(FIcons.circleAlert),
        );
      }
    }
  }

  void _showDeleteConfirmation(Vod vod) {
    final l10n = AppLocalizations.of(context);
    showFDialog(
      context: context,
      builder: (dialogContext, style, animation) {
        return FDialog(
          animation: animation,
          title: Text(l10n.settingsVodsDeleteTitle),
          body: Text(l10n.settingsVodsDeleteConfirm),
          actions: [
            FButton(
              variant: FButtonVariant.outline,
              onPress: () => Navigator.of(dialogContext).pop(),
              child: Text(l10n.cancel),
            ),
            FButton(
              variant: FButtonVariant.destructive,
              onPress: () {
                Navigator.of(dialogContext).pop();
                _deleteVod(vod.id);
              },
              child: Text(l10n.delete),
            ),
          ],
        );
      },
    );
  }

  void _showEditDialog(Vod vod) {
    final titleController = TextEditingController(text: vod.title);
    final descriptionController =
        TextEditingController(text: vod.description ?? '');
    final formKey = GlobalKey<FormState>();
    var visibility = vod.visibility;
    var isSaving = false;

    showFDialog(
      context: context,
      builder: (dialogContext, style, animation) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            final l10n = AppLocalizations.of(context);
            final colors = context.theme.colors;
            final typography = context.theme.typography;

            Future<void> save() async {
              if (!formKey.currentState!.validate()) return;
              setDialogState(() => isSaving = true);

              try {
                final repo = ref.read(vodRepositoryProvider);
                final response = await repo.updateVod(
                  vodId: vod.id,
                  title: titleController.text.trim(),
                  description: descriptionController.text.trim(),
                  visibility: visibility,
                );

                if (!context.mounted) return;

                if (response.success) {
                  final updatedVod = vod.copyWith(
                    title: titleController.text.trim(),
                    description: descriptionController.text.trim(),
                    visibility: visibility,
                  );
                  setState(() {
                    final index =
                        _vods.indexWhere((v) => v.id == updatedVod.id);
                    if (index != -1) {
                      _vods[index] = updatedVod;
                    }
                  });
                  showFToast(
                    context: context,
                    title: Text(l10n.settingsVodsUpdateSuccess),
                    icon: const Icon(FIcons.check),
                  );
                  Navigator.of(context).pop();
                } else {
                  final errorMsg =
                      getLocalizedApiMessage(context, response.key);
                  showFToast(
                    context: context,
                    title: Text(errorMsg),
                    icon: const Icon(FIcons.circleAlert),
                  );
                }
              } on DioException catch (_) {
                if (context.mounted) {
                  showFToast(
                    context: context,
                    title: Text(l10n.fetchError),
                    icon: const Icon(FIcons.circleAlert),
                  );
                }
              } finally {
                if (context.mounted) setDialogState(() => isSaving = false);
              }
            }

            return FDialog(
              animation: animation,
              title: Text(l10n.settingsVodsEditTitle),
              body: Form(
                key: formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    FTextFormField(
                      control: FTextFieldControl.managed(
                        controller: titleController,
                      ),
                      label: Text(l10n.settingsStreamStreamTitle),
                      maxLength: FieldLimits.vodTitleMaxLength,
                      autovalidateMode: AutovalidateMode.onUserInteraction,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return l10n.errorTitleRequired;
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 12),
                    FTextFormField(
                      control: FTextFieldControl.managed(
                        controller: descriptionController,
                      ),
                      label: Text(l10n.settingsStreamStreamDescription),
                      maxLength: FieldLimits.vodDescriptionMaxLength,
                      maxLines: 3,
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Text(
                          l10n.settingsVodsVisibility,
                          style: typography.sm
                              .copyWith(fontWeight: FontWeight.w600),
                        ),
                        const Spacer(),
                        if (visibility == 'public')
                          FButton(
                            onPress: () {},
                            prefix: Icon(FIcons.eye,
                                size: 14,
                                color: colors.primaryForeground),
                            child: Text(l10n.settingsVodsPublic),
                          )
                        else
                          FButton(
                            variant: FButtonVariant.outline,
                            onPress: () {
                              setDialogState(() => visibility = 'public');
                            },
                            prefix: Icon(FIcons.eye,
                                size: 14, color: colors.primary),
                            child: Text(l10n.settingsVodsPublic),
                          ),
                        const SizedBox(width: 8),
                        if (visibility == 'private')
                          FButton(
                            onPress: () {},
                            prefix: Icon(FIcons.eyeOff,
                                size: 14,
                                color: colors.primaryForeground),
                            child: Text(l10n.settingsVodsPrivate),
                          )
                        else
                          FButton(
                            variant: FButtonVariant.outline,
                            onPress: () {
                              setDialogState(() => visibility = 'private');
                            },
                            prefix: Icon(FIcons.eyeOff,
                                size: 14,
                                color: colors.mutedForeground),
                            child: Text(l10n.settingsVodsPrivate),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              actions: [
                FButton(
                  variant: FButtonVariant.outline,
                  onPress: () => Navigator.of(context).pop(),
                  child: Text(l10n.cancel),
                ),
                FButton(
                  onPress: isSaving ? null : save,
                  child: isSaving
                      ? const SizedBox(
                          height: 16,
                          width: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(l10n.save),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return FScaffold(
      header: FHeader.nested(
        title: Text(l10n.settingsVodsTitle),
        suffixes: [
          FButton.icon(
            onPress: () async {
              final result = await Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => const UploadVodScreen(),
                ),
              );
              if (result == true) _fetchVods();
            },
            child: const Icon(FIcons.upload),
          ),
        ],
      ),
      child: _buildContent(l10n),
    );
  }

  Widget _buildContent(AppLocalizations l10n) {
    if (_isLoading) {
      return LoadingIndicator(message: l10n.loading);
    }

    if (_error != null) {
      return ErrorDisplay(
        title: l10n.errorGeneralTitle,
        message: _error,
        onRetry: _fetchVods,
      );
    }

    if (_vods.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(FIcons.film,
                  size: 48,
                  color: context.theme.colors.mutedForeground),
              const SizedBox(height: 16),
              Text(
                l10n.settingsVodsNoVods,
                style: context.theme.typography.lg
                    .copyWith(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: 8),
              Text(
                l10n.settingsVodsNoVodsDescription,
                style: context.theme.typography.sm.copyWith(
                  color: context.theme.colors.mutedForeground,
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _fetchVods,
      child: ListView.builder(
        padding: const EdgeInsets.all(12),
        itemCount: _vods.length,
        itemBuilder: (context, index) {
          final vod = _vods[index];
          return _VodManageCard(
            vod: vod,
            onEdit: () => _showEditDialog(vod),
            onDelete: () => _showDeleteConfirmation(vod),
          );
        },
      ),
    );
  }
}

class _VodManageCard extends StatelessWidget {
  final Vod vod;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _VodManageCard({
    required this.vod,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.theme.colors;
    final typography = context.theme.typography;
    final l10n = AppLocalizations.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.card,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: colors.border, width: 0.5),
        ),
        child: Row(
          children: [
            // Thumbnail
            ClipRRect(
              borderRadius:
                  const BorderRadius.horizontal(left: Radius.circular(12)),
              child: SizedBox(
                width: 120,
                height: 80,
                child: vod.thumbnailUrl != null
                    ? CachedNetworkImage(
                        imageUrl: '${AppConfig.apiUrl}/${vod.thumbnailUrl}',
                        fit: BoxFit.cover,
                        errorWidget: (_, _, _) => ColoredBox(
                          color: colors.muted,
                          child: const Center(child: Icon(FIcons.film)),
                        ),
                      )
                    : ColoredBox(
                        color: colors.muted,
                        child: const Center(child: Icon(FIcons.film)),
                      ),
              ),
            ),
            // Info
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      vod.title,
                      style:
                          typography.sm.copyWith(fontWeight: FontWeight.w600),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: vod.isPublic
                                ? colors.primary.withValues(alpha: 0.1)
                                : colors.muted,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            vod.isPublic
                                ? l10n.settingsVodsPublic
                                : l10n.settingsVodsPrivate,
                            style: typography.xs.copyWith(
                              color: vod.isPublic
                                  ? colors.primary
                                  : colors.mutedForeground,
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          l10n.homeViewCount(vod.viewCount),
                          style: typography.xs
                              .copyWith(color: colors.mutedForeground),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            // Actions
            Column(
              children: [
                IconButton(
                  icon: const Icon(FIcons.squarePen, size: 18),
                  onPressed: onEdit,
                ),
                IconButton(
                  icon: Icon(FIcons.trash, size: 18, color: colors.destructive),
                  onPressed: onDelete,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

