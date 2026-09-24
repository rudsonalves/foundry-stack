import '/core/resources/storage_keys.dart';
import '/core/result/result.dart';
import '/core/services/secure_storage/local_secure_storage.dart';
import '/domain/common/onboarding/models/onboarding_draft_load.dart';
import '/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'codec/onboarding_draft_codec.dart';
import 'onboarding_draft_repository.dart';

class OnboardingDraftRepositoryImpl implements OnboardingDraftRepository {
  final LocalSecureStorage _storage;
  final DateTime Function() _now;

  OnboardingDraftRepositoryImpl({
    required LocalSecureStorage storage,
  }) : _storage = storage,
       _now = DateTime.now;

  OnboardingDraftRepositoryImpl.withNow({
    required LocalSecureStorage storage,
    required DateTime Function() now,
  }) : _storage = storage,
       _now = now;

  @override
  AsyncResult<OnboardingDraftLoad> load() async {
    final result = await _storage.read(
      StorageKeys.onboardingRegistrationDraft,
    );

    if (result.isFailure) {
      final err = result.error!;

      if (err.code == AppErrorCode.storageNotFound) {
        return const Success(OnboardingDraftLoad.absent());
      }

      return Failure(err);
    }

    try {
      final draft = OnboardingDraftCodec.decode(result.value!);

      if (draft.isExpiredAt(_now())) {
        return await _deleteAndReturnAbsent();
      }

      return Success(OnboardingDraftLoad.present(draft));
    } on FormatException {
      final deleteResult = await _storage.delete(
        StorageKeys.onboardingRegistrationDraft,
      );

      if (deleteResult.isFailure) {
        return Failure(deleteResult.error!);
      }
    }

    return Success(OnboardingDraftLoad.absent());
  }

  @override
  AsyncResult<Unit> save(OnboardingRegistrationDraft draft) {
    return _storage.write(
      StorageKeys.onboardingRegistrationDraft,
      OnboardingDraftCodec.encode(draft),
    );
  }

  @override
  AsyncResult<Unit> delete() {
    return _storage.delete(
      StorageKeys.onboardingRegistrationDraft,
    );
  }

  AsyncResult<OnboardingDraftLoad> _deleteAndReturnAbsent() async {
    final result = await _storage.delete(
      StorageKeys.onboardingRegistrationDraft,
    );

    if (result.isFailure) {
      return Failure(result.error!);
    }

    return const Success(OnboardingDraftLoad.absent());
  }
}
