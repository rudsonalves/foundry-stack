import 'package:foundry_stack_mobile/core/resources/storage_keys.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/core/services/secure_storage/local_secure_storage.dart';
import 'package:foundry_stack_mobile/data/repositories/onboarding/onboarding_draft_repository_impl.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final now = DateTime.utc(2026, 9, 18, 12);
  late _FakeLocalSecureStorage storage;
  late OnboardingDraftRepositoryImpl repository;

  setUp(() {
    storage = _FakeLocalSecureStorage();
    repository = OnboardingDraftRepositoryImpl.withNow(
      storage: storage,
      now: () => now,
    );
  });

  test('saves and loads a draft using its exclusive key', () async {
    final draft = OnboardingRegistrationDraft(
      step: OnboardingStep.email,
      name: 'Ada',
      expiresAt: DateTime.utc(2026, 9, 19),
    );

    final saveResult = await repository.save(draft);
    final loadResult = await repository.load();

    expect(saveResult.isSuccess, isTrue);
    expect(storage.writeKeys, [StorageKeys.onboardingRegistrationDraft]);
    expect(loadResult.isSuccess, isTrue);
    expect(loadResult.value?.hasDraft, isTrue);
    expect(loadResult.value?.draft?.step, OnboardingStep.email);
    expect(loadResult.value?.draft?.name, 'Ada');
    expect(storage.deleteKeys, isEmpty);
  });

  test('deletes a draft at the exact expiration instant', () async {
    storage.values['auth_token'] = 'access-token';
    await repository.save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.userName,
        expiresAt: now,
      ),
    );

    final result = await repository.load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.hasDraft, isFalse);
    expect(storage.deleteKeys, [StorageKeys.onboardingRegistrationDraft]);
    expect(storage.values['auth_token'], 'access-token');
    expect(storage.deleteAllCalls, 0);
  });

  test('deletes a draft after its expiration', () async {
    await repository.save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.userName,
        expiresAt: now.subtract(const Duration(microseconds: 1)),
      ),
    );

    final result = await repository.load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.hasDraft, isFalse);
    expect(storage.deleteKeys, [StorageKeys.onboardingRegistrationDraft]);
  });

  test('preserves cleanup failure for an expired draft', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'delete failed',
    );
    await repository.save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.userName,
        expiresAt: now,
      ),
    );
    storage.deleteResult = const Failure(error);

    final result = await repository.load();

    expect(identical(result.error, error), isTrue);
    expect(
      storage.values,
      contains(StorageKeys.onboardingRegistrationDraft),
    );
  });

  test('returns absence when the draft key does not exist', () async {
    final result = await repository.load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.hasDraft, isFalse);
    expect(storage.deleteKeys, isEmpty);
  });

  test('deletes corrupted content and returns absence', () async {
    storage.values[StorageKeys.onboardingRegistrationDraft] = 'not-json';
    storage.values['auth_token'] = 'access-token';

    final result = await repository.load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.hasDraft, isFalse);
    expect(storage.deleteKeys, [StorageKeys.onboardingRegistrationDraft]);
    expect(storage.values['auth_token'], 'access-token');
    expect(storage.deleteAllCalls, 0);
  });

  test('deletes an incompatible version and returns absence', () async {
    storage.values[StorageKeys.onboardingRegistrationDraft] =
        '{"schema_version":2}';

    final result = await repository.load();

    expect(result.isSuccess, isTrue);
    expect(result.value?.hasDraft, isFalse);
    expect(storage.deleteKeys, [StorageKeys.onboardingRegistrationDraft]);
  });

  test('preserves a storage read failure', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'read failed',
    );
    storage.readResult = const Failure(error);

    final result = await repository.load();

    expect(identical(result.error, error), isTrue);
  });

  test('preserves a storage write failure', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'write failed',
    );
    storage.writeResult = const Failure(error);

    final result = await repository.save(
      OnboardingRegistrationDraft(
        step: OnboardingStep.userName,
        expiresAt: DateTime.utc(2026, 9, 19),
      ),
    );

    expect(identical(result.error, error), isTrue);
  });

  test('preserves cleanup failure for corrupted content', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'delete failed',
    );
    storage.values[StorageKeys.onboardingRegistrationDraft] = 'not-json';
    storage.deleteResult = const Failure(error);

    final result = await repository.load();

    expect(identical(result.error, error), isTrue);
  });

  test('deletes only the onboarding draft key', () async {
    storage.values[StorageKeys.onboardingRegistrationDraft] = 'draft';
    storage.values['refresh_token'] = 'refresh-token';

    final result = await repository.delete();

    expect(result.isSuccess, isTrue);
    expect(storage.deleteKeys, [StorageKeys.onboardingRegistrationDraft]);
    expect(storage.values['refresh_token'], 'refresh-token');
    expect(storage.deleteAllCalls, 0);
  });
}

class _FakeLocalSecureStorage implements LocalSecureStorage {
  final values = <String, String>{};
  final readKeys = <String>[];
  final writeKeys = <String>[];
  final deleteKeys = <String>[];
  Result<String>? readResult;
  Result<Unit>? writeResult;
  Result<Unit>? deleteResult;
  int deleteAllCalls = 0;

  @override
  AsyncResult<String> read(String key) async {
    readKeys.add(key);
    final configured = readResult;
    if (configured != null) return configured;

    final value = values[key];
    if (value == null) {
      return const Failure(
        AppError(
          code: AppErrorCode.storageNotFound,
          message: 'key not found',
        ),
      );
    }
    return Success(value);
  }

  @override
  AsyncResult<Unit> write(String key, String value) async {
    writeKeys.add(key);
    final result = writeResult ?? const Success(unit);
    if (result.isSuccess) values[key] = value;
    return result;
  }

  @override
  AsyncResult<Unit> delete(String key) async {
    deleteKeys.add(key);
    final result = deleteResult ?? const Success(unit);
    if (result.isSuccess) values.remove(key);
    return result;
  }

  @override
  AsyncResult<Unit> deleteAll() async {
    deleteAllCalls++;
    values.clear();
    return const Success(unit);
  }

  @override
  AsyncResult<List<String>> keysWithPrefix(String pattern) async {
    return Success(
      values.keys.where((key) => key.startsWith(pattern)).toList(),
    );
  }
}
