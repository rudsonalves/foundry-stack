import '/core/result/command.dart';
import '/domain/usecases/bootstrap/bootstrap_usecase.dart';
import '/domain/usecases/bootstrap/models/bootstrap_destination.dart';

class SplashViewmodel {
  final BootstrapUsecase _bootstrapUsecase;

  SplashViewmodel(this._bootstrapUsecase) {
    initialize = Command0(_bootstrapUsecase.initialize);
  }

  late final Command0<BootstrapDestination> initialize;
}
