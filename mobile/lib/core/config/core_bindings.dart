import 'package:auto_injector/auto_injector.dart';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../resources/app_env.dart';
import '../services/auth_token_store/auth_token_store.dart';
import '../services/auth_token_store/secure_auth_token_store.dart';
import '../services/client_http/client/rest_client.dart';
import '../services/client_http/dio/dio_factory.dart';
import '../services/client_http/dio/dio_rest_client.dart';
import '../services/client_http/interceptors/auth/auth_interceptor.dart';
import '../services/secure_storage/flutter_local_secure_storage.dart';
import '../services/secure_storage/local_secure_storage.dart';

void registerCoreBindings(
  AutoInjector injector, {
  String? baseUrl,
}) {
  injector
    ..addSingleton<FlutterSecureStorage>(FlutterSecureStorage.new)
    ..addSingleton<LocalSecureStorage>(
      () => FlutterLocalSecureStorageLocalStorage(
        storage: injector.get<FlutterSecureStorage>(),
      ),
    )
    ..addSingleton<AuthTokenStore>(
      () => SecureAuthTokenStore(
        injector.get<LocalSecureStorage>(),
      ),
    )
    ..addSingleton<Dio>(() => DioFactory.create(baseUrl: baseUrl))
    ..addSingleton<AuthInterceptor>(
      () => AuthInterceptor(
        authDio: injector.get<Dio>(),
        authTokenStore: injector.get<AuthTokenStore>(),
        baseUrl: baseUrl ?? AppEnv.baseUrl,
      ),
    )
    ..addSingleton<RestClient>(() {
      final dio = injector.get<Dio>();
      dio.interceptors.add(injector.get<AuthInterceptor>());
      return DioRestClient(dio);
    });
}
