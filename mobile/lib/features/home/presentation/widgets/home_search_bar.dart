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
    final colorScheme = Theme.of(context).colorScheme;
    return SizedBox(
      height: 44,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colorScheme.surface,
          border: Border.all(color: colorScheme.outlineVariant),
          borderRadius: BorderRadius.circular(12),
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: Row(
            children: <Widget>[
              Expanded(
                child: TextField(
                  controller: controller,
                  onChanged: onQueryChanged,
                  onSubmitted: (_) => onRunSearch(),
                  decoration: const InputDecoration(
                    hintText: '"Babe Ruth", "1927 Yankees"...',
                    border: InputBorder.none,
                    isDense: true,
                    contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                  ),
                ),
              ),
              SizedBox(
                height: double.infinity,
                child: FilledButton(
                  style: FilledButton.styleFrom(
                    shape: const RoundedRectangleBorder(borderRadius: BorderRadius.zero),
                    minimumSize: const Size(56, 44),
                    padding: const EdgeInsets.symmetric(horizontal: 14),
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                  onPressed: loading ? null : onRunSearch,
                  child: const Text('Go'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
