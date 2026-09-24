import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '/app/routing/routes.dart';
import '/core/result/result.dart';
import 'viewmodel/splash_viewmodel.dart';

class SplashPage extends StatefulWidget {
  final SplashViewmodel viewmodel;

  const SplashPage({
    super.key,
    required this.viewmodel,
  });

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage>
    with SingleTickerProviderStateMixin {
  SplashViewmodel get _viewModel => widget.viewmodel;

  late final AnimationController _animationController;
  late final Animation<double> _logoScale;
  late final Animation<double> _logoOpacity;

  bool get _isConnectionFailure => switch (_viewModel.initialize.error?.code) {
    AppErrorCode.networkError || AppErrorCode.timeout => true,
    _ => false,
  };

  String get _errorTitle => _isConnectionFailure
      ? 'Não foi possível conectar.'
      : 'Não foi possível preparar este app para acesso.';

  String get _errorMessage => _isConnectionFailure
      ? 'Verifique sua conexão e tente novamente.'
      : 'Verifique o armazenamento do dispositivo e tente novamente.';

  @override
  void initState() {
    super.initState();

    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 2200),
    );

    _logoScale = Tween<double>(begin: 0.8, end: 1).animate(
      CurvedAnimation(
        parent: _animationController,
        curve: Curves.easeOutCubic,
      ),
    );

    _logoOpacity = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(
        parent: _animationController,
        curve: const Interval(0.1, 0.75, curve: Curves.easeOut),
      ),
    );

    _viewModel.initialize.addListener(_onInitializeChanged);
    _animationController.addStatusListener(_onAnimationStatusChanged);
    _animationController.forward();
    _viewModel.initialize.execute();
  }

  @override
  void dispose() {
    _viewModel.initialize.removeListener(_onInitializeChanged);
    _animationController.removeStatusListener(_onAnimationStatusChanged);
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      body: Container(
        width: double.infinity,
        height: double.infinity,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [
              colorScheme.surface,
              colorScheme.surfaceContainerHighest.withValues(alpha: 0.7),
            ],
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
          ),
        ),
        child: Center(
          child: AnimatedBuilder(
            animation: Listenable.merge([
              _animationController,
              _viewModel.initialize,
            ]),
            builder: (context, _) {
              final showRecoverableError = _viewModel.initialize.isFailure;

              return Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Opacity(
                      opacity: _logoOpacity.value,
                      child: Transform.scale(
                        scale: _logoScale.value,
                        child: Card(
                          color: colorScheme.onPrimary,
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  Icons.casino_outlined,
                                  size: 72,
                                  color: colorScheme.primary,
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  'FoundryStack',
                                  style: Theme.of(
                                    context,
                                  ).textTheme.headlineSmall,
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                    if (showRecoverableError) ...[
                      const SizedBox(height: 28),
                      ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 420),
                        child: Column(
                          spacing: 14,
                          children: [
                            Text(
                              _errorTitle,
                              style: Theme.of(context).textTheme.titleMedium,
                              textAlign: TextAlign.center,
                            ),
                            Text(
                              _errorMessage,
                              style: Theme.of(context).textTheme.bodyMedium
                                  ?.copyWith(
                                    color: colorScheme.onSurfaceVariant,
                                  ),
                              textAlign: TextAlign.center,
                            ),
                            FilledButton.icon(
                              onPressed: _viewModel.initialize.isRunning
                                  ? null
                                  : _retryInitialize,
                              icon: const Icon(Icons.refresh_rounded),
                              label: const Text('Tentar novamente'),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  void _retryInitialize() {
    _viewModel.initialize.execute();
  }

  void _onInitializeChanged() {
    if (_viewModel.initialize.isRunning || _viewModel.initialize.isFailure) {
      return;
    }

    _navigateWhenReady();
  }

  void _onAnimationStatusChanged(AnimationStatus status) {
    if (status != AnimationStatus.completed) return;

    _navigateWhenReady();
  }

  void _navigateWhenReady() {
    if (!_animationController.isCompleted || !_viewModel.initialize.isSuccess) {
      return;
    }

    final destination = _viewModel.initialize.value;
    if (destination == null) return;

    switch (destination) {
      case .login:
        context.goNamed(AuthRoutes.login.routeName);
      case .onboarding:
        context.goNamed(BaseRoutes.register.routeName);
      case .home:
        context.goNamed(BaseRoutes.home.routeName);
    }
  }
}
