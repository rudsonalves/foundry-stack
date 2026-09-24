import '../services/auth_token_store/auth_token_store.dart';
import '../services/client_http/client/rest_client.dart';
import '../services/secure_storage/local_secure_storage.dart';

class CoreDependencies {
  final RestClient restClient;
  final LocalSecureStorage localSecureStorage;
  final AuthTokenStore authTokenStore;

  const CoreDependencies({
    required this.restClient,
    required this.localSecureStorage,
    required this.authTokenStore,
  });
}
