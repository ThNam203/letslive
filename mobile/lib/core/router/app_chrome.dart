/// The bars of the signed-in shell (each tab's top bar and the bottom
/// navigation) that slide away while the user scrolls a tab's content down
/// and come back on scrolling up, like the YouTube app.
///
/// [AppShell] owns the state and listens to the scrolls of its tabs;
/// [TabScaffold] is the screen frame a tab root uses so its top bar takes
/// part.
library;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';

/// How long a bar takes to slide in or out.
const Duration _slideDuration = Duration(milliseconds: 200);

/// Publishes whether the shell's bars are showing to the tab screens below.
class AppChrome extends InheritedNotifier<ValueNotifier<bool>> {
  const AppChrome({
    required ValueNotifier<bool> visible,
    required super.child,
    super.key,
  }) : super(notifier: visible);

  /// Whether the bars show. True outside a shell (a screen pumped alone in a
  /// test, or one pushed over the tabs).
  static bool visibleOf(BuildContext context) =>
      context
          .dependOnInheritedWidgetOfExactType<AppChrome>()
          ?.notifier
          ?.value ??
      true;
}

/// Updates [visible] from one scroll event of a tab's content:
/// - a vertical drag towards the end of the list hides the bars, one back
///   towards the start shows them;
/// - reaching the top shows them, so a list too short to scroll back never
///   leaves them hidden.
///
/// Horizontal lists (card strips, tab pages) are ignored. Returns false so
/// the event keeps bubbling.
bool updateChromeOnScroll(
  ValueNotifier<bool> visible,
  ScrollNotification notification,
) {
  final metrics = notification.metrics;
  if (metrics.axis != Axis.vertical) return false;
  if (notification is UserScrollNotification) {
    switch (notification.direction) {
      case ScrollDirection.reverse:
        visible.value = false;
      case ScrollDirection.forward:
        visible.value = true;
      case ScrollDirection.idle:
        break;
    }
  } else if (notification is ScrollUpdateNotification &&
      metrics.extentBefore <= 0) {
    visible.value = true;
  }
  return false;
}

/// [child] cut to a height between its own and zero: [visible] slides it
/// in, otherwise it slides out towards [alignment]'s opposite edge
/// (`Alignment.bottomCenter` for a top bar, which slides up and away).
class CollapsibleBar extends StatelessWidget {
  const CollapsibleBar({
    required this.visible,
    required this.alignment,
    required this.child,
    super.key,
  });

  final bool visible;
  final Alignment alignment;
  final Widget child;

  @override
  Widget build(BuildContext context) => ClipRect(
    child: AnimatedAlign(
      duration: _slideDuration,
      curve: Curves.easeOut,
      alignment: alignment,
      heightFactor: visible ? 1 : 0,
      child: child,
    ),
  );
}

/// The frame of a tab's root screen: [appBar] on top, hidden together with
/// the bottom bar (see [AppChrome]), then [body].
///
/// The bar sits in the body rather than in `Scaffold.appBar` so its height
/// can animate. The status bar area stays clear and keeps icons that suit
/// the theme while the bar is away.
class TabScaffold extends StatelessWidget {
  const TabScaffold({required this.appBar, required this.body, super.key});

  final Widget appBar;
  final Widget body;

  @override
  Widget build(BuildContext context) => ThemedStatusBar(
    child: Scaffold(
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            CollapsibleBar(
              visible: AppChrome.visibleOf(context),
              alignment: Alignment.bottomCenter,
              child: appBar,
            ),
            Expanded(child: body),
          ],
        ),
      ),
    ),
  );
}

/// Status bar icons that read on the theme's background, for screens with no
/// `AppBar` in `Scaffold.appBar` (which would otherwise set them).
class ThemedStatusBar extends StatelessWidget {
  const ThemedStatusBar({required this.child, super.key});

  final Widget child;

  @override
  Widget build(BuildContext context) => AnnotatedRegion<SystemUiOverlayStyle>(
    value: Theme.of(context).brightness == Brightness.dark
        ? SystemUiOverlayStyle.light
        : SystemUiOverlayStyle.dark,
    child: child,
  );
}
