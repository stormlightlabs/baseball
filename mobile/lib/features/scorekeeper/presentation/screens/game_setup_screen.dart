import 'package:bigfly_mobile/features/scorekeeper/data/models/scorecard_models.dart';
import 'package:bigfly_mobile/features/scorekeeper/data/repositories/scorecard_repository.dart';
import 'package:bigfly_mobile/features/scorekeeper/presentation/scorekeeper_routes.dart';
import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';

class GameSetupScreen extends StatefulWidget {
  const GameSetupScreen({super.key, required this.repository});

  final ScorecardRepository repository;

  @override
  State<GameSetupScreen> createState() => _GameSetupScreenState();
}

class _GameSetupScreenState extends State<GameSetupScreen> {
  final _formKey = GlobalKey<FormState>();
  final _awayController = TextEditingController();
  final _homeController = TextEditingController();
  final _venueController = TextEditingController();
  DateTime _date = DateTime.now();
  bool _saving = false;

  @override
  void dispose() {
    _awayController.dispose();
    _homeController.dispose();
    _venueController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('New Scorecard Game')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: <Widget>[
          Form(
            key: _formKey,
            child: Column(
              children: <Widget>[
                TextFormField(
                  controller: _awayController,
                  decoration: const InputDecoration(labelText: 'Away Team Name'),
                  validator: (value) => _required(value, 'Away team is required'),
                ),
                const SizedBox(height: 10),
                TextFormField(
                  controller: _homeController,
                  decoration: const InputDecoration(labelText: 'Home Team Name'),
                  validator: (value) => _required(value, 'Home team is required'),
                ),
                const SizedBox(height: 10),
                TextFormField(
                  controller: _venueController,
                  decoration: const InputDecoration(labelText: 'Venue (optional)'),
                ),
                const SizedBox(height: 14),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: Text(
                        'Date: ${_date.year}-${_date.month.toString().padLeft(2, '0')}-${_date.day.toString().padLeft(2, '0')}',
                      ),
                    ),
                    TextButton(onPressed: _saving ? null : _pickDate, child: const Text('Select Date')),
                  ],
                ),
                const SizedBox(height: 20),
                FilledButton.icon(
                  onPressed: _saving ? null : _startScoring,
                  icon: _saving
                      ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                      : const Icon(Icons.play_arrow),
                  label: const Text('Start Scoring'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _date,
      firstDate: DateTime(1900),
      lastDate: DateTime(2100),
    );
    if (picked == null) {
      return;
    }
    setState(() => _date = picked);
  }

  Future<void> _startScoring() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() => _saving = true);
    final uuid = const Uuid().v4();
    final awayName = _awayController.text.trim();
    final homeName = _homeController.text.trim();

    await widget.repository.createGame(
      ScorecardGameDraft(
        uuid: uuid,
        awayTeamName: awayName,
        awayTeamAbbreviation: _abbr(awayName),
        homeTeamName: homeName,
        homeTeamAbbreviation: _abbr(homeName),
        venue: _venueController.text.trim().isEmpty ? null : _venueController.text.trim(),
        gameDate: _date,
        status: ScorecardStatus.inProgress,
        lineups: const <ScorecardLineupSlot>[],
      ),
    );

    if (!mounted) {
      return;
    }
    setState(() => _saving = false);
    Navigator.of(context).pushReplacementNamed(ScorekeeperRoutes.activeGame(uuid));
  }

  static String _abbr(String value) {
    final tokens = value
        .trim()
        .split(RegExp(r'\s+'))
        .where((token) => token.isNotEmpty)
        .map((token) => token[0].toUpperCase())
        .take(3)
        .join();
    if (tokens.length == 3) {
      return tokens;
    }
    final compact = value.replaceAll(RegExp(r'[^A-Za-z]'), '').toUpperCase();
    if (compact.length >= 3) {
      return compact.substring(0, 3);
    }
    return compact.padRight(3, 'X');
  }

  static String? _required(String? value, String message) {
    if (value == null || value.trim().isEmpty) {
      return message;
    }
    return null;
  }
}
