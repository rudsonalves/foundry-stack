enum AppMode { dev, staging, prod }

abstract final class AppEnv {
  static const _baseUrl = String.fromEnvironment('BASE_URL');
  static const _connectTimeout = int.fromEnvironment(
    'CONNECT_TIMEOUT',
    defaultValue: 10000,
  );
  static const _receiveTimeout = int.fromEnvironment(
    'RECEIVE_TIMEOUT',
    defaultValue: 10000,
  );
  static const _appMode = String.fromEnvironment(
    'APP_MODE',
    defaultValue: 'dev',
  );
  static const _authClientId = String.fromEnvironment(
    'AUTH_CLIENT_ID',
    defaultValue: 'foundry-stack-mobile',
  );
  static const _appAccessToken = String.fromEnvironment('APP_ACCESS_TOKEN');

  static String get authClientId {
    String clientId = _authClientId.trim().toLowerCase();

    // Validar AUTH_CLIENT_ID contra JWT_CLIENT_IDS no fluxo de build/deploy.
    if (clientId.isEmpty) {
      clientId = 'foundry-stack-mobile';
    }

    return clientId;
  }

  static String get appAccessToken => _appAccessToken;

  static String get baseUrl {
    final uri = Uri.tryParse(_baseUrl);
    if (_baseUrl.isEmpty || uri == null || !uri.hasScheme) {
      throw StateError('BASE_URL não definida ou inválida: $_baseUrl');
    }
    return _baseUrl;
  }

  static int get connectTimeout => _connectTimeout;
  static int get receiveTimeout => _receiveTimeout;

  static AppMode get mode => switch (_appMode.toLowerCase()) {
    'dev' => AppMode.dev,
    'staging' => AppMode.staging,
    'prod' => AppMode.prod,
    final value => throw StateError('APP_MODE inválido: $value'),
  };
}
