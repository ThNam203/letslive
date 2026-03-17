import 'package:flutter_test/flutter_test.dart';

import 'package:letslive_mobile/app.dart';

void main() {
  testWidgets('App renders', (WidgetTester tester) async {
    await tester.pumpWidget(const App());
  });
}
