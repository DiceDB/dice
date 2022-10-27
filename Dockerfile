FROM golang:1.17

RUN apt-get update && apt-get -y dist-upgrade
RUN apt install -y netcat

# Set destination for COPY
WORKDIR /app

# copy project files to /app dir
COPY . .


# Build
RUN go build -o /dice

# Run
CMD [ "/dice" ]
