import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_registration_draft.dart';
import 'package:foundry_stack_mobile/domain/common/onboarding/models/onboarding_step.dart';
import 'package:foundry_stack_mobile/domain/usecases/onboarding/onboarding_usecase.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/email/viewmodel/onboarding_email_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/email_verification/viewmodel/onboarding_email_verification_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/password/viewmodel/onboarding_password_viewmodel.dart';
import 'package:foundry_stack_mobile/ui/pages/onboarding/user_name/viewmodel/onboarding_user_name_viewmodel.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  final draft = OnboardingRegistrationDraft(
    step: OnboardingStep.userName,
    name: 'Ada',
    email: 'ada@example.com',
    expiresAt: DateTime.utc(2026, 9, 20),
  );

  test('user name ViewModel delegates the name to the use case', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingUserNameViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.advanceNameCommand.execute('Grace');

    expect(usecase.name, 'Grace');
    expect(viewmodel.advanceNameCommand.isSuccess, isTrue);
    expect(viewmodel.name, 'Ada');
  });

  test('email ViewModel delegates the email to the use case', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingEmailViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.advanceEmailCommand.execute('grace@example.com');

    expect(usecase.email, 'grace@example.com');
    expect(viewmodel.advanceEmailCommand.isSuccess, isTrue);
  });

  test('email ViewModel delegates returning to the previous step', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingEmailViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.returnToPreviousStepCommand.execute();

    expect(usecase.returnCalls, 1);
    expect(viewmodel.returnToPreviousStepCommand.isSuccess, isTrue);
  });

  test('email ViewModel delegates registration cancellation', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingEmailViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.cancelRegistrationCommand.execute();

    expect(usecase.cancelCalls, 1);
    expect(viewmodel.cancelRegistrationCommand.isSuccess, isTrue);
  });

  test('verification ViewModel delegates confirmation and resend', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingEmailVerificationViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.confirmVerificationCommand.execute('012345');
    await viewmodel.resendVerificationCommand.execute();

    expect(usecase.otp, '012345');
    expect(usecase.resendCalls, 1);
    expect(viewmodel.confirmVerificationCommand.isSuccess, isTrue);
    expect(viewmodel.resendVerificationCommand.isSuccess, isTrue);
  });

  test(
    'verification ViewModel delegates returning to the previous step',
    () async {
      final usecase = _FakeOnboardingUsecase(draft);
      final viewmodel = OnboardingEmailVerificationViewmodel(usecase);
      addTearDown(viewmodel.dispose);

      await viewmodel.returnToPreviousStepCommand.execute();

      expect(usecase.returnCalls, 1);
      expect(viewmodel.returnToPreviousStepCommand.isSuccess, isTrue);
    },
  );

  test('password ViewModel delegates the password to the use case', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingPasswordViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.createAccountCommand.execute('secret123');

    expect(usecase.password, 'secret123');
    expect(viewmodel.createAccountCommand.isSuccess, isTrue);
  });

  test('password ViewModel delegates returning to the previous step', () async {
    final usecase = _FakeOnboardingUsecase(draft);
    final viewmodel = OnboardingPasswordViewmodel(usecase);
    addTearDown(viewmodel.dispose);

    await viewmodel.returnToPreviousStepCommand.execute();

    expect(usecase.returnCalls, 1);
    expect(viewmodel.returnToPreviousStepCommand.isSuccess, isTrue);
  });

  test('email ViewModel disposes its commands', () {
    final viewmodel = OnboardingEmailViewmodel(
      _FakeOnboardingUsecase(draft),
    );

    viewmodel.dispose();

    expect(
      () => viewmodel.advanceEmailCommand.addListener(() {}),
      throwsFlutterError,
    );
    expect(
      () => viewmodel.returnToPreviousStepCommand.addListener(() {}),
      throwsFlutterError,
    );
    expect(
      () => viewmodel.cancelRegistrationCommand.addListener(() {}),
      throwsFlutterError,
    );
  });

  test('email verification ViewModel disposes its commands', () {
    final viewmodel = OnboardingEmailVerificationViewmodel(
      _FakeOnboardingUsecase(draft),
    );

    viewmodel.dispose();

    expect(
      () => viewmodel.confirmVerificationCommand.addListener(() {}),
      throwsFlutterError,
    );
    expect(
      () => viewmodel.resendVerificationCommand.addListener(() {}),
      throwsFlutterError,
    );
    expect(
      () => viewmodel.returnToPreviousStepCommand.addListener(() {}),
      throwsFlutterError,
    );
  });

  test('password ViewModel disposes its commands', () {
    final viewmodel = OnboardingPasswordViewmodel(
      _FakeOnboardingUsecase(draft),
    );

    viewmodel.dispose();

    expect(
      () => viewmodel.createAccountCommand.addListener(() {}),
      throwsFlutterError,
    );
    expect(
      () => viewmodel.returnToPreviousStepCommand.addListener(() {}),
      throwsFlutterError,
    );
  });

  test('user name ViewModel disposes its commands', () {
    final viewmodel = OnboardingUserNameViewmodel(
      _FakeOnboardingUsecase(draft),
    );

    viewmodel.dispose();

    expect(
      () => viewmodel.advanceNameCommand.addListener(() {}),
      throwsFlutterError,
    );
  });
}

class _FakeOnboardingUsecase implements OnboardingUsecase {
  _FakeOnboardingUsecase(this.currentDraft);

  @override
  OnboardingRegistrationDraft? currentDraft;

  String? name;
  String? email;
  String? otp;
  String? password;
  int resendCalls = 0;
  int returnCalls = 0;
  int cancelCalls = 0;

  @override
  AsyncResult<OnboardingRegistrationDraft> initialize() async {
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithName(String name) async {
    this.name = name;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> advanceWithEmail(
    String email,
  ) async {
    this.email = email;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> returnToPreviousStep() async {
    returnCalls++;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> confirmEmailVerification(
    String otp,
  ) async {
    this.otp = otp;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<OnboardingRegistrationDraft> resendEmailVerification() async {
    resendCalls++;
    return Success(currentDraft!);
  }

  @override
  AsyncResult<Unit> createAccount(String password) async {
    this.password = password;
    return const Success(unit);
  }

  @override
  AsyncResult<Unit> cancelRegistration() async {
    cancelCalls++;
    return const Success(unit);
  }
}
