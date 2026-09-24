import 'package:go_router/go_router.dart';

import '/app/dependencies/app_container.dart';
import '/app/routing/routes.dart';
import '/ui/pages/auth/login/login_page.dart';
import '/ui/pages/auth/login/viewmodel/login_viewmodel.dart';
import '../animations_page/app_custom_transaction.dart';

List<RouteBase> authRoutes({required AppContainer appContainer}) => [
  GoRoute(
    path: AuthRoutes.login.routePath,
    name: AuthRoutes.login.routeName,
    pageBuilder: (context, state) => AppCustomTransactionPage(
      key: state.pageKey,
      child: LoginPage(
        viewmodel: appContainer.get<LoginViewmodel>(),
      ),
    ),
  ),
];
