import 'package:bigfly_mobile/features/scorekeeper/presentation/widgets/export_sheet.dart';
import 'package:flutter/material.dart';

class ScorecardExportLauncherScreen extends StatefulWidget {
  const ScorecardExportLauncherScreen({super.key, required this.gameUuid});

  final String gameUuid;

  @override
  State<ScorecardExportLauncherScreen> createState() => _ScorecardExportLauncherScreenState();
}

class _ScorecardExportLauncherScreenState extends State<ScorecardExportLauncherScreen> {
  var _opened = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_opened) {
      return;
    }
    _opened = true;

    WidgetsBinding.instance.addPostFrameCallback((_) async {
      if (!mounted) {
        return;
      }

      await showModalBottomSheet<void>(
        context: context,
        isScrollControlled: true,
        builder: (_) => ExportSheet(gameUuid: widget.gameUuid),
      );

      if (mounted && Navigator.of(context).canPop()) {
        Navigator.of(context).pop();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: SizedBox.shrink());
  }
}
