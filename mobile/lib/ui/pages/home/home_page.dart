import 'package:material_ui/material_ui.dart';

import '../../components/base/safe_scaffold.dart';
import 'viewmodel/home_viewmodel.dart';

class HomePage extends StatefulWidget {
  final HomeViewmodel viewmodel;

  const HomePage({super.key, required this.viewmodel});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  @override
  Widget build(BuildContext context) {
    return SafeScaffold(
      appBar: AppBar(
        title: const Text('FoundryStack'),
      ),
      body: Column(
        mainAxisAlignment: .center,
        crossAxisAlignment: .center,
        children: [
          Text('No data'),
        ],
      ),
    );
  }
}
