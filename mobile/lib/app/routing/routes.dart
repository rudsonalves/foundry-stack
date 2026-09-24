abstract interface class AppRoutes {
  String get routePath;
  String get routeName;
}

enum BaseRoutes implements AppRoutes {
  splash('/splash', 'Splash'),
  home('/home', 'Home'),
  register('/register', 'Register'),
  forgotPassword('/forgot-password', 'ForgotPassword');

  const BaseRoutes(this.routePath, this.routeName);

  @override
  final String routePath;

  @override
  final String routeName;
}

enum AuthRoutes implements AppRoutes {
  login('/login', 'Login');

  const AuthRoutes(this.routePath, this.routeName);

  @override
  final String routePath;

  @override
  final String routeName;
}
