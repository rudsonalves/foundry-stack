import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '/core/result/result.dart';
import '/core/services/logging/console_log.dart';
import 'local_secure_storage.dart';
import 'storage_app_error.dart';

class FlutterLocalSecureStorageLocalStorage implements LocalSecureStorage {
  final FlutterSecureStorage storage;

  FlutterLocalSecureStorageLocalStorage({required this.storage});

  final _log = ConsoleLog('FlutterLocalSecureStorageLocalStorage');

  @override
  AsyncResult<List<String>> keysWithPrefix(String pattern) async {
    try {
      final allKeys = await storage.readAll();
      final filteredKeys = allKeys.keys
          .where((key) => key.startsWith(pattern))
          .toList();
      return Success(filteredKeys);
    } on AppError catch (error) {
      return Failure(error);
    } catch (err, stack) {
      _log.error('[readKeys]: $err', error: err, stack: stack);
      return Failure(
        StorageAppError.storage(
          message: 'Failed to read keys with prefix: $pattern',
          details: err,
        ),
      );
    }
  }

  @override
  AsyncResult<Unit> delete(String key) async {
    try {
      await storage.delete(key: key);
      return Success(unit);
    } on AppError catch (error) {
      return Failure(error);
    } catch (err, stack) {
      _log.error('[delete]: $err', error: err, stack: stack);
      return Failure(
        StorageAppError.storage(
          message: 'Failed to delete key: $key',
          details: err,
        ),
      );
    }
  }

  @override
  AsyncResult<Unit> deleteAll() async {
    try {
      await storage.deleteAll();
      return Success(unit);
    } on AppError catch (error) {
      return Failure(error);
    } catch (err, stack) {
      _log.error('[deleteAll]: $err', error: err, stack: stack);
      return Failure(
        StorageAppError.storage(
          message: 'Failed to delete all keys',
          details: err,
        ),
      );
    }
  }

  @override
  AsyncResult<String> read(String key) async {
    try {
      final value = await storage.read(key: key);
      if (value == null) {
        return Failure(StorageAppError.notFound(key));
      }
      return Success(value);
    } on AppError catch (error) {
      return Failure(error);
    } catch (err, stack) {
      _log.error('[read]: $err', error: err, stack: stack);
      return Failure(
        StorageAppError.storage(
          message: 'Failed to read key: $key',
          details: err,
        ),
      );
    }
  }

  @override
  AsyncResult<Unit> write(String key, String value) async {
    try {
      await storage.write(key: key, value: value);
      return Success(unit);
    } on AppError catch (error) {
      return Failure(error);
    } catch (err, stack) {
      _log.error('[write]: $err', error: err, stack: stack);
      return Failure(
        StorageAppError.storage(
          message: 'Failed to write key: $key',
          details: err,
        ),
      );
    }
  }
}
