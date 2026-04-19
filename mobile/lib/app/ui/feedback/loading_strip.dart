import 'package:flutter/material.dart';

class LoadingStrip extends StatelessWidget {
  const LoadingStrip({super.key, this.minHeight = 2});

  final double minHeight;

  @override
  Widget build(BuildContext context) {
    return LinearProgressIndicator(minHeight: minHeight);
  }
}
