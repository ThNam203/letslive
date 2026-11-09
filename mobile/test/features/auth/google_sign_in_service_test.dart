import 'package:flutter_test/flutter_test.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:letslive_mobile/features/auth/data/google_sign_in_service.dart';
import 'package:mocktail/mocktail.dart';

class MockGoogleSignIn extends Mock implements GoogleSignIn {}

class MockGoogleAccount extends Mock implements GoogleSignInAccount {}

void main() {
  late MockGoogleSignIn signIn;
  late GoogleSignInService service;

  setUp(() {
    signIn = MockGoogleSignIn();
    service = GoogleSignInService(serverClientId: 'web-id', signIn: signIn);
    final account = MockGoogleAccount();
    when(() => account.authentication)
        .thenReturn(const GoogleSignInAuthentication(idToken: 'id-token'));
    when(() => signIn.authenticate()).thenAnswer((_) async => account);
  });

  test('a failed initialize is retried on the next attempt', () async {
    var calls = 0;
    when(() => signIn.initialize(serverClientId: 'web-id'))
        .thenAnswer((_) async {
          if (++calls == 1) throw StateError('play services unavailable');
        });

    await expectLater(service.obtainIdToken(), throwsStateError);
    expect(await service.obtainIdToken(), 'id-token');
    expect(calls, 2);
  });

  test('a successful initialize runs only once', () async {
    when(() => signIn.initialize(serverClientId: 'web-id'))
        .thenAnswer((_) async {});

    await service.obtainIdToken();
    await service.obtainIdToken();

    verify(() => signIn.initialize(serverClientId: 'web-id')).called(1);
  });

  test('a missing id token is an error', () async {
    when(() => signIn.initialize(serverClientId: 'web-id'))
        .thenAnswer((_) async {});
    final account = MockGoogleAccount();
    when(() => account.authentication)
        .thenReturn(const GoogleSignInAuthentication(idToken: null));
    when(() => signIn.authenticate()).thenAnswer((_) async => account);

    await expectLater(
      service.obtainIdToken(),
      throwsA(isA<GoogleSignInException>()),
    );
  });
}
