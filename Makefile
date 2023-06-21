GO = go
BUILD = build
SERVER = chat-server
HASHERDIR = pkg/hasher
SRVSOURCE = api/server
CMD = cmd
BIN = bin

server: ${HASHERDIR}/hasher.go ${SRVSOURCE}/server.go ${SRVSOURCE}/util.go ${SRVSOURCE}/router.go
	${GO} ${BUILD} -o ${BIN}/${SERVER}.exe ${CMD}/${SERVER}.go