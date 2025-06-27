package me.vipulgupta.dice.example;

import static me.vipulgupta.dice.Reponse.Status.Status_ERR;

import com.googlecode.lanterna.TerminalSize;
import com.googlecode.lanterna.gui2.BasicWindow;
import com.googlecode.lanterna.gui2.BorderLayout;
import com.googlecode.lanterna.gui2.BorderLayout.Location;
import com.googlecode.lanterna.gui2.Borders;
import com.googlecode.lanterna.gui2.Button;
import com.googlecode.lanterna.gui2.Direction;
import com.googlecode.lanterna.gui2.LinearLayout;
import com.googlecode.lanterna.gui2.MultiWindowTextGUI;
import com.googlecode.lanterna.gui2.Panel;
import com.googlecode.lanterna.gui2.TextBox;
import com.googlecode.lanterna.gui2.Window;
import com.googlecode.lanterna.screen.Screen;
import com.googlecode.lanterna.terminal.DefaultTerminalFactory;
import java.io.IOException;
import java.util.Objects;
import java.util.Set;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import me.vipulgupta.dice.Exceptions.DiceDbException;
import me.vipulgupta.dice.Reponse.Response;

public class ChatRoom {

  String username;
  DiceDbChatBackend diceDbChatBackend;
  BlockingQueue<Response> messageQueue;
  TextBox messagesBox;
  ExecutorService executorService;
  Screen screen;
  MultiWindowTextGUI gui;
  BasicWindow window;

  ChatRoom(String username, DiceDbChatBackend diceDbChatBackend)
      throws IOException, DiceDbException, InterruptedException {
    this.username = username;
    this.diceDbChatBackend = diceDbChatBackend;
    this.messageQueue = this.diceDbChatBackend.register();
    this.executorService = Executors.newFixedThreadPool(2);
    this.initUi();
    this.receiveMessage();
    this.sendFirstMessage();
    this.gui.addWindowAndWait(this.window);
  }

  public void sendMessage(String message) {
    try {
      diceDbChatBackend.broadcast(message);
    } catch (DiceDbException e) {
      System.err.println("Error sending message: " + e.getMessage());
    }
  }

  public void receiveMessage() {
    this.executorService.submit(() -> {
      while (true) {
        try {
          Response response = messageQueue.take();
          if (response.getStatus() == Status_ERR) {
            System.out.println("Stopping receiving message: " + response.getMessage());
            break;
          }
          String message = response.getGETRes().getValue();
          messagesBox.addLine(message);
        } catch (InterruptedException e) {
          System.out.println("Stopping receiving message: " + e.getMessage());
        }
      }
    });
  }

  public void sendFirstMessage() {
    try {
      String message = username + ": Joined the chat room!";
      diceDbChatBackend.broadcast(message);
    } catch (DiceDbException e) {
      System.err.println("Error sending message: " + e.getMessage());
    }
  }

  public void initUi() throws IOException {
    this.screen = new DefaultTerminalFactory().setInitialTerminalSize(new TerminalSize(80, 20))
        .createScreen();
    this.screen.startScreen();
    this.gui = new MultiWindowTextGUI(this.screen);
    this.window = new BasicWindow("DiceDB Chat Room");
    this.window.setHints(Set.of(Window.Hint.CENTERED,
        Window.Hint.NO_POST_RENDERING,
        Window.Hint.EXPANDED));
    Panel rootPanel = new Panel(new BorderLayout());

    this.messagesBox = new TextBox(new TerminalSize(80, 12), TextBox.Style.MULTI_LINE);
    this.messagesBox.setReadOnly(true);
    rootPanel.addComponent(this.messagesBox, Location.TOP);

    Panel inputPanel = new Panel(new LinearLayout(Direction.HORIZONTAL));
    TextBox inputBox = new TextBox("Type Message Here...")
        .setPreferredSize(new TerminalSize(60, 1));

    Button sendButton = new Button("Send").setSize(new TerminalSize(20, 1));
    sendButton.addListener((button1) -> {
      String message = inputBox.getText();
      if (message.isEmpty()) {
        return;
      }
      if (message.equalsIgnoreCase("exit")) {
        this.close();
        return;
      }
      inputBox.setText("");
      sendMessage(username + ": " + message);
    });

    inputPanel.addComponent(inputBox.withBorder(Borders.singleLine()));
    inputPanel.addComponent(sendButton.withBorder(Borders.singleLine()));

    rootPanel.addComponent(inputPanel);
    this.window.setComponent(rootPanel);

    inputBox.setCaretPosition(inputBox.getCaretPosition().getColumn());
    inputBox.takeFocus();
  }

  public void close() {
    try {
      messageQueue.put(Response.newBuilder().setStatus(Status_ERR)
          .setMessage("Stopping receiving message").build());
      executorService.shutdown();
      if (!executorService.awaitTermination(10, java.util.concurrent.TimeUnit.SECONDS)) {
        executorService.shutdownNow();
      }
      this.diceDbChatBackend.close();
      this.screen.close();
    } catch (IOException e) {
      System.err.println("Error closing chat room: " + e.getMessage());
    } catch (InterruptedException e) {
      Thread.currentThread().interrupt();
    }

  }

}
