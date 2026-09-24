import 'dart:async';

import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/viewmodel/onboarding_coordinator_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final now = DateTime.utc(2026, 9, 19, 12);
  late _FakeOnboardingUsecase usecase;
  late OnboardingCoordinatorViewmodel viewmodel;

  OnboardingRegistrationDraft draft(OnboardingStep step) {
    return OnboardingRegistrationDraft(
      step: step,
      expiresAt: now.add(const Duration(hours: 24)),
    );
  }

  setUp(() {
    usecase = _FakeOnboardingUsecase(draft(OnboardingStep.userName));
    viewmodel = OnboardingCoordinatorViewmodel(usecase);
  });

  tearDown(() => viewmodel.dispose());

  test('exposes the step owned by the use case', () {
    expect(viewmodel.step, OnboardingStep.userName);

    usecase.currentDraft = draft(OnboardingStep.emailVerification);

    expect(viewmodel.step, OnboardingStep.emailVerification);
  });

  test('initializes and exposes the restored step', () async {
    usecase.initializeResult = Success(draft(OnboardingStep.password));

    await viewmodel.initializeCommand.execute();

    expect(viewmodel.initializeCommand.isSuccess, isTrue);
    expect(viewmodel.step, OnboardingStep.password);
    expect(usecase.initializeCalls, 1);
  });

  test('preserves an initialization failure', () async {
    const error = AppError(
      code: AppErrorCode.storageError,
      message: 'read failed',
    );
    usecase.initializeResult = const Failure(error);

    await viewmodel.initializeCommand.execute();

    expect(viewmodel.initializeCommand.isFailure, isTrue);
    expect(viewmodel.initializeCommand.error, same(error));
    expect(viewmodel.step, OnboardingStep.userName);
  });

  test('prevents a repeated initialization while running', () async {
    final completer = Completer<Result<OnboardingRegistrationDraft>>();
    usecase.initializeFuture = completer.future;

    final first = viewmodel.initializeCommand.execute();
    final second = viewmodel.initializeCommand.execute();

    expect(usecase.initializeCalls, 1);
    completer.complete(Success(draft(OnboardingStep.email)));
    await Future.wait([first, second]);

    expect(usecase.initializeCalls, 1);
    expect(viewmodel.step, OnboardingStep.email);
  });
}

class _FakeOnboardingUsecase implements OnboardingUsecase {
  _FakeOnboardingUsecase(this.currentDraft)
    : initializeResult = Success(currentDraft!);

  @override
  OnboardingRegistrationDraft? currentDraft;

  Result<OnboardingRegistrationDraft> initializeResult;
  Future<Result<OnboardingRegistrationDraft>>? initializeFuture;
  int initializeCalls = 0;

  @override
  AsyncResult<OnboardingRegistrationDraft> initialize() async {
    initializeCalls++;
    final result = await (initializeFuture ?? Future.value(initializeResult));
    if (result.isSuccess) currentDraft = result.value!;
    return result;
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(String email) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() {
    throw UnimplementedError();
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) {
    throw UnimplementedError();
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() {
    throw UnimplementedError();
  }

  @override
  AsyncResult<Unit> cancelRegistration() {
    throw UnimplementedError();
  }

  @override
  AsyncResult<Unit> createAccount(String password) {
    throw UnimplementedError();
  }
}
