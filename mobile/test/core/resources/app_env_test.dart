import 'package:flutter_test/flutter_test.dart';
import 'package:foundry_stack_mobile/core/resources/app_env.dart';

const expectedAuthClientId = String.fromEnvironment(
  'EXPECTED_AUTH_CLIENT_ID',
  defaultValue: 'foundry-stack-mobile',
);

void main() {
  group('AppEnv', () {
    test('exposes the expected authClientId value', () {
      expect(AppEnv.authClientId, expectedAuthClientId);
    });

    test('exposes the expected default values for other settings', () {
      expect(AppEnv.connectTimeout, 10000);
      expect(AppEnv.receiveTimeout, 10000);
      expect(AppEnv.mode, AppMode.dev);
    });

    test('throws when BASE_URL is not defined', () {
      expect(() => AppEnv.baseUrl, throwsStateError);
    });
  });
}
