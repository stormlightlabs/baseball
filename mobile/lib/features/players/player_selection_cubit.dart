import 'package:flutter_bloc/flutter_bloc.dart';

class PlayerSelectionCubit extends Cubit<String?> {
  PlayerSelectionCubit() : super(null);

  void selectPlayer(String playerId) => emit(playerId);
}
