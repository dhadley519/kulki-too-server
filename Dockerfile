FROM golang:alpine3.17 as webbuilder

WORKDIR /kulki

RUN apk add npm
RUN apk add openssh
RUN apk add git

RUN mkdir -p /root/.ssh
RUN chmod 0700 /root/.ssh
RUN ssh-keyscan github.com > /root/.ssh/known_hosts

COPY ssh/* /root/.ssh/
RUN chmod 600 /root/.ssh/id_ed25519
RUN chmod 600 /root/.ssh/id_ed25519.pub
RUN chmod 600 /root/.ssh/config

RUN git clone --depth 1 --branch react git@github.com:dhadley519/kulki-too.git

WORKDIR /kulki/kulki-too

RUN npm install
RUN npm run build

FROM golang:alpine3.17 as gobuilder

WORKDIR /kulki

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

RUN mkdir -p ./database
RUN mkdir -p ./game
RUN mkdir -p ./web

COPY main.go ./main.go
COPY database/*.go ./database/
COPY game/*.go ./game/
COPY web/*.go ./web/

RUN go build -v -o ./kulki ./main.go

FROM golang:alpine3.17

WORKDIR /kulki

RUN mkdir -p ./public/assets
COPY --from=webbuilder kulki/kulki-too/dist/index.html ./public/index.html
COPY --from=webbuilder kulki/kulki-too/dist/assets/index.js ./public/assets/index.js
COPY --from=webbuilder kulki/kulki-too/dist/assets/index.css ./public/assets/index.css
COPY --from=gobuilder kulki/kulki ./

EXPOSE 8080

ENV POSTGRES_USER=skinny_dog
ARG POSTGRES_PASSWORD
ENV POSTGRES_PASSWORD=$POSTGRES_PASSWORD
ENV POSTGRES_DB=db

CMD ["/kulki/kulki"]