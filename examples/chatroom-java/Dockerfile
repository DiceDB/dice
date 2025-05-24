# Use OpenJDK 21 base image
FROM openjdk:21-jdk

WORKDIR /app

COPY target/chatroom.jar .

ENTRYPOINT ["java", "-jar", "chatroom.jar"]
CMD []