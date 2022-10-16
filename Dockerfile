FROM golang:1.17

RUN apt-get update && apt-get -y dist-upgrade
RUN apt-get -y install build-essential libssl-dev libffi-dev libblas3 libc6 liblapack3 gcc python3-dev python3-pip cython3
RUN apt-get -y install python3-numpy python3-scipy 
RUN apt install -y netcat

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY . .


# Build
RUN go build -o /dice

# Run
CMD [ "/dice" ]
