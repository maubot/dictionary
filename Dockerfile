FROM maubot/plugin-base AS builder

COPY . /go/src/maubot.xyz/dictionary
RUN go build -buildmode=plugin -o /maubot-plugins/dictionary.mbp maubot.xyz/dictionary

FROM alpine:latest
COPY --from=builder /maubot-plugins/dictionary.mbp /maubot-plugins/dictionary.mbp
VOLUME /output
CMD ["cp", "/maubot-plugins/dictionary.mbp", "/output/dictionary.mbp"]
