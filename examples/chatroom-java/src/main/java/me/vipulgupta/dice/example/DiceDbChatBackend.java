package me.vipulgupta.dice.example;

import java.util.List;
import java.util.concurrent.BlockingQueue;
import me.vipulgupta.dice.DiceDbManager;
import me.vipulgupta.dice.Exceptions.DiceDbException;
import me.vipulgupta.dice.Reponse.Response;

class DiceDbChatBackend {

  BlockingQueue<Response> messageQueue;
  DiceDbManager diceDbManager;
  String username;

  DiceDbChatBackend(String username) throws DiceDbException, InterruptedException {
    this.diceDbManager = new DiceDbManager("localhost", 7379, 2, 2);
    this.username = username;
  }

  public BlockingQueue<Response> register() throws DiceDbException, InterruptedException {
    messageQueue = diceDbManager.watch("GET.WATCH", List.of("last_message"));
    messageQueue.take(); // consume the first message
    return this.messageQueue;
  }

  public void broadcast(String message) throws DiceDbException {
    this.diceDbManager.fire("SET", List.of("last_message", message));
  }

  public void close() {
    diceDbManager.close();
  }
}
