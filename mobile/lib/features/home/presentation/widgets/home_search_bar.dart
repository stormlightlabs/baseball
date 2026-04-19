import 'package:flutter/material.dart';

class HomeSearchBar extends StatelessWidget {
  const HomeSearchBar({
    super.key,
    required this.controller,
    required this.loading,
    required this.onQueryChanged,
    required this.onRunSearch,
  });

  final TextEditingController controller;
  final bool loading;
  final ValueChanged<String> onQueryChanged;
  final VoidCallback onRunSearch;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(
          child: TextField(
            controller: controller,
            onChanged: onQueryChanged,
            onSubmitted: (_) => onRunSearch(),
            decoration: const InputDecoration(
              hintText: '"Babe Ruth", "1927 Yankees"...',
              border: OutlineInputBorder(),
              isDense: true,
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(onPressed: loading ? null : onRunSearch, child: const Text('Go')),
      ],
    );
  }
}
