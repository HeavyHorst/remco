FROM golang:buster AS build

WORKDIR /src
COPY . .
RUN make

FROM scratch AS bin
COPY --from=build /src/bin/remco /remco

