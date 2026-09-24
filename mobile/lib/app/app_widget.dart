import 'package:material_ui/material_ui.dart';
import 'package:go_router/go_router.dart';

import '../ui/components/themes/material_theme.dart';
import '../ui/components/themes/text_theme.dart';
import 'dependencies/root_container.dart';
import 'routing/router.dart';

class AppWidget extends StatefulWidget {
  final RootContainer rootContainer;

  const AppWidget({required this.rootContainer, super.key});

  @override
  State<AppWidget> createState() => _AppWidgetState();
}

class _AppWidgetState extends State<AppWidget> {
  late final MaterialTheme _materialTheme;
  late final GoRouter _router;

  @override
  void initState() {
    super.initState();
    _router = router(
      appContainer: widget.rootContainer.appContainer,
      createOnboardingScope: widget.rootContainer.createOnboardingScope,
      createPasswordResetScope: widget.rootContainer.createPasswordResetScope,
    );
  }

  @override
  void dispose() {
    _router.dispose();
    widget.rootContainer.dispose();
    super.dispose();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    // Initialize once with context-dependent resources
    final textTheme = createTextTheme(Theme.of(context).textTheme);

    _materialTheme = MaterialTheme(textTheme);
  }

  @override
  Widget build(BuildContext context) {
    final brightness = View.of(context).platformDispatcher.platformBrightness;

    final baseTheme = _resolveBaseTheme(brightness);
    final theme = _buildAppTheme(baseTheme);

    return MaterialApp.router(
      title: 'FoundryStack',
      theme: theme,
      debugShowCheckedModeBanner: false,
      routerConfig: _router,
    );
  }

  ThemeData _resolveBaseTheme(Brightness brightness) {
    return brightness == Brightness.light
        ? _materialTheme.light()
        : _materialTheme.dark();
  }

  ThemeData _buildAppTheme(ThemeData base) {
    final colorScheme = base.colorScheme;

    return base.copyWith(
      appBarTheme: base.appBarTheme.copyWith(
        backgroundColor: colorScheme.primaryContainer,
        foregroundColor: colorScheme.onPrimaryContainer,
        titleTextStyle: base.textTheme.titleLarge?.copyWith(
          fontWeight: FontWeight.w600,
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colorScheme.primaryContainer.withValues(alpha: 0.5),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(24),
          borderSide: BorderSide.none,
        ),
      ),
    );
  }
}
