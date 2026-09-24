import 'package:auto_injector/auto_injector.dart';

import '/data/repositories/password_reset/password_reset_repository.dart';
import '/data/repositories/password_reset/password_reset_repository_impl.dart';
import '/data/services/apis/password_reset/password_reset_service.dart';
import '/domain/usecases/password_reset/password_reset_usecase.dart';

void registerPasswordResetBindings(AutoInjector injector) {
  injector
    ..add<PasswordResetService>(PasswordResetService.new)
    ..add<PasswordResetRepository>(PasswordResetRepositoryImpl.new)
    ..addSingleton<PasswordResetUsecase>(
      PasswordResetUsecase.new,
      config: BindConfig(
        onDispose: (usecase) => usecase.dispose(),
      ),
    );
}
