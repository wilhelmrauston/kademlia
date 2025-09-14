FROM golang

WORKDIR /app
RUN cd /app

RUN apt-get update 
RUN git clone https://github.com/wilhelmrauston/kademlia.git
RUN cd kademlia && git pull

RUN cd kademlia && go build -o main .

EXPOSE 8001

ENTRYPOINT [ "kademlia/main" ]