import '/core/services/auth_token_store/auth_token_store.dart';
import '/core/services/client_http/client/rest_client.dart';
import '../../core/services/secure_storage/local_secure_storage.dart';

class AppDependencies {
  final RestClient restClient;
  final AuthTokenStore authTokenStore;
  final LocalSecureStorage localSecureStorage;

  const AppDependencies({
    required this.restClient,
    required this.authTokenStore,
    required this.localSecureStorage,
  });
}
