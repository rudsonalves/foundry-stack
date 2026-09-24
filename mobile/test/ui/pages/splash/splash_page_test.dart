import 'dart:async';

import 'package:foundry_stack_mobile/app/routing/routes.dart';
import 'package:foundry_stack_mobile/core/result/result.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/bootstrap_usecase.dart';
import 'package:foundry_stack_mobile/domain/usecases/bootstrap/models/bootstrap_destination.dart';
import 'package:foundry_stack_mobile/ui/pages/splash/splash_page.dart';
import 'package:foundry_stack_mobile/ui/pages/splash/viewmodel/splash_viewmodel.dart';
import 'package:material_ui/material_ui.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

void main() {
  for (final brightness in [Brightness.light, Brightness.dark]) {
    testWidgets('renders the temporary brand in ${brightness.name} theme', (
      tester,
    ) async {
      final completer = Completer<Result<BootstrapDestination>>();
      final usecase = _FakeBootstrapUsecase(() => completer.future);

      await tester.pumpWidget(
        MaterialApp(
          theme: ThemeData(brightness: brightness),
          home: SplashPage(viewmodel: SplashViewmodel(usecase)),
        ),
      );
      await tester.pump();

      expect(find.byIcon(Icons.casino_outlined), findsOneWidget);
      expect(find.text('FoundryStack'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('shows the animation before replacing Splash with Login', (
    tester,
  ) async {
    final usecase = _FakeBootstrapUsecase(
      () async => const Success(BootstrapDestination.login),
    );
    final router = _testRouter(SplashViewmodel(usecase));

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1100));

    expect(find.text('FoundryStack'), findsOneWidget);
    expect(find.text('Login destination'), findsNothing);

    await tester.pumpAndSettle();

    expect(find.text('Login destination'), findsOneWidget);
    expect(usecase.initializeCalls, 1);
    expect(router.canPop(), isFalse);
  });

  testWidgets('replaces Splash with Home after the animation', (tester) async {
    final usecase = _FakeBootstrapUsecase(
      () async => const Success(BootstrapDestination.home),
    );
    final router = _testRouter(SplashViewmodel(usecase));

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 2199));

    expect(find.text('FoundryStack'), findsOneWidget);
    expect(find.text('Home destination'), findsNothing);

    await tester.pumpAndSettle();

    expect(find.text('Home destination'), findsOneWidget);
    expect(usecase.initializeCalls, 1);
    expect(router.canPop(), isFalse);
  });

  testWidgets('replaces Splash with Onboarding after the animation', (
    tester,
  ) async {
    final usecase = _FakeBootstrapUsecase(
      () async => const Success(BootstrapDestination.onboarding),
    );
    final router = _testRouter(SplashViewmodel(usecase));

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.pumpAndSettle();

    expect(find.text('Onboarding destination'), findsOneWidget);
    expect(usecase.initializeCalls, 1);
    expect(router.canPop(), isFalse);
  });

  for (final errorCode in [
    AppErrorCode.networkError,
    AppErrorCode.timeout,
  ]) {
    testWidgets('shows the connection message for ${errorCode.name}', (
      tester,
    ) async {
      final usecase = _FakeBootstrapUsecase(
        () async => Failure(
          AppError(code: errorCode, message: '${errorCode.name} failure'),
        ),
      );

      await tester.pumpWidget(
        MaterialApp(home: SplashPage(viewmodel: SplashViewmodel(usecase))),
      );
      await tester.pump();

      expect(find.text('Não foi possível conectar.'), findsOneWidget);
      expect(
        find.text('Verifique sua conexão e tente novamente.'),
        findsOneWidget,
      );
      expect(find.text('Tentar novamente'), findsOneWidget);
    });
  }

  testWidgets('shows the local failure message for a storage error', (
    tester,
  ) async {
    final usecase = _FakeBootstrapUsecase(
      () async => const Failure(
        AppError(
          code: AppErrorCode.storageError,
          message: 'storage failure',
        ),
      ),
    );

    await tester.pumpWidget(
      MaterialApp(home: SplashPage(viewmodel: SplashViewmodel(usecase))),
    );
    await tester.pump();

    expect(
      find.text('Não foi possível preparar este app para acesso.'),
      findsOneWidget,
    );
    expect(
      find.text(
        'Verifique o armazenamento do dispositivo e tente novamente.',
      ),
      findsOneWidget,
    );
    expect(find.text('Tentar novamente'), findsOneWidget);
  });

  testWidgets('retries once and hides the previous failure while loading', (
    tester,
  ) async {
    final retryCompleter = Completer<Result<BootstrapDestination>>();
    late _FakeBootstrapUsecase usecase;
    usecase = _FakeBootstrapUsecase(() {
      if (usecase.initializeCalls == 1) {
        return Future.value(
          const Failure(
            AppError(
              code: AppErrorCode.networkError,
              message: 'network failure',
            ),
          ),
        );
      }

      return retryCompleter.future;
    });
    final router = _testRouter(SplashViewmodel(usecase));

    await tester.pumpWidget(MaterialApp.router(routerConfig: router));
    await tester.pump();

    final retryButton = find.text('Tentar novamente');
    expect(retryButton, findsOneWidget);

    await tester.tap(retryButton);
    await tester.tap(retryButton);
    await tester.pump();

    expect(usecase.initializeCalls, 2);
    expect(find.text('Não foi possível conectar.'), findsNothing);
    expect(find.text('Tentar novamente'), findsNothing);

    retryCompleter.complete(const Success(BootstrapDestination.login));
    await tester.pumpAndSettle();

    expect(find.text('Login destination'), findsOneWidget);
    expect(usecase.initializeCalls, 2);
  });

  testWidgets('does not navigate when disposed before initialization ends', (
    tester,
  ) async {
    final completer = Completer<Result<BootstrapDestination>>();
    final usecase = _FakeBootstrapUsecase(() => completer.future);

    await tester.pumpWidget(
      MaterialApp(home: SplashPage(viewmodel: SplashViewmodel(usecase))),
    );
    await tester.pumpWidget(
      const MaterialApp(home: Text('Disposed destination')),
    );

    completer.complete(const Success(BootstrapDestination.home));
    await tester.pump();

    expect(find.text('Disposed destination'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}

GoRouter _testRouter(SplashViewmodel viewmodel) => GoRouter(
  initialLocation: BaseRoutes.splash.routePath,
  routes: [
    GoRoute(
      path: BaseRoutes.splash.routePath,
      name: BaseRoutes.splash.routeName,
      builder: (context, state) => SplashPage(viewmodel: viewmodel),
    ),
    GoRoute(
      path: AuthRoutes.login.routePath,
      name: AuthRoutes.login.routeName,
      builder: (context, state) => const Scaffold(
        body: Text('Login destination'),
      ),
    ),
    GoRoute(
      path: BaseRoutes.register.routePath,
      name: BaseRoutes.register.routeName,
      builder: (context, state) => const Scaffold(
        body: Text('Onboarding destination'),
      ),
    ),
    GoRoute(
      path: BaseRoutes.home.routePath,
      name: BaseRoutes.home.routeName,
      builder: (context, state) => const Scaffold(
        body: Text('Home destination'),
      ),
    ),
  ],
);

class _FakeBootstrapUsecase implements BootstrapUsecase {
  _FakeBootstrapUsecase(this._initialize);

  final AsyncResult<BootstrapDestination> Function() _initialize;
  int initializeCalls = 0;

  @override
  AsyncResult<BootstrapDestination> initialize() {
    initializeCalls++;
    return _initialize();
  }
}
