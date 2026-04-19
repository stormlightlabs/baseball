import 'package:flutter/material.dart';

class InlineErrorText extends StatelessWidget {
  const InlineErrorText(this.message, {super.key});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Text(
      message,
      style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Theme.of(context).colorScheme.error),
    );
  }
}
