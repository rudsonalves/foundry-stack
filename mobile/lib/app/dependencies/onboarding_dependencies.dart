import '/core/services/client_http/client/rest_client.dart';
import '/core/services/secure_storage/local_secure_storage.dart';

class OnboardingDependencies {
  final RestClient restClient;
  final LocalSecureStorage localSecureStorage;

  const OnboardingDependencies({
    required this.restClient,
    required this.localSecureStorage,
  });
}
