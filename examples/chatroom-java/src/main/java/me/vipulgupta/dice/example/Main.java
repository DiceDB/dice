package me.vipulgupta.dice.example;

import java.io.IOException;
import me.vipulgupta.dice.Exceptions.DiceDbException;

public class Main {

  public static void main(String[] args) throws DiceDbException, InterruptedException, IOException {

    String username = args.length > 0 ? args[0] : "Anonymous";
    DiceDbChatBackend chatBackend = new DiceDbChatBackend(username);
    ChatRoom chatRoom = new ChatRoom(username, chatBackend);
    Runtime.getRuntime().addShutdownHook(new Thread(chatRoom::close));

  }

}
